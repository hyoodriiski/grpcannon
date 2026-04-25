// Package loadgen provides a high-level orchestration layer for gRPC load
// generation runs. It composes rate limiting, circuit breaking, concurrency
// control, metrics collection, and live progress reporting into a single
// blocking Run call, returning an aggregate Result when the run completes.
package loadgen
