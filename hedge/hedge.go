// Package hedge implements a hedged request strategy that issues a
// duplicate request after a configurable delay if the original has not
// yet returned, taking whichever response arrives first.
package hedge

import (
	"context"
	"errors"
	"time"
)

// ErrNoInvoker is returned when Run is called with a nil invoker.
var ErrNoInvoker = errors.New("hedge: invoker must not be nil")

// InvokerFunc is the function signature for a single RPC attempt.
type InvokerFunc func(ctx context.Context) error

// Hedge holds the configuration for hedged requests.
type Hedge struct {
	delay time.Duration
}

// New creates a Hedge that issues a second attempt after delay.
// A zero or negative delay disables hedging (only one attempt is made).
func New(delay time.Duration) *Hedge {
	return &Hedge{delay: delay}
}

// Run executes fn and, if no result has arrived within the hedge delay,
// fires a second concurrent attempt. The first successful result wins;
// if both fail the last error is returned.
func (h *Hedge) Run(ctx context.Context, fn InvokerFunc) error {
	if fn == nil {
		return ErrNoInvoker
	}

	type result struct{ err error }
	results := make(chan result, 2)

	launch := func(c context.Context) {
		results <- result{err: fn(c)}
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go launch(ctx)

	if h.delay <= 0 {
		r := <-results
		return r.err
	}

	timer := time.NewTimer(h.delay)
	defer timer.Stop()

	select {
	case r := <-results:
		if r.err == nil {
			return nil
		}
		// first attempt failed before hedge fired — wait for hedge or give up
		select {
		case r2 := <-results:
			return r2.err
		case <-timer.C:
			go launch(ctx)
			r2 := <-results
			return r2.err
		}
	case <-timer.C:
		go launch(ctx)
	}

	// wait for whichever finishes first
	var lastErr error
	for i := 0; i < 2; i++ {
		r := <-results
		if r.err == nil {
			return nil
		}
		lastErr = r.err
	}
	return lastErr
}
