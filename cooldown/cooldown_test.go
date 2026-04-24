package cooldown

import (
	"testing"
	"time"
)

func TestAllow_ZeroWindow_AlwaysAllows(t *testing.T) {
	c := New(0)
	for i := 0; i < 5; i++ {
		if !c.Allow("k") {
			t.Fatal("expected allow with zero window")
		}
	}
}

func TestAllow_FirstCall_Succeeds(t *testing.T) {
	c := New(time.Second)
	if !c.Allow("key") {
		t.Fatal("expected first Allow to succeed")
	}
}

func TestAllow_SecondCall_Blocked(t *testing.T) {
	c := New(time.Second)
	c.Allow("key")
	if c.Allow("key") {
		t.Fatal("expected second Allow within window to be blocked")
	}
}

func TestAllow_AfterWindowExpires_Succeeds(t *testing.T) {
	now := time.Unix(1000, 0)
	c := New(time.Second)
	c.nowFn = func() time.Time { return now }
	c.Allow("key")

	// advance past the window
	c.nowFn = func() time.Time { return now.Add(2 * time.Second) }
	if !c.Allow("key") {
		t.Fatal("expected Allow to succeed after window expires")
	}
}

func TestAllow_DifferentKeys_Independent(t *testing.T) {
	c := New(time.Second)
	c.Allow("a")
	if !c.Allow("b") {
		t.Fatal("expected different key to be allowed")
	}
	if c.Allow("a") {
		t.Fatal("expected same key to be blocked")
	}
}

func TestReset_AllowsImmediately(t *testing.T) {
	c := New(time.Second)
	c.Allow("key")
	c.Reset("key")
	if !c.Allow("key") {
		t.Fatal("expected Allow after Reset to succeed")
	}
}

func TestRemaining_ZeroWindow(t *testing.T) {
	c := New(0)
	if r := c.Remaining("key"); r != 0 {
		t.Fatalf("expected 0, got %v", r)
	}
}

func TestRemaining_NoEntry(t *testing.T) {
	c := New(time.Second)
	if r := c.Remaining("key"); r != 0 {
		t.Fatalf("expected 0 for unknown key, got %v", r)
	}
}

func TestRemaining_WithinWindow(t *testing.T) {
	now := time.Unix(1000, 0)
	c := New(time.Second)
	c.nowFn = func() time.Time { return now }
	c.Allow("key")

	c.nowFn = func() time.Time { return now.Add(200 * time.Millisecond) }
	r := c.Remaining("key")
	if r <= 0 || r > time.Second {
		t.Fatalf("unexpected remaining %v", r)
	}
}
