package circuit_test

import (
	"testing"
	"time"

	"github.com/grpcannon/circuit"
)

func TestNew_DefaultsThreshold(t *testing.T) {
	b := circuit.New(0, time.Second)
	if b == nil {
		t.Fatal("expected non-nil breaker")
	}
	// threshold clamped to 1 — one failure should open it
	b.RecordFailure()
	if b.CurrentState() != circuit.StateOpen {
		t.Errorf("expected StateOpen after one failure with threshold=1")
	}
}

func TestAllow_ClosedState(t *testing.T) {
	b := circuit.New(3, time.Second)
	if err := b.Allow(); err != nil {
		t.Errorf("expected nil error in closed state, got %v", err)
	}
}

func TestRecordFailure_OpensAfterThreshold(t *testing.T) {
	b := circuit.New(3, time.Second)
	b.RecordFailure()
	b.RecordFailure()
	if b.CurrentState() != circuit.StateClosed {
		t.Error("should still be closed after 2 failures")
	}
	b.RecordFailure()
	if b.CurrentState() != circuit.StateOpen {
		t.Error("should be open after 3 failures")
	}
}

func TestAllow_OpenState_ReturnsErr(t *testing.T) {
	b := circuit.New(1, time.Hour)
	b.RecordFailure()
	if err := b.Allow(); err != circuit.ErrOpen {
		t.Errorf("expected ErrOpen, got %v", err)
	}
}

func TestAllow_TransitionsToHalfOpen(t *testing.T) {
	b := circuit.New(1, time.Millisecond)
	b.RecordFailure()
	time.Sleep(5 * time.Millisecond)
	if err := b.Allow(); err != nil {
		t.Errorf("expected nil in half-open transition, got %v", err)
	}
	if b.CurrentState() != circuit.StateHalfOpen {
		t.Errorf("expected StateHalfOpen, got %v", b.CurrentState())
	}
}

func TestRecordSuccess_ClosesBreakerAndResetsFailures(t *testing.T) {
	b := circuit.New(2, time.Hour)
	b.RecordFailure()
	b.RecordSuccess()
	if b.CurrentState() != circuit.StateClosed {
		t.Error("expected StateClosed after success")
	}
	// one more failure should not open it (counter reset)
	b.RecordFailure()
	if b.CurrentState() != circuit.StateClosed {
		t.Error("expected StateClosed — counter should have reset")
	}
}

func TestConcurrentSafe(t *testing.T) {
	b := circuit.New(100, time.Second)
	done := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func() {
			_ = b.Allow()
			b.RecordFailure()
			b.RecordSuccess()
			done <- struct{}{}
		}()
	}
	for i := 0; i < 50; i++ {
		<-done
	}
}
