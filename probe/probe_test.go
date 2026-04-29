package probe_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"grpcannon/probe"
)

func TestNew_DefaultInterval(t *testing.T) {
	p := probe.New(func(_ context.Context) error { return nil }, 0)
	if p == nil {
		t.Fatal("expected non-nil probe")
	}
}

func TestHealthy_InitiallyFalse(t *testing.T) {
	p := probe.New(func(_ context.Context) error { return nil }, time.Hour)
	if p.Healthy() {
		t.Fatal("probe should not be healthy before first check")
	}
}

func TestStart_ImmediateCheck_Healthy(t *testing.T) {
	p := probe.New(func(_ context.Context) error { return nil }, time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	if !p.Healthy() {
		t.Fatal("expected healthy after successful check")
	}
}

func TestStart_ImmediateCheck_Unhealthy(t *testing.T) {
	sentinel := errors.New("down")
	p := probe.New(func(_ context.Context) error { return sentinel }, time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	s := p.Status()
	if s.Healthy {
		t.Fatal("expected unhealthy")
	}
	if !errors.Is(s.Err, sentinel) {
		t.Fatalf("unexpected error: %v", s.Err)
	}
}

func TestStart_PeriodicChecks(t *testing.T) {
	var calls atomic.Int64
	p := probe.New(func(_ context.Context) error {
		calls.Add(1)
		return nil
	}, 20*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p.Start(ctx)
	time.Sleep(90 * time.Millisecond)
	if calls.Load() < 3 {
		t.Fatalf("expected at least 3 checks, got %d", calls.Load())
	}
}

func TestStop_HaltsChecks(t *testing.T) {
	var calls atomic.Int64
	p := probe.New(func(_ context.Context) error {
		calls.Add(1)
		return nil
	}, 20*time.Millisecond)
	ctx := context.Background()
	p.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	p.Stop()
	before := calls.Load()
	time.Sleep(60 * time.Millisecond)
	after := calls.Load()
	if after != before {
		t.Fatalf("expected no new checks after Stop, got %d extra", after-before)
	}
}

func TestStatus_AtTimestamp(t *testing.T) {
	p := probe.New(func(_ context.Context) error { return nil }, time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	before := time.Now()
	p.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	s := p.Status()
	if s.At.Before(before) {
		t.Fatal("expected At timestamp to be after test start")
	}
}
