package label

import (
	"strings"
	"sync"
	"testing"
)

func TestNew_Empty(t *testing.T) {
	s := New()
	if s.Len() != 0 {
		t.Fatalf("expected 0 labels, got %d", s.Len())
	}
}

func TestAdd_And_Get(t *testing.T) {
	s := New()
	s.Add("env", "prod")
	v, ok := s.Get("env")
	if !ok {
		t.Fatal("expected key 'env' to exist")
	}
	if v != "prod" {
		t.Fatalf("expected 'prod', got %q", v)
	}
}

func TestAdd_EmptyKeyIgnored(t *testing.T) {
	s := New()
	s.Add("", "value")
	s.Add("   ", "value")
	if s.Len() != 0 {
		t.Fatalf("expected 0 labels after empty-key adds, got %d", s.Len())
	}
}

func TestGet_NotFound(t *testing.T) {
	s := New()
	_, ok := s.Get("missing")
	if ok {
		t.Fatal("expected key not to be found")
	}
}

func TestAdd_Overwrite(t *testing.T) {
	s := New()
	s.Add("region", "us-east")
	s.Add("region", "eu-west")
	v, _ := s.Get("region")
	if v != "eu-west" {
		t.Fatalf("expected 'eu-west', got %q", v)
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	s := New()
	s.Add("a", "1")
	s.Add("b", "2")
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	// mutating copy must not affect original
	delete(all, "a")
	if s.Len() != 2 {
		t.Fatal("original set was mutated through All() copy")
	}
}

func TestString_ContainsKeyValue(t *testing.T) {
	s := New()
	s.Add("method", "SayHello")
	str := s.String()
	if !strings.Contains(str, "method=SayHello") {
		t.Fatalf("expected 'method=SayHello' in %q", str)
	}
}

func TestAdd_ConcurrentSafe(t *testing.T) {
	s := New()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			s.Add(strings.Repeat("k", n+1), "v")
			_ = s.Len()
		}(i)
	}
	wg.Wait()
}
