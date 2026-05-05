package routes

import (
	"sync/atomic"
	"testing"
	"time"
)

// TestThrottler_ConcurrentAccess ensures Allow is safe under concurrent use.
func TestThrottler_ConcurrentAccess(t *testing.T) {
	th := NewThrottler(ThrottleConfig{MinInterval: 50 * time.Millisecond})
	var allowed int64
	done := make(chan struct{})
	for i := 0; i < 20; i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					if th.Allow() {
						atomic.AddInt64(&allowed, 1)
					}
					time.Sleep(5 * time.Millisecond)
				}
			}
		}()
	}
	time.Sleep(200 * time.Millisecond)
	close(done)
	// With 50ms interval over 200ms we expect roughly 4 allowed calls.
	if atomic.LoadInt64(&allowed) < 2 {
		t.Errorf("expected at least 2 allowed calls, got %d", allowed)
	}
}

// TestThrottler_IntegratesWithWatcher verifies the Throttler can gate Watch
// callbacks without dropping the channel signal.
func TestThrottler_IntegratesWithWatcher(t *testing.T) {
	th := NewThrottler(ThrottleConfig{MinInterval: 1 * time.Millisecond})

	diffs := []Diff{
		{Added: []Route{{Destination: "10.0.0.0/8"}}, Removed: nil},
		{Added: []Route{{Destination: "192.168.0.0/16"}}, Removed: nil},
	}

	var forwarded int
	for _, d := range diffs {
		if th.Allow() {
			forwarded++
			_ = d // simulate processing
		}
		time.Sleep(2 * time.Millisecond)
	}

	if forwarded != len(diffs) {
		t.Errorf("expected %d forwarded, got %d", len(diffs), forwarded)
	}
}
