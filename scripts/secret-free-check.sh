#!/usr/bin/env bash
# scripts/secret-free-check.sh
#
# Secret-free artifact assertion gate (PRIV-01 / T-03-SECRETFREE).
#
# Builds the omarchy speech with --skip-iso, then asserts that none of the
# following secret-bearing filenames appear anywhere in the shared arch-profile/
# tree:
#
#   pane.yaml            — private pane (must never enter shared artifacts)
#   identity.age         — age X25519 private key (must never enter shared artifacts)
#   private-injection.tar — first-boot injection tar (local only, next to ISO)
#
# These files are produced or managed by `debateos pane` and `debateos build`
# and are designed to stay on the user's local machine.  If any of them appear
# inside the arch-profile/ tree it means a path containment bug exists that
# could expose secrets in the published ISO.
#
# Usage:
#   bash scripts/secret-free-check.sh [--speech-dir <path>] [--profile <name>]
#
# Defaults:
#   --speech-dir  examples/omarchy
#   --profile     vanilla-arch
#
# Environment:
#   DEBATEOS_BIN   path to the debateos binary (default: ./debateos or built from source)
#
# Exit codes:
#   0  Secret-free: none of the checked filenames found in the profile tree
#   1  Secrets found in profile tree (or script error) — investigate immediately

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# ─── Argument defaults ─────────────────────────────────────────────────────
SPEECH_DIR="${REPO_ROOT}/examples/omarchy"
PROFILE="vanilla-arch"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --speech-dir) SPEECH_DIR="$2"; shift 2 ;;
        --profile)    PROFILE="$2";    shift 2 ;;
        *) echo "unknown argument: $1" >&2; exit 1 ;;
    esac
done

# ─── Resolve debateos binary ───────────────────────────────────────────────
if [ -n "${DEBATEOS_BIN:-}" ]; then
    DEBATEOS="${DEBATEOS_BIN}"
elif command -v debateos &>/dev/null; then
    DEBATEOS="debateos"
elif [ -x "${REPO_ROOT}/debateos" ]; then
    DEBATEOS="${REPO_ROOT}/debateos"
else
    echo "=== Building debateos binary ==="
    (cd "${REPO_ROOT}" && go build -o debateos ./cmd/debateos)
    DEBATEOS="${REPO_ROOT}/debateos"
fi

echo "=== DebateOS Secret-Free Artifact Check (PRIV-01) ==="
echo "  debateos: ${DEBATEOS}"
echo "  speech:   ${SPEECH_DIR}"
echo "  profile:  ${PROFILE}"
echo ""

# ─── Temporary build directory ─────────────────────────────────────────────
WORK_DIR="$(mktemp -d /tmp/debateos-secretfree-XXXXXX)"
OUT_DIR="${WORK_DIR}/out"
PROFILE_DIR="${OUT_DIR}/arch-profile"

cleanup() { rm -rf "${WORK_DIR}"; }
trap cleanup EXIT

mkdir -p "${OUT_DIR}"

# ─── Build (profile emission only — no mkarchiso required) ─────────────────
echo "--- Running: debateos build --skip-iso ---"
(cd "${REPO_ROOT}" && "${DEBATEOS}" build \
    --dir  "${SPEECH_DIR}" \
    --profile "${PROFILE}" \
    --out  "${OUT_DIR}" \
    --skip-iso)
echo ""

if [ ! -d "${PROFILE_DIR}" ]; then
    echo "ERROR: arch-profile/ not found: ${PROFILE_DIR}" >&2
    exit 1
fi

echo "--- Checking arch-profile/ tree for secret filenames ---"
echo "  Profile dir: ${PROFILE_DIR}"
echo ""

# Secret filenames that must be absent from the shared profile tree.
SECRET_PATTERNS=(
    "pane.yaml"
    "identity.age"
    "private-injection.tar"
)

FOUND=0
for PATTERN in "${SECRET_PATTERNS[@]}"; do
    # Use find (not grep -r) to check for exact filename matches.
    MATCHES="$(find "${PROFILE_DIR}" -name "${PATTERN}" 2>/dev/null || true)"
    if [ -n "${MATCHES}" ]; then
        echo "FAIL: '${PATTERN}' found in arch-profile/ tree:" >&2
        echo "${MATCHES}" | while read -r M; do echo "  ${M}" >&2; done
        FOUND=$((FOUND + 1))
    else
        echo "  OK: '${PATTERN}' absent from profile tree"
    fi
done

# Also grep for the string "pane.yaml" appearing inside any file content
# (could indicate a config file accidentally referencing the private pane path).
GREP_MATCHES="$(grep -rl "pane.yaml" "${PROFILE_DIR}" 2>/dev/null || true)"
if [ -n "${GREP_MATCHES}" ]; then
    echo "" >&2
    echo "WARN: 'pane.yaml' string found in profile tree file contents:" >&2
    echo "${GREP_MATCHES}" | while read -r M; do echo "  ${M}" >&2; done
    echo "  Review these files to confirm they do not reference private pane paths." >&2
    # This is a warning, not a hard failure (the string may be in a comment or doc).
    # Hard failure is reserved for the filename check above.
fi

echo ""

if [ "${FOUND}" -gt 0 ]; then
    echo "=== SECRET-FREE CHECK FAIL: ${FOUND} secret file(s) found in profile tree ===" >&2
    echo "" >&2
    echo "This is a PRIV-01 violation.  The following must be ensured:" >&2
    echo "  - pane.yaml must only exist in ~/.config/debateos/ (never in arch-profile/)" >&2
    echo "  - identity.age must only exist in ~/.config/debateos/ (never in arch-profile/)" >&2
    echo "  - private-injection.tar must be written next to the ISO (outDir), not in profileDir" >&2
    exit 1
fi

echo "=== SECRET-FREE CHECK OK: no secret files in profile tree ==="
exit 0
