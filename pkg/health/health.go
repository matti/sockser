package health

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/matti/sockser/pkg/globals"
	"github.com/matti/sockser/pkg/types"
	"golang.org/x/net/proxy"
)

func healthcheck(upstream *types.Upstream) {
	started := time.Now()

	dialSocksProxy, err := proxy.SOCKS5("tcp", upstream.Address, nil, proxy.Direct)
	if err != nil {
		log.Println("err", upstream.Address, "dialer", err)
		upstream.Healthy = false
		return
	}
	transport := &http.Transport{Dial: dialSocksProxy.Dial}

	client := &http.Client{
		Transport: transport,
		Timeout:   globals.Config.Timeout,
	}

	resp, err := client.Get(globals.Config.HealthcheckUrl)
	if err != nil {
		//log.Println("err", upstream.address, "get", err)
		upstream.Healthy = false
		return
	}
	resp.Body.Close()
	took := time.Since(started)
	upstream.Healthy = true
	upstream.Rtt = took
}

func Run(upstreams []*types.Upstream) {
	var inflight sync.WaitGroup

	go func() {
		for {
			log.Println("-- index: ", globals.Config.Index, " ------------------------")

			for _, upstream := range upstreams {
				log.Println("upstream", upstream.Address, upstream.Healthy, globals.Best == upstream, upstream.Rtt)
			}
			log.Println("fallback", globals.Best == globals.Config.Fallback)
			time.Sleep(globals.Config.Stats)
		}
	}()

	for {
		started := time.Now()
		for _, upstream := range upstreams {
			inflight.Add(1)
			go func(u *types.Upstream) {
				healthcheck(u)
				inflight.Done()
			}(upstream)
		}
		inflight.Wait()
		took := time.Since(started)
		remaining := time.Second*1 - took

		if upstreams[globals.Config.Index].Healthy {
			globals.Best = upstreams[globals.Config.Index]
		} else {
			fallback := true
			for _, upstream := range upstreams {
				if upstream.Healthy {
					globals.Best = upstream
					fallback = false
					break
				}
			}

			if fallback {
				globals.Best = globals.Config.Fallback
			}
		}

		if remaining > 0 {
			time.Sleep(remaining)
		}
	}
}
