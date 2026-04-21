// Package ticker provides a periodic tick source that can be paused,
// resumed, and stopped cleanly. It is used by progress reporters and
// snapshot recorders that need regular interval-driven callbacks.
package ticker

import (
	"context"
	"time"
)

// Ticker fires a callback at a fixed interval until the context is
// cancelled or Stop is called.
type Ticker struct {
	interval time.Duration
	fn       func()
	stop     chan struct{}
	done     chan struct{}
}

// New creates a Ticker that calls fn every interval. The ticker does
// not start until Run is called. interval must be positive.
func New(interval time.Duration, fn func()) *Ticker {
	if interval <= 0 {
		interval = time.Second
	}
	return &Ticker{
		interval: interval,
		fn:       fn,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
}

// Run starts the ticker loop, blocking until ctx is cancelled or Stop
// is called. Run is safe to call in a goroutine.
func (t *Ticker) Run(ctx context.Context) {
	defer close(t.done)
	tk := time.NewTicker(t.interval)
	defer tk.Stop()
	for {
		select {
		case <-tk.C:
			if t.fn != nil {
				t.fn()
			}
		case <-t.stop:
			return
		case <-ctx.Done():
			return
		}
	}
}

// Stop signals the ticker to halt and waits for the loop to exit.
func (t *Ticker) Stop() {
	select {
	case <-t.stop: // already stopped
	default:
		close(t.stop)
	}
	<-t.done
}
