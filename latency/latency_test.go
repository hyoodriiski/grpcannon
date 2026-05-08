package latency

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestNew_DefaultCap(t *testing.T) {
	tr := New(0)
	if tr.cap != 1024 {
		t.Fatalf("expected cap 1024, got %d", tr.cap)
	}
}

func TestRecord_And_Samples(t *testing.T) {
	tr := New(4)
	tr.Record(10 * time.Millisecond)
	tr.Record(20 * time.Millisecond)
	tr.Record(5 * time.Millisecond)

	samples := tr.Samples()
	if len(samples) != 3 {
		t.Fatalf("expected 3 samples, got %d", len(samples))
	}
	// Samples must be sorted.
	if samples[0] != 5*time.Millisecond || samples[2] != 20*time.Millisecond {
		t.Fatalf("samples not sorted: %v", samples)
	}
}

func TestRecord_RingWrap(t *testing.T) {
	tr := New(3)
	for i := 0; i < 6; i++ {
		tr.Record(time.Duration(i+1) * time.Millisecond)
	}
	// Buffer has wrapped; should still hold exactly cap samples.
	if len(tr.Samples()) != 3 {
		t.Fatalf("expected 3 samples after wrap, got %d", len(tr.Samples()))
	}
}

func TestPercentile_Empty(t *testing.T) {
	tr := New(10)
	if tr.Percentile(99) != 0 {
		t.Fatal("expected 0 for empty tracker")
	}
}

func TestPercentile_Values(t *testing.T) {
	tr := New(100)
	for i := 1; i <= 100; i++ {
		tr.Record(time.Duration(i) * time.Millisecond)
	}
	p50 := tr.Percentile(50)
	if p50 < 49*time.Millisecond || p50 > 51*time.Millisecond {
		t.Fatalf("unexpected p50: %v", p50)
	}
	p99 := tr.Percentile(99)
	if p99 < 98*time.Millisecond || p99 > 100*time.Millisecond {
		t.Fatalf("unexpected p99: %v", p99)
	}
}

func TestReset_ClearsSamples(t *testing.T) {
	tr := New(10)
	tr.Record(1 * time.Millisecond)
	tr.Reset()
	if len(tr.Samples()) != 0 {
		t.Fatal("expected 0 samples after reset")
	}
}

func TestRecord_ConcurrentSafe(t *testing.T) {
	tr := New(512)
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			tr.Record(time.Duration(n) * time.Microsecond)
		}(i)
	}
	wg.Wait()
}

func TestGuard_RecordsLatency(t *testing.T) {
	tr := New(10)
	next := func(_ context.Context) error {
		time.Sleep(5 * time.Millisecond)
		return nil
	}
	guarded := Guard(tr, next)
	if err := guarded(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tr.Samples()) != 1 {
		t.Fatal("expected 1 sample recorded")
	}
	if tr.Samples()[0] < 5*time.Millisecond {
		t.Fatal("recorded latency too low")
	}
}

func TestGuard_RecordsOnError(t *testing.T) {
	tr := New(10)
	errBoom := errors.New("boom")
	next := func(_ context.Context) error { return errBoom }
	guarded := Guard(tr, next)
	if err := guarded(context.Background()); !errors.Is(err, errBoom) {
		t.Fatalf("expected errBoom, got %v", err)
	}
	if len(tr.Samples()) != 1 {
		t.Fatal("expected sample recorded even on error")
	}
}
