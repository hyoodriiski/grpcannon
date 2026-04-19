// Package warmup implements a configurable warm-up phase for grpcannon.
// It fires RPC calls concurrently for a fixed duration before the timed
// measurement window begins, allowing connection pools and server JIT
// caches to stabilise.
package warmup
