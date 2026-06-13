#!/usr/bin/env bash
# scripts/determinism-test.sh
#
# Determinism gate for DebateOS speech builds (BLD-03 / T-03-DET).
#
# Runs the full resolve + translate pipeline TWICE into independent clean
# directories, then produces deterministic tarballs from each and compares
# sha256 checksums.  Both checksums MUST be identical — otherwise there is
# a non-determinism bug in the build pipeline.
#
# What is tested:
#   debateos build --skip-iso  →  resolved.json + arch-profile/ (two runs)
#   SOURCE_DATE_EPOCH derived from sha256(resolved.json) — same algorithm as
#   manifest.py derive_source_date_epoch / cli/build.DeriveEpoch
#   GNU tar (--sort=name --mtime=@EPOCH --owner=0 --group=0 --numeric-owner
#            --pax-option=...delete=atime,delete=ctime) + gzip -n
#
# What is NOT tested (deferred — documented host limitation):
#   Full mkarchiso ISO build — requires a non-Proxmox host with devtmpfs access.
#   Full-ISO determinism uses the same script + flags on a capable host.
#
# Usage:
#   bash scripts/determinism-test.sh [--speech-dir <path>] [--profile <name>]
#
# Defaults:
#   --speech-dir  examples/omarchy
#   --profile     vanilla-arch
#
# Environment:
#   DEBATEOS_BIN   path to the debateos binary (default: ./debateos or go run ./cmd/debateos)
#
# Exit codes:
#   0  Determinism OK (both tarballs have identical sha256)
#   1  Determinism FAIL or script error

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

echo "=== DebateOS Determinism Gate (BLD-03) ==="
echo "  debateos: ${DEBATEOS}"
echo "  speech:   ${SPEECH_DIR}"
echo "  profile:  ${PROFILE}"
echo ""

# ─── Temporary directories ─────────────────────────────────────────────────
WORK_DIR="$(mktemp -d /tmp/debateos-det-XXXXXX)"
OUT1="${WORK_DIR}/run1"
OUT2="${WORK_DIR}/run2"
TAR1="${WORK_DIR}/run1.tar.gz"
TAR2="${WORK_DIR}/run2.tar.gz"

cleanup() { rm -rf "${WORK_DIR}"; }
trap cleanup EXIT

mkdir -p "${OUT1}" "${OUT2}"

# ─── Run 1 ─────────────────────────────────────────────────────────────────
echo "--- Run 1 (resolve + translate) ---"
(cd "${REPO_ROOT}" && "${DEBATEOS}" build \
    --dir  "${SPEECH_DIR}" \
    --profile "${PROFILE}" \
    --out  "${OUT1}" \
    --skip-iso)
echo ""

# ─── Run 2 ─────────────────────────────────────────────────────────────────
echo "--- Run 2 (resolve + translate) ---"
(cd "${REPO_ROOT}" && "${DEBATEOS}" build \
    --dir  "${SPEECH_DIR}" \
    --profile "${PROFILE}" \
    --out  "${OUT2}" \
    --skip-iso)
echo ""

# ─── Derive SOURCE_DATE_EPOCH from resolved.json ────────────────────────────
# Algorithm mirrors cli/build.DeriveEpoch + manifest.py derive_source_date_epoch:
#   SHA-256(resolved.json) → first 4 bytes as big-endian uint32 → MIN + raw % (MAX-MIN)
RESOLVED1="${OUT1}/resolved.json"

if ! command -v python3 &>/dev/null; then
    echo "ERROR: python3 is required to derive SOURCE_DATE_EPOCH" >&2
    exit 1
fi

EPOCH="$(python3 - "${RESOLVED1}" <<'PYEOF'
import hashlib, struct, sys

MIN_EPOCH = 1577836800  # 2020-01-01T00:00:00Z (mirrors cli/build.epochMin)
MAX_EPOCH = 2208988800  # 2040-01-01T00:00:00Z (mirrors cli/build.epochMax)

content = open(sys.argv[1], "rb").read()
digest  = hashlib.sha256(content).digest()
raw     = struct.unpack(">I", digest[:4])[0]
print(MIN_EPOCH + (raw % (MAX_EPOCH - MIN_EPOCH)))
PYEOF
)"

echo "--- Derived SOURCE_DATE_EPOCH: ${EPOCH} ---"
echo ""

# Export SOURCE_DATE_EPOCH so the gzip stream header uses the deterministic
# epoch rather than the current wall-clock time.  Without this export, GNU
# tar's -z/-czf pipes through gzip, which embeds time.Now() in the gzip
# header regardless of --mtime — two runs separated by even one second
# produce different SHA-256 checksums.  We additionally pipe explicitly
# through `gzip -n` (--no-name: suppress filename/timestamp in header)
# for defence-in-depth on any gzip version that does not honour
# SOURCE_DATE_EPOCH.
#
# Reference: https://reproducible-builds.org/docs/source-date-epoch/
# Mirror of cli/build comment noting "gzip -n" requirement (BLD-03 / CR-02).
export SOURCE_DATE_EPOCH="${EPOCH}"

# ─── Deterministic tar (03-RESEARCH.md Pattern 5, verified flags) ──────────
# Flags:
#   --sort=name                       alphabetical entry order
#   --mtime=@EPOCH                    uniform modification time
#   --owner=0 --group=0               no host-specific uid/gid
#   --numeric-owner                   numeric owner fields (no user/group name lookup)
#   --pax-option=...delete=atime,...  strip atime/ctime from PAX extended headers
#                                     (Pitfall 2: stale PAX timestamps break determinism)
#   gzip -n                           suppress filename+timestamp in gzip header
#                                     (defence-in-depth alongside SOURCE_DATE_EPOCH)
#
# The arch-profile/ tree is tarred (not resolved.json or private-injection.tar,
# which contain timestamps or are empty on this host path).
PROFILE_DIR1="${OUT1}/arch-profile"
PROFILE_DIR2="${OUT2}/arch-profile"

if [ ! -d "${PROFILE_DIR1}" ]; then
    echo "ERROR: arch-profile/ not found in run1 output: ${PROFILE_DIR1}" >&2
    exit 1
fi
if [ ! -d "${PROFILE_DIR2}" ]; then
    echo "ERROR: arch-profile/ not found in run2 output: ${PROFILE_DIR2}" >&2
    exit 1
fi

echo "--- Creating deterministic tarballs ---"

tar \
    --sort=name \
    --mtime="@${EPOCH}" \
    --owner=0 --group=0 --numeric-owner \
    --pax-option=exthdr.name=%d/PaxHeaders/%f,delete=atime,delete=ctime \
    -c -C "${PROFILE_DIR1}" . | gzip -n > "${TAR1}"

tar \
    --sort=name \
    --mtime="@${EPOCH}" \
    --owner=0 --group=0 --numeric-owner \
    --pax-option=exthdr.name=%d/PaxHeaders/%f,delete=atime,delete=ctime \
    -c -C "${PROFILE_DIR2}" . | gzip -n > "${TAR2}"

echo ""

# ─── SHA-256 comparison ─────────────────────────────────────────────────────
SHA1="$(sha256sum "${TAR1}" | cut -d' ' -f1)"
SHA2="$(sha256sum "${TAR2}" | cut -d' ' -f1)"

echo "Run 1 sha256: ${SHA1}"
echo "Run 2 sha256: ${SHA2}"
echo ""

if [ "${SHA1}" = "${SHA2}" ]; then
    echo "=== DETERMINISM OK: ${SHA1} ==="
    exit 0
else
    echo "=== DETERMINISM FAIL: ${SHA1} != ${SHA2} ===" >&2
    echo "" >&2
    echo "The two resolve+translate runs produced different profile trees." >&2
    echo "Check for:" >&2
    echo "  - Timestamps in generated files (look for \$(date) calls)" >&2
    echo "  - Non-deterministic dict iteration in generator.py" >&2
    echo "  - Missing SOURCE_DATE_EPOCH propagation in manifest.py" >&2
    echo "  - PAX header atime/ctime not stripped (add --pax-option delete=atime,delete=ctime)" >&2
    exit 1
fi
