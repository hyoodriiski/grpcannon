package runner

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/grpcannon/config"
)

func defaultCfg() *config.Config {
	return &config.Config{
		Address:        "localhost:50051",
		Concurrency:    4,
		TotalRequests:  20,
		Timeout:        2 * time.Second,
	}
}

func TestRun_AllSuccess(t *testing.T) {
	r := New(defaultCfg())
	results := r.Run(context.Background(), func(_ context.Context) error { return nil })
	if len(results) != 20 {
		t.Fatalf("expected 20 results, got %d", len(results))
	}
	for _, res := range results {
		if res.Err != nil {
			t.Errorf("unexpected error: %v", res.Err)
		}
	}
}

func TestRun_SomeErrors(t *testing.T) {
	r := New(defaultCfg())
	i := 0
	results := r.Run(context.Background(), func(_ context.Context) error {
		i++
		if i%2 == 0 {
			return errors.New("rpc error")
		}
		return nil
	})
	s := Compute(results)
	if s.Errors == 0 {
		t.Error("expected some errors")
	}
}

func TestCompute_Percentiles(t *testing.T) {
	results := make([]Result, 100)
	for i := range results {
		results[i] = Result{Latency: time.Duration(i+1) * time.Millisecond}
	}
	s := Compute(results)
	if s.Total != 100 {
		t.Errorf("expected 100, got %d", s.Total)
	}
	if s.Min != time.Millisecond {
		t.Errorf("unexpected min: %v", s.Min)
	}
}

func TestPrint_Output(t *testing.T) {
	s := Stats{Total: 10, Errors: 1, Min: time.Millisecond, Max: 10 * time.Millisecond,
		Mean: 5 * time.Millisecond, P50: 5 * time.Millisecond, P95: 9 * time.Millisecond, P99: 10 * time.Millisecond}
	var buf bytes.Buffer
	Print(&buf, s)
	if !strings.Contains(buf.String(), "p99") {
		t.Error("expected p99 in output")
	}
}
