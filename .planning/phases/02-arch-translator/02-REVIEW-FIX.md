---
phase: 02-arch-translator
fixed_at: 2026-06-13T00:00:00Z
review_path: .planning/phases/02-arch-translator/02-REVIEW.md
iteration: 1
findings_in_scope: 16
fixed: 16
skipped: 0
status: all_fixed
---

# Phase 02: Arch Translator — Code Review Fix Report

**Fixed at:** 2026-06-13
**Source review:** `.planning/phases/02-arch-translator/02-REVIEW.md`
**Iteration:** 1

**Summary:**
- Findings in scope: 16
- Fixed: 16
- Skipped: 0

## Fixed Issues

### CR-01: Installer template missing custom repo configuration

**Files modified:** `translators/arch/templates/installer.sh.tpl`, `translators/arch/tests/test_profile.py`
**Commit:** d276af9
**Applied fix:** Added a `custom_repos` block before `pacstrap` that reads `custom_repos` from
`build-manifest.json`, imports GPG keys via `pacman-key --recv-keys` / `--lsign-key`, appends
repo sections to the live `/etc/pacman.conf`, and runs `pacman -Sy`. Tests added for presence
of `custom_repos` and `pacman.conf` references in the generated installer.

### CR-02: Installer template has no bootloader installation step

**Files modified:** `translators/arch/templates/installer.sh.tpl`, `translators/arch/tests/test_profile.py`
**Commit:** d276af9
**Applied fix:** Added a data-driven bootloader block after `mkinitcpio -P`. Detects limine via
`manifest.bootloader` or by checking `target_packages | index("limine")`, installs limine BIOS +
EFI + writes limine.cfg if needed; defaults to `bootctl install` + loader entries for systemd-boot.
Kernel params from `manifest.kernel_params` are passed through to both paths. Test added asserting
`bootctl` is present in the generated installer (systemd-boot fallback). Commit shared with CR-01/CR-03.

### CR-03: NVMe partition naming produces wrong device paths

**Files modified:** `translators/arch/templates/installer.sh.tpl`, `translators/arch/tests/test_profile.py`
**Commit:** d276af9
**Applied fix:** Replaced the unconditional `EFI_PART="${TARGET_DISK}1"` lines with a bash
conditional that tests `[[ "$TARGET_DISK" =~ nvme|loop ]]` and uses `p1`/`p2` suffix for
NVMe/loop devices, falling back to `1`/`2` for SATA. Test added asserting `nvme` and `p1`
appear in the generated installer.

### CR-04: File-asset dst path validation runs after all profile files are written

**Files modified:** `translators/arch/profile.py`, `translators/arch/tests/test_profile.py`
**Commit:** 8fb5fe6
**Applied fix:** Moved the `file_assets` dst validation loop to the very start of
`emit_profile_tree`, before `os.makedirs()` and any write calls. A path traversal dst
now raises `ValueError` before any files are created. Tests added: `test_traversal_dst_no_files_written`
verifies that after a traversal ValueError the output directory contains no files.

### WR-01: Trust warnings miss OptionalTrustAll sig_level

**Files modified:** `translators/arch/manifest.py`, `translators/arch/variant.py`, `translators/arch/profile.py`, `translators/arch/tests/test_manifest.py`
**Commit:** 2471883
**Applied fix:** Extended the `if sig == "Never"` block to `elif sig == "OptionalTrustAll"` in
`manifest.py`, `variant.py`, and `profile.py._build_repo_section`. OM-001's omarchy repo now
produces a `trust_warning` entry and a pacman.conf comment block. Distinct messages for the two
levels (signatures fully bypassed vs unsigned packages accepted). Test added.

### WR-02: `_sanitize_dst` accepts empty string and '.' dst

**Files modified:** `translators/arch/profile.py`, `translators/arch/tests/test_profile.py`
**Commit:** 8fb5fe6
**Applied fix:** Added an explicit guard at the top of `_sanitize_dst` that raises `ValueError`
when `dst` is empty or `dst.strip()` is `"."`. Tests `test_empty_dst_raises` and
`test_dot_dst_raises` added. Fix shares commit with CR-04 (both are in `profile.py`).

### WR-03: sysctl drop-in file defaults to literal "null" when field absent

**Files modified:** `translators/arch/templates/installer.sh.tpl`
**Commit:** d276af9
**Applied fix:** Changed `jq -r '.file'` to `jq -r '.file // "99-debateos.conf"'` with an
additional `if [[ -z ... || == "null" ]]` guard so absent or null `drop_in_file` fields default
to `99-debateos.conf`. Shares commit with CR-01/CR-02/CR-03.

### WR-04: Dead code in `contract.py`

**Files modified:** `translators/arch/contract.py`
**Commit:** 79c69c9
**Applied fix:** Replaced `data[key] = [] if key != "explanations" else []` with `data[key] = []`
and added a clarifying comment. Both branches of the ternary returned `[]`; the conditional was dead.

### WR-05: TMPDIR environment variable clobbered in `arch-validate-iso.sh`

**Files modified:** `scripts/arch-validate-iso.sh`
**Commit:** 08a55f5
**Applied fix:** Renamed `TMPDIR` to `_VALIDATE_TMPDIR` throughout the script (assignment, trap,
and `SFS_PATH` reference). Added a comment explaining why the rename is needed.

### WR-06: Northstar check success message hardcodes Applied/Skipped counts

**Files modified:** `scripts/arch-northstar-check.sh`
**Commit:** 340085d
**Applied fix:** Replaced the static `"Applied=99 Skipped=35"` string with a `grep -oE` extraction
from `go-test-omarchy.log` so the pass message reflects the actual resolver output.

### WR-07: `--skip-build` comment in northstar script is inaccurate

**Files modified:** `scripts/arch-northstar-check.sh`
**Commit:** 340085d
**Applied fix:** Corrected the option header to say "Skip Step 4 only (Docker ISO build +
structural validation). Runs Steps 1-3 (resolve + translate + mechanical equivalence)." Shares
commit with WR-06.

### WR-08: `firstrun.service.tpl` is never loaded — dead template file

**Files modified:** `translators/arch/firstrun.py`
**Commit:** d46b63f
**Applied fix:** Replaced the `_UNIT_TEMPLATE` inline string constant with a `_load_unit_template()`
function that reads `templates/firstrun.service.tpl` at render time. The `.tpl` file is now the
single authoritative source. The inline duplicate was removed.

### IN-01: Invariant-1 violations in intent fields of OM-095 and OM-099

**Files modified:** `examples/omarchy/opinions/OM-095.yaml`, `examples/omarchy/opinions/OM-099.yaml`
**Commit:** 129cdcd
**Applied fix:** Rewrote both intent fields to use capability-level language free of tool names.
OM-095: "Deploy the Omarchy boot splash theme assets and activate them as the system default."
OM-099: "Configure the bootloader with snapshot boot-menu integration, set up initramfs generation,
register the EFI boot entry, and enable the snapshot synchronization service." Examples validation
(TestExampleOmarchy) passes after the change.

### IN-02: `generator.py` reads resolved speech file twice

**Files modified:** `translators/arch/generator.py`
**Commit:** 45b3881
**Applied fix:** Consolidated the two `open()` calls (text-mode `json.load` via
`load_resolved_speech` + binary read for `resolved_bytes`) into a single binary `open()` that
yields both `resolved_bytes` and the parsed dict. The `load_resolved_speech` import is removed;
key defaulting is done inline matching contract.py's behavior.

### IN-03: `cmd/resolve-json` pretty-print changes key ordering from canonical form

**Files modified:** `cmd/resolve-json/main.go`
**Commit:** 459b75d
**Applied fix:** Removed the `json.Unmarshal` → `json.MarshalIndent` round-trip that was sorting
keys alphabetically via `map[string]interface{}`. Now writes `CanonicalJSON` output bytes directly
to stdout followed by a newline. The unused `"encoding/json"` import was removed. The translator's
`json.loads()` handles compact JSON without issue.

### IN-04: Test files open files without context managers

**Files modified:** `translators/arch/tests/test_generator.py`, `translators/arch/tests/test_profile.py`
**Commit:** 91ea3a2
**Applied fix:** Replaced all `open(...).read()` and `json.load(open(...))` patterns with
`with open(...) as fh:` context managers in both test files. Covers all 16 unclosed opens
identified (line 94 in test_generator.py; multiple in test_profile.py).

## Skipped Issues

None — all 16 findings were fixed.

---

**Gate results:**
- `pytest translators/arch/tests -q`: 135 passed (7 new tests added for CR-01, CR-02, CR-03, CR-04, WR-01, WR-02 RED→GREEN)
- `go test ./... -count=1`: all packages GREEN (examples, graph, hardware, parse, patch, resolve)
- `bash scripts/arch-northstar-check.sh --skip-build`: 16/16 PASS (ARCH-02 north-star gate)

_Fixed: 2026-06-13T00:00:00Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
