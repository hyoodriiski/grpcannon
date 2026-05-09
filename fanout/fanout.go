// Package fanout provides a mechanism for broadcasting a single invocation
// to multiple targets concurrently, collecting all results.
package fanout

import (
	"context"
	"sync"
	"time"
)

// Result holds the outcome of a single fanout target invocation.
type Result struct {
	// Index is the position of the target in the original slice.
	Index int
	// Latency is the time taken for the invocation.
	Latency time.Duration
	// Err holds any error returned by the target.
	Err error
}

// InvokerFn is a function that performs a single call against a target.
type InvokerFn func(ctx context.Context) error

// Fan broadcasts a call to all provided invokers concurrently and waits for
// all of them to complete or the context to be cancelled. Results are returned
// in the order they complete, not necessarily the order of the invokers slice.
func Fan(ctx context.Context, invokers []InvokerFn) []Result {
	if len(invokers) == 0 {
		return nil
	}

	results := make([]Result, len(invokers))
	var wg sync.WaitGroup

	for i, fn := range invokers {
		if fn == nil {
			results[i] = Result{Index: i, Err: nil}
			continue
		}

		wg.Add(1)
		go func(idx int, invoke InvokerFn) {
			defer wg.Done()
			start := time.Now()
			err := invoke(ctx)
			results[idx] = Result{
				Index:   idx,
				Latency: time.Since(start),
				Err:     err,
			}
		}(i, fn)
	}

	wg.Wait()
	return results
}

// Errors returns only the results that contain a non-nil error.
func Errors(results []Result) []Result {
	var out []Result
	for _, r := range results {
		if r.Err != nil {
			out = append(out, r)
		}
	}
	return out
}

// SuccessCount returns the number of results with no error.
func SuccessCount(results []Result) int {
	count := 0
	for _, r := range results {
		if r.Err == nil {
			count++
		}
	}
	return count
}
