package routes

import (
	"testing"
)

// TestTagger_IntegratesWithFilter verifies that tagging works correctly
// on a snapshot that has been filtered, simulating real pipeline usage.
func TestTagger_IntegratesWithFilter(t *testing.T) {
	snap := []Route{
		{Destination: "192.168.0.0/16", Gateway: "0.0.0.0", Iface: "eth0", Protocol: "kernel", Flags: "U"},
		{Destination: "10.0.0.0/8", Gateway: "10.0.0.1", Iface: "eth1", Protocol: "static", Flags: "UG"},
		{Destination: "8.8.8.0/24", Gateway: "203.0.113.1", Iface: "eth0", Protocol: "bgp", Flags: "UG"},
	}

	filtered := Filter(snap, FilterOptions{ExcludeLocal: false})
	tagger := NewTagger(DefaultTagConfig())

	tagged := 0
	for _, r := range filtered {
		if tags := tagger.Tag(r); len(tags) > 0 {
			tagged++
		}
	}

	// 192.168.x and 10.x should be tagged as private
	if tagged < 2 {
		t.Errorf("expected at least 2 tagged private routes, got %d", tagged)
	}
}

// TestTagger_IntegratesWithDiff verifies tagging on a real diff output.
func TestTagger_IntegratesWithDiff(t *testing.T) {
	old := []Route{
		{Destination: "10.0.0.0/8", Gateway: "10.0.0.1", Iface: "eth0"},
	}
	new := []Route{
		{Destination: "10.0.0.0/8", Gateway: "10.0.0.1", Iface: "eth0"},
		{Destination: "0.0.0.0/0", Gateway: "203.0.113.1", Iface: "eth0"},
	}

	d := Compare(old, new)
	if len(d.Added) != 1 {
		t.Fatalf("expected 1 added route, got %d", len(d.Added))
	}

	tagger := NewTagger(DefaultTagConfig())
	tagMap := tagger.TagDiff(d)

	tags, ok := tagMap["0.0.0.0/0"]
	if !ok {
		t.Fatal("expected tag entry for added default route")
	}
	if !containsStr(tags, "default") {
		t.Errorf("expected 'default' tag, got %v", tags)
	}
}
