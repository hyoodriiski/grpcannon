package output

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func baseSummary() Summary {
	return Summary{
		Total:       100,
		Errors:      2,
		Duration:    5 * time.Second,
		Throughput:  20.0,
		P50:         12 * time.Millisecond,
		P95:         45 * time.Millisecond,
		P99:         80 * time.Millisecond,
		MaxLatency:  100 * time.Millisecond,
		MeanLatency: 15 * time.Millisecond,
	}
}

func TestWriteText_ContainsHeaders(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteText(&buf, baseSummary()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"grpcannon", "Total requests", "Throughput", "P50", "P95", "P99", "Mean", "Max"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q", want)
		}
	}
}

func TestWriteText_Values(t *testing.T) {
	var buf bytes.Buffer
	s := baseSummary()
	_ = WriteText(&buf, s)
	out := buf.String()
	if !strings.Contains(out, "100") {
		t.Error("expected total 100 in output")
	}
	if !strings.Contains(out, "20.00") {
		t.Error("expected throughput 20.00 in output")
	}
}

func TestWriteText_ZeroValues(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteText(&buf, Summary{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output for zero summary")
	}
}
