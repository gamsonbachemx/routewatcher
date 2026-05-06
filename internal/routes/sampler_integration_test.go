package routes

import (
	"sync"
	"testing"
)

func TestSampler_ConcurrentAccess(t *testing.T) {
	s := NewSampler(SamplerConfig{Rate: 0.5, Enabled: true, Seed: 7})
	d := sampleDiffForSampler()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Sample(d)
		}()
	}
	wg.Wait()
}

func TestSampler_IntegratesWithPipeline(t *testing.T) {
	sampler := NewSampler(SamplerConfig{Rate: 1.0, Enabled: true, Seed: 42})
	deduper := NewDeduplicator(DefaultDedupeConfig())

	d := Diff{
		Added:   []Route{{Destination: "172.16.0.0/12", Gateway: "10.0.0.1", Iface: "eth1"}},
		Removed: []Route{},
	}

	// Sampler passes, deduper allows first occurrence
	if !sampler.Sample(d) {
		t.Fatal("sampler should pass at rate 1.0")
	}
	if deduper.IsDuplicate(d) {
		t.Fatal("first diff should not be a duplicate")
	}

	// Sampler still passes, but deduper blocks repeat
	if !sampler.Sample(d) {
		t.Fatal("sampler should still pass at rate 1.0")
	}
	if !deduper.IsDuplicate(d) {
		t.Fatal("second identical diff should be duplicate")
	}
}

func TestSampler_ZeroRateDropsAll(t *testing.T) {
	s := NewSampler(SamplerConfig{Rate: 0.0, Enabled: true, Seed: 0})
	diffs := []Diff{
		sampleDiffForSampler(),
		sampleDiffForSampler(),
	}
	out := s.SampleDiffs(diffs)
	if len(out) != 0 {
		t.Errorf("expected zero diffs from zero-rate sampler, got %d", len(out))
	}
}
