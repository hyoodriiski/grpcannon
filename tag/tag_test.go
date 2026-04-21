package tag_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/example/grpcannon/tag"
)

func TestNew_Empty(t *testing.T) {
	b := tag.New()
	if b.Len() != 0 {
		t.Fatalf("expected 0 tags, got %d", b.Len())
	}
}

func TestSet_And_Get(t *testing.T) {
	b := tag.New()
	b.Set("env", "prod")
	v, ok := b.Get("env")
	if !ok {
		t.Fatal("expected key to be found")
	}
	if v != "prod" {
		t.Fatalf("expected prod, got %s", v)
	}
}

func TestSet_EmptyKeyIgnored(t *testing.T) {
	b := tag.New()
	b.Set("", "value")
	if b.Len() != 0 {
		t.Fatal("expected empty key to be ignored")
	}
}

func TestGet_NotFound(t *testing.T) {
	b := tag.New()
	_, ok := b.Get("missing")
	if ok {
		t.Fatal("expected key not to be found")
	}
}

func TestSet_Overwrite(t *testing.T) {
	b := tag.New()
	b.Set("region", "us-east")
	b.Set("region", "eu-west")
	v, _ := b.Get("region")
	if v != "eu-west" {
		t.Fatalf("expected eu-west, got %s", v)
	}
}

func TestDelete(t *testing.T) {
	b := tag.New()
	b.Set("k", "v")
	b.Delete("k")
	if b.Len() != 0 {
		t.Fatal("expected tag to be deleted")
	}
}

func TestSnapshot_Immutable(t *testing.T) {
	b := tag.New()
	b.Set("a", "1")
	snap := b.Snapshot()
	b.Set("b", "2")
	if _, ok := snap["b"]; ok {
		t.Fatal("snapshot should not reflect later mutations")
	}
}

func TestTags_String_Deterministic(t *testing.T) {
	tags := tag.Tags{"z": "last", "a": "first", "m": "mid"}
	s := tags.String()
	expected := "a=first,m=mid,z=last"
	if s != expected {
		t.Fatalf("expected %q, got %q", expected, s)
	}
}

func TestSet_ConcurrentSafe(t *testing.T) {
	b := tag.New()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			b.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("val%d", i))
			_, _ = b.Get(fmt.Sprintf("key%d", i))
		}(i)
	}
	wg.Wait()
	if b.Len() == 0 {
		t.Fatal("expected tags to be recorded")
	}
}
