package routes

import (
	"testing"
	"time"
)

func sampleSuppressDiff() Diff {
	return Diff{
		Added: []Route{
			{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0"},
		},
		Removed: []Route{},
	}
}

func TestDefaultSuppressConfig(t *testing.T) {
	cfg := DefaultSuppressConfig()
	if cfg.Window <= 0 {
		t.Error("expected positive Window")
	}
	if cfg.MaxSuppressed <= 0 {
		t.Error("expected positive MaxSuppressed")
	}
}

func TestSuppressor_FirstCallNotSuppressed(t *testing.T) {
	s := NewSuppressor(DefaultSuppressConfig())
	d := sampleSuppressDiff()
	if s.IsSuppressed(d) {
		t.Error("expected first call to not be suppressed")
	}
}

func TestSuppressor_SecondCallWithinWindowSuppressed(t *testing.T) {
	s := NewSuppressor(DefaultSuppressConfig())
	d := sampleSuppressDiff()
	s.IsSuppressed(d)
	if !s.IsSuppressed(d) {
		t.Error("expected second call within window to be suppressed")
	}
}

func TestSuppressor_DifferentDiffNotSuppressed(t *testing.T) {
	s := NewSuppressor(DefaultSuppressConfig())
	d1 := sampleSuppressDiff()
	d2 := Diff{
		Added: []Route{
			{Destination: "172.16.0.0/12", Gateway: "10.0.0.1", Iface: "eth1"},
		},
		Removed: []Route{},
	}
	s.IsSuppressed(d1)
	if s.IsSuppressed(d2) {
		t.Error("expected different diff to not be suppressed")
	}
}

func TestSuppressor_WindowExpiredAllowsAgain(t *testing.T) {
	cfg := SuppressConfig{Window: 10 * time.Millisecond, MaxSuppressed: 100}
	s := NewSuppressor(cfg)
	d := sampleSuppressDiff()
	s.IsSuppressed(d)
	time.Sleep(20 * time.Millisecond)
	if s.IsSuppressed(d) {
		t.Error("expected diff to be allowed after window expires")
	}
}

func TestSuppressor_EmptyDiffAlwaysSuppressed(t *testing.T) {
	s := NewSuppressor(DefaultSuppressConfig())
	d := Diff{}
	if !s.IsSuppressed(d) {
		t.Error("expected empty diff to always be suppressed")
	}
}

func TestSuppressor_StatsAndReset(t *testing.T) {
	s := NewSuppressor(DefaultSuppressConfig())
	s.IsSuppressed(sampleSuppressDiff())
	if s.Stats() != 1 {
		t.Errorf("expected 1 tracked fingerprint, got %d", s.Stats())
	}
	s.Reset()
	if s.Stats() != 0 {
		t.Errorf("expected 0 after reset, got %d", s.Stats())
	}
}

func TestSuppressor_MaxSuppressedEvictsOldest(t *testing.T) {
	cfg := SuppressConfig{Window: time.Minute, MaxSuppressed: 2}
	s := NewSuppressor(cfg)

	d1 := Diff{Added: []Route{{Destination: "1.0.0.0/8"}}}
	d2 := Diff{Added: []Route{{Destination: "2.0.0.0/8"}}}
	d3 := Diff{Added: []Route{{Destination: "3.0.0.0/8"}}}

	s.IsSuppressed(d1)
	s.IsSuppressed(d2)
	s.IsSuppressed(d3) // should evict d1

	if s.Stats() != 2 {
		t.Errorf("expected 2 tracked entries after eviction, got %d", s.Stats())
	}
}
