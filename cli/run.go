package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/example/grpcannon/config"
	"github.com/example/grpcannon/dialer"
	"github.com/example/grpcannon/metrics"
	"github.com/example/grpcannon/output"
	"github.com/example/grpcannon/worker"
)

// RunOptions holds CLI-level overrides that supplement the config file.
type RunOptions struct {
	ConfigPath string
	Address    string
	Method     string
	Concurrency int
	Requests   int
	Timeout    time.Duration
	JSONOutput bool
}

// Execute loads config, wires dependencies, and runs the load test.
func Execute(opts RunOptions) error {
	cfg, err := loadConfig(opts)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	conn, err := dialer.Connect(cfg.Address, cfg.Timeout)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	collector := metrics.NewCollector()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout*time.Duration(cfg.Requests+1))
	defer cancel()

	pool := worker.NewPool(cfg.Concurrency, func(ctx context.Context) error {
		_, latency, err := invokeMethod(ctx, conn, cfg.Method)
		collector.Record(latency, err)
		return nil
	})

	pool.Run(ctx, cfg.Requests)

	summary := metrics.Summarise(collector)

	if opts.JSONOutput {
		return output.WriteJSONSummary(os.Stdout, summary)
	}
	return output.WriteText(os.Stdout, summary)
}

func loadConfig(opts RunOptions) (*config.Config, error) {
	var cfg *config.Config
	var err error
	if opts.ConfigPath != "" {
		cfg, err = config.LoadFromFile(opts.ConfigPath)
		if err != nil {
			return nil, err
		}
	} else {
		cfg = config.DefaultConfig()
	}
	if opts.Address != "" {
		cfg.Address = opts.Address
	}
	if opts.Method != "" {
		cfg.Method = opts.Method
	}
	if opts.Concurrency > 0 {
		cfg.Concurrency = opts.Concurrency
	}
	if opts.Requests > 0 {
		cfg.Requests = opts.Requests
	}
	if opts.Timeout > 0 {
		cfg.Timeout = opts.Timeout
	}
	return cfg, cfg.Validate()
}
