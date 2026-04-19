package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type rawConfig struct {
	Address     string `json:"address"`
	ProtoFile   string `json:"proto_file"`
	Service     string `json:"service"`
	Method      string `json:"method"`
	Data        string `json:"data"`
	Concurrency int    `json:"concurrency"`
	Requests    int    `json:"requests"`
	TimeoutSec  int    `json:"timeout_seconds"`
	Insecure    bool   `json:"insecure"`
}

// LoadFromFile reads a JSON config file and returns a validated Config.
func LoadFromFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening config file: %w", err)
	}
	defer f.Close()

	var raw rawConfig
	if err := json.NewDecoder(f).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decoding config file: %w", err)
	}

	cfg := DefaultConfig()
	if raw.Address != "" {
		cfg.Address = raw.Address
	}
	if raw.ProtoFile != "" {
		cfg.ProtoFile = raw.ProtoFile
	}
	if raw.Service != "" {
		cfg.Service = raw.Service
	}
	if raw.Method != "" {
		cfg.Method = raw.Method
	}
	if raw.Data != "" {
		cfg.Data = raw.Data
	}
	if raw.Concurrency > 0 {
		cfg.Concurrency = raw.Concurrency
	}
	if raw.Requests > 0 {
		cfg.Requests = raw.Requests
	}
	if raw.TimeoutSec > 0 {
		cfg.Timeout = time.Duration(raw.TimeoutSec) * time.Second
	}
	cfg.Insecure = raw.Insecure

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return cfg, nil
}
