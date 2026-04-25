package routes

import (
	"testing"
)

const sampleOutput = `default via 192.168.1.1 dev eth0 proto dhcp metric 100
10.0.0.0/8 via 10.1.0.1 dev tun0 metric 50
192.168.1.0/24 dev eth0 proto kernel scope link src 192.168.1.42
`

func TestParseOutput(t *testing.T) {
	snap, err := parseOutput(sampleOutput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snap.Routes) != 3 {
		t.Fatalf("expected 3 routes, got %d", len(snap.Routes))
	}
	if snap.Routes[0].Destination != "default" {
		t.Errorf("expected default, got %s", snap.Routes[0].Destination)
	}
	if snap.Routes[0].Gateway != "192.168.1.1" {
		t.Errorf("expected gateway 192.168.1.1, got %s", snap.Routes[0].Gateway)
	}
	if snap.Routes[0].Iface != "eth0" {
		t.Errorf("expected iface eth0, got %s", snap.Routes[0].Iface)
	}
}

func TestCompare_AddedAndRemoved(t *testing.T) {
	old := &Snapshot{Routes: []Route{
		{Destination: "10.0.0.0/8", Gateway: "10.1.0.1", Iface: "tun0", Metric: "50"},
		{Destination: "default", Gateway: "192.168.1.1", Iface: "eth0", Metric: "100"},
	}}
	new := &Snapshot{Routes: []Route{
		{Destination: "default", Gateway: "192.168.1.1", Iface: "eth0", Metric: "100"},
		{Destination: "172.16.0.0/12", Gateway: "172.16.0.1", Iface: "eth1", Metric: "200"},
	}}

	diff := Compare(old, new)

	if len(diff.Removed) != 1 || diff.Removed[0].Destination != "10.0.0.0/8" {
		t.Errorf("expected 10.0.0.0/8 to be removed, got %+v", diff.Removed)
	}
	if len(diff.Added) != 1 || diff.Added[0].Destination != "172.16.0.0/12" {
		t.Errorf("expected 172.16.0.0/12 to be added, got %+v", diff.Added)
	}
}

func TestCompare_NoChanges(t *testing.T) {
	routes := []Route{
		{Destination: "default", Gateway: "10.0.0.1", Iface: "eth0", Metric: "100"},
	}
	old := &Snapshot{Routes: routes}
	new := &Snapshot{Routes: routes}

	diff := Compare(old, new)
	if !diff.IsEmpty() {
		t.Errorf("expected empty diff, got added=%v removed=%v", diff.Added, diff.Removed)
	}
}
