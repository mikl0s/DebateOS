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

// TestWasmContentType verifies that .wasm files are served with Content-Type application/wasm.
func TestWasmContentType(t *testing.T) {
	handler := embeddedui.NewUIHandler()

	req := httptest.NewRequest(http.MethodGet, "/debateos.wasm", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	// The placeholder embed doesn't have a real .wasm file, but we can verify
	// the Content-Type detection works when such a file is present.
	// If the file doesn't exist (placeholder), we expect a 404.
	// When a real WASM is present, we test the Content-Type.
	ct := resp.Header.Get("Content-Type")
	if resp.StatusCode == http.StatusOK {
		if !strings.Contains(ct, "application/wasm") {
			t.Errorf("expected Content-Type application/wasm for .wasm file, got %q", ct)
		}
	} else if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 200 or 404 for .wasm file, got %d", resp.StatusCode)
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
