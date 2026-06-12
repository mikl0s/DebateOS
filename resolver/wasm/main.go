//go:build js && wasm

// Package main is the WASM entrypoint for the DebateOS resolver.
// It registers a global JavaScript function (debateosResolve) that accepts
// a JSON string containing speech, opinions, and hardware fields, calls the
// native resolver.Resolve + resolver.CanonicalJSON, and returns the result as
// a JSON string to the JS caller.
//
// Build: GOOS=js GOARCH=wasm go build ./resolver/wasm/ -o debateos.wasm
// Load via wasm_exec.js from $(go env GOROOT)/lib/wasm/wasm_exec.js.
// Do NOT commit a copy of wasm_exec.js — always reference the runtime copy.
//
// T-01-15 mitigation: scripts reference $(go env GOROOT)/lib/wasm/go_js_wasm_exec
// rather than any committed copy.
//
// T-01-16 mitigation: malformed JS input returns a structured {"error": "..."} JSON
// response; the function never panics the WASM runtime.
package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"syscall/js"

	"go.yaml.in/yaml/v3"

	"github.com/mikl0s/debateos/resolver"
	"github.com/mikl0s/debateos/resolver/hardware"
	"github.com/mikl0s/debateos/resolver/resolve"
)

// resolveInput is the JSON-decoded payload accepted by debateosResolve.
// It mirrors the structure of a loaded composition: speech, opinions, hardware.
type resolveInput struct {
	Speech   *resolver.Speech       `json:"speech"   yaml:"speech"`
	Opinions []resolver.Opinion     `json:"opinions" yaml:"opinions"`
	Hardware hardware.HardwareProfile `json:"hardware" yaml:"hardware"`
}

// resolveOutput is the JSON-encoded response returned by debateosResolve.
// On success: {"result": "<canonical JSON string>"}
// On error:   {"error": "<error message>"}
type resolveOutput struct {
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

// debateosResolveFunc is the js.Func implementation.
// It accepts a single string argument (JSON-encoded resolveInput) and returns
// a string (JSON-encoded resolveOutput).
func debateosResolveFunc(_ js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return errorResponse("missing argument: expected a JSON string")
	}

	inputStr := args[0].String()
	if inputStr == "" {
		return errorResponse("empty input")
	}

	// Try JSON decode first (primary production path); fall back to YAML.
	// JSON: use DisallowUnknownFields so typos like "speeech" are caught early.
	// YAML: use KnownFields(true) for the same reason (IN-04).
	var input resolveInput
	jsonDec := json.NewDecoder(strings.NewReader(inputStr))
	jsonDec.DisallowUnknownFields()
	if err := jsonDec.Decode(&input); err != nil {
		// Try YAML as a fallback (handles YAML-formatted input from tests/CLI).
		yamlDec := yaml.NewDecoder(strings.NewReader(inputStr))
		yamlDec.KnownFields(true)
		if yamlErr := yamlDec.Decode(&input); yamlErr != nil {
			return errorResponse(fmt.Sprintf("parse error (json: %v; yaml: %v)", err, yamlErr))
		}
	}

	if input.Speech == nil {
		return errorResponse("speech field is required")
	}

	rs, resolveErr := resolve.Resolve(input.Speech, input.Opinions, input.Hardware)
	if rs == nil {
		// Nil ResolvedSpeech means an internal error (should not happen in practice).
		if resolveErr != nil {
			return errorResponse(fmt.Sprintf("resolve error (no result): %v", resolveErr))
		}
		return errorResponse("resolve returned nil result without error")
	}

	// Always produce canonical JSON — even on hard conflict (partial result with
	// explanations so the caller can display the conflict text).
	canonical, marshalErr := resolve.CanonicalJSON(rs)
	if marshalErr != nil {
		return errorResponse(fmt.Sprintf("canonical JSON error: %v", marshalErr))
	}

	out := resolveOutput{Result: string(canonical)}
	if resolveErr != nil {
		// Attach the error message alongside the partial result so JS callers can
		// distinguish clean resolves from hard conflicts.
		out.Error = resolveErr.Error()
	}

	outBytes, err := json.Marshal(out)
	if err != nil {
		return errorResponse(fmt.Sprintf("output marshal error: %v", err))
	}
	return string(outBytes)
}

// errorResponse encodes an error-only resolveOutput as a JSON string.
func errorResponse(msg string) string {
	b, _ := json.Marshal(resolveOutput{Error: msg})
	return string(b)
}

func init() {
	// Register the global resolve function in init() so it is available to both
	// the production main() runtime (loaded by wasm_exec.js) AND to the WASM
	// test runner (go test -exec go_js_wasm_exec), which calls init but not main.
	js.Global().Set("debateosResolve", js.FuncOf(debateosResolveFunc))
}

func main() {
	// Block the main goroutine indefinitely so the WASM runtime stays alive.
	// JS callers drive execution via callbacks; we never exit voluntarily.
	// The debateosResolve function is already registered in init().
	select {}
}
