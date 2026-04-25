package routes

import (
	"io"
	"os"
	"time"
)

// SummaryConfig controls periodic summary reporting.
type SummaryConfig struct {
	Interval time.Duration
	Output   io.Writer
}

// DefaultSummaryConfig returns a SummaryConfig with sensible defaults.
func DefaultSummaryConfig() SummaryConfig {
	return SummaryConfig{
		Interval: 5 * time.Minute,
		Output:   os.Stdout,
	}
}

// Summarizer collects diffs and periodically emits summaries.
type Summarizer struct {
	cfg    SummaryConfig
	buffer []Diff
	ticker *time.Ticker
	done   chan struct{}
}

// NewSummarizer creates and starts a Summarizer.
func NewSummarizer(cfg SummaryConfig) *Summarizer {
	s := &Summarizer{
		cfg:  cfg,
		done: make(chan struct{}),
	}
	s.ticker = time.NewTicker(cfg.Interval)
	go s.run()
	return s
}

// Record adds a Diff to the current window buffer.
func (s *Summarizer) Record(d Diff) {
	s.buffer = append(s.buffer, d)
}

// Stop halts the summarizer and flushes any remaining diffs.
func (s *Summarizer) Stop() {
	s.ticker.Stop()
	close(s.done)
}

func (s *Summarizer) flush(start, end time.Time) {
	if len(s.buffer) == 0 {
		return
	}
	summary := Summarize(s.buffer, start, end)
	s.buffer = nil
	s.cfg.Output.Write([]byte(FormatSummary(summary))) //nolint:errcheck
}

func (s *Summarizer) run() {
	windowStart := time.Now()
	for {
		select {
		case t := <-s.ticker.C:
			s.flush(windowStart, t)
			windowStart = t
		case <-s.done:
			s.flush(windowStart, time.Now())
			return
		}
	}
}
