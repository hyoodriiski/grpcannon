package snapshot_test

import (
	"sync"
	"testing"
	"time"

	"github.com/example/grpcannon/snapshot"
)

func TestErrorRate_Zero(t *testing.T) {
	s := snapshot.Snapshot{}
	if got := s.ErrorRate(); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestErrorRate_Calculated(t *testing.T) {
	s := snapshot.Snapshot{Requests: 10, Errors: 2}
	if got := s.ErrorRate(); got != 0.2 {
		t.Fatalf("expected 0.2, got %v", got)
	}
}

func TestAvgLatency_Zero(t *testing.T) {
	s := snapshot.Snapshot{}
	if got := s.AvgLatency(); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestAvgLatency_Calculated(t *testing.T) {
	s := snapshot.Snapshot{
		Requests:     4,
		TotalLatency: 400 * time.Millisecond,
	}
	if got := s.AvgLatency(); got != 100*time.Millisecond {
		t.Fatalf("expected 100ms, got %v", got)
	}
}

func TestRecorder_Record(t *testing.T) {
	r := &snapshot.Recorder{}
	r.Record(10*time.Millisecond, false)
	r.Record(20*time.Millisecond, true)

	s := r.Take()
	if s.Requests != 2 {
		t.Fatalf("expected 2 requests, got %d", s.Requests)
	}
	if s.Errors != 1 {
		t.Fatalf("expected 1 error, got %d", s.Errors)
	}
	if s.TotalLatency != 30*time.Millisecond {
		t.Fatalf("expected 30ms total latency, got %v", s.TotalLatency)
	}
}

func TestRecorder_ConcurrentSafe(t *testing.T) {
	r := &snapshot.Recorder{}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Record(time.Millisecond, false)
		}()
	}
	wg.Wait()
	if s := r.Take(); s.Requests != 100 {
		t.Fatalf("expected 100 requests, got %d", s.Requests)
	}
}

func TestDelta(t *testing.T) {
	a := snapshot.Snapshot{Requests: 5, Errors: 1, TotalLatency: 50 * time.Millisecond}
	b := snapshot.Snapshot{Requests: 15, Errors: 3, TotalLatency: 150 * time.Millisecond}
	d := snapshot.Delta(a, b)
	if d.Requests != 10 {
		t.Fatalf("expected 10, got %d", d.Requests)
	}
	if d.Errors != 2 {
		t.Fatalf("expected 2, got %d", d.Errors)
	}
	if d.TotalLatency != 100*time.Millisecond {
		t.Fatalf("expected 100ms, got %v", d.TotalLatency)
	}
}
