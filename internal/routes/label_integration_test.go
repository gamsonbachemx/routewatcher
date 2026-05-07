package routes

import "testing"

func TestLabeler_IntegratesWithDiff(t *testing.T) {
	snap1 := Snapshot{
		Routes: []Route{
			{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0"},
		},
	}
	snap2 := Snapshot{
		Routes: []Route{
			{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0"},
			{Destination: "172.16.0.0/12", Gateway: "10.0.0.1", Iface: "eth1"},
		},
	}
	d := Compare(snap1, snap2)
	if len(d.Added) != 1 {
		t.Fatalf("expected 1 added route, got %d", len(d.Added))
	}

	cfg := DefaultLabelConfig()
	cfg.StaticLabels = map[string]string{"source": "routewatcher"}
	cfg.PrefixRules = map[string]string{"172.": "link-local"}
	l := NewLabeler(cfg)
	labeled := l.LabelDiff(d)

	if labeled.Added[0].Labels["source"] != "routewatcher" {
		t.Errorf("expected source=routewatcher, got %q", labeled.Added[0].Labels["source"])
	}
	if labeled.Added[0].Labels["prefix"] != "link-local" {
		t.Errorf("expected prefix=link-local, got %q", labeled.Added[0].Labels["prefix"])
	}
}

func TestLabeler_IntegratesWithAnnotator(t *testing.T) {
	annotCfg := DefaultAnnotateConfig()
	annotCfg.Rules = map[string]string{"10.0.0.0/8": "internal-network"}
	annotator := NewAnnotator(annotCfg)

	labelCfg := DefaultLabelConfig()
	labelCfg.StaticLabels = map[string]string{"monitored": "true"}
	labeler := NewLabeler(labelCfg)

	d := Diff{
		Added: []Route{
			{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "eth0"},
		},
	}

	annotated := annotator.AnnotateDiff(d)
	labeled := labeler.LabelDiff(annotated)

	if labeled.Added[0].Labels["monitored"] != "true" {
		t.Error("expected monitored=true label")
	}
	if labeled.Added[0].Annotation == "" {
		t.Error("expected annotation to be set by annotator")
	}
}
