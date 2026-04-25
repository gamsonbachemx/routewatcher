package routes

// Diff represents the changes between two routing table snapshots.
type Diff struct {
	Added   []Route
	Removed []Route
}

// IsEmpty returns true when there are no changes.
func (d Diff) IsEmpty() bool {
	return len(d.Added) == 0 && len(d.Removed) == 0
}

// Compare computes the diff between an old and new snapshot.
func Compare(old, new *Snapshot) Diff {
	oldMap := indexRoutes(old.Routes)
	newMap := indexRoutes(new.Routes)

	var added, removed []Route

	for key, route := range newMap {
		if _, exists := oldMap[key]; !exists {
			added = append(added, route)
		}
	}

	for key, route := range oldMap {
		if _, exists := newMap[key]; !exists {
			removed = append(removed, route)
		}
	}

	return Diff{Added: added, Removed: removed}
}

// indexRoutes builds a map keyed by the route's string representation.
func indexRoutes(routes []Route) map[string]Route {
	m := make(map[string]Route, len(routes))
	for _, r := range routes {
		m[r.String()] = r
	}
	return m
}
