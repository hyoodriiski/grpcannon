package tee_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/your-org/grpcannon/tee"
)

type errWriter struct{ err error }

func (e *errWriter) Write(_ []byte) (int, error) { return 0, e.err }

func TestNew_NilTargetsIgnored(t *testing.T) {
	w := tee.New(nil, nil)
	if w.Len() != 0 {
		t.Fatalf("expected 0 targets, got %d", w.Len())
	}
}

func TestWrite_FansOutToAll(t *testing.T) {
	var a, b bytes.Buffer
	w := tee.New(&a, &b)

	_, err := io.WriteString(w, "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.String() != "hello" {
		t.Errorf("a: got %q, want %q", a.String(), "hello")
	}
	if b.String() != "hello" {
		t.Errorf("b: got %q, want %q", b.String(), "hello")
	}
}

func TestWrite_ReturnsFirstError(t *testing.T) {
	expected := errors.New("disk full")
	w := tee.New(&errWriter{err: expected})

	_, err := io.WriteString(w, "data")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "disk full") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestAdd_NilIgnored(t *testing.T) {
	w := tee.New()
	w.Add(nil)
	if w.Len() != 0 {
		t.Fatalf("expected 0 targets after adding nil, got %d", w.Len())
	}
}

func TestAdd_AppendsTarget(t *testing.T) {
	var buf bytes.Buffer
	w := tee.New()
	w.Add(&buf)

	if w.Len() != 1 {
		t.Fatalf("expected 1 target, got %d", w.Len())
	}
	io.WriteString(w, "ok")
	if buf.String() != "ok" {
		t.Errorf("got %q, want %q", buf.String(), "ok")
	}
}

func TestWrite_ConcurrentSafe(t *testing.T) {
	var buf bytes.Buffer
	w := tee.New(&buf)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			io.WriteString(w, "x")
		}()
	}
	wg.Wait()

	if buf.Len() != 50 {
		t.Errorf("expected 50 bytes, got %d", buf.Len())
	}
}
