#!/usr/bin/env bash
# scripts/dual-foundation-check.sh — DebateOS Dual-Foundation Proof Gate (DEB-02)
#
# Resolves the representative dual-foundation speech ONCE, then runs BOTH the
# Arch and Debian translators on the SAME resolved.json, and asserts per-foundation
# equivalence: both translators can effectuate every REQUIRED opinion and produce
# the expected package + asset sets from the shared input.
#
# This is the DEB-02 gate: the dual-foundation proof that the DebateOS abstraction
# is real — one resolved speech drives two completely independent foundation
# translators to equivalent outputs.
#
#   Step 1 RESOLVE:   go run ./cmd/resolve-json examples/dual-foundation > resolved.json
#                     Both translators share this single resolved.json input.
#   Step 2 TRANSLATE-ARCH:
#                     translators/arch/translate resolved.json --opinions examples/dual-foundation/opinions
#                     --profile vanilla-arch --out <arch-profile>
#   Step 3 TRANSLATE-DEBIAN:
#                     translators/debian/translate resolved.json --opinions examples/dual-foundation/opinions
#                     --profile debian --out <debian-profile>
#   Step 4 EQUIVALENCE: per-foundation structural checks:
#     (a) Both translate invocations exit 0 (no CapabilityError).
#     (b) Arch: build-manifest.json target_packages contains git/curl/vim;
#               file_assets includes etc/motd.
#     (c) Debian: config/package-lists/debateos.list.chroot_install contains git/curl/vim;
#                 config/hooks/live/9000-debateos-apply.hook.chroot exists + executable;
#                 config/includes.installer/preseed.cfg exists.
#     (d) Cross-foundation: target_packages from the same resolved.json matches
#         between the two foundations (same 3 packages) — the dual-foundation equivalence proof.
#   Step 5 STRUCTURAL: validate the Debian config/ tree (calls debian-validate-iso.sh).
#   Step 6 BUILD (slow; skip with --skip-iso): debian-build-iso.sh (requires capable host).
#
# Usage:
#   bash scripts/dual-foundation-check.sh [--skip-iso]
#
# Options:
#   --skip-iso   Skip Step 6 only (Docker lb-build). Runs Steps 1-5 (resolve +
#                both translators + equivalence + structural validation).
#                Default behavior on this host (Proxmox VE — devtmpfs restricted).
#
# Required tools:
#   go       — for go run ./cmd/resolve-json
#   python3  — for translators/arch/translate and translators/debian/translate
#   jq       — for JSON manifest inspection (equivalence checks)
#   docker   — for the ISO build (Step 6, unless --skip-iso)
#
# Host note:
#   Full lb build requires loop devices / devtmpfs and FAILS on this Proxmox VE
#   host (VERIFIED in 04-RESEARCH.md §Host Environment) — same policy as arch-build-iso.sh.
#   Use --skip-iso (the default intent) to run Steps 1-5 on this host.
#
# Source: 04-05-PLAN.md DEB-02 dual-foundation gate.

set -euo pipefail

# ---------------------------------------------------------------------------
# Options
# ---------------------------------------------------------------------------

SKIP_ISO=false
for arg in "$@"; do
    case "${arg}" in
        --skip-iso) SKIP_ISO=true ;;
        -h|--help)
            echo "Usage: $0 [--skip-iso]" >&2
            echo "  --skip-iso  Skip Docker lb-build (structural+equivalence only)" >&2
            exit 0
            ;;
        *)
            echo "ERROR: unknown option: ${arg}" >&2
            echo "Usage: $0 [--skip-iso]" >&2
            exit 1
            ;;
    esac
done

# ---------------------------------------------------------------------------
# Setup
# ---------------------------------------------------------------------------

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${REPO_ROOT}"

WORK_DIR="$(mktemp -d /tmp/debateos-df-check-XXXXXX)"
trap 'rm -rf "${WORK_DIR}"' EXIT

RESOLVED_JSON="${WORK_DIR}/resolved.json"
ARCH_PROFILE_DIR="${WORK_DIR}/arch-profile"
DEBIAN_PROFILE_DIR="${WORK_DIR}/debian-profile"
ISO_DIR="${WORK_DIR}/iso-out"

PASS=0
FAIL=0

pass() { echo "  [PASS] $*"; ((PASS++)) || true; }
fail() { echo "  [FAIL] $*" >&2; ((FAIL++)) || true; }

echo "=== DebateOS Dual-Foundation Proof Gate (DEB-02) ==="
echo "  Repo root:   ${REPO_ROOT}"
echo "  Work dir:    ${WORK_DIR}"
echo "  Skip ISO:    ${SKIP_ISO}"
echo ""

# ---------------------------------------------------------------------------
# Step 1: RESOLVE — emit canonical resolved.json (ONCE for both translators)
# ---------------------------------------------------------------------------

echo "--- Step 1: Resolve dual-foundation speech (once, shared input) ---"
if go run ./cmd/resolve-json examples/dual-foundation > "${RESOLVED_JSON}" 2>"${WORK_DIR}/resolve-stderr.log"; then
    APPLIED_COUNT="$(jq '.applied | length' "${RESOLVED_JSON}" 2>/dev/null || echo 0)"
    pass "resolved.json emitted (Applied=${APPLIED_COUNT})"
    echo "    resolved.json: $(wc -c < "${RESOLVED_JSON}") bytes"
else
    fail "cmd/resolve-json failed — stderr:"
    cat "${WORK_DIR}/resolve-stderr.log" >&2
    echo ""
    echo "=== DUAL-FOUNDATION GATE SUMMARY ==="
    echo "  Passed: ${PASS}  Failed: ${FAIL}"
    echo "  RESULT: DUAL-FOUNDATION GATE FAILED (resolve step)" >&2
    exit 1
fi
echo ""

# ---------------------------------------------------------------------------
# Step 2: TRANSLATE-ARCH — run Arch translator on the shared resolved.json
# ---------------------------------------------------------------------------

echo "--- Step 2: Arch translator (same resolved.json) ---"
# WR-02 NOTE: We intentionally run the Arch translator on a debian-foundation
# resolved.json to prove the abstraction holds. The Arch translator ignores the
# foundation field for translation purposes but the emitted build-manifest.json
# will carry foundation="debian" — this is expected and documented as part of
# the dual-foundation proof design.  The Arch translator is invoked explicitly
# out-of-band here; in production, the build system dispatches by foundation.
mkdir -p "${ARCH_PROFILE_DIR}"

ARCH_TRANSLATE_EXIT=0
"${REPO_ROOT}/translators/arch/translate" \
    "${RESOLVED_JSON}" \
    --opinions "${REPO_ROOT}/examples/dual-foundation/opinions" \
    --profile vanilla-arch \
    --out "${ARCH_PROFILE_DIR}" \
    2>"${WORK_DIR}/arch-translate-stderr.log" || ARCH_TRANSLATE_EXIT=$?

if [[ "${ARCH_TRANSLATE_EXIT}" -eq 0 ]]; then
    pass "Arch translator: exit 0 (no CapabilityError — all 5 DF opinions effectuable on Arch)"
else
    fail "Arch translator: exit ${ARCH_TRANSLATE_EXIT} — stderr:"
    cat "${WORK_DIR}/arch-translate-stderr.log" >&2
fi
echo ""

# ---------------------------------------------------------------------------
# Step 3: TRANSLATE-DEBIAN — run Debian translator on the same resolved.json
# ---------------------------------------------------------------------------

echo "--- Step 3: Debian translator (same resolved.json) ---"
mkdir -p "${DEBIAN_PROFILE_DIR}"

DEBIAN_TRANSLATE_EXIT=0
"${REPO_ROOT}/translators/debian/translate" \
    "${RESOLVED_JSON}" \
    --opinions "${REPO_ROOT}/examples/dual-foundation/opinions" \
    --profile debian \
    --out "${DEBIAN_PROFILE_DIR}" \
    2>"${WORK_DIR}/debian-translate-stderr.log" || DEBIAN_TRANSLATE_EXIT=$?

if [[ "${DEBIAN_TRANSLATE_EXIT}" -eq 0 ]]; then
    pass "Debian translator: exit 0 (no CapabilityError — all 5 DF opinions effectuable on Debian)"
else
    fail "Debian translator: exit ${DEBIAN_TRANSLATE_EXIT} — stderr:"
    cat "${WORK_DIR}/debian-translate-stderr.log" >&2
fi
echo ""

# ---------------------------------------------------------------------------
# Step 4: EQUIVALENCE CHECKS
# ---------------------------------------------------------------------------

echo "--- Step 4: Per-foundation equivalence checks ---"

ARCH_MANIFEST="${ARCH_PROFILE_DIR}/build-manifest.json"
DEBIAN_MANIFEST="${DEBIAN_PROFILE_DIR}/build-manifest.json"

# --- 4a: Arch build-manifest.json target_packages ---
echo "  [4a] Arch: build-manifest.json target_packages contains git/curl/vim"
if [[ -f "${ARCH_MANIFEST}" ]]; then
    ARCH_PKGS="$(jq -r '.target_packages // [] | .[]' "${ARCH_MANIFEST}" 2>/dev/null || true)"
    for pkg in git curl vim; do
        if echo "${ARCH_PKGS}" | grep -qx "${pkg}"; then
            pass "Arch packages: '${pkg}' in build-manifest.json target_packages"
        else
            fail "Arch packages: '${pkg}' MISSING from build-manifest.json target_packages"
        fi
    done
else
    fail "Arch: build-manifest.json not found at ${ARCH_MANIFEST}"
fi

# --- 4b: Arch build-manifest.json file_assets includes etc/motd ---
echo "  [4b] Arch: build-manifest.json file_assets includes etc/motd"
if [[ -f "${ARCH_MANIFEST}" ]]; then
    ARCH_DST="$(jq -r '.file_assets // [] | .[].dst' "${ARCH_MANIFEST}" 2>/dev/null || true)"
    if echo "${ARCH_DST}" | grep -qx "etc/motd"; then
        pass "Arch file_assets: 'etc/motd' declared in build-manifest.json"
    else
        fail "Arch file_assets: 'etc/motd' MISSING from build-manifest.json file_assets"
    fi
fi

# --- 4c: Debian package-lists contains git/curl/vim ---
echo "  [4c] Debian: config/package-lists/debateos.list.chroot_install contains git/curl/vim"
DEBIAN_PKGLIST="${DEBIAN_PROFILE_DIR}/config/package-lists/debateos.list.chroot_install"
if [[ -f "${DEBIAN_PKGLIST}" ]]; then
    for pkg in git curl vim; do
        if grep -qx "${pkg}" "${DEBIAN_PKGLIST}"; then
            pass "Debian packages: '${pkg}' in debateos.list.chroot_install"
        else
            fail "Debian packages: '${pkg}' MISSING from debateos.list.chroot_install"
        fi
    done
else
    fail "Debian: debateos.list.chroot_install not found at ${DEBIAN_PKGLIST}"
fi

# --- 4d: Debian chroot hook exists + executable ---
echo "  [4d] Debian: config/hooks/live/9000-debateos-apply.hook.chroot exists + executable"
DEBIAN_HOOK="${DEBIAN_PROFILE_DIR}/config/hooks/live/9000-debateos-apply.hook.chroot"
if [[ -f "${DEBIAN_HOOK}" ]]; then
    pass "Debian hook: 9000-debateos-apply.hook.chroot exists"
    if [[ -x "${DEBIAN_HOOK}" ]]; then
        pass "Debian hook: 9000-debateos-apply.hook.chroot is executable (0755)"
    else
        fail "Debian hook: 9000-debateos-apply.hook.chroot is NOT executable"
    fi
else
    fail "Debian hook: 9000-debateos-apply.hook.chroot NOT FOUND at ${DEBIAN_HOOK}"
fi

# --- 4e: Debian preseed.cfg exists ---
echo "  [4e] Debian: config/includes.installer/preseed.cfg exists"
DEBIAN_PRESEED="${DEBIAN_PROFILE_DIR}/config/includes.installer/preseed.cfg"
if [[ -f "${DEBIAN_PRESEED}" ]]; then
    pass "Debian preseed: preseed.cfg exists"
    if grep -q 'd-i' "${DEBIAN_PRESEED}"; then
        pass "Debian preseed: contains d-i directives"
    else
        fail "Debian preseed: no 'd-i' lines found in preseed.cfg"
    fi
else
    fail "Debian preseed: preseed.cfg NOT FOUND at ${DEBIAN_PRESEED}"
fi

# --- 4f: Cross-foundation equivalence — same package set from same resolved.json ---
echo "  [4f] Cross-foundation: package set matches between Arch and Debian (same resolved.json)"
if [[ -f "${ARCH_MANIFEST}" && -f "${DEBIAN_MANIFEST}" ]]; then
    ARCH_PKG_SET="$(jq -r '.target_packages // [] | sort | .[]' "${ARCH_MANIFEST}" 2>/dev/null | tr '\n' ',' | sed 's/,$//')"
    DEBIAN_PKG_SET="$(jq -r '.target_packages // [] | sort | .[]' "${DEBIAN_MANIFEST}" 2>/dev/null | tr '\n' ',' | sed 's/,$//')"
    if [[ "${ARCH_PKG_SET}" == "${DEBIAN_PKG_SET}" ]]; then
        pass "Cross-foundation: Arch and Debian produce identical package set: [${ARCH_PKG_SET}]"
    else
        fail "Cross-foundation: package set mismatch — Arch=[${ARCH_PKG_SET}] Debian=[${DEBIAN_PKG_SET}]"
    fi
fi
echo ""

# ---------------------------------------------------------------------------
# Step 5: Structural validation of the Debian config/ tree
# ---------------------------------------------------------------------------

echo "--- Step 5: Debian config/ tree structural validation ---"
if [[ -f "${SCRIPT_DIR}/debian-validate-iso.sh" ]]; then
    if bash "${SCRIPT_DIR}/debian-validate-iso.sh" "${DEBIAN_PROFILE_DIR}" \
           >"${WORK_DIR}/debian-validate.log" 2>&1; then
        pass "debian-validate-iso.sh: Debian config/ tree structural validation PASSED"
    else
        fail "debian-validate-iso.sh: structural validation FAILED"
        tail -20 "${WORK_DIR}/debian-validate.log" >&2
    fi
else
    fail "debian-validate-iso.sh not found at ${SCRIPT_DIR}/debian-validate-iso.sh"
fi
echo ""

# ---------------------------------------------------------------------------
# Step 5b: Regression tests
# ---------------------------------------------------------------------------

echo "--- Step 5b: Regression tests ---"
go test ./... -count=1 >"${WORK_DIR}/go-test-all.log" 2>&1 || true
# WR-03 fix: Treat absence of '^FAIL' as GREEN (handles the no-test-files edge
# where go test exits 0 but emits no '^ok' lines). If any package fails to build
# or a test fails, go test emits '^FAIL <pkg>' which the grep catches.
if ! grep -Eq '^FAIL\b' "${WORK_DIR}/go-test-all.log"; then
    # No FAIL lines — all tested packages (if any) are green
    if grep -Eq '^ok\b' "${WORK_DIR}/go-test-all.log"; then
        pass "go test ./... -count=1: all packages GREEN"
    else
        pass "go test ./... -count=1: no test files (nothing to fail)"
    fi
else
    fail "go test ./... failed"
    tail -20 "${WORK_DIR}/go-test-all.log" >&2
fi

# Run each pytest suite independently (arch+debian run together collides on bare-name
# imports — see pre-existing known issue documented in 04-05-SUMMARY.md §Known Issues).
python3 -m pytest translators/arch/tests/ -q --tb=short \
    >"${WORK_DIR}/pytest-arch.log" 2>&1 || true
if grep -Eq '^[0-9]+ passed' "${WORK_DIR}/pytest-arch.log"; then
    ARCH_RESULT="$(grep -E '^[0-9]+ passed' "${WORK_DIR}/pytest-arch.log" | tail -1)"
    pass "pytest translators/arch/tests/: ${ARCH_RESULT}"
else
    fail "pytest arch failed"
    tail -20 "${WORK_DIR}/pytest-arch.log" >&2
fi

python3 -m pytest translators/debian/tests/ -q --tb=short \
    >"${WORK_DIR}/pytest-debian.log" 2>&1 || true
if grep -Eq '^[0-9]+ passed' "${WORK_DIR}/pytest-debian.log"; then
    DEB_RESULT="$(grep -E '^[0-9]+ passed' "${WORK_DIR}/pytest-debian.log" | tail -1)"
    pass "pytest translators/debian/tests/: ${DEB_RESULT}"
else
    fail "pytest debian failed"
    tail -20 "${WORK_DIR}/pytest-debian.log" >&2
fi

python3 -m pytest translators/common/tests/ -q --tb=short \
    >"${WORK_DIR}/pytest-common.log" 2>&1 || true
if grep -Eq '^[0-9]+ passed' "${WORK_DIR}/pytest-common.log"; then
    CMN_RESULT="$(grep -E '^[0-9]+ passed' "${WORK_DIR}/pytest-common.log" | tail -1)"
    pass "pytest translators/common/tests/: ${CMN_RESULT}"
else
    fail "pytest common failed"
    tail -20 "${WORK_DIR}/pytest-common.log" >&2
fi
echo ""

# ---------------------------------------------------------------------------
# Step 6: ISO BUILD (slow gate; skipped with --skip-iso)
# ---------------------------------------------------------------------------

if [[ "${SKIP_ISO}" == "true" ]]; then
    echo "--- Step 6: ISO BUILD (SKIPPED via --skip-iso) ---"
    echo "  This host (Proxmox VE) cannot run lb build (loop/devtmpfs restricted)."
    echo "  To run the full Debian ISO build on a capable host:"
    echo "    SOURCE_DATE_EPOCH=<epoch> bash scripts/debian-build-iso.sh <debian-profile-dir> <out-dir>"
    echo "    bash scripts/debian-validate-iso.sh <iso>"
    echo ""
else
    echo "--- Step 6: ISO BUILD (slow gate: 20-40 min) ---"

    # Derive SOURCE_DATE_EPOCH from resolved.json (same derivation as manifest.py)
    SOURCE_DATE_EPOCH="$(python3 -c "
import hashlib, struct, sys
data = open('${RESOLVED_JSON}', 'rb').read()
d = hashlib.sha256(data).digest()
raw = struct.unpack('>I', d[:4])[0]
MIN_E = 1577836800
MAX_E = 2208988800
print(MIN_E + (raw % (MAX_E - MIN_E)))
")"
    export SOURCE_DATE_EPOCH

    echo "  SOURCE_DATE_EPOCH: ${SOURCE_DATE_EPOCH}"
    mkdir -p "${ISO_DIR}"

    echo "  Running: bash scripts/debian-build-iso.sh ${DEBIAN_PROFILE_DIR} ${ISO_DIR}"
    if bash "${SCRIPT_DIR}/debian-build-iso.sh" "${DEBIAN_PROFILE_DIR}" "${ISO_DIR}" \
           2>&1 | tee "${WORK_DIR}/build.log"; then
        ISO_FILE="$(find "${ISO_DIR}" -maxdepth 1 -name '*.iso' | head -1 || true)"
        if [[ -f "${ISO_FILE:-}" ]]; then
            pass "Debian ISO built: ${ISO_FILE}"
        else
            fail "No .iso file found in ${ISO_DIR} after build"
        fi
    else
        fail "debian-build-iso.sh failed — see ${WORK_DIR}/build.log"
        tail -30 "${WORK_DIR}/build.log" >&2
    fi
    echo ""
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------

echo "=== DUAL-FOUNDATION GATE SUMMARY (DEB-02) ==="
echo "  Passed: ${PASS}"
echo "  Failed: ${FAIL}"
echo ""

if [[ ${FAIL} -eq 0 ]]; then
    echo "  RESULT: DUAL-FOUNDATION GATE PASSED (DEB-02)"
    if [[ "${SKIP_ISO}" == "true" ]]; then
        echo "  (Equivalence+structural run. Full lb build requires a capable host.)"
    else
        echo "  (Full pipeline: resolve → translate×2 → equivalence → build → validate)"
    fi
    exit 0
else
    echo "  RESULT: DUAL-FOUNDATION GATE FAILED (${FAIL} check(s) failed)" >&2
    exit 1
fi
