// Package validate implements the `debateos validate` subcommand.
//
// It parses a speech directory (speech.yaml + points/ + opinions/),
// schema-validates all referenced documents, and runs a clean-resolve check.
// Returns 0 on success (fully valid + cleanly-resolving speech), non-zero on
// any parse, schema-validation, or hard-conflict resolve failure.
//
// Signature: Run(args []string, stdout, stderr io.Writer) int
// Never calls os.Exit — main() is the sole caller of os.Exit.
package validate

import (
	"flag"
	"fmt"
	"io"

	"github.com/mikl0s/debateos/cli/config"
	"github.com/mikl0s/debateos/cli/internal/loader"
)

// Run is the validate subcommand entry point. It parses args (--dir flag),
// resolves the speech directory, and reports OK or the failure reason.
func Run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	dir := fs.String("dir", "", "speech directory (overrides DEBATEOS_DIR)")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	speechDir := *dir
	if speechDir == "" {
		var err error
		speechDir, err = config.DebateOSDir()
		if err != nil {
			fmt.Fprintf(stderr, "validate: %v\n", err)
			return 1
		}
	}

	rs, err := loader.ResolveDir(speechDir)
	if err != nil {
		// Surface partial hard-conflict explanations if available.
		if rs != nil {
			for _, ex := range rs.Explanations {
				fmt.Fprintf(stderr, "validate: conflict: %s\n", ex.Text)
			}
		}
		fmt.Fprintf(stderr, "validate: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "validate: OK — Applied=%d Skipped=%d Dropped=%d\n",
		len(rs.Applied), len(rs.Skipped), len(rs.Dropped))
	return 0
}
