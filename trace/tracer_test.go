package trace_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/grpcannon/trace"
)

func TestStart_RecordsSpan(t *testing.T) {
	tr := trace.New()
	finish := tr.Start("pkg.Svc/Method", nil)
	time.Sleep(2 * time.Millisecond)
	finish(nil)

	spans := tr.Spans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Method != "pkg.Svc/Method" {
		t.Errorf("unexpected method: %s", spans[0].Method)
	}
	if spans[0].Latency < 2*time.Millisecond {
		t.Errorf("latency too small: %v", spans[0].Latency)
	}
	if spans[0].Err != nil {
		t.Errorf("expected nil error, got %v", spans[0].Err)
	}
}

func TestStart_RecordsError(t *testing.T) {
	tr := trace.New()
	sentinel := errors.New("rpc error")
	finish := tr.Start("pkg.Svc/Fail", nil)
	finish(sentinel)

	spans := tr.Spans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if !errors.Is(spans[0].Err, sentinel) {
		t.Errorf("expected sentinel error, got %v", spans[0].Err)
	}
}

func TestStart_RecordsLabels(t *testing.T) {
	tr := trace.New()
	labels := map[string]string{"env": "test", "region": "us-east"}
	finish := tr.Start("pkg.Svc/Method", labels)
	finish(nil)

	spans := tr.Spans()
	if spans[0].Labels["env"] != "test" {
		t.Errorf("missing label env")
	}
}

func TestReset_ClearsSpans(t *testing.T) {
	tr := trace.New()
	tr.Start("m", nil)(nil)
	tr.Reset()
	if len(tr.Spans()) != 0 {
		t.Error("expected empty spans after reset")
	}
}

func TestStart_ConcurrentSafe(t *testing.T) {
	tr := trace.New()
	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			finish := tr.Start("pkg.Svc/M", nil)
			finish(nil)
		}()
	}
	wg.Wait()
	if len(tr.Spans()) != goroutines {
		t.Errorf("expected %d spans, got %d", goroutines, len(tr.Spans()))
	}
}
