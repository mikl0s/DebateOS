// Package config resolves the DebateOS configuration directory.
//
// The directory is resolved in this order:
//  1. DEBATEOS_DIR environment variable (non-empty) — used verbatim.
//  2. os.UserConfigDir() joined with "debateos" (XDG: ~/.config/debateos).
//  3. Error with guidance to set DEBATEOS_DIR.
//
// Tests always set DEBATEOS_DIR via t.Setenv("DEBATEOS_DIR", t.TempDir())
// so that os.UserConfigDir (which reads HOME/XDG_CONFIG_HOME) is not called
// in CI environments where HOME may be unset.
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// DebateOSDir returns the path to the DebateOS configuration directory.
//
// It checks DEBATEOS_DIR first (CI / test override), then falls back to
// os.UserConfigDir() + "/debateos". Returns a non-nil error when neither
// source is available, with a message that names DEBATEOS_DIR so the user
// knows exactly what to set.
func DebateOSDir() (string, error) {
	if d := os.Getenv("DEBATEOS_DIR"); d != "" {
		return d, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine config dir: %w\n(set DEBATEOS_DIR or HOME)", err)
	}
	return filepath.Join(base, "debateos"), nil
}
