package pacing

import (
	"context"
	"time"
)

// InvokerFn is the signature for a gRPC invocation function.
type InvokerFn func(ctx context.Context) (time.Duration, error)

// Guard wraps next with pacing: it calls p.Wait before each invocation.
// If Wait returns an error (e.g. context cancelled) it is returned immediately.
func Guard(p *Pacer, next InvokerFn) InvokerFn {
	return func(ctx context.Context) (time.Duration, error) {
		if err := p.Wait(ctx); err != nil {
			return 0, err
		}
		return next(ctx)
	}
}
