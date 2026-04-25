package routes

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
)

// Route represents a single routing table entry.
type Route struct {
	Destination string
	Gateway     string
	Iface       string
	Flags       string
	Metric      string
}

// Snapshot holds a point-in-time capture of the routing table.
type Snapshot struct {
	Routes []Route
}

// String returns a human-readable key for a route used in diffing.
func (r Route) String() string {
	return fmt.Sprintf("%s via %s dev %s metric %s", r.Destination, r.Gateway, r.Iface, r.Metric)
}

// Capture reads the current routing table using `ip route show`.
func Capture() (*Snapshot, error) {
	out, err := exec.Command("ip", "route", "show").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run ip route show: %w", err)
	}
	return parseOutput(string(out))
}

// parseOutput parses the output of `ip route show` into a Snapshot.
func parseOutput(output string) (*Snapshot, error) {
	var routes []Route
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		r, err := parseLine(line)
		if err != nil {
			continue // skip unparseable lines
		}
		routes = append(routes, r)
	}
	return &Snapshot{Routes: routes}, nil
}

// parseLine parses a single line from `ip route show`.
func parseLine(line string) (Route, error) {
	fields := strings.Fields(line)
	if len(fields) < 1 {
		return Route{}, fmt.Errorf("empty line")
	}
	r := Route{Destination: fields[0]}
	// Validate or normalise destination
	if r.Destination != "default" {
		_, _, err := net.ParseCIDR(r.Destination)
		if err != nil {
			if net.ParseIP(r.Destination) == nil {
				return Route{}, fmt.Errorf("invalid destination: %s", r.Destination)
			}
		}
	}
	for i := 1; i < len(fields)-1; i++ {
		switch fields[i] {
		case "via":
			r.Gateway = fields[i+1]
		case "dev":
			r.Iface = fields[i+1]
		case "metric":
			r.Metric = fields[i+1]
		}
	}
	return r, nil
}
