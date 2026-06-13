package config_test

import (
	"strings"
	"testing"

	"github.com/mikl0s/debateos/cli/config"
)

func TestDebateOSDir_EnvOverride(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DEBATEOS_DIR", tmp)

	got, err := config.DebateOSDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != tmp {
		t.Fatalf("got %q, want %q", got, tmp)
	}
}

func TestDebateOSDir_XDGFallback(t *testing.T) {
	// Unset DEBATEOS_DIR; HOME must be set for os.UserConfigDir to succeed.
	t.Setenv("DEBATEOS_DIR", "")
	// Leave HOME as-is (it is set on this host); we just verify the path ends with /debateos.

	got, err := config.DebateOSDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(got, "/debateos") {
		t.Fatalf("expected path ending with /debateos, got %q", got)
	}
}

func TestDebateOSDir_NoHomeError(t *testing.T) {
	// Clear all env vars that os.UserConfigDir relies on.
	t.Setenv("DEBATEOS_DIR", "")
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	_, err := config.DebateOSDir()
	if err == nil {
		t.Fatal("expected error when HOME and XDG_CONFIG_HOME are unset, got nil")
	}
	if !strings.Contains(err.Error(), "DEBATEOS_DIR") {
		t.Fatalf("error message %q does not contain 'DEBATEOS_DIR'", err.Error())
	}
}
