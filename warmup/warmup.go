// Package warmup provides a pre-load warm-up phase before the main run.
package warmup

import (
	"context"
	"time"
)

// Config holds warm-up configuration.
type Config struct {
	Duration    time.Duration
	Concurrency int
}

// Invoker is the function signature used to invoke a single RPC.
type Invoker func(ctx context.Context) error

// Run executes the warm-up phase by invoking fn concurrently for the
// specified duration. Errors during warm-up are silently discarded.
func Run(ctx context.Context, cfg Config, fn Invoker) {
	if cfg.Duration <= 0 || cfg.Concurrency <= 0 {
		return
	}

	deadline, cancel := context.WithTimeout(ctx, cfg.Duration)
	defer cancel()

	work := make(chan struct{})
	go func() {
		defer close(work)
		for {
			select {
			case <-deadline.Done():
				return
			case work <- struct{}{}:
			}
		}
	}()

	done := make(chan struct{})
	for i := 0; i < cfg.Concurrency; i++ {
		go func() {
			for range work {
				_ = fn(deadline)
			}
			done <- struct{}{}
		}()
	}

	for i := 0; i < cfg.Concurrency; i++ {
		<-done
	}
}
