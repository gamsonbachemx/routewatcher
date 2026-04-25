package routes

import (
	"flag"
	"strings"
	"time"
)

// Config holds all runtime configuration for routewatcher.
type Config struct {
	Interval     time.Duration
	OutputFormat string
	Filter       *Filter
}

// multiStringFlag allows a flag to be specified multiple times.
type multiStringFlag []string

func (m *multiStringFlag) String() string  { return strings.Join(*m, ",") }
func (m *multiStringFlag) Set(v string) error { *m = append(*m, v); return nil }

// ParseFlags reads CLI flags and returns a populated Config.
func ParseFlags() *Config {
	var (
		interval     = flag.Duration("interval", 5*time.Second, "polling interval (e.g. 5s, 1m)")
		format       = flag.String("format", "text", "output format: text or json")
		excludeLocal = flag.Bool("exclude-local", false, "exclude local/loopback routes")
		ifaces       multiStringFlag
		protos       multiStringFlag
	)

	flag.Var(&ifaces, "iface", "filter by interface name (repeatable)")
	flag.Var(&protos, "proto", "filter by protocol (repeatable)")
	flag.Parse()

	var f *Filter
	if *excludeLocal || len(ifaces) > 0 || len(protos) > 0 {
		f = &Filter{
			ExcludeLocal: *excludeLocal,
			Interfaces:   []string(ifaces),
			Protocols:    []string(protos),
		}
	}

	return &Config{
		Interval:     *interval,
		OutputFormat: *format,
		Filter:       f,
	}
}
