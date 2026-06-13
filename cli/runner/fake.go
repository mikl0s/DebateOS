package runner

import "strings"

// FakeRunner is a test double that records every call to Run and Output.
// It satisfies the Runner interface and lets tests assert on exact argv
// without starting real subprocesses.
//
// Usage:
//
//	f := &FakeRunner{Outputs: map[string][]byte{"echo hello": []byte("hello")}}
//	f.Run("docker", "build", ".")      // appends "docker build ." to f.Calls
//	f.Output("echo", "hello")          // appends "echo hello"; returns []byte("hello")
type FakeRunner struct {
	// Calls records each invocation as "name arg1 arg2 ..." in order.
	Calls []string

	// Err is returned by every Run and Output call (nil = success).
	Err error

	// Outputs maps join-key ("name arg1 arg2 ...") to the []byte to return
	// from Output. Missing keys return nil (no output).
	Outputs map[string][]byte
}

// Run records the call and returns f.Err.
func (f *FakeRunner) Run(name string, args ...string) error {
	key := joinKey(name, args)
	f.Calls = append(f.Calls, key)
	return f.Err
}

// Output records the call, returns f.Outputs[key] and f.Err.
func (f *FakeRunner) Output(name string, args ...string) ([]byte, error) {
	key := joinKey(name, args)
	f.Calls = append(f.Calls, key)
	return f.Outputs[key], f.Err
}

// joinKey builds the canonical call key used for Calls and Outputs lookup.
func joinKey(name string, args []string) string {
	if len(args) == 0 {
		return name
	}
	return name + " " + strings.Join(args, " ")
}
