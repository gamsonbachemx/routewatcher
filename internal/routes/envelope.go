package routes

import (
	"encoding/json"
	"fmt"
	"time"
)

// EnvelopeVersion is the current envelope schema version.
const EnvelopeVersion = "v1"

// DefaultEnvelopeConfig returns a sensible default EnvelopeConfig.
func DefaultEnvelopeConfig() EnvelopeConfig {
	return EnvelopeConfig{
		Source:  "routewatcher",
		Version: EnvelopeVersion,
	}
}

// EnvelopeConfig controls how diffs are wrapped.
type EnvelopeConfig struct {
	Source  string
	Version string
}

// Envelope wraps a Diff with metadata for downstream consumers.
type Envelope struct {
	Version   string    `json:"version"`
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp"`
	Sequence  uint64    `json:"sequence"`
	Diff      Diff      `json:"diff"`
}

// Enveloper wraps outgoing Diff values in an Envelope.
type Enveloper struct {
	cfg EnvelopeConfig
	seq uint64
}

// NewEnveloper creates an Enveloper with the given config.
func NewEnveloper(cfg EnvelopeConfig) *Enveloper {
	return &Enveloper{cfg: cfg}
}

// Wrap returns an Envelope containing the given Diff.
func (e *Enveloper) Wrap(d Diff) Envelope {
	e.seq++
	return Envelope{
		Version:   e.cfg.Version,
		Source:    e.cfg.Source,
		Timestamp: time.Now().UTC(),
		Sequence:  e.seq,
		Diff:      d,
	}
}

// Marshal serialises the Envelope as JSON.
func (env Envelope) Marshal() ([]byte, error) {
	b, err := json.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("envelope: marshal: %w", err)
	}
	return b, nil
}
