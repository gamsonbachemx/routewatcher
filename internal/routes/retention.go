package routes

import (
	"sync"
	"time"
)

// DefaultRetentionConfig returns a RetentionConfig with sensible defaults.
func DefaultRetentionConfig() RetentionConfig {
	return RetentionConfig{
		MaxAge:      24 * time.Hour,
		MaxEntries:  500,
		PurgeEvery:  10 * time.Minute,
	}
}

// RetentionConfig controls how long diff history entries are kept.
type RetentionConfig struct {
	MaxAge     time.Duration
	MaxEntries int
	PurgeEvery time.Duration
}

// RetentionManager prunes old diff entries based on age and count limits.
type RetentionManager struct {
	mu      sync.Mutex
	cfg     RetentionConfig
	entries []retentionEntry
	stop    chan struct{}
}

type retentionEntry struct {
	recordedAt time.Time
	diff       Diff
}

// NewRetentionManager creates a RetentionManager and starts its background purge loop.
func NewRetentionManager(cfg RetentionConfig) *RetentionManager {
	rm := &RetentionManager{
		cfg:  cfg,
		stop: make(chan struct{}),
	}
	go rm.loop()
	return rm
}

// Add records a new diff entry.
func (rm *RetentionManager) Add(d Diff) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.entries = append(rm.entries, retentionEntry{recordedAt: time.Now(), diff: d})
}

// Entries returns a snapshot of currently retained diff entries.
func (rm *RetentionManager) Entries() []Diff {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	out := make([]Diff, len(rm.entries))
	for i, e := range rm.entries {
		out[i] = e.diff
	}
	return out
}

// Purge removes entries that exceed MaxAge or MaxEntries limits.
func (rm *RetentionManager) Purge() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	cutoff := time.Now().Add(-rm.cfg.MaxAge)
	filtered := rm.entries[:0]
	for _, e := range rm.entries {
		if e.recordedAt.After(cutoff) {
			filtered = append(filtered, e)
		}
	}
	if len(filtered) > rm.cfg.MaxEntries {
		filtered = filtered[len(filtered)-rm.cfg.MaxEntries:]
	}
	rm.entries = filtered
}

// Stop halts the background purge loop.
func (rm *RetentionManager) Stop() {
	close(rm.stop)
}

func (rm *RetentionManager) loop() {
	ticker := time.NewTicker(rm.cfg.PurgeEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rm.Purge()
		case <-rm.stop:
			return
		}
	}
}
