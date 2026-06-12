#!/usr/bin/env bash
# scripts/check-coverage.sh
#
# Coverage gate for the resolver packages.
# Fails (exit non-zero) if total coverage is below 90%.
# Passes (exit 0) at or above 90%.
#
# Usage: bash scripts/check-coverage.sh
#        (run from the module root)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
THRESHOLD=90

echo "=== DebateOS Coverage Gate (threshold: ${THRESHOLD}%) ==="
echo ""

PROFILE="$(mktemp /tmp/debateos-cover-XXXXXX.out)"
trap 'rm -f "$PROFILE"' EXIT

# Run tests with coverage over all resolver packages.
echo "--- Running go test -coverprofile over ./resolver/... ---"
(cd "$REPO_ROOT" && go test -coverprofile="$PROFILE" -covermode=atomic ./resolver/... -count=1 2>&1)
echo ""

# Parse total coverage from the profile.
# `go tool cover` output: "total: (statements)   91.3%"
TOTAL_LINE="$(cd "$REPO_ROOT" && go tool cover -func="$PROFILE" 2>/dev/null | grep '^total:' || true)"
echo "Coverage report line: $TOTAL_LINE"

# Extract the percentage value (strip the % sign).
COVERAGE="$(echo "$TOTAL_LINE" | awk '{print $3}' | tr -d '%')"

if [ -z "$COVERAGE" ]; then
    echo "ERROR: could not parse total coverage from profile"
    exit 1
fi

echo ""
echo "Total coverage: ${COVERAGE}%"
echo "Threshold:      ${THRESHOLD}%"

# Compare using awk (bc may not be installed).
PASS="$(awk -v cov="$COVERAGE" -v thr="$THRESHOLD" 'BEGIN { print (cov+0 >= thr+0) ? "1" : "0" }')"

if [ "$PASS" = "1" ]; then
    echo ""
    echo "=== COVERAGE OK: ${COVERAGE}% >= ${THRESHOLD}% ==="
    exit 0
else
    echo ""
    echo "=== COVERAGE FAIL: ${COVERAGE}% < ${THRESHOLD}% ==="
    echo "Add tests to bring coverage above ${THRESHOLD}%."
    exit 1
fi
