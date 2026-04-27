package relay_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/your-org/grpcannon/relay"
)

func TestRegister_NilHandlerIgnored(t *testing.T) {
	r := relay.New[int]()
	r.Register(nil)
	if r.Len() != 0 {
		t.Fatalf("expected 0 handlers, got %d", r.Len())
	}
}

func TestSend_DeliversToAllHandlers(t *testing.T) {
	r := relay.New[string]()
	var got []string
	var mu sync.Mutex
	add := func(s string) {
		mu.Lock()
		got = append(got, s)
		mu.Unlock()
	}
	r.Register(add)
	r.Register(add)
	r.Send("hello")
	mu.Lock()
	defer mu.Unlock()
	if len(got) != 2 {
		t.Fatalf("expected 2 deliveries, got %d", len(got))
	}
}

func TestSend_NoHandlers_NoOp(t *testing.T) {
	r := relay.New[int]()
	// must not panic
	r.Send(42)
}

func TestPump_ForwardsValues(t *testing.T) {
	r := relay.New[int]()
	var count atomic.Int64
	r.Register(func(int) { count.Add(1) })

	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	ch <- 3
	close(ch)

	r.Pump(context.Background(), ch)
	if count.Load() != 3 {
		t.Fatalf("expected 3, got %d", count.Load())
	}
}

func TestPump_StopsOnContextCancel(t *testing.T) {
	r := relay.New[int]()
	var count atomic.Int64
	r.Register(func(int) { count.Add(1) })

	ch := make(chan int) // unbuffered, never written
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	r.Pump(ctx, ch)
	if count.Load() != 0 {
		t.Fatalf("expected 0, got %d", count.Load())
	}
}

func TestRegister_ConcurrentSafe(t *testing.T) {
	r := relay.New[int]()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Register(func(int) {})
			r.Send(1)
		}()
	}
	wg.Wait()
}
