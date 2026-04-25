package routes

import "strings"

// Filter holds criteria for including or excluding routes from a diff.
type Filter struct {
	// Interfaces limits output to routes on the specified interfaces.
	Interfaces []string
	// Protocols limits output to routes with the specified protocols (e.g. "kernel", "static").
	Protocols []string
	// ExcludeLocal drops local/loopback routes when true.
	ExcludeLocal bool
}

// Apply returns a new snapshot containing only the routes that match the filter.
func (f *Filter) Apply(snap Snapshot) Snapshot {
	if f == nil {
		return snap
	}

	filtered := make(Snapshot, 0, len(snap))
	for _, r := range snap {
		if f.ExcludeLocal && isLocal(r) {
			continue
		}
		if len(f.Interfaces) > 0 && !containsStr(f.Interfaces, r.Iface) {
			continue
		}
		if len(f.Protocols) > 0 && !containsStr(f.Protocols, r.Protocol) {
			continue
		}
		filtered = append(filtered, r)
	}
	return filtered
}

func isLocal(r Route) bool {
	return strings.HasPrefix(r.Destination, "127.") ||
		r.Destination == "::1" ||
		r.Iface == "lo"
}

func containsStr(list []string, s string) bool {
	for _, v := range list {
		if strings.EqualFold(v, s) {
			return true
		}
	}
	return false
}
