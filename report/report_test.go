package report

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func makeLatencies(vals ...int) []time.Duration {
	out := make([]time.Duration, len(vals))
	for i, v := range vals {
		out[i] = time.Duration(v) * time.Millisecond
	}
	return out
}

func TestNew_Throughput(t *testing.T) {
	lat := makeLatencies(10, 20, 30, 40, 50)
	s := New(lat, 1, 5*time.Second)
	if s.Total != 5 {
		t.Errorf("expected Total=5, got %d", s.Total)
	}
	if s.Failures != 1 {
		t.Errorf("expected Failures=1, got %d", s.Failures)
	}
	if s.Successes != 4 {
		t.Errorf("expected Successes=4, got %d", s.Successes)
	}
	if s.Throughput != 1.0 {
		t.Errorf("expected Throughput=1.0, got %.2f", s.Throughput)
	}
}

func TestNew_ZeroDuration(t *testing.T) {
	s := New(makeLatencies(5, 10), 0, 0)
	if s.Throughput != 0 {
		t.Errorf("expected Throughput=0 for zero duration")
	}
}

func TestPrint_ContainsFields(t *testing.T) {
	s := Summary{
		Total: 100, Successes: 95, Failures: 5,
		Duration:   2 * time.Second,
		Throughput: 50.0,
		P50: 10 * time.Millisecond,
		P95: 45 * time.Millisecond,
		P99: 99 * time.Millisecond,
	}
	var buf bytes.Buffer
	Print(s, &buf)
	out := buf.String()
	for _, want := range []string{"100", "95", "5", "50.00", "10ms", "45ms", "99ms"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}
