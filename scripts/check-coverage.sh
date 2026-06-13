#!/usr/bin/env bash
# scripts/check-coverage.sh
#
# Two-gate coverage check for DebateOS (D19 / Phase 3 extension):
#   Gate 1: ./resolver/... must be >= 90%   (original gate, unchanged)
#   Gate 2: ./cli/...      must be >= 85%   (added Phase 3 Plan 04)
#
# Fails (exit non-zero) if EITHER gate is below threshold.
# Passes (exit 0) when both gates are at or above their thresholds.
#
# Usage: bash scripts/check-coverage.sh
#        (run from the module root)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

RESOLVER_THRESHOLD=90
CLI_THRESHOLD=85

echo "=== DebateOS Coverage Gate ==="
echo "  resolver/ threshold: ${RESOLVER_THRESHOLD}%"
echo "  cli/      threshold: ${CLI_THRESHOLD}%"
echo ""

# Helper: run coverage for a package path and check against a threshold.
# Returns 0 if coverage >= threshold, 1 otherwise.
# Prints the result line to stdout.
check_package_coverage() {
    local LABEL="$1"
    local PKG="$2"
    local THRESHOLD="$3"

    local PROFILE
    PROFILE="$(mktemp /tmp/debateos-cover-XXXXXX.out)"
    trap "rm -f '${PROFILE}'" RETURN

    echo "--- Running go test -coverprofile over ${PKG} ---"
    (cd "${REPO_ROOT}" && go test -coverprofile="${PROFILE}" -covermode=atomic "${PKG}" -count=1 2>&1)
    echo ""

    local TOTAL_LINE
    TOTAL_LINE="$(cd "${REPO_ROOT}" && go tool cover -func="${PROFILE}" 2>/dev/null | grep '^total:' || true)"
    echo "Coverage report line (${LABEL}): ${TOTAL_LINE}"

    local COVERAGE
    COVERAGE="$(echo "${TOTAL_LINE}" | awk '{print $3}' | tr -d '%')"

    if [ -z "${COVERAGE}" ]; then
        echo "ERROR: could not parse total coverage for ${PKG}" >&2
        return 1
    fi

    echo ""
    echo "Total ${LABEL} coverage: ${COVERAGE}%"
    echo "Threshold:               ${THRESHOLD}%"

    local PASS
    PASS="$(awk -v cov="${COVERAGE}" -v thr="${THRESHOLD}" 'BEGIN { print (cov+0 >= thr+0) ? "1" : "0" }')"

    if [ "${PASS}" = "1" ]; then
        echo "=== ${LABEL} COVERAGE OK: ${COVERAGE}% >= ${THRESHOLD}% ==="
        return 0
    else
        echo "=== ${LABEL} COVERAGE FAIL: ${COVERAGE}% < ${THRESHOLD}% ===" >&2
        echo "Add tests to bring ${LABEL} coverage above ${THRESHOLD}%." >&2
        return 1
    fi
}

# Track overall pass/fail across both gates.
OVERALL=0

echo ""
check_package_coverage "resolver/" "./resolver/..." "${RESOLVER_THRESHOLD}" || OVERALL=1

echo ""
check_package_coverage "cli/" "./cli/..." "${CLI_THRESHOLD}" || OVERALL=1

echo ""
if [ "${OVERALL}" = "0" ]; then
    echo "=== ALL COVERAGE GATES PASSED ==="
    exit 0
else
    echo "=== ONE OR MORE COVERAGE GATES FAILED ===" >&2
    exit 1
fi
