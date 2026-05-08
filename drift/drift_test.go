package drift

import (
	"testing"
	"time"
)

func TestNew_DefaultsInvalidFactor(t *testing.T) {
	d := New(10*time.Millisecond, 0.5, 10)
	if d.factor != 2.0 {
		t.Fatalf("expected factor=2.0, got %v", d.factor)
	}
}

func TestNew_DefaultsInvalidCap(t *testing.T) {
	d := New(10*time.Millisecond, 2.0, 0)
	if d.cap != 100 {
		t.Fatalf("expected cap=100, got %v", d.cap)
	}
}

func TestRecord_BelowThreshold_NoError(t *testing.T) {
	d := New(100*time.Millisecond, 2.0, 10)
	if err := d.Record(50 * time.Millisecond); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRecord_AboveThreshold_ReturnsDrifted(t *testing.T) {
	d := New(10*time.Millisecond, 2.0, 10)
	var last error
	for i := 0; i < 5; i++ {
		last = d.Record(100 * time.Millisecond)
	}
	if last != ErrDrifted {
		t.Fatalf("expected ErrDrifted, got %v", last)
	}
}

func TestReset_ClearsSamples(t *testing.T) {
	d := New(10*time.Millisecond, 2.0, 10)
	for i := 0; i < 5; i++ {
		_ = d.Record(100 * time.Millisecond)
	}
	d.Reset()
	if avg := d.Avg(); avg != 0 {
		t.Fatalf("expected avg=0 after reset, got %v", avg)
	}
}

func TestAvg_Calculated(t *testing.T) {
	d := New(100*time.Millisecond, 2.0, 10)
	_ = d.Record(20 * time.Millisecond)
	_ = d.Record(40 * time.Millisecond)
	got := d.Avg()
	want := 30 * time.Millisecond
	if got != want {
		t.Fatalf("expected avg=%v, got %v", want, got)
	}
}

func TestRecord_WindowCap_TrimsOldSamples(t *testing.T) {
	d := New(10*time.Millisecond, 2.0, 3)
	// Fill with high values then overwrite with low values.
	for i := 0; i < 3; i++ {
		_ = d.Record(500 * time.Millisecond)
	}
	for i := 0; i < 3; i++ {
		_ = d.Record(1 * time.Millisecond)
	}
	if avg := d.Avg(); avg > 10*time.Millisecond {
		t.Fatalf("expected low avg after window trim, got %v", avg)
	}
}
