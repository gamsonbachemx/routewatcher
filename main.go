package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/user/routewatcher/watcher"
)

const (
	defaultInterval = 5 * time.Second
	version         = "0.1.0"
)

func main() {
	var (
		interval    = flag.Duration("interval", defaultInterval, "polling interval for route table checks")
		once        = flag.Bool("once", false, "print current routing table and exit")
		showVersion = flag.Bool("version", false, "print version and exit")
		jsonOutput  = flag.Bool("json", false, "output diffs in JSON format")
		verbose     = flag.Bool("verbose", false, "enable verbose logging")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("routewatcher v%s\n", version)
		os.Exit(0)
	}

	cfg := watcher.Config{
		Interval:   *interval,
		JSONOutput: *jsonOutput,
		Verbose:    *verbose,
	}

	w, err := watcher.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error initializing watcher: %v\n", err)
		os.Exit(1)
	}

	if *once {
		// Snapshot mode: print current routes and exit
		routes, err := w.Snapshot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading routing table: %v\n", err)
			os.Exit(1)
		}
		for _, r := range routes {
			fmt.Println(r)
		}
		os.Exit(0)
	}

	// Watch mode: continuously monitor for changes
	fmt.Printf("routewatcher v%s — watching routing table every %s\n", version, *interval)
	if err := w.Watch(); err != nil {
		fmt.Fprintf(os.Stderr, "watcher error: %v\n", err)
		os.Exit(1)
	}
}
