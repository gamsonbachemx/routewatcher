package routes

import (
	"fmt"
	"strings"
	"time"
)

// Summary holds aggregated statistics over a window of diffs.
type Summary struct {
	WindowStart  time.Time
	WindowEnd    time.Time
	TotalDiffs   int
	TotalAdded   int
	TotalRemoved int
	TopInterfaces []InterfaceCount
	TopProtocols  []ProtocolCount
}

// InterfaceCount pairs an interface name with a change count.
type InterfaceCount struct {
	Iface string
	Count int
}

// ProtocolCount pairs a protocol with a change count.
type ProtocolCount struct {
	Protocol string
	Count int
}

// Summarize aggregates a slice of Diffs into a Summary.
func Summarize(diffs []Diff, start, end time.Time) Summary {
	s := Summary{
		WindowStart: start,
		WindowEnd:   end,
		TotalDiffs:  len(diffs),
	}

	ifaceMap := map[string]int{}
	protoMap := map[string]int{}

	for _, d := range diffs {
		s.TotalAdded += len(d.Added)
		s.TotalRemoved += len(d.Removed)

		for _, r := range append(d.Added, d.Removed...) {
			if r.Iface != "" {
				ifaceMap[r.Iface]++
			}
			if r.Protocol != "" {
				protoMap[r.Protocol]++
			}
		}
	}

	for iface, count := range ifaceMap {
		s.TopInterfaces = append(s.TopInterfaces, InterfaceCount{Iface: iface, Count: count})
	}
	for proto, count := range protoMap {
		s.TopProtocols = append(s.TopProtocols, ProtocolCount{Protocol: proto, Count: count})
	}

	return s
}

// FormatSummary returns a human-readable summary string.
func FormatSummary(s Summary) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Summary [%s → %s]\n",
		s.WindowStart.Format(time.RFC3339),
		s.WindowEnd.Format(time.RFC3339))
	fmt.Fprintf(&sb, "  Diffs:   %d\n", s.TotalDiffs)
	fmt.Fprintf(&sb, "  Added:   %d\n", s.TotalAdded)
	fmt.Fprintf(&sb, "  Removed: %d\n", s.TotalRemoved)

	if len(s.TopInterfaces) > 0 {
		sb.WriteString("  Interfaces:\n")
		for _, ic := range s.TopInterfaces {
			fmt.Fprintf(&sb, "    %-15s %d changes\n", ic.Iface, ic.Count)
		}
	}
	if len(s.TopProtocols) > 0 {
		sb.WriteString("  Protocols:\n")
		for _, pc := range s.TopProtocols {
			fmt.Fprintf(&sb, "    %-15s %d changes\n", pc.Protocol, pc.Count)
		}
	}
	return sb.String()
}
