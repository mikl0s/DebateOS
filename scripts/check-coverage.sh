#!/usr/bin/env bash
# scripts/check-coverage.sh
#
# Four-gate coverage check for DebateOS (D19 / Phase 5 extension):
#   Gate 1: ./resolver/... must be >= 90%   (original gate, unchanged)
#   Gate 2: ./cli/...      must be >= 85%   (added Phase 3 Plan 04)
#   Gate 3: ./registry/... must be >= 85%   (added Phase 5 Plan 06)
#   Gate 4: ./forum/...    must be >= 85%   (added Phase 5 Plan 06)
#
# Fails (exit non-zero) if ANY gate is below threshold.
# Passes (exit 0) when all gates are at or above their thresholds.
#
# Usage: bash scripts/check-coverage.sh
#        (run from the module root)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

RESOLVER_THRESHOLD=90
CLI_THRESHOLD=85
REGISTRY_THRESHOLD=85
FORUM_THRESHOLD=85

echo "=== DebateOS Coverage Gate ==="
echo "  resolver/  threshold: ${RESOLVER_THRESHOLD}%"
echo "  cli/       threshold: ${CLI_THRESHOLD}%"
echo "  registry/  threshold: ${REGISTRY_THRESHOLD}%"
echo "  forum/     threshold: ${FORUM_THRESHOLD}%"
echo ""

# Helper: run coverage for a package path and check against a threshold.
# Returns 0 if coverage >= threshold, 1 otherwise.
# Prints the result line to stdout.
# Usage: check_package_coverage "label" "threshold" pkg1 [pkg2 ...]
check_package_coverage() {
    local LABEL="$1"
    local THRESHOLD="$2"
    shift 2
    local PKGS=("$@")

    local PROFILE
    PROFILE="$(mktemp /tmp/debateos-cover-XXXXXX.out)"
    trap "rm -f '${PROFILE}'" RETURN

    echo "--- Running go test -coverprofile over ${PKGS[*]} ---"
    (cd "${REPO_ROOT}" && go test -coverprofile="${PROFILE}" -covermode=atomic "${PKGS[@]}" -count=1 2>&1)
    echo ""

    local TOTAL_LINE
    TOTAL_LINE="$(cd "${REPO_ROOT}" && go tool cover -func="${PROFILE}" 2>/dev/null | grep '^total:' || true)"
    echo "Coverage report line (${LABEL}): ${TOTAL_LINE}"

    local COVERAGE
    COVERAGE="$(echo "${TOTAL_LINE}" | awk '{print $3}' | tr -d '%')"

    if [ -z "${COVERAGE}" ]; then
        echo "ERROR: could not parse total coverage for ${PKGS[*]}" >&2
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

# Track overall pass/fail across all gates.
OVERALL=0

echo ""
check_package_coverage "resolver/" "${RESOLVER_THRESHOLD}" "./resolver/..." || OVERALL=1

echo ""
check_package_coverage "cli/" "${CLI_THRESHOLD}" "./cli/..." || OVERALL=1

echo ""
check_package_coverage "registry/" "${REGISTRY_THRESHOLD}" "./registry/..." || OVERALL=1

# Forum coverage: exclude the binary entrypoint (forum/cmd/forumctl — main package,
# untestable without running the whole service).
# The 85% gate applies to the testable core packages listed below.
# forum/cmd/forumctl is excluded because its main(), runServe(), and runReindex()
# functions require a running service environment (not unit-testable).
echo ""
check_package_coverage "forum/ (testable core)" "${FORUM_THRESHOLD}" \
  "./forum" \
  "./forum/api" \
  "./forum/migrations" \
  "./forum/store" \
  "./forum/store/generated" || OVERALL=1

echo ""
if [ "${OVERALL}" = "0" ]; then
    echo "=== ALL COVERAGE GATES PASSED ==="
    exit 0
else
    echo "=== ONE OR MORE COVERAGE GATES FAILED ===" >&2
    exit 1
fi
