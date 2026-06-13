package compose_test

import (
	"bytes"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mikl0s/debateos/cli/compose"
)

// omarchyDir returns the path to the omarchy example directory relative to
// the module root, using the test file's location to avoid working-dir issues.
func omarchyDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// file is <repo>/cli/compose/compose_test.go; go up 2 dirs to reach repo root.
	root := filepath.Join(filepath.Dir(file), "..", "..")
	return filepath.Join(root, "examples", "omarchy")
}

func TestCompose_OmarchyHappyPath(t *testing.T) {
	t.Setenv("DEBATEOS_DIR", omarchyDir(t))

	var stdout, stderr bytes.Buffer
	code := compose.Run([]string{}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d\nstdout: %s\nstderr: %s",
			code, stdout.String(), stderr.String())
	}

	out := stdout.String()
	// Must print Applied/Skipped counts.
	if !strings.Contains(out, "Applied:") {
		t.Errorf("output missing 'Applied:'; got:\n%s", out)
	}
	if !strings.Contains(out, "Skipped:") {
		t.Errorf("output missing 'Skipped:'; got:\n%s", out)
	}
	// Must print at least one explanation line.
	if !strings.Contains(out, "  -") {
		t.Errorf("output missing explanation lines (expected lines starting with '  -'); got:\n%s", out)
	}
}

func TestCompose_DirFlag(t *testing.T) {
	dir := omarchyDir(t)
	t.Setenv("DEBATEOS_DIR", "") // clear so --dir flag is the only source

	var stdout, stderr bytes.Buffer
	code := compose.Run([]string{"--dir", dir}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0 with --dir flag, got %d\nstdout: %s\nstderr: %s",
			code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Applied:") {
		t.Errorf("--dir flag output missing 'Applied:'; got:\n%s", stdout.String())
	}
}

func TestCompose_MissingDir(t *testing.T) {
	t.Setenv("DEBATEOS_DIR", "/nonexistent/path/that/should/not/exist")

	var stdout, stderr bytes.Buffer
	code := compose.Run([]string{}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit for missing dir, got 0")
	}
	if stderr.Len() == 0 {
		t.Error("expected error message on stderr for missing dir")
	}
}

func TestCompose_InvalidFlag(t *testing.T) {
	t.Setenv("DEBATEOS_DIR", omarchyDir(t))

	var stdout, stderr bytes.Buffer
	code := compose.Run([]string{"--unknown-flag"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected non-zero exit for unknown flag, got 0")
	}
}
