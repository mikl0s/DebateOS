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
	"bytes"
	"embed"
	"io/fs"
	"net/http"
)

// WebFS is the embedded filesystem containing the SvelteKit build output.
// The //go:embed all: prefix includes dotfiles such as .nojekyll.
//
//go:embed all:web
var WebFS embed.FS

// bufferingRecorder buffers the file server's response so that if it returns
// a 404 we can discard it and serve 404.html instead. This avoids sending a
// partial 404 body before switching to the SPA shell.
type bufferingRecorder struct {
	header http.Header
	buf    bytes.Buffer
	code   int
}

func newBufferingRecorder() *bufferingRecorder {
	return &bufferingRecorder{
		header: make(http.Header),
		code:   http.StatusOK,
	}
}

func (br *bufferingRecorder) Header() http.Header {
	return br.header
}

func (br *bufferingRecorder) WriteHeader(code int) {
	br.code = code
}

func (br *bufferingRecorder) Write(b []byte) (int, error) {
	return br.buf.Write(b)
}

// flush writes the buffered response to the real ResponseWriter.
func (br *bufferingRecorder) flush(w http.ResponseWriter) {
	for k, vals := range br.header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(br.code)
	_, _ = w.Write(br.buf.Bytes())
}

// NewUIHandler returns an http.Handler that serves the embedded UI at the root.
// It uses fs.Sub to strip the "web/" prefix so files are served at "/".
//
// SPA fallback (WR-05): SvelteKit's adapter-static generates 404.html as the
// SPA shell for client-side routes (/debate/, /export/, /browse/).
// net/http.FileServer does not know about 404.html; it returns a plain HTTP 404
// for any unrecognised path. This wrapper intercepts 404 responses from the
// file server and re-serves 404.html so deep-linked SPA routes work offline.
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
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Buffer the file server response so we can intercept 404 before any
		// bytes are flushed to the real connection.
		br := newBufferingRecorder()
		fileServer.ServeHTTP(br, r)

		if br.code == http.StatusNotFound {
			// Serve 404.html (the SvelteKit SPA shell) for unrecognised routes.
			// This lets the client-side router handle /debate/, /export/, etc.
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/404.html"
			http.ServeFileFS(w, r2, sub, "404.html")
			return
		}

		// Not a 404 — flush the buffered response to the real writer.
		br.flush(w)
	})
}

// ServeUI starts an HTTP server at addr that serves the embedded UI.
// It blocks until the server exits or an error occurs.
func ServeUI(addr string) error {
	return http.ListenAndServe(addr, NewUIHandler())
}
