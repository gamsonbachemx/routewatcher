package routes

import (
	"math/rand"
	"sync"
	"time"
)

// DefaultSamplerConfig returns a SamplerConfig with sensible defaults.
func DefaultSamplerConfig() SamplerConfig {
	return SamplerConfig{
		Rate:    1.0,
		Seed:    time.Now().UnixNano(),
		Enabled: true,
	}
}

// SamplerConfig controls probabilistic sampling of route diffs.
type SamplerConfig struct {
	// Rate is the fraction of diffs to pass through (0.0–1.0).
	Rate    float64
	Seed    int64
	Enabled bool
}

// Sampler probabilistically forwards diffs based on a configured rate.
type Sampler struct {
	cfg SamplerConfig
	rng *rand.Rand
	mu  sync.Mutex
}

// NewSampler creates a Sampler from the given config.
// A Rate of 1.0 passes all diffs; 0.0 drops all diffs.
func NewSampler(cfg SamplerConfig) *Sampler {
	return &Sampler{
		cfg: cfg,
		rng: rand.New(rand.NewSource(cfg.Seed)), //nolint:gosec
	}
}

// Sample returns true if the diff should be forwarded based on the sampling rate.
func (s *Sampler) Sample(_ Diff) bool {
	if !s.cfg.Enabled {
		return true
	}
	if s.cfg.Rate <= 0.0 {
		return false
	}
	if s.cfg.Rate >= 1.0 {
		return true
	}
	s.mu.Lock()
	v := s.rng.Float64()
	s.mu.Unlock()
	return v < s.cfg.Rate
}

// SampleDiffs filters a slice of diffs according to the sampling rate.
func (s *Sampler) SampleDiffs(diffs []Diff) []Diff {
	out := make([]Diff, 0, len(diffs))
	for _, d := range diffs {
		if s.Sample(d) {
			out = append(out, d)
		}
	}
	return out
}
