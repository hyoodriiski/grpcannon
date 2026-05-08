// Package pacing provides a request pacing controller that smooths
// outgoing load by spacing requests evenly across a time window.
package pacing

import (
	"context"
	"sync"
	"time"
)

// Pacer spaces requests evenly at a target rate (requests per second).
type Pacer struct {
	mu       sync.Mutex
	interval time.Duration
	last     time.Time
}

// New returns a Pacer that allows at most rps requests per second.
// If rps is zero or negative the Pacer imposes no delay.
func New(rps float64) *Pacer {
	var interval time.Duration
	if rps > 0 {
		interval = time.Duration(float64(time.Second) / rps)
	}
	return &Pacer{interval: interval}
}

// Wait blocks until the next request slot is available or ctx is cancelled.
// Returns ctx.Err() if the context is cancelled while waiting.
func (p *Pacer) Wait(ctx context.Context) error {
	if p.interval <= 0 {
		return ctx.Err()
	}

	p.mu.Lock()
	now := time.Now()
	var delay time.Duration
	if !p.last.IsZero() {
		next := p.last.Add(p.interval)
		if next.After(now) {
			delay = next.Sub(now)
		}
	}
	p.last = now.Add(delay)
	p.mu.Unlock()

	if delay <= 0 {
		return ctx.Err()
	}

	select {
	case <-time.After(delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Interval returns the configured inter-request spacing.
func (p *Pacer) Interval() time.Duration {
	return p.interval
}
