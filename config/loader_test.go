package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "grpcannon-config-*.json")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoadFromFile_Valid(t *testing.T) {
	path := writeTempConfig(t, `{
		"address": "localhost:50051",
		"proto_file": "svc.proto",
		"service": "pkg.Svc",
		"method": "Call",
		"concurrency": 20,
		"requests": 200,
		"timeout_seconds": 10,
		"insecure": true
	}`)
	defer os.Remove(path)

	cfg, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Concurrency != 20 {
		t.Errorf("expected concurrency 20, got %d", cfg.Concurrency)
	}
	if cfg.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", cfg.Timeout)
	}
	if !cfg.Insecure {
		t.Error("expected insecure true")
	}
}

func TestLoadFromFile_Missing(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadFromFile_InvalidJSON(t *testing.T) {
	path := writeTempConfig(t, `{not valid json}`)
	defer os.Remove(path)
	_, err := LoadFromFile(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadFromFile_FailsValidation(t *testing.T) {
	path := writeTempConfig(t, `{"address": "", "proto_file": "x.proto", "service": "s", "method": "m"}`)
	defer os.Remove(path)
	_, err := LoadFromFile(path)
	if err == nil {
		t.Error("expected validation error")
	}
}
