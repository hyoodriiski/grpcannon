package backoff_test

import (
	"testing"
	"time"

	"github.com/example/grpcannon/backoff"
)

func TestDefault_Values(t *testing.T) {
	s := backoff.Default()
	if s.InitialDelay != 50*time.Millisecond {
		t.Errorf("expected InitialDelay 50ms, got %v", s.InitialDelay)
	}
	if s.MaxDelay != 2*time.Second {
		t.Errorf("expected MaxDelay 2s, got %v", s.MaxDelay)
	}
	if s.Multiplier != 2.0 {
		t.Errorf("expected Multiplier 2.0, got %v", s.Multiplier)
	}
}

func TestDelay_ZeroAttempt(t *testing.T) {
	s := backoff.Strategy{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		Jitter:       0,
	}
	d := s.Delay(0)
	if d != 100*time.Millisecond {
		t.Errorf("expected 100ms, got %v", d)
	}
}

func TestDelay_Grows(t *testing.T) {
	s := backoff.Strategy{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
		Jitter:       0,
	}
	d0 := s.Delay(0)
	d1 := s.Delay(1)
	d2 := s.Delay(2)
	if d1 <= d0 {
		t.Errorf("expected delay to grow: d0=%v d1=%v", d0, d1)
	}
	if d2 <= d1 {
		t.Errorf("expected delay to grow: d1=%v d2=%v", d1, d2)
	}
}

func TestDelay_CappedAtMax(t *testing.T) {
	s := backoff.Strategy{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     200 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       0,
	}
	for i := 5; i < 20; i++ {
		d := s.Delay(i)
		if d > s.MaxDelay {
			t.Errorf("attempt %d: delay %v exceeds MaxDelay %v", i, d, s.MaxDelay)
		}
	}
}

func TestDelay_NegativeAttempt(t *testing.T) {
	s := backoff.Default()
	s.Jitter = 0
	d := s.Delay(-3)
	if d != s.InitialDelay {
		t.Errorf("expected InitialDelay for negative attempt, got %v", d)
	}
}

func TestSteps_Length(t *testing.T) {
	s := backoff.Default()
	steps := s.Steps(5)
	if len(steps) != 5 {
		t.Errorf("expected 5 steps, got %d", len(steps))
	}
}

func TestSteps_NonDecreasing(t *testing.T) {
	s := backoff.Strategy{
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   3.0,
		Jitter:       0,
	}
	steps := s.Steps(6)
	for i := 1; i < len(steps); i++ {
		if steps[i] < steps[i-1] {
			t.Errorf("steps[%d]=%v < steps[%d]=%v", i, steps[i], i-1, steps[i-1])
		}
	}
}
