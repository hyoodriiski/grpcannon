// Package cascade implements ordered fallback execution for gRPC invokers.
// Primary use-case: retry across multiple backends before surfacing an error.
package cascade
