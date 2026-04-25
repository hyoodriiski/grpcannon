package slope

import (
	"math"
	"testing"
	"time"
)

func TestNew_ClampsWindow(t *testing.T) {
	s := New(0)
	if s.window != 2 {
		t.Fatalf("expected window=2, got %d", s.window)
	}
	s2 := New(1)
	if s2.window != 2 {
		t.Fatalf("expected window=2, got %d", s2.window)
	}
}

func TestRate_TooFewSamples(t *testing.T) {
	s := New(10)
	if r := s.Rate(); r != 0 {
		t.Fatalf("expected 0 with no samples, got %f", r)
	}
	s.Record(1.0)
	if r := s.Rate(); r != 0 {
		t.Fatalf("expected 0 with one sample, got %f", r)
	}
}

func TestRate_PositiveSlope(t *testing.T) {
	s := New(10)
	// Inject points manually with known timestamps.
	now := time.Now()
	s.mu.Lock()
	s.samples = []Point{
		{At: now, Value: 0},
		{At: now.Add(1 * time.Second), Value: 10},
		{At: now.Add(2 * time.Second), Value: 20},
	}
	s.mu.Unlock()

	rate := s.Rate()
	if math.Abs(rate-10.0) > 0.01 {
		t.Fatalf("expected slope ~10.0, got %f", rate)
	}
}

func TestRate_NegativeSlope(t *testing.T) {
	s := New(10)
	now := time.Now()
	s.mu.Lock()
	s.samples = []Point{
		{At: now, Value: 30},
		{At: now.Add(1 * time.Second), Value: 20},
		{At: now.Add(2 * time.Second), Value: 10},
	}
	s.mu.Unlock()

	rate := s.Rate()
	if math.Abs(rate+10.0) > 0.01 {
		t.Fatalf("expected slope ~-10.0, got %f", rate)
	}
}

func TestRate_WindowTrimming(t *testing.T) {
	s := New(3)
	for i := 0; i < 10; i++ {
		s.Record(float64(i))
		time.Sleep(1 * time.Millisecond)
	}
	s.mu.Lock()
	n := len(s.samples)
	s.mu.Unlock()
	if n != 3 {
		t.Fatalf("expected 3 retained samples, got %d", n)
	}
}

func TestReset_ClearsSamples(t *testing.T) {
	s := New(5)
	s.Record(1)
	s.Record(2)
	s.Reset()
	if r := s.Rate(); r != 0 {
		t.Fatalf("expected 0 after reset, got %f", r)
	}
}

func TestRate_FlatLine_ZeroSlope(t *testing.T) {
	s := New(10)
	now := time.Now()
	s.mu.Lock()
	s.samples = []Point{
		{At: now, Value: 5},
		{At: now.Add(1 * time.Second), Value: 5},
		{At: now.Add(2 * time.Second), Value: 5},
	}
	s.mu.Unlock()
	rate := s.Rate()
	if math.Abs(rate) > 0.001 {
		t.Fatalf("expected slope ~0, got %f", rate)
	}
}
