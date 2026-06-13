// Package loader provides a shared speech-loading pipeline for the debateos
// CLI subcommands. It replicates the loading logic from cmd/resolve-json/main.go
// so compose and validate can share the same assembly steps without duplicating
// them or causing the cmd/resolve-json pipeline to drift.
//
// The exported ResolveDir function is the primary entry point:
//
//	rs, err := loader.ResolveDir(speechDir)
package loader

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

// ResolveDir loads a speech directory (speech.yaml + points/ + opinions/) and
// runs resolve.Resolve, returning the resolved speech. On hard conflict the
// partial ResolvedSpeech is returned alongside a non-nil error.
func ResolveDir(dir string) (*resolve.ResolvedSpeech, error) {
	// 1. Load speech.yaml.
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

	// 2. Load points/.
	points, err := loadPoints(filepath.Join(dir, "points"))
	if err != nil {
		return nil, fmt.Errorf("load points: %w", err)
	}

	// 3. Load opinions/.
	opIdx, err := loadOpinions(filepath.Join(dir, "opinions"))
	if err != nil {
		return nil, fmt.Errorf("load opinions: %w", err)
	}

	// 4. Assemble flat []resolver.Opinion in speech order.
	opinions, err := assembleOpinions(speech, points, opIdx)
	if err != nil {
		return nil, fmt.Errorf("assemble opinions: %w", err)
	}

	// 5. Build hardware profile from speech.
	hw := hardware.HardwareProfile{}
	if speech.Hardware != nil {
		hw.Predicates = speech.Hardware.Predicates
		hw.Facts = speech.Hardware.Facts
		hw.PCIIDs = speech.Hardware.PCIIDs
	}

	// 6. Resolve.
	rs, err := resolve.Resolve(speech, opinions, hw)
	if err != nil {
		return rs, fmt.Errorf("resolve: %w", err)
	}

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
		if decErr := yaml.NewDecoder(f).Decode(&pt); decErr != nil {
			f.Close()
			return nil, fmt.Errorf("decode %s: %w", e.Name(), decErr)
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
		op, parseErr := parse.ParseOpinion(f)
		f.Close()
		if parseErr != nil {
			return nil, fmt.Errorf("parse %s: %w", e.Name(), parseErr)
		}
		out[op.ID] = *op
	}
	return out, nil
}

// assembleOpinions builds the flat []resolver.Opinion by expanding speech.Points
// through the loaded point files, deduplicating by opinion ID (stable first-occurrence).
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
