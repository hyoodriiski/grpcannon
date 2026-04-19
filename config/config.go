package config

import (
	"errors"
	"time"
)

// Config holds all configuration for a grpcannon run.
type Config struct {
	Address     string        `json:"address"`
	ProtoFile   string        `json:"proto_file"`
	Service     string        `json:"service"`
	Method      string        `json:"method"`
	Data        string        `json:"data"`
	Concurrency int           `json:"concurrency"`
	Requests    int           `json:"requests"`
	Timeout     time.Duration `json:"timeout"`
	Insecure    bool          `json:"insecure"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Concurrency: 10,
		Requests:    100,
		Timeout:     5 * time.Second,
		Insecure:    false,
	}
}

// Validate checks that required fields are set and values are sane.
func (c *Config) Validate() error {
	if c.Address == "" {
		return errors.New("address is required")
	}
	if c.ProtoFile == "" {
		return errors.New("proto_file is required")
	}
	if c.Service == "" {
		return errors.New("service is required")
	}
	if c.Method == "" {
		return errors.New("method is required")
	}
	if c.Concurrency <= 0 {
		return errors.New("concurrency must be greater than 0")
	}
	if c.Requests <= 0 {
		return errors.New("requests must be greater than 0")
	}
	if c.Timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}
	return nil
}
