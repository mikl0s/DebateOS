// Package registry implements the DebateOS static registry index generator
// (REG-01). It scans point and opinion YAML files from a fixture directory,
// validates each via resolver/parse (strict, no silent skips), computes
// foundation-compatibility from the translators' capabilities.json, and emits
// a deterministic JSON index plus minimal static browse HTML for GitHub Pages.
//
// Git is authoritative — the index is a derived, idempotent cache (D12).
// Identical inputs always produce byte-identical JSON output.
package registry

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mikl0s/debateos/registry/index"
	resolver "github.com/mikl0s/debateos/resolver"
	"github.com/mikl0s/debateos/resolver/parse"
)

// GenerateIndex scans fixturesDir/points and fixturesDir/opinions, parses and
// validates each YAML file via resolver/parse (KnownFields strict), builds
// the opinions map, and assembles a RegistryIndex with foundation_compat
// computed from the caps map.
//
// Parameters:
//   - fixturesDir: root directory containing points/ and opinions/ sub-dirs
//   - caps: map[foundationID]→[]token loaded from translators/*/capabilities.json
//   - generatedAt: ISO 8601 timestamp to embed; caller controls for determinism
//
// Returns a wrapped error naming the offending file on any parse failure.
func GenerateIndex(fixturesDir string, caps map[string][]string, generatedAt string) (*index.RegistryIndex, error) {
	pointsDir := filepath.Join(fixturesDir, "points")
	opinionsDir := filepath.Join(fixturesDir, "opinions")

	// Load and validate all opinions first (they are referenced by points).
	opinions, err := loadOpinions(opinionsDir)
	if err != nil {
		return nil, err
	}

	// Load and validate all points.
	points, err := loadPoints(pointsDir)
	if err != nil {
		return nil, err
	}

	// Sort points by ID for deterministic output.
	sort.Slice(points, func(i, j int) bool {
		return points[i].ID < points[j].ID
	})

	entries := make([]index.PointEntry, 0, len(points))
	for _, pt := range points {
		entry := buildEntry(pt, opinions, caps)
		entries = append(entries, entry)
	}

	return &index.RegistryIndex{
		Schema:      index.SchemaVersion,
		GeneratedAt: generatedAt,
		Points:      entries,
	}, nil
}

// loadOpinions reads all .yaml files from dir, parses each via
// parse.ParseOpinion (strict), and returns a map ID→Opinion.
// Returns a wrapped error naming the file on any parse failure.
func loadOpinions(dir string) (map[resolver.OpinionID]resolver.Opinion, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[resolver.OpinionID]resolver.Opinion{}, nil
		}
		return nil, fmt.Errorf("read opinions dir %s: %w", dir, err)
	}

	out := make(map[resolver.OpinionID]resolver.Opinion, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", e.Name(), err)
		}
		op, parseErr := parse.ParseOpinion(f)
		f.Close()
		if parseErr != nil {
			// Error MUST name the offending file (T-05-01).
			return nil, fmt.Errorf("validate %s: %w", e.Name(), parseErr)
		}
		out[op.ID] = *op
	}
	return out, nil
}

// loadPoints reads all .yaml files from dir, parses each via parse.ParsePoint
// (strict), and returns a slice of Points sorted by ID.
func loadPoints(dir string) ([]resolver.Point, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read points dir %s: %w", dir, err)
	}

	var out []resolver.Point
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", e.Name(), err)
		}
		pt, parseErr := parse.ParsePoint(f)
		f.Close()
		if parseErr != nil {
			return nil, fmt.Errorf("validate %s: %w", e.Name(), parseErr)
		}
		out = append(out, *pt)
	}
	return out, nil
}

// buildEntry constructs a PointEntry from a Point, resolving member IDs to
// opinion metadata, computing foundation-compat, and sorting slices for
// deterministic JSON output.
func buildEntry(
	pt resolver.Point,
	opinions map[resolver.OpinionID]resolver.Opinion,
	caps map[string][]string,
) index.PointEntry {
	// Collect member IDs, sorted for determinism.
	memberIDs := make([]string, 0, len(pt.Members))
	for _, m := range pt.Members {
		memberIDs = append(memberIDs, string(m.ID))
	}
	sort.Strings(memberIDs)

	// Compute foundation compatibility.
	compat := index.ComputeCompat(pt, opinions, caps)

	tags := []string{}

	return index.PointEntry{
		ID:               pt.ID,
		Name:             pt.Name,
		Intent:           pt.Intent,
		Curator:          pt.Curator,
		Members:          memberIDs,
		FoundationCompat: compat,
		Tags:             tags,
	}
}

// LoadCapabilities reads the two capabilities.json files (arch and debian
// translators) and returns a map[foundationID]→[]token for use in GenerateIndex.
//
// The returned map always has keys "arch" and "debian".
func LoadCapabilities(archPath, debianPath string) (map[string][]string, error) {
	type capsFile struct {
		Capabilities []string `json:"capabilities"`
	}

	readFile := func(path, foundation string) ([]string, error) {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read capabilities.json for %s: %w", foundation, err)
		}
		var cf capsFile
		if err := json.Unmarshal(data, &cf); err != nil {
			return nil, fmt.Errorf("parse capabilities.json for %s: %w", foundation, err)
		}
		return cf.Capabilities, nil
	}

	archCaps, err := readFile(archPath, "arch")
	if err != nil {
		return nil, err
	}
	debianCaps, err := readFile(debianPath, "debian")
	if err != nil {
		return nil, err
	}

	return map[string][]string{
		"arch":   archCaps,
		"debian": debianCaps,
	}, nil
}

// browseHTMLTpl is the minimal static browse HTML template (no JS, Pages-static).
// One table row per point; compat badges rendered as plain text symbols.
var browseHTMLTpl = template.Must(template.New("browse").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>DebateOS Registry — Point Browser</title>
  <style>
    body { font-family: system-ui, sans-serif; max-width: 960px; margin: 2rem auto; padding: 0 1rem; }
    h1 { font-size: 1.5rem; }
    table { border-collapse: collapse; width: 100%; }
    th, td { text-align: left; padding: 0.5rem; border-bottom: 1px solid #ddd; }
    th { background: #f5f5f5; }
    .compat-yes { color: #16a34a; font-weight: bold; }
    .compat-no  { color: #dc2626; }
    small { color: #6b7280; }
  </style>
</head>
<body>
<h1>DebateOS Registry</h1>
<p><small>Generated: {{.GeneratedAt}} — schema v{{.Schema}}</small></p>
<p>That's just your opinion, man. Browse {{len .Points}} points below.</p>
<table>
  <thead>
    <tr>
      <th>ID</th>
      <th>Name</th>
      <th>Curator</th>
      <th>Foundation Compat</th>
    </tr>
  </thead>
  <tbody>
  {{range .Points}}
    <tr>
      <td><code>{{.ID}}</code></td>
      <td>{{.Name}}</td>
      <td>{{.Curator}}</td>
      <td>{{range .FoundationCompat}}{{if .Compatible}}<span class="compat-yes">{{.Foundation}} ✓</span>{{else}}<span class="compat-no">{{.Foundation}} ✗</span>{{end}} {{end}}</td>
    </tr>
  {{end}}
  </tbody>
</table>
</body>
</html>
`))

// EmitHTML writes a minimal static browse page (no JS, GitHub Pages compatible)
// to w. One table row per point with text-based foundation compat badges.
func EmitHTML(idx *index.RegistryIndex, w io.Writer) error {
	return browseHTMLTpl.Execute(w, idx)
}
