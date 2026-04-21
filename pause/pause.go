// Package pause provides a mechanism to temporarily pause and resume
// load generation during a gRPC cannon run.
package pause

import (
	"context"
	"sync"
)

// Controller allows callers to pause and resume work.
type Controller struct {
	mu     sync.Mutex
	cond   *sync.Cond
	paused bool
	closed bool
}

// New returns a ready-to-use Controller.
func New() *Controller {
	c := &Controller{}
	c.cond = sync.NewCond(&c.mu)
	return c
}

// Pause signals that workers should stop and wait.
func (c *Controller) Pause() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.paused = true
}

// Resume signals that workers may continue.
func (c *Controller) Resume() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.paused = false
	c.cond.Broadcast()
}

// Close unblocks all waiters permanently.
func (c *Controller) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	c.paused = false
	c.cond.Broadcast()
}

// Wait blocks until the controller is not paused, the context is done,
// or Close has been called. Returns false if the context was cancelled.
func (c *Controller) Wait(ctx context.Context) bool {
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		c.mu.Lock()
		c.cond.Broadcast()
		c.mu.Unlock()
		close(done)
	}()

	c.mu.Lock()
	for c.paused && !c.closed {
		select {
		case <-ctx.Done():
			c.mu.Unlock()
			<-done
			return false
		default:
		}
		c.cond.Wait()
	}
	c.mu.Unlock()
	return ctx.Err() == nil
}

// Paused reports whether the controller is currently paused.
func (c *Controller) Paused() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.paused
}
