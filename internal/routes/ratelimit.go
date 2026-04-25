package routes

import (
	"sync"
	"time"
)

// RateLimitConfig holds configuration for the rate limiter.
type RateLimitConfig struct {
	// MaxEvents is the maximum number of diff events allowed within Window.
	MaxEvents int
	// Window is the duration over which MaxEvents is measured.
	Window time.Duration
}

// DefaultRateLimitConfig returns a sensible default rate limit config.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		MaxEvents: 10,
		Window:    time.Minute,
	}
}

// RateLimiter tracks event counts over a sliding window and signals when
// the configured threshold is exceeded.
type RateLimiter struct {
	cfg    RateLimitConfig
	mu     sync.Mutex
	events []time.Time
}

// NewRateLimiter creates a new RateLimiter with the given config.
func NewRateLimiter(cfg RateLimitConfig) *RateLimiter {
	if cfg.MaxEvents <= 0 {
		cfg.MaxEvents = DefaultRateLimitConfig().MaxEvents
	}
	if cfg.Window <= 0 {
		cfg.Window = DefaultRateLimitConfig().Window
	}
	return &RateLimiter{cfg: cfg}
}

// Allow records a new event and returns true if the event is within the
// allowed rate, or false if the rate limit has been exceeded.
func (r *RateLimiter) Allow() bool {
	now := time.Now()
	cutoff := now.Add(-r.cfg.Window)

	r.mu.Lock()
	defer r.mu.Unlock()

	// Evict events outside the window.
	valid := r.events[:0]
	for _, t := range r.events {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	r.events = valid

	if len(r.events) >= r.cfg.MaxEvents {
		return false
	}

	r.events = append(r.events, now)
	return true
}

// Count returns the number of events currently within the active window.
func (r *RateLimiter) Count() int {
	now := time.Now()
	cutoff := now.Add(-r.cfg.Window)

	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	for _, t := range r.events {
		if t.After(cutoff) {
			count++
		}
	}
	return count
}

// Reset clears all recorded events.
func (r *RateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = nil
}
