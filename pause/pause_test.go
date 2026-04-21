package pause_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/example/grpcannon/pause"
)

func TestPaused_InitiallyFalse(t *testing.T) {
	c := pause.New()
	if c.Paused() {
		t.Fatal("expected not paused initially")
	}
}

func TestPause_SetsPaused(t *testing.T) {
	c := pause.New()
	c.Pause()
	if !c.Paused() {
		t.Fatal("expected paused after Pause()")
	}
}

func TestResume_ClearsPaused(t *testing.T) {
	c := pause.New()
	c.Pause()
	c.Resume()
	if c.Paused() {
		t.Fatal("expected not paused after Resume()")
	}
}

func TestWait_NotPaused_ReturnsImmediately(t *testing.T) {
	c := pause.New()
	ctx := context.Background()
	if !c.Wait(ctx) {
		t.Fatal("expected Wait to return true when not paused")
	}
}

func TestWait_BlocksUntilResume(t *testing.T) {
	c := pause.New()
	c.Pause()

	var wg sync.WaitGroup
	wg.Add(1)
	result := make(chan bool, 1)
	go func() {
		defer wg.Done()
		result <- c.Wait(context.Background())
	}()

	time.Sleep(20 * time.Millisecond)
	c.Resume()
	wg.Wait()

	if !<-result {
		t.Fatal("expected Wait to return true after Resume")
	}
}

func TestWait_ContextCancelled_ReturnsFalse(t *testing.T) {
	c := pause.New()
	c.Pause()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	if c.Wait(ctx) {
		t.Fatal("expected Wait to return false on context cancellation")
	}
}

func TestClose_UnblocksWaiters(t *testing.T) {
	c := pause.New()
	c.Pause()

	var wg sync.WaitGroup
	wg.Add(1)
	result := make(chan bool, 1)
	go func() {
		defer wg.Done()
		result <- c.Wait(context.Background())
	}()

	time.Sleep(20 * time.Millisecond)
	c.Close()
	wg.Wait()

	if !<-result {
		t.Fatal("expected Wait to return true after Close")
	}
}
