// Package pane_test contains tests for the age identity, encrypt/decrypt,
// and pane set/get/list/backup/restore functionality.
package pane_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mikl0s/debateos/cli/pane"
	"github.com/mikl0s/debateos/cli/runner"
)

// ─── Task 1: age identity + encrypt/decrypt ───────────────────────────────

// TestIdentityCreation verifies that LoadOrCreateIdentity generates identity.age
// 0600 on first call and returns the same identity on subsequent calls.
func TestIdentityCreation(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)

	id1, err := pane.LoadOrCreateIdentity(dir)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	if id1 == nil {
		t.Fatal("identity is nil")
	}

	// identity.age must be created 0600
	idPath := filepath.Join(dir, "identity.age")
	info, err := os.Stat(idPath)
	if err != nil {
		t.Fatalf("identity.age not found: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("identity.age perm = %04o, want 0600", perm)
	}

	// second call returns same public key (same identity)
	id2, err := pane.LoadOrCreateIdentity(dir)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if id1.Recipient().String() != id2.Recipient().String() {
		t.Errorf("identity changed between calls: %s vs %s",
			id1.Recipient(), id2.Recipient())
	}
}

// TestAgeRoundTrip verifies EncryptFile + DecryptFile are byte-identical.
func TestAgeRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)

	id, err := pane.LoadOrCreateIdentity(dir)
	if err != nil {
		t.Fatalf("identity: %v", err)
	}

	// write plaintext source
	plaintext := []byte("api-key: super-secret-value\ntoken: abc123\n")
	src := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(src, plaintext, 0600); err != nil {
		t.Fatal(err)
	}

	// encrypt
	enc := filepath.Join(dir, "source.txt.age")
	if err := pane.EncryptFile(id, src, enc); err != nil {
		t.Fatalf("EncryptFile: %v", err)
	}

	// encrypted file must be 0600
	info, err := os.Stat(enc)
	if err != nil {
		t.Fatalf("encrypted file not found: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("encrypted file perm = %04o, want 0600", perm)
	}

	// encrypted != plaintext (prove real encryption)
	encBytes, _ := os.ReadFile(enc)
	if bytes.Equal(encBytes, plaintext) {
		t.Error("encrypted file equals plaintext — no encryption occurred")
	}

	// decrypt
	dec := filepath.Join(dir, "source.txt.dec")
	if err := pane.DecryptFile(id, enc, dec); err != nil {
		t.Fatalf("DecryptFile: %v", err)
	}

	// must be byte-identical to original
	got, _ := os.ReadFile(dec)
	if !bytes.Equal(got, plaintext) {
		t.Errorf("round-trip mismatch:\nwant: %q\n got: %q", plaintext, got)
	}
}

// TestAgeRoundTripWrongIdentity verifies that decrypting with a different
// identity returns an error (proves real encryption, not a copy).
func TestAgeRoundTripWrongIdentity(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)

	id1, err := pane.LoadOrCreateIdentity(dir)
	if err != nil {
		t.Fatal(err)
	}

	// encrypt with id1
	plaintext := []byte("secret: value\n")
	src := filepath.Join(dir, "plain.txt")
	if err := os.WriteFile(src, plaintext, 0600); err != nil {
		t.Fatal(err)
	}
	enc := filepath.Join(dir, "plain.txt.age")
	if err := pane.EncryptFile(id1, src, enc); err != nil {
		t.Fatal(err)
	}

	// generate a second identity in a different dir
	dir2 := t.TempDir()
	id2, err := pane.LoadOrCreateIdentity(dir2)
	if err != nil {
		t.Fatal(err)
	}

	// decrypt with id2 must fail
	out := filepath.Join(dir, "plain.txt.bad")
	if err := pane.DecryptFile(id2, enc, out); err == nil {
		t.Error("expected error decrypting with wrong identity, got nil")
	}
}

// ─── Task 2: pane set/get/list/backup/restore ─────────────────────────────

// TestPaneSetGet verifies set writes pane.yaml (0600) and get reads it back.
func TestPaneSetGet(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)

	fr := &runner.FakeRunner{}

	// set
	code := pane.Run([]string{"set", "mykey", "myvalue"}, &bytes.Buffer{}, &bytes.Buffer{}, fr)
	if code != 0 {
		t.Fatalf("set returned %d", code)
	}

	// pane.yaml must be 0600
	paneYAML := filepath.Join(dir, "pane.yaml")
	info, err := os.Stat(paneYAML)
	if err != nil {
		t.Fatalf("pane.yaml not found: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("pane.yaml perm = %04o, want 0600", perm)
	}

	// get
	var stdout bytes.Buffer
	code = pane.Run([]string{"get", "mykey"}, &stdout, &bytes.Buffer{}, fr)
	if code != 0 {
		t.Fatalf("get returned %d", code)
	}
	if got := strings.TrimSpace(stdout.String()); got != "myvalue" {
		t.Errorf("get: want %q got %q", "myvalue", got)
	}
}

// TestPaneGetMissing verifies get of a missing key returns non-zero exit.
func TestPaneGetMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)
	fr := &runner.FakeRunner{}

	code := pane.Run([]string{"get", "nonexistent"}, &bytes.Buffer{}, &bytes.Buffer{}, fr)
	if code == 0 {
		t.Error("expected non-zero exit for missing key")
	}
}

// TestPaneList verifies list prints all keys.
func TestPaneList(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)
	fr := &runner.FakeRunner{}

	pane.Run([]string{"set", "alpha", "1"}, &bytes.Buffer{}, &bytes.Buffer{}, fr)
	pane.Run([]string{"set", "beta", "2"}, &bytes.Buffer{}, &bytes.Buffer{}, fr)

	var stdout bytes.Buffer
	code := pane.Run([]string{"list"}, &stdout, &bytes.Buffer{}, fr)
	if code != 0 {
		t.Fatalf("list returned %d", code)
	}
	out := stdout.String()
	if !strings.Contains(out, "alpha") {
		t.Errorf("list missing 'alpha': %q", out)
	}
	if !strings.Contains(out, "beta") {
		t.Errorf("list missing 'beta': %q", out)
	}
}

// TestPanePermissions verifies pane.yaml stays 0600 after multiple set calls.
func TestPanePermissions(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)
	fr := &runner.FakeRunner{}

	pane.Run([]string{"set", "k1", "v1"}, &bytes.Buffer{}, &bytes.Buffer{}, fr)
	pane.Run([]string{"set", "k2", "v2"}, &bytes.Buffer{}, &bytes.Buffer{}, fr)

	info, err := os.Stat(filepath.Join(dir, "pane.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("pane.yaml perm after multiple sets = %04o, want 0600", perm)
	}
}

// TestPaneBackup verifies backup encrypts and invokes git add/commit/push via FakeRunner.
func TestPaneBackup(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)
	fr := &runner.FakeRunner{}

	// set a value first so pane.yaml exists
	pane.Run([]string{"set", "secret", "supersecret"}, &bytes.Buffer{}, &bytes.Buffer{}, fr)
	fr.Calls = nil // reset call log

	code := pane.Run([]string{"backup"}, &bytes.Buffer{}, &bytes.Buffer{}, fr)
	if code != 0 {
		t.Fatalf("backup returned %d", code)
	}

	// pane.yaml.age must exist and be 0600
	ageFile := filepath.Join(dir, "pane.yaml.age")
	info, err := os.Stat(ageFile)
	if err != nil {
		t.Fatalf("pane.yaml.age not found: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("pane.yaml.age perm = %04o, want 0600", perm)
	}

	// FakeRunner must have recorded git add, git commit, git push (in order)
	calls := strings.Join(fr.Calls, "\n")
	if !strings.Contains(calls, "git add") {
		t.Errorf("backup: no 'git add' call in: %v", fr.Calls)
	}
	if !strings.Contains(calls, "git commit") {
		t.Errorf("backup: no 'git commit' call in: %v", fr.Calls)
	}
	if !strings.Contains(calls, "git push") {
		t.Errorf("backup: no 'git push' call in: %v", fr.Calls)
	}

	// plaintext pane.yaml must NOT be staged (only .age committed)
	for _, c := range fr.Calls {
		if strings.Contains(c, "git add") && strings.Contains(c, "pane.yaml") &&
			!strings.Contains(c, "pane.yaml.age") {
			t.Errorf("backup staged plaintext pane.yaml: %q", c)
		}
	}
}

// TestPaneRestore verifies restore decrypts pane.yaml.age back to pane.yaml
// with byte-identical content, and pane.yaml stays 0600.
func TestPaneRestore(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)
	fr := &runner.FakeRunner{}

	// set a value and backup
	pane.Run([]string{"set", "restorekey", "restoreval"}, &bytes.Buffer{}, &bytes.Buffer{}, fr)
	pane.Run([]string{"backup"}, &bytes.Buffer{}, &bytes.Buffer{}, fr)
	fr.Calls = nil

	// capture original pane.yaml content
	orig, err := os.ReadFile(filepath.Join(dir, "pane.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	// delete pane.yaml to simulate loss
	if err := os.Remove(filepath.Join(dir, "pane.yaml")); err != nil {
		t.Fatal(err)
	}

	// restore
	code := pane.Run([]string{"restore"}, &bytes.Buffer{}, &bytes.Buffer{}, fr)
	if code != 0 {
		t.Fatalf("restore returned %d", code)
	}

	// pane.yaml must be recreated with original content
	restored, err := os.ReadFile(filepath.Join(dir, "pane.yaml"))
	if err != nil {
		t.Fatalf("pane.yaml not restored: %v", err)
	}
	if !bytes.Equal(orig, restored) {
		t.Errorf("restore mismatch:\nwant: %q\n got: %q", orig, restored)
	}

	// pane.yaml must be 0600 after restore
	info, err := os.Stat(filepath.Join(dir, "pane.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("restored pane.yaml perm = %04o, want 0600", perm)
	}

	// restore must have called git pull or git fetch
	calls := strings.Join(fr.Calls, "\n")
	if !strings.Contains(calls, "git pull") && !strings.Contains(calls, "git fetch") {
		t.Errorf("restore: no 'git pull'/'git fetch' in: %v", fr.Calls)
	}
}

// ─── Additional error-path and edge-case tests ────────────────────────────

// TestPaneRunNoArgs verifies that calling Run with no args returns an error.
func TestPaneRunNoArgs(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)
	fr := &runner.FakeRunner{}

	var stdout, stderr bytes.Buffer
	code := pane.Run([]string{}, &stdout, &stderr, fr)
	if code == 0 {
		t.Fatal("expected non-zero exit for no args, got 0")
	}
	if stderr.Len() == 0 {
		t.Error("expected usage on stderr")
	}
}

// TestPaneRunUnknownVerb verifies that an unknown verb returns an error.
func TestPaneRunUnknownVerb(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)
	fr := &runner.FakeRunner{}

	var stdout, stderr bytes.Buffer
	code := pane.Run([]string{"frobulate"}, &stdout, &stderr, fr)
	if code == 0 {
		t.Fatal("expected non-zero exit for unknown verb, got 0")
	}
	if !strings.Contains(stderr.String(), "frobulate") {
		t.Errorf("expected verb name in error: %s", stderr.String())
	}
}

// TestPaneSetNoArgs verifies that `pane set` without key/value returns an error.
func TestPaneSetNoArgs(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)
	fr := &runner.FakeRunner{}

	var stdout, stderr bytes.Buffer
	code := pane.Run([]string{"set"}, &stdout, &stderr, fr)
	if code == 0 {
		t.Fatal("expected non-zero exit for 'pane set' with no args, got 0")
	}
}

// TestPaneGetNotFound verifies that `pane get <missing-key>` returns non-zero.
func TestPaneGetNotFound(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)
	fr := &runner.FakeRunner{}

	var stdout, stderr bytes.Buffer
	code := pane.Run([]string{"get", "does-not-exist"}, &stdout, &stderr, fr)
	if code == 0 {
		t.Fatal("expected non-zero exit for missing key, got 0")
	}
}

// TestPaneGetNoArgs verifies that `pane get` with no key returns an error.
func TestPaneGetNoArgs(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)
	fr := &runner.FakeRunner{}

	var stdout, stderr bytes.Buffer
	code := pane.Run([]string{"get"}, &stdout, &stderr, fr)
	if code == 0 {
		t.Fatal("expected non-zero exit for 'pane get' with no key, got 0")
	}
}

// TestPaneSetGetRoundTrip verifies `pane set` then `pane get` returns the value.
func TestPaneSetGetRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)
	fr := &runner.FakeRunner{}

	// set
	var out, errBuf bytes.Buffer
	code := pane.Run([]string{"set", "mykey", "myvalue"}, &out, &errBuf, fr)
	if code != 0 {
		t.Fatalf("set returned %d: %s", code, errBuf.String())
	}

	// get
	out.Reset()
	errBuf.Reset()
	code = pane.Run([]string{"get", "mykey"}, &out, &errBuf, fr)
	if code != 0 {
		t.Fatalf("get returned %d: %s", code, errBuf.String())
	}
	if !strings.Contains(out.String(), "myvalue") {
		t.Errorf("expected 'myvalue' in get output, got: %s", out.String())
	}
}

// TestPaneListMultipleKeys verifies `pane list` shows all keys that were set.
func TestPaneListMultipleKeys(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)
	fr := &runner.FakeRunner{}

	pane.Run([]string{"set", "alpha", "1"}, &bytes.Buffer{}, &bytes.Buffer{}, fr)
	pane.Run([]string{"set", "beta", "2"}, &bytes.Buffer{}, &bytes.Buffer{}, fr)

	var out, errBuf bytes.Buffer
	code := pane.Run([]string{"list"}, &out, &errBuf, fr)
	if code != 0 {
		t.Fatalf("list returned %d: %s", code, errBuf.String())
	}
	got := out.String()
	if !strings.Contains(got, "alpha") {
		t.Errorf("expected 'alpha' in list output: %s", got)
	}
	if !strings.Contains(got, "beta") {
		t.Errorf("expected 'beta' in list output: %s", got)
	}
}

// TestPaneDirFlag verifies that --dir flag overrides DEBATEOS_DIR.
func TestPaneDirFlag(t *testing.T) {
	dir := t.TempDir()
	// deliberately set DEBATEOS_DIR to something else
	t.Setenv("DEBATEOS_DIR", "/nonexistent-should-not-be-used")
	fr := &runner.FakeRunner{}

	var out, errBuf bytes.Buffer
	code := pane.Run([]string{"--dir", dir, "set", "flagtest", "flagvalue"}, &out, &errBuf, fr)
	if code != 0 {
		t.Fatalf("set with --dir flag returned %d: %s", code, errBuf.String())
	}

	// Now get it
	out.Reset()
	errBuf.Reset()
	code = pane.Run([]string{"--dir", dir, "get", "flagtest"}, &out, &errBuf, fr)
	if code != 0 {
		t.Fatalf("get with --dir flag returned %d: %s", code, errBuf.String())
	}
	if !strings.Contains(out.String(), "flagvalue") {
		t.Errorf("expected 'flagvalue' in output: %s", out.String())
	}
}

// TestPaneBackupFailsWithNoPaneYaml verifies that backup fails gracefully
// when pane.yaml does not exist (nothing to back up).
func TestPaneBackupFailsWithNoPaneYaml(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)
	fr := &runner.FakeRunner{}

	// Don't create pane.yaml — backup should fail at encrypt step.
	var out, errBuf bytes.Buffer
	code := pane.Run([]string{"backup"}, &out, &errBuf, fr)
	if code == 0 {
		t.Fatal("expected non-zero exit for backup with missing pane.yaml, got 0")
	}
}

// TestPaneBackupGitAddError verifies that a git add failure from the Runner
// propagates to a non-zero exit.
func TestPaneBackupGitAddError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)

	// First set a value so pane.yaml exists.
	fr := &runner.FakeRunner{}
	pane.Run([]string{"set", "k", "v"}, &bytes.Buffer{}, &bytes.Buffer{}, fr)

	// Now use a runner that fails on git add.
	fr2 := &runner.FakeRunner{Err: fmt.Errorf("git add failed")}
	var out, errBuf bytes.Buffer
	code := pane.Run([]string{"backup"}, &out, &errBuf, fr2)
	if code == 0 {
		t.Fatal("expected non-zero exit when git add fails, got 0")
	}
}

// TestPaneRestoreGitPullError verifies that a git pull failure propagates
// to a non-zero exit.
func TestPaneRestoreGitPullError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)

	fr := &runner.FakeRunner{Err: fmt.Errorf("git pull failed")}
	var out, errBuf bytes.Buffer
	code := pane.Run([]string{"restore"}, &out, &errBuf, fr)
	if code == 0 {
		t.Fatal("expected non-zero exit when git pull fails, got 0")
	}
}

// TestPaneRestoreNoAgeFile verifies that restore fails when pane.yaml.age
// does not exist (no backup to restore from).
func TestPaneRestoreNoAgeFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)

	// FakeRunner succeeds for git pull but no .age file exists.
	fr := &runner.FakeRunner{}
	var out, errBuf bytes.Buffer
	code := pane.Run([]string{"restore"}, &out, &errBuf, fr)
	if code == 0 {
		t.Fatal("expected non-zero exit when pane.yaml.age is missing, got 0")
	}
}

// TestPaneEncryptFileMissingSrc verifies that EncryptFile fails cleanly
// when the source file does not exist.
func TestPaneEncryptFileMissingSrc(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)

	id, err := pane.LoadOrCreateIdentity(dir)
	if err != nil {
		t.Fatalf("LoadOrCreateIdentity: %v", err)
	}

	err = pane.EncryptFile(id, filepath.Join(dir, "nonexistent.yaml"), filepath.Join(dir, "out.age"))
	if err == nil {
		t.Fatal("expected error for missing src, got nil")
	}
}

// TestPaneDecryptFileMissingSrc verifies that DecryptFile fails cleanly
// when the source encrypted file does not exist.
func TestPaneDecryptFileMissingSrc(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)

	id, err := pane.LoadOrCreateIdentity(dir)
	if err != nil {
		t.Fatalf("LoadOrCreateIdentity: %v", err)
	}

	err = pane.DecryptFile(id, filepath.Join(dir, "nonexistent.age"), filepath.Join(dir, "out.yaml"))
	if err == nil {
		t.Fatal("expected error for missing src, got nil")
	}
}

// TestPaneLoadOrCreateIdentityCorrupted verifies that a corrupted identity.age
// file returns an error rather than silently generating a new identity.
func TestPaneLoadOrCreateIdentityCorrupted(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)

	// Write a corrupt identity file.
	if err := os.WriteFile(filepath.Join(dir, "identity.age"), []byte("not-a-valid-age-key\n"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := pane.LoadOrCreateIdentity(dir)
	if err == nil {
		t.Fatal("expected error for corrupted identity.age, got nil")
	}
}

// TestPaneBackupGitCommitError verifies that a git commit failure propagates
// to a non-zero exit (the git add succeeds but commit fails).
func TestPaneBackupGitCommitError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)

	// Set a value so pane.yaml exists.
	fr := &runner.FakeRunner{}
	pane.Run([]string{"set", "key", "val"}, &bytes.Buffer{}, &bytes.Buffer{}, fr)

	// Runner fails only on git commit (not on git add).
	gitCommitErr := fmt.Errorf("git commit failed")
	callCount := 0
	fr2 := &runner.FakeRunnerFunc{
		RunFn: func(name string, args ...string) error {
			callCount++
			// First call is git add (should succeed).
			if callCount == 1 {
				return nil
			}
			// Second call is git commit (fail).
			return gitCommitErr
		},
	}

	var out, errBuf bytes.Buffer
	code := pane.Run([]string{"backup"}, &out, &errBuf, fr2)
	if code == 0 {
		t.Fatal("expected non-zero exit when git commit fails, got 0")
	}
}

// TestPaneSetBadLoadPane verifies that cmdSet fails gracefully when the existing
// pane.yaml is corrupt YAML (not the write path, but the load-before-set path).
func TestPaneSetBadLoadPane(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)
	fr := &runner.FakeRunner{}

	// Write a corrupt pane.yaml.
	if err := os.WriteFile(filepath.Join(dir, "pane.yaml"), []byte("this: [bad: {"), 0600); err != nil {
		t.Fatal(err)
	}

	var out, errBuf bytes.Buffer
	code := pane.Run([]string{"set", "k", "v"}, &out, &errBuf, fr)
	if code == 0 {
		t.Fatal("expected non-zero exit when pane.yaml is corrupt, got 0")
	}
}

// TestPaneGetBadLoadPane verifies that cmdGet fails gracefully when pane.yaml is corrupt.
func TestPaneGetBadLoadPane(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)
	fr := &runner.FakeRunner{}

	if err := os.WriteFile(filepath.Join(dir, "pane.yaml"), []byte("this: [bad: {"), 0600); err != nil {
		t.Fatal(err)
	}

	var out, errBuf bytes.Buffer
	code := pane.Run([]string{"get", "anykey"}, &out, &errBuf, fr)
	if code == 0 {
		t.Fatal("expected non-zero exit when pane.yaml is corrupt, got 0")
	}
}

// TestPaneListBadLoadPane verifies that cmdList fails gracefully when pane.yaml is corrupt.
func TestPaneListBadLoadPane(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEBATEOS_DIR", dir)
	fr := &runner.FakeRunner{}

	if err := os.WriteFile(filepath.Join(dir, "pane.yaml"), []byte("this: [bad: {"), 0600); err != nil {
		t.Fatal(err)
	}

	var out, errBuf bytes.Buffer
	code := pane.Run([]string{"list"}, &out, &errBuf, fr)
	if code == 0 {
		t.Fatal("expected non-zero exit when pane.yaml is corrupt, got 0")
	}
}
