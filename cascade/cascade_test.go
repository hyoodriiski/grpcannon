package cascade_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"grpcannon/cascade"
)

func successInvoker(latency time.Duration) cascade.Invoker {
	return func(_ context.Context) (time.Duration, error) {
		return latency, nil
	}
}

func errorInvoker(msg string) cascade.Invoker {
	return func(_ context.Context) (time.Duration, error) {
		return 0, errors.New(msg)
	}
}

func TestNew_NilInvokersIgnored(t *testing.T) {
	c := cascade.New(nil, successInvoker(time.Millisecond), nil)
	if c.Len() != 1 {
		t.Fatalf("expected 1 invoker, got %d", c.Len())
	}
}

func TestRun_NoInvokers_ReturnsError(t *testing.T) {
	c := cascade.New()
	_, err := c.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for empty cascade")
	}
}

func TestRun_FirstSucceeds(t *testing.T) {
	c := cascade.New(successInvoker(5*time.Millisecond), errorInvoker("should not reach"))
	res, err := c.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Stage != 0 {
		t.Errorf("expected stage 0, got %d", res.Stage)
	}
	if res.Latency != 5*time.Millisecond {
		t.Errorf("unexpected latency: %v", res.Latency)
	}
}

func TestRun_FallsBackToSecond(t *testing.T) {
	c := cascade.New(errorInvoker("primary down"), successInvoker(2*time.Millisecond))
	res, err := c.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Stage != 1 {
		t.Errorf("expected stage 1, got %d", res.Stage)
	}
}

func TestRun_AllFail_ReturnsLastError(t *testing.T) {
	c := cascade.New(errorInvoker("err1"), errorInvoker("err2"))
	_, err := c.Run(context.Background())
	if err == nil {
		t.Fatal("expected error when all invokers fail")
	}
	if !errors.Is(err, err) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRun_CancelledContext_StopsEarly(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	called := 0
	inv := func(_ context.Context) (time.Duration, error) {
		called++
		return 0, nil
	}
	c := cascade.New(inv, inv)
	_, err := c.Run(ctx)
	if err == nil {
		t.Fatal("expected context error")
	}
	if called != 0 {
		t.Errorf("expected no calls after cancel, got %d", called)
	}
}
