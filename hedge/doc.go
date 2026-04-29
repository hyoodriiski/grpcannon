// Package hedge provides a hedged-request wrapper that transparently
// fires a duplicate RPC after a configurable delay, returning whichever
// response arrives first to reduce tail latency.
package hedge
