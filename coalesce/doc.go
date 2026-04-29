// Package coalesce provides a call-coalescing primitive that merges concurrent
// requests for the same key into a single in-flight invocation, reducing
// redundant work under high concurrency.
package coalesce
