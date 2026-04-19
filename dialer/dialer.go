package dialer

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Options holds connection options for the gRPC dialer.
type Options struct {
	Address    string
	Timeout    time.Duration
	TLSEnabled bool
	TLSConfig  *tls.Config
}

// Connect establishes a gRPC client connection using the provided options.
func Connect(ctx context.Context, opts Options) (*grpc.ClientConn, error) {
	if opts.Address == "" {
		return nil, fmt.Errorf("dialer: address must not be empty")
	}

	if opts.Timeout <= 0 {
		opts.Timeout = 10 * time.Second
	}

	dialCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	var creds grpc.DialOption
	if opts.TLSEnabled {
		tlsCfg := opts.TLSConfig
		if tlsCfg == nil {
			tlsCfg = &tls.Config{MinVersion: tls.VersionTLS12}
		}
		creds = grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg))
	} else {
		creds = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	conn, err := grpc.DialContext(dialCtx, opts.Address, creds, grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("dialer: failed to connect to %s: %w", opts.Address, err)
	}

	return conn, nil
}
