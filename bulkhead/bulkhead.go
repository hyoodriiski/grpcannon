// Package bulkhead provides a concurrency isolation primitive that limits
// the number of simultaneous in-flight calls to a given resource, preventing
// one slow dependency from exhausting the whole worker pool.
package bulkhead

import (
	"context"
	"errors"
	"sync"
)

// ErrFull is returned when the bulkhead has no remaining capacity.
var ErrFull = errors.New("bulkhead: at capacity")

// Bulkhead limits concurrent access to a resource.
type Bulkhead struct {
	mu      sync.Mutex
	max     int
	active  int
	closed  bool
}

// New creates a Bulkhead that allows at most max concurrent acquisitions.
// If max is zero or negative, no limit is enforced.
func New(max int) *Bulkhead {
	if max < 0 {
		max = 0
	}
	return &Bulkhead{max: max}
}

// Acquire attempts to enter the bulkhead. It returns ErrFull when the
// concurrency limit is reached, or an error if ctx is already done.
func (b *Bulkhead) Acquire(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return errors.New("bulkhead: closed")
	}
	if b.max > 0 && b.active >= b.max {
		return ErrFull
	}
	b.active++
	return nil
}

// Release decrements the active counter. It must be called once for every
// successful Acquire.
func (b *Bulkhead) Release() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.active > 0 {
		b.active--
	}
}

// Active returns the current number of in-flight acquisitions.
func (b *Bulkhead) Active() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.active
}

// Close prevents future acquisitions.
func (b *Bulkhead) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
}
