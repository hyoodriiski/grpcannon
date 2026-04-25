// Package loadgen orchestrates a full load generation run, wiring together
// workers, rate limiting, circuit breaking, and progress reporting.
package loadgen

import (
	"context"
	"time"

	"github.com/yourorg/grpcannon/circuit"
	"github.com/yourorg/grpcannon/metrics"
	"github.com/yourorg/grpcannon/progress"
	"github.com/yourorg/grpcannon/ratelimit"
	"github.com/yourorg/grpcannon/worker"
)

// Config holds the parameters for a single load generation run.
type Config struct {
	Concurrency int
	TotalCalls  int
	RatePerSec  int
	Timeout     time.Duration
	Invoker     worker.InvokerFn
}

// Result captures the aggregate outcome of a run.
type Result struct {
	Summary  *metrics.Summary
	Duration time.Duration
	Errors   int
}

// Run executes the load generation described by cfg and returns a Result.
// It wires together a rate limiter, a circuit breaker, a worker pool, and a
// progress reporter, then blocks until all calls complete or ctx is cancelled.
func Run(ctx context.Context, cfg Config) (*Result, error) {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 1
	}

	rl := ratelimit.New(cfg.RatePerSec)
	cb := circuit.New(circuit.Options{})
	col := metrics.NewCollector()
	prog := progress.New(progress.Options{Interval: time.Second})

	prog.Start(ctx)
	defer prog.Stop()

	guardedInvoker := func(ctx context.Context) error {
		if err := rl.Wait(ctx); err != nil {
			return err
		}
		return circuit.RunWithBreaker(cb, func() error {
			start := time.Now()
			err := cfg.Invoker(ctx)
			col.Record(time.Since(start), err)
			prog.Record(err)
			return err
		})
	}

	pool := worker.NewPool(cfg.Concurrency, guardedInvoker)

	start := time.Now()
	pool.Run(ctx, cfg.TotalCalls)
	elapsed := time.Since(start)

	sum := col.Summarise()
	return &Result{
		Summary:  sum,
		Duration: elapsed,
		Errors:   col.ErrorCount(),
	}, nil
}
