package routes

import (
	"testing"
)

func sampleDiffForSampler() Diff {
	return Diff{
		Added:   []Route{{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0"}},
		Removed: []Route{},
	}
}

func TestDefaultSamplerConfig(t *testing.T) {
	cfg := DefaultSamplerConfig()
	if cfg.Rate != 1.0 {
		t.Errorf("expected default rate 1.0, got %f", cfg.Rate)
	}
	if !cfg.Enabled {
		t.Error("expected default sampler to be enabled")
	}
}

func TestSampler_RateOne_AlwaysPasses(t *testing.T) {
	s := NewSampler(SamplerConfig{Rate: 1.0, Enabled: true, Seed: 42})
	for i := 0; i < 100; i++ {
		if !s.Sample(sampleDiffForSampler()) {
			t.Fatal("expected all diffs to pass at rate 1.0")
		}
	}
}

func TestSampler_RateZero_AlwaysBlocks(t *testing.T) {
	s := NewSampler(SamplerConfig{Rate: 0.0, Enabled: true, Seed: 42})
	for i := 0; i < 100; i++ {
		if s.Sample(sampleDiffForSampler()) {
			t.Fatal("expected all diffs to be blocked at rate 0.0")
		}
	}
}

func TestSampler_Disabled_AlwaysPasses(t *testing.T) {
	s := NewSampler(SamplerConfig{Rate: 0.0, Enabled: false, Seed: 42})
	for i := 0; i < 20; i++ {
		if !s.Sample(sampleDiffForSampler()) {
			t.Fatal("expected disabled sampler to always pass")
		}
	}
}

func TestSampler_PartialRate_Approximate(t *testing.T) {
	s := NewSampler(SamplerConfig{Rate: 0.5, Enabled: true, Seed: 99})
	passed := 0
	total := 1000
	for i := 0; i < total; i++ {
		if s.Sample(sampleDiffForSampler()) {
			passed++
		}
	}
	// Expect roughly 50% ± 10%
	if passed < 350 || passed > 650 {
		t.Errorf("expected ~50%% pass rate, got %d/%d", passed, total)
	}
}

func TestSampler_SampleDiffs_FiltersSlice(t *testing.T) {
	s := NewSampler(SamplerConfig{Rate: 1.0, Enabled: true, Seed: 1})
	diffs := []Diff{sampleDiffForSampler(), sampleDiffForSampler(), sampleDiffForSampler()}
	out := s.SampleDiffs(diffs)
	if len(out) != 3 {
		t.Errorf("expected 3 diffs at rate 1.0, got %d", len(out))
	}

	s2 := NewSampler(SamplerConfig{Rate: 0.0, Enabled: true, Seed: 1})
	out2 := s2.SampleDiffs(diffs)
	if len(out2) != 0 {
		t.Errorf("expected 0 diffs at rate 0.0, got %d", len(out2))
	}
}
