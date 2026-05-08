// Package scatter fans a single invocation out to N concurrent workers
// and collects all results, returning once every worker has finished or
// the context is cancelled.
package scatter

import (
	"context"
	"sync"
	"time"
)

// Result holds the outcome of a single worker invocation.
type Result struct {
	Index   int
	Latency time.Duration
	Err     error
}

// Invoker is the function signature each worker calls.
type Invoker func(ctx context.Context) error

// Scatter fans out fn across n concurrent goroutines and returns all results.
// If ctx is cancelled before all workers finish, remaining workers are
// interrupted and their results include ctx.Err().
func Scatter(ctx context.Context, n int, fn Invoker) []Result {
	if n <= 0 || fn == nil {
		return nil
	}

	results := make([]Result, n)
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			start := time.Now()
			err := fn(ctx)
			results[i] = Result{
				Index:   i,
				Latency: time.Since(start),
				Err:     err,
			}
		}()
	}

	wg.Wait()
	return results
}

// Errors returns only the non-nil errors from a result slice.
func Errors(results []Result) []error {
	var errs []error
	for _, r := range results {
		if r.Err != nil {
			errs = append(errs, r.Err)
		}
	}
	return errs
}

// SuccessCount returns the number of results with a nil error.
func SuccessCount(results []Result) int {
	count := 0
	for _, r := range results {
		if r.Err == nil {
			count++
		}
	}
	return count
}
