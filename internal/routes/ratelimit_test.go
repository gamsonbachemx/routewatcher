package routes

import (
	"testing"
	"time"
)

func TestDefaultRateLimitConfig(t *testing.T) {
	cfg := DefaultRateLimitConfig()
	if cfg.MaxEvents <= 0 {
		t.Errorf("expected positive MaxEvents, got %d", cfg.MaxEvents)
	}
	if cfg.Window <= 0 {
		t.Errorf("expected positive Window, got %v", cfg.Window)
	}
}

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := NewRateLimiter(RateLimitConfig{MaxEvents: 3, Window: time.Second})
	for i := 0; i < 3; i++ {
		if !rl.Allow() {
			t.Errorf("expected Allow() == true on call %d", i+1)
		}
	}
}

func TestRateLimiter_BlocksAtLimit(t *testing.T) {
	rl := NewRateLimiter(RateLimitConfig{MaxEvents: 2, Window: time.Second})
	rl.Allow()
	rl.Allow()
	if rl.Allow() {
		t.Error("expected Allow() == false after limit reached")
	}
}

func TestRateLimiter_Count(t *testing.T) {
	rl := NewRateLimiter(RateLimitConfig{MaxEvents: 5, Window: time.Second})
	rl.Allow()
	rl.Allow()
	if got := rl.Count(); got != 2 {
		t.Errorf("expected Count() == 2, got %d", got)
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := NewRateLimiter(RateLimitConfig{MaxEvents: 2, Window: time.Second})
	rl.Allow()
	rl.Allow()
	rl.Reset()
	if got := rl.Count(); got != 0 {
		t.Errorf("expected Count() == 0 after Reset, got %d", got)
	}
	if !rl.Allow() {
		t.Error("expected Allow() == true after Reset")
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	rl := NewRateLimiter(RateLimitConfig{MaxEvents: 2, Window: 50 * time.Millisecond})
	rl.Allow()
	rl.Allow()
	if rl.Allow() {
		t.Error("expected Allow() == false while at limit")
	}
	time.Sleep(60 * time.Millisecond)
	if !rl.Allow() {
		t.Error("expected Allow() == true after window expired")
	}
}

func TestNewRateLimiter_DefaultsOnZeroValues(t *testing.T) {
	rl := NewRateLimiter(RateLimitConfig{})
	if rl.cfg.MaxEvents != DefaultRateLimitConfig().MaxEvents {
		t.Errorf("expected default MaxEvents %d, got %d", DefaultRateLimitConfig().MaxEvents, rl.cfg.MaxEvents)
	}
	if rl.cfg.Window != DefaultRateLimitConfig().Window {
		t.Errorf("expected default Window %v, got %v", DefaultRateLimitConfig().Window, rl.cfg.Window)
	}
}
