package cache_test

import (
	"testing"
	"time"

	"github.com/grpcannon/cache"
)

func TestSet_And_Get(t *testing.T) {
	c := cache.New(time.Minute)
	c.Set("key", "value")
	v, err := c.Get("key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.(string) != "value" {
		t.Fatalf("expected value, got %v", v)
	}
}

func TestGet_NotFound(t *testing.T) {
	c := cache.New(time.Minute)
	_, err := c.Get("missing")
	if err != cache.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGet_Expired(t *testing.T) {
	c := cache.New(0)
	c.SetTTL("key", "val", time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	_, err := c.Get("key")
	if err != cache.ErrExpired {
		t.Fatalf("expected ErrExpired, got %v", err)
	}
}

func TestDelete_RemovesKey(t *testing.T) {
	c := cache.New(time.Minute)
	c.Set("k", 1)
	c.Delete("k")
	_, err := c.Get("k")
	if err != cache.ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestFlush_ClearsAll(t *testing.T) {
	c := cache.New(time.Minute)
	c.Set("a", 1)
	c.Set("b", 2)
	c.Flush()
	if c.Len() != 0 {
		t.Fatalf("expected empty cache after flush, got %d", c.Len())
	}
}

func TestLen_CountsEntries(t *testing.T) {
	c := cache.New(time.Minute)
	if c.Len() != 0 {
		t.Fatal("expected 0")
	}
	c.Set("x", true)
	c.Set("y", true)
	if c.Len() != 2 {
		t.Fatalf("expected 2, got %d", c.Len())
	}
}

func TestZeroTTL_NeverExpires(t *testing.T) {
	c := cache.New(0)
	c.Set("k", "v")
	time.Sleep(5 * time.Millisecond)
	_, err := c.Get("k")
	if err != nil {
		t.Fatalf("expected no expiry with zero TTL, got %v", err)
	}
}
