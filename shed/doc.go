// Package shed provides probabilistic load-shedding for gRPC calls.
//
// A Shed tracks the number of in-flight requests and rejects new ones once a
// configurable maximum is reached, returning ErrShed to the caller. Use
// Guard to wrap an Invoker with automatic acquire/release semantics.
package shed
