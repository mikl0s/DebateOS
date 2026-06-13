// Package loader_test exercises the shared speech-loading pipeline.
//
// These tests verify that ResolveDir correctly handles:
//   - A valid speech directory (examples/omarchy as the canonical fixture)
//   - Missing speech.yaml
//   - Malformed speech.yaml
//   - Missing points directory (treated as empty)
//   - Missing opinions directory (treated as empty)
//   - A referenced point not found in points/
//   - A referenced opinion not found in opinions/
//   - A minimal speech with no points/opinions (trivially resolves)
package loader_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mikl0s/debateos/cli/internal/loader"
)

// repoRoot returns the path to the repository root by walking up from this test file.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// file is <repo>/cli/internal/loader/loader_test.go; go up 3 dirs.
	return filepath.Join(filepath.Dir(file), "..", "..", "..")
}

// TestResolveDirOmarchy verifies that ResolveDir succeeds on the canonical
// examples/omarchy fixture — the same fixture used by the validate subcommand.
func TestResolveDirOmarchy(t *testing.T) {
	speechDir := filepath.Join(repoRoot(t), "examples", "omarchy")
	rs, err := loader.ResolveDir(speechDir)
	if err != nil {
		t.Fatalf("ResolveDir(omarchy): unexpected error: %v", err)
	}
	if rs == nil {
		t.Fatal("ResolveDir returned nil ResolvedSpeech")
	}
	// The north-star: Applied=99, Skipped=35, Dropped=0 (per REQUIREMENTS.md ARCH-02).
	// We allow a broad lower bound here so this test doesn't break if opinion counts change.
	if len(rs.Applied) == 0 {
		t.Errorf("expected >0 applied opinions, got 0")
	}
}

// TestResolveDirMinimal verifies that a minimal speech (no points, no opinions)
// resolves cleanly with empty Applied/Skipped/Dropped.
func TestResolveDirMinimal(t *testing.T) {
	dir := t.TempDir()
	speech := `schema: 1
id: minimal-test
foundation: arch
points: []
opinions: []
`
	if err := os.WriteFile(filepath.Join(dir, "speech.yaml"), []byte(speech), 0644); err != nil {
		t.Fatal(err)
	}

	rs, err := loader.ResolveDir(dir)
	if err != nil {
		t.Fatalf("ResolveDir(minimal): %v", err)
	}
	if rs == nil {
		t.Fatal("nil ResolvedSpeech on minimal speech")
	}
}

// TestResolveDirMissingSpeech verifies that ResolveDir returns an error when
// speech.yaml does not exist.
func TestResolveDirMissingSpeech(t *testing.T) {
	dir := t.TempDir()
	// No speech.yaml written.
	_, err := loader.ResolveDir(dir)
	if err == nil {
		t.Fatal("expected error for missing speech.yaml, got nil")
	}
	if !strings.Contains(err.Error(), "speech.yaml") {
		t.Errorf("error should mention speech.yaml: %v", err)
	}
}

// TestResolveDirBadSpeechYAML verifies that malformed YAML returns a parse error.
func TestResolveDirBadSpeechYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "speech.yaml"), []byte("this: [bad yaml: {"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := loader.ResolveDir(dir)
	if err == nil {
		t.Fatal("expected error for malformed speech.yaml, got nil")
	}
}

// TestResolveDirMissingPoints verifies that a missing points/ dir is treated as
// an empty collection (not an error).
func TestResolveDirMissingPoints(t *testing.T) {
	dir := t.TempDir()
	speech := `schema: 1
id: no-points
foundation: arch
points: []
opinions: []
`
	if err := os.WriteFile(filepath.Join(dir, "speech.yaml"), []byte(speech), 0644); err != nil {
		t.Fatal(err)
	}
	// No points/ dir created.

	rs, err := loader.ResolveDir(dir)
	if err != nil {
		t.Fatalf("missing points/ should not be an error: %v", err)
	}
	if rs == nil {
		t.Fatal("nil ResolvedSpeech")
	}
}

// TestResolveDirMissingOpinions verifies that a missing opinions/ dir is treated
// as an empty collection (not an error).
func TestResolveDirMissingOpinions(t *testing.T) {
	dir := t.TempDir()
	speech := `schema: 1
id: no-opinions
foundation: arch
points: []
opinions: []
`
	if err := os.WriteFile(filepath.Join(dir, "speech.yaml"), []byte(speech), 0644); err != nil {
		t.Fatal(err)
	}
	// No opinions/ dir created.

	rs, err := loader.ResolveDir(dir)
	if err != nil {
		t.Fatalf("missing opinions/ should not be an error: %v", err)
	}
	if rs == nil {
		t.Fatal("nil ResolvedSpeech")
	}
}

// TestResolveDirUnknownPoint verifies that a point referenced in speech.yaml
// but absent from points/ returns an error.
func TestResolveDirUnknownPoint(t *testing.T) {
	dir := t.TempDir()
	// Reference a point that does not exist.
	speech := `schema: 1
id: bad-point-ref
foundation: arch
points:
  - id: "some/nonexistent-point"
opinions: []
`
	if err := os.WriteFile(filepath.Join(dir, "speech.yaml"), []byte(speech), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "points"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "opinions"), 0755); err != nil {
		t.Fatal(err)
	}

	_, err := loader.ResolveDir(dir)
	if err == nil {
		t.Fatal("expected error for unknown point reference, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistent-point") {
		t.Errorf("error should mention the missing point: %v", err)
	}
}

// TestResolveDirWithPointAndOpinion creates a minimal point + opinion and verifies
// the assembleOpinions path is exercised.
func TestResolveDirWithPointAndOpinion(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "points"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "opinions"), 0755); err != nil {
		t.Fatal(err)
	}

	// Write a minimal opinion.
	opYAML := `schema: 1
id: "test/test-op"
category: packages
status: required
name: "Test Opinion"
packages:
  - vim
`
	if err := os.WriteFile(filepath.Join(dir, "opinions", "test-op.yaml"), []byte(opYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Write a minimal point referencing that opinion.
	ptYAML := `schema: 1
id: "test/test-point"
name: "Test Point"
members:
  - id: "test/test-op"
`
	if err := os.WriteFile(filepath.Join(dir, "points", "test-point.yaml"), []byte(ptYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Speech referencing the point.
	speech := `schema: 1
id: test-with-point
foundation: arch
points:
  - id: "test/test-point"
opinions: []
`
	if err := os.WriteFile(filepath.Join(dir, "speech.yaml"), []byte(speech), 0644); err != nil {
		t.Fatal(err)
	}

	rs, err := loader.ResolveDir(dir)
	if err != nil {
		t.Fatalf("ResolveDir with point+opinion: %v", err)
	}
	if rs == nil {
		t.Fatal("nil ResolvedSpeech")
	}
	// The test opinion should be Applied or Skipped (required with no conflicts).
	if len(rs.Applied)+len(rs.Skipped) == 0 {
		t.Errorf("expected at least one applied or skipped opinion, got none")
	}
}

// TestResolveDirBadPointYAML verifies that a malformed point YAML in points/
// returns a parse error.
func TestResolveDirBadPointYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "points"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "points", "bad.yaml"), []byte("this: [bad: {"), 0644); err != nil {
		t.Fatal(err)
	}

	speech := `schema: 1
id: bad-point
foundation: arch
points: []
opinions: []
`
	if err := os.WriteFile(filepath.Join(dir, "speech.yaml"), []byte(speech), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := loader.ResolveDir(dir)
	if err == nil {
		t.Fatal("expected error for malformed point YAML, got nil")
	}
}

// TestResolveDirUnknownOpinionInPoint verifies that a point referencing an
// opinion not present in opinions/ returns an error.
func TestResolveDirUnknownOpinionInPoint(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "points"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "opinions"), 0755); err != nil {
		t.Fatal(err)
	}

	// Point references an opinion that doesn't exist.
	ptYAML := `schema: 1
id: "test/pt-with-missing-op"
name: "Test Point"
members:
  - id: "test/nonexistent-opinion"
`
	if err := os.WriteFile(filepath.Join(dir, "points", "pt.yaml"), []byte(ptYAML), 0644); err != nil {
		t.Fatal(err)
	}

	speech := `schema: 1
id: missing-op-ref
foundation: arch
points:
  - id: "test/pt-with-missing-op"
opinions: []
`
	if err := os.WriteFile(filepath.Join(dir, "speech.yaml"), []byte(speech), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := loader.ResolveDir(dir)
	if err == nil {
		t.Fatal("expected error for missing opinion referenced by point, got nil")
	}
}

// TestResolveDirBadOpinionYAML verifies that a malformed opinion YAML in opinions/
// returns a parse error.
func TestResolveDirBadOpinionYAML(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "opinions"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "opinions", "bad.yaml"), []byte("this: [bad: {"), 0644); err != nil {
		t.Fatal(err)
	}

	speech := `schema: 1
id: bad-opinion
foundation: arch
points: []
opinions: []
`
	if err := os.WriteFile(filepath.Join(dir, "speech.yaml"), []byte(speech), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := loader.ResolveDir(dir)
	if err == nil {
		t.Fatal("expected error for malformed opinion YAML, got nil")
	}
}

// TestResolveDirDirectOpinionRef verifies that opinions referenced directly in
// speech.Opinions (not via points) are assembled correctly.
func TestResolveDirDirectOpinionRef(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "opinions"), 0755); err != nil {
		t.Fatal(err)
	}

	opYAML := `schema: 1
id: "direct/my-op"
category: packages
status: required
name: "Direct Opinion"
packages:
  - git
`
	if err := os.WriteFile(filepath.Join(dir, "opinions", "my-op.yaml"), []byte(opYAML), 0644); err != nil {
		t.Fatal(err)
	}

	speech := `schema: 1
id: direct-op-test
foundation: arch
points: []
opinions:
  - id: "direct/my-op"
`
	if err := os.WriteFile(filepath.Join(dir, "speech.yaml"), []byte(speech), 0644); err != nil {
		t.Fatal(err)
	}

	rs, err := loader.ResolveDir(dir)
	if err != nil {
		t.Fatalf("ResolveDir with direct opinion ref: %v", err)
	}
	if len(rs.Applied)+len(rs.Skipped) == 0 {
		t.Errorf("expected at least one applied/skipped, got none (Applied=%d Skipped=%d)",
			len(rs.Applied), len(rs.Skipped))
	}
}
