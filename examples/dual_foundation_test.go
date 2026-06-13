// Package examples_test — TestExampleDualFoundation
//
// Loads the dual-foundation representative speech (examples/dual-foundation/)
// and asserts that resolve.Resolve returns no hard conflicts, a non-empty
// InstallOrder, and all 5 required opinions applied on a baseline hardware
// profile (empty predicates/pci_ids).
//
// This test is the DEB-02 clean-resolution gate: one foundation-neutral speech
// that proves the DebateOS abstraction works across foundations.
//
// License: CC0-1.0 (examples/LICENSE).
package examples_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	yaml "go.yaml.in/yaml/v3"

	"github.com/mikl0s/debateos/resolver"
	"github.com/mikl0s/debateos/resolver/hardware"
	"github.com/mikl0s/debateos/resolver/parse"
	"github.com/mikl0s/debateos/resolver/resolve"
)

// loadDualFoundationSpeech reads and parses examples/dual-foundation/speech.yaml.
func loadDualFoundationSpeech(t *testing.T) *resolver.Speech {
	t.Helper()
	root := findRoot(t)
	f, err := os.Open(filepath.Join(root, "examples", "dual-foundation", "speech.yaml"))
	if err != nil {
		t.Fatalf("loadDualFoundationSpeech: open speech.yaml: %v", err)
	}
	defer f.Close()
	sp, err := parse.ParseSpeech(f)
	if err != nil {
		t.Fatalf("loadDualFoundationSpeech: parse speech.yaml: %v", err)
	}
	return sp
}

// loadDualFoundationPoints reads all point YAML files from
// examples/dual-foundation/points/ and returns a map from point ID → resolver.Point.
func loadDualFoundationPoints(t *testing.T) map[string]resolver.Point {
	t.Helper()
	root := findRoot(t)
	dir := filepath.Join(root, "examples", "dual-foundation", "points")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("loadDualFoundationPoints: readdir: %v", err)
	}
	out := make(map[string]resolver.Point, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		f, err := os.Open(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("loadDualFoundationPoints: open %s: %v", e.Name(), err)
		}
		var pt resolver.Point
		if err := yaml.NewDecoder(f).Decode(&pt); err != nil {
			f.Close()
			t.Fatalf("loadDualFoundationPoints: decode %s: %v", e.Name(), err)
		}
		f.Close()
		out[pt.ID] = pt
	}
	return out
}

// loadDualFoundationOpinions reads all DF-*.yaml files from
// examples/dual-foundation/opinions/ and returns them indexed by opinion ID.
func loadDualFoundationOpinions(t *testing.T) map[resolver.OpinionID]resolver.Opinion {
	t.Helper()
	root := findRoot(t)
	dir := filepath.Join(root, "examples", "dual-foundation", "opinions")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("loadDualFoundationOpinions: readdir: %v", err)
	}
	out := make(map[resolver.OpinionID]resolver.Opinion, 5)
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "DF-") || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		f, err := os.Open(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("loadDualFoundationOpinions: open %s: %v", e.Name(), err)
		}
		op, err := parse.ParseOpinion(f)
		f.Close()
		if err != nil {
			t.Fatalf("loadDualFoundationOpinions: parse %s: %v", e.Name(), err)
		}
		out[op.ID] = *op
	}
	return out
}

// TestExampleDualFoundation loads the dual-foundation representative composition
// and asserts that it resolves clean on a baseline hardware profile:
//   - no hard conflicts
//   - all 5 required opinions applied (len(Applied) == 5, len(InstallOrder) == 5)
//   - zero dropped, zero skipped
//   - every applied opinion ID has the "DF-" prefix
//
// This is the DEB-02 resolve-once gate (foundation-neutral proof).
func TestExampleDualFoundation(t *testing.T) {
	speech := loadDualFoundationSpeech(t)
	points := loadDualFoundationPoints(t)
	opIdx := loadDualFoundationOpinions(t)

	// Sanity: speech must reference exactly 2 points.
	if len(speech.Points) != 2 {
		t.Errorf("TestExampleDualFoundation: expected speech to reference 2 points, got %d", len(speech.Points))
	}

	// Sanity: 5 opinion files loaded.
	if len(opIdx) != 5 {
		t.Errorf("TestExampleDualFoundation: expected 5 opinion files, got %d", len(opIdx))
	}

	opinions := assembleOpinionsFromSpeech(t, speech, points, opIdx)

	// Build a baseline hardware profile: no special hardware.
	// All 5 opinions have no hardware_condition so they must all apply.
	hw := hardware.HardwareProfile{}
	if speech.Hardware != nil {
		hw.Predicates = speech.Hardware.Predicates
		hw.Facts = speech.Hardware.Facts
		hw.PCIIDs = speech.Hardware.PCIIDs
	}

	rs, err := resolve.Resolve(speech, opinions, hw)
	if err != nil {
		// A non-nil error means a hard conflict or cycle was detected.
		// Surface the explanation text for easier debugging.
		if rs != nil {
			for _, ex := range rs.Explanations {
				if strings.Contains(ex.Text, "Hard conflict") {
					t.Errorf("TestExampleDualFoundation: hard conflict detected: %s", ex.Text)
				}
			}
		}
		t.Fatalf("TestExampleDualFoundation: resolve returned unexpected error: %v", err)
	}
	if rs == nil {
		t.Fatal("TestExampleDualFoundation: Resolve returned nil *ResolvedSpeech")
	}

	// Assert no hard conflict in explanations (DEB-02 gate).
	for _, ex := range rs.Explanations {
		if strings.Contains(ex.Text, "Hard conflict") {
			t.Errorf("TestExampleDualFoundation: hard conflict in explanations: %s", ex.Text)
		}
	}

	// Assert all 5 opinions applied (no hardware gates, no conflicts).
	if len(rs.Applied) != 5 {
		t.Errorf("TestExampleDualFoundation: expected 5 applied opinions, got %d (applied=%v)", len(rs.Applied), rs.Applied)
	}

	// Assert install order has all 5.
	if len(rs.InstallOrder) != 5 {
		t.Errorf("TestExampleDualFoundation: expected InstallOrder length 5, got %d", len(rs.InstallOrder))
	}

	// Assert zero dropped.
	if len(rs.Dropped) != 0 {
		t.Errorf("TestExampleDualFoundation: expected 0 dropped opinions, got %d: %v", len(rs.Dropped), rs.Dropped)
	}

	// Assert zero skipped.
	if len(rs.Skipped) != 0 {
		t.Errorf("TestExampleDualFoundation: expected 0 skipped opinions, got %d: %v", len(rs.Skipped), rs.Skipped)
	}

	// Assert every applied opinion has DF- prefix.
	for _, id := range rs.Applied {
		if !strings.HasPrefix(string(id), "DF-") {
			t.Errorf("TestExampleDualFoundation: Applied opinion %q does not have DF- prefix", id)
		}
	}

	t.Logf("TestExampleDualFoundation: Applied=%d Skipped=%d Dropped=%d InstallOrder=%d",
		len(rs.Applied), len(rs.Skipped), len(rs.Dropped), len(rs.InstallOrder))
}
