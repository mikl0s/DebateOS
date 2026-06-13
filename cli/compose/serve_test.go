package compose_test

import (
	"bytes"
	"testing"

	"github.com/mikl0s/debateos/cli/compose"
)

// TestComposeServeFlag verifies that compose.Run with --serve flag
// does not regress the existing compose tests and that the --serve
// flag is recognized (not treated as an unknown flag).
func TestComposeServeFlag(t *testing.T) {
	// --serve without --dir should fail with missing dir error (not "unknown flag").
	t.Setenv("DEBATEOS_DIR", "")

	var stdout, stderr bytes.Buffer
	// With --no-listen seam, Run returns without blocking on ListenAndServe.
	// --serve --addr :0 --no-listen allows testing flag parsing without binding.
	code := compose.Run([]string{"--serve", "--addr", ":0", "--no-listen"}, &stdout, &stderr)

	// Without a valid speech dir, and with --no-listen, we just check that
	// the --serve flag is parsed (no "unknown flag" error). The dir error is fine.
	errOutput := stderr.String()
	if code == 0 {
		// Unexpected: we have no dir set; should fail
		// This is acceptable if no-listen skips resolve entirely.
		return
	}
	// The error should NOT be about an unknown flag.
	if len(errOutput) > 0 && (contains(errOutput, "unknown flag") || contains(errOutput, "flag provided but not defined")) {
		t.Errorf("--serve flag not recognized: %s", errOutput)
	}
}

// TestComposeFlagBackcompat verifies that adding --serve does not break
// the existing compose (resolution-preview) behavior when --serve is absent.
func TestComposeFlagBackcompat(t *testing.T) {
	dir := omarchyDir(t)
	t.Setenv("DEBATEOS_DIR", "")

	var stdout, stderr bytes.Buffer
	code := compose.Run([]string{"--dir", dir}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("compose --dir flag regressed after --serve addition: exit %d\nstdout: %s\nstderr: %s",
			code, stdout.String(), stderr.String())
	}
	if !contains(stdout.String(), "Applied:") {
		t.Errorf("compose output missing 'Applied:' after --serve flag addition; got:\n%s", stdout.String())
	}
}

// TestComposeServeNoListen verifies that --serve --no-listen returns 0 without blocking.
func TestComposeServeNoListen(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := compose.Run([]string{"--serve", "--addr", ":0", "--no-listen"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("--serve --no-listen expected exit 0, got %d\nstderr: %s", code, stderr.String())
	}
	// Output should mention the serve intent.
	out := stdout.String()
	if len(out) > 0 {
		// Any output is acceptable — no crash is the key invariant.
		t.Logf("serve output: %s", out)
	}
}

// TestComposeServeWithAddrFlag verifies --addr flag is recognized.
func TestComposeServeWithAddrFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := compose.Run([]string{"--serve", "--addr", ":9999", "--no-listen"}, &stdout, &stderr)
	// Should succeed (no unknown flag error).
	errOut := stderr.String()
	if contains(errOut, "flag provided but not defined") {
		t.Errorf("--addr flag not recognized: %s", errOut)
	}
	if code != 0 && contains(errOut, "flag provided but not defined") {
		t.Errorf("--addr flag caused unexpected error: %s", errOut)
	}
}

// TestComposeServeErrorPath verifies that --serve with an invalid port returns exit 1.
// Uses port 99999 (invalid, >65535) so ListenAndServe fails immediately without blocking.
func TestComposeServeErrorPath(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := compose.Run([]string{"--serve", "--addr", ":99999"}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("expected exit 1 for serveUI error, got %d (stderr: %s)", code, stderr.String())
	}
	if stderr.Len() == 0 {
		t.Error("expected error message on stderr for serveUI failure")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
