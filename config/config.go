package config

import (
	"errors"
	"time"
)

// Config holds all runtime configuration for grpcannon.
type Config struct {
	Address     string        `json:"address"`
	Concurrency int           `json:"concurrency"`
	Requests    int           `json:"requests"`
	Timeout     time.Duration `json:"timeout"`
	RPS         int           `json:"rps"`          // max requests per second, 0 = unlimited
	OutputJSON  bool          `json:"output_json"`
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Concurrency: 10,
		Requests:    100,
		Timeout:     5 * time.Second,
		RPS:         0,
	}
}

// Validate returns an error if the Config is not usable.
func (c Config) Validate() error {
	if c.Address == "" {
		return errors.New("address is required")
	}
	if c.Concurrency <= 0 {
		return errors.New("concurrency must be greater than 0")
	}
	if c.Timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}
	return nil
}
