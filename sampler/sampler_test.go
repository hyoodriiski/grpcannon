package sampler_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/grpcannon/sampler"
)

func TestNew_ZeroRate_RecordsNothing(t *testing.T) {
	s := sampler.New(0)
	for i := 0; i < 1000; i++ {
		s.Record("SomeMethod", 1.5, nil)
	}
	if s.Len() != 0 {
		t.Fatalf("expected 0 samples with rate=0, got %d", s.Len())
	}
}

func TestNew_FullRate_RecordsAll(t *testing.T) {
	s := sampler.New(1.0)
	const n = 200
	for i := 0; i < n; i++ {
		s.Record("Ping", float64(i), nil)
	}
	if s.Len() != n {
		t.Fatalf("expected %d samples, got %d", n, s.Len())
	}
}

func TestNew_ClampAboveOne(t *testing.T) {
	s := sampler.New(5.0)
	s.Record("X", 1.0, nil)
	if s.Len() != 1 {
		t.Fatalf("rate clamped to 1.0 should record every call")
	}
}

func TestNew_ClampBelowZero(t *testing.T) {
	s := sampler.New(-3.0)
	s.Record("X", 1.0, nil)
	if s.Len() != 0 {
		t.Fatalf("rate clamped to 0.0 should record nothing")
	}
}

func TestRecord_StoresFields(t *testing.T) {
	s := sampler.New(1.0)
	sentErr := errors.New("rpc error")
	s.Record("MyMethod", 42.5, sentErr)

	samples := s.Samples()
	if len(samples) != 1 {
		t.Fatalf("expected 1 sample, got %d", len(samples))
	}
	got := samples[0]
	if got.Method != "MyMethod" {
		t.Errorf("method: want MyMethod, got %s", got.Method)
	}
	if got.LatencyMs != 42.5 {
		t.Errorf("latency: want 42.5, got %f", got.LatencyMs)
	}
	if got.Error != sentErr {
		t.Errorf("error: want %v, got %v", sentErr, got.Error)
	}
	if got.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}
}

func TestReset_ClearsSamples(t *testing.T) {
	s := sampler.New(1.0)
	s.Record("A", 1.0, nil)
	s.Record("B", 2.0, nil)
	s.Reset()
	if s.Len() != 0 {
		t.Fatalf("expected 0 after reset, got %d", s.Len())
	}
}

func TestRecord_ConcurrentSafe(t *testing.T) {
	s := sampler.New(1.0)
	const goroutines = 50
	const perG = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < perG; j++ {
				s.Record("Concurrent", float64(j), nil)
			}
		}()
	}
	wg.Wait()
	if s.Len() != goroutines*perG {
		t.Fatalf("expected %d samples, got %d", goroutines*perG, s.Len())
	}
}
