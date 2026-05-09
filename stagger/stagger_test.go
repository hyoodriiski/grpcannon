package stagger_test

import (
	"context"
	"testing"
	"time"

	"grpcannon/stagger"
)

func TestNew_ZeroConcurrency_ReturnsErr(t *testing.T) {
	_, err := stagger.New(time.Second, 0)
	if err == nil {
		t.Fatal("expected error for zero concurrency")
	}
}

func TestNew_ValidConcurrency_ComputesDelay(t *testing.T) {
	s, err := stagger.New(time.Second, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Delay() != 250*time.Millisecond {
		t.Fatalf("expected 250ms, got %v", s.Delay())
	}
}

func TestWait_IndexZero_ReturnsImmediately(t *testing.T) {
	s, _ := stagger.New(time.Second, 10)
	start := time.Now()
	if err := s.Wait(context.Background(), 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if time.Since(start) > 20*time.Millisecond {
		t.Fatal("index 0 should return immediately")
	}
}

func TestWait_ContextCancelled_ReturnsErr(t *testing.T) {
	s, _ := stagger.New(10*time.Second, 1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.Wait(ctx, 1); err == nil {
		t.Fatal("expected context error")
	}
}

func TestWait_SmallDelay_Elapses(t *testing.T) {
	s, _ := stagger.New(100*time.Millisecond, 10) // 10ms per step
	start := time.Now()
	if err := s.Wait(context.Background(), 2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	elapsed := time.Since(start)
	if elapsed < 15*time.Millisecond {
		t.Fatalf("expected at least 15ms, got %v", elapsed)
	}
}
