package main

// forumctl_test.go — tests for the forumctl binary entrypoint utilities.
// These live in the main package (not _test) to access unexported helpers.

import (
	"os"
	"testing"
)

// TestEnvOrDefault verifies the envOrDefault helper function.
func TestEnvOrDefault(t *testing.T) {
	const testKey = "DEBATEOS_TEST_ENV_KEY_XYZ"

	// Unset: should return default.
	os.Unsetenv(testKey)
	got := envOrDefault(testKey, "fallback")
	if got != "fallback" {
		t.Errorf("unset env: expected 'fallback', got %q", got)
	}

	// Set to empty string: should return default (empty env vars are ignored).
	os.Setenv(testKey, "")
	defer os.Unsetenv(testKey)
	got = envOrDefault(testKey, "fallback")
	if got != "fallback" {
		t.Errorf("empty env: expected 'fallback', got %q", got)
	}

	// Set to a real value: should return that value.
	os.Setenv(testKey, "realvalue")
	got = envOrDefault(testKey, "fallback")
	if got != "realvalue" {
		t.Errorf("set env: expected 'realvalue', got %q", got)
	}
}

// TestEnvOrDefaultMultiple tests multiple env keys at once.
func TestEnvOrDefaultMultiple(t *testing.T) {
	// Verify that different keys work independently.
	os.Setenv("DEBATEOS_TEST_A", "valueA")
	os.Setenv("DEBATEOS_TEST_B", "valueB")
	defer func() {
		os.Unsetenv("DEBATEOS_TEST_A")
		os.Unsetenv("DEBATEOS_TEST_B")
	}()

	if got := envOrDefault("DEBATEOS_TEST_A", "x"); got != "valueA" {
		t.Errorf("key A: expected 'valueA', got %q", got)
	}
	if got := envOrDefault("DEBATEOS_TEST_B", "y"); got != "valueB" {
		t.Errorf("key B: expected 'valueB', got %q", got)
	}
	if got := envOrDefault("DEBATEOS_TEST_MISSING", "default"); got != "default" {
		t.Errorf("missing key: expected 'default', got %q", got)
	}
}
