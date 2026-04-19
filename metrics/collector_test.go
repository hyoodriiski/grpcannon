package metrics

import (
	"bytes"
	"sync"
	"testing"
	"time"
)

func TestRecord_ConcurrentSafe(t *testing.T) {
	c := NewCollector()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c.Record(time.Duration(i)*time.Millisecond, i%10 == 0)
		}(i)
	}
	wg.Wait()
	s := c.Snapshot()
	if s.Total != 100 {
		t.Fatalf("expected 100 total, got %d", s.Total)
	}
	if s.Errors != 10 {
		t.Fatalf("expected 10 errors, got %d", s.Errors)
	}
	if len(s.Latencies) != 90 {
		t.Fatalf("expected 90 latencies, got %d", len(s.Latencies))
	}
}

func TestErrorRate(t *testing.T) {
	s := Snapshot{Total: 4, Errors: 1}
	if s.ErrorRate() != 0.25 {
		t.Fatalf("expected 0.25, got %f", s.ErrorRate())
	}
}

func TestErrorRate_Zero(t *testing.T) {
	s := Snapshot{}
	if s.ErrorRate() != 0 {
		t.Fatal("expected 0 for empty snapshot")
	}
}

func TestSummarise_Percentiles(t *testing.T) {
	lats := make([]time.Duration, 100)
	for i := range lats {
		lats[i] = time.Duration(i+1) * time.Millisecond
	}
	s := Summarise(Snapshot{Latencies: lats, Total: 100})
	if s.P50 != 50*time.Millisecond {
		t.Errorf("P50 want 50ms got %v", s.P50)
	}
	if s.Min != time.Millisecond {
		t.Errorf("Min want 1ms got %v", s.Min)
	}
	if s.Max != 100*time.Millisecond {
		t.Errorf("Max want 100ms got %v", s.Max)
	}
}

func TestSummaryPrint(t *testing.T) {
	s := Summary{Total: 10, Errors: 1, ErrorRate: 0.1,
		Min: time.Millisecond, Mean: 5 * time.Millisecond, Max: 10 * time.Millisecond,
		P50: 5 * time.Millisecond, P95: 9 * time.Millisecond, P99: 10 * time.Millisecond}
	var buf bytes.Buffer
	s.Print(&buf)
	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}
