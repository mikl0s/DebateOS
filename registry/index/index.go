// Package index defines the RegistryIndex and PointEntry types that form the
// schema of the static JSON index emitted by the registry generator (REG-01).
//
// Schema version 1 is stable. All slices in the emitted index are sorted for
// deterministic, byte-identical JSON output across repeated runs on identical
// inputs (D12 — index is a derived, idempotent cache).
package index

const SchemaVersion = 1

// RegistryIndex is the top-level JSON object written to registry/index.json
// and served on GitHub Pages.
type RegistryIndex struct {
	Schema      int          `json:"schema"`
	GeneratedAt string       `json:"generated_at"`
	Points      []PointEntry `json:"points"`
}

// PointEntry is one row in the index: the public metadata for a single Point
// consumed by the Debate UI (point discovery) and the Forum (what it indexes).
type PointEntry struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Intent          string           `json:"intent,omitempty"`
	Curator         string           `json:"curator,omitempty"`
	Members         []string         `json:"members"`
	FoundationCompat []FoundationCompat `json:"foundation_compat"`
	CommitDate      string           `json:"commit_date,omitempty"`
	Tags            []string         `json:"tags"`
}
