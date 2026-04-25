package routes

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

// multiStringFlag allows repeated -flag value usage.
type multiStringFlag []string

func (m *multiStringFlag) String() string  { return strings.Join(*m, ",") }
func (m *multiStringFlag) Set(v string) error { *m = append(*m, v); return nil }

// Config holds all CLI-parsed configuration for routewatcher.
type Config struct {
	Interval    time.Duration
	Format      string
	OutputPath  string
	ExcludeLocal bool
	Interfaces  multiStringFlag
	Protocols   multiStringFlag
	AlertThreshold int
	WebhookURL  string
	WebhookHeaders multiStringFlag
}

// ParseFlags parses os.Args and returns a populated Config.
func ParseFlags() *Config {
	cfg := &Config{}
	flag.DurationVar(&cfg.Interval, "interval", 5*time.Second, "polling interval")
	flag.StringVar(&cfg.Format, "format", "text", "output format: text or json")
	flag.StringVar(&cfg.OutputPath, "output", "", "file path to write diffs (stdout if empty)")
	flag.BoolVar(&cfg.ExcludeLocal, "exclude-local", false, "exclude local/loopback routes")
	flag.Var(&cfg.Interfaces, "iface", "filter by interface (repeatable)")
	flag.Var(&cfg.Protocols, "proto", "filter by protocol (repeatable)")
	flag.IntVar(&cfg.AlertThreshold, "alert-threshold", 1, "minimum changes to trigger alert")
	flag.StringVar(&cfg.WebhookURL, "webhook", "", "webhook URL for change notifications")
	flag.Var(&cfg.WebhookHeaders, "webhook-header", "extra header for webhook (Key:Value, repeatable)")
	flag.Parse()
	return cfg
}

// NewAlerterFromConfig builds an Alerter from the parsed config.
func NewAlerterFromConfig(cfg *Config, out *os.File) *Alerter {
	ac := DefaultAlertConfig()
	ac.Threshold = cfg.AlertThreshold
	if out != nil {
		ac.Output = out
	}
	return NewAlerter(ac)
}

// NewNotifierFromConfig builds a Notifier if a webhook URL is configured.
// Returns nil and no error when no URL is set.
func NewNotifierFromConfig(cfg *Config) (*Notifier, error) {
	if cfg.WebhookURL == "" {
		return nil, nil
	}
	headers := make(map[string]string)
	for _, h := range cfg.WebhookHeaders {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid webhook header %q: expected Key:Value", h)
		}
		headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return NewNotifier(WebhookConfig{
		URL:     cfg.WebhookURL,
		Headers: headers,
	})
}
