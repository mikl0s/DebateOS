// serve.go — embedded UI serving for `debateos compose --serve`.
//
// ServeUI is a thin wrapper over embeddedui.ServeUI that is called by
// compose.Run when the --serve flag is set. It blocks on ListenAndServe.
//
// The embedded build is the BASE_PATH= (root) SvelteKit build produced by
// scripts/build-ui-dual.sh. It serves the Debate UI at localhost root so the
// resolver WASM loads from "/" — not from "/debateos/" (the Pages subpath).
package compose

import (
	embeddedui "github.com/mikl0s/debateos/cli/embed"
)

// serveUI starts the embedded UI server at addr and blocks until it exits.
// It delegates to embeddedui.ServeUI which uses fs.Sub + http.FileServer.
//
// Exposed as a package-level variable so tests can inject a no-op (noListen
// seam) without spinning up a real server.
var serveUI = embeddedui.ServeUI
