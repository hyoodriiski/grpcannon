package drain_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/example/grpcannon/drain"
)

func TestAcquire_AfterDrain_ReturnsFalse(t *testing.T) {
	d := drain.New(time.Second)
	// drain immediately with no in-flight work
	if err := d.Drain(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Acquire() {
		t.Fatal("expected Acquire to return false after Drain")
	}
}

func TestDrain_NoWork_ReturnsImmediately(t *testing.T) {
	d := drain.New(time.Second)
	start := time.Now()
	if err := d.Drain(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 200*time.Millisecond {
		t.Fatalf("Drain took too long: %v", elapsed)
	}
}

func TestDrain_WaitsForInFlight(t *testing.T) {
	d := drain.New(2 * time.Second)
	if !d.Acquire() {
		t.Fatal("expected Acquire to succeed")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond)
		d.Release()
	}()

	if err := d.Drain(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wg.Wait()
}

func TestDrain_Timeout_ReturnsDeadlineExceeded(t *testing.T) {
	d := drain.New(50 * time.Millisecond)
	if !d.Acquire() {
		t.Fatal("expected Acquire to succeed")
	}
	defer d.Release() // never released in time — intentional

	err := d.Drain(context.Background())
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestDrain_ContextCancelled(t *testing.T) {
	d := drain.New(5 * time.Second)
	if !d.Acquire() {
		t.Fatal("expected Acquire to succeed")
	}
	defer d.Release()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	err := d.Drain(ctx)
	if err != context.Canceled {
		t.Fatalf("expected Canceled, got %v", err)
	}
}
