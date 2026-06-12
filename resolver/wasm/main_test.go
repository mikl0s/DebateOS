//go:build js && wasm

// Package wasm smoke-tests the WASM entrypoint exported by main.go.
// This file is compiled only for GOOS=js GOARCH=wasm (go_js_wasm_exec target).
//
// TestWasmEntryPointSmoke verifies that the exported debateosResolve js.Func
// produces a non-empty string when called with a minimal valid JSON payload,
// and that the result parses as valid JSON.
package main

import (
	"encoding/json"
	"syscall/js"
	"testing"
)

// TestWasmEntryPointSmoke calls the registered debateosResolve global with a
// minimal valid speech + opinions payload and asserts the response is non-empty
// valid JSON. This proves the WASM glue does not alter the output format.
func TestWasmEntryPointSmoke(t *testing.T) {
	// Minimal speech: one required opinion, no conflicts, no hardware.
	input := `{
		"speech": {
			"schema": 1,
			"id": "wasm-smoke",
			"foundation": "arch",
			"points": [],
			"opinions": [{"id": "OM-006"}]
		},
		"opinions": [
			{
				"schema": 1,
				"id": "OM-006",
				"name": "Wayland compositor stack",
				"category": "package-install",
				"status": "required",
				"packages": ["hyprland"]
			}
		],
		"hardware": {}
	}`

	fn := js.Global().Get("debateosResolve")
	if fn.IsUndefined() || fn.IsNull() {
		t.Fatal("debateosResolve is not registered as a global JS function")
	}

	result := fn.Invoke(js.ValueOf(input))
	if result.IsUndefined() || result.IsNull() {
		t.Fatal("debateosResolve returned undefined/null")
	}

	resultStr := result.String()
	if resultStr == "" {
		t.Fatal("debateosResolve returned empty string")
	}

	// Must be valid JSON.
	var parsed interface{}
	if err := json.Unmarshal([]byte(resultStr), &parsed); err != nil {
		t.Fatalf("debateosResolve output is not valid JSON: %v\noutput: %s", err, resultStr)
	}

	// Must not be an error-only response with no output.
	m, ok := parsed.(map[string]interface{})
	if !ok {
		t.Fatalf("debateosResolve output top-level is not a JSON object: %s", resultStr)
	}

	// Either "result" or "error" key must be present.
	_, hasResult := m["result"]
	_, hasError := m["error"]
	if !hasResult && !hasError {
		t.Fatalf("debateosResolve output has neither 'result' nor 'error' key: %s", resultStr)
	}

	// For a clean input, result must be present and non-empty.
	if !hasResult {
		t.Fatalf("debateosResolve returned error for clean input: %s", resultStr)
	}
	if m["result"] == nil || m["result"] == "" {
		t.Fatalf("debateosResolve 'result' is empty for clean input: %s", resultStr)
	}
}
