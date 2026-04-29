// Package coalesce deduplicates in-flight calls for identical keys,
// ensuring only one real invocation runs at a time per key.
package coalesce

import (
	"context"
	"sync"
)

// Result holds the outcome of a coalesced call.
type Result struct {
	Value interface{}
	Err   error
}

// call represents a single in-flight or completed call.
type call struct {
	wg  sync.WaitGroup
	res Result
}

// Group manages deduplication of concurrent calls by key.
type Group struct {
	mu    sync.Mutex
	calls map[string]*call
}

// New returns a new Group ready for use.
func New() *Group {
	return &Group{
		calls: make(map[string]*call),
	}
}

// Do executes fn for the given key, suppressing duplicate concurrent calls.
// All callers sharing the same key while fn is in-flight receive the same
// result. fn receives the provided context.
func (g *Group) Do(ctx context.Context, key string, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if c, ok := g.calls[key]; ok {
		g.mu.Unlock()
		// Wait for the in-flight call to finish, respecting context cancellation.
		done := make(chan struct{})
		go func() {
			c.wg.Wait()
			close(done)
		}()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-done:
			return c.res.Value, c.res.Err
		}
	}

	c := &call{}
	c.wg.Add(1)
	g.calls[key] = c
	g.mu.Unlock()

	c.res.Value, c.res.Err = fn(ctx)
	c.wg.Done()

	g.mu.Lock()
	delete(g.calls, key)
	g.mu.Unlock()

	return c.res.Value, c.res.Err
}

// InFlight returns the number of keys currently being executed.
func (g *Group) InFlight() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return len(g.calls)
}
