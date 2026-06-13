package runner_test

import (
	"testing"

	"github.com/mikl0s/debateos/cli/runner"
)

func TestFakeRunner_Run(t *testing.T) {
	f := &runner.FakeRunner{}
	err := f.Run("echo", "hello", "world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(f.Calls))
	}
	if f.Calls[0] != "echo hello world" {
		t.Fatalf("got call %q, want %q", f.Calls[0], "echo hello world")
	}
}

func TestFakeRunner_RunError(t *testing.T) {
	f := &runner.FakeRunner{Err: errTestError}
	err := f.Run("docker", "build", ".")
	if err != errTestError {
		t.Fatalf("got %v, want errTestError", err)
	}
	if len(f.Calls) != 1 {
		t.Fatalf("expected 1 call recorded even on error, got %d", len(f.Calls))
	}
}

func TestFakeRunner_Output(t *testing.T) {
	f := &runner.FakeRunner{
		Outputs: map[string][]byte{
			"echo hello": []byte("hello"),
		},
	}
	out, err := f.Output("echo", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "hello" {
		t.Fatalf("got %q, want %q", string(out), "hello")
	}
	if len(f.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(f.Calls))
	}
	if f.Calls[0] != "echo hello" {
		t.Fatalf("got call %q, want %q", f.Calls[0], "echo hello")
	}
}

func TestFakeRunner_OutputMissing(t *testing.T) {
	f := &runner.FakeRunner{}
	out, err := f.Output("git", "rev-parse", "HEAD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty output for unregistered key, got %q", string(out))
	}
}

func TestExecRunner_Run(t *testing.T) {
	r := runner.ExecRunner{}
	// `true` exits 0 on all POSIX systems
	if err := r.Run("true"); err != nil {
		t.Fatalf("ExecRunner.Run(true) unexpected error: %v", err)
	}
}

func TestExecRunner_Output(t *testing.T) {
	r := runner.ExecRunner{}
	out, err := r.Output("echo", "-n", "ping")
	if err != nil {
		t.Fatalf("ExecRunner.Output unexpected error: %v", err)
	}
	if string(out) != "ping" {
		t.Fatalf("got %q, want %q", string(out), "ping")
	}
}

// errTestError is a sentinel error for testing.
var errTestError = &testErr{}

type testErr struct{}

func (e *testErr) Error() string { return "test error" }
