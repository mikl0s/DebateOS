// cmd/resolve-json — DebateOS resolve-json helper
//
// Loads a speech composition from a directory (speech.yaml + points/ + opinions/),
// resolves it, and writes the canonical ResolvedSpeech JSON to stdout.
//
// This is the Phase 2 north-star pipeline seed and the Phase 3 CLI seed.
// The resolved.json it emits is the input contract for the Arch translator.
//
// Usage:
//
//	go run ./cmd/resolve-json <speech-dir>
//
// Example:
//
//	go run ./cmd/resolve-json examples/omarchy > resolved.json
//
// Arguments:
//
//	<speech-dir>   Directory containing:
//	               - speech.yaml   (the composition speech)
//	               - points/       (YAML files, one per point)
//	               - opinions/     (YAML files, one per opinion, e.g. OM-*.yaml)
//
// Output:
//
//	Canonical ResolvedSpeech JSON on stdout (deterministic; T-01-12).
//	Informational messages on stderr.
//
// Exit codes:
//
//	0   Success; JSON written to stdout.
//	1   Error (parse failure, hard conflict, or IO error).
//
// Source: examples/omarchy_test.go pattern (assembleOpinionsFromSpeech).
// Phase 3 CLI seed: this binary will become the `debateos resolve` subcommand.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "go.yaml.in/yaml/v3"

	"github.com/mikl0s/debateos/resolver"
	"github.com/mikl0s/debateos/resolver/hardware"
	"github.com/mikl0s/debateos/resolver/parse"
	"github.com/mikl0s/debateos/resolver/resolve"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Fprintf(os.Stderr, "usage: go run ./cmd/resolve-json <speech-dir>\n")
		fmt.Fprintf(os.Stderr, "example: go run ./cmd/resolve-json examples/omarchy > resolved.json\n")
		os.Exit(1)
	}

	dir := os.Args[1]

	rs, err := resolveDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve-json: %v\n", err)
		os.Exit(1)
	}

	out, err := resolve.CanonicalJSON(rs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve-json: canonical JSON: %v\n", err)
		os.Exit(1)
	}

	// IN-03: Write the canonical bytes directly so that SOURCE_DATE_EPOCH derivation
	// is consistent with CanonicalJSON output (re-marshalling via map[string]interface{}
	// reorders keys alphabetically, producing a non-canonical byte sequence).
	// The translator's json.loads() handles compact JSON without issue.
	os.Stdout.Write(out) //nolint:errcheck
	os.Stdout.Write([]byte("\n"))
}

// resolveDir loads a speech directory and runs resolve.Resolve on it.
func resolveDir(dir string) (*resolve.ResolvedSpeech, error) {
	// --- 1. Load speech.yaml ---
	speechPath := filepath.Join(dir, "speech.yaml")
	sf, err := os.Open(speechPath)
	if err != nil {
		return nil, fmt.Errorf("open speech.yaml: %w", err)
	}
	defer sf.Close()

	speech, err := parse.ParseSpeech(sf)
	if err != nil {
		return nil, fmt.Errorf("parse speech.yaml: %w", err)
	}
	fmt.Fprintf(os.Stderr, "resolve-json: loaded speech %q (foundation: %s, %d points)\n",
		speech.ID, speech.Foundation, len(speech.Points))

	// --- 2. Load points/ directory ---
	points, err := loadPoints(filepath.Join(dir, "points"))
	if err != nil {
		return nil, fmt.Errorf("load points: %w", err)
	}
	fmt.Fprintf(os.Stderr, "resolve-json: loaded %d point files\n", len(points))

	// --- 3. Load opinions/ directory ---
	opIdx, err := loadOpinions(filepath.Join(dir, "opinions"))
	if err != nil {
		return nil, fmt.Errorf("load opinions: %w", err)
	}
	fmt.Fprintf(os.Stderr, "resolve-json: loaded %d opinion files\n", len(opIdx))

	// --- 4. Assemble flat []resolver.Opinion in speech order ---
	opinions, err := assembleOpinions(speech, points, opIdx)
	if err != nil {
		return nil, fmt.Errorf("assemble opinions: %w", err)
	}
	fmt.Fprintf(os.Stderr, "resolve-json: assembled %d opinions for resolution\n", len(opinions))

	// --- 5. Build hardware profile from speech ---
	hw := hardware.HardwareProfile{}
	if speech.Hardware != nil {
		hw.Predicates = speech.Hardware.Predicates
		hw.Facts = speech.Hardware.Facts
		hw.PCIIDs = speech.Hardware.PCIIDs
	}

	// --- 6. Resolve ---
	rs, err := resolve.Resolve(speech, opinions, hw)
	if err != nil {
		if rs != nil {
			// Surface hard conflict explanations for diagnostics
			for _, ex := range rs.Explanations {
				if strings.Contains(ex.Text, "Hard conflict") {
					fmt.Fprintf(os.Stderr, "resolve-json: CONFLICT: %s\n", ex.Text)
				}
			}
		}
		return rs, fmt.Errorf("resolve: %w", err)
	}

	fmt.Fprintf(os.Stderr, "resolve-json: Applied=%d Skipped=%d Dropped=%d InstallOrder=%d\n",
		len(rs.Applied), len(rs.Skipped), len(rs.Dropped), len(rs.InstallOrder))

	return rs, nil
}

// loadPoints reads all .yaml files from dir and returns a map point-id → Point.
func loadPoints(dir string) (map[string]resolver.Point, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]resolver.Point{}, nil
		}
		return nil, err
	}

	out := make(map[string]resolver.Point, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		f, err := os.Open(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", e.Name(), err)
		}
		var pt resolver.Point
		if err := yaml.NewDecoder(f).Decode(&pt); err != nil {
			f.Close()
			return nil, fmt.Errorf("decode %s: %w", e.Name(), err)
		}
		f.Close()
		out[pt.ID] = pt
	}
	return out, nil
}

// loadOpinions reads all .yaml files from dir and returns a map id → Opinion.
func loadOpinions(dir string) (map[resolver.OpinionID]resolver.Opinion, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[resolver.OpinionID]resolver.Opinion{}, nil
		}
		return nil, err
	}

	out := make(map[resolver.OpinionID]resolver.Opinion, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		f, err := os.Open(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", e.Name(), err)
		}
		op, err := parse.ParseOpinion(f)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", e.Name(), err)
		}
		out[op.ID] = *op
	}
	return out, nil
}

// assembleOpinions builds the flat []resolver.Opinion by expanding speech.Points
// through the loaded point files, deduplicating by opinion ID (stable first-occurrence).
// This mirrors the pattern in examples/omarchy_test.go::assembleOpinionsFromSpeech.
func assembleOpinions(
	speech *resolver.Speech,
	points map[string]resolver.Point,
	opIdx map[resolver.OpinionID]resolver.Opinion,
) ([]resolver.Opinion, error) {
	seen := make(map[resolver.OpinionID]bool)
	var out []resolver.Opinion

	for _, pref := range speech.Points {
		pt, ok := points[pref.ID]
		if !ok {
			return nil, fmt.Errorf("point %q referenced in speech not found in points/", pref.ID)
		}
		for _, m := range pt.Members {
			if seen[m.ID] {
				continue
			}
			seen[m.ID] = true
			op, exists := opIdx[m.ID]
			if !exists {
				return nil, fmt.Errorf("opinion %q referenced in point %q not found in opinions/", m.ID, pt.ID)
			}
			out = append(out, op)
		}
	}

	// Also include any opinions referenced directly in speech.Opinions.
	for _, oref := range speech.Opinions {
		if seen[oref.ID] {
			continue
		}
		seen[oref.ID] = true
		op, exists := opIdx[oref.ID]
		if !exists {
			return nil, fmt.Errorf("direct opinion %q in speech not found in opinions/", oref.ID)
		}
		out = append(out, op)
	}

	return out, nil
}
