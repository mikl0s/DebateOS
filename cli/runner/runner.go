// Package runner defines the Runner interface and its ExecRunner production
// implementation. All external subprocess invocations in the debateos CLI
// (docker, git, translators/arch/translate, sha256sum) go through this
// interface so tests can record and assert calls without real binaries.
//
// Security: ExecRunner.Run / ExecRunner.Output always use exec.Command(name,
// args...) with variadic args — never "sh -c" string interpolation. This
// prevents arg-injection attacks on user-controlled paths (T-03-AI).
package runner

import "os/exec"

// Runner is the interface for external subprocess invocations.
type Runner interface {
	// Run executes the named binary with args and returns any non-zero exit
	// error. Combined stdout/stderr goes to the inherited file descriptors.
	Run(name string, args ...string) error

	// Output executes the named binary with args and returns its stdout as
	// bytes. Stderr goes to the inherited file descriptor.
	Output(name string, args ...string) ([]byte, error)
}

// ExecRunner is the production Runner that delegates to os/exec.
type ExecRunner struct{}

// Run executes the named binary with args using exec.Command.
func (ExecRunner) Run(name string, args ...string) error {
	return exec.Command(name, args...).Run()
}

// Output executes the named binary with args using exec.Command and returns stdout.
func (ExecRunner) Output(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}
