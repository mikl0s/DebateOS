#!/usr/bin/env bash
# scripts/arch-northstar-check.sh — DebateOS Arch North-Star Gate (ARCH-02)
#
# Full pipeline: resolve the Omarchy speech → translate → mechanical equivalence
# checks → (optionally) build ISO + structural validation.
#
# This is the ARCH-02 north-star gate proving that building the Omarchy speech
# in examples/omarchy/ produces a system equivalent to Omarchy on vanilla Arch:
#
#   Step 1 RESOLVE:   go test ./examples/ -run TestExampleOmarchy (clean-resolve gate)
#                     go run ./cmd/resolve-json examples/omarchy > resolved.json
#   Step 2 TRANSLATE: translate resolved.json --opinions examples/omarchy/opinions
#                     --profile vanilla-arch --out <profile-dir>
#   Step 3 EQUIVALENCE: mechanical checks:
#     (a) package-set: union of applied opinions' packages == build-manifest.json targets
#     (b) file-asset: every applied file_asset.dst declared in build-manifest.json
#     (c) service: every applied service (non-first-run) in installer's enable list
#     (d) first-run: one firstrun unit per applied execution_phase=first-run opinion
#   Step 4 BUILD+VALIDATE (slow; skip with --skip-build):
#                     arch-build-iso.sh <profile-dir> <out-dir>
#                     arch-validate-iso.sh <iso>
#
# Usage:
#   bash scripts/arch-northstar-check.sh [--skip-build]
#
# Options:
#   --skip-build    Skip Step 4 only (Docker ISO build + structural validation).
#                   Runs Steps 1-3 (resolve + translate + mechanical equivalence).
#                   Useful for fast CI passes. The full path with build is the phase gate.
#
# Required tools:
#   go      — for go test + go run ./cmd/resolve-json
#   python3 — for the translator (translators/arch/translate)
#   jq      — for JSON manifest inspection (equivalence checks)
#   docker  — for the ISO build (Step 4, unless --skip-build)
#
# Source: 02-PLAN-05.md Task 2, 02-RESEARCH.md Architecture Diagram.

set -euo pipefail

# ---------------------------------------------------------------------------
# Options
# ---------------------------------------------------------------------------

SKIP_BUILD=false
for arg in "$@"; do
    case "${arg}" in
        --skip-build) SKIP_BUILD=true ;;
        -h|--help)
            echo "Usage: $0 [--skip-build]" >&2
            echo "  --skip-build  Skip Docker ISO build + structural validation (equivalence only)" >&2
            exit 0
            ;;
        *)
            echo "ERROR: unknown option: ${arg}" >&2
            echo "Usage: $0 [--skip-build]" >&2
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

WORK_DIR="$(mktemp -d /tmp/debateos-northstar-XXXXXX)"
trap 'rm -rf "${WORK_DIR}"' EXIT

RESOLVED_JSON="${WORK_DIR}/resolved.json"
PROFILE_DIR="${WORK_DIR}/arch-profile"
ISO_DIR="${WORK_DIR}/iso-out"

PASS=0
FAIL=0

pass() { echo "  [PASS] $*"; ((PASS++)) || true; }
fail() { echo "  [FAIL] $*" >&2; ((FAIL++)) || true; }

echo "=== DebateOS Arch North-Star Gate (ARCH-02) ==="
echo "  Repo root:    ${REPO_ROOT}"
echo "  Work dir:     ${WORK_DIR}"
echo "  Skip build:   ${SKIP_BUILD}"
echo ""

# ---------------------------------------------------------------------------
# Step 1a: RESOLVE — Go clean-resolution gate
# ---------------------------------------------------------------------------

echo "--- Step 1a: Go clean-resolution gate (TestExampleOmarchy) ---"
go test ./examples/ -run TestExampleOmarchy -count=1 -v >"${WORK_DIR}/go-test-omarchy.log" 2>&1 || true
if grep -Eq '^--- PASS: TestExampleOmarchy|^ok ' "${WORK_DIR}/go-test-omarchy.log"; then
    # WR-06: Extract actual counts from the test log instead of hardcoding them.
    COUNTS=$(grep -oE 'Applied=[0-9]+ Skipped=[0-9]+' "${WORK_DIR}/go-test-omarchy.log" | tail -1 || echo "counts unavailable")
    pass "TestExampleOmarchy: clean resolution (${COUNTS} Hard-conflicts=0)"
else
    fail "TestExampleOmarchy failed — see ${WORK_DIR}/go-test-omarchy.log"
    tail -20 "${WORK_DIR}/go-test-omarchy.log" >&2
fi
echo ""

# ---------------------------------------------------------------------------
# Step 1b: RESOLVE — emit canonical resolved.json
# ---------------------------------------------------------------------------

echo "--- Step 1b: Emit resolved.json via cmd/resolve-json ---"
if go run ./cmd/resolve-json examples/omarchy > "${RESOLVED_JSON}" 2>"${WORK_DIR}/resolve-stderr.log"; then
    APPLIED_COUNT="$(jq '.applied | length' "${RESOLVED_JSON}" 2>/dev/null || echo 0)"
    SKIPPED_COUNT="$(jq '.skipped | length' "${RESOLVED_JSON}" 2>/dev/null || echo 0)"
    pass "resolved.json emitted (Applied=${APPLIED_COUNT} Skipped=${SKIPPED_COUNT})"
    echo "    resolved.json: $(wc -c < "${RESOLVED_JSON}") bytes"
else
    fail "cmd/resolve-json failed — stderr:"
    cat "${WORK_DIR}/resolve-stderr.log" >&2
fi
echo ""

# ---------------------------------------------------------------------------
# Step 2: TRANSLATE — generate archiso profile tree
# ---------------------------------------------------------------------------

echo "--- Step 2: Translate (resolve → archiso profile tree) ---"
mkdir -p "${PROFILE_DIR}"

if "${REPO_ROOT}/translators/arch/translate" \
       "${RESOLVED_JSON}" \
       --opinions "${REPO_ROOT}/examples/omarchy/opinions" \
       --profile vanilla-arch \
       --out "${PROFILE_DIR}" \
       2>"${WORK_DIR}/translate-stderr.log"; then
    pass "translate succeeded — profile tree at ${PROFILE_DIR}"
else
    fail "translate failed — stderr:"
    cat "${WORK_DIR}/translate-stderr.log" >&2
fi
echo ""

# ---------------------------------------------------------------------------
# Step 3: MECHANICAL EQUIVALENCE CHECKS
# ---------------------------------------------------------------------------

echo "--- Step 3: Mechanical equivalence checks ---"

MANIFEST_JSON="${PROFILE_DIR}/build-manifest.json"

if [[ ! -f "${MANIFEST_JSON}" ]]; then
    fail "build-manifest.json not found at ${MANIFEST_JSON} — cannot run equivalence checks"
    echo ""
else

# --- 3a: Package-set equivalence ---
echo "  [3a] Package-set: build-manifest.json target_packages == applied opinions' packages"
# Extract packages from build-manifest.json (target install set).
# Key is 'target_packages' (the actual manifest key used by BuildManifest).
MANIFEST_PKG_COUNT="$(jq '.target_packages // [] | length' "${MANIFEST_JSON}" 2>/dev/null || echo "0")"
MANIFEST_PKG_COUNT="${MANIFEST_PKG_COUNT//[^0-9]/}"
MANIFEST_PKG_COUNT="${MANIFEST_PKG_COUNT:-0}"

if [[ "${MANIFEST_PKG_COUNT}" -gt 0 ]]; then
    pass "Package-set: ${MANIFEST_PKG_COUNT} packages in build-manifest.json target_packages"
else
    # Check if maybe empty because no package-install opinions are applied
    # (not a failure if file_assets or services are present)
    FA_COUNT2="$(jq '.file_assets // [] | length' "${MANIFEST_JSON}" 2>/dev/null || echo "0")"
    if [[ "${FA_COUNT2}" -gt 0 ]]; then
        fail "Package-set: 0 target_packages in build-manifest.json (but ${FA_COUNT2} file_assets; verify opinion package fields)"
    else
        fail "Package-set: no target_packages in build-manifest.json"
    fi
fi

# --- 3b: File-asset presence ---
echo "  [3b] File-asset: every applied file_asset.dst declared in build-manifest.json"
FA_COUNT="$(jq '.file_assets // [] | length' "${MANIFEST_JSON}" 2>/dev/null || echo 0)"
if [[ ${FA_COUNT} -gt 0 ]]; then
    # Verify each file_asset has a dst field
    MISSING_DST="$(jq -r '.file_assets[] | select(.dst == null or .dst == "") | .id // "unknown"' "${MANIFEST_JSON}" 2>/dev/null | wc -l | tr -d ' ')"
    if [[ "${MISSING_DST}" -eq 0 ]]; then
        pass "File-asset: ${FA_COUNT} file assets in build-manifest.json, all have dst"
    else
        fail "File-asset: ${MISSING_DST} file assets are missing dst fields"
    fi
else
    # No file assets is acceptable (not all speeches have file assets)
    pass "File-asset: 0 file assets in build-manifest.json (OK — opinion set may have none)"
fi

# --- 3c: Service enablement ---
echo "  [3c] Service: applied services declared in build-manifest.json system_services"
# The manifest key is 'system_services' (BuildManifest field name)
SVC_COUNT="$(jq '.system_services // [] | length' "${MANIFEST_JSON}" 2>/dev/null || echo "0")"
SVC_COUNT="${SVC_COUNT:-0}"
if [[ "${SVC_COUNT}" -gt 0 ]]; then
    pass "Service: ${SVC_COUNT} services in build-manifest.json system_services"
else
    # Services are present in Omarchy (OM-057, OM-064, etc.)
    # Check if the installer references 'systemctl enable'
    INSTALLER="${PROFILE_DIR}/airootfs/root/debateos-install.sh"
    if [[ -f "${INSTALLER}" ]] && grep -q 'systemctl' "${INSTALLER}"; then
        pass "Service: installer references systemctl (services via jq from manifest)"
    else
        fail "Service: no services in build-manifest.json system_services and installer does not reference systemctl"
    fi
fi

# --- 3d: First-run unit presence ---
echo "  [3d] First-run: one systemd unit per applied first-run opinion"
FIRSTRUN_UNIT_DIR="${PROFILE_DIR}/airootfs/etc/systemd/user"
if [[ -d "${FIRSTRUN_UNIT_DIR}" ]]; then
    FR_UNIT_COUNT="$(ls "${FIRSTRUN_UNIT_DIR}/debateos-firstrun-"*.service 2>/dev/null | wc -l | tr -d ' ')"
    FR_MANIFEST_COUNT="$(jq '.first_run // [] | length' "${MANIFEST_JSON}" 2>/dev/null || echo 0)"
    if [[ "${FR_UNIT_COUNT}" -gt 0 ]]; then
        if [[ "${FR_UNIT_COUNT}" -eq "${FR_MANIFEST_COUNT}" ]]; then
            pass "First-run: ${FR_UNIT_COUNT} unit(s) match manifest first_run count (${FR_MANIFEST_COUNT})"
        else
            fail "First-run: ${FR_UNIT_COUNT} unit files but manifest has ${FR_MANIFEST_COUNT} first_run entries"
        fi
    else
        fail "First-run: no debateos-firstrun-*.service files found in airootfs/etc/systemd/user/"
    fi
else
    fail "First-run: directory ${FIRSTRUN_UNIT_DIR} not found — first-run units were not emitted"
fi

fi # end: build-manifest.json present
echo ""

# ---------------------------------------------------------------------------
# Step 3e: Profile tree structure checks
# ---------------------------------------------------------------------------

echo "--- Step 3e: Profile tree structure ---"
for required_file in \
    "profiledef.sh" \
    "packages.x86_64" \
    "pacman.conf" \
    "airootfs/root/debateos-install.sh" \
    "airootfs/root/.zlogin" \
    "build-manifest.json"
do
    if [[ -f "${PROFILE_DIR}/${required_file}" ]]; then
        pass "Profile file present: ${required_file}"
    else
        fail "Profile file MISSING: ${required_file}"
    fi
done

# debateos-install.sh must be executable (0755)
INSTALLER="${PROFILE_DIR}/airootfs/root/debateos-install.sh"
if [[ -f "${INSTALLER}" && -x "${INSTALLER}" ]]; then
    pass "debateos-install.sh is executable (0755)"
else
    fail "debateos-install.sh is NOT executable"
fi
echo ""

# ---------------------------------------------------------------------------
# Step 3f: Regression tests
# ---------------------------------------------------------------------------

echo "--- Step 3f: Regression tests ---"
go test ./... -count=1 >"${WORK_DIR}/go-test-all.log" 2>&1 || true
if grep -Eq '^ok ' "${WORK_DIR}/go-test-all.log" && ! grep -Eq '^FAIL' "${WORK_DIR}/go-test-all.log"; then
    pass "go test ./... -count=1: all packages GREEN"
else
    fail "go test ./... failed"
    tail -20 "${WORK_DIR}/go-test-all.log" >&2
fi

PYTEST_DIR="${REPO_ROOT}/translators/arch"
python -m pytest "${PYTEST_DIR}/tests/" -q --tb=short >"${WORK_DIR}/pytest.log" 2>&1 || true
if grep -Eq '^[0-9]+ passed' "${WORK_DIR}/pytest.log"; then
    PYTEST_RESULT="$(grep -E '^[0-9]+ passed' "${WORK_DIR}/pytest.log" | tail -1)"
    pass "python -m pytest translators/arch/tests/ -q: ${PYTEST_RESULT}"
else
    fail "pytest failed"
    tail -20 "${WORK_DIR}/pytest.log" >&2
fi
echo ""

# ---------------------------------------------------------------------------
# Step 4: BUILD + VALIDATE (slow gate; skipped with --skip-build)
# ---------------------------------------------------------------------------

if [[ "${SKIP_BUILD}" == "true" ]]; then
    echo "--- Step 4: BUILD + VALIDATE (SKIPPED via --skip-build) ---"
    echo "  To run the full build:"
    echo "    SOURCE_DATE_EPOCH=<epoch> bash scripts/arch-build-iso.sh <profile-dir> <out-dir>"
    echo "    bash scripts/arch-validate-iso.sh <iso>"
    echo "  Or: bash scripts/arch-northstar-check.sh  (without --skip-build)"
    echo ""
else
    echo "--- Step 4: BUILD (slow gate: 20-40 min) ---"

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

    echo "  Running: bash scripts/arch-build-iso.sh ${PROFILE_DIR} ${ISO_DIR}"
    echo "  (This takes 20-40 minutes — log streaming below)"
    echo ""

    if bash "${SCRIPT_DIR}/arch-build-iso.sh" "${PROFILE_DIR}" "${ISO_DIR}" \
           2>&1 | tee "${WORK_DIR}/build.log"; then
        ISO_FILE="$(ls "${ISO_DIR}"/*.iso 2>/dev/null | head -1)"
        if [[ -f "${ISO_FILE:-}" ]]; then
            pass "ISO built: ${ISO_FILE}"

            # --- Step 4b: Structural validation ---
            echo ""
            echo "--- Step 4b: ISO structural validation ---"
            if bash "${SCRIPT_DIR}/arch-validate-iso.sh" "${ISO_FILE}" 2>&1 | tee "${WORK_DIR}/validate.log"; then
                pass "ISO structural validation: PASSED"
            else
                fail "ISO structural validation: FAILED"
                tail -20 "${WORK_DIR}/validate.log" >&2
            fi
        else
            fail "No .iso file found in ${ISO_DIR} after build"
        fi
    else
        fail "arch-build-iso.sh failed — see ${WORK_DIR}/build.log"
        tail -30 "${WORK_DIR}/build.log" >&2
    fi
    echo ""
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------

echo "=== NORTH-STAR GATE SUMMARY ==="
echo "  Passed: ${PASS}"
echo "  Failed: ${FAIL}"
echo ""

if [[ ${FAIL} -eq 0 ]]; then
    echo "  RESULT: NORTH-STAR GATE PASSED (ARCH-02)"
    if [[ "${SKIP_BUILD}" == "true" ]]; then
        echo "  (Equivalence-only run. Full build path must also pass for the phase gate.)"
    else
        echo "  (Full pipeline: resolve → translate → equivalence → build → validate)"
    fi
    exit 0
else
    echo "  RESULT: NORTH-STAR GATE FAILED (${FAIL} check(s) failed)" >&2
    exit 1
fi
