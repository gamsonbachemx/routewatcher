package routes

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

// DedupeConfig holds configuration for the deduplicator.
type DedupeConfig struct {
	// TTL is how long a seen diff fingerprint is remembered.
	TTL time.Duration
}

// DefaultDedupeConfig returns sensible defaults.
func DefaultDedupeConfig() DedupeConfig {
	return DedupeConfig{
		TTL: 5 * time.Minute,
	}
}

type dedupEntry struct {
	expiry time.Time
}

// Deduplicator suppresses repeated identical diffs within a TTL window.
type Deduplicator struct {
	mu     sync.Mutex
	seen   map[string]dedupEntry
	config DedupeConfig
}

// NewDeduplicator creates a Deduplicator with the given config.
func NewDeduplicator(cfg DedupeConfig) *Deduplicator {
	return &Deduplicator{
		seen:   make(map[string]dedupEntry),
		config: cfg,
	}
}

// IsDuplicate returns true if an identical diff was seen within the TTL.
// If not a duplicate, the fingerprint is recorded.
func (d *Deduplicator) IsDuplicate(diff Diff) bool {
	key := fingerprintDiff(diff)
	now := time.Now()

	d.mu.Lock()
	defer d.mu.Unlock()

	// Evict expired entries lazily.
	for k, e := range d.seen {
		if now.After(e.expiry) {
			delete(d.seen, k)
		}
	}

	if e, ok := d.seen[key]; ok && now.Before(e.expiry) {
		return true
	}

	d.seen[key] = dedupEntry{expiry: now.Add(d.config.TTL)}
	return false
}

// Reset clears all remembered fingerprints.
func (d *Deduplicator) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.seen = make(map[string]dedupEntry)
}

// fingerprintDiff produces a stable hash string for a Diff.
func fingerprintDiff(diff Diff) string {
	h := sha256.New()
	for _, r := range diff.Added {
		fmt.Fprintf(h, "A:%s:%s:%s\n", r.Destination, r.Gateway, r.Iface)
	}
	for _, r := range diff.Removed {
		fmt.Fprintf(h, "R:%s:%s:%s\n", r.Destination, r.Gateway, r.Iface)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
