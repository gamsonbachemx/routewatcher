package routes

import (
	"testing"
)

func sampleSnapshot() Snapshot {
	return Snapshot{
		{Destination: "0.0.0.0/0", Gateway: "192.168.1.1", Iface: "eth0", Protocol: "static", Metric: 100},
		{Destination: "127.0.0.0/8", Gateway: "", Iface: "lo", Protocol: "kernel", Metric: 0},
		{Destination: "192.168.1.0/24", Gateway: "", Iface: "eth0", Protocol: "kernel", Metric: 0},
		{Destination: "10.0.0.0/8", Gateway: "10.0.0.1", Iface: "wg0", Protocol: "static", Metric: 50},
	}
}

func TestFilter_NilReturnsAll(t *testing.T) {
	snap := sampleSnapshot()
	var f *Filter
	result := f.Apply(snap)
	if len(result) != len(snap) {
		t.Fatalf("expected %d routes, got %d", len(snap), len(result))
	}
}

func TestFilter_ExcludeLocal(t *testing.T) {
	f := &Filter{ExcludeLocal: true}
	result := f.Apply(sampleSnapshot())
	for _, r := range result {
		if isLocal(r) {
			t.Errorf("local route not excluded: %+v", r)
		}
	}
}

func TestFilter_ByInterface(t *testing.T) {
	f := &Filter{Interfaces: []string{"eth0"}}
	result := f.Apply(sampleSnapshot())
	if len(result) != 2 {
		t.Fatalf("expected 2 routes for eth0, got %d", len(result))
	}
	for _, r := range result {
		if r.Iface != "eth0" {
			t.Errorf("unexpected iface %q", r.Iface)
		}
	}
}

func TestFilter_ByProtocol(t *testing.T) {
	f := &Filter{Protocols: []string{"static"}}
	result := f.Apply(sampleSnapshot())
	if len(result) != 2 {
		t.Fatalf("expected 2 static routes, got %d", len(result))
	}
}

func TestFilter_Combined(t *testing.T) {
	f := &Filter{
		ExcludeLocal: true,
		Interfaces:   []string{"eth0", "wg0"},
		Protocols:    []string{"static"},
	}
	result := f.Apply(sampleSnapshot())
	if len(result) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(result))
	}
}

func TestFilter_EmptySnapshot(t *testing.T) {
	f := &Filter{ExcludeLocal: true}
	result := f.Apply(Snapshot{})
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %d", len(result))
	}
}
