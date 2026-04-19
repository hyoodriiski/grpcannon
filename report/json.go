package report

import (
	"encoding/json"
	"os"
	"time"
)

type jsonSummary struct {
	Total      int     `json:"total"`
	Successes  int     `json:"successes"`
	Failures   int     `json:"failures"`
	DurationMs int64   `json:"duration_ms"`
	Throughput float64 `json:"throughput_rps"`
	P50Ms      int64   `json:"p50_ms"`
	P95Ms      int64   `json:"p95_ms"`
	P99Ms      int64   `json:"p99_ms"`
}

// WriteJSON serialises the Summary as JSON to the given file path.
func WriteJSON(s Summary, path string) error {
	j := jsonSummary{
		Total:      s.Total,
		Successes:  s.Successes,
		Failures:   s.Failures,
		DurationMs: s.Duration.Milliseconds(),
		Throughput: s.Throughput,
		P50Ms:      s.P50.Milliseconds(),
		P95Ms:      s.P95.Milliseconds(),
		P99Ms:      s.P99.Milliseconds(),
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(j)
}

// MarshalJSON returns the Summary as a JSON byte slice.
func MarshalJSON(s Summary) ([]byte, error) {
	j := jsonSummary{
		Total:      s.Total,
		Successes:  s.Successes,
		Failures:   s.Failures,
		DurationMs: s.Duration.Milliseconds(),
		Throughput: s.Throughput,
		P50Ms:      s.P50 / time.Millisecond,
		P95Ms:      s.P95 / time.Millisecond,
		P99Ms:      s.P99 / time.Millisecond,
	}
	return json.MarshalIndent(j, "", "  ")
}
