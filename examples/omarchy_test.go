// Package examples_test — TestExampleOmarchy
//
// Loads the full Omarchy north-star composition (134 opinions, 32 points,
// 1 speech targeting vanilla Arch) and asserts that resolve.Resolve returns
// no hard conflicts, a non-empty InstallOrder, and non-empty Applied set
// where every applied opinion ID has the OM- prefix.
//
// Hardware-gated opinions resolve to Skipped on the baseline vanilla-arch
// hardware profile (empty predicates/pci_ids) — this is expected and correct.
//
// This test is the ARCH-02 clean-resolution gate.
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

// loadOmarchySpeech reads and parses examples/omarchy/speech.yaml.
func loadOmarchySpeech(t *testing.T) *resolver.Speech {
	t.Helper()
	root := findRoot(t)
	f, err := os.Open(filepath.Join(root, "examples", "omarchy", "speech.yaml"))
	if err != nil {
		t.Fatalf("loadOmarchySpeech: open speech.yaml: %v", err)
	}
	defer f.Close()
	sp, err := parse.ParseSpeech(f)
	if err != nil {
		t.Fatalf("loadOmarchySpeech: parse speech.yaml: %v", err)
	}
	return sp
}

// loadOmarchyPoints reads all point YAML files from examples/omarchy/points/
// and returns a map from point ID → resolver.Point.
func loadOmarchyPoints(t *testing.T) map[string]resolver.Point {
	t.Helper()
	root := findRoot(t)
	dir := filepath.Join(root, "examples", "omarchy", "points")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("loadOmarchyPoints: readdir: %v", err)
	}
	out := make(map[string]resolver.Point, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		f, err := os.Open(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("loadOmarchyPoints: open %s: %v", e.Name(), err)
		}
		var pt resolver.Point
		if err := yaml.NewDecoder(f).Decode(&pt); err != nil {
			f.Close()
			t.Fatalf("loadOmarchyPoints: decode %s: %v", e.Name(), err)
		}
		f.Close()
		out[pt.ID] = pt
	}
	return out
}

// loadOmarchyOpinions reads all OM-*.yaml files from examples/omarchy/opinions/
// and returns them indexed by opinion ID.
func loadOmarchyOpinions(t *testing.T) map[resolver.OpinionID]resolver.Opinion {
	t.Helper()
	root := findRoot(t)
	dir := filepath.Join(root, "examples", "omarchy", "opinions")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("loadOmarchyOpinions: readdir: %v", err)
	}
	out := make(map[resolver.OpinionID]resolver.Opinion, 134)
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "OM-") || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		f, err := os.Open(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("loadOmarchyOpinions: open %s: %v", e.Name(), err)
		}
		op, err := parse.ParseOpinion(f)
		f.Close()
		if err != nil {
			t.Fatalf("loadOmarchyOpinions: parse %s: %v", e.Name(), err)
		}
		out[op.ID] = *op
	}
	return out
}

// assembleOpinionsFromSpeech builds the flat []resolver.Opinion that resolve.Resolve
// expects by expanding the speech's point references through the point files.
// Ordering: opinions appear in the order the points are listed in the speech,
// and within each point in the order members are listed.  Duplicates across
// points are silently deduplicated (stable first-occurrence wins).
func assembleOpinionsFromSpeech(
	t *testing.T,
	speech *resolver.Speech,
	points map[string]resolver.Point,
	opIdx map[resolver.OpinionID]resolver.Opinion,
) []resolver.Opinion {
	t.Helper()
	seen := make(map[resolver.OpinionID]bool)
	var out []resolver.Opinion
	for _, pref := range speech.Points {
		pt, ok := points[pref.ID]
		if !ok {
			t.Fatalf("assembleOpinionsFromSpeech: point %q referenced in speech not found in points/", pref.ID)
		}
		for _, m := range pt.Members {
			if seen[m.ID] {
				continue
			}
			seen[m.ID] = true
			op, exists := opIdx[m.ID]
			if !exists {
				t.Fatalf("assembleOpinionsFromSpeech: opinion %q referenced in point %q not found in opinions/", m.ID, pt.ID)
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
			t.Fatalf("assembleOpinionsFromSpeech: direct opinion %q in speech not found in opinions/", oref.ID)
		}
		out = append(out, op)
	}
	return out
}

// TestExampleOmarchy loads the full Omarchy north-star composition and asserts
// that it resolves with no hard conflicts on a vanilla-arch baseline profile.
func TestExampleOmarchy(t *testing.T) {
	speech := loadOmarchySpeech(t)
	points := loadOmarchyPoints(t)
	opIdx := loadOmarchyOpinions(t)

	// Sanity: speech must reference 32 points.
	if len(speech.Points) != 32 {
		t.Errorf("TestExampleOmarchy: expected speech to reference 32 points, got %d", len(speech.Points))
	}

	// Sanity: 134 opinion files loaded.
	if len(opIdx) != 134 {
		t.Errorf("TestExampleOmarchy: expected 134 opinion files, got %d", len(opIdx))
	}

	opinions := assembleOpinionsFromSpeech(t, speech, points, opIdx)

	// Build a baseline vanilla-arch hardware profile: no special hardware.
	// Hardware-gated opinions will land in Skipped (expected).
	hw := hardware.HardwareProfile{}
	if speech.Hardware != nil {
		hw.Predicates = speech.Hardware.Predicates
		hw.Facts = speech.Hardware.Facts
		hw.PCIIDs = speech.Hardware.PCIIDs
	}

	rs, err := resolve.Resolve(speech, opinions, hw)
	if err != nil {
		// A non-nil error means a hard conflict or cycle was detected.
		// The partial ResolvedSpeech may carry the explanation — surface it.
		if rs != nil {
			for _, ex := range rs.Explanations {
				if strings.Contains(ex.Text, "Hard conflict") {
					t.Errorf("TestExampleOmarchy: hard conflict detected: %s", ex.Text)
				}
			}
		}
		t.Fatalf("TestExampleOmarchy: resolve returned unexpected error: %v", err)
	}
	if rs == nil {
		t.Fatal("TestExampleOmarchy: Resolve returned nil *ResolvedSpeech")
	}

	// Assert no hard conflict in explanations (ARCH-02 gate).
	for _, ex := range rs.Explanations {
		if strings.Contains(ex.Text, "Hard conflict") {
			t.Errorf("TestExampleOmarchy: hard conflict in explanations: %s", ex.Text)
		}
	}

	// Assert non-empty install order.
	if len(rs.InstallOrder) == 0 {
		t.Error("TestExampleOmarchy: InstallOrder is empty — expected at least one opinion in install order")
	}

	// Assert non-empty Applied set.
	if len(rs.Applied) == 0 {
		t.Error("TestExampleOmarchy: Applied is empty — expected at least some opinions to be applied")
	}

	// Assert every applied opinion has OM- prefix.
	for _, id := range rs.Applied {
		if !strings.HasPrefix(string(id), "OM-") {
			t.Errorf("TestExampleOmarchy: Applied opinion %q does not have OM- prefix", id)
		}
	}

	// Hardware-gated opinions on a vanilla-arch profile land in Skipped (not error).
	// Spot-check: OM-068 (NVIDIA GPU) must be in Skipped, not Applied, since we
	// declared no PCI IDs.
	nvidiaInSkipped := false
	for _, id := range rs.Skipped {
		if id == "OM-068" {
			nvidiaInSkipped = true
			break
		}
	}
	if !nvidiaInSkipped {
		// OM-068 may also have been dropped (rule 1) if it conflicted; just ensure
		// it is not in Applied on a baseline profile.
		for _, id := range rs.Applied {
			if id == "OM-068" {
				t.Error("TestExampleOmarchy: OM-068 (NVIDIA GPU driver) is in Applied on a vanilla-arch baseline — expected Skipped")
			}
		}
	}

	t.Logf("TestExampleOmarchy: Applied=%d Skipped=%d Dropped=%d InstallOrder=%d",
		len(rs.Applied), len(rs.Skipped), len(rs.Dropped), len(rs.InstallOrder))
}
