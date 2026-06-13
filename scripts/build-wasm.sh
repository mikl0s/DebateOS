#!/usr/bin/env bash
# scripts/build-wasm.sh — Build the DebateOS Go-WASM resolver and copy wasm_exec.js
#
# Outputs (both gitignored per web/.gitignore — T-05-03):
#   web/static/debateos.wasm   (~4.3 MB)
#   web/static/wasm_exec.js    (copied from GOROOT at build time — version-matched)
#
# Usage:
#   bash scripts/build-wasm.sh           # from repo root
#   BASE_PATH=/debateos npm run build    # full Pages build (call this script first)
#
# T-05-03 mitigation: wasm_exec.js is always sourced from $(go env GOROOT)/lib/wasm/
# so the JS runtime glue matches the Go toolchain that compiled the WASM binary.
# Never commit wasm_exec.js — it changes between Go versions.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${REPO_ROOT}/web/static"

echo "==> Building debateos.wasm (GOOS=js GOARCH=wasm)..."
GOOS=js GOARCH=wasm go build \
  -o "${OUT_DIR}/debateos.wasm" \
  ./resolver/wasm/

WASM_SIZE=$(du -sh "${OUT_DIR}/debateos.wasm" | cut -f1)
echo "    Size: ${WASM_SIZE}  →  ${OUT_DIR}/debateos.wasm"

echo "==> Copying wasm_exec.js from GOROOT..."
GOROOT_WASM="$(go env GOROOT)/lib/wasm/wasm_exec.js"
if [ ! -f "${GOROOT_WASM}" ]; then
  echo "ERROR: wasm_exec.js not found at ${GOROOT_WASM}" >&2
  exit 1
fi
cp "${GOROOT_WASM}" "${OUT_DIR}/wasm_exec.js"
echo "    Source: ${GOROOT_WASM}"
echo "    Dest:   ${OUT_DIR}/wasm_exec.js"

echo "==> Done. WASM artifacts ready in ${OUT_DIR}/"
echo "    (Both files are gitignored — do not commit them.)"
