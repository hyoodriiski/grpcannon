// Package drift provides a latency-drift detector that compares a rolling
// average of observed call durations against a configured baseline, returning
// an error when the ratio exceeds the allowed threshold.
package drift
