package cache

import (
	"context"
	"time"
)

// InvokerFn is the signature used by grpcannon invokers.
type InvokerFn func(ctx context.Context, method string, payload interface{}) (interface{}, error)

// Guard wraps an InvokerFn with a cache layer. Responses are cached by method
// for the given TTL. Errors are never cached.
func Guard(c *Cache, ttl time.Duration, next InvokerFn) InvokerFn {
	return func(ctx context.Context, method string, payload interface{}) (interface{}, error) {
		if v, err := c.Get(method); err == nil {
			return v, nil
		}
		resp, err := next(ctx, method, payload)
		if err != nil {
			return nil, err
		}
		c.SetTTL(method, resp, ttl)
		return resp, nil
	}
}
