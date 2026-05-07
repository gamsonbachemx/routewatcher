package routes

import (
	"fmt"
	"io"
	"os"
)

// DefaultEnvelopeCmdConfig returns defaults for the envelope sub-command.
func DefaultEnvelopeCmdConfig() EnvelopeCmdConfig {
	return EnvelopeCmdConfig{
		Envelope: DefaultEnvelopeConfig(),
		Output:   os.Stdout,
	}
}

// EnvelopeCmdConfig configures RunEnvelopeEmit.
type EnvelopeCmdConfig struct {
	Envelope EnvelopeConfig
	Output   io.Writer
}

// RunEnvelopeEmit wraps the supplied Diff in an Envelope, serialises it as
// JSON and writes it to cfg.Output followed by a newline.
func RunEnvelopeEmit(d Diff, cfg EnvelopeCmdConfig) error {
	e := NewEnveloper(cfg.Envelope)
	env := e.Wrap(d)

	b, err := env.Marshal()
	if err != nil {
		return fmt.Errorf("envelope emit: %w", err)
	}

	if _, err := fmt.Fprintf(cfg.Output, "%s\n", b); err != nil {
		return fmt.Errorf("envelope emit: write: %w", err)
	}
	return nil
}

// RunEnvelopeShow pretty-prints envelope metadata (version, source, sequence)
// for the supplied Diff without the full diff body.
func RunEnvelopeShow(d Diff, cfg EnvelopeCmdConfig) error {
	e := NewEnveloper(cfg.Envelope)
	env := e.Wrap(d)

	_, err := fmt.Fprintf(cfg.Output,
		"version=%s source=%s seq=%d timestamp=%s added=%d removed=%d\n",
		env.Version,
		env.Source,
		env.Sequence,
		env.Timestamp.Format("2006-01-02T15:04:05Z"),
		len(env.Diff.Added),
		len(env.Diff.Removed),
	)
	if err != nil {
		return fmt.Errorf("envelope show: %w", err)
	}
	return nil
}
