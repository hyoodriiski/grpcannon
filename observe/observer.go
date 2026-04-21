// Package observe provides a lightweight event observer for broadcasting
// named events to registered listeners during a load test run.
package observe

import "sync"

// Handler is a function that receives an event payload.
type Handler func(event string, payload any)

// Observer holds a set of named event listeners.
type Observer struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

// New returns a new Observer with an empty handler map.
func New() *Observer {
	return &Observer{
		handlers: make(map[string][]Handler),
	}
}

// On registers a Handler for the given event name.
// Nil handlers are silently ignored.
func (o *Observer) On(event string, h Handler) {
	if h == nil || event == "" {
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	o.handlers[event] = append(o.handlers[event], h)
}

// Emit broadcasts payload to every handler registered for event.
// Handlers are called synchronously in registration order.
func (o *Observer) Emit(event string, payload any) {
	o.mu.RLock()
	hs := make([]Handler, len(o.handlers[event]))
	copy(hs, o.handlers[event])
	o.mu.RUnlock()

	for _, h := range hs {
		h(event, payload)
	}
}

// Off removes all handlers registered for event.
func (o *Observer) Off(event string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	delete(o.handlers, event)
}

// Events returns the list of event names that have at least one handler.
func (o *Observer) Events() []string {
	o.mu.RLock()
	defer o.mu.RUnlock()
	names := make([]string, 0, len(o.handlers))
	for k := range o.handlers {
		names = append(names, k)
	}
	return names
}
