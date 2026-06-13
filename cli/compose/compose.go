// Package compose implements the `debateos compose` subcommand.
//
// It resolves the configured speech directory and prints a human-readable
// resolution preview to stdout: Applied/Skipped/Dropped counts followed by
// each explanation .Text line. Returns 0 on success, 1 on error.
//
// Signature: Run(args []string, stdout, stderr io.Writer) int
// Never calls os.Exit — main() is the sole caller of os.Exit.
package compose

import (
	"flag"
	"fmt"
	"io"

	"github.com/mikl0s/debateos/cli/config"
	"github.com/mikl0s/debateos/cli/internal/loader"
)

// Run is the compose subcommand entry point. It parses args (--dir flag),
// resolves the speech directory, and prints a resolution preview.
func Run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("compose", flag.ContinueOnError)
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
			fmt.Fprintf(stderr, "compose: %v\n", err)
			return 1
		}
	}

	rs, err := loader.ResolveDir(speechDir)
	if err != nil {
		fmt.Fprintf(stderr, "compose: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Resolution preview for: %s\n", speechDir)
	fmt.Fprintf(stdout, "Applied: %d  Skipped: %d  Dropped: %d\n",
		len(rs.Applied), len(rs.Skipped), len(rs.Dropped))
	fmt.Fprintln(stdout)

	if len(rs.Explanations) > 0 {
		fmt.Fprintln(stdout, "Explanations:")
		for _, ex := range rs.Explanations {
			fmt.Fprintf(stdout, "  - %s\n", ex.Text)
		}
	}

	return 0
}
