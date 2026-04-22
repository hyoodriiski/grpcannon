package budget_test

import (
	"errors"
	"sync"
	"testing"

	"grpcannon/budget"
)

func TestNew_ClampsThreshold(t *testing.T) {
	b := budget.New(-0.5)
	if b.Rate() != 0 {
		t.Fatalf("expected 0 rate, got %f", b.Rate())
	}
}

func TestRecord_NoErrors_NotExhausted(t *testing.T) {
	b := budget.New(0.1)
	for i := 0; i < 10; i++ {
		if err := b.Record(false); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if b.Exhausted() {
		t.Fatal("budget should not be exhausted")
	}
}

func TestRecord_ExceedsThreshold_ReturnsExhausted(t *testing.T) {
	b := budget.New(0.1) // 10% threshold
	// 9 successes, 1 error → ~10% — just at threshold, not over
	for i := 0; i < 9; i++ {
		_ = b.Record(false)
	}
	// one more error tips it over
	_ = b.Record(true)
	// now add another error to push rate over 10%
	err := b.Record(true)
	if !errors.Is(err, budget.ErrExhausted) {
		t.Fatalf("expected ErrExhausted, got %v", err)
	}
	if !b.Exhausted() {
		t.Fatal("budget should be exhausted")
	}
}

func TestRate_Calculated(t *testing.T) {
	b := budget.New(0.5)
	_ = b.Record(false)
	_ = b.Record(true)
	if got := b.Rate(); got != 0.5 {
		t.Fatalf("expected 0.5, got %f", got)
	}
}

func TestRate_ZeroTotal(t *testing.T) {
	b := budget.New(0.1)
	if b.Rate() != 0 {
		t.Fatal("expected 0 rate when no records")
	}
}

func TestReset_ClearsState(t *testing.T) {
	b := budget.New(0.0) // zero tolerance
	_ = b.Record(true)
	if !b.Exhausted() {
		t.Fatal("should be exhausted")
	}
	b.Reset()
	if b.Exhausted() {
		t.Fatal("should not be exhausted after reset")
	}
	if b.Rate() != 0 {
		t.Fatal("rate should be 0 after reset")
	}
}

func TestRecord_ConcurrentSafe(t *testing.T) {
	b := budget.New(0.5)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = b.Record(i%2 == 0)
		}(i)
	}
	wg.Wait()
}

func TestGuard_SkipsWhenExhausted(t *testing.T) {
	b := budget.New(0.0) // zero tolerance
	_ = b.Record(true)  // exhaust immediately

	called := false
	err := budget.Guard(b, func() error {
		called = true
		return nil
	})
	if called {
		t.Fatal("invoker should not have been called")
	}
	if !errors.Is(err, budget.ErrExhausted) {
		t.Fatalf("expected ErrExhausted, got %v", err)
	}
}

func TestGuard_RecordsCallError(t *testing.T) {
	b := budget.New(0.5)
	sentinel := errors.New("call failed")
	err := budget.Guard(b, func() error { return sentinel })
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if b.Rate() != 1.0 {
		t.Fatalf("expected error rate 1.0, got %f", b.Rate())
	}
}
