package evict_test

import (
	"fmt"
	"sync"
	"testing"

	"grpcannon/evict"
)

func TestNew_DefaultsCap(t *testing.T) {
	c := evict.New(0, nil)
	if c.Len() != 0 {
		t.Fatalf("expected empty cache")
	}
	// Should accept at least one entry without panicking.
	c.Set("k", 1)
	if c.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", c.Len())
	}
}

func TestSet_And_Get(t *testing.T) {
	c := evict.New(4, nil)
	c.Set("a", 42)
	v, ok := c.Get("a")
	if !ok {
		t.Fatal("expected key to be present")
	}
	if v.(int) != 42 {
		t.Fatalf("expected 42, got %v", v)
	}
}

func TestGet_NotFound(t *testing.T) {
	c := evict.New(4, nil)
	_, ok := c.Get("missing")
	if ok {
		t.Fatal("expected key to be absent")
	}
}

func TestEviction_LRU(t *testing.T) {
	evicted := []string{}
	c := evict.New(3, func(key string, _ any) {
		evicted = append(evicted, key)
	})

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	// Access "a" so "b" becomes LRU.
	c.Get("a")
	// Adding "d" should evict "b".
	c.Set("d", 4)

	if len(evicted) != 1 || evicted[0] != "b" {
		t.Fatalf("expected 'b' evicted, got %v", evicted)
	}
}

func TestSet_UpdateExisting_NoEviction(t *testing.T) {
	evictions := 0
	c := evict.New(2, func(_ string, _ any) { evictions++ })
	c.Set("x", 1)
	c.Set("x", 2) // update, not insert
	if evictions != 0 {
		t.Fatalf("expected no evictions, got %d", evictions)
	}
	v, _ := c.Get("x")
	if v.(int) != 2 {
		t.Fatalf("expected updated value 2, got %v", v)
	}
}

func TestCache_ConcurrentSafe(t *testing.T) {
	c := evict.New(16, nil)
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i%8)
			c.Set(key, i)
			c.Get(key)
		}(i)
	}
	wg.Wait()
}
