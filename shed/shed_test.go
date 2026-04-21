package shed_test

import (
	"context"
	"sync"
	"testing"

	"github.com/example/grpcannon/shed"
)

func TestNew_ZeroMax_NeverSheds(t *testing.T) {
	s := shed.New(0)
	for i := 0; i < 1000; i++ {
		release, err := s.Acquire(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		release()
	}
}

func TestAcquire_WithinLimit_Succeeds(t *testing.T) {
	s := shed.New(3)
	var releases []func()
	for i := 0; i < 3; i++ {
		release, err := s.Acquire(context.Background())
		if err != nil {
			t.Fatalf("unexpected error on acquire %d: %v", i, err)
		}
		releases = append(releases, release)
	}
	if got := s.Inflight(); got != 3 {
		t.Fatalf("expected 3 inflight, got %d", got)
	}
	for _, r := range releases {
		r()
	}
	if got := s.Inflight(); got != 0 {
		t.Fatalf("expected 0 inflight after release, got %d", got)
	}
}

func TestAcquire_ExceedsLimit_ReturnsShed(t *testing.T) {
	s := shed.New(2)
	r1, _ := s.Acquire(context.Background())
	r2, _ := s.Acquire(context.Background())
	defer r1()
	defer r2()

	_, err := s.Acquire(context.Background())
	if err != shed.ErrShed {
		t.Fatalf("expected ErrShed, got %v", err)
	}
	if got := s.Inflight(); got != 2 {
		t.Fatalf("inflight should remain 2 after shed, got %d", got)
	}
}

func TestAcquire_CancelledContext_ReturnsErr(t *testing.T) {
	s := shed.New(10)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := s.Acquire(ctx)
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

func TestAcquire_ConcurrentSafe(t *testing.T) {
	const max = 5
	s := shed.New(max)
	var wg sync.WaitGroup
	var shed_count, ok_count int64
	var mu sync.Mutex

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			release, err := s.Acquire(context.Background())
			mu.Lock()
			if err != nil {
				shed_count++
			} else {
				ok_count++
				defer release()
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	if s.Inflight() != 0 {
		t.Fatalf("expected 0 inflight after all goroutines done, got %d", s.Inflight())
	}
	_ = shed_count
	_ = ok_count
}
