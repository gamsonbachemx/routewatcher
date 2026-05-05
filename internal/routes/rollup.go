package routes

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// DefaultRollupConfig returns a RollupConfig with sensible defaults.
func DefaultRollupConfig() RollupConfig {
	return RollupConfig{
		Window:   30 * time.Second,
		MaxDiffs: 100,
	}
}

// RollupConfig controls how diffs are accumulated before being emitted.
type RollupConfig struct {
	// Window is the duration over which diffs are rolled up.
	Window time.Duration
	// MaxDiffs is the maximum number of diffs to accumulate before forcing a flush.
	MaxDiffs int
}

// Rollup accumulates route diffs over a time window and emits a merged summary.
type Rollup struct {
	cfg    RollupConfig
	mu     sync.Mutex
	buf    []Diff
	ticker *time.Ticker
	stop   chan struct{}
	out    func([]Diff)
}

// NewRollup creates a Rollup that calls out with accumulated diffs on each flush.
func NewRollup(cfg RollupConfig, out func([]Diff)) *Rollup {
	if out == nil {
		out = func([]Diff) {}
	}
	return &Rollup{
		cfg:  cfg,
		stop: make(chan struct{}),
		out:  out,
	}
}

// Start begins the rollup ticker in the background.
func (r *Rollup) Start() {
	r.ticker = time.NewTicker(r.cfg.Window)
	go func() {
		for {
			select {
			case <-r.ticker.C:
				r.Flush()
			case <-r.stop:
				return
			}
		}
	}()
}

// Stop halts the rollup ticker and performs a final flush.
func (r *Rollup) Stop() {
	close(r.stop)
	if r.ticker != nil {
		r.ticker.Stop()
	}
	r.Flush()
}

// Add appends a diff to the rollup buffer. If MaxDiffs is reached, a flush is triggered.
func (r *Rollup) Add(d Diff) {
	r.mu.Lock()
	r.buf = append(r.buf, d)
	should := r.cfg.MaxDiffs > 0 && len(r.buf) >= r.cfg.MaxDiffs
	r.mu.Unlock()
	if should {
		r.Flush()
	}
}

// Flush emits all buffered diffs and clears the buffer.
func (r *Rollup) Flush() {
	r.mu.Lock()
	if len(r.buf) == 0 {
		r.mu.Unlock()
		return
	}
	batch := make([]Diff, len(r.buf))
	copy(batch, r.buf)
	r.buf = r.buf[:0]
	r.mu.Unlock()
	r.out(batch)
}

// FormatRollup returns a human-readable summary of a batch of diffs.
func FormatRollup(batch []Diff) string {
	if len(batch) == 0 {
		return "rollup: no changes"
	}
	added, removed := 0, 0
	for _, d := range batch {
		added += len(d.Added)
		removed += len(d.Removed)
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("rollup: %d diff(s), +%d added, -%d removed", len(batch), added, removed))
	return sb.String()
}
