package valve

import (
	"context"
	"time"
)

// InvokerFn is a function that performs a single gRPC invocation.
type InvokerFn func(ctx context.Context, method string, payload []byte) (time.Duration, error)

// Guard wraps next so that every call waits for the valve to be open before
// proceeding. If the valve is closed and ctx expires, the error is returned
// immediately without calling next.
func Guard(v *Valve, next InvokerFn) InvokerFn {
	return func(ctx context.Context, method string, payload []byte) (time.Duration, error) {
		if err := v.Wait(ctx); err != nil {
			return 0, err
		}
		return next(ctx, method, payload)
	}
}
