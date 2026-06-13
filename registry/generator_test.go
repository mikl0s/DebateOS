package registry_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mikl0s/debateos/registry"
	"github.com/mikl0s/debateos/registry/index"
)

// fixturesDir returns the path to testdata/fixtures relative to this test file.
func fixturesDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(wd, "testdata", "fixtures")
}

// goldenDir returns the path to testdata/golden relative to this test file.
func goldenDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(wd, "testdata", "golden")
}

// capabilitiesForTest returns a minimal caps map covering the sample fixtures
// without depending on real capabilities.json path in tests.
func capabilitiesForTest() map[string][]string {
	return map[string][]string{
		"arch": {
			"install-packages",
			"deploy-config-file-tree",
			"enable-systemd-service",
			"write-sysctl-drop-in",
			"deploy-sddm-theme", // arch-only
		},
		"debian": {
			"install-packages",
			"deploy-config-file-tree",
			"enable-systemd-service",
			"write-sysctl-drop-in",
			// deploy-sddm-theme NOT in debian
		},
	}
}

// TestGenerateIndex verifies that GenerateIndex produces a RegistryIndex with
// the fixture point present, foundation_compat populated, and members sorted.
func TestGenerateIndex(t *testing.T) {
	caps := capabilitiesForTest()
	const fixedAt = "2026-01-01T00:00:00Z"

	idx, err := registry.GenerateIndex(fixturesDir(t), caps, fixedAt)
	if err != nil {
		t.Fatalf("GenerateIndex returned unexpected error: %v", err)
	}

	if idx.Schema != index.SchemaVersion {
		t.Errorf("schema: expected %d got %d", index.SchemaVersion, idx.Schema)
	}
	if idx.GeneratedAt != fixedAt {
		t.Errorf("generated_at: expected %q got %q", fixedAt, idx.GeneratedAt)
	}
	if len(idx.Points) == 0 {
		t.Fatal("expected at least one point in index, got 0")
	}

	// Find the sample point
	var sampleEntry *index.PointEntry
	for i := range idx.Points {
		if idx.Points[i].ID == "registry-test/sample-point" {
			sampleEntry = &idx.Points[i]
			break
		}
	}
	if sampleEntry == nil {
		t.Fatal("sample-point not found in index")
	}

	// Members must be sorted by ID
	for i := 1; i < len(sampleEntry.Members); i++ {
		if sampleEntry.Members[i] < sampleEntry.Members[i-1] {
			t.Errorf("members not sorted: %s before %s", sampleEntry.Members[i-1], sampleEntry.Members[i])
		}
	}

	// FoundationCompat must be populated
	if len(sampleEntry.FoundationCompat) == 0 {
		t.Fatal("expected foundation_compat to be populated")
	}

	// The point has member SMP-002 which requires deploy-sddm-theme (arch-only)
	// so debian must be incompatible
	var debianFC *index.FoundationCompat
	for i := range sampleEntry.FoundationCompat {
		if sampleEntry.FoundationCompat[i].Foundation == "debian" {
			debianFC = &sampleEntry.FoundationCompat[i]
			break
		}
	}
	if debianFC == nil {
		t.Fatal("debian foundation_compat entry missing")
	}
	if debianFC.Compatible {
		t.Error("debian should not be compatible (missing deploy-sddm-theme)")
	}
	found := false
	for _, m := range debianFC.Missing {
		if m == "deploy-sddm-theme" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("deploy-sddm-theme should be in debian Missing, got %v", debianFC.Missing)
	}
}

// TestGenerateIndexValidation verifies that a malformed opinion YAML in the
// fixture dir causes GenerateIndex to return a non-nil error naming the file.
func TestGenerateIndexValidation(t *testing.T) {
	// Create a temp fixture dir with a malformed opinion
	tmpDir := t.TempDir()
	pointsDir := filepath.Join(tmpDir, "points")
	opinionsDir := filepath.Join(tmpDir, "opinions")
	if err := os.MkdirAll(pointsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(opinionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write a valid point
	validPoint := `schema: 1
id: test/bad-point
name: Bad Point
curator: test@example.com
members:
- id: BAD-001
  status: required
`
	if err := os.WriteFile(filepath.Join(pointsDir, "bad-point.yaml"), []byte(validPoint), 0o644); err != nil {
		t.Fatal(err)
	}

	// Write a malformed opinion (invalid YAML / missing required field)
	malformedOpinion := `schema: 1
id: BAD-001
# Missing required 'status' field
name: Bad Opinion
category: package-install
unknown_field_xyz: should_fail_strict_parse
`
	if err := os.WriteFile(filepath.Join(opinionsDir, "BAD-001.yaml"), []byte(malformedOpinion), 0o644); err != nil {
		t.Fatal(err)
	}

	caps := capabilitiesForTest()
	_, err := registry.GenerateIndex(tmpDir, caps, "2026-01-01T00:00:00Z")
	if err == nil {
		t.Fatal("expected error for malformed opinion YAML, got nil")
	}

	// Error must mention the offending file
	if !strings.Contains(err.Error(), "BAD-001.yaml") {
		t.Errorf("error should name offending file BAD-001.yaml, got: %v", err)
	}
}

// TestDeterminism verifies that marshaling the same RegistryIndex twice yields
// byte-identical JSON and that a fixed generatedAt is stable (no time.Now()).
func TestDeterminism(t *testing.T) {
	caps := capabilitiesForTest()
	const fixedAt = "2026-01-01T00:00:00Z"

	idx1, err := registry.GenerateIndex(fixturesDir(t), caps, fixedAt)
	if err != nil {
		t.Fatalf("first GenerateIndex failed: %v", err)
	}
	idx2, err := registry.GenerateIndex(fixturesDir(t), caps, fixedAt)
	if err != nil {
		t.Fatalf("second GenerateIndex failed: %v", err)
	}

	b1, err := json.MarshalIndent(idx1, "", "  ")
	if err != nil {
		t.Fatalf("marshal idx1: %v", err)
	}
	b2, err := json.MarshalIndent(idx2, "", "  ")
	if err != nil {
		t.Fatalf("marshal idx2: %v", err)
	}

	if !bytes.Equal(b1, b2) {
		t.Errorf("two runs produced different JSON:\nrun1: %s\n\nrun2: %s", b1, b2)
	}
}

// TestGoldenIndex verifies that the emitted JSON exactly matches the committed
// golden file registry/testdata/golden/index.json byte-for-byte.
func TestGoldenIndex(t *testing.T) {
	caps := capabilitiesForTest()
	const fixedAt = "2026-01-01T00:00:00Z"

	idx, err := registry.GenerateIndex(fixturesDir(t), caps, fixedAt)
	if err != nil {
		t.Fatalf("GenerateIndex failed: %v", err)
	}

	emitted, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// Append trailing newline to match committed file
	emitted = append(emitted, '\n')

	goldenPath := filepath.Join(goldenDir(t), "index.json")
	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		// Golden doesn't exist yet: write it (first run) and pass
		if os.IsNotExist(err) {
			if err := os.MkdirAll(goldenDir(t), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(goldenPath, emitted, 0o644); err != nil {
				t.Fatal(err)
			}
			t.Logf("golden/index.json written for first time at %s", goldenPath)
			return
		}
		t.Fatalf("read golden: %v", err)
	}

	if !bytes.Equal(emitted, golden) {
		t.Errorf("emitted JSON differs from golden.\nEmitted:\n%s\n\nGolden:\n%s", emitted, golden)
	}
}

// TestEmitHTML verifies that EmitHTML produces a non-empty JS-free browse page
// that contains the point name and compat badges.
func TestEmitHTML(t *testing.T) {
	caps := capabilitiesForTest()
	const fixedAt = "2026-01-01T00:00:00Z"

	idx, err := registry.GenerateIndex(fixturesDir(t), caps, fixedAt)
	if err != nil {
		t.Fatalf("GenerateIndex failed: %v", err)
	}

	var buf bytes.Buffer
	if err := registry.EmitHTML(idx, &buf); err != nil {
		t.Fatalf("EmitHTML returned error: %v", err)
	}

	html := buf.String()
	if len(html) == 0 {
		t.Fatal("EmitHTML produced empty output")
	}

	// Must contain the point name
	if !strings.Contains(html, "Sample Point") {
		t.Error("HTML should contain point name 'Sample Point'")
	}
	// Must contain compat badges (text-based, no JS)
	if !strings.Contains(html, "arch") {
		t.Error("HTML should contain 'arch' foundation badge")
	}
	if !strings.Contains(html, "debian") {
		t.Error("HTML should contain 'debian' foundation badge")
	}
	// Must NOT contain <script> tags (static, no JS)
	if strings.Contains(html, "<script") {
		t.Error("HTML must be JS-free (no <script> tags)")
	}
}
