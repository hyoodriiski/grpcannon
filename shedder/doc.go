// Package shedder provides adaptive load shedding driven by a rolling
// error-rate window. It is intended to protect downstream services from
// cascading failures by proactively rejecting requests when the recent
// error rate rises above a configurable threshold.
package shedder
