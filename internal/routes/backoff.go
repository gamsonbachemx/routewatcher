package routes

import (
	"math"
	"sync"
	"time"
)

// DefaultBackoffConfig returns a BackoffConfig with sensible defaults.
func DefaultBackoffConfig() BackoffConfig {
	return BackoffConfig{
		InitialInterval: 500 * time.Millisecond,
		MaxInterval:     30 * time.Second,
		Multiplier:      2.0,
		MaxAttempts:     8,
	}
}

// BackoffConfig controls exponential backoff behaviour.
type BackoffConfig struct {
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	MaxAttempts     int
}

// Backoff implements an exponential backoff strategy.
type Backoff struct {
	cfg      BackoffConfig
	attempt  int
	mu       sync.Mutex
}

// NewBackoff creates a new Backoff with the provided config.
func NewBackoff(cfg BackoffConfig) *Backoff {
	if cfg.Multiplier <= 1.0 {
		cfg.Multiplier = 2.0
	}
	if cfg.InitialInterval <= 0 {
		cfg.InitialInterval = 500 * time.Millisecond
	}
	if cfg.MaxInterval <= 0 {
		cfg.MaxInterval = 30 * time.Second
	}
	return &Backoff{cfg: cfg}
}

// Next returns the next backoff duration and reports whether another attempt
// is allowed. Once MaxAttempts is reached, allowed is false.
func (b *Backoff) Next() (d time.Duration, allowed bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cfg.MaxAttempts > 0 && b.attempt >= b.cfg.MaxAttempts {
		return 0, false
	}

	interval := float64(b.cfg.InitialInterval) * math.Pow(b.cfg.Multiplier, float64(b.attempt))
	if interval > float64(b.cfg.MaxInterval) {
		interval = float64(b.cfg.MaxInterval)
	}

	b.attempt++
	return time.Duration(interval), true
}

// Reset resets the attempt counter.
func (b *Backoff) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.attempt = 0
}

// Attempts returns the current attempt count.
func (b *Backoff) Attempts() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.attempt
}
