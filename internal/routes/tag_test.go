package routes

import (
	"testing"
)

func sampleTagRoute(dest string) Route {
	return Route{Destination: dest, Gateway: "via 10.0.0.1", Iface: "eth0"}
}

func TestDefaultTagConfig(t *testing.T) {
	cfg := DefaultTagConfig()
	if len(cfg.Rules) == 0 {
		t.Fatal("expected non-empty default rules")
	}
	if _, ok := cfg.Rules["private"]; !ok {
		t.Error("expected 'private' tag in defaults")
	}
	if _, ok := cfg.Rules["default"]; !ok {
		t.Error("expected 'default' tag in defaults")
	}
}

func TestTagger_PrivateRoute(t *testing.T) {
	tagger := NewTagger(DefaultTagConfig())
	r := sampleTagRoute("192.168.1.0/24")
	tags := tagger.Tag(r)
	if !containsStr(tags, "private") {
		t.Errorf("expected 'private' tag for %s, got %v", r.Destination, tags)
	}
}

func TestTagger_DefaultRoute(t *testing.T) {
	tagger := NewTagger(DefaultTagConfig())
	r := sampleTagRoute("0.0.0.0/0")
	tags := tagger.Tag(r)
	if !containsStr(tags, "default") {
		t.Errorf("expected 'default' tag for %s, got %v", r.Destination, tags)
	}
}

func TestTagger_NoMatchReturnsEmpty(t *testing.T) {
	tagger := NewTagger(DefaultTagConfig())
	r := sampleTagRoute("8.8.8.0/24")
	tags := tagger.Tag(r)
	if len(tags) != 0 {
		t.Errorf("expected no tags for public route, got %v", tags)
	}
}

func TestTagger_CustomRule(t *testing.T) {
	cfg := TagConfig{
		Rules: map[string][]string{
			"corp": {"172.31."},
		},
	}
	tagger := NewTagger(cfg)
	r := sampleTagRoute("172.31.5.0/24")
	tags := tagger.Tag(r)
	if !containsStr(tags, "corp") {
		t.Errorf("expected 'corp' tag, got %v", tags)
	}
}

func TestTagger_TagDiff(t *testing.T) {
	tagger := NewTagger(DefaultTagConfig())
	d := Diff{
		Added:   []Route{sampleTagRoute("10.0.0.0/8")},
		Removed: []Route{sampleTagRoute("0.0.0.0/0")},
	}
	result := tagger.TagDiff(d)
	if tags, ok := result["10.0.0.0/8"]; !ok || !containsStr(tags, "private") {
		t.Errorf("expected 'private' tag for added route, got %v", result)
	}
	if tags, ok := result["0.0.0.0/0"]; !ok || !containsStr(tags, "default") {
		t.Errorf("expected 'default' tag for removed route, got %v", result)
	}
}

func TestTagger_TagDiff_NoMatches(t *testing.T) {
	tagger := NewTagger(DefaultTagConfig())
	d := Diff{
		Added:   []Route{sampleTagRoute("8.8.8.0/24")},
		Removed: []Route{},
	}
	result := tagger.TagDiff(d)
	if len(result) != 0 {
		t.Errorf("expected empty tag map for public routes, got %v", result)
	}
}
