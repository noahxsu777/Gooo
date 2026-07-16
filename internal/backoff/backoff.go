package backoff

import (
	"math/rand"
	"time"
)

type Exponential struct {
	Base    time.Duration
	Max     time.Duration
	Jitter  float64
	Retries int
}

func (e Exponential) Duration(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	d := e.Base << attempt
	if d > e.Max {
		d = e.Max
	}
	if e.Jitter <= 0 {
		return d
	}
	factor := 1 + ((rand.Float64()*2 - 1) * e.Jitter)
	if factor < 0.1 {
		factor = 0.1
	}
	return time.Duration(float64(d) * factor)
}

func (e Exponential) CanRetry(attempt int) bool {
	if e.Retries <= 0 {
		return true
	}
	return attempt < e.Retries
}
