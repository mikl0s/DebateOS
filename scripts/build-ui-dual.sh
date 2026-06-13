#!/usr/bin/env bash
# scripts/build-ui-dual.sh — Dual-delivery SvelteKit build
#
# Builds the Debate UI twice from the same SvelteKit codebase:
#
#   1. BASE_PATH=/debateos build → dist/pages/   (GitHub Pages deployment artifact)
#   2. BASE_PATH=          build → cli/embed/web/ (go:embed target for `debateos compose --serve`)
#
# The embedded build (BASE_PATH=) is the one committed under cli/embed/web/.
# It is served at localhost root so assets are at "/" not "/debateos/".
#
# T-05-18 mitigation: the wrong base-path build embedded in the binary would
# cause 404s for all assets. This script ensures the BASE_PATH= build goes to
# the embed target and the BASE_PATH=/debateos build goes to the Pages target.
#
# WASM artifacts (debateos.wasm, wasm_exec.js) are built first so the SvelteKit
# build can find them in web/static/. They are copied into cli/embed/web/ as
# committed embed assets — the WASM binary MUST be present in the embedded FS
# for the offline serve to function.
#
# Note: web/static/debateos.wasm and web/static/wasm_exec.js are gitignored
# (T-05-03). The embedded copies in cli/embed/web/ ARE committed because they
# are required for the binary to serve the resolver offline.
#
# Usage:
#   bash scripts/build-ui-dual.sh         # from repo root
#   SKIP_PAGES=1 bash scripts/build-ui-dual.sh  # skip Pages build (embed only)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

PAGES_OUT="${REPO_ROOT}/dist/pages"
EMBED_OUT="${REPO_ROOT}/cli/embed/web"
WEB_DIR="${REPO_ROOT}/web"

echo "=== DebateOS Dual-Delivery UI Build ==="
echo "  Pages target:  ${PAGES_OUT}"
echo "  Embed target:  ${EMBED_OUT}"
echo ""

# Step 1: Build the WASM resolver + copy wasm_exec.js
echo "--> Step 1: Building WASM resolver..."
bash "${REPO_ROOT}/scripts/build-wasm.sh"
echo ""

# Step 2: Install npm dependencies if node_modules is missing.
if [ ! -d "${WEB_DIR}/node_modules" ]; then
    echo "--> Step 2: Installing npm dependencies..."
    npm --prefix "${WEB_DIR}" install
else
    echo "--> Step 2: npm dependencies already installed."
fi
echo ""

# Step 3: Pages build (BASE_PATH=/debateos)
if [ "${SKIP_PAGES:-}" != "1" ]; then
    echo "--> Step 3: Building Pages artifact (BASE_PATH=/debateos)..."
    rm -rf "${WEB_DIR}/build"
    BASE_PATH=/debateos npm --prefix "${WEB_DIR}" run build
    mkdir -p "${PAGES_OUT}"
    rm -rf "${PAGES_OUT:?}"/*
    cp -r "${WEB_DIR}/build/." "${PAGES_OUT}/"
    echo "    Pages build → ${PAGES_OUT}"
    echo ""
else
    echo "--> Step 3: Skipping Pages build (SKIP_PAGES=1)."
    echo ""
fi

# Step 4: Embed build (BASE_PATH= empty for localhost root)
echo "--> Step 4: Building embed artifact (BASE_PATH= for localhost root)..."
rm -rf "${WEB_DIR}/build"
BASE_PATH= npm --prefix "${WEB_DIR}" run build

# Sync build output into cli/embed/web (the go:embed target).
# rsync is preferred for efficiency; fall back to cp if unavailable.
echo "--> Syncing build output to cli/embed/web..."
rm -rf "${EMBED_OUT:?}"/*
cp -r "${WEB_DIR}/build/." "${EMBED_OUT}/"

echo ""
echo "--> Embed build → ${EMBED_OUT}"
echo "    Files:"
find "${EMBED_OUT}" -type f | head -20

echo ""
echo "=== Dual-Delivery Build Complete ==="
echo "  Pages artifact:  ${PAGES_OUT}  (deploy with: push to gh-pages branch)"
echo "  Embed artifact:  ${EMBED_OUT}  (committed; served by \`debateos compose --serve\`)"
echo ""
echo "Commit cli/embed/web/ to include the embedded UI in the binary:"
echo "  git add cli/embed/web/"
echo "  git commit -m 'chore: update embedded UI build'"
