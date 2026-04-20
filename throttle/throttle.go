// Package throttle provides a concurrency limiter using a semaphore.
package throttle

import (
	"context"
	"errors"
)

// Throttle limits the number of concurrent operations.
type Throttle struct {
	sem chan struct{}
}

// New creates a Throttle that allows at most n concurrent acquisitions.
// If n <= 0, an error is returned.
func New(n int) (*Throttle, error) {
	if n <= 0 {
		return nil, errors.New("throttle: n must be greater than zero")
	}
	return &Throttle{sem: make(chan struct{}, n)}, nil
}

// Acquire blocks until a slot is available or ctx is cancelled.
func (t *Throttle) Acquire(ctx context.Context) error {
	select {
	case t.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release frees a previously acquired slot.
func (t *Throttle) Release() {
	select {
	case <-t.sem:
	default:
	}
}

// Cap returns the maximum concurrency allowed.
func (t *Throttle) Cap() int {
	return cap(t.sem)
}

// InFlight returns the number of currently acquired slots.
func (t *Throttle) InFlight() int {
	return len(t.sem)
}
