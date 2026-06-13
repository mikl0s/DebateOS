// Package embeddedui provides the embedded SvelteKit UI build for offline serving.
//
// The web/ subdirectory is the go:embed target. It is populated by
// scripts/build-ui-dual.sh with a BASE_PATH= (empty) SvelteKit build so that
// `debateos compose --serve` can serve the UI at localhost root without a
// subpath — byte-identical to the GitHub Pages build except for the base path.
//
// cli/embed/web/index.html is a placeholder committed to the repo so the
// //go:embed directive always compiles. The build script overwrites it with the
// real SvelteKit output.
//
// T-05-18 mitigation: the embedded build is always the BASE_PATH= build (not
// the Pages BASE_PATH=/debateos build). See scripts/build-ui-dual.sh.
package embeddedui

import (
	"embed"
	"io/fs"
	"net/http"
)

// WebFS is the embedded filesystem containing the SvelteKit build output.
// The //go:embed all: prefix includes dotfiles such as .nojekyll.
//
//go:embed all:web
var WebFS embed.FS

// NewUIHandler returns an http.Handler that serves the embedded UI at the root.
// It uses fs.Sub to strip the "web/" prefix so files are served at "/".
//
// net/http auto-detects Content-Type for .wasm files as "application/wasm"
// (Go 1.17+), so no explicit override is required.
func NewUIHandler() http.Handler {
	sub, err := fs.Sub(WebFS, "web")
	if err != nil {
		// This cannot happen at runtime — the "web" directory is always present
		// in the embedded FS because index.html is committed.
		panic("embeddedui: fs.Sub(WebFS, \"web\") failed: " + err.Error())
	}
	return http.FileServer(http.FS(sub))
}

// ServeUI starts an HTTP server at addr that serves the embedded UI.
// It blocks until the server exits or an error occurs.
func ServeUI(addr string) error {
	return http.ListenAndServe(addr, NewUIHandler())
}
