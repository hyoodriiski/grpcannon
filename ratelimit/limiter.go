package ratelimit

import (
	"context"
	"time"
)

// Limiter controls the rate of requests per second.
type Limiter struct {
	ticker *time.Ticker
	done   chan struct{}
}

// New creates a Limiter that allows rps requests per second.
// If rps <= 0, no limiting is applied.
func New(rps int) *Limiter {
	if rps <= 0 {
		return &Limiter{}
	}
	interval := time.Second / time.Duration(rps)
	return &Limiter{
		ticker: time.NewTicker(interval),
		done:   make(chan struct{}),
	}
}

// Wait blocks until the next request slot is available or ctx is cancelled.
func (l *Limiter) Wait(ctx context.Context) error {
	if l.ticker == nil {
		return nil
	}
	select {
	case <-l.ticker.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-l.done:
		return context.Canceled
	}
}

// Stop releases resources held by the Limiter.
func (l *Limiter) Stop() {
	if l.ticker == nil {
		return
	}
	l.ticker.Stop()
	close(l.done)
}
