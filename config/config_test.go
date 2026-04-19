package config

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Concurrency != 10 {
		t.Errorf("expected concurrency 10, got %d", cfg.Concurrency)
	}
	if cfg.Requests != 100 {
		t.Errorf("expected requests 100, got %d", cfg.Requests)
	}
	if cfg.Timeout != 5*time.Second {
		t.Errorf("expected timeout 5s, got %v", cfg.Timeout)
	}
}

func TestValidate_MissingAddress(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err == nil || err.Error() != "address is required" {
		t.Errorf("expected address error, got %v", err)
	}
}

func TestValidate_Valid(t *testing.T) {
	cfg := &Config{
		Address:     "localhost:50051",
		ProtoFile:   "service.proto",
		Service:     "helloworld.Greeter",
		Method:      "SayHello",
		Concurrency: 5,
		Requests:    50,
		Timeout:     3 * time.Second,
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidate_InvalidConcurrency(t *testing.T) {
	cfg := &Config{
		Address:     "localhost:50051",
		ProtoFile:   "service.proto",
		Service:     "helloworld.Greeter",
		Method:      "SayHello",
		Concurrency: 0,
		Requests:    50,
		Timeout:     3 * time.Second,
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected concurrency error, got nil")
	}
}

func TestValidate_InvalidTimeout(t *testing.T) {
	cfg := &Config{
		Address:     "localhost:50051",
		ProtoFile:   "service.proto",
		Service:     "helloworld.Greeter",
		Method:      "SayHello",
		Concurrency: 5,
		Requests:    50,
		Timeout:     0,
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected timeout error, got nil")
	}
}
