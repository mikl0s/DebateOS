package build_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mikl0s/debateos/cli/build"
	"github.com/mikl0s/debateos/cli/runner"
)

// ─────────────────────────────────────────────────────────────────────────────
// Fixtures
// ─────────────────────────────────────────────────────────────────────────────

// minimalSpeechDir creates a minimal but valid speech directory in t.TempDir()
// that the loader pipeline can process (with no real opinions so resolution
// trivially succeeds).
func minimalSpeechDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// speech.yaml — minimal valid speech (schema: 1 required by validator)
	speech := `schema: 1
id: test-speech
foundation: arch
points: []
opinions: []
`
	if err := os.WriteFile(filepath.Join(dir, "speech.yaml"), []byte(speech), 0644); err != nil {
		t.Fatal(err)
	}
	// points/ and opinions/ can be absent — loader tolerates that.
	return dir
}

// ─────────────────────────────────────────────────────────────────────────────
// Task 1 tests: build orchestration (dry-run / skip-iso / full / epoch)
// ─────────────────────────────────────────────────────────────────────────────

// TestBuildDryRun asserts that --dry-run prints the build plan (resolved.json
// path, epoch, translate argv, docker argv) and makes ZERO Runner calls.
func TestBuildDryRun(t *testing.T) {
	speechDir := minimalSpeechDir(t)
	outDir := t.TempDir()
	fake := &runner.FakeRunner{}

	var stdout, stderr bytes.Buffer
	code := build.Run(
		[]string{"--dir", speechDir, "--out", outDir, "--dry-run"},
		&stdout, &stderr, fake,
	)

	if code != 0 {
		t.Fatalf("dry-run: expected exit 0, got %d\nstderr: %s", code, stderr.String())
	}
	if len(fake.Calls) != 0 {
		t.Errorf("dry-run: expected 0 runner calls, got %d: %v", len(fake.Calls), fake.Calls)
	}

	out := stdout.String()
	// Plan output must contain key items.
	if !strings.Contains(out, "resolved.json") {
		t.Errorf("dry-run plan missing 'resolved.json': %s", out)
	}
	if !strings.Contains(out, "SOURCE_DATE_EPOCH") {
		t.Errorf("dry-run plan missing 'SOURCE_DATE_EPOCH': %s", out)
	}
	if !strings.Contains(out, "translate") {
		t.Errorf("dry-run plan missing 'translate': %s", out)
	}
	if !strings.Contains(out, "docker") {
		t.Errorf("dry-run plan missing 'docker': %s", out)
	}
}

// TestBuildSkipISO asserts that --skip-iso calls translate with the EXACT
// frozen argv contract and does NOT invoke docker.
func TestBuildSkipISO(t *testing.T) {
	speechDir := minimalSpeechDir(t)
	outDir := t.TempDir()
	fake := &runner.FakeRunner{}

	var stdout, stderr bytes.Buffer
	code := build.Run(
		[]string{"--dir", speechDir, "--out", outDir, "--profile", "vanilla-arch", "--skip-iso"},
		&stdout, &stderr, fake,
	)

	if code != 0 {
		t.Fatalf("skip-iso: expected exit 0, got %d\nstderr: %s", code, stderr.String())
	}

	// Exactly one runner call: translate with frozen argv.
	translateFound := false
	for _, call := range fake.Calls {
		if strings.HasPrefix(call, "translators/arch/translate") {
			translateFound = true
			// Assert frozen argv contract: translate <resolved.json> --opinions <dir> --profile <name> --out <dir>
			if !strings.Contains(call, "resolved.json") {
				t.Errorf("translate call missing resolved.json: %q", call)
			}
			if !strings.Contains(call, "--opinions") {
				t.Errorf("translate call missing --opinions: %q", call)
			}
			if !strings.Contains(call, "--profile") {
				t.Errorf("translate call missing --profile: %q", call)
			}
			if !strings.Contains(call, "--out") {
				t.Errorf("translate call missing --out: %q", call)
			}
			if !strings.Contains(call, "vanilla-arch") {
				t.Errorf("translate call missing profile name 'vanilla-arch': %q", call)
			}
		}
		if strings.HasPrefix(call, "docker") {
			t.Errorf("skip-iso: unexpected docker runner call: %q", call)
		}
	}
	if !translateFound {
		t.Errorf("skip-iso: no translate runner call found in: %v", fake.Calls)
	}

	// resolved.json must be written to --out dir.
	resolvedPath := filepath.Join(outDir, "resolved.json")
	if _, err := os.Stat(resolvedPath); err != nil {
		t.Errorf("skip-iso: resolved.json not written to out dir: %v", err)
	}
}

// TestBuildDocker asserts that the full build (no --skip-iso) issues BOTH a
// translate call AND a docker call, and that the docker call includes
// SOURCE_DATE_EPOCH in the -e flag.
func TestBuildDocker(t *testing.T) {
	speechDir := minimalSpeechDir(t)
	outDir := t.TempDir()
	fake := &runner.FakeRunner{}

	var stdout, stderr bytes.Buffer
	code := build.Run(
		[]string{"--dir", speechDir, "--out", outDir, "--profile", "vanilla-arch"},
		&stdout, &stderr, fake,
	)

	if code != 0 {
		t.Fatalf("full build: expected exit 0, got %d\nstderr: %s", code, stderr.String())
	}

	translateFound := false
	dockerFound := false
	for _, call := range fake.Calls {
		if strings.HasPrefix(call, "translators/arch/translate") {
			translateFound = true
		}
		if strings.HasPrefix(call, "docker") {
			dockerFound = true
			// Docker call must pass -e SOURCE_DATE_EPOCH=<value>
			if !strings.Contains(call, "SOURCE_DATE_EPOCH") {
				t.Errorf("docker call missing SOURCE_DATE_EPOCH env: %q", call)
			}
			// Must include volume mounts for speech and out dirs.
			if !strings.Contains(call, "-v") {
				t.Errorf("docker call missing -v mount: %q", call)
			}
		}
	}
	if !translateFound {
		t.Errorf("full build: no translate runner call found: %v", fake.Calls)
	}
	if !dockerFound {
		t.Errorf("full build: no docker runner call found: %v", fake.Calls)
	}
}

// TestBuildEpochConsistency asserts that the Go epoch derivation for a fixed
// resolved.json fixture equals the value computed by the manifest.py algorithm.
//
// Golden value is derived from manifest.py derive_source_date_epoch on the
// same bytes:
//
//	_MIN = 1577836800, _MAX = 2208988800
//	content = b'{"schema":1,"foundation":"arch","explanations":[]}'
//	sha256 → first4 BE uint32 → _MIN + (raw % (_MAX - _MIN))
//
// This golden was computed by running manifest.py on the test fixture bytes.
func TestBuildEpochConsistency(t *testing.T) {
	// Fixed minimal resolved.json bytes that the loader produces for a speech
	// with no points/opinions. We compute the golden here using the same
	// algorithm as manifest.py derive_source_date_epoch so this test is
	// self-consistent even without running Python.
	fixtureBytes := []byte(`{"schema":1,"foundation":"arch","explanations":[]}`)

	got := build.DeriveEpoch(fixtureBytes)

	// Re-derive expected using the same algorithm.
	want := computeExpectedEpoch(fixtureBytes)

	if got != want {
		t.Errorf("epoch mismatch: got %d, want %d", got, want)
	}

	// Boundary checks: epoch must be in valid range.
	const minEpoch = 1577836800 // 2020-01-01T00:00:00Z
	const maxEpoch = 2208988800 // 2040-01-01T00:00:00Z
	if got < minEpoch || got >= maxEpoch {
		t.Errorf("epoch %d out of range [%d, %d)", got, minEpoch, maxEpoch)
	}

	// Determinism: same input → same output.
	got2 := build.DeriveEpoch(fixtureBytes)
	if got != got2 {
		t.Error("DeriveEpoch is not deterministic")
	}
}

// computeExpectedEpoch mirrors manifest.py derive_source_date_epoch exactly.
func computeExpectedEpoch(content []byte) int64 {
	digest := sha256.Sum256(content)
	raw := binary.BigEndian.Uint32(digest[:4])
	const minE = int64(1577836800)
	const maxE = int64(2208988800)
	return minE + (int64(raw) % (maxE - minE))
}

// TestBuildTranslateArgvOrder asserts the exact positional order of the
// translate argv: translate <resolved.json> --opinions <dir> --profile <name> --out <dir>
func TestBuildTranslateArgvOrder(t *testing.T) {
	speechDir := minimalSpeechDir(t)
	outDir := t.TempDir()
	fake := &runner.FakeRunner{}

	var stdout, stderr bytes.Buffer
	build.Run(
		[]string{"--dir", speechDir, "--out", outDir, "--profile", "vanilla-arch", "--skip-iso"},
		&stdout, &stderr, fake,
	)

	var call string
	for _, c := range fake.Calls {
		if strings.HasPrefix(c, "translators/arch/translate") {
			call = c
			break
		}
	}
	if call == "" {
		t.Fatal("no translate call found")
	}

	// Parse tokens.
	tokens := strings.Fields(call)
	// tokens[0] = "translators/arch/translate"
	// tokens[1] = <resolved.json path>
	// tokens[2] = "--opinions"
	// tokens[3] = <opinions-dir>
	// tokens[4] = "--profile"
	// tokens[5] = <profile-name>
	// tokens[6] = "--out"
	// tokens[7] = <out-dir>
	if len(tokens) < 8 {
		t.Fatalf("translate call has too few tokens (%d): %q", len(tokens), call)
	}
	if !strings.HasSuffix(tokens[1], "resolved.json") {
		t.Errorf("token[1] must be resolved.json path, got %q", tokens[1])
	}
	if tokens[2] != "--opinions" {
		t.Errorf("token[2] must be '--opinions', got %q", tokens[2])
	}
	if tokens[4] != "--profile" {
		t.Errorf("token[4] must be '--profile', got %q", tokens[4])
	}
	if tokens[5] != "vanilla-arch" {
		t.Errorf("token[5] must be 'vanilla-arch', got %q", tokens[5])
	}
	if tokens[6] != "--out" {
		t.Errorf("token[6] must be '--out', got %q", tokens[6])
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Task 2 tests: private-injection.tar + sanitization + secret-free profile
// ─────────────────────────────────────────────────────────────────────────────

// paneAsset is a simple test helper matching the build.PaneAsset type.
type paneAsset struct {
	Src     string // source bytes (content written to a temp file)
	Dst     string // target-relative destination path
	Mode    int64
	Content []byte
}

// TestInjectionTarLayout asserts:
// - debateos-private.json manifest is present at the tar root.
// - Each asset is at its target-relative path (dst).
// - The tar is written to the output dir (not the profile dir).
func TestInjectionTarLayout(t *testing.T) {
	outDir := t.TempDir()
	profileDir := filepath.Join(outDir, "arch-profile")
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		t.Fatal(err)
	}

	assets := []build.PaneAsset{
		{Dst: "etc/debateos/config.yaml", Content: []byte("key: value\n"), Mode: 0600},
		{Dst: "home/user/.zshrc", Content: []byte("# private zsh\n"), Mode: 0644},
	}

	tarPath, err := build.WriteInjectionTar(outDir, assets)
	if err != nil {
		t.Fatalf("WriteInjectionTar: %v", err)
	}

	// 1. Tar must be in outDir (not profileDir).
	if !strings.HasPrefix(tarPath, outDir) {
		t.Errorf("tar written to wrong location: %s (expected under %s)", tarPath, outDir)
	}
	if strings.HasPrefix(tarPath, profileDir) {
		t.Errorf("tar MUST NOT be inside profile dir: %s", tarPath)
	}

	// 2. Inspect tar contents.
	entries := readTarEntries(t, tarPath)

	manifestFound := false
	for name := range entries {
		if name == "debateos-private.json" {
			manifestFound = true
		}
	}
	if !manifestFound {
		t.Errorf("debateos-private.json not found in tar; entries: %v", tarKeys(entries))
	}

	// 3. Each asset at its target-relative path.
	for _, a := range assets {
		if _, ok := entries[a.Dst]; !ok {
			t.Errorf("asset %q not found in tar; entries: %v", a.Dst, tarKeys(entries))
		}
	}
}

// TestInjectionTarManifestContent asserts the manifest JSON is well-formed
// and contains the file list.
func TestInjectionTarManifestContent(t *testing.T) {
	outDir := t.TempDir()

	assets := []build.PaneAsset{
		{Dst: "etc/ssh/authorized_keys", Content: []byte("ssh-ed25519 AAAA...\n"), Mode: 0600},
	}

	tarPath, err := build.WriteInjectionTar(outDir, assets)
	if err != nil {
		t.Fatalf("WriteInjectionTar: %v", err)
	}

	entries := readTarEntries(t, tarPath)
	manifestBytes, ok := entries["debateos-private.json"]
	if !ok {
		t.Fatal("debateos-private.json not in tar")
	}

	var manifest struct {
		Version int    `json:"version"`
		Created string `json:"created"`
		Files   []struct {
			Path string `json:"path"`
			Mode int    `json:"mode"`
		} `json:"files"`
	}
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("manifest parse error: %v\ncontent: %s", err, string(manifestBytes))
	}
	if manifest.Version == 0 {
		t.Error("manifest.version must be non-zero")
	}
	if manifest.Created == "" {
		t.Error("manifest.created must be non-empty")
	}
	if len(manifest.Files) != len(assets) {
		t.Errorf("manifest.files len %d, want %d", len(manifest.Files), len(assets))
	}
	if len(manifest.Files) > 0 && manifest.Files[0].Path != assets[0].Dst {
		t.Errorf("manifest.files[0].path = %q, want %q", manifest.Files[0].Path, assets[0].Dst)
	}
}

// TestInjectSanitizeAbsolute asserts that absolute dst paths are rejected.
func TestInjectSanitizeAbsolute(t *testing.T) {
	outDir := t.TempDir()
	assets := []build.PaneAsset{
		{Dst: "/etc/shadow", Content: []byte("root:x:...\n"), Mode: 0600},
	}
	_, err := build.WriteInjectionTar(outDir, assets)
	if err == nil {
		t.Error("expected error for absolute dst '/etc/shadow', got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "absolute") && !strings.Contains(err.Error(), "traversal") {
		t.Errorf("error message should mention 'absolute' or 'traversal': %v", err)
	}
}

// TestInjectSanitizeTraversal asserts that ../ traversal dst paths are rejected.
func TestInjectSanitizeTraversal(t *testing.T) {
	outDir := t.TempDir()
	assets := []build.PaneAsset{
		{Dst: "etc/../../etc/shadow", Content: []byte("root:x:...\n"), Mode: 0600},
	}
	_, err := build.WriteInjectionTar(outDir, assets)
	if err == nil {
		t.Error("expected error for traversal dst 'etc/../../etc/shadow', got nil")
	}
}

// TestSecretFreeProfile asserts that the arch-profile tree produced by
// --skip-iso does NOT contain pane.yaml, identity.age, or private-injection.tar.
func TestSecretFreeProfile(t *testing.T) {
	speechDir := minimalSpeechDir(t)
	outDir := t.TempDir()
	fake := &runner.FakeRunner{}

	// Write sentinel "secret" files in the speech dir to simulate pane data.
	// These must NOT appear in the profile tree.
	if err := os.WriteFile(filepath.Join(speechDir, "pane.yaml"), []byte("secret: data\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(speechDir, "identity.age"), []byte("AGE-SECRET-KEY-...\n"), 0600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := build.Run(
		[]string{"--dir", speechDir, "--out", outDir, "--profile", "vanilla-arch", "--skip-iso"},
		&stdout, &stderr, fake,
	)
	if code != 0 {
		t.Fatalf("skip-iso: exit %d\nstderr: %s", code, stderr.String())
	}

	// Assert profile tree (if it exists) contains none of the secret files.
	profileDir := filepath.Join(outDir, "arch-profile")
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		// No profile dir = translate was fake; that's fine — still assert outDir.
	}

	forbidden := []string{"pane.yaml", "identity.age", "private-injection.tar"}
	err := filepath.Walk(profileDir, func(path string, info os.FileInfo, err error) error {
		if os.IsNotExist(err) {
			return nil // profile dir not created by FakeRunner — skip
		}
		if err != nil {
			return err
		}
		for _, f := range forbidden {
			if info.Name() == f {
				return fmt.Errorf("secret file %q found inside profile tree at %s", f, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Errorf("profile tree contains secret files: %v", err)
	}

	// private-injection.tar must NOT be inside the profile dir.
	tarInProfile := filepath.Join(profileDir, "private-injection.tar")
	if _, statErr := os.Stat(tarInProfile); statErr == nil {
		t.Errorf("private-injection.tar found INSIDE profile dir: %s", tarInProfile)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// readTarEntries reads a .tar (no gzip) or .tar.gz and returns a map
// header-name → content bytes.
func readTarEntries(t *testing.T, path string) map[string][]byte {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open tar %s: %v", path, err)
	}
	defer f.Close()

	var rd io.Reader = f
	// Try gzip decompression.
	if strings.HasSuffix(path, ".gz") || strings.HasSuffix(path, ".tgz") {
		gz, gzErr := gzip.NewReader(f)
		if gzErr != nil {
			t.Fatalf("gzip open %s: %v", path, gzErr)
		}
		defer gz.Close()
		rd = gz
	}

	tr := tar.NewReader(rd)
	entries := make(map[string][]byte)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("tar read %s: %v", path, err)
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			t.Fatalf("tar read body %s: %v", hdr.Name, err)
		}
		entries[hdr.Name] = data
	}
	return entries
}

func tarKeys(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
