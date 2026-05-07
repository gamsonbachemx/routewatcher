package routes

import "testing"

func TestTransformer_IntegratesWithFilter(t *testing.T) {
	snap := Snapshot{
		{Destination: "127.0.0.1", Gateway: "", Iface: "LO", Protocol: "kernel", Flags: "U"},
		{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "ETH0", Protocol: "static"},
	}

	// First filter local, then transform
	filtered := Filter(snap, FilterConfig{ExcludeLocal: true})
	tr := NewTransformer(DefaultTransformConfig())
	out := tr.Apply(filtered)

	if len(out) != 1 {
		t.Fatalf("expected 1 route after filter+transform, got %d", len(out))
	}
	if out[0].Iface != "eth0" {
		t.Errorf("expected iface eth0, got %q", out[0].Iface)
	}
}

func TestTransformer_IntegratesWithDiff(t *testing.T) {
	before := Snapshot{
		{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "ETH0", Protocol: "kernel"},
	}
	after := Snapshot{
		{Destination: "10.0.0.0/8", Gateway: "192.168.1.1", Iface: "ETH0", Protocol: "kernel"},
		{Destination: "172.16.0.0/12", Gateway: "192.168.1.1", Iface: "ETH1", Protocol: "static"},
	}

	d := Compare(before, after)
	tr := NewTransformer(DefaultTransformConfig())
	out := tr.ApplyDiff(d)

	if len(out.Added) != 1 {
		t.Fatalf("expected 1 added route, got %d", len(out.Added))
	}
	if out.Added[0].Iface != "eth1" {
		t.Errorf("expected iface eth1 after transform, got %q", out.Added[0].Iface)
	}
}

func TestTransformer_ChainedWithAnnotator(t *testing.T) {
	snap := Snapshot{
		{Destination: "0.0.0.0/0", Gateway: "10.0.0.1", Iface: "ETH0", Protocol: "static"},
	}

	tr := NewTransformer(DefaultTransformConfig())
	transformed := tr.Apply(snap)

	annotCfg := DefaultAnnotateConfig()
	annotCfg.Rules = []AnnotateRule{
		{MatchDestination: "0.0.0.0/0", Note: "default-gw"},
	}
	annotr := NewAnnotator(annotCfg)
	annotated := annotr.Annotate(transformed)

	if annotated[0].Iface != "eth0" {
		t.Errorf("expected eth0 after transform, got %q", annotated[0].Iface)
	}
	if annotated[0].Note != "default-gw" {
		t.Errorf("expected note default-gw, got %q", annotated[0].Note)
	}
}
