// Package policy provides a composable execution policy that combines
// retry, exponential backoff, and circuit-breaker protection into a
// single, reusable abstraction for gRPC call sites.
//
// # Overview
//
// A Policy wraps an outbound gRPC call and transparently applies:
//
//   - Retry logic with configurable attempt limits and retryable status codes.
//   - Exponential backoff with optional jitter between retry attempts.
//   - Circuit-breaker protection that opens after a configurable failure
//     threshold and half-opens after a cooldown period.
//
// # Usage
//
//	p := policy.New(
//		policy.WithMaxAttempts(3),
//		policy.WithBackoff(50*time.Millisecond, 2.0),
//		policy.WithCircuitBreaker(5, 10*time.Second),
//	)
//
//	err := p.Execute(ctx, func(ctx context.Context) error {
//		_, err := client.SomeRPC(ctx, req)
//		return err
//	})
//
// Policies are safe for concurrent use by multiple goroutines.
package policy
