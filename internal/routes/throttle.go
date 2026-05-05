package routes

import (
	"sync"
	"time"
)

// ThrottleConfig holds configuration for the Throttler.
type ThrottleConfig struct {
	// MinInterval is the minimum duration between forwarded diffs.
	MinInterval time.Duration
}

// DefaultThrottleConfig returns a ThrottleConfig with sensible defaults.
func DefaultThrottleConfig() ThrottleConfig {
	return ThrottleConfig{
		MinInterval: 5 * time.Second,
	}
}

// Throttler suppresses diffs that arrive too soon after the previous one.
type Throttler struct {
	cfg     ThrottleConfig
	mu      sync.Mutex
	lastAt  time.Time
	skipped int
}

// NewThrottler creates a Throttler with the given config.
func NewThrottler(cfg ThrottleConfig) *Throttler {
	return &Throttler{cfg: cfg}
}

// Allow returns true if the diff should be forwarded, false if it should be
// suppressed because it arrived within MinInterval of the last allowed diff.
func (t *Throttler) Allow() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	if t.lastAt.IsZero() || now.Sub(t.lastAt) >= t.cfg.MinInterval {
		t.lastAt = now
		return true
	}
	t.skipped++
	return false
}

// Skipped returns the number of diffs suppressed since the last reset.
func (t *Throttler) Skipped() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.skipped
}

// Reset clears the last-seen timestamp and skipped counter.
func (t *Throttler) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastAt = time.Time{}
	t.skipped = 0
}
