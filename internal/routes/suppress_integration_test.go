package routes

import (
	"sync"
	"testing"
	"time"
)

func TestSuppressor_ConcurrentAccess(t *testing.T) {
	s := NewSuppressor(DefaultSuppressConfig())
	d := sampleSuppressDiff()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.IsSuppressed(d)
		}()
	}
	wg.Wait()

	if s.Stats() != 1 {
		t.Errorf("expected 1 fingerprint after concurrent access, got %d", s.Stats())
	}
}

func TestSuppressor_IntegratesWithPipeline(t *testing.T) {
	suppressCfg := SuppressConfig{Window: time.Second, MaxSuppressed: 50}
	suppressor := NewSuppressor(suppressCfg)

	pipelineCfg := DefaultPipelineConfig()
	pipeline := NewPipeline(pipelineCfg)

	d := sampleSuppressDiff()

	// Pipeline passes the diff through its own deduplication;
	// suppressor adds an additional layer on top.
	forwarded := 0
	for i := 0; i < 5; i++ {
		pipeline.Process(d, func(out Diff) {
			if !suppressor.IsSuppressed(out) {
				forwarded++
			}
		})
	}

	// Pipeline deduplication blocks repeats; suppressor would also block any
	// that slip through. Only the first should reach the forwarded counter.
	if forwarded != 1 {
		t.Errorf("expected 1 forwarded diff, got %d", forwarded)
	}
}

func TestSuppressor_ResetAllowsReplay(t *testing.T) {
	cfg := SuppressConfig{Window: time.Minute, MaxSuppressed: 100}
	s := NewSuppressor(cfg)
	d := sampleSuppressDiff()

	s.IsSuppressed(d) // record
	if !s.IsSuppressed(d) {
		t.Fatal("expected suppression before reset")
	}

	s.Reset()

	if s.IsSuppressed(d) {
		t.Error("expected diff to pass through after reset")
	}
}
