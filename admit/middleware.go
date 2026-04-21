package admit

import (
	"context"
	"time"
)

// InvokerFn is a function that performs a single gRPC invocation.
type InvokerFn func(ctx context.Context, method string) (time.Duration, error)

// Guard wraps an InvokerFn with admission control. If the Controller rejects
// the request, Guard returns ErrRejected without calling the underlying
// invoker, preserving a zero latency value.
func Guard(c *Controller, next InvokerFn) InvokerFn {
	return func(ctx context.Context, method string) (time.Duration, error) {
		release, err := c.Admit(ctx)
		if err != nil {
			return 0, err
		}
		defer release()
		return next(ctx, method)
	}
}
