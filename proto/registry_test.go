package proto

import (
	"testing"
)

func TestRegister_Valid(t *testing.T) {
	r := NewRegistry()
	err := r.Register(MethodInfo{FullMethod: "/svc/Method", InputType: "Req", OutputType: "Resp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegister_EmptyMethod(t *testing.T) {
	r := NewRegistry()
	err := r.Register(MethodInfo{})
	if err == nil {
		t.Fatal("expected error for empty FullMethod")
	}
}

func TestLookup_Found(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(MethodInfo{FullMethod: "/svc/Hello"})
	info, ok := r.Lookup("/svc/Hello")
	if !ok {
		t.Fatal("expected method to be found")
	}
	if info.FullMethod != "/svc/Hello" {
		t.Errorf("got %s, want /svc/Hello", info.FullMethod)
	}
}

func TestLookup_NotFound(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Lookup("/svc/Missing")
	if ok {
		t.Fatal("expected method not to be found")
	}
}

func TestList_ReturnsAll(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(MethodInfo{FullMethod: "/svc/A"})
	_ = r.Register(MethodInfo{FullMethod: "/svc/B"})
	names := r.List()
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
	}
}

func TestRegister_Concurrent(t *testing.T) {
	r := NewRegistry()
	done := make(chan struct{})
	for i := 0; i < 20; i++ {
		go func(n int) {
			_ = r.Register(MethodInfo{FullMethod: fmt.Sprintf("/svc/M%d", n)})
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < 20; i++ {
		<-done
	}
	if len(r.List()) != 20 {
		t.Errorf("expected 20 methods, got %d", len(r.List()))
	}
}
