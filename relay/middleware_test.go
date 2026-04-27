package relay_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/grpcannon/relay"
)

func TestTap_SuccessEmitsResult(t *testing.T) {
	r := relay.New[relay.Result]()
	var received []relay.Result
	r.Register(func(res relay.Result) {
		received = append(received, res)
	})

	invoker := relay.Tap(r, func(ctx context.Context) error {
		return nil
	})

	if err := invoker(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(received) != 1 {
		t.Fatalf("expected 1 result, got %d", len(received))
	}
	if received[0].Err != nil {
		t.Fatalf("expected nil err in result")
	}
}

func TestTap_ErrorEmitsResult(t *testing.T) {
	r := relay.New[relay.Result]()
	var received []relay.Result
	r.Register(func(res relay.Result) {
		received = append(received, res)
	})

	sentinel := errors.New("rpc failed")
	invoker := relay.Tap(r, func(ctx context.Context) error {
		return sentinel
	})

	err := invoker(context.Background())
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if len(received) != 1 {
		t.Fatalf("expected 1 result, got %d", len(received))
	}
	if !errors.Is(received[0].Err, sentinel) {
		t.Fatalf("result should carry sentinel error")
	}
}

func TestTap_MultipleHandlers_AllReceive(t *testing.T) {
	r := relay.New[relay.Result]()
	count := 0
	r.Register(func(relay.Result) { count++ })
	r.Register(func(relay.Result) { count++ })

	invoker := relay.Tap(r, func(ctx context.Context) error { return nil })
	_ = invoker(context.Background())

	if count != 2 {
		t.Fatalf("expected 2, got %d", count)
	}
}
