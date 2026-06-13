#!/usr/bin/env bash
# generate-index.sh — regenerate the static registry index (REG-01).
#
# Wraps cmd/gen-index: validates every point/opinion via resolver/parse and
# emits registry/index.json + registry/index.html. Deterministic — re-running
# on unchanged source produces byte-identical output, so the CI commit step
# (.github/workflows/registry-index.yml) is a no-op when nothing changed.
#
# Source of record is Git (the point/opinion YAML); index.json is a derived
# cache. Usage: scripts/generate-index.sh [source-dir] [out-dir]
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

SOURCE_DIR="${1:-examples/omarchy}"
OUT_DIR="${2:-registry}"

# Deterministic generatedAt: the source dir's last git commit date (ISO-8601,
# UTC). Falls back to empty if not in a git repo or the path is untracked —
# gen-index renders an empty string rather than a churning wall-clock value.
GENERATED_AT="$(git log -1 --format=%cI -- "$SOURCE_DIR" 2>/dev/null || true)"
export GENERATED_AT

echo "generate-index: source=$SOURCE_DIR out=$OUT_DIR generated_at=${GENERATED_AT:-<none>}"
go run ./cmd/gen-index "$SOURCE_DIR" "$OUT_DIR"
