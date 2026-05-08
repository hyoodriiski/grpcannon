// Package mirror provides a fan-out Sink that forwards every recorded Result
// to a primary collector and an arbitrary number of secondary collectors.
// It is useful for capturing load test observations in multiple backends
// simultaneously without modifying the invoker pipeline.
package mirror
