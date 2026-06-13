package embeddedui_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	embeddedui "github.com/mikl0s/debateos/cli/embed"
)

// TestServeUI verifies that NewUIHandler serves the embedded index.html at root.
func TestServeUI(t *testing.T) {
	handler := embeddedui.NewUIHandler()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET / expected 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}

	if len(body) == 0 {
		t.Fatal("expected non-empty body for /")
	}

	// The placeholder (or real build) must contain html marker.
	if !strings.Contains(string(body), "<html") && !strings.Contains(string(body), "<!doctype") && !strings.Contains(string(body), "<!DOCTYPE") {
		t.Errorf("expected HTML body at root, got: %q", string(body)[:min(len(body), 200)])
	}
}

// TestWasmContentType verifies that a real embedded .wasm file is served with
// Content-Type application/wasm. With the SPA fallback active, a missing .wasm
// file returns the SPA shell (200 + text/html) rather than a raw 404.
// This test checks the Content-Type only when the embedded FS actually contains
// debateos.wasm; otherwise it confirms the SPA fallback kicked in.
func TestWasmContentType(t *testing.T) {
	// Check if debateos.wasm is actually present in the embedded FS.
	_, wasmErr := embeddedui.WebFS.Open("web/debateos.wasm")
	wasmPresent := wasmErr == nil

	handler := embeddedui.NewUIHandler()

	req := httptest.NewRequest(http.MethodGet, "/debateos.wasm", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if wasmPresent {
		// Real WASM file present: expect 200 with correct Content-Type.
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 for real .wasm file, got %d", resp.StatusCode)
		}
		if !strings.Contains(ct, "application/wasm") {
			t.Errorf("expected Content-Type application/wasm for .wasm file, got %q", ct)
		}
	} else {
		// No WASM in placeholder embed: SPA fallback returns the SPA shell (200 + HTML).
		// This is correct behaviour — the fallback is working as intended.
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected SPA fallback 200 for missing .wasm, got %d", resp.StatusCode)
		}
		t.Logf("debateos.wasm not embedded (placeholder build) — SPA fallback active, Content-Type: %q", ct)
	}
}

// TestWebFSNotEmpty verifies that the embedded FS contains at least one file.
func TestWebFSNotEmpty(t *testing.T) {
	// We verify the embed.FS is accessible.
	f, err := embeddedui.WebFS.Open("web/index.html")
	if err != nil {
		t.Fatalf("expected web/index.html in embedded FS: %v", err)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("reading embedded index.html: %v", err)
	}
	if len(content) == 0 {
		t.Error("embedded index.html is empty")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestServeUIInvalidAddr verifies that ServeUI returns an error for an invalid address.
// Port 99999 exceeds the valid range (0-65535) so net.Listen fails immediately,
// allowing this test to exercise the ServeUI code path without blocking.
func TestServeUIInvalidAddr(t *testing.T) {
	err := embeddedui.ServeUI(":99999")
	if err == nil {
		t.Error("expected error for invalid port 99999, got nil")
	}
}

// TestSPAFallback verifies that deep-linked SPA routes return 200 with the
// 404.html SPA shell content, not a raw HTTP 404 (WR-05).
// SvelteKit adapter-static generates 404.html as the client-side router entry
// point. Without the SPA fallback, `debateos compose --serve` would return a
// Go file-server 404 for /debate/, /export/, etc.
func TestSPAFallback(t *testing.T) {
	handler := embeddedui.NewUIHandler()

	// These paths have no pre-rendered static file in the embedded FS —
	// only the SPA shell (404.html) can handle them.
	spaPaths := []string{
		"/debate/",
		"/export/",
		"/browse/",
		"/nonexistent-deep-link",
	}

	for _, path := range spaPaths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			resp := rec.Result()
			defer resp.Body.Close()

			// Should NOT be a raw 404 from the Go file server.
			// The SPA shell (404.html) returns 200.
			if resp.StatusCode == http.StatusNotFound {
				t.Errorf("GET %s returned 404 — SPA fallback not working; deep links will break offline", path)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("reading body for %s: %v", path, err)
			}

			// The SPA shell must be HTML.
			if len(body) > 0 {
				bodyStr := string(body)
				if !strings.Contains(bodyStr, "<html") && !strings.Contains(bodyStr, "<!doctype") && !strings.Contains(bodyStr, "<!DOCTYPE") {
					t.Errorf("GET %s: expected HTML SPA shell, got content starting with: %q", path, bodyStr[:min(len(bodyStr), 100)])
				}
			}
		})
	}
}
