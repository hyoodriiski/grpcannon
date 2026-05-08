package mirror_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"grpcannon/mirror"
)

// captureSink records every Result it receives.
type captureSink struct {
	mu      sync.Mutex
	results []mirror.Result
}

func (c *captureSink) Record(r mirror.Result) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.results = append(c.results, r)
}

func (c *captureSink) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.results)
}

func TestNew_NilSecondaryIgnored(t *testing.T) {
	p := &captureSink{}
	m := mirror.New(p, nil, nil)
	if got := m.Len(); got != 1 {
		t.Fatalf("expected 1 sink, got %d", got)
	}
}

func TestRecord_DeliveredToPrimary(t *testing.T) {
	p := &captureSink{}
	m := mirror.New(p)
	m.Record(mirror.Result{Latency: 5 * time.Millisecond})
	if p.Len() != 1 {
		t.Fatalf("primary should have 1 result, got %d", p.Len())
	}
}

func TestRecord_FansOutToSecondary(t *testing.T) {
	p := &captureSink{}
	s1 := &captureSink{}
	s2 := &captureSink{}
	m := mirror.New(p, s1, s2)
	r := mirror.Result{Latency: 10 * time.Millisecond, Err: errors.New("boom")}
	m.Record(r)
	for i, s := range []*captureSink{p, s1, s2} {
		if s.Len() != 1 {
			t.Errorf("sink %d: expected 1 result, got %d", i, s.Len())
		}
	}
}

func TestAdd_NilIgnored(t *testing.T) {
	m := mirror.New(nil)
	m.Add(nil)
	if got := m.Len(); got != 0 {
		t.Fatalf("expected 0 sinks after adding nil, got %d", got)
	}
}

func TestAdd_AppendsSecondary(t *testing.T) {
	p := &captureSink{}
	m := mirror.New(p)
	s := &captureSink{}
	m.Add(s)
	m.Record(mirror.Result{Latency: time.Millisecond})
	if s.Len() != 1 {
		t.Fatalf("added sink should have 1 result, got %d", s.Len())
	}
}

func TestRecord_ConcurrentSafe(t *testing.T) {
	p := &captureSink{}
	s := &captureSink{}
	m := mirror.New(p, s)
	var wg sync.WaitGroup
	const n = 200
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			m.Record(mirror.Result{Latency: time.Microsecond})
		}()
	}
	wg.Wait()
	if p.Len() != n {
		t.Fatalf("primary: expected %d results, got %d", n, p.Len())
	}
	if s.Len() != n {
		t.Fatalf("secondary: expected %d results, got %d", n, s.Len())
	}
}
