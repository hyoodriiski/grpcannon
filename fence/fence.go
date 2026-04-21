// Package fence provides a concurrency gate that limits the number of
// goroutines allowed to proceed past a checkpoint at any one time.
package fence

import (
	"context"
	"errors"
	"sync/atomic"
)

// ErrFenceClosed is returned when Acquire is called on a closed Fence.
var ErrFenceClosed = errors.New("fence: gate is closed")

// Fence is a concurrency gate with a configurable maximum pass-through count.
type Fence struct {
	max     int64
	current atomic.Int64
	closed  atomic.Bool
	gate    chan struct{}
}

// New creates a Fence that allows at most max concurrent passes.
// max must be greater than zero.
func New(max int) (*Fence, error) {
	if max <= 0 {
		return nil, errors.New("fence: max must be greater than zero")
	}
	f := &Fence{
		max:  int64(max),
		gate: make(chan struct{}, max),
	}
	for i := 0; i < max; i++ {
		f.gate <- struct{}{}
	}
	return f, nil
}

// Acquire blocks until a slot is available or ctx is cancelled.
// Returns ErrFenceClosed if the Fence has been closed.
func (f *Fence) Acquire(ctx context.Context) error {
	if f.closed.Load() {
		return ErrFenceClosed
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-f.gate:
		if f.closed.Load() {
			f.gate <- struct{}{}
			return ErrFenceClosed
		}
		f.current.Add(1)
		return nil
	}
}

// Release returns a slot to the Fence. Each Acquire must be paired with
// exactly one Release.
func (f *Fence) Release() {
	f.current.Add(-1)
	f.gate <- struct{}{}
}

// Active returns the number of goroutines currently past the gate.
func (f *Fence) Active() int64 {
	return f.current.Load()
}

// Close permanently closes the Fence. Subsequent Acquire calls return
// ErrFenceClosed.
func (f *Fence) Close() {
	f.closed.Store(true)
}
