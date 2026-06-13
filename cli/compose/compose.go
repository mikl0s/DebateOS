// Package compose implements the `debateos compose` subcommand.
//
// It resolves the configured speech directory and prints a human-readable
// resolution preview to stdout: Applied/Skipped/Dropped counts followed by
// each explanation .Text line. Returns 0 on success, 1 on error.
//
// When the --serve flag is set, compose starts an HTTP server that serves
// the embedded SvelteKit UI at the given --addr (default ":8080"). The server
// blocks until it exits. The resolution preview is still printed before the
// server starts if a speech directory is configured.
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

// Run is the compose subcommand entry point. It parses args (--dir, --serve,
// --addr flags), optionally resolves the speech directory and prints a preview,
// and optionally starts the embedded UI HTTP server.
func Run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("compose", flag.ContinueOnError)
	fs.SetOutput(stderr)
	dir := fs.String("dir", "", "speech directory (overrides DEBATEOS_DIR)")
	serve := fs.Bool("serve", false, "serve the embedded Debate UI on localhost")
	addr := fs.String("addr", ":8080", "address to listen on when --serve is set")
	noListen := fs.Bool("no-listen", false, "parse flags and resolve but do not bind to addr (testing seam)")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	// --serve mode: start the embedded UI server.
	if *serve {
		if !*noListen {
			fmt.Fprintf(stdout, "Serving DebateOS UI at http://localhost%s\n", *addr)
			fmt.Fprintf(stdout, "Press Ctrl+C to stop.\n")
			if err := serveUI(*addr); err != nil {
				fmt.Fprintf(stderr, "compose --serve: %v\n", err)
				return 1
			}
		}
		// --no-listen: just validate flags, don't bind.
		return 0
	}

	// Resolution preview mode (original behavior — unchanged).
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
