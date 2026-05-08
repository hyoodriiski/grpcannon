package mirror

import (
	"context"
	"time"
)

// InvokerFn is the signature used throughout grpcannon for a single RPC call.
type InvokerFn func(ctx context.Context) error

// Tap wraps next so that each invocation's latency and error are forwarded to
// the provided Mirror.  The original error (if any) is always returned to the
// caller unchanged.
//
//	guarded := mirror.Tap(m, invoker.Call)
func Tap(m *Mirror, next InvokerFn) InvokerFn {
	if m == nil || next == nil {
		return next
	}
	return func(ctx context.Context) error {
		start := time.Now()
		err := next(ctx)
		m.Record(Result{
			Latency: time.Since(start),
			Err:     err,
		})
		return err
	}
}
