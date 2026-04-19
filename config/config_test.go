package config_test

import (
	"testing"
	"time"

	"github.com/example/grpcannon/config"
)

func TestDefaultConfig(t *testing.T) {
	c := config.DefaultConfig()
	if c.Concurrency != 10 {
		t.Errorf("expected concurrency 10, got %d", c.Concurrency)
	}
	if c.Timeout != 5*time.Second {
		t.Errorf("expected timeout 5s, got %v", c.Timeout)
	}
	if c.RPS != 0 {
		t.Errorf("expected rps 0, got %d", c.RPS)
	}
}

func TestValidate_MissingAddress(t *testing.T) {
	c := config.DefaultConfig()
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for missing address")
	}
}

func TestValidate_Valid(t *testing.T) {
	c := config.DefaultConfig()
	c.Address = "localhost:50051"
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_InvalidConcurrency(t *testing.T) {
	c := config.DefaultConfig()
	c.Address = "localhost:50051"
	c.Concurrency = 0
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for zero concurrency")
	}
}

func TestValidate_InvalidTimeout(t *testing.T) {
	c := config.DefaultConfig()
	c.Address = "localhost:50051"
	c.Timeout = 0
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for zero timeout")
	}
}

func TestValidate_RPSField(t *testing.T) {
	c := config.DefaultConfig()
	c.Address = "localhost:50051"
	c.RPS = 50
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected error with rps set: %v", err)
	}
}
