// Package debounce provides a mechanism to coalesce rapid successive calls
// into a single invocation after a quiet period has elapsed.
package debounce

import (
	"sync"
	"time"
)

// Func is the signature of a debounced function.
type Func func()

// Debouncer delays execution of fn until after wait has elapsed since the
// last call to Trigger. Concurrent calls to Trigger reset the timer.
type Debouncer struct {
	wait  time.Duration
	fn    Func
	mu    sync.Mutex
	timer *time.Timer
}

// New returns a Debouncer that will invoke fn no sooner than wait after the
// final call to Trigger. A zero or negative wait fires on the next scheduler
// tick (equivalent to time.AfterFunc(0, fn)).
func New(wait time.Duration, fn Func) *Debouncer {
	if fn == nil {
		panic("debounce: fn must not be nil")
	}
	if wait < 0 {
		wait = 0
	}
	return &Debouncer{wait: wait, fn: fn}
}

// Trigger resets the debounce timer. If a timer is already running it is
// stopped and restarted.
func (d *Debouncer) Trigger() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.wait, d.fn)
}

// Cancel stops a pending invocation. It is safe to call Cancel even if no
// invocation is pending.
func (d *Debouncer) Cancel() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
}
