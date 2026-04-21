package window_test

import (
	"testing"
	"time"

	"github.com/example/grpcannon/window"
)

func TestNew_PanicsOnZeroBuckets(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero buckets")
		}
	}()
	window.New(time.Second, 0)
}

func TestNew_PanicsOnNegativeDuration(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for non-positive duration")
		}
	}()
	window.New(-time.Second, 4)
}

func TestRecord_CountsRequests(t *testing.T) {
	w := window.New(time.Second, 10)
	w.Record(false)
	w.Record(false)
	w.Record(false)
	reqs, errs := w.Counts()
	if reqs != 3 {
		t.Fatalf("expected 3 requests, got %d", reqs)
	}
	if errs != 0 {
		t.Fatalf("expected 0 errors, got %d", errs)
	}
}

func TestRecord_CountsErrors(t *testing.T) {
	w := window.New(time.Second, 10)
	w.Record(false)
	w.Record(true)
	w.Record(true)
	reqs, errs := w.Counts()
	if reqs != 3 {
		t.Fatalf("expected 3 requests, got %d", reqs)
	}
	if errs != 2 {
		t.Fatalf("expected 2 errors, got %d", errs)
	}
}

func TestCounts_ZeroInitially(t *testing.T) {
	w := window.New(time.Second, 5)
	reqs, errs := w.Counts()
	if reqs != 0 || errs != 0 {
		t.Fatalf("expected zero counts, got reqs=%d errs=%d", reqs, errs)
	}
}

func TestRecord_ConcurrentSafe(t *testing.T) {
	w := window.New(time.Second, 10)
	done := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func(i int) {
			w.Record(i%2 == 0)
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < 50; i++ {
		<-done
	}
	reqs, _ := w.Counts()
	if reqs != 50 {
		t.Fatalf("expected 50 requests, got %d", reqs)
	}
}
