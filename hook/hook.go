// Package hook provides before/after lifecycle hooks for gRPC invocation runs.
package hook

import "context"

// Phase indicates when a hook is executed.
type Phase int

const (
	// BeforeRun is called once before the load test begins.
	BeforeRun Phase = iota
	// AfterRun is called once after the load test completes.
	AfterRun
)

// Fn is a function that can be registered as a hook.
type Fn func(ctx context.Context, phase Phase) error

// Runner holds registered hooks and executes them in registration order.
type Runner struct {
	before []Fn
	after  []Fn
}

// New returns an empty Runner.
func New() *Runner {
	return &Runner{}
}

// Register adds fn to the set of hooks for the given phase.
func (r *Runner) Register(phase Phase, fn Fn) {
	if fn == nil {
		return
	}
	switch phase {
	case BeforeRun:
		r.before = append(r.before, fn)
	case AfterRun:
		r.after = append(r.after, fn)
	}
}

// RunBefore executes all BeforeRun hooks in order.
// It stops and returns the first error encountered.
func (r *Runner) RunBefore(ctx context.Context) error {
	return r.run(ctx, BeforeRun, r.before)
}

// RunAfter executes all AfterRun hooks in order.
// It stops and returns the first error encountered.
func (r *Runner) RunAfter(ctx context.Context) error {
	return r.run(ctx, AfterRun, r.after)
}

func (r *Runner) run(ctx context.Context, phase Phase, fns []Fn) error {
	for _, fn := range fns {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := fn(ctx, phase); err != nil {
			return err
		}
	}
	return nil
}
