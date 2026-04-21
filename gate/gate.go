// Package gate provides a simple on/off gate that can block callers
// until it is opened, useful for coordinating staged load test starts.
package gate

import (
	"context"
	"sync"
)

// Gate blocks callers on Wait until it is opened.
type Gate struct {
	mu     sync.RWMutex
	open   bool
	notify chan struct{}
}

// New returns a closed Gate.
func New() *Gate {
	return &Gate{
		notify: make(chan struct{}),
	}
}

// Open unblocks all current and future callers of Wait.
func (g *Gate) Open() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.open {
		g.open = true
		close(g.notify)
	}
}

// Close resets the gate so that future callers of Wait will block again.
// Callers already unblocked by a previous Open are not affected.
func (g *Gate) Close() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.open {
		g.open = false
		g.notify = make(chan struct{})
	}
}

// IsOpen reports whether the gate is currently open.
func (g *Gate) IsOpen() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.open
}

// Wait blocks until the gate is open or ctx is cancelled.
// Returns ctx.Err() if the context is done before the gate opens.
func (g *Gate) Wait(ctx context.Context) error {
	g.mu.RLock()
	if g.open {
		g.mu.RUnlock()
		return nil
	}
	ch := g.notify
	g.mu.RUnlock()

	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
