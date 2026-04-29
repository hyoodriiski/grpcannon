// Package policy composes retry, backoff, and circuit-breaker behaviour
// into a single reusable execution policy.
package policy

import (
	"context"
	"errors"
	"time"

	"github.com/nickcorin/grpcannon/backoff"
	"github.com/nickcorin/grpcannon/circuit"
)

// ErrCircuitOpen is returned when the circuit breaker is open.
var ErrCircuitOpen = errors.New("policy: circuit open")

// Options configures the execution policy.
type Options struct {
	MaxAttempts int
	Breaker     *circuit.Breaker
	Backoff     backoff.Backoff
}

// Policy wraps retry and circuit-breaker logic.
type Policy struct {
	opts Options
}

// New creates a Policy with the provided options.
// If MaxAttempts is zero it defaults to 1.
func New(opts Options) *Policy {
	if opts.MaxAttempts <= 0 {
		opts.MaxAttempts = 1
	}
	if opts.Backoff == (backoff.Backoff{}) {
		opts.Backoff = backoff.Default()
	}
	return &Policy{opts: opts}
}

// Execute runs fn according to the policy, retrying on error with backoff
// and honouring the circuit breaker when one is configured.
func (p *Policy) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
	var last error
	for attempt := 0; attempt < p.opts.MaxAttempts; attempt++ {
		if p.opts.Breaker != nil {
			if err := p.opts.Breaker.Allow(); err != nil {
				return ErrCircuitOpen
			}
		}

		last = fn(ctx)

		if p.opts.Breaker != nil {
			p.opts.Breaker.Record(last)
		}

		if last == nil {
			return nil
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		if attempt < p.opts.MaxAttempts-1 {
			delay := p.opts.Backoff.Delay(attempt)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return last
}
