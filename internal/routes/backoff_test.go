package routes

import (
	"testing"
	"time"
)

func TestDefaultBackoffConfig(t *testing.T) {
	cfg := DefaultBackoffConfig()
	if cfg.InitialInterval != 500*time.Millisecond {
		t.Errorf("expected 500ms initial interval, got %v", cfg.InitialInterval)
	}
	if cfg.MaxInterval != 30*time.Second {
		t.Errorf("expected 30s max interval, got %v", cfg.MaxInterval)
	}
	if cfg.Multiplier != 2.0 {
		t.Errorf("expected multiplier 2.0, got %v", cfg.Multiplier)
	}
	if cfg.MaxAttempts != 8 {
		t.Errorf("expected 8 max attempts, got %d", cfg.MaxAttempts)
	}
}

func TestBackoff_FirstAttemptUsesInitialInterval(t *testing.T) {
	b := NewBackoff(BackoffConfig{
		InitialInterval: 100 * time.Millisecond,
		MaxInterval:     10 * time.Second,
		Multiplier:      2.0,
		MaxAttempts:     5,
	})
	d, ok := b.Next()
	if !ok {
		t.Fatal("expected allowed=true on first attempt")
	}
	if d != 100*time.Millisecond {
		t.Errorf("expected 100ms, got %v", d)
	}
}

func TestBackoff_IntervalsGrowExponentially(t *testing.T) {
	b := NewBackoff(BackoffConfig{
		InitialInterval: 100 * time.Millisecond,
		MaxInterval:     10 * time.Second,
		Multiplier:      2.0,
		MaxAttempts:     10,
	})
	expected := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		400 * time.Millisecond,
		800 * time.Millisecond,
	}
	for i, want := range expected {
		d, ok := b.Next()
		if !ok {
			t.Fatalf("attempt %d: expected allowed", i)
		}
		if d != want {
			t.Errorf("attempt %d: expected %v, got %v", i, want, d)
		}
	}
}

func TestBackoff_CapsAtMaxInterval(t *testing.T) {
	b := NewBackoff(BackoffConfig{
		InitialInterval: 1 * time.Second,
		MaxInterval:     3 * time.Second,
		Multiplier:      2.0,
		MaxAttempts:     10,
	})
	for i := 0; i < 6; i++ {
		d, _ := b.Next()
		if d > 3*time.Second {
			t.Errorf("attempt %d: interval %v exceeded max", i, d)
		}
	}
}

func TestBackoff_BlocksAfterMaxAttempts(t *testing.T) {
	b := NewBackoff(BackoffConfig{
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     1 * time.Second,
		Multiplier:      2.0,
		MaxAttempts:     3,
	})
	for i := 0; i < 3; i++ {
		_, ok := b.Next()
		if !ok {
			t.Fatalf("attempt %d should be allowed", i)
		}
	}
	_, ok := b.Next()
	if ok {
		t.Error("expected allowed=false after MaxAttempts")
	}
}

func TestBackoff_ResetRestoresAttempts(t *testing.T) {
	b := NewBackoff(BackoffConfig{
		InitialInterval: 50 * time.Millisecond,
		MaxInterval:     1 * time.Second,
		Multiplier:      2.0,
		MaxAttempts:     2,
	})
	b.Next()
	b.Next()
	if _, ok := b.Next(); ok {
		t.Fatal("expected blocked before reset")
	}
	b.Reset()
	if b.Attempts() != 0 {
		t.Errorf("expected 0 attempts after reset, got %d", b.Attempts())
	}
	if _, ok := b.Next(); !ok {
		t.Error("expected allowed after reset")
	}
}
