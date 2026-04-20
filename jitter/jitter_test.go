package jitter_test

import (
	"testing"
	"time"

	"github.com/grpcannon/jitter"
)

const base = 100 * time.Millisecond

func TestFull_ZeroBase(t *testing.T) {
	j := jitter.New(jitter.Full)
	if got := j.Apply(0, 0); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestFull_InRange(t *testing.T) {
	j := jitter.New(jitter.Full)
	for i := 0; i < 200; i++ {
		got := j.Apply(base, 0)
		if got < 0 || got > base {
			t.Fatalf("Full: %v out of [0, %v]", got, base)
		}
	}
}

func TestEqual_InRange(t *testing.T) {
	j := jitter.New(jitter.Equal)
	half := base / 2
	for i := 0; i < 200; i++ {
		got := j.Apply(base, 0)
		if got < half || got > base {
			t.Fatalf("Equal: %v out of [%v, %v]", got, half, base)
		}
	}
}

func TestDecorelated_FirstCall_UsesBase(t *testing.T) {
	j := jitter.New(jitter.Decorrelated)
	for i := 0; i < 200; i++ {
		got := j.Apply(base, 0)
		if got < base || got > base*3 {
			t.Fatalf("Decorrelated(prev=0): %v out of [%v, %v]", got, base, base*3)
		}
	}
}

func TestDecorelated_WithPrev(t *testing.T) {
	j := jitter.New(jitter.Decorrelated)
	prev := 200 * time.Millisecond
	for i := 0; i < 200; i++ {
		got := j.Apply(base, prev)
		if got < base || got > prev*3 {
			t.Fatalf("Decorrelated: %v out of [%v, %v]", got, base, prev*3)
		}
	}
}

func TestFull_NegativeBase(t *testing.T) {
	j := jitter.New(jitter.Full)
	if got := j.Apply(-1*time.Second, 0); got != 0 {
		t.Fatalf("expected 0 for negative base, got %v", got)
	}
}
