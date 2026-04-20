// Package circuit provides a circuit breaker implementation for protecting
// gRPC load test workers from cascading failures. When consecutive errors
// exceed a configured threshold the breaker opens and rejects calls until
// a cooldown period has elapsed.
package circuit
