package types

import "time"

type Upstream struct {
	Address string
	Healthy bool
	Rtt     time.Duration
}

type Config struct {
	HealthcheckUrl string
	Timeout        time.Duration
	Fallback       *Upstream
	Index          int
	Stats          time.Duration
}
