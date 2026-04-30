package routes

import (
	"strings"
)

// TagConfig holds configuration for route tagging.
type TagConfig struct {
	// Rules maps a tag name to a list of prefix patterns (e.g. "10." or "via 192.168.")
	Rules map[string][]string
}

// DefaultTagConfig returns a TagConfig with sensible built-in tags.
func DefaultTagConfig() TagConfig {
	return TagConfig{
		Rules: map[string][]string{
			"private": {"10.", "172.16.", "172.17.", "172.18.", "172.19.",
				"172.20.", "172.21.", "172.22.", "172.23.", "172.24.",
				"172.25.", "172.26.", "172.27.", "172.28.", "172.29.",
				"172.30.", "172.31.", "192.168."},
			"default": {"0.0.0.0", "::/0"},
			"loopback": {"127.", "::1"},
		},
	}
}

// Tagger assigns string tags to Route entries based on configured rules.
type Tagger struct {
	cfg TagConfig
}

// NewTagger creates a Tagger from the given TagConfig.
func NewTagger(cfg TagConfig) *Tagger {
	return &Tagger{cfg: cfg}
}

// Tag returns the set of tags that apply to the given Route's destination.
// A route may match multiple tags.
func (t *Tagger) Tag(r Route) []string {
	var tags []string
	for tag, prefixes := range t.cfg.Rules {
		for _, prefix := range prefixes {
			if strings.HasPrefix(r.Destination, prefix) {
				tags = append(tags, tag)
				break
			}
		}
	}
	return tags
}

// TagDiff annotates each route in a Diff's Added and Removed slices with
// matching tags, returned as a map keyed by route destination.
func (t *Tagger) TagDiff(d Diff) map[string][]string {
	result := make(map[string][]string)
	for _, r := range d.Added {
		if tags := t.Tag(r); len(tags) > 0 {
			result[r.Destination] = tags
		}
	}
	for _, r := range d.Removed {
		if tags := t.Tag(r); len(tags) > 0 {
			result[r.Destination] = tags
		}
	}
	return result
}
