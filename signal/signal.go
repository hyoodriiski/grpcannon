// Package signal provides a graceful shutdown coordinator that listens for
// OS signals and notifies registered handlers in order.
package signal

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Handler is a function called when a shutdown signal is received.
type Handler func(ctx context.Context)

// Watcher listens for OS signals and invokes registered handlers.
type Watcher struct {
	mu       sync.Mutex
	handlers []Handler
	signals  []os.Signal
}

// New creates a Watcher that listens for the provided signals.
// If no signals are provided it defaults to SIGINT and SIGTERM.
func New(sigs ...os.Signal) *Watcher {
	if len(sigs) == 0 {
		sigs = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}
	return &Watcher{signals: sigs}
}

// Register adds a handler to be called on shutdown. Nil handlers are ignored.
func (w *Watcher) Register(h Handler) {
	if h == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.handlers = append(w.handlers, h)
}

// Wait blocks until a signal is received, then calls all registered handlers
// with the provided context and returns.
func (w *Watcher) Wait(ctx context.Context) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, w.signals...)
	defer signal.Stop(ch)

	select {
	case <-ctx.Done():
	case <-ch:
	}

	w.mu.Lock()
	handlers := make([]Handler, len(w.handlers))
	copy(handlers, w.handlers)
	w.mu.Unlock()

	for _, h := range handlers {
		h(ctx)
	}
}
