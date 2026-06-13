#!/usr/bin/env bash
# build/docker/entrypoint.sh
#
# Entrypoint for the DebateOS Docker build image (BLD-01).
#
# Mounts expected by the caller:
#   /speech   — speech directory (mounted read-only: -v <speech-dir>:/speech:ro)
#   /out      — output directory for ISO and artifacts (-v <out-dir>:/out)
#
# The entrypoint invokes `debateos build` with the appropriate flags.
# --skip-iso is passed automatically when SKIP_ISO=1 is set in the container env.
#
# Usage (local channel — channel 1):
#   docker run --rm \
#     -v /path/to/speech:/speech:ro \
#     -v /path/to/out:/out \
#     [-e SKIP_ISO=1] \
#     [-e SOURCE_DATE_EPOCH=<epoch>] \
#     ghcr.io/mikl0s/debateos:latest
#
# Usage (GitHub Actions — channel 2):
#   See build/actions/build-speech.yml (workflow_call reusable workflow).
#
# Security (T-03-CTX):
#   The container receives only the speech directory and output directory.
#   pane.yaml, identity.age, and private-injection.tar are NEVER mounted into
#   the container — they stay on the user's local machine.
#
# Deferred verification:
#   Full mkarchiso ISO build requires a non-Proxmox host (host devtmpfs restriction).
#   Use SKIP_ISO=1 (or --skip-iso) for profile-emission-only builds on restricted hosts.

set -euo pipefail

SPEECH_DIR="${SPEECH_DIR:-/speech}"
OUT_DIR="${OUT_DIR:-/out}"
PROFILE="${PROFILE:-vanilla-arch}"

# Change to the debateos tooling root so relative translator paths resolve.
cd /debateos

# Build the --skip-iso flag if requested.
SKIP_ISO_FLAG=""
if [ "${SKIP_ISO:-0}" = "1" ]; then
    SKIP_ISO_FLAG="--skip-iso"
fi

echo "=== DebateOS Build ==="
echo "  speech:  ${SPEECH_DIR}"
echo "  out:     ${OUT_DIR}"
echo "  profile: ${PROFILE}"
[ -n "${SKIP_ISO_FLAG}" ] && echo "  mode:    profile-only (--skip-iso)"
echo ""

exec /usr/local/bin/debateos build \
    --dir  "${SPEECH_DIR}" \
    --profile "${PROFILE}" \
    --out  "${OUT_DIR}" \
    ${SKIP_ISO_FLAG}
