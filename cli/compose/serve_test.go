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
