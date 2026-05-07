package routes

import (
	"testing"
)

func sampleLabelDiff() Diff {
	return Diff{
		Added: []Route{
			{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0"},
			{Destination: "172.16.0.0/12", Gateway: "192.168.1.1", Iface: "eth0"},
		},
		Removed: []Route{
			{Destination: "192.168.0.0/16", Gateway: "10.0.0.1", Iface: "eth1"},
		},
	}
}

func TestDefaultLabelConfig(t *testing.T) {
	cfg := DefaultLabelConfig()
	if !cfg.Enabled {
		t.Error("expected Enabled to be true by default")
	}
	if cfg.StaticLabels == nil {
		t.Error("expected StaticLabels to be non-nil")
	}
	if cfg.PrefixRules == nil {
		t.Error("expected PrefixRules to be non-nil")
	}
}

func TestLabeler_StaticLabels(t *testing.T) {
	cfg := DefaultLabelConfig()
	cfg.StaticLabels = map[string]string{"env": "prod", "region": "us-east"}
	l := NewLabeler(cfg)
	d := l.LabelDiff(sampleLabelDiff())
	for _, r := range append(d.Added, d.Removed...) {
		if r.Labels["env"] != "prod" {
			t.Errorf("expected env=prod, got %q", r.Labels["env"])
		}
		if r.Labels["region"] != "us-east" {
			t.Errorf("expected region=us-east, got %q", r.Labels["region"])
		}
	}
}

func TestLabeler_PrefixRule(t *testing.T) {
	cfg := DefaultLabelConfig()
	cfg.PrefixRules = map[string]string{"10.": "private-10", "172.": "private-172"}
	l := NewLabeler(cfg)
	d := l.LabelDiff(sampleLabelDiff())
	if d.Added[0].Labels["prefix"] != "private-10" {
		t.Errorf("expected prefix=private-10, got %q", d.Added[0].Labels["prefix"])
	}
	if d.Added[1].Labels["prefix"] != "private-172" {
		t.Errorf("expected prefix=private-172, got %q", d.Added[1].Labels["prefix"])
	}
}

func TestLabeler_DisabledPassesThrough(t *testing.T) {
	cfg := DefaultLabelConfig()
	cfg.Enabled = false
	cfg.StaticLabels = map[string]string{"env": "prod"}
	l := NewLabeler(cfg)
	d := l.LabelDiff(sampleLabelDiff())
	for _, r := range d.Added {
		if len(r.Labels) != 0 {
			t.Errorf("expected no labels when disabled, got %v", r.Labels)
		}
	}
}

func TestLabeler_PreservesExistingLabels(t *testing.T) {
	cfg := DefaultLabelConfig()
	cfg.StaticLabels = map[string]string{"env": "staging"}
	l := NewLabeler(cfg)
	d := Diff{
		Added: []Route{
			{Destination: "10.0.0.0/8", Labels: map[string]string{"owner": "team-a"}},
		},
	}
	out := l.LabelDiff(d)
	if out.Added[0].Labels["owner"] != "team-a" {
		t.Error("expected existing label owner=team-a to be preserved")
	}
	if out.Added[0].Labels["env"] != "staging" {
		t.Error("expected static label env=staging to be applied")
	}
}
