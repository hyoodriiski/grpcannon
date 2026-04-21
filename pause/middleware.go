package pause

import (
	"context"
	"errors"
)

// ErrPauseCancelled is returned by Guard when the context is cancelled
// while waiting for the controller to resume.
var ErrPauseCancelled = errors.New("pause: context cancelled while waiting")

// Guard waits until the controller is not paused before calling fn.
// If the context is cancelled during the wait, ErrPauseCancelled is
// returned and fn is not called.
func Guard(ctx context.Context, c *Controller, fn func(context.Context) error) error {
	if !c.Wait(ctx) {
		return ErrPauseCancelled
	}
	return fn(ctx)
}

// GuardedInvoker wraps an invocation function so that every call
// first checks the pause controller before proceeding.
type GuardedInvoker struct {
	ctrl *Controller
	invoke func(context.Context) error
}

// NewGuardedInvoker creates a GuardedInvoker that gates calls through ctrl.
func NewGuardedInvoker(ctrl *Controller, invoke func(context.Context) error) *GuardedInvoker {
	return &GuardedInvoker{ctrl: ctrl, invoke: invoke}
}

// Call blocks if the controller is paused, then delegates to the
// underlying invoke function.
func (g *GuardedInvoker) Call(ctx context.Context) error {
	return Guard(ctx, g.ctrl, g.invoke)
}
