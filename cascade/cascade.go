// Package cascade provides ordered fallback execution across multiple invokers.
// If the primary invoker fails, each subsequent invoker is tried in order.
package cascade

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Invoker is a function that performs a single gRPC-style call.
type Invoker func(ctx context.Context) (time.Duration, error)

// Result holds the outcome of a single stage in the cascade.
type Result struct {
	Stage    int
	Latency  time.Duration
	Err      error
}

// Cascade tries each invoker in order, returning the first success.
// If all invokers fail the last error is returned.
type Cascade struct {
	invokers []Invoker
}

// New creates a Cascade from the provided invokers.
// Nil invokers are silently ignored.
func New(invokers ...Invoker) *Cascade {
	filtered := make([]Invoker, 0, len(invokers))
	for _, inv := range invokers {
		if inv != nil {
			filtered = append(filtered, inv)
		}
	}
	return &Cascade{invokers: filtered}
}

// Run executes the cascade, trying each invoker until one succeeds.
// It returns the Result of the first successful stage, or the last
// failure if all stages are exhausted.
func (c *Cascade) Run(ctx context.Context) (Result, error) {
	if len(c.invokers) == 0 {
		return Result{}, errors.New("cascade: no invokers registered")
	}

	var last error
	for i, inv := range c.invokers {
		if ctx.Err() != nil {
			return Result{Stage: i}, ctx.Err()
		}
		latency, err := inv(ctx)
		if err == nil {
			return Result{Stage: i, Latency: latency}, nil
		}
		last = fmt.Errorf("cascade: stage %d: %w", i, err)
	}
	return Result{Stage: len(c.invokers) - 1}, last
}

// Len returns the number of registered invokers.
func (c *Cascade) Len() int { return len(c.invokers) }
