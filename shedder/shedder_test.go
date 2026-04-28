package shedder_test

import (
	"context"
	"testing"
	"time"

	"github.com/grpcannon/shedder"
)

func TestNew_ZeroThreshold_NeverSheds(t *testing.T) {
	s := shedder.New(0, time.Second)
	for i := 0; i < 100; i++ {
		s.Record(true)
	}
	if err := s.Allow(context.Background()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestAllow_BelowThreshold_Succeeds(t *testing.T) {
	s := shedder.New(0.5, time.Minute)
	// 2 errors out of 10 = 0.2 rate, below 0.5 threshold
	for i := 0; i < 8; i++ {
		s.Record(false)
	}
	for i := 0; i < 2; i++ {
		s.Record(true)
	}
	if err := s.Allow(context.Background()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestAllow_AboveThreshold_Sheds(t *testing.T) {
	s := shedder.New(0.5, time.Minute)
	// 8 errors out of 10 = 0.8 rate, above 0.5 threshold
	for i := 0; i < 2; i++ {
		s.Record(false)
	}
	for i := 0; i < 8; i++ {
		s.Record(true)
	}
	if err := s.Allow(context.Background()); err != shedder.ErrShed {
		t.Fatalf("expected ErrShed, got %v", err)
	}
}

func TestRate_Calculated(t *testing.T) {
	s := shedder.New(0.9, time.Minute)
	s.Record(false)
	s.Record(true)
	s.Record(true)
	got := s.Rate()
	want := 2.0 / 3.0
	if got < want-0.01 || got > want+0.01 {
		t.Fatalf("rate = %.4f, want ~%.4f", got, want)
	}
}

func TestRate_ZeroSamples(t *testing.T) {
	s := shedder.New(0.5, time.Minute)
	if r := s.Rate(); r != 0 {
		t.Fatalf("expected 0, got %f", r)
	}
}

func TestNew_ClampsThreshold(t *testing.T) {
	// Should not panic or shed on negative threshold (treated as 0).
	s := shedder.New(-1, time.Second)
	s.Record(true)
	if err := s.Allow(context.Background()); err != nil {
		t.Fatalf("expected nil for clamped-zero threshold, got %v", err)
	}
}

func TestWindow_ExpiredSamplesIgnored(t *testing.T) {
	s := shedder.New(0.5, 50*time.Millisecond)
	// Record errors that will expire.
	for i := 0; i < 10; i++ {
		s.Record(true)
	}
	time.Sleep(60 * time.Millisecond)
	// After expiry the window is empty → rate 0 → allowed.
	if err := s.Allow(context.Background()); err != nil {
		t.Fatalf("expected nil after window expiry, got %v", err)
	}
}
