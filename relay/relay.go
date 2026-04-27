// Package relay fans a single result stream out to multiple consumers.
package relay

import (
	"context"
	"sync"
)

// Handler is a function called with each relayed value.
type Handler[T any] func(T)

// Relay broadcasts values sent to it to all registered handlers.
type Relay[T any] struct {
	mu       sync.RWMutex
	handlers []Handler[T]
}

// New returns an empty Relay.
func New[T any]() *Relay[T] {
	return &Relay[T]{}
}

// Register adds a handler that will receive every future value.
// Nil handlers are silently ignored.
func (r *Relay[T]) Register(h Handler[T]) {
	if h == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers = append(r.handlers, h)
}

// Send delivers v to every registered handler synchronously.
func (r *Relay[T]) Send(v T) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, h := range r.handlers {
		h(v)
	}
}

// Pump reads from ch until it is closed or ctx is cancelled,
// forwarding each value to all registered handlers.
func (r *Relay[T]) Pump(ctx context.Context, ch <-chan T) {
	for {
		select {
		case <-ctx.Done():
			return
		case v, ok := <-ch:
			if !ok {
				return
			}
			r.Send(v)
		}
	}
}

// Len returns the number of registered handlers.
func (r *Relay[T]) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.handlers)
}
