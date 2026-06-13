package validate_test

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mikl0s/debateos/cli/validate"
)

// omarchyDir returns the path to the omarchy example directory.
func omarchyDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// file is <repo>/cli/validate/validate_test.go; go up 2 dirs to reach repo root.
	root := filepath.Join(filepath.Dir(file), "..", "..")
	return filepath.Join(root, "examples", "omarchy")
}

func TestValidateOmarchy(t *testing.T) {
	t.Setenv("DEBATEOS_DIR", omarchyDir(t))

	var stdout, stderr bytes.Buffer
	code := validate.Run([]string{}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0 for omarchy, got %d\nstdout: %s\nstderr: %s",
			code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "OK") {
		t.Errorf("expected 'OK' in stdout, got: %s", stdout.String())
	}
}

func TestValidate_DirFlag(t *testing.T) {
	dir := omarchyDir(t)
	t.Setenv("DEBATEOS_DIR", "") // clear env so --dir is the only source

	var stdout, stderr bytes.Buffer
	code := validate.Run([]string{"--dir", dir}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0 with --dir flag, got %d\nstdout: %s\nstderr: %s",
			code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "OK") {
		t.Errorf("expected 'OK' in stdout, got: %s", stdout.String())
	}
}

func TestValidate_BadSpeech(t *testing.T) {
	// Create a temp dir with a broken speech.yaml
	tmp := t.TempDir()
	t.Setenv("DEBATEOS_DIR", tmp)

	badSpeech := `this is not valid yaml: [{`
	if err := os.WriteFile(filepath.Join(tmp, "speech.yaml"), []byte(badSpeech), 0644); err != nil {
		t.Fatalf("write speech.yaml: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := validate.Run([]string{}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit for bad speech.yaml, got 0")
	}
	if stderr.Len() == 0 {
		t.Error("expected error message on stderr for bad speech")
	}
}

func TestValidate_MissingDir(t *testing.T) {
	t.Setenv("DEBATEOS_DIR", "/nonexistent/path/that/should/not/exist")

	var stdout, stderr bytes.Buffer
	code := validate.Run([]string{}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit for missing dir, got 0")
	}
}

func TestValidate_InvalidFlag(t *testing.T) {
	t.Setenv("DEBATEOS_DIR", omarchyDir(t))

	var stdout, stderr bytes.Buffer
	code := validate.Run([]string{"--unknown-flag"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit for unknown flag, got 0")
	}
}
