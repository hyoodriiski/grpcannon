package signal

import (
	"context"
)

// InvokerFn is a generic invocation function compatible with grpcannon workers.
type InvokerFn func(ctx context.Context) error

// Guard wraps an InvokerFn so that it returns immediately with ctx.Err() once
// the watcher's context has been cancelled, without invoking next.
func Guard(ctx context.Context, next InvokerFn) InvokerFn {
	return func(callCtx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		return next(callCtx)
	}
}
