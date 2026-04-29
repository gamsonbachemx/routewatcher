package routes

import (
	"sync"
	"testing"
)

func TestMetrics_ConcurrentRecording(t *testing.T) {
	m := NewMetrics(DefaultMetricsConfig())

	const goroutines = 50
	const ops = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < ops; j++ {
				m.RecordPoll()
				m.RecordChange()
			}
		}()
	}
	wg.Wait()

	snap := m.Snapshot()
	expected := int64(goroutines * ops)
	if snap.PollCount != expected {
		t.Errorf("expected PollCount=%d, got %d", expected, snap.PollCount)
	}
	if snap.ChangeCount != expected {
		t.Errorf("expected ChangeCount=%d, got %d", expected, snap.ChangeCount)
	}
}

func TestMetrics_IntegratesWithWatcher(t *testing.T) {
	cfg := DefaultMetricsConfig()
	m := NewMetrics(cfg)

	// Simulate watcher integration: record polls and changes as a watcher would.
	routes1 := []Route{
		{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0", Protocol: "static"},
	}
	routes2 := []Route{
		{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0", Protocol: "static"},
		{Destination: "172.16.0.0/12", Gateway: "192.168.1.254", Iface: "eth0", Protocol: "bgp"},
	}

	prev := Snapshot{Routes: routes1}
	curr := Snapshot{Routes: routes2}

	m.RecordPoll()
	diff := Compare(prev, curr)
	if diff.HasChanges() {
		m.RecordChange()
	}

	snap := m.Snapshot()
	if snap.PollCount != 1 {
		t.Errorf("expected 1 poll, got %d", snap.PollCount)
	}
	if snap.ChangeCount != 1 {
		t.Errorf("expected 1 change, got %d", snap.ChangeCount)
	}
}
