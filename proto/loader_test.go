package proto

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempSchema(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "schema*.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString(content)
	_ = f.Close()
	return f.Name()
}

func TestLoadSchema_Valid(t *testing.T) {
	path := writeTempSchema(t, `{"methods":[{"full_method":"/svc/Hello","input_type":"Req","output_type":"Resp"}]}`)
	r := NewRegistry()
	if err := LoadSchema(path, r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := r.Lookup("/svc/Hello"); !ok {
		t.Error("expected /svc/Hello to be registered")
	}
}

func TestLoadSchema_Missing(t *testing.T) {
	r := NewRegistry()
	err := LoadSchema(filepath.Join(t.TempDir(), "no.json"), r)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadSchema_InvalidJSON(t *testing.T) {
	path := writeTempSchema(t, `not json`)
	r := NewRegistry()
	if err := LoadSchema(path, r); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestLoadSchema_NoMethods(t *testing.T) {
	path := writeTempSchema(t, `{"methods":[]}`)
	r := NewRegistry()
	if err := LoadSchema(path, r); err == nil {
		t.Fatal("expected error for empty methods")
	}
}

func TestLoadSchema_MultipleMethod(t *testing.T) {
	path := writeTempSchema(t, `{"methods":[{"full_method":"/svc/A"},{"full_method":"/svc/B"}]}`)
	r := NewRegistry()
	if err := LoadSchema(path, r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.List()) != 2 {
		t.Errorf("expected 2 methods, got %d", len(r.List()))
	}
}
