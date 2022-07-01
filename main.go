package main

import (
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/matti/betterio"
	"github.com/matti/sockser/pkg/globals"
	"github.com/matti/sockser/pkg/health"
	"github.com/matti/sockser/pkg/types"
)

func main() {
	var addresses []string
	addresses = append(addresses, "127.0.0.1:1081")
	addresses = append(addresses, "127.0.0.1:1082")

	rand.Seed(time.Now().UnixNano())
	globals.Config = &types.Config{
		HealthcheckUrl: "http://127.0.0.1:8000",
		Timeout:        3 * time.Second,
		Fallback:       &types.Upstream{Address: "127.0.0.1:2080"},
		Index:          rand.Intn(len(addresses)),
		Stats:          3 * time.Second,
	}

	var upstreams []*types.Upstream
	for _, address := range addresses {
		upstreams = append(upstreams, &types.Upstream{
			Address: address,
		})
	}

	go health.Run(upstreams)

	for {
		if globals.Best != nil {
			break
		}

		time.Sleep(time.Millisecond * 100)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:1080")
	if err != nil {
		panic(err)
	}

	log.Println("listen 127.0.0.1:1080")
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				log.Printf("accept temp err: %v", ne)
				continue
			}

			log.Panicln("accept", err)
			os.Exit(1)
		}

		go func() {
			for {
				if handle(conn, globals.Best) {
					break
				}

				if err := betterio.CheckReaderOpen(conn); err != nil {
					break
				}

				log.Println(conn.RemoteAddr().String(), "retry")
				time.Sleep(time.Millisecond * 100)
			}
		}()
	}

}

func handle(conn net.Conn, upstream *types.Upstream) bool {
	defer conn.Close()

	var upstreamConn net.Conn

	dialer := net.Dialer{
		Timeout: 500 * time.Millisecond,
	}
	upstreamConn, err := dialer.Dial("tcp", upstream.Address)
	if err != nil {
		log.Println(conn.RemoteAddr().String(), "dial err to", upstream.Address, err)
		return false
	}
	defer upstreamConn.Close()

	// log.Println(conn.RemoteAddr().String(), "copying")
	betterio.CopyBidirUntilCloseAndReturnBytesWritten(conn, upstreamConn)
	return true
}
