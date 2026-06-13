---
phase: 02-arch-translator
reviewed: 2026-06-12T00:00:00Z
depth: standard
files_reviewed: 22
files_reviewed_list:
  - translators/arch/capabilities.py
  - translators/arch/contract.py
  - translators/arch/manifest.py
  - translators/arch/variant.py
  - translators/arch/firstrun.py
  - translators/arch/profile.py
  - translators/arch/generator.py
  - translators/arch/translate
  - translators/arch/capabilities.json
  - translators/arch/Dockerfile
  - translators/arch/profiles/vanilla-arch.yaml
  - translators/arch/profiles/cachyos.yaml
  - translators/arch/profiles/garuda.yaml
  - translators/arch/templates/installer.sh.tpl
  - translators/arch/templates/profiledef.sh.tpl
  - translators/arch/templates/pacman.conf.tpl
  - translators/arch/templates/firstrun.service.tpl
  - translators/arch/tests/test_capability_gate.py
  - translators/arch/tests/test_firstrun.py
  - translators/arch/tests/test_manifest.py
  - translators/arch/tests/test_profile.py
  - translators/arch/tests/test_variant.py
  - translators/arch/tests/test_generator.py
  - scripts/arch-build-iso.sh
  - scripts/arch-validate-iso.sh
  - scripts/arch-northstar-check.sh
  - cmd/resolve-json/main.go
  - examples/omarchy_test.go
  - examples/omarchy/speech.yaml
  - examples/omarchy/opinions/OM-001.yaml
  - examples/omarchy/opinions/OM-002.yaml
  - examples/omarchy/opinions/OM-095.yaml
  - examples/omarchy/opinions/OM-097.yaml
  - examples/omarchy/opinions/OM-099.yaml
  - examples/omarchy/opinions/OM-101.yaml
  - examples/omarchy/opinions/OM-114.yaml
findings:
  critical: 4
  warning: 8
  info: 4
  total: 16
status: issues_found
fixes_applied: true
resolved_ids:
  - CR-01
  - CR-02
  - CR-03
  - CR-04
  - WR-01
  - WR-02
  - WR-03
  - WR-04
  - WR-05
  - WR-06
  - WR-07
  - WR-08
  - IN-01
  - IN-02
  - IN-03
  - IN-04
fixed_at: 2026-06-13T00:00:00Z
---

# Phase 02: Arch Translator — Code Review Report

**Reviewed:** 2026-06-12
**Depth:** standard
**Files Reviewed:** 34 (across 22 primary + spot-checks)
**Status:** issues_found

## Summary

The Arch translator Python generator, profile emitter, and shell scripts are structurally sound for the nominal path (vanilla-arch, SATA disk, no custom repos). Four blockers prevent correct real-world operation: the installer template omits custom repo configuration (Omarchy packages from the private repo will not be found at install time), the installer template has no bootloader installation step (installed system will not boot), NVMe partition naming produces wrong device paths, and file-asset dst validation runs *after* all files are written (security gate fires too late). Eight warnings cover trust-warning coverage gaps, an unused dead-code branch, post-write validation ordering, missing drop-in file null guard, two TMPDIR clobbers, a misleading northstar pass message, and a dead template file. Four info items cover minor quality issues.

---

## Narrative Findings (AI reviewer)

## Critical Issues

### CR-01: Installer template missing custom repo configuration — Omarchy packages not installable

**File:** `translators/arch/templates/installer.sh.tpl:62-76`

**Issue:** The installer reads `keyring_install_before_repos` from `build-manifest.json` and installs those keyring packages, but it never adds the `custom_repos` entries from the manifest to the target system's `pacman.conf` before running `pacstrap`. OM-001 registers the Omarchy custom package repository (`https://packages.omarchy.org/stable`) with a GPG key that must be imported before packages from that repo can be installed. Without the repo configured and key imported, `pacstrap /mnt ... $TARGET_PKGS` will fail with "target not found" for any package that is in the Omarchy repo rather than in standard Arch repos (e.g., `limine`, `limine-mkinitcpio-hook`, `limine-snapper-sync`, `omarchy-*` packages). The ARCH-02 north-star equivalence gate cannot pass for the full Omarchy opinion set.

**Fix:** After the keyring install step and before `pacstrap`, read `custom_repos` from the manifest and:
1. Import each repo's GPG key fingerprint via `pacman-key --recv-keys` + `pacman-key --lsign-key`
2. Append the repo section to `/etc/pacman.conf` (the live ISO's pacman.conf)
3. Run `pacman -Sy` to sync the new repo

```bash
# After KEYRING_PKGS install, before pacstrap:
CUSTOM_REPOS=$(jq -c '.custom_repos // []' "$MANIFEST")
if [[ "$(echo "$CUSTOM_REPOS" | jq 'length')" -gt 0 ]]; then
    echo "Configuring custom repos..."
    echo "$CUSTOM_REPOS" | jq -r '.[] | "[" + .name + "]\nServer = " + .url + "\nSigLevel = " + .sig_level' \
        >> /etc/pacman.conf
    echo "$CUSTOM_REPOS" | jq -r '.[] | .keyring // empty' | while read -r key; do
        [[ -n "$key" ]] && pacman-key --recv-keys "$key" && pacman-key --lsign-key "$key" || true
    done
    pacman -Sy --noconfirm
fi
```

---

### CR-02: Installer template has no bootloader installation step — system will not boot

**File:** `translators/arch/templates/installer.sh.tpl:130-133`

**Issue:** The installer partitions the disk, runs `pacstrap`, generates `/etc/fstab`, configures locale, enables services, and runs `mkinitcpio -P`, but it never installs or configures a bootloader. The EFI partition is mounted at `/mnt/boot/efi` but no bootloader binary is written to it and no EFI boot entry is registered. The installed system will fail to boot because there is no bootloader pointing to the kernel. This is a functional blocker for every install the ISO produces.

**Fix:** Add a bootloader installation block after `mkinitcpio -P`. For the vanilla-arch/Omarchy target using limine (OM-099):

```bash
# After mkinitcpio -P:
BOOTLOADER=$(jq -r '.bootloader // ""' "$MANIFEST" 2>/dev/null || echo "")
if [[ "$BOOTLOADER" == "limine" ]] || jq -e '.target_packages | index("limine")' "$MANIFEST" >/dev/null 2>&1; then
    arch-chroot /mnt limine bios-install "$TARGET_DISK" || true
    mkdir -p /mnt/boot/efi/EFI/limine
    cp /mnt/usr/share/limine/limine-bios.sys /mnt/boot/efi/EFI/limine/
    arch-chroot /mnt limine enroll-config /boot/limine.cfg || true
else
    # Fallback: systemd-boot
    arch-chroot /mnt bootctl install
fi
```

---

### CR-03: NVMe partition naming produces wrong device paths — installer fails on NVMe drives

**File:** `translators/arch/templates/installer.sh.tpl:39-40`

**Issue:** The installer uses:
```bash
EFI_PART="${TARGET_DISK}1"
ROOT_PART="${TARGET_DISK}2"
```
For SATA/SCSI drives (`/dev/sda`), this produces `/dev/sda1` and `/dev/sda2` (correct). For NVMe drives (`/dev/nvme0n1`), this produces `/dev/nvme0n11` and `/dev/nvme0n12` — invalid device paths. NVMe partition names require a `p` separator: `/dev/nvme0n1p1`, `/dev/nvme0n1p2`. Modern Omarchy hardware (Framework, ASUS ROG) typically uses NVMe storage. The `mkfs.fat`, `mkfs.btrfs`, and all subsequent `mount` commands will fail on the invalid paths.

**Fix:**
```bash
# Determine partition suffix based on disk type
if [[ "$TARGET_DISK" =~ nvme|loop ]]; then
    EFI_PART="${TARGET_DISK}p1"
    ROOT_PART="${TARGET_DISK}p2"
else
    EFI_PART="${TARGET_DISK}1"
    ROOT_PART="${TARGET_DISK}2"
fi
```

---

### CR-04: File-asset dst path validation runs after all profile files are written

**File:** `translators/arch/profile.py:234-239`

**Issue:** `_sanitize_dst` (T-02-08 security gate) is called in a loop *after* `_write_profiledef`, `_write_packages_x86_64`, `_write_pacman_conf`, `_write_installer`, `_write_zlogin`, `_write_firstrun_units`, and `_write_build_manifest` have all completed. The comment explicitly states this is intentional: "Done last so the tree is fully written before raising on invalid assets." This means a path traversal attempt in any `file_asset.dst` field does not prevent the profile tree from being written to disk. A caller that catches the `ValueError` and continues (or a bug that ignores exceptions) would use a profile tree that passed partial validation. The security gate should fire *before* any file I/O.

**Fix:** Move the `file_asset` dst validation to the beginning of `emit_profile_tree`, before any writes:

```python
def emit_profile_tree(out_dir, manifest, variant):
    out_dir = str(out_dir)
    # T-02-08: Validate BEFORE any file I/O — fail fast, no partial output.
    for fa in manifest.file_assets:
        _sanitize_dst(fa.get("dst", ""))

    os.makedirs(out_dir, exist_ok=True)
    # ... rest of writes ...
```

Also add rejection of empty/dot dst values (see WR-02).

---

## Warnings

### WR-01: Trust warnings miss `OptionalTrustAll` sig_level — OM-001's repo silently passes

**File:** `translators/arch/manifest.py:252`, `translators/arch/variant.py:112`, `translators/arch/profile.py:162`

**Issue:** All three trust-warning paths only trigger when `sig_level == "Never"`. OM-001 declares `sig_level: OptionalTrustAll`, which permits unsigned packages from the Omarchy repo and carries the same security risk as `Never` (no per-package signature verification). The research notes this is already surfaced via resolver Explanation (T-01-10), but the translator itself emits no `trust_warning` entry for it, and the generated `pacman.conf` comment block is also silent. Users relying on `build-manifest.json`'s `trust_warnings` list for audit would miss this.

**Fix:** Extend the trust-warning check to include `OptionalTrustAll`:
```python
# In manifest.py line 252, variant.py line 112, profile.py line 162:
if repo.get("sig_level") in ("Never", "OptionalTrustAll"):
    trust_warnings.append(...)
```

---

### WR-02: `_sanitize_dst` accepts empty string and `.` dst — both map to sentinel root

**File:** `translators/arch/profile.py:105-143`

**Issue:** `_sanitize_dst("")` and `_sanitize_dst(".")` both pass validation (the normalized join equals the sentinel directory itself, which the `joined == sentinel` check explicitly allows) and return `"."` as the normalized relative path. If a `file_asset` has `dst: ""` or `dst: "."`, no `ValueError` is raised. Any downstream code attempting to use that as a file path destination would either fail with an opaque OS error or write to an unexpected location.

**Fix:** Add explicit guards at the top of `_sanitize_dst`:
```python
def _sanitize_dst(dst: str) -> str:
    if not dst or dst.strip() in (".", ""):
        raise ValueError(
            f"file_asset dst is empty or '.' — a concrete relative path is required (T-02-08)."
        )
    # ... existing checks ...
```

---

### WR-03: `sysctl_params` drop-in file is `"null"` string when `drop_in_file` field is absent

**File:** `translators/arch/templates/installer.sh.tpl:115-117`

**Issue:** The sysctl drop-in generation pipeline:
```bash
jq -r '.sysctl_params | group_by(.drop_in_file) | .[] | {file: .[0].drop_in_file, params: .} | @json' "$MANIFEST" \
  | while IFS= read -r group_json; do
      drop_file=$(echo "$group_json" | jq -r '.file')
      echo "$group_json" | jq -r '.params[] | "\(.key) = \(.value)"' >> "/mnt/etc/sysctl.d/$drop_file"
  done
```
If any `SysctlParam` has an empty or absent `drop_in_file` field, `jq -r '.file'` outputs the string `"null"` (not the empty string), and the params are appended to `/mnt/etc/sysctl.d/null` — a file named `null` that sysctl ignores silently.

**Fix:**
```bash
drop_file=$(echo "$group_json" | jq -r '.file // "99-debateos.conf"')
if [[ -z "$drop_file" || "$drop_file" == "null" ]]; then
    drop_file="99-debateos.conf"
fi
```

---

### WR-04: Dead code in `contract.py` — both branches of ternary return identical value

**File:** `translators/arch/contract.py:53`

**Issue:**
```python
data[key] = [] if key != "explanations" else []
```
Both the `if` and `else` branches return `[]`. The ternary has no effect. The original intent was likely `{} if key == "explanations" else []` (since explanations is a list of dicts, `[]` is correct but the conditional is pointless) or perhaps a different type for the explanations default. In any case, the conditional is dead code and masks what the intended default was.

**Fix:** Remove the dead conditional:
```python
data[key] = []
```
Or if the intent was to distinguish explanations from other keys, document why both are `[]`.

---

### WR-05: TMPDIR environment variable clobbered in `arch-validate-iso.sh`

**File:** `scripts/arch-validate-iso.sh:89`

**Issue:**
```bash
TMPDIR="$(mktemp -d)"
trap 'rm -rf "${TMPDIR}"' EXIT
```
`TMPDIR` is the POSIX standard environment variable that `mktemp`, `tmpfile`, and other tools read to determine where to create temporary files. Overwriting it with a generated path means any subsequent call to `mktemp` in the same shell (or subshells that inherit the environment) would create temp files inside the about-to-be-deleted directory. In this script the subsequent `mktemp` call for `SFS_PATH` is not present, but the clobber is a latent risk for any future maintenance that adds temp file creation.

**Fix:** Use a non-colliding name:
```bash
_VALIDATE_TMPDIR="$(mktemp -d)"
trap 'rm -rf "${_VALIDATE_TMPDIR}"' EXIT
SFS_PATH="${_VALIDATE_TMPDIR}/airootfs.sfs"
```

---

### WR-06: `arch-northstar-check.sh` success message hardcodes counts that may not match reality

**File:** `scripts/arch-northstar-check.sh:97`

**Issue:**
```bash
pass "TestExampleOmarchy: clean resolution (Applied=99 Skipped=35 Hard-conflicts=0)"
```
The pass message hardcodes `Applied=99 Skipped=35`. This string is emitted regardless of the actual resolver output — it is a static string, not derived from the test log. If the actual counts differ (which they will change as opinions are added or hardware-conditional logic is refined), the diagnostic output silently contradicts reality. A developer debugging a count regression would see a passing check with wrong counts.

**Fix:** Extract the actual counts from the test log:
```bash
COUNTS=$(grep -oE 'Applied=[0-9]+ Skipped=[0-9]+' "${WORK_DIR}/go-test-omarchy.log" | tail -1 || echo "counts unavailable")
pass "TestExampleOmarchy: clean resolution (${COUNTS} Hard-conflicts=0)"
```

---

### WR-07: `--skip-build` comment in northstar script is inaccurate

**File:** `scripts/arch-northstar-check.sh:27-28`

**Issue:** The script header says `--skip-build` skips "Steps 3 (build) and 4 (validate)", but the script has no Step 3 that is a build step — Step 3 is the mechanical equivalence checks, and Step 4 is the actual build+validate. The `--skip-build` flag skips only Step 4. The comment creates confusion about which steps run under `--skip-build`. The script's summary message does correctly qualify: `(Equivalence-only run. Full build path must also pass for the phase gate.)`, which is good, but the header is wrong.

**Fix:** Correct the header comment:
```bash
#   --skip-build    Skip Step 4 only (Docker ISO build + structural validation).
#                   Runs Steps 1-3 (resolve + translate + mechanical equivalence).
```

---

### WR-08: `firstrun.service.tpl` is never loaded — dead template file

**File:** `translators/arch/templates/firstrun.service.tpl`

**Issue:** `firstrun.py` defines `_UNIT_TEMPLATE` as a hardcoded string literal and renders it with `str.format()`. It never calls `_load_template("firstrun.service.tpl")`. The `.tpl` file exists in `templates/` and is identical to the inline string, but the two are separate copies with no enforcement mechanism. Future edits to `firstrun.service.tpl` will have no effect on generated units; edits to `_UNIT_TEMPLATE` in `firstrun.py` will not be reflected in the file. This is a two-source-of-truth situation that guarantees silent divergence.

**Fix:** Either:
- Delete `templates/firstrun.service.tpl` and keep the inline string (simpler)
- Or change `firstrun.py` to call `_load_template("firstrun.service.tpl")` at module level and remove `_UNIT_TEMPLATE`

The second option is preferred for consistency with how `profiledef.sh.tpl` and `pacman.conf.tpl` are handled in `profile.py`.

---

## Info

### IN-01: Invariant-1 violations in `intent` fields of OM-095 and OM-099

**File:** `examples/omarchy/opinions/OM-095.yaml:3-4`, `examples/omarchy/opinions/OM-099.yaml:3-8`

**Issue:** Invariant 1 (docs/02-concepts.md) requires that `intent` fields express *what* the opinion achieves, free of distro-specific mechanics. OM-095's intent mentions "Plymouth boot splash theme" (a specific Linux tool) and "Omarchy branded". OM-099's intent mentions "Limine bootloader", "mkinitcpio hooks", "Plymouth", "btrfs-overlayfs", "unified kernel image", "Thunderbolt module", and "Limine EFI boot entry" — tool and implementation details throughout. Opinions with distro-leaking intent fields fail schema validation against spirit if not letter.

**Fix:** Rewrite intent to be tool-agnostic:
- OM-095: "Deploy the Omarchy boot splash theme assets and activate them as the system default."
- OM-099: "Configure the bootloader with snapshot boot-menu integration, set up initramfs generation, register the EFI boot entry, and enable the snapshot synchronization service."

---

### IN-02: `generator.py` reads resolved speech file twice

**File:** `translators/arch/generator.py:85,92`

**Issue:**
```python
resolved = load_resolved_speech(resolved_path)   # line 85 — text mode parse
# ...
with open(resolved_path, "rb") as fh:             # line 92 — binary read for hash
    resolved_bytes = fh.read()
```
The file is opened twice. If the file is modified between reads (race condition, highly unlikely in practice), the parsed `resolved` dict and the `resolved_bytes` used for `SOURCE_DATE_EPOCH` derivation would be inconsistent. Additionally, `load_resolved_speech` already reads the file in its entirety; the bytes could be derived from that read.

**Fix:** Consolidate into one read:
```python
with open(resolved_path, "rb") as fh:
    resolved_bytes = fh.read()
resolved = json.loads(resolved_bytes.decode("utf-8"))
# Apply the same key-defaulting as load_resolved_speech:
for key in ("applied", "skipped", "dropped", "install_order", "explanations"):
    if key not in resolved:
        resolved[key] = []
```

---

### IN-03: `cmd/resolve-json` pretty-print round-trip changes key ordering from canonical form

**File:** `cmd/resolve-json/main.go:75-88`

**Issue:** `CanonicalJSON` marshals a typed Go struct (field order = struct declaration order). `resolve-json` then does `json.Unmarshal(out, &prettyBuf)` into `interface{}`, which decodes JSON objects as `map[string]interface{}`. `json.MarshalIndent` on a map sorts keys alphabetically. The pretty-printed output has keys in a different order than the canonical form, meaning the `SOURCE_DATE_EPOCH` derived from `resolve-json`'s output differs from what `CanonicalJSON` alone would produce. This is consistent (same input → same pretty output) but it means `SOURCE_DATE_EPOCH` is derived from a non-canonical representation.

**Fix (advisory):** This is low-risk since the derivation is consistent. To restore true canonical derivation, write the canonical bytes directly with a newline appended:
```go
os.Stdout.Write(out)
os.Stdout.Write([]byte("\n"))
```
If human-readable output is desired, use a separate `--pretty` flag rather than always re-encoding.

---

### IN-04: `test_generator.py` opens `build-manifest.json` without context manager

**File:** `translators/arch/tests/test_generator.py:94`

**Issue:**
```python
data = json.load(open(bm))
```
`open()` without `with` leaves the file handle unclosed until garbage collection. On CPython this is immediately collected, but it is a resource-leak pattern that fails on PyPy and may trigger `ResourceWarning` in test output.

**Fix:**
```python
with open(bm) as fh:
    data = json.load(fh)
```
Similar unclosed opens exist in `test_profile.py` lines 103, 122, 143, 190, 237.

---

_Reviewed: 2026-06-12T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
