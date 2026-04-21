package observe_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/grpcannon/observe"
)

func TestOn_NilHandlerIgnored(t *testing.T) {
	o := observe.New()
	o.On("evt", nil) // must not panic
	o.Emit("evt", nil)
}

func TestOn_EmptyEventIgnored(t *testing.T) {
	o := observe.New()
	called := false
	o.On("", func(_ string, _ any) { called = true })
	o.Emit("", nil)
	if called {
		t.Fatal("handler for empty event should not be registered")
	}
}

func TestEmit_CallsHandlerWithPayload(t *testing.T) {
	o := observe.New()
	var got any
	o.On("ping", func(_ string, p any) { got = p })
	o.Emit("ping", 42)
	if got != 42 {
		t.Fatalf("expected 42, got %v", got)
	}
}

func TestEmit_MultipleHandlers_AllCalled(t *testing.T) {
	o := observe.New()
	var count int32
	for i := 0; i < 3; i++ {
		o.On("tick", func(_ string, _ any) { atomic.AddInt32(&count, 1) })
	}
	o.Emit("tick", nil)
	if atomic.LoadInt32(&count) != 3 {
		t.Fatalf("expected 3 calls, got %d", count)
	}
}

func TestEmit_UnknownEvent_NoOp(t *testing.T) {
	o := observe.New()
	o.Emit("ghost", "data") // must not panic
}

func TestOff_RemovesHandlers(t *testing.T) {
	o := observe.New()
	called := false
	o.On("x", func(_ string, _ any) { called = true })
	o.Off("x")
	o.Emit("x", nil)
	if called {
		t.Fatal("handler should have been removed")
	}
}

func TestEvents_ReturnsRegisteredNames(t *testing.T) {
	o := observe.New()
	o.On("a", func(_ string, _ any) {})
	o.On("b", func(_ string, _ any) {})
	evts := o.Events()
	if len(evts) != 2 {
		t.Fatalf("expected 2 events, got %d", len(evts))
	}
}

func TestEmit_ConcurrentSafe(t *testing.T) {
	o := observe.New()
	var count int32
	o.On("c", func(_ string, _ any) { atomic.AddInt32(&count, 1) })
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			o.Emit("c", nil)
		}()
	}
	wg.Wait()
	if atomic.LoadInt32(&count) != 50 {
		t.Fatalf("expected 50, got %d", count)
	}
}
