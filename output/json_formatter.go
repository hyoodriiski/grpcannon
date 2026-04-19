package output

import (
	"encoding/json"
	"io"
)

type jsonSummary struct {
	Total       int     `json:"total_requests"`
	Errors      int     `json:"errors"`
	DurationMs  float64 `json:"duration_ms"`
	Throughput  float64 `json:"throughput_rps"`
	MeanMs      float64 `json:"mean_ms"`
	P50Ms       float64 `json:"p50_ms"`
	P95Ms       float64 `json:"p95_ms"`
	P99Ms       float64 `json:"p99_ms"`
	MaxMs       float64 `json:"max_ms"`
}

// WriteJSONSummary encodes s as JSON into w.
func WriteJSONSummary(w io.Writer, s Summary) error {
	js := jsonSummary{
		Total:      s.Total,
		Errors:     s.Errors,
		DurationMs: float64(s.Duration.Milliseconds()),
		Throughput: s.Throughput,
		MeanMs:     float64(s.MeanLatency.Microseconds()) / 1000.0,
		P50Ms:      float64(s.P50.Microseconds()) / 1000.0,
		P95Ms:      float64(s.P95.Microseconds()) / 1000.0,
		P99Ms:      float64(s.P99.Microseconds()) / 1000.0,
		MaxMs:      float64(s.MaxLatency.Microseconds()) / 1000.0,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(js)
}
