// Package valve provides a flow-control primitive that can be opened or
// closed to allow or block invocations. When closed all calls block until
// the valve is reopened or the caller's context expires.
package valve

import (
	"context"
	"sync"
)

// Valve gates invocations behind an open/closed state.
type Valve struct {
	mu     sync.RWMutex
	open   bool
	notify chan struct{}
}

// New returns a Valve. If open is true the valve starts in the open state.
func New(open bool) *Valve {
	v := &Valve{notify: make(chan struct{})}
	v.open = open
	return v
}

// Open allows blocked and future callers to proceed.
func (v *Valve) Open() {
	v.mu.Lock()
	defer v.mu.Unlock()
	if !v.open {
		v.open = true
		close(v.notify)
		v.notify = make(chan struct{})
	}
}

// Close causes future calls to Wait to block until Open is called.
func (v *Valve) Close() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.open = false
}

// IsOpen reports the current state.
func (v *Valve) IsOpen() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.open
}

// Wait blocks until the valve is open or ctx is cancelled.
func (v *Valve) Wait(ctx context.Context) error {
	v.mu.RLock()
	if v.open {
		v.mu.RUnlock()
		return nil
	}
	ch := v.notify
	v.mu.RUnlock()

	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
