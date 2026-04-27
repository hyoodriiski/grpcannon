package relay

import "context"

// Invoker is the function signature used throughout grpcannon for a single RPC call.
type Invoker func(ctx context.Context) error

// Result carries the outcome of one RPC invocation.
type Result struct {
	Err error
}

// Tap wraps an Invoker, emitting a Result to the relay after every call.
// The relay receives the result regardless of whether the call succeeded.
func Tap(r *Relay[Result], next Invoker) Invoker {
	return func(ctx context.Context) error {
		err := next(ctx)
		r.Send(Result{Err: err})
		return err
	}
}
