// Package backoff provides exponential backoff strategies for retrying
// failed gRPC calls with configurable jitter and maximum delay.
package backoff

import (
	"math"
	"math/rand"
	"time"
)

// Strategy defines the parameters for exponential backoff.
type Strategy struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	Jitter       float64 // fraction in [0, 1]
}

// Default returns a Strategy with sensible defaults.
func Default() Strategy {
	return Strategy{
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     2 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.2,
	}
}

// Delay returns the backoff duration for the given attempt (0-indexed).
// It applies exponential growth capped at MaxDelay, then adds random jitter.
func (s Strategy) Delay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	base := float64(s.InitialDelay) * math.Pow(s.Multiplier, float64(attempt))
	if base > float64(s.MaxDelay) {
		base = float64(s.MaxDelay)
	}
	if s.Jitter > 0 {
		// jitter in [-jitter*base, +jitter*base]
		delta := s.Jitter * base
		base += (rand.Float64()*2 - 1) * delta
		if base < 0 {
			base = 0
		}
	}
	return time.Duration(base)
}

// Steps returns a slice of delay durations for n attempts.
func (s Strategy) Steps(n int) []time.Duration {
	delays := make([]time.Duration, n)
	for i := 0; i < n; i++ {
		delays[i] = s.Delay(i)
	}
	return delays
}
