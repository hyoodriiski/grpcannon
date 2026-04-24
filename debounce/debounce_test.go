package debounce_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/grpcannon/debounce"
)

func TestNew_NilFn_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil fn")
		}
	}()
	debounce.New(10*time.Millisecond, nil)
}

func TestTrigger_FiresAfterWait(t *testing.T) {
	var count int32
	d := debounce.New(30*time.Millisecond, func() {
		atomic.AddInt32(&count, 1)
	})

	d.Trigger()
	time.Sleep(60 * time.Millisecond)

	if got := atomic.LoadInt32(&count); got != 1 {
		t.Fatalf("expected 1 invocation, got %d", got)
	}
}

func TestTrigger_CoalescesRapidCalls(t *testing.T) {
	var count int32
	d := debounce.New(50*time.Millisecond, func() {
		atomic.AddInt32(&count, 1)
	})

	// Fire 10 times in quick succession.
	for i := 0; i < 10; i++ {
		d.Trigger()
		time.Sleep(5 * time.Millisecond)
	}

	time.Sleep(100 * time.Millisecond)

	if got := atomic.LoadInt32(&count); got != 1 {
		t.Fatalf("expected 1 coalesced invocation, got %d", got)
	}
}

func TestCancel_PreventsInvocation(t *testing.T) {
	var count int32
	d := debounce.New(40*time.Millisecond, func() {
		atomic.AddInt32(&count, 1)
	})

	d.Trigger()
	d.Cancel()
	time.Sleep(80 * time.Millisecond)

	if got := atomic.LoadInt32(&count); got != 0 {
		t.Fatalf("expected 0 invocations after cancel, got %d", got)
	}
}

func TestCancel_Idempotent(t *testing.T) {
	d := debounce.New(20*time.Millisecond, func() {})
	// Cancel with no pending timer should not panic.
	d.Cancel()
	d.Cancel()
}

func TestNew_NegativeWait_TreatedAsZero(t *testing.T) {
	var count int32
	d := debounce.New(-1*time.Second, func() {
		atomic.AddInt32(&count, 1)
	})
	d.Trigger()
	time.Sleep(20 * time.Millisecond)
	if got := atomic.LoadInt32(&count); got != 1 {
		t.Fatalf("expected 1 invocation with zero wait, got %d", got)
	}
}
