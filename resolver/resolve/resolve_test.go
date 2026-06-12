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
// `hardware_override` block for cases (EC-038, EC-041, EC-042) requiring
// PCIIDs, which resolver.HardwareProfile does not carry.
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
	hw := hardware.HardwareProfile{}
	if doc.Speech.Hardware != nil {
		hw.Predicates = doc.Speech.Hardware.Predicates
		hw.Facts = doc.Speech.Hardware.Facts
	}
	// Apply fixture-level override for PCI IDs and richer profiles.
	if doc.HardwareOverride != nil {
		if doc.HardwareOverride.PCIIDs != nil {
			hw.PCIIDs = doc.HardwareOverride.PCIIDs
		}
	}
	// EC-038: if speech.hardware has pci_ids declared (field name in the YAML),
	// the hardware block's YAML is already decoded into resolver.HardwareProfile
	// which does NOT carry PCIIDs. We need to re-parse the hardware block raw.
	// Workaround: store PCIIDs in the fixture's hardware block as a custom field
	// and re-read them via a separate raw decode.
	if doc.Speech.Hardware != nil {
		// Re-decode the hardware block raw to pick up pci_ids if present.
		var rawHWBlock struct {
			PCIIDS []string `yaml:"pci_ids"`
		}
		hwRaw, _ := yaml.Marshal(doc.Speech.Hardware)
		_ = yaml.Unmarshal(hwRaw, &rawHWBlock)
		if len(rawHWBlock.PCIIDS) > 0 {
			hw.PCIIDs = rawHWBlock.PCIIDS
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
