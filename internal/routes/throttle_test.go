package routes

import (
	"testing"
	"time"
)

func TestDefaultThrottleConfig(t *testing.T) {
	cfg := DefaultThrottleConfig()
	if cfg.MinInterval != 5*time.Second {
		t.Errorf("expected 5s, got %v", cfg.MinInterval)
	}
}

func TestThrottler_FirstCallAllowed(t *testing.T) {
	th := NewThrottler(DefaultThrottleConfig())
	if !th.Allow() {
		t.Error("first call should be allowed")
	}
}

func TestThrottler_SecondCallWithinIntervalBlocked(t *testing.T) {
	th := NewThrottler(ThrottleConfig{MinInterval: 10 * time.Second})
	th.Allow() // first — allowed
	if th.Allow() {
		t.Error("second call within interval should be blocked")
	}
}

func TestThrottler_SkippedCount(t *testing.T) {
	th := NewThrottler(ThrottleConfig{MinInterval: 10 * time.Second})
	th.Allow()
	th.Allow()
	th.Allow()
	if th.Skipped() != 2 {
		t.Errorf("expected 2 skipped, got %d", th.Skipped())
	}
}

func TestThrottler_AllowedAfterInterval(t *testing.T) {
	th := NewThrottler(ThrottleConfig{MinInterval: 10 * time.Millisecond})
	th.Allow()
	time.Sleep(20 * time.Millisecond)
	if !th.Allow() {
		t.Error("call after interval should be allowed")
	}
}

func TestThrottler_Reset(t *testing.T) {
	th := NewThrottler(ThrottleConfig{MinInterval: 10 * time.Second})
	th.Allow()
	th.Allow()
	th.Reset()
	if th.Skipped() != 0 {
		t.Errorf("expected 0 after reset, got %d", th.Skipped())
	}
	if !th.Allow() {
		t.Error("first call after reset should be allowed")
	}
}
