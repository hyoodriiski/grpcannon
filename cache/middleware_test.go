package cache_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grpcannon/cache"
)

func TestGuard_CachesResponse(t *testing.T) {
	c := cache.New(time.Minute)
	calls := 0
	next := func(_ context.Context, _ string, _ interface{}) (interface{}, error) {
		calls++
		return "resp", nil
	}
	guarded := cache.Guard(c, time.Minute, next)

	for i := 0; i < 3; i++ {
		v, err := guarded(context.Background(), "method", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.(string) != "resp" {
			t.Fatalf("unexpected response: %v", v)
		}
	}
	if calls != 1 {
		t.Fatalf("expected 1 upstream call, got %d", calls)
	}
}

func TestGuard_DoesNotCacheErrors(t *testing.T) {
	c := cache.New(time.Minute)
	calls := 0
	next := func(_ context.Context, _ string, _ interface{}) (interface{}, error) {
		calls++
		return nil, errors.New("upstream error")
	}
	guarded := cache.Guard(c, time.Minute, next)

	for i := 0; i < 3; i++ {
		_, err := guarded(context.Background(), "method", nil)
		if err == nil {
			t.Fatal("expected error")
		}
	}
	if calls != 3 {
		t.Fatalf("expected 3 upstream calls, got %d", calls)
	}
}

func TestGuard_ExpiredEntry_Refetches(t *testing.T) {
	c := cache.New(0)
	calls := 0
	next := func(_ context.Context, _ string, _ interface{}) (interface{}, error) {
		calls++
		return "v", nil
	}
	guarded := cache.Guard(c, time.Millisecond, next)

	_, _ = guarded(context.Background(), "m", nil)
	time.Sleep(5 * time.Millisecond)
	_, _ = guarded(context.Background(), "m", nil)

	if calls != 2 {
		t.Fatalf("expected 2 calls after expiry, got %d", calls)
	}
}
