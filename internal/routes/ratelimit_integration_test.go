package routes

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestRateLimiter_ConcurrentAccess verifies the rate limiter is safe for
// concurrent use and never allows more events than MaxEvents within the window.
func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	const maxEvents = 20
	rl := NewRateLimiter(RateLimitConfig{
		MaxEvents: maxEvents,
		Window:    time.Second,
	})

	var allowed atomic.Int64
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if rl.Allow() {
				allowed.Add(1)
			}
		}()
	}
	wg.Wait()

	if got := allowed.Load(); got > int64(maxEvents) {
		t.Errorf("allowed %d events, want <= %d", got, maxEvents)
	}
}

// TestRateLimiter_WithWatcher simulates a watcher loop that is gated by a
// rate limiter, ensuring downstream handlers are not called too frequently.
func TestRateLimiter_WithWatcher(t *testing.T) {
	rl := NewRateLimiter(RateLimitConfig{
		MaxEvents: 3,
		Window:    200 * time.Millisecond,
	})

	var processed int
	simulatedEvents := 8

	for i := 0; i < simulatedEvents; i++ {
		if rl.Allow() {
			processed++
		}
	}

	if processed > 3 {
		t.Errorf("processed %d events, expected at most 3 within window", processed)
	}

	// After window expires, events should be allowed again.
	time.Sleep(210 * time.Millisecond)
	if !rl.Allow() {
		t.Error("expected event to be allowed after window reset")
	}
}

// TestRateLimiter_ZeroMaxEvents verifies that a rate limiter configured with
// zero MaxEvents blocks all events immediately.
func TestRateLimiter_ZeroMaxEvents(t *testing.T) {
	rl := NewRateLimiter(RateLimitConfig{
		MaxEvents: 0,
		Window:    time.Second,
	})

	for i := 0; i < 5; i++ {
		if rl.Allow() {
			t.Errorf("event %d was allowed, expected all events to be blocked with MaxEvents=0", i)
		}
	}
}
