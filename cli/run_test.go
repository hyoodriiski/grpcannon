package cli

import (
	"testing"
	"time"
)

func TestLoadConfig_Defaults(t *testing.T) {
	opts := RunOptions{
		Address: "localhost:50051",
		Method:  "pkg.Service/Method",
	}
	cfg, err := loadConfig(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Address != "localhost:50051" {
		t.Errorf("expected address override, got %s", cfg.Address)
	}
	if cfg.Method != "pkg.Service/Method" {
		t.Errorf("expected method override, got %s", cfg.Method)
	}
}

func TestLoadConfig_Overrides(t *testing.T) {
	opts := RunOptions{
		Address:     "host:9090",
		Method:      "svc/Call",
		Concurrency: 8,
		Requests:    50,
		Timeout:     3 * time.Second,
	}
	cfg, err := loadConfig(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Concurrency != 8 {
		t.Errorf("expected concurrency 8, got %d", cfg.Concurrency)
	}
	if cfg.Requests != 50 {
		t.Errorf("expected requests 50, got %d", cfg.Requests)
	}
	if cfg.Timeout != 3*time.Second {
		t.Errorf("expected timeout 3s, got %v", cfg.Timeout)
	}
}

func TestLoadConfig_MissingAddress(t *testing.T) {
	opts := RunOptions{Method: "svc/Call"}
	_, err := loadConfig(opts)
	if err == nil {
		t.Fatal("expected validation error for missing address")
	}
}

func TestLoadConfig_MissingMethod(t *testing.T) {
	opts := RunOptions{Address: "localhost:50051"}
	_, err := loadConfig(opts)
	if err == nil {
		t.Fatal("expected validation error for missing method")
	}
}
