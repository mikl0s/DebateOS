---
phase: 04-debian-translator
plan: "03"
subsystem: translators/debian
tags: [deb-01, deb-03, capability-gate, apt-sig-level, preseed, chroot-hook, tdd]
dependency_graph:
  requires: [04-01-SUMMARY.md, 04-02-SUMMARY.md]
  provides: [translators/debian/, translators/debian/translate]
  affects: [04-04-PLAN.md, 04-05-PLAN.md]
tech_stack:
  added: []
  patterns: [TDD RED/GREEN, %%SENTINEL%% template replacement, shim re-export, _sanitize_dst security gate, apply_variant apt mapping]
key_files:
  created:
    - translators/debian/__init__.py
    - translators/debian/pytest.ini
    - translators/debian/capabilities.py
    - translators/debian/capabilities.json
    - translators/debian/contract.py
    - translators/debian/manifest.py
    - translators/debian/variant.py
    - translators/debian/generator.py
    - translators/debian/profile.py
    - translators/debian/translate
    - translators/debian/profiles/debian.yaml
    - translators/debian/templates/preseed.cfg.tpl
    - translators/debian/templates/chroot-install.hook.tpl
    - translators/debian/tests/__init__.py
    - translators/debian/tests/fixtures/df_resolved.json
    - translators/debian/tests/fixtures/df_opinions.json
    - translators/debian/tests/test_capability_gate.py
    - translators/debian/tests/test_variant.py
    - translators/debian/tests/test_profile.py
    - translators/debian/tests/test_generator.py
  modified: []
decisions:
  - "[04-03] capabilities.json declares 45 foundation-neutral apt-effectuable tokens; excludes mkinitcpio/limine/pacman-AUR (DEB-03 Findings 3/4)"
  - "[04-03] apply_variant(variant) takes only the variant dict (no base_repos arg) — Debian has no fixed base repo list; all repos come from opinions' custom_repos or variant YAML"
  - "[04-03] manifest.py + contract.py shim re-exports (same pattern as arch) — common/ is single source of truth; bare-name imports work unchanged"
  - "[04-03] %%SENTINEL%% replacement for both preseed.cfg.tpl and chroot-install.hook.tpl — avoids str.format KeyError on ${SHELL_VAR} syntax (T-04-06, mirrors T-02-01)"
  - "[04-03] package-lists use .list.chroot_install suffix (both live + installed system) — .list.chroot would make packages live-only (Pitfall 1)"
  - "[04-03] fail-fast gate: all file_asset dst paths validated BEFORE any file I/O in emit_profile_tree — no partial output on path-traversal error (T-04-05 / CR-04)"
metrics:
  duration: "~35 min"
  completed: "2026-06-13T14:30:00Z"
  tasks_completed: 3
  files_changed: 20
---

# Phase 4 Plan 03: Debian Translator Core Summary

**One-liner:** Debian translator with capability gate (45 tokens, no mkinitcpio/limine), apt sig_level → signed-by/trusted=yes mapping, and live-build config/ tree emitter (preseed.cfg + chroot hook + package-lists) via RED/GREEN TDD.

## What Was Built

### Task 1 — RED: Capability gate + apt-source variant + manifest reuse tests (commit a33d8ac)

Written BEFORE any implementation. Tests cover:

- `test_capability_gate.py` (18 tests): 5 dual-foundation tokens required in capabilities.json; mkinitcpio/limine excluded; CapabilityError names opinion+token+"composition time" for required; nice-to-have dropped not raised
- `test_variant.py` (16 tests): debian.yaml loads with initramfs-tools/grub2 defaults; sig_level Required/RequiredDatabaseOptional → signed-by; OptionalTrustAll/Never → trusted=yes + LOUD WARNING comment; manifest reuse from common/
- Fixtures: df_resolved.json (foundation: debian, 5 DF opinions applied) + df_opinions.json (5 DF + 2 ARCH-ONLY opinions for gate tests)

RED confirmed: collection fails on `No module named 'capabilities'` and `No module named 'variant'`.

### Task 2 — GREEN: capabilities.py/json + variant.py + debian.yaml (commit 72de11f)

**capabilities.py** — mirrors arch/capabilities.py; error message says "Debian translator"; points to debian/capabilities.json.

**capabilities.json** — 45 foundation-neutral apt-effectuable tokens:

| Included (5 dual-foundation tokens) | Excluded (Arch-only) |
|--------------------------------------|----------------------|
| install-packages | configure-mkinitcpio-hooks-and-modules |
| deploy-config-file-tree | write-mkinitcpio-config-drop-in |
| enable-systemd-service | write-mkinitcpio-module-list |
| write-sysctl-drop-in | manage-limine-bootloader-installation |
| add-user-to-group | write-bootloader-entry-tool-drop-in |

Plus 40 additional foundation-neutral tokens (systemd, apt, sysctl, network, etc.).

**variant.py** — `load_variant_profile` + `apply_variant(variant)` (Debian-specific sig_level → apt mapping):

| Schema sig_level | apt option | Trust warning |
|-----------------|-----------|---------------|
| Required | `[signed-by=/etc/apt/trusted.gpg.d/NAME.asc]` | None |
| RequiredDatabaseOptional | `[signed-by=/etc/apt/trusted.gpg.d/NAME.asc]` | None |
| OptionalTrustAll | `[trusted=yes]` | Yes (T-04-07) |
| Never | `[trusted=yes]` + LOUD `# WARNING:` comment | Yes + louder (T-04-07) |

**profiles/debian.yaml** — Debian stable: `initramfs: initramfs-tools`, `bootloader: grub2`, `filesystem: ext4`, `repos: []`. Structure accommodates future ubuntu.yaml (ARCH-04 invariant).

34 tests GREEN.

### Task 3 — RED→GREEN: profile.py + generator.py + translate (commit 456a73a)

**profile.py** — `emit_profile_tree(out_dir, manifest, variant)` emits the live-build config/ tree:

```
out_dir/
  config/
    includes.installer/preseed.cfg           (d-i automation; %%HASHED_PASSWORD%%)
    hooks/live/9000-debateos-apply.hook.chroot  (packages + services + sysctl + groups; 0755)
    package-lists/debateos.list.chroot_install  (target packages; .chroot_install = both envs)
    archives/<repo>.list.chroot_install          (custom apt repos with signed-by/trusted=yes)
    archives/<repo>.key.chroot                   (keyring stubs for signed-by repos)
    includes.chroot/etc/systemd/user/            (first-run oneshot units via common.firstrun)
  build-manifest.json
```

- `_sanitize_dst` rejects absolute paths and `../` traversal (T-04-05 mirror of arch T-02-08)
- %%SENTINEL%% replacement only — never `str.format` on raw opinion data (T-04-06)
- Fail-fast: all dst paths validated BEFORE any `os.makedirs` (no partial output on gate failure)

**generator.py** — `generate()` pipeline:
1. Load resolved speech + opinion bodies
2. Load capabilities (Debian set)
3. `check_capabilities()` BEFORE `BuildManifest.from_resolved()` (SC-3 / DEB-01)
4. Build BuildManifest (from common/)
5. Load variant profile
6. `emit_profile_tree()`

`CapabilityError` fires before file I/O → empty out_dir on failure (tested).

**translate** — argv-stable shell wrapper; default `--profile debian`, default `--out ./debian-profile`; PYTHONPATH=REPO_ROOT for module resolution. Mirrors arch/translate exactly.

**manifest.py + contract.py** — shim re-exports from common/ (same pattern as arch); bare-name imports work for tests and generator.

57 new tests GREEN (34 Task 2 + 23 Task 3 additional). Total: 75 Debian tests GREEN.

## Verification

```
python3 -m pytest translators/debian/tests/ -q
# 75 passed in 0.21s

python3 -m pytest translators/arch/tests/ translators/common/tests/ -q
# 171 passed in 0.65s  (no regressions)

go test ./... -count=1
# all packages ok  (no regressions)

./translators/debian/translate translators/debian/tests/fixtures/df_resolved.json \
  --opinions translators/debian/tests/fixtures/df_opinions.json \
  --profile debian --out /tmp/deb-profile
# Profile tree generated at: /tmp/deb-profile
# preseed.cfg: FOUND
# chroot hook: FOUND + executable (0755)
```

## Config/ Tree Layout (for Plan 04-05 structural validation)

```
config/
  includes.installer/
    preseed.cfg                         — d-i automation (not package install)
  hooks/live/
    9000-debateos-apply.hook.chroot     — opinion effectuation at lb chroot time (0755)
  package-lists/
    debateos.list.chroot_install        — target packages (both live + installed)
  archives/
    debateos-<name>.list.chroot_install — custom apt repos
    debateos-<name>.key.chroot          — keyring stubs for signed-by repos
  includes.chroot/
    etc/systemd/user/
      debateos-firstrun-<id>.service    — first-run systemd user units
build-manifest.json                     — full manifest dict (target_packages, file_assets, ...)
```

## Debian Capability Token Set (final, for Plan 04-05)

45 tokens declared. The 5 dual-foundation tokens required for the DF speech proof:

| Token | Source |
|-------|--------|
| install-packages | DF-001 |
| deploy-config-file-tree | DF-002 |
| enable-systemd-service | DF-003 |
| write-sysctl-drop-in | DF-004 |
| add-user-to-group | DF-005 |

Arch-only tokens correctly absent: configure-mkinitcpio-*, manage-limine-*, write-mkinitcpio-*.

## Deviations from Plan

**1. [Rule 2 - Missing shims] Added manifest.py + contract.py re-export shims**
- **Found during:** Task 3 RED — tests imported `from manifest import BuildManifest` and got stdlib's profile module instead
- **Fix:** Created `translators/debian/manifest.py` and `contract.py` as re-export shims from `common/`, exactly mirroring `translators/arch/manifest.py` and `contract.py`
- **Files modified:** translators/debian/manifest.py, translators/debian/contract.py
- **Commit:** 456a73a (included in Task 3 implementation commit)

**2. [Rule 1 - Behaviour clarification] apply_variant signature: no base_repos arg**
- **Found during:** Task 2 GREEN — plan referred to `apply_variant(variant, base_pacman_repos)` but Debian has no fixed base repo list
- **Fix:** `apply_variant(variant)` takes only the variant dict. Custom repos come from opinions' `custom_repos` field via the manifest, not from a hardcoded base set. Tests updated accordingly.
- **Reason:** Debian uses standard mirrors implicitly; the translator never rewrites them (unlike Arch where pacman.conf must list `[core]` and `[extra]` explicitly).

## Known Stubs

- `preseed.cfg` contains `%%USERNAME%%`, `%%USER_FULLNAME%%`, `%%HASHED_PASSWORD%%`, `%%PKGSEL_PACKAGES%%` sentinels — these are intentional build-time placeholders, not runtime stubs. Plan 04-05 documents how to replace them before lb build.
- `config/archives/<repo>.key.chroot` files are stub stubs containing the keyring URL, not the actual GPG key — actual key fetch happens at lb build time inside Docker (Plan 04-05).

## Threat Flags

No new security-relevant surface beyond the plan's threat model. All T-04-05 through T-04-09 mitigations applied:
- T-04-05: _sanitize_dst with fail-fast gate implemented and tested
- T-04-06: %%SENTINEL%% replacement only (no str.format with opinion data)
- T-04-07: OptionalTrustAll/Never → [trusted=yes] + LOUD WARNING comment; Required → signed-by
- T-04-08: preseed.cfg uses %%HASHED_PASSWORD%% sentinel, never plaintext
- T-04-09: accepted (private assets not in config/ tree — same policy as Arch)

## Self-Check: PASSED

Files exist:
- translators/debian/translate: FOUND
- translators/debian/capabilities.json: FOUND
- translators/debian/capabilities.py: FOUND
- translators/debian/variant.py: FOUND
- translators/debian/profile.py: FOUND
- translators/debian/generator.py: FOUND
- translators/debian/profiles/debian.yaml: FOUND
- translators/debian/templates/preseed.cfg.tpl: FOUND
- translators/debian/templates/chroot-install.hook.tpl: FOUND
- translators/debian/tests/test_capability_gate.py: FOUND
- translators/debian/tests/test_variant.py: FOUND
- translators/debian/tests/test_profile.py: FOUND
- translators/debian/tests/test_generator.py: FOUND
- .planning/phases/04-debian-translator/04-03-SUMMARY.md: FOUND

Commits verified:
- a33d8ac: test(04-03) — Task 1 RED (7 files)
- 72de11f: feat(04-03) — Task 2 GREEN capabilities + variant + debian.yaml (4 files)
- 456a73a: feat(04-03) — Task 3 GREEN profile + generator + translate (9 files)
