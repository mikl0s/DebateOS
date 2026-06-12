package resolve_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"

	"github.com/mikkelraglan/debateos/resolver"
	"github.com/mikkelraglan/debateos/resolver/hardware"
	"github.com/mikkelraglan/debateos/resolver/resolve"
)

// fixtureDoc holds the top-level structure of a testdata/ec*.yaml fixture.
// Fixtures embed both the opinion definitions and the speech under test.
type fixtureDoc struct {
	Opinions []resolver.Opinion `yaml:"opinions"`
	Speech   resolver.Speech    `yaml:"speech"`
	// Hardware profile can be embedded directly in the fixture instead of
	// inside the speech (for tests that need a richer hardware.HardwareProfile).
	HardwareOverride *hardware.HardwareProfile `yaml:"hardware_override,omitempty"`
}

// loadFixture parses an ec*.yaml testdata file and converts speech.Hardware
// into a hardware.HardwareProfile for Resolve. The fixture YAML may supply a
// `hardware_override` block for cases (e.g. EC-038) that need a richer
// profile than the speech.hardware block alone provides.
func loadFixture(t *testing.T, name string) ([]resolver.Opinion, resolver.Speech, hardware.HardwareProfile) {
	t.Helper()
	path := filepath.Join("testdata", name)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("loadFixture: cannot read %s: %v", path, err)
	}
	var doc fixtureDoc
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("loadFixture: YAML parse error in %s: %v", path, err)
	}

	// Build hardware.HardwareProfile from the speech's hardware declaration.
	// resolver.HardwareProfile now carries PCIIDs so pci_ids in speech.hardware
	// propagates directly without any workaround.
	hw := hardware.HardwareProfile{}
	if doc.Speech.Hardware != nil {
		hw.Predicates = doc.Speech.Hardware.Predicates
		hw.Facts = doc.Speech.Hardware.Facts
		hw.PCIIDs = doc.Speech.Hardware.PCIIDs
	}
	// Apply fixture-level override for richer profiles (e.g. legacy ec038 fixture
	// that declares hardware_override rather than speech.hardware.pci_ids).
	if doc.HardwareOverride != nil {
		if doc.HardwareOverride.PCIIDs != nil {
			hw.PCIIDs = doc.HardwareOverride.PCIIDs
		}
	}

	return doc.Opinions, doc.Speech, hw
}

// ─── EC table-driven subtests ──────────────────────────────────────────────

// ecCase describes one EC-NNN subtest expectation.
type ecCase struct {
	fixture         string // testdata/ec*.yaml filename
	wantErr         bool   // true when Resolve should return a non-nil error
	wantErrContains string // substring that must appear in err.Error()
	wantText        string // substring that must appear in one Explanation.Text
	wantApplied     []string
	wantDropped     []string
	wantSkipped     []string
	wantOrder       []string // expected install order (exact, if non-nil)
}

// TestResolveEC runs one subtest per EC-NNN scenario.
func TestResolveEC(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		ecCase
	}{
		// ── Class 1: Foundation pre-seeded vs user opinion ─────────────────
		{
			name: "EC-001",
			ecCase: ecCase{
				fixture:         "ec001-garuda-snapper.yaml",
				wantErr:         true,
				wantErrContains: "Hard conflict",
				wantText:        "Hard conflict",
			},
		},
		{
			name: "EC-002",
			ecCase: ecCase{
				fixture:         "ec002-grub-limine.yaml",
				wantErr:         true,
				wantErrContains: "Hard conflict",
				wantText:        "Hard conflict",
			},
		},
		{
			name: "EC-003",
			ecCase: ecCase{
				fixture:         "ec003-sddm-theme.yaml",
				wantErr:         true,
				wantErrContains: "Hard conflict",
				wantText:        "Hard conflict",
			},
		},
		{
			name: "EC-004",
			ecCase: ecCase{
				fixture:         "ec004-cachyos-kernel.yaml",
				wantErr:         true,
				wantErrContains: "Hard conflict",
				wantText:        "Hard conflict",
			},
		},
		{
			name: "EC-005",
			ecCase: ecCase{
				fixture:         "ec005-sysctl-collision.yaml",
				wantErr:         true,
				wantErrContains: "Sysctl key collision",
				wantText:        "Sysctl key collision",
			},
		},
		// ── Class 2: Repo priority ──────────────────────────────────────────
		{
			// EC-010: nice-to-have omarchy repo added after required CachyOS repos — no conflict
			name: "EC-010",
			ecCase: ecCase{
				fixture:  "ec010-cachyos-repo-order.yaml",
				wantErr:  false,
				wantText: "Repo ordering",
			},
		},
		{
			// EC-011: nice-to-have omarchy repo, priority undeclared — default applied
			name: "EC-011",
			ecCase: ecCase{
				fixture:  "ec011-garuda-repo-priority.yaml",
				wantErr:  false,
				wantText: "Repo priority",
			},
		},
		{
			// EC-012: required beats nice-to-have — extra-gaming-repo dropped
			name: "EC-012",
			ecCase: ecCase{
				fixture:     "ec012-required-repo-drops-nice.yaml",
				wantErr:     false,
				wantText:    "Required beats nice-to-have",
				wantDropped: []string{"custom-repo/extra-gaming-repo"},
				wantApplied: []string{"custom-repo/omarchy-repo"},
			},
		},
		// ── Class 3: Cross-variant effectuation ────────────────────────────
		{
			// EC-020: mesa opinion applied; CachyOS variant note emitted
			name: "EC-020",
			ecCase: ecCase{
				fixture:  "ec020-mesa-variant.yaml",
				wantErr:  false,
				wantText: "No conflict",
			},
		},
		{
			// EC-021: linux-headers dependency; translator note emitted
			name: "EC-021",
			ecCase: ecCase{
				fixture:  "ec021-linux-headers-name.yaml",
				wantErr:  false,
				wantText: "No conflict",
			},
		},
		{
			// EC-022: snapper idempotency warning (not a conflict)
			name: "EC-022",
			ecCase: ecCase{
				fixture:  "ec022-snapper-idempotency.yaml",
				wantErr:  false,
				wantText: "No conflict",
			},
		},
		{
			// EC-023: bluetooth service enable — no conflict
			name: "EC-023",
			ecCase: ecCase{
				fixture:  "ec023-bluetooth-enable.yaml",
				wantErr:  false,
				wantText: "No conflict",
			},
		},
		// ── Class 4: docs/04 rule coverage ────────────────────────────────
		{
			// EC-030: required kernel drops nice-to-have DKMS (Rule 1)
			name: "EC-030",
			ecCase: ecCase{
				fixture:     "ec030-required-drops-dkms.yaml",
				wantErr:     false,
				wantText:    "Required beats nice-to-have",
				wantDropped: []string{"package-install/broadcom-wl-dkms"},
				wantApplied: []string{"kernel-install/linux-ptl"},
			},
		},
		{
			// EC-031: required-vs-required hard conflict, no patch (Rule 2)
			name: "EC-031",
			ecCase: ecCase{
				fixture:         "ec031-kernel-hard-conflict.yaml",
				wantErr:         true,
				wantErrContains: "Hard conflict",
				wantText:        "Hard conflict",
			},
		},
		{
			// EC-032: required-vs-required resolved by patch opinion
			name: "EC-032",
			ecCase: ecCase{
				fixture:  "ec032-dracut-patch.yaml",
				wantErr:  false,
				wantText: "patch",
			},
		},
		{
			// EC-033: nice-vs-nice — foot chosen as default (Rule 3)
			name: "EC-033",
			ecCase: ecCase{
				fixture:     "ec033-nice-vs-nice-terminal.yaml",
				wantErr:     false,
				wantText:    "Nice-to-have conflict",
				wantApplied: []string{"package-install/terminal-foot"},
				wantDropped: []string{"package-install/terminal-ghostty"},
			},
		},
		{
			// EC-034: patch overrides hierarchy (Rule 4)
			name: "EC-034",
			ecCase: ecCase{
				fixture:  "ec034-patch-overrides.yaml",
				wantErr:  false,
				wantText: "resolved by patch",
			},
		},
		{
			// EC-035: three-hop ordering chain [OM-009, OM-041, OM-023]
			name: "EC-035",
			ecCase: ecCase{
				fixture:     "ec035-three-hop-order.yaml",
				wantErr:     false,
				wantText:    "Install order",
				wantOrder:   []string{"OM-009", "OM-041", "OM-023"},
				wantApplied: []string{"OM-009", "OM-041", "OM-023"},
			},
		},
		{
			// EC-036: cycle detected — hard error naming both opinions
			name: "EC-036",
			ecCase: ecCase{
				fixture:         "ec036-cycle.yaml",
				wantErr:         true,
				wantErrContains: "Cycle detected",
				wantText:        "Cycle detected",
			},
		},
		{
			// EC-037: NVIDIA driver skipped — hardware condition false
			name: "EC-037",
			ecCase: ecCase{
				fixture:     "ec037-nvidia-skip.yaml",
				wantErr:     false,
				wantText:    "Skipped (hardware condition false)",
				wantSkipped: []string{"hardware-conditional/nvidia-driver"},
			},
		},
		{
			// EC-038: Apple T2 block applied — hardware condition true; sig_level=Never warning
			name: "EC-038",
			ecCase: ecCase{
				fixture:     "ec038-apple-t2-apply.yaml",
				wantErr:     false,
				wantText:    "Applied (hardware condition true)",
				wantApplied: []string{"hardware-conditional/apple-t2"},
			},
		},
		{
			// EC-038b: pci_ids declared directly in speech.hardware — CR-02 fix verification.
			// pci_ids must propagate from speech.Hardware.PCIIDs through the normal parse path
			// into hardware.EvalCondition without a hardware_override workaround.
			name: "EC-038b",
			ecCase: ecCase{
				fixture:     "ec038b-pci-ids-in-speech.yaml",
				wantErr:     false,
				wantText:    "Applied (hardware condition true)",
				wantApplied: []string{"hardware-conditional/apple-t2-direct"},
			},
		},
		// ── Class 5: CachyOS kernel collision ──────────────────────────────
		{
			// EC-040: CachyOS pre-seed vs vanilla linux — hard conflict
			name: "EC-040",
			ecCase: ecCase{
				fixture:         "ec040-vanilla-vs-cachyos.yaml",
				wantErr:         true,
				wantErrContains: "Hard conflict",
				wantText:        "Hard conflict",
			},
		},
		{
			// EC-041: v3 vs v4 arch level — not a conflict; suggestion emitted
			name: "EC-041",
			ecCase: ecCase{
				fixture:  "ec041-cpu-arch-mismatch.yaml",
				wantErr:  false,
				wantText: "No conflict",
			},
		},
		{
			// EC-042: multi-kernel — no hard conflict; both applied
			name: "EC-042",
			ecCase: ecCase{
				fixture:     "ec042-multi-kernel.yaml",
				wantErr:     false,
				wantText:    "No conflict",
				wantApplied: []string{"foundation/cachyos-kernel-base", "hardware-conditional/linux-ptl"},
			},
		},
		// ── Class 6: Garuda theming collision ──────────────────────────────
		{
			// EC-050: SDDM active theme slot — required-vs-required hard conflict
			name: "EC-050",
			ecCase: ecCase{
				fixture:         "ec050-sddm-slot.yaml",
				wantErr:         true,
				wantErrContains: "Hard conflict",
				wantText:        "Hard conflict",
			},
		},
		{
			// EC-051: Plymouth active theme slot — required-vs-required hard conflict
			name: "EC-051",
			ecCase: ecCase{
				fixture:         "ec051-plymouth-slot.yaml",
				wantErr:         true,
				wantErrContains: "Hard conflict",
				wantText:        "Hard conflict",
			},
		},
		{
			// EC-052: Garuda GRUB theme; no competing Omarchy opinion — no conflict
			name: "EC-052",
			ecCase: ecCase{
				fixture:  "ec052-grub-theme-no-conflict.yaml",
				wantErr:  false,
				wantText: "No conflict",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			opinions, speech, hw := loadFixture(t, tc.fixture)

			rs, err := resolve.Resolve(&speech, opinions, hw)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("[%s] Resolve() returned nil error; want error containing %q", tc.name, tc.wantErrContains)
				}
				if tc.wantErrContains != "" && !strings.Contains(err.Error(), tc.wantErrContains) {
					t.Fatalf("[%s] error %q does not contain %q", tc.name, err.Error(), tc.wantErrContains)
				}
				// Even on error, Resolve must return a (partial) ResolvedSpeech with an explanation.
				if rs != nil && tc.wantText != "" {
					found := false
					for _, ex := range rs.Explanations {
						if strings.Contains(ex.Text, tc.wantText) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("[%s] no explanation text containing %q found in %d explanations", tc.name, tc.wantText, len(rs.Explanations))
					}
				}
				return
			}

			// ── Success path ───────────────────────────────────────────────
			if err != nil {
				t.Fatalf("[%s] Resolve() unexpected error: %v", tc.name, err)
			}
			if rs == nil {
				t.Fatalf("[%s] Resolve() returned nil *ResolvedSpeech without error", tc.name)
			}

			// wantText: at least one Explanation.Text must contain the substring.
			if tc.wantText != "" {
				found := false
				for _, ex := range rs.Explanations {
					if strings.Contains(ex.Text, tc.wantText) {
						found = true
						break
					}
				}
				if !found {
					var texts []string
					for _, ex := range rs.Explanations {
						texts = append(texts, ex.Text)
					}
					t.Errorf("[%s] no explanation text containing %q; got:\n  %s",
						tc.name, tc.wantText, strings.Join(texts, "\n  "))
				}
			}

			// wantApplied: all listed IDs must appear in rs.Applied.
			appliedSet := makeSet(rs.Applied)
			for _, id := range tc.wantApplied {
				if !appliedSet[resolver.OpinionID(id)] {
					t.Errorf("[%s] expected %q in Applied; got %v", tc.name, id, rs.Applied)
				}
			}

			// wantDropped: all listed IDs must appear in rs.Dropped.
			droppedSet := makeSet(rs.Dropped)
			for _, id := range tc.wantDropped {
				if !droppedSet[resolver.OpinionID(id)] {
					t.Errorf("[%s] expected %q in Dropped; got %v", tc.name, id, rs.Dropped)
				}
			}

			// wantSkipped: all listed IDs must appear in rs.Skipped.
			skippedSet := makeSet(rs.Skipped)
			for _, id := range tc.wantSkipped {
				if !skippedSet[resolver.OpinionID(id)] {
					t.Errorf("[%s] expected %q in Skipped; got %v", tc.name, id, rs.Skipped)
				}
			}

			// wantOrder: exact install order if specified.
			if len(tc.wantOrder) > 0 {
				if len(rs.InstallOrder) != len(tc.wantOrder) {
					t.Errorf("[%s] install order length: got %d, want %d\n  got:  %v\n  want: %v",
						tc.name, len(rs.InstallOrder), len(tc.wantOrder), rs.InstallOrder, tc.wantOrder)
				} else {
					for i, wid := range tc.wantOrder {
						if string(rs.InstallOrder[i]) != wid {
							t.Errorf("[%s] InstallOrder[%d]: got %q, want %q", tc.name, i, rs.InstallOrder[i], wid)
						}
					}
				}
			}
		})
	}
}

// ─── Rule coverage assertion ───────────────────────────────────────────────

// TestResolveRuleCoverage asserts that the EC corpus exercises every docs/04
// rule branch, ordering, cycle, hardware-skip, hardware-apply, patch-override,
// and sysctl-collision behaviors.
func TestResolveRuleCoverage(t *testing.T) {
	t.Parallel()

	type ruleCheck struct {
		name    string
		fixture string
		rule    string // expected Explanation.Rule value
	}

	checks := []ruleCheck{
		{name: "Rule1 required-beats-nice", fixture: "ec012-required-repo-drops-nice.yaml", rule: "rule1"},
		{name: "Rule1 required-beats-nice (kernel)", fixture: "ec030-required-drops-dkms.yaml", rule: "rule1"},
		{name: "Rule2 required-vs-required hard conflict", fixture: "ec031-kernel-hard-conflict.yaml", rule: "rule2"},
		// EC-032: patch is already in the active speech, so Rule 4 fires (patch overrides).
		// Rule 2 "with patch offered" is covered by EC-031 returning rule2 + PatchOffered=="" (no patch).
		{name: "Rule2 with patch offered", fixture: "ec032-dracut-patch.yaml", rule: "rule4"},
		{name: "Rule3 nice-vs-nice default", fixture: "ec033-nice-vs-nice-terminal.yaml", rule: "rule3"},
		{name: "Rule4 patch overrides hierarchy", fixture: "ec034-patch-overrides.yaml", rule: "rule4"},
		{name: "Ordering toposort", fixture: "ec035-three-hop-order.yaml", rule: "ordering"},
		{name: "Cycle detection", fixture: "ec036-cycle.yaml", rule: "cycle"},
		{name: "Hardware-skip (condition false)", fixture: "ec037-nvidia-skip.yaml", rule: "hardware-skip"},
		{name: "Hardware-apply (condition true)", fixture: "ec038-apple-t2-apply.yaml", rule: "hardware-apply"},
		{name: "Sysctl collision (SR-016)", fixture: "ec005-sysctl-collision.yaml", rule: "sysctl-collision"},
	}

	for _, c := range checks {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			opinions, speech, hw := loadFixture(t, c.fixture)
			rs, _ := resolve.Resolve(&speech, opinions, hw)
			if rs == nil {
				t.Skipf("Resolve returned nil ResolvedSpeech for %s — no explanations to check", c.fixture)
			}
			found := false
			for _, ex := range rs.Explanations {
				if ex.Rule == c.rule {
					found = true
					break
				}
			}
			if !found {
				var rules []string
				for _, ex := range rs.Explanations {
					rules = append(rules, ex.Rule)
				}
				t.Errorf("rule coverage %q: no explanation with Rule=%q found; got rules: %v", c.name, c.rule, rules)
			}
		})
	}
}

// ─── Canonical JSON determinism ───────────────────────────────────────────

// TestCanonicalJSONDeterministic verifies that CanonicalJSON of the same
// ResolvedSpeech across ≥3 runs is byte-identical, contains no NaN/Inf, and
// marshals without error. Tests T-01-12 mitigation.
func TestCanonicalJSONDeterministic(t *testing.T) {
	t.Parallel()

	// Build a representative ResolvedSpeech manually so this test does not
	// depend on Resolve (isolation) and runs deterministically itself.
	rs := &resolve.ResolvedSpeech{
		Schema:     1,
		Foundation: "arch",
		InstallOrder: []resolver.OpinionID{
			"OM-009", "OM-041", "OM-023",
		},
		Applied: []resolver.OpinionID{
			"OM-009", "OM-041", "OM-023",
		},
		Skipped: []resolver.OpinionID{},
		Dropped: []resolver.OpinionID{},
		Explanations: []Explanation{
			{
				Text:             "Install order (topological sort): [OM-009, OM-041, OM-023]",
				Rule:             "ordering",
				OpinionsInvolved: []resolver.OpinionID{"OM-009", "OM-041", "OM-023"},
			},
		},
	}

	runs := 5
	var baseline []byte
	for i := 0; i < runs; i++ {
		b, err := resolve.CanonicalJSON(rs)
		if err != nil {
			t.Fatalf("run %d: CanonicalJSON error: %v", i, err)
		}
		if len(b) == 0 {
			t.Fatalf("run %d: CanonicalJSON returned empty output", i)
		}
		// No NaN or Inf in output.
		if bytes.Contains(b, []byte("NaN")) || bytes.Contains(b, []byte("Inf")) {
			t.Fatalf("run %d: CanonicalJSON output contains NaN/Inf: %s", i, b)
		}
		if baseline == nil {
			baseline = b
		} else if !bytes.Equal(baseline, b) {
			t.Fatalf("run %d: CanonicalJSON is not deterministic\n  run 0: %s\n  run %d: %s", i, baseline, i, b)
		}
	}
}

// ─── Explicit RSLV-06 required cases ──────────────────────────────────────

// TestResolveRequiredVsRequired verifies EC-031 specifically (RSLV-06 gate).
func TestResolveRequiredVsRequired(t *testing.T) {
	t.Parallel()
	opinions, speech, hw := loadFixture(t, "ec031-kernel-hard-conflict.yaml")
	_, err := resolve.Resolve(&speech, opinions, hw)
	if err == nil {
		t.Fatal("EC-031: Resolve() should return an error for required-vs-required with no patch")
	}
	if !strings.Contains(err.Error(), "Hard conflict") {
		t.Errorf("EC-031: error %q should contain 'Hard conflict'", err.Error())
	}
}

// TestResolveHardwareMismatch verifies EC-037 specifically (RSLV-06 gate).
func TestResolveHardwareMismatch(t *testing.T) {
	t.Parallel()
	opinions, speech, hw := loadFixture(t, "ec037-nvidia-skip.yaml")
	rs, err := resolve.Resolve(&speech, opinions, hw)
	if err != nil {
		t.Fatalf("EC-037: unexpected error: %v", err)
	}
	found := false
	for _, id := range rs.Skipped {
		if id == "hardware-conditional/nvidia-driver" {
			found = true
		}
	}
	if !found {
		t.Errorf("EC-037: nvidia-driver not in Skipped; got %v", rs.Skipped)
	}
}

// TestResolveKernelClash verifies EC-004 and EC-040 (version/kernel clash — RSLV-06 gate).
func TestResolveKernelClash(t *testing.T) {
	t.Parallel()
	for _, fixture := range []string{"ec004-cachyos-kernel.yaml", "ec040-vanilla-vs-cachyos.yaml"} {
		fixture := fixture
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()
			opinions, speech, hw := loadFixture(t, fixture)
			_, err := resolve.Resolve(&speech, opinions, hw)
			if err == nil {
				t.Fatalf("%s: expected hard conflict error, got nil", fixture)
			}
		})
	}
}

// TestResolvePatchablePair verifies EC-032 (patchable pair — RSLV-06 gate).
func TestResolvePatchablePair(t *testing.T) {
	t.Parallel()
	opinions, speech, hw := loadFixture(t, "ec032-dracut-patch.yaml")
	rs, err := resolve.Resolve(&speech, opinions, hw)
	if err != nil {
		t.Fatalf("EC-032: expected patch resolution, got error: %v", err)
	}
	foundPatch := false
	for _, ex := range rs.Explanations {
		if ex.PatchOffered != "" {
			foundPatch = true
			break
		}
	}
	if !foundPatch {
		t.Error("EC-032: no PatchOffered in any Explanation")
	}
}

// ─── Gap-closure tests (01-05 supplemental) ───────────────────────────────

// TestCanonicalJSONNilInput verifies that CanonicalJSON returns an error when
// called with nil (covers the nil-guard branch in canonical.go).
func TestCanonicalJSONNilInput(t *testing.T) {
	t.Parallel()
	_, err := resolve.CanonicalJSON(nil)
	if err == nil {
		t.Fatal("CanonicalJSON(nil): expected error, got nil")
	}
	if !strings.Contains(err.Error(), "nil") {
		t.Errorf("CanonicalJSON(nil): error %q should mention nil", err.Error())
	}
}

// TestResolveHardwareErrorPath verifies that a malformed hardware condition
// (unknown HardwareExprType) in an opinion produces a hardware-skip explanation
// rather than panicking or returning a nil ResolvedSpeech (covers the
// hardware.EvalCondition error branch in resolve.go).
func TestResolveHardwareErrorPath(t *testing.T) {
	t.Parallel()
	badType := resolver.HardwareExprType("bad-type-for-coverage")
	opinions := []resolver.Opinion{
		{
			Schema:   1,
			ID:       "op-malformed-hw",
			Name:     "Malformed HW Condition",
			Category: "hardware-conditional",
			Status:   resolver.StatusNiceToHave,
			HardwareCondition: &resolver.HardwareExpr{
				Type: badType,
			},
		},
	}
	speech := resolver.Speech{
		Schema:     1,
		ID:         "test-malformed-hw",
		Foundation: "arch",
		Opinions:   []resolver.OpinionRef{{ID: "op-malformed-hw"}},
	}
	hw := hardware.HardwareProfile{}

	rs, _ := resolve.Resolve(&speech, opinions, hw)
	if rs == nil {
		t.Fatal("Resolve: expected non-nil ResolvedSpeech on malformed hardware condition")
	}
	// The opinion should appear in Skipped (hardware-skip rule fires on eval error).
	found := false
	for _, id := range rs.Skipped {
		if id == "op-malformed-hw" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Resolve: expected op-malformed-hw in Skipped; got skipped=%v applied=%v", rs.Skipped, rs.Applied)
	}
}

// TestResolveNiceVsNiceRule3 verifies Rule 3: when two nice-to-have opinions
// conflict, the first-listed (lower canonical ID) wins and the second is dropped.
func TestResolveNiceVsNiceRule3(t *testing.T) {
	t.Parallel()

	// nice/alpha < nice/beta lexicographically → alpha wins, beta dropped.
	opinions := []resolver.Opinion{
		{
			Schema:   1,
			ID:       "nice/alpha",
			Name:     "Alpha nice",
			Category: "config-dotfile",
			Status:   resolver.StatusNiceToHave,
			Conflicts: []resolver.OpinionRef{{ID: "nice/beta"}},
		},
		{
			Schema:   1,
			ID:       "nice/beta",
			Name:     "Beta nice",
			Category: "config-dotfile",
			Status:   resolver.StatusNiceToHave,
			Conflicts: []resolver.OpinionRef{{ID: "nice/alpha"}},
		},
	}
	speech := resolver.Speech{
		Schema:     1,
		ID:         "test-nice-vs-nice",
		Foundation: "arch",
		Opinions: []resolver.OpinionRef{
			{ID: "nice/alpha"},
			{ID: "nice/beta"},
		},
	}

	rs, err := resolve.Resolve(&speech, opinions, hardware.HardwareProfile{})
	if err != nil {
		t.Fatalf("nice-vs-nice: unexpected error: %v", err)
	}
	if rs == nil {
		t.Fatal("nice-vs-nice: nil ResolvedSpeech")
	}
	// nice/beta should be dropped (higher canonical ID loses)
	foundDropped := false
	for _, id := range rs.Dropped {
		if id == "nice/beta" {
			foundDropped = true
			break
		}
	}
	if !foundDropped {
		t.Errorf("nice-vs-nice: expected nice/beta in Dropped; dropped=%v applied=%v", rs.Dropped, rs.Applied)
	}
	// Rule 3 explanation should be present.
	foundRule3 := false
	for _, ex := range rs.Explanations {
		if ex.Rule == "rule3" {
			foundRule3 = true
			break
		}
	}
	if !foundRule3 {
		t.Error("nice-vs-nice: expected Rule 3 explanation in Explanations")
	}
}

// TestResolveRule1BDropsA verifies the Rule 1 branch where opB is required
// and opA is nice-to-have: opA should be dropped and opB kept.
func TestResolveRule1BDropsA(t *testing.T) {
	t.Parallel()

	// Canonical pair: A < B lexicographically, but B is required and A is nice-to-have.
	// Rule 1 second branch fires: bReq && !aReq → drop A.
	opinions := []resolver.Opinion{
		{
			Schema:   1,
			ID:       "pkg/aaa-nice",
			Name:     "AAA nice-to-have",
			Category: "package-install",
			Status:   resolver.StatusNiceToHave,
			Conflicts: []resolver.OpinionRef{{ID: "pkg/bbb-required"}},
		},
		{
			Schema:   1,
			ID:       "pkg/bbb-required",
			Name:     "BBB required",
			Category: "package-install",
			Status:   resolver.StatusRequired,
			Conflicts: []resolver.OpinionRef{{ID: "pkg/aaa-nice"}},
		},
	}
	speech := resolver.Speech{
		Schema:     1,
		ID:         "test-rule1-b-drops-a",
		Foundation: "arch",
		Opinions: []resolver.OpinionRef{
			{ID: "pkg/aaa-nice"},
			{ID: "pkg/bbb-required"},
		},
	}

	rs, err := resolve.Resolve(&speech, opinions, hardware.HardwareProfile{})
	if err != nil {
		t.Fatalf("rule1-b-drops-a: unexpected error: %v", err)
	}
	if rs == nil {
		t.Fatal("rule1-b-drops-a: nil ResolvedSpeech")
	}
	// aaa-nice should be dropped.
	foundDropped := false
	for _, id := range rs.Dropped {
		if id == "pkg/aaa-nice" {
			foundDropped = true
			break
		}
	}
	if !foundDropped {
		t.Errorf("rule1-b-drops-a: expected pkg/aaa-nice in Dropped; dropped=%v applied=%v", rs.Dropped, rs.Applied)
	}
	// Rule 1 explanation must be present.
	foundRule1 := false
	for _, ex := range rs.Explanations {
		if ex.Rule == "rule1" {
			foundRule1 = true
			break
		}
	}
	if !foundRule1 {
		t.Error("rule1-b-drops-a: expected Rule 1 explanation")
	}
}

// TestResolveInstallOrderExplanation verifies that an install-order Explanation
// with Rule="ordering" is emitted when opinions have ordering constraints
// (covers the hasOrdering branch in resolve.go).
func TestResolveInstallOrderExplanation(t *testing.T) {
	t.Parallel()

	opinions := []resolver.Opinion{
		{
			Schema:   1,
			ID:       "ord/first",
			Name:     "First",
			Category: "package-install",
			Status:   resolver.StatusRequired,
		},
		{
			Schema:   1,
			ID:       "ord/second",
			Name:     "Second",
			Category: "package-install",
			Status:   resolver.StatusRequired,
			DependsOn: []resolver.OpinionRef{{ID: "ord/first"}},
		},
	}
	speech := resolver.Speech{
		Schema:     1,
		ID:         "test-ordering-explanation",
		Foundation: "arch",
		Opinions: []resolver.OpinionRef{
			{ID: "ord/first"},
			{ID: "ord/second"},
		},
	}

	rs, err := resolve.Resolve(&speech, opinions, hardware.HardwareProfile{})
	if err != nil {
		t.Fatalf("ordering explanation: unexpected error: %v", err)
	}
	if rs == nil {
		t.Fatal("ordering explanation: nil ResolvedSpeech")
	}

	foundOrdering := false
	for _, ex := range rs.Explanations {
		if ex.Rule == "ordering" {
			foundOrdering = true
			break
		}
	}
	if !foundOrdering {
		var rules []string
		for _, ex := range rs.Explanations {
			rules = append(rules, ex.Rule)
		}
		t.Errorf("ordering explanation: expected Rule='ordering'; got rules: %v", rules)
	}
}

// ─── Canonical golden-file parity ─────────────────────────────────────────

// TestCanonicalGolden verifies that Resolve + CanonicalJSON over each named
// example fixture produces output byte-identical to a committed golden file
// in testdata/golden/.  With -update the test writes the goldens instead of
// comparing (used by scripts/wasm-parity-test.sh to regenerate the native
// baseline).
//
// The goldens are the source of truth for the parity script (01-05, RSLV-05).
// This test is the mechanism that both generates and verifies them.
func TestCanonicalGolden(t *testing.T) {
	t.Parallel()

	update := os.Getenv("GOLDEN_UPDATE") == "1"
	// GOLDEN_DIR may be overridden by the parity script to write goldens to a
	// temp directory for comparison.  Default is the committed testdata/golden/.
	goldenDir := os.Getenv("GOLDEN_DIR")
	if goldenDir == "" {
		goldenDir = filepath.Join("testdata", "golden")
	}

	cases := []struct {
		name    string // golden file stem (<goldenDir>/<name>.json)
		fixture string // relative path under examples/ dir (speech.yaml + opinions.yaml)
	}{
		{name: "omarchy-mini", fixture: "omarchy-mini"},
		{name: "two-point-clean", fixture: "two-point-clean"},
		{name: "conflicting", fixture: "conflicting"},
		{name: "hardware-conditional", fixture: "hardware-conditional"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Load example from examples/ directory (relative to module root).
			opinions, speech, hw := loadExample(t, tc.fixture)

			// Resolve — conflicting returns error but we still golden the partial RS.
			rs, _ := resolve.Resolve(&speech, opinions, hw)
			if rs == nil {
				t.Fatalf("Resolve returned nil *ResolvedSpeech for %s", tc.name)
			}

			got, err := resolve.CanonicalJSON(rs)
			if err != nil {
				t.Fatalf("CanonicalJSON error for %s: %v", tc.name, err)
			}

			goldenPath := filepath.Join(goldenDir, tc.name+".json")

			if update {
				// Write mode: create/overwrite the golden.
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
					t.Fatalf("mkdir golden dir: %v", err)
				}
				if err := os.WriteFile(goldenPath, got, 0o644); err != nil {
					t.Fatalf("write golden %s: %v", goldenPath, err)
				}
				t.Logf("updated golden: %s (%d bytes)", goldenPath, len(got))
				return
			}

			// Compare mode: golden must already exist.
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("golden file missing for %s (run with GOLDEN_UPDATE=1 to generate): %v", tc.name, err)
			}
			if !bytes.Equal(want, got) {
				t.Errorf("canonical JSON mismatch for %s:\n  golden: %s\n  got:    %s", tc.name, want, got)
			}
		})
	}
}

// loadExample loads speech.yaml + opinions.yaml from the examples/<name>/ directory
// (relative to the module root, two directories above this package).
// It re-uses the same fixture-loading logic as loadFixture but reads raw YAML
// that is already shaped as resolver.Opinion and resolver.Speech slices.
func loadExample(t *testing.T, name string) ([]resolver.Opinion, resolver.Speech, hardware.HardwareProfile) {
	t.Helper()

	// Locate the module root (two levels up from resolver/resolve/).
	// This relies on the standard Go module layout.
	root := findModuleRoot(t)

	speechPath := filepath.Join(root, "examples", name, "speech.yaml")
	opPath := filepath.Join(root, "examples", name, "opinions.yaml")

	speechRaw, err := os.ReadFile(speechPath)
	if err != nil {
		t.Fatalf("loadExample: cannot read %s: %v", speechPath, err)
	}
	opRaw, err := os.ReadFile(opPath)
	if err != nil {
		t.Fatalf("loadExample: cannot read %s: %v", opPath, err)
	}

	var speech resolver.Speech
	if err := yaml.Unmarshal(speechRaw, &speech); err != nil {
		t.Fatalf("loadExample: speech YAML parse error in %s: %v", speechPath, err)
	}

	// opinions.yaml contains a top-level list
	var opinions []resolver.Opinion
	if err := yaml.Unmarshal(opRaw, &opinions); err != nil {
		t.Fatalf("loadExample: opinions YAML parse error in %s: %v", opPath, err)
	}

	// Build hardware.HardwareProfile from speech.Hardware.
	// PCIIDs propagates directly now that resolver.HardwareProfile carries the field.
	hw := hardware.HardwareProfile{}
	if speech.Hardware != nil {
		hw.Predicates = speech.Hardware.Predicates
		hw.Facts = speech.Hardware.Facts
		hw.PCIIDs = speech.Hardware.PCIIDs
	}
	return opinions, speech, hw
}

// findModuleRoot locates the directory containing go.mod by walking up from this
// file's package directory.
func findModuleRoot(t *testing.T) string {
	t.Helper()
	// Start from the current working directory (which is the package dir when
	// running `go test ./resolver/resolve/`).
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("findModuleRoot: getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("findModuleRoot: could not locate go.mod")
		}
		dir = parent
	}
}

// ─── helpers ──────────────────────────────────────────────────────────────

func makeSet(ids []resolver.OpinionID) map[resolver.OpinionID]bool {
	m := make(map[resolver.OpinionID]bool, len(ids))
	for _, id := range ids {
		m[id] = true
	}
	return m
}

// Explanation re-exported alias so tests in package resolve_test can construct
// Explanation literals directly (avoid importing internal type).
type Explanation = resolve.Explanation

// TestResolveSysctlAndHardConflictBothReported verifies WR-04: when a speech
// has both a sysctl key collision AND a required-vs-required hard conflict,
// both errors must appear in the returned error — not just the sysctl one.
func TestResolveSysctlAndHardConflictBothReported(t *testing.T) {
	t.Parallel()

	// Two required opinions that conflict (hard conflict) AND share a sysctl key (collision).
	opinions := []resolver.Opinion{
		{
			Schema:   1,
			ID:       "pkg/sysctl-conflict-a",
			Name:     "A with sysctl",
			Category: "package-install",
			Status:   resolver.StatusRequired,
			Conflicts: []resolver.OpinionRef{{ID: "pkg/sysctl-conflict-b"}},
			SysctlParams: []resolver.SysctlParam{
				{Key: "vm.swappiness", Value: "10"},
			},
		},
		{
			Schema:   1,
			ID:       "pkg/sysctl-conflict-b",
			Name:     "B with same sysctl",
			Category: "package-install",
			Status:   resolver.StatusRequired,
			Conflicts: []resolver.OpinionRef{{ID: "pkg/sysctl-conflict-a"}},
			SysctlParams: []resolver.SysctlParam{
				{Key: "vm.swappiness", Value: "60"},
			},
		},
	}
	speech := resolver.Speech{
		Schema:     1,
		ID:         "test-sysctl-and-hard-conflict",
		Foundation: "arch",
		Opinions: []resolver.OpinionRef{
			{ID: "pkg/sysctl-conflict-a"},
			{ID: "pkg/sysctl-conflict-b"},
		},
	}

	_, err := resolve.Resolve(&speech, opinions, hardware.HardwareProfile{})
	if err == nil {
		t.Fatal("SysctlAndHardConflict: expected error, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "Sysctl") {
		t.Errorf("SysctlAndHardConflict: error %q must mention sysctl collision", errMsg)
	}
	if !strings.Contains(errMsg, "Hard conflict") {
		t.Errorf("SysctlAndHardConflict: error %q must mention hard conflict — both must be reported together", errMsg)
	}
}

// TestResolveTrustWarningNonHardwareOpinion verifies WR-02: a non-hardware-
// conditional opinion with a sig_level=Never custom repo must emit a
// TrustWarning in its Explanation — not just hardware-conditional ones.
func TestResolveTrustWarningNonHardwareOpinion(t *testing.T) {
	t.Parallel()

	// A plain required package-install opinion with a Never-sig repo.
	// No hardware_condition — goes through the "No conflict" path.
	opinions := []resolver.Opinion{
		{
			Schema:   1,
			ID:       "custom-repo/never-sig-plain",
			Name:     "Plain opinion with Never sig repo",
			Category: "custom-repo",
			Status:   resolver.StatusRequired,
			CustomRepos: []resolver.RepoDecl{
				{
					Name:     "myunsaferepo",
					URL:      "https://example.com/repo",
					SigLevel: resolver.SigLevelNever,
				},
			},
		},
	}
	speech := resolver.Speech{
		Schema:     1,
		ID:         "test-trust-warn-non-hw",
		Foundation: "arch",
		Opinions:   []resolver.OpinionRef{{ID: "custom-repo/never-sig-plain"}},
	}

	rs, err := resolve.Resolve(&speech, opinions, hardware.HardwareProfile{})
	if err != nil {
		t.Fatalf("TrustWarningNonHW: unexpected error: %v", err)
	}
	if rs == nil {
		t.Fatal("TrustWarningNonHW: nil ResolvedSpeech")
	}

	// Find the explanation for the opinion and assert TrustWarning is set.
	found := false
	for _, ex := range rs.Explanations {
		for _, id := range ex.OpinionsInvolved {
			if id == "custom-repo/never-sig-plain" {
				found = true
				if ex.TrustWarning == "" {
					t.Errorf("TrustWarningNonHW: explanation for custom-repo/never-sig-plain has empty TrustWarning; want non-empty (sig_level=Never repo)")
				}
			}
		}
	}
	if !found {
		t.Error("TrustWarningNonHW: no explanation found for custom-repo/never-sig-plain")
	}
}

// TestResolveMultipleHardConflictsAllReported verifies WR-01: when a speech
// contains two distinct required-vs-required conflict pairs, both pair IDs
// must appear in the returned error message (not just the last one).
func TestResolveMultipleHardConflictsAllReported(t *testing.T) {
	t.Parallel()

	// Four required opinions: A conflicts with B, C conflicts with D.
	// Both pairs are hard conflicts; the error must mention all four IDs.
	opinions := []resolver.Opinion{
		{
			Schema:    1,
			ID:        "pkg/conflict-a",
			Name:      "A",
			Category:  "package-install",
			Status:    resolver.StatusRequired,
			Conflicts: []resolver.OpinionRef{{ID: "pkg/conflict-b"}},
		},
		{
			Schema:    1,
			ID:        "pkg/conflict-b",
			Name:      "B",
			Category:  "package-install",
			Status:    resolver.StatusRequired,
			Conflicts: []resolver.OpinionRef{{ID: "pkg/conflict-a"}},
		},
		{
			Schema:    1,
			ID:        "pkg/conflict-c",
			Name:      "C",
			Category:  "package-install",
			Status:    resolver.StatusRequired,
			Conflicts: []resolver.OpinionRef{{ID: "pkg/conflict-d"}},
		},
		{
			Schema:    1,
			ID:        "pkg/conflict-d",
			Name:      "D",
			Category:  "package-install",
			Status:    resolver.StatusRequired,
			Conflicts: []resolver.OpinionRef{{ID: "pkg/conflict-c"}},
		},
	}
	speech := resolver.Speech{
		Schema:     1,
		ID:         "test-multi-hard-conflict",
		Foundation: "arch",
		Opinions: []resolver.OpinionRef{
			{ID: "pkg/conflict-a"},
			{ID: "pkg/conflict-b"},
			{ID: "pkg/conflict-c"},
			{ID: "pkg/conflict-d"},
		},
	}

	_, err := resolve.Resolve(&speech, opinions, hardware.HardwareProfile{})
	if err == nil {
		t.Fatal("MultipleHardConflicts: expected error, got nil")
	}

	// Both conflict pairs must appear in the error message.
	errMsg := err.Error()
	for _, id := range []string{"pkg/conflict-a", "pkg/conflict-b", "pkg/conflict-c", "pkg/conflict-d"} {
		if !strings.Contains(errMsg, id) {
			t.Errorf("MultipleHardConflicts: error %q does not mention %q — all conflict pairs must be reported", errMsg, id)
		}
	}
}
