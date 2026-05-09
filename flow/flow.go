// Package flow provides a token-bucket style request flow controller
// that smooths bursts by enforcing a steady emission rate.
package flow

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrClosed is returned when Wait is called on a closed Flow.
var ErrClosed = errors.New("flow: controller closed")

// Flow controls the rate at which callers are allowed to proceed.
type Flow struct {
	mu       sync.Mutex
	interval time.Duration
	last     time.Time
	closed   bool
}

// New creates a Flow that allows at most rps requests per second.
// If rps is zero or negative the Flow imposes no delay.
func New(rps float64) *Flow {
	var interval time.Duration
	if rps > 0 {
		interval = time.Duration(float64(time.Second) / rps)
	}
	return &Flow{interval: interval}
}

// Wait blocks until the caller is allowed to proceed or the context is
// cancelled. Returns ErrClosed if the Flow has been closed.
func (f *Flow) Wait(ctx context.Context) error {
	f.mu.Lock()
	if f.closed {
		f.mu.Unlock()
		return ErrClosed
	}
	if f.interval == 0 {
		f.mu.Unlock()
		return nil
	}
	now := time.Now()
	delay := f.interval - now.Sub(f.last)
	if delay <= 0 {
		f.last = now
		f.mu.Unlock()
		return nil
	}
	f.last = now.Add(delay)
	f.mu.Unlock()

	select {
	case <-time.After(delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close permanently closes the Flow. Subsequent calls to Wait return ErrClosed.
func (f *Flow) Close() {
	f.mu.Lock()
	f.closed = true
	f.mu.Unlock()
}
