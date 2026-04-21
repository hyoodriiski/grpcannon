package quota_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/grpcannon/quota"
)

func TestNew_ZeroMax_Unlimited(t *testing.T) {
	q := quota.New(0)
	for i := 0; i < 1000; i++ {
		if err := q.Acquire(); err != nil {
			t.Fatalf("unexpected error at iteration %d: %v", i, err)
		}
	}
	if got := q.Remaining(); got != -1 {
		t.Fatalf("expected -1 for unlimited, got %d", got)
	}
}

func TestNew_NegativeMax_TreatedAsUnlimited(t *testing.T) {
	q := quota.New(-5)
	if err := q.Acquire(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAcquire_WithinLimit_Succeeds(t *testing.T) {
	q := quota.New(3)
	for i := 0; i < 3; i++ {
		if err := q.Acquire(); err != nil {
			t.Fatalf("expected success at %d: %v", i, err)
		}
	}
	if got := q.Used(); got != 3 {
		t.Fatalf("expected used=3, got %d", got)
	}
}

func TestAcquire_ExceedsLimit_ReturnsErr(t *testing.T) {
	q := quota.New(2)
	_ = q.Acquire()
	_ = q.Acquire()
	if err := q.Acquire(); err != quota.ErrQuotaExceeded {
		t.Fatalf("expected ErrQuotaExceeded, got %v", err)
	}
	if got := q.Remaining(); got != 0 {
		t.Fatalf("expected remaining=0, got %d", got)
	}
}

func TestClose_PreventsAcquire(t *testing.T) {
	q := quota.New(100)
	q.Close()
	if err := q.Acquire(); err != quota.ErrQuotaExceeded {
		t.Fatalf("expected ErrQuotaExceeded after close, got %v", err)
	}
}

func TestAcquire_ConcurrentSafe(t *testing.T) {
	const limit = 50
	q := quota.New(limit)
	var wg sync.WaitGroup
	successes := make(chan struct{}, 200)
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if q.Acquire() == nil {
				successes <- struct{}{}
			}
		}()
	}
	wg.Wait()
	close(successes)
	count := 0
	for range successes {
		count++
	}
	if count != limit {
		t.Fatalf("expected exactly %d successes, got %d", limit, count)
	}
}

func TestGuard_SkipsCallOnExhaustedQuota(t *testing.T) {
	q := quota.New(1)
	called := 0
	fn := quota.Guard(q, func(_ context.Context) (time.Duration, error) {
		called++
		return time.Millisecond, nil
	})
	if _, err := fn(context.Background()); err != nil {
		t.Fatalf("first call should succeed: %v", err)
	}
	_, err := fn(context.Background())
	if err != quota.ErrQuotaExceeded {
		t.Fatalf("expected ErrQuotaExceeded, got %v", err)
	}
	if called != 1 {
		t.Fatalf("inner fn should be called once, got %d", called)
	}
}
