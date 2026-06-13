// cmd/debateos — the debateos CLI entry point.
//
// Dispatches subcommands to their respective packages. Each subcommand package
// exposes a Run(args []string, stdout, stderr io.Writer) int function that
// returns an exit code without ever calling os.Exit. main() is the ONLY place
// that calls os.Exit.
//
// Usage:
//
//	debateos <command> [flags]
//
// Commands:
//
//	compose   Print a resolution preview with explanations.
//	validate  Parse, schema-validate, and clean-resolve the speech (CI-friendly).
//	build     Resolve → translate → docker ISO build (--dry-run / --skip-iso supported).
//	pane      Manage the private pane (set/get/list/backup/restore).
package main

import (
	"fmt"
	"os"

	"github.com/mikl0s/debateos/cli/build"
	"github.com/mikl0s/debateos/cli/compose"
	"github.com/mikl0s/debateos/cli/pane"
	"github.com/mikl0s/debateos/cli/runner"
	"github.com/mikl0s/debateos/cli/validate"
)

const usage = `usage: debateos <command> [flags]

Commands:
  compose   Print a resolution preview with full explanations.
  validate  Parse + schema-validate + clean-resolve gate (CI-friendly; exits non-zero on failure).
  build     Resolve speech → emit arch profile → docker ISO build (--dry-run / --skip-iso).
  pane      Manage the private pane: set/get/list/backup/restore (see 'debateos pane --help').

Run 'debateos <command> --help' for per-command flags.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "compose":
		os.Exit(compose.Run(os.Args[2:], os.Stdout, os.Stderr))
	case "validate":
		os.Exit(validate.Run(os.Args[2:], os.Stdout, os.Stderr))
	case "build":
		os.Exit(build.Run(os.Args[2:], os.Stdout, os.Stderr, runner.ExecRunner{}))
	case "pane":
		os.Exit(pane.Run(os.Args[2:], os.Stdout, os.Stderr, runner.ExecRunner{}))
	default:
		fmt.Fprintf(os.Stderr, "debateos: unknown command %q\n\n", os.Args[1])
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}
}
