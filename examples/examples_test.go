// Package examples_test exercises the four example compositions end-to-end
// through parse → resolve.  Each example is a speech.yaml + opinions.yaml pair
// in a subdirectory of examples/.  These tests encode RSLV-06 for the examples
// and serve as the human-readable demonstration (Invariant 3) that a person can
// understand a composition and its resolution from the YAML alone.
//
// License for examples/: CC0-1.0 (examples/LICENSE).
package examples_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"

	"github.com/mikl0s/debateos/resolver"
	"github.com/mikl0s/debateos/resolver/hardware"
	"github.com/mikl0s/debateos/resolver/resolve"
)

// ─── helpers ──────────────────────────────────────────────────────────────

// loadExample reads speech.yaml and opinions.yaml from examples/<name>/ and
// returns the typed structs ready for resolve.Resolve.
func loadExample(t *testing.T, name string) ([]resolver.Opinion, resolver.Speech, hardware.HardwareProfile) {
	t.Helper()

	// Determine examples directory: this file lives in examples/, so its parent
	// is the module root.
	root := findRoot(t)
	dir := filepath.Join(root, "examples", name)

	speechRaw, err := os.ReadFile(filepath.Join(dir, "speech.yaml"))
	if err != nil {
		t.Fatalf("loadExample(%s): cannot read speech.yaml: %v", name, err)
	}
	opRaw, err := os.ReadFile(filepath.Join(dir, "opinions.yaml"))
	if err != nil {
		t.Fatalf("loadExample(%s): cannot read opinions.yaml: %v", name, err)
	}

	var speech resolver.Speech
	if err := yaml.Unmarshal(speechRaw, &speech); err != nil {
		t.Fatalf("loadExample(%s): speech YAML error: %v", name, err)
	}

	var opinions []resolver.Opinion
	if err := yaml.Unmarshal(opRaw, &opinions); err != nil {
		t.Fatalf("loadExample(%s): opinions YAML error: %v", name, err)
	}

	hw := hardware.HardwareProfile{}
	if speech.Hardware != nil {
		hw.Predicates = speech.Hardware.Predicates
		hw.Facts = speech.Hardware.Facts
	}
	return opinions, speech, hw
}

// findRoot walks up from the test working directory until go.mod is found.
func findRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("findRoot: getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("findRoot: could not locate go.mod")
		}
		dir = parent
	}
}

// ─── Test: omarchy-mini ────────────────────────────────────────────────────

// TestExampleOmarchyMini loads the omarchy-mini composition (a coherent subset
// of real OM-NNN Omarchy opinions) and asserts that it resolves cleanly with a
// non-empty InstallOrder and Explanations.
func TestExampleOmarchyMini(t *testing.T) {
	opinions, speech, hw := loadExample(t, "omarchy-mini")

	rs, err := resolve.Resolve(&speech, opinions, hw)
	if err != nil {
		t.Fatalf("omarchy-mini: unexpected resolve error: %v", err)
	}
	if rs == nil {
		t.Fatal("omarchy-mini: Resolve returned nil *ResolvedSpeech")
	}
	if len(rs.InstallOrder) == 0 {
		t.Error("omarchy-mini: InstallOrder is empty — expected at least one opinion in install order")
	}
	if len(rs.Explanations) == 0 {
		t.Error("omarchy-mini: Explanations is empty — expected at least one explanation")
	}
	if len(rs.Applied) == 0 {
		t.Error("omarchy-mini: Applied is empty — expected opinions to be applied")
	}
	// All applied opinions must reference real OM-NNN IDs.
	for _, id := range rs.Applied {
		if !strings.HasPrefix(string(id), "OM-") {
			t.Errorf("omarchy-mini: Applied opinion %q does not have OM-NNN prefix", id)
		}
	}
}

// ─── Test: two-point-clean ─────────────────────────────────────────────────

// TestExampleTwoPointClean loads the two-point-clean composition (two
// non-conflicting opinions) and asserts that all opinions are Applied with no
// conflicts.
func TestExampleTwoPointClean(t *testing.T) {
	opinions, speech, hw := loadExample(t, "two-point-clean")

	rs, err := resolve.Resolve(&speech, opinions, hw)
	if err != nil {
		t.Fatalf("two-point-clean: unexpected resolve error: %v", err)
	}
	if rs == nil {
		t.Fatal("two-point-clean: Resolve returned nil *ResolvedSpeech")
	}
	if len(rs.Dropped) != 0 {
		t.Errorf("two-point-clean: expected no dropped opinions; got %v", rs.Dropped)
	}
	if len(rs.Applied) != len(opinions) {
		t.Errorf("two-point-clean: expected all %d opinions applied; got %d applied (%v)",
			len(opinions), len(rs.Applied), rs.Applied)
	}
	// All Explanations should be rule="no-conflict" (no rule1/2/3/4 firings).
	for _, ex := range rs.Explanations {
		if ex.Rule == "rule1" || ex.Rule == "rule2" || ex.Rule == "rule3" || ex.Rule == "rule4" {
			t.Errorf("two-point-clean: unexpected conflict rule %q in explanation: %s", ex.Rule, ex.Text)
		}
	}
}

// ─── Test: conflicting ─────────────────────────────────────────────────────

// TestExampleConflicting loads the deliberately conflicting composition and
// asserts that resolution surfaces a hard conflict — either as a non-nil error
// or via an Explanation containing "Hard conflict".
func TestExampleConflicting(t *testing.T) {
	opinions, speech, hw := loadExample(t, "conflicting")

	rs, err := resolve.Resolve(&speech, opinions, hw)

	// Either an error is returned OR the ResolvedSpeech contains a "Hard conflict"
	// explanation.  Both signal that the conflict was surfaced, not silenced.
	surfaced := false

	if err != nil && strings.Contains(err.Error(), "Hard conflict") {
		surfaced = true
	}
	if rs != nil {
		for _, ex := range rs.Explanations {
			if strings.Contains(ex.Text, "Hard conflict") {
				surfaced = true
				break
			}
		}
	}

	if !surfaced {
		t.Error("conflicting: expected a hard conflict to be surfaced (error or Explanation.Text containing 'Hard conflict')")
	}
}

// ─── Test: hardware-conditional ───────────────────────────────────────────

// TestExampleHardwareConditional verifies the hardware-conditional composition
// in two scenarios:
//  1. Matching hardware profile: the gated opinion appears in Applied.
//  2. Non-matching hardware profile: the gated opinion appears in Skipped.
func TestExampleHardwareConditional(t *testing.T) {
	opinions, speech, _ := loadExample(t, "hardware-conditional")

	// Find the hardware-gated opinion ID by looking for one with a HardwareCondition.
	var gatedID resolver.OpinionID
	for _, op := range opinions {
		if op.HardwareCondition != nil {
			gatedID = op.ID
			break
		}
	}
	if gatedID == "" {
		t.Fatal("hardware-conditional: no opinion with hardware_condition found in opinions.yaml")
	}

	t.Run("matching-hardware", func(t *testing.T) {
		// Use the speech as written (its hardware profile should match the condition).
		hw := hardware.HardwareProfile{}
		if speech.Hardware != nil {
			hw.Predicates = speech.Hardware.Predicates
			hw.Facts = speech.Hardware.Facts
		}
		// Add the PCI ID for NVIDIA to the profile so the condition is true.
		hw.PCIIDs = []string{"10de:2204"} // NVIDIA RTX 3090 — Turing+

		rs, err := resolve.Resolve(&speech, opinions, hw)
		if err != nil {
			t.Fatalf("matching-hardware: unexpected resolve error: %v", err)
		}
		found := false
		for _, id := range rs.Applied {
			if id == gatedID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("matching-hardware: gated opinion %q not in Applied; applied=%v skipped=%v",
				gatedID, rs.Applied, rs.Skipped)
		}
	})

	t.Run("non-matching-hardware", func(t *testing.T) {
		// Use an empty hardware profile — condition will be false.
		hw := hardware.HardwareProfile{}

		rs, err := resolve.Resolve(&speech, opinions, hw)
		if err != nil {
			t.Fatalf("non-matching-hardware: unexpected resolve error: %v", err)
		}
		found := false
		for _, id := range rs.Skipped {
			if id == gatedID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("non-matching-hardware: gated opinion %q not in Skipped; skipped=%v applied=%v",
				gatedID, rs.Skipped, rs.Applied)
		}
		// Check explanation text
		skipText := false
		for _, ex := range rs.Explanations {
			if strings.Contains(ex.Text, string(gatedID)) && strings.Contains(ex.Text, "hardware condition") {
				skipText = true
				break
			}
		}
		if !skipText {
			t.Errorf("non-matching-hardware: expected a hardware-condition explanation for %q", gatedID)
		}
	})
}
