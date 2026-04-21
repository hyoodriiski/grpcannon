package gate_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"grpcannon/gate"
)

func TestNew_InitiallyClosed(t *testing.T) {
	g := gate.New()
	if g.IsOpen() {
		t.Fatal("expected gate to be closed initially")
	}
}

func TestOpen_SetsOpenTrue(t *testing.T) {
	g := gate.New()
	g.Open()
	if !g.IsOpen() {
		t.Fatal("expected gate to be open after Open()")
	}
}

func TestClose_AfterOpen_ClosesGate(t *testing.T) {
	g := gate.New()
	g.Open()
	g.Close()
	if g.IsOpen() {
		t.Fatal("expected gate to be closed after Close()")
	}
}

func TestWait_OpenGate_ReturnsImmediately(t *testing.T) {
	g := gate.New()
	g.Open()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := g.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWait_ClosedGate_BlocksUntilOpen(t *testing.T) {
	g := gate.New()
	var wg sync.WaitGroup
	wg.Add(1)
	var waitErr error
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		waitErr = g.Wait(ctx)
	}()
	time.Sleep(20 * time.Millisecond)
	g.Open()
	wg.Wait()
	if waitErr != nil {
		t.Fatalf("expected nil error, got %v", waitErr)
	}
}

func TestWait_ContextCancelled_ReturnsErr(t *testing.T) {
	g := gate.New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := g.Wait(ctx); err == nil {
		t.Fatal("expected context error, got nil")
	}
}

func TestOpen_Idempotent(t *testing.T) {
	g := gate.New()
	g.Open()
	g.Open() // should not panic
	if !g.IsOpen() {
		t.Fatal("gate should remain open")
	}
}

func TestWait_ConcurrentCallers_AllUnblocked(t *testing.T) {
	g := gate.New()
	const n = 20
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			errs[idx] = g.Wait(ctx)
		}(i)
	}
	time.Sleep(10 * time.Millisecond)
	g.Open()
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: unexpected error %v", i, err)
		}
	}
}
