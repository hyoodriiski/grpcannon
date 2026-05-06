package ceiling

import "context"

// InvokerFn is a function that performs a single gRPC invocation.
type InvokerFn func(ctx context.Context) error

// Guard wraps next with the ceiling limiter. If the ceiling is reached the
// call is rejected immediately with ErrAbove without invoking next.
func Guard(c *Ceiling, next InvokerFn) InvokerFn {
	return func(ctx context.Context) error {
		if err := c.Acquire(ctx); err != nil {
			return err
		}
		defer c.Release()
		return next(ctx)
	}
}
