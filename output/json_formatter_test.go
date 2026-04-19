package output

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestWriteJSONSummary_ValidJSON(t *testing.T) {
	var buf bytes.Buffer
	s := Summary{
		Total:       50,
		Errors:      1,
		Duration:    2 * time.Second,
		Throughput:  25.0,
		P50:         10 * time.Millisecond,
		P95:         30 * time.Millisecond,
		P99:         60 * time.Millisecond,
		MaxLatency:  70 * time.Millisecond,
		MeanLatency: 12 * time.Millisecond,
	}
	if err := WriteJSONSummary(&buf, s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

func TestWriteJSONSummary_Fields(t *testing.T) {
	var buf bytes.Buffer
	s := baseSummary()
	_ = WriteJSONSummary(&buf, s)
	var out map[string]interface{}
	_ = json.Unmarshal(buf.Bytes(), &out)
	for _, key := range []string{"total_requests", "errors", "throughput_rps", "p50_ms", "p95_ms", "p99_ms", "mean_ms", "max_ms", "duration_ms"} {
		if _, ok := out[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}
}

func TestWriteJSONSummary_ThroughputValue(t *testing.T) {
	var buf bytes.Buffer
	s := baseSummary()
	_ = WriteJSONSummary(&buf, s)
	var out map[string]interface{}
	_ = json.Unmarshal(buf.Bytes(), &out)
	if got := out["throughput_rps"].(float64); got != 20.0 {
		t.Errorf("expected throughput 20.0, got %v", got)
	}
}
