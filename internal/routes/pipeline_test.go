package routes

import (
	"testing"
	"time"
)

func samplePipelineDiff() Diff {
	return Diff{
		Added:   []Route{{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0"}},
		Removed: nil,
	}
}

func TestDefaultPipelineConfig(t *testing.T) {
	cfg := DefaultPipelineConfig()
	if cfg.Throttle.MinInterval != 5*time.Second {
		t.Errorf("unexpected throttle interval: %v", cfg.Throttle.MinInterval)
	}
	if cfg.RateLimit.MaxEvents < 1 {
		t.Errorf("expected positive MaxEvents, got %d", cfg.RateLimit.MaxEvents)
	}
}

func TestPipeline_FirstDiffForwarded(t *testing.T) {
	cfg := DefaultPipelineConfig()
	cfg.Throttle.MinInterval = 0
	p := NewPipeline(cfg)
	if !p.ShouldForward(samplePipelineDiff()) {
		t.Error("first unique diff should be forwarded")
	}
}

func TestPipeline_DuplicateBlocked(t *testing.T) {
	cfg := DefaultPipelineConfig()
	cfg.Throttle.MinInterval = 0
	p := NewPipeline(cfg)
	d := samplePipelineDiff()
	p.ShouldForward(d)
	if p.ShouldForward(d) {
		t.Error("duplicate diff should be blocked by deduplicator")
	}
}

func TestPipeline_ThrottleBlocks(t *testing.T) {
	cfg := DefaultPipelineConfig()
	cfg.Throttle.MinInterval = 10 * time.Second
	cfg.Dedupe.Window = 0 // disable dedupe TTL so second call isn't a dup
	p := NewPipeline(cfg)

	d1 := Diff{Added: []Route{{Destination: "10.0.0.0/8"}}, Removed: nil}
	d2 := Diff{Added: []Route{{Destination: "172.16.0.0/12"}}, Removed: nil}

	p.ShouldForward(d1) // primes throttle
	if p.ShouldForward(d2) {
		t.Error("second diff within throttle window should be blocked")
	}
}

func TestPipeline_Reset(t *testing.T) {
	cfg := DefaultPipelineConfig()
	cfg.Throttle.MinInterval = 0
	p := NewPipeline(cfg)
	d := samplePipelineDiff()
	p.ShouldForward(d)
	p.Reset()
	if !p.ShouldForward(d) {
		t.Error("after reset the same diff should be forwarded again")
	}
}
