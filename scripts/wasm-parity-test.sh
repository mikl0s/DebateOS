#!/usr/bin/env bash
# scripts/wasm-parity-test.sh
#
# WASM parity test: asserts that native and WASM canonical-JSON outputs are
# byte-identical to the committed golden files in
# resolver/resolve/testdata/golden/.
#
# Exit 0 = PARITY OK
# Exit non-zero = parity failure or guard condition not met
#
# T-01-14 mitigation: parity is automated (not inspection).
# T-01-15 mitigation: uses $(go env GOROOT)/lib/wasm/go_js_wasm_exec (no
#                     committed wasm_exec.js copy in this repo).
#
# Usage: bash scripts/wasm-parity-test.sh
#        (run from the module root)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
GOLDEN_DIR="$REPO_ROOT/resolver/resolve/testdata/golden"
WASM_EXEC="$(go env GOROOT)/lib/wasm/go_js_wasm_exec"

echo "=== DebateOS WASM Parity Test ==="
echo "REPO_ROOT: $REPO_ROOT"
echo "GOLDEN_DIR: $GOLDEN_DIR"
echo "WASM_EXEC: $WASM_EXEC"
echo ""

# ── GUARD: golden dir must exist and hold >= 4 files ──────────────────────
if [ ! -d "$GOLDEN_DIR" ]; then
    echo "ERROR: golden dir does not exist: $GOLDEN_DIR"
    echo "Run: GOLDEN_UPDATE=1 go test ./resolver/resolve/ -run TestCanonicalGolden"
    exit 1
fi

GOLDEN_COUNT="$(ls "$GOLDEN_DIR" | wc -l)"
if [ "$GOLDEN_COUNT" -lt 4 ]; then
    echo "ERROR: golden dir holds only $GOLDEN_COUNT files (need >= 4): $GOLDEN_DIR"
    echo "Run: GOLDEN_UPDATE=1 go test ./resolver/resolve/ -run TestCanonicalGolden"
    exit 1
fi
echo "GUARD OK: $GOLDEN_COUNT golden files found in $GOLDEN_DIR"
echo ""

# ── NATIVE run ────────────────────────────────────────────────────────────
NATIVE_TMP="$(mktemp -d)"
trap 'rm -rf "$NATIVE_TMP" "${WASM_TMP:-}"' EXIT

echo "--- Running NATIVE TestCanonicalGolden (writes to $NATIVE_TMP) ---"
(cd "$REPO_ROOT" && GOLDEN_UPDATE=1 GOLDEN_DIR="$NATIVE_TMP" \
    go test ./resolver/resolve/ -run TestCanonicalGolden -count=1 2>&1)

NATIVE_COUNT="$(ls "$NATIVE_TMP" | wc -l)"
if [ "$NATIVE_COUNT" -eq 0 ]; then
    echo "ERROR: native test produced no golden files in $NATIVE_TMP"
    exit 1
fi
echo "Native: $NATIVE_COUNT files written"
echo ""

# diff native output against committed goldens
echo "--- Diffing native output against committed goldens ---"
if ! diff -r "$GOLDEN_DIR" "$NATIVE_TMP"; then
    echo "PARITY FAIL: native output differs from committed goldens"
    exit 1
fi
echo "Native output matches committed goldens."
echo ""

# ── WASM run ──────────────────────────────────────────────────────────────
WASM_TMP="$(mktemp -d)"

if [ ! -f "$WASM_EXEC" ]; then
    echo "ERROR: go_js_wasm_exec not found at $WASM_EXEC"
    echo "Ensure Go $( go version) is installed; expected at \$(go env GOROOT)/lib/wasm/go_js_wasm_exec"
    exit 1
fi

echo "--- Running WASM TestCanonicalGolden (writes to $WASM_TMP) ---"
(cd "$REPO_ROOT" && GOLDEN_UPDATE=1 GOLDEN_DIR="$WASM_TMP" \
    GOOS=js GOARCH=wasm go test \
    -exec="$WASM_EXEC" \
    ./resolver/resolve/ \
    -run TestCanonicalGolden \
    -count=1 2>&1)

WASM_COUNT="$(ls "$WASM_TMP" | wc -l)"
if [ "$WASM_COUNT" -eq 0 ]; then
    echo "ERROR: WASM test produced no golden files in $WASM_TMP"
    exit 1
fi
echo "WASM: $WASM_COUNT files written"
echo ""

# diff WASM output against committed goldens
echo "--- Diffing WASM output against committed goldens ---"
if ! diff -r "$GOLDEN_DIR" "$WASM_TMP"; then
    echo "PARITY FAIL: WASM output differs from committed goldens"
    exit 1
fi
echo "WASM output matches committed goldens."
echo ""

echo "=== PARITY OK ==="
