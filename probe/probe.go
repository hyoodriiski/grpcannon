// Package probe provides periodic health-checking for a target invoker.
// A Probe calls a user-supplied check function on a configurable interval
// and exposes the most recent healthy/unhealthy status.
package probe

import (
	"context"
	"sync"
	"time"
)

// Status represents the result of a single probe check.
type Status struct {
	Healthy bool
	Err     error
	At      time.Time
}

// CheckFn is the function called on each probe tick.
type CheckFn func(ctx context.Context) error

// Probe runs a periodic health check and tracks the current status.
type Probe struct {
	mu       sync.RWMutex
	status   Status
	check    CheckFn
	interval time.Duration
	stop     chan struct{}
	done     chan struct{}
}

// New creates a new Probe that calls check every interval.
// A zero or negative interval defaults to 5 seconds.
func New(check CheckFn, interval time.Duration) *Probe {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	return &Probe{
		check:    check,
		interval: interval,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
}

// Start begins the probe loop in a background goroutine.
// It runs an immediate check, then repeats on each tick.
func (p *Probe) Start(ctx context.Context) {
	go func() {
		defer close(p.done)
		p.run(ctx)
		t := time.NewTicker(p.interval)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				p.run(ctx)
			case <-p.stop:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop halts the probe loop and waits for it to exit.
func (p *Probe) Stop() {
	close(p.stop)
	<-p.done
}

// Status returns the most recent probe status.
func (p *Probe) Status() Status {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status
}

// Healthy reports whether the last check succeeded.
func (p *Probe) Healthy() bool {
	return p.Status().Healthy
}

func (p *Probe) run(ctx context.Context) {
	err := p.check(ctx)
	p.mu.Lock()
	p.status = Status{Healthy: err == nil, Err: err, At: time.Now()}
	p.mu.Unlock()
}
