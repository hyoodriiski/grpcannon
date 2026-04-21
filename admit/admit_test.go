package admit_test

import (
	"context"
	"sync"
	"testing"

	"github.com/grpcannon/admit"
)

func TestNew_ZeroMax_Unlimited(t *testing.T) {
	c := admit.New(0)
	for i := 0; i < 1000; i++ {
		rel, err := c.Admit(context.Background())
		if err != nil {
			t.Fatalf("unexpected rejection at i=%d: %v", i, err)
		}
		rel()
	}
}

func TestAdmit_WithinLimit_Succeeds(t *testing.T) {
	c := admit.New(3)
	var releases []func()
	for i := 0; i < 3; i++ {
		rel, err := c.Admit(context.Background())
		if err != nil {
			t.Fatalf("expected admission, got: %v", err)
		}
		releases = append(releases, rel)
	}
	if got := c.InFlight(); got != 3 {
		t.Fatalf("expected 3 in-flight, got %d", got)
	}
	for _, r := range releases {
		r()
	}
	if got := c.InFlight(); got != 0 {
		t.Fatalf("expected 0 in-flight after release, got %d", got)
	}
}

func TestAdmit_ExceedsLimit_ReturnsRejected(t *testing.T) {
	c := admit.New(2)
	rel1, _ := c.Admit(context.Background())
	rel2, _ := c.Admit(context.Background())
	defer rel1()
	defer rel2()

	_, err := c.Admit(context.Background())
	if err != admit.ErrRejected {
		t.Fatalf("expected ErrRejected, got %v", err)
	}
}

func TestAdmit_ReleaseRestoresCapacity(t *testing.T) {
	c := admit.New(1)
	rel, err := c.Admit(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	rel()

	rel2, err := c.Admit(context.Background())
	if err != nil {
		t.Fatalf("expected admission after release, got: %v", err)
	}
	rel2()
}

func TestAdmit_ConcurrentSafe(t *testing.T) {
	const cap = 50
	c := admit.New(cap)
	var wg sync.WaitGroup
	admitted := make(chan func(), cap*2)

	for i := 0; i < cap*2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rel, err := c.Admit(context.Background())
			if err == nil {
				admitted <- rel
			}
		}()
	}
	wg.Wait()
	close(admitted)
	for r := range admitted {
		r()
	}
	if got := c.InFlight(); got != 0 {
		t.Fatalf("expected 0 in-flight, got %d", got)
	}
}
