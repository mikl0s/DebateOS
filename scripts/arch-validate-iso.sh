#!/usr/bin/env bash
# scripts/arch-validate-iso.sh — DebateOS Arch ISO structural validator
#
# Validates that a built Arch ISO is structurally bootable:
#   1. ISO9660 / El Torito primary volume descriptor (xorriso pvd_info)
#   2. EFI boot entries present under /EFI (systemd-boot UEFI path)
#   3. airootfs.sfs present under /arch/x86_64/ (Arch live-env squashfs)
#   4. debateos-install.sh present in airootfs (the generated installer)
#   5. .zlogin present in airootfs root/ (installer hook, Pattern 1)
#   6. At least one first-run unit present in airootfs (systemd user units)
#
# NOTE: This is the STRUCTURAL bootability gate. QEMU boot smoke is documented
# as an OPTIONAL manual step (CONTEXT.md — no QEMU on build host).
#
# Usage:
#   bash scripts/arch-validate-iso.sh <iso-file>
#
# Arguments:
#   <iso-file>   Path to the .iso file produced by arch-build-iso.sh.
#
# Dependencies (all available inside the Docker build image):
#   xorriso      — ISO9660 inspection (part of libisoburn; archiso dep)
#   unsquashfs   — squashfs listing (part of squashfs-tools; archiso dep)
#   mktemp       — temporary directory for squashfs extraction
#
# If xorriso or unsquashfs are not on the host, run inside the build container:
#   docker run --rm -v $(pwd):/work archlinux:base-devel@sha256:dd60dfcca90f1... \
#     bash -c "pacman -Sy --noconfirm libisoburn squashfs-tools && \
#              bash /work/scripts/arch-validate-iso.sh /work/<iso>"
#
# Exit codes:
#   0   All structural checks passed.
#   1   One or more checks failed (details printed to stderr).
#
# Source: 02-RESEARCH.md ISO Structural Validation code example.

set -euo pipefail

# ---------------------------------------------------------------------------
# Argument validation
# ---------------------------------------------------------------------------

if [[ $# -ne 1 ]]; then
    echo "ERROR: wrong number of arguments" >&2
    echo "Usage: $0 <iso-file>" >&2
    exit 1
fi

ISO="${1}"

if [[ ! -f "${ISO}" ]]; then
    echo "ERROR: ISO file not found: ${ISO}" >&2
    exit 1
fi

# ---------------------------------------------------------------------------
# Dependency checks
# ---------------------------------------------------------------------------

MISSING_DEPS=()
for cmd in xorriso unsquashfs; do
    if ! command -v "${cmd}" &>/dev/null; then
        MISSING_DEPS+=("${cmd}")
    fi
done

if [[ ${#MISSING_DEPS[@]} -gt 0 ]]; then
    echo "ERROR: missing required tools: ${MISSING_DEPS[*]}" >&2
    echo "  Install on Arch: pacman -Sy libisoburn squashfs-tools" >&2
    echo "  Or run inside the Docker build image (see script header)." >&2
    exit 1
fi

# ---------------------------------------------------------------------------
# Setup
# ---------------------------------------------------------------------------

PASS=0
FAIL=0

pass() { echo "  [PASS] $*"; ((PASS++)) || true; }
fail() { echo "  [FAIL] $*" >&2; ((FAIL++)) || true; }

echo "=== DebateOS Arch ISO Structural Validation ==="
echo "  ISO: ${ISO}"
echo "  Size: $(du -sh "${ISO}" | cut -f1)"
echo ""

TMPDIR="$(mktemp -d)"
trap 'rm -rf "${TMPDIR}"' EXIT

# ---------------------------------------------------------------------------
# Check 1: ISO9660 Primary Volume Descriptor (El Torito)
# ---------------------------------------------------------------------------

echo "--- Check 1: ISO9660 / El Torito PVD ---"
PVD_OUTPUT="$(xorriso -indev "${ISO}" -pvd_info 2>&1 || true)"
if echo "${PVD_OUTPUT}" | grep -Eq 'Volume id|System Area|ISO 9660'; then
    pass "ISO9660 primary volume descriptor found"
else
    fail "ISO9660 PVD not found (xorriso pvd_info output did not contain 'Volume id')"
    echo "  xorriso output:" >&2
    echo "${PVD_OUTPUT}" | head -20 >&2
fi
echo ""

# ---------------------------------------------------------------------------
# Check 2: Boot entries (EFI / syslinux)
# ---------------------------------------------------------------------------

echo "--- Check 2: Boot entries ---"
FIND_OUTPUT="$(xorriso -indev "${ISO}" -find / -type f 2>&1 || true)"

# Check for EFI (systemd-boot UEFI)
if echo "${FIND_OUTPUT}" | grep -qi '/EFI'; then
    pass "EFI boot entries found (UEFI / systemd-boot)"
else
    fail "No EFI entries found under /EFI — UEFI boot may not work"
fi

# Check for syslinux (BIOS fallback) — directory or cfg file
if echo "${FIND_OUTPUT}" | grep -qi 'syslinux\|isolinux'; then
    pass "Syslinux/isolinux entries found (BIOS fallback)"
else
    fail "No syslinux/isolinux entries found — BIOS boot may not work"
fi
echo ""

# ---------------------------------------------------------------------------
# Check 3: airootfs.sfs presence
# ---------------------------------------------------------------------------

echo "--- Check 3: airootfs.sfs ---"
if echo "${FIND_OUTPUT}" | grep -q '/arch/x86_64/airootfs.sfs'; then
    pass "airootfs.sfs found at /arch/x86_64/airootfs.sfs"
else
    fail "airootfs.sfs not found at /arch/x86_64/airootfs.sfs"
    echo "  Files under /arch/x86_64/:" >&2
    echo "${FIND_OUTPUT}" | grep -i '/arch/x86_64' | head -10 >&2
fi
echo ""

# ---------------------------------------------------------------------------
# Check 4-6: squashfs content (installer + .zlogin + first-run units)
# ---------------------------------------------------------------------------

echo "--- Check 4-6: squashfs content (installer / .zlogin / first-run units) ---"

# Extract just the airootfs.sfs from the ISO
SFS_PATH="${TMPDIR}/airootfs.sfs"
if xorriso -indev "${ISO}" -osirrox on -extract /arch/x86_64/airootfs.sfs "${SFS_PATH}" 2>/dev/null; then
    pass "airootfs.sfs extracted successfully"

    # Get squashfs listing
    SFS_LISTING="$(unsquashfs -l "${SFS_PATH}" 2>/dev/null || true)"

    # Check 4: debateos-install.sh
    if echo "${SFS_LISTING}" | grep -q 'debateos-install.sh'; then
        pass "debateos-install.sh found in airootfs"
    else
        fail "debateos-install.sh NOT found in airootfs — installer is missing"
        echo "  Files under root/ in squashfs:" >&2
        echo "${SFS_LISTING}" | grep 'root/' | head -20 >&2
    fi

    # Check 5: .zlogin
    if echo "${SFS_LISTING}" | grep -q '\.zlogin'; then
        pass ".zlogin found in airootfs root/ (installer hook, Pattern 1)"
    else
        fail ".zlogin NOT found in airootfs root/ — installer auto-invocation hook missing"
    fi

    # Check 6: first-run units (at least one debateos-firstrun-*.service)
    if echo "${SFS_LISTING}" | grep -Eq 'debateos-firstrun-[A-Z0-9-]+\.service'; then
        FR_COUNT="$(echo "${SFS_LISTING}" | grep -c 'debateos-firstrun-.*\.service' || echo 0)"
        pass "${FR_COUNT} first-run unit(s) found in airootfs (etc/systemd/user/)"
    else
        fail "No first-run units (debateos-firstrun-*.service) found in airootfs"
        echo "  Systemd units in squashfs:" >&2
        echo "${SFS_LISTING}" | grep -i 'systemd' | head -20 >&2
    fi
else
    fail "Failed to extract airootfs.sfs from ISO"
    echo "  Checks 4-6 skipped (could not extract squashfs)" >&2
    ((FAIL += 3)) || true
fi
echo ""

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------

echo "=== VALIDATION SUMMARY ==="
echo "  Passed: ${PASS}"
echo "  Failed: ${FAIL}"
echo ""

if [[ ${FAIL} -eq 0 ]]; then
    echo "  RESULT: STRUCTURAL VALIDATION PASSED"
    echo "  The ISO is structurally bootable (ISO9660 + El Torito + boot entries + installer)."
    echo "  QEMU boot smoke is an OPTIONAL manual step (no QEMU on build host)."
    exit 0
else
    echo "  RESULT: STRUCTURAL VALIDATION FAILED (${FAIL} check(s) failed)" >&2
    exit 1
fi
