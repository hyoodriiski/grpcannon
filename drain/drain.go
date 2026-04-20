// Package drain provides a graceful shutdown helper that waits for
// in-flight requests to complete before allowing the process to exit.
package drain

import (
	"context"
	"sync"
	"time"
)

// Drainer tracks active requests and blocks until they finish or the
// supplied deadline expires.
type Drainer struct {
	mu      sync.Mutex
	wg      sync.WaitGroup
	closed  bool
	timeout time.Duration
}

// New returns a Drainer that will wait at most timeout for in-flight
// work to finish. A zero timeout means wait forever.
func New(timeout time.Duration) *Drainer {
	return &Drainer{timeout: timeout}
}

// Acquire signals that a new unit of work has started.
// It returns false if the Drainer has already been closed, meaning no
// new work should begin.
func (d *Drainer) Acquire() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return false
	}
	d.wg.Add(1)
	return true
}

// Release signals that a unit of work has finished.
func (d *Drainer) Release() {
	d.wg.Done()
}

// Drain closes the Drainer to new acquisitions and blocks until all
// in-flight work completes or ctx is cancelled.
// Returns context.DeadlineExceeded when the internal timeout fires, or
// ctx.Err() when the parent context is cancelled first.
func (d *Drainer) Drain(ctx context.Context) error {
	d.mu.Lock()
	d.closed = true
	d.mu.Unlock()

	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()

	var timer <-chan time.Time
	if d.timeout > 0 {
		t := time.NewTimer(d.timeout)
		defer t.Stop()
		timer = t.C
	}

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-timer:
		return context.DeadlineExceeded
	}
}
