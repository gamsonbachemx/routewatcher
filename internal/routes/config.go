package routes

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// multiStringFlag allows repeated -flag value usage.
type multiStringFlag []string

func (m *multiStringFlag) String() string {
	if m == nil {
		return ""
	}
	result := ""
	for i, v := range *m {
		if i > 0 {
			result += ","
		}
		result += v
	}
	return result
}

func (m *multiStringFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

// Config holds all CLI configuration for routewatcher.
type Config struct {
	Interval        time.Duration
	Format          string
	OutputPath      string
	WebhookURL      string
	WebhookHeader   string
	AlertThreshold  int
	FilterIfaces    multiStringFlag
	FilterProtocols multiStringFlag
	ExcludeLocal    bool
	SummaryInterval time.Duration
	HistoryFile     string
	HistoryMax      int
}

// ParseFlags parses os.Args and returns a populated Config.
func ParseFlags() Config {
	var cfg Config
	flag.DurationVar(&cfg.Interval, "interval", 5*time.Second, "polling interval")
	flag.StringVar(&cfg.Format, "format", "text", "output format: text or json")
	flag.StringVar(&cfg.OutputPath, "output", "", "file path to export diffs (optional)")
	flag.StringVar(&cfg.WebhookURL, "webhook", "", "webhook URL for notifications")
	flag.StringVar(&cfg.WebhookHeader, "webhook-header", "", "extra header for webhook (Key:Value)")
	flag.IntVar(&cfg.AlertThreshold, "alert-threshold", 5, "number of changes to trigger alert")
	flag.Var(&cfg.FilterIfaces, "iface", "filter by interface (repeatable)")
	flag.Var(&cfg.FilterProtocols, "protocol", "filter by protocol (repeatable)")
	flag.BoolVar(&cfg.ExcludeLocal, "exclude-local", false, "exclude local/loopback routes")
	flag.DurationVar(&cfg.SummaryInterval, "summary-interval", 5*time.Minute, "interval for periodic summaries")
	flag.StringVar(&cfg.HistoryFile, "history-file", "", "path to persist diff history")
	flag.IntVar(&cfg.HistoryMax, "history-max", 100, "maximum history entries to keep")
	flag.Parse()
	return cfg
}

// NewAlerterFromConfig builds an Alerter from the given Config.
func NewAlerterFromConfig(cfg Config) *Alerter {
	ac := DefaultAlertConfig()
	ac.Threshold = cfg.AlertThreshold
	return NewAlerter(ac)
}

// NewNotifierFromConfig builds a Notifier from the given Config, or nil.
func NewNotifierFromConfig(cfg Config) (*Notifier, error) {
	if cfg.WebhookURL == "" {
		return nil, nil
	}
	nc := NotifierConfig{URL: cfg.WebhookURL}
	if cfg.WebhookHeader != "" {
		nc.Headers = map[string]string{}
		var k, v string
		if _, err := fmt.Sscanf(cfg.WebhookHeader, "%s", &k); err == nil {
			nc.Headers[k] = v
		}
	}
	return NewNotifier(nc)
}

// NewSummarizerFromConfig builds a Summarizer from the given Config.
func NewSummarizerFromConfig(cfg Config) *Summarizer {
	sc := SummaryConfig{
		Interval: cfg.SummaryInterval,
		Output:   os.Stdout,
	}
	return NewSummarizer(sc)
}
