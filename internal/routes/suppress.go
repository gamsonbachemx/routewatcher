package routes

import (
	"sync"
	"time"
)

// DefaultSuppressConfig returns a SuppressConfig with sensible defaults.
func DefaultSuppressConfig() SuppressConfig {
	return SuppressConfig{
		Window:    2 * time.Minute,
		MaxSuppressed: 100,
	}
}

// SuppressConfig controls how route change suppression behaves.
type SuppressConfig struct {
	// Window is the duration during which repeated identical diffs are suppressed.
	Window time.Duration
	// MaxSuppressed is the maximum number of distinct fingerprints tracked.
	MaxSuppressed int
}

type suppressEntry struct {
	count     int
	firstSeen time.Time
	lastSeen  time.Time
}

// Suppressor tracks recently seen diffs and suppresses repeats within a window.
type Suppressor struct {
	mu      sync.Mutex
	cfg     SuppressConfig
	seen    map[string]*suppressEntry
}

// NewSuppressor creates a Suppressor with the given config.
func NewSuppressor(cfg SuppressConfig) *Suppressor {
	return &Suppressor{
		cfg:  cfg,
		seen: make(map[string]*suppressEntry),
	}
}

// IsSuppressed returns true if the diff has been seen recently and should be
// suppressed. It always records the diff on first call within the window.
func (s *Suppressor) IsSuppressed(d Diff) bool {
	if len(d.Added) == 0 && len(d.Removed) == 0 {
		return true
	}

	fp := fingerprintDiff(d)
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	if entry, ok := s.seen[fp]; ok {
		if now.Sub(entry.firstSeen) < s.cfg.Window {
			entry.count++
			entry.lastSeen = now
			return true
		}
		// Window expired — reset
		entry.count = 1
		entry.firstSeen = now
		entry.lastSeen = now
		return false
	}

	// Evict oldest entry if at capacity
	if len(s.seen) >= s.cfg.MaxSuppressed {
		var oldest string
		var oldestTime time.Time
		for k, v := range s.seen {
			if oldest == "" || v.firstSeen.Before(oldestTime) {
				oldest = k
				oldestTime = v.firstSeen
			}
		}
		delete(s.seen, oldest)
	}

	s.seen[fp] = &suppressEntry{count: 1, firstSeen: now, lastSeen: now}
	return false
}

// Stats returns the number of currently tracked fingerprints.
func (s *Suppressor) Stats() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.seen)
}

// Reset clears all tracked suppression state.
func (s *Suppressor) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seen = make(map[string]*suppressEntry)
}
