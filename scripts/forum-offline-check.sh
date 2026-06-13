#!/usr/bin/env bash
# scripts/forum-offline-check.sh — Invariant-4 offline gate (FORM-05)
#
# Proves that compose→resolve works with the Forum process completely offline.
# This is the phase's hardest guarantee: the core DebateOS pipeline (compose →
# resolve-json) must succeed even when no Forum service is running.
#
# Steps:
#   1. Ensure no forum/forumctl process is running.
#   2. Run `debateos compose --dir examples/dual-foundation/` (resolves locally).
#   3. Run `resolve-json examples/dual-foundation` and assert the resolved JSON
#      contains Applied opinions (non-empty, parseable).
#   4. Print INVARIANT-4 PASS.
#
# On any assertion failure the script exits non-zero, causing CI to fail.
#
# T-05-19 mitigation: Forum is NOT in the critical path of compose→resolve.
# This script proves it by running the full pipeline with Forum offline.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EXAMPLE_DIR="${REPO_ROOT}/examples/dual-foundation"

echo "=== INVARIANT-4 GATE: Forum-Offline compose→resolve ==="
echo ""

# Step 1: Ensure no forum/forumctl process is running.
echo "--> Step 1: Ensuring no forum process is running..."
pkill -f forumctl 2>/dev/null || true
pkill -f "forum/cmd" 2>/dev/null || true
# Brief pause to let any dying processes exit cleanly.
sleep 0.2
echo "    OK — no forum processes running."
echo ""

# Step 2: Run compose on dual-foundation example.
echo "--> Step 2: Running debateos compose --dir ${EXAMPLE_DIR}..."
COMPOSE_OUT="$(go run "${REPO_ROOT}/cmd/debateos" compose --dir "${EXAMPLE_DIR}" 2>&1)"
echo "${COMPOSE_OUT}"
echo ""

# Assert compose returned Applied count.
if ! echo "${COMPOSE_OUT}" | grep -q "Applied:"; then
    echo "FAIL: compose output missing 'Applied:' — compose did not resolve successfully." >&2
    exit 1
fi

APPLIED_COUNT="$(echo "${COMPOSE_OUT}" | grep -oP 'Applied:\s*\K[0-9]+')"
if [ -z "${APPLIED_COUNT}" ] || [ "${APPLIED_COUNT}" -lt 1 ]; then
    echo "FAIL: compose reported 0 applied opinions — expected at least 1." >&2
    exit 1
fi
echo "--> Compose: Applied=${APPLIED_COUNT}  OK"
echo ""

# Step 3: Run resolve-json and assert non-empty parsed JSON with Applied count.
echo "--> Step 3: Running resolve-json on ${EXAMPLE_DIR}..."
RESOLVE_OUT="$(go run "${REPO_ROOT}/cmd/resolve-json" "${EXAMPLE_DIR}" 2>&1)"
# Separate JSON line (last line with {) from diagnostic output.
RESOLVE_JSON="$(echo "${RESOLVE_OUT}" | grep '^{' | tail -1)"

if [ -z "${RESOLVE_JSON}" ]; then
    echo "FAIL: resolve-json produced no JSON output." >&2
    echo "Full output: ${RESOLVE_OUT}" >&2
    exit 1
fi

# Validate JSON and assert applied count.
APPLIED_RESOLVE="$(python3 -c "
import json, sys
d = json.loads(sys.stdin.read())
applied = d.get('applied', [])
print(len(applied))
" <<< "${RESOLVE_JSON}")"

if [ -z "${APPLIED_RESOLVE}" ] || [ "${APPLIED_RESOLVE}" -lt 1 ]; then
    echo "FAIL: resolve-json returned 0 applied opinions — expected at least 1." >&2
    echo "JSON: ${RESOLVE_JSON}" >&2
    exit 1
fi

echo ""
echo "--> resolve-json: Applied=${APPLIED_RESOLVE}  OK"
echo ""

echo "=== INVARIANT-4 PASS: compose→resolve works with Forum offline ==="
echo "    Compose Applied=${APPLIED_COUNT}  resolve-json Applied=${APPLIED_RESOLVE}"
echo "    The Forum process is NOT required for the core pipeline (T-05-19 mitigated)."
