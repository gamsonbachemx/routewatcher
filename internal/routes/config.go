package routes

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// multiStringFlag allows a flag to be specified multiple times.
type multiStringFlag []string

func (m *multiStringFlag) String() string {
	return fmt.Sprintf("%v", *m)
}

func (m *multiStringFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

// Config holds all CLI configuration for routewatcher.
type Config struct {
	Interval   time.Duration
	Format     string
	Interfaces multiStringFlag
	Protocols  multiStringFlag
	ExcludeLocal bool
	// Alert settings
	AlertEnabled    bool
	AlertMinChanges int
	AlertLevel      AlertLevel
}

// ParseFlags parses command-line flags into a Config.
func ParseFlags() Config {
	cfg := Config{}

	flag.DurationVar(&cfg.Interval, "interval", 5*time.Second, "polling interval")
	flag.StringVar(&cfg.Format, "format", "text", "output format: text or json")
	flag.BoolVar(&cfg.ExcludeLocal, "exclude-local", false, "exclude local/loopback routes")
	flag.Var(&cfg.Interfaces, "iface", "filter by interface (repeatable)")
	flag.Var(&cfg.Protocols, "proto", "filter by protocol (repeatable)")

	// Alert flags
	flag.BoolVar(&cfg.AlertEnabled, "alert", false, "enable alerting on route changes")
	flag.IntVar(&cfg.AlertMinChanges, "alert-min-changes", 1, "minimum changes to trigger alert")
	alertLevel := flag.String("alert-level", "WARNING", "alert severity level: INFO, WARNING, CRITICAL")

	flag.Parse()

	cfg.AlertLevel = AlertLevel(*alertLevel)
	return cfg
}

// NewAlerterFromConfig builds an Alerter from the alert-related Config fields.
// Returns nil if alerting is disabled.
func NewAlerterFromConfig(cfg Config) *Alerter {
	if !cfg.AlertEnabled {
		return nil
	}
	return NewAlerter(AlertConfig{
		MinChanges: cfg.AlertMinChanges,
		Level:      cfg.AlertLevel,
		Output:     os.Stderr,
	})
}
