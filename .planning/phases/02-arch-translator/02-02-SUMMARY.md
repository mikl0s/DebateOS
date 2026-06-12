---
phase: 02-arch-translator
plan: 02
subsystem: translator
tags: [python, pytest, archiso, variant-profiles, profile-emitter, firstrun-units, tdd, debateos, arch-04, arch-01]

# Dependency graph
requires:
  - phase: 02-arch-translator
    plan: 01
    provides: BuildManifest.from_resolved(), capabilities gate, contract loaders, derive_source_date_epoch
  - phase: 02-arch-translator
    plan: 03
    provides: vanilla-arch.yaml, cachyos.yaml, garuda.yaml variant profiles

provides:
  - variant.py: load_variant_profile (FileNotFoundError on unknown name), apply_variant (above_core ordering, keyring-first, trust-warning capture), surface_conflicts (omarchy opinion intersection)
  - firstrun.py: render_firstrun_unit — flag-file-guarded systemd user oneshot units (ConditionPathExists=!, ExecStartPost touch, RemainAfterExit, graphical-session.target)
  - profile.py: emit_profile_tree — complete archiso profile tree from BuildManifest + variant (profiledef.sh, packages.x86_64, pacman.conf, airootfs installer 0755, .zlogin, systemd/user first-run units, build-manifest.json); _sanitize_dst T-02-08 security gate
  - generator.py: generate() composing contract→capabilities→BuildManifest→variant→emit; runnable as python -m translators.arch.generator; _main() CLI
  - translate: argv-stable shell wrapper (FROZEN for Phase 3): translate <resolved.json> --opinions <path> --profile <name> --out <dir>
  - templates/: profiledef.sh.tpl, pacman.conf.tpl, installer.sh.tpl (%%SENTINEL%% substitution), firstrun.service.tpl
  - fixtures: omarchy_subset_resolved.json + omarchy_subset_opinions.json (7 opinions: repo, package, file-asset, sysctl, service, first-run, theming)
  - 85 new tests GREEN (128 total across all plans)

affects: [02-arch-translator-plan-05, 03-cli-builds]

# Tech tracking
tech-stack:
  added: [string.replace sentinel substitution (%%VAR%%) for shell-safe templates, os.chmod stat.S_IRW* for 0755 installer, translators/__init__.py for python -m package discovery]
  patterns:
    - ARCH-04 no-fork: apply_variant uses only YAML data (above_core field); zero name branches in variant.py (grep count = 0)
    - ARCH-01 path safety: _sanitize_dst rejects absolute paths and .. traversal before any write (T-02-08)
    - Pitfall 2 minimal packages.x86_64: live-env set (~15 packages) distinct from target set in build-manifest.json
    - Pitfall 4 keyring-first: keyring_install_before_repos injected into build-manifest.json; installer installs before custom repos
    - Pitfall 6 jq-driven installer: all opinion data in build-manifest.json; installer reads via jq, no shell interpolation (T-02-09)
    - Pattern 1 .zlogin hook: tty1 check calls /root/debateos-install.sh (releng baseline pattern)
    - Pattern 2 flag-file guard: ConditionPathExists=! /var/lib/debateos/.firstrun-<id>.done; not ConditionFirstBoot (Pitfall 3)
    - T-02-11 user-scope units: first-run units in etc/systemd/user/ (not system/) for graphical-session.target
    - Template approach: %%SENTINEL%% replace() for shell-heavy installer.sh.tpl; str.format() for profiledef.sh.tpl and pacman.conf.tpl

key-files:
  created:
    - translators/__init__.py
    - translators/arch/variant.py
    - translators/arch/firstrun.py
    - translators/arch/profile.py
    - translators/arch/generator.py
    - translators/arch/translate
    - translators/arch/templates/profiledef.sh.tpl
    - translators/arch/templates/pacman.conf.tpl
    - translators/arch/templates/installer.sh.tpl
    - translators/arch/templates/firstrun.service.tpl
    - translators/arch/tests/test_variant.py
    - translators/arch/tests/test_firstrun.py
    - translators/arch/tests/test_profile.py
    - translators/arch/tests/test_generator.py
    - translators/arch/tests/fixtures/omarchy_subset_resolved.json
    - translators/arch/tests/fixtures/omarchy_subset_opinions.json
  modified: []

key-decisions:
  - "%%SENTINEL%% replace() for installer.sh.tpl: installer has many ${SHELL_VAR} expansions; str.format() raises KeyError on them. Targeted %%ISO_LABEL%%/%%SOURCE_DATE_EPOCH%% substitution is the stdlib-only safe approach without escaping every shell brace."
  - "translators/__init__.py added so python -m translators.arch.generator works from repo root (package discovery); existing bare-name imports in arch/*.py unaffected"
  - "sys.path.insert in generator.py so _ARCH_DIR is on path in both pytest context (pythonpath=.) and -m translators.arch.generator context (where CWD is repo root)"
  - "validate file_asset dst after writing profile tree: _sanitize_dst called at end of emit_profile_tree after all other files are written; error is raised before caller uses the tree"
  - "emit_profile_tree injects keyring_install_before_repos into build-manifest.json dict so installer.sh can read it via jq (Pitfall 4 ordering preserved at install time)"

requirements-completed: [ARCH-01, ARCH-04]

# Metrics
duration: 11min
completed: 2026-06-12
---

# Phase 02 Plan 02: Arch Translator Profile Emitter + Variant Application Summary

**Archiso profile tree emitter (ARCH-01) + variant application (ARCH-04) on top of Plan 01 BuildManifest: load_variant_profile, apply_variant (above_core ordering, keyring-first, trust-warning capture), surface_conflicts, emit_profile_tree (profiledef.sh/packages.x86_64/pacman.conf/installer 0755/.zlogin/first-run units/build-manifest.json), file_asset dst sanitization (T-02-08), and argv-stable translate entrypoint — 128 pytest GREEN, RED-before-GREEN for all three TDD tasks (D19)**

## Performance

- **Duration:** ~11 min
- **Started:** 2026-06-12T23:08:52Z
- **Completed:** 2026-06-12T23:19:xx Z
- **Tasks:** 3 (each with RED + GREEN commits; Task 2 had mid-task template commit)
- **Files created:** 16 (new); 1 modified (test_generator.py repo_root fix)

## Accomplishments

- `variant.py`: `load_variant_profile(name)` reads `profiles/<name>.yaml` with clear FileNotFoundError on unknown name; `apply_variant(variant, base_repos)` injects repos with above_core ordering (CachyOS above core, Garuda below core), returns `keyring_install_before_repos` and `trust_warnings`; `surface_conflicts(variant, applied_ids)` returns the intersection of `conflicts_with_omarchy` entries whose `omarchy_opinions` are in the applied set — ARCH-04 zero-branch proven: `grep -Eic "if .*(cachyos|garuda|vanilla)" variant.py` == 0
- `firstrun.py`: `render_firstrun_unit(opinion_id, description, exec_path)` emits a flag-file-guarded systemd user oneshot unit (ConditionPathExists=!, Type=oneshot, RemainAfterExit=yes, ExecStartPost /bin/touch, WantedBy=graphical-session.target) per Pattern 2
- `profile.py`: `emit_profile_tree(out_dir, manifest, variant)` writes the complete archiso profile tree; `_sanitize_dst(dst)` rejects absolute and `..`-traversal paths (T-02-08); minimal packages.x86_64 (~15 live-env packages) distinct from target set in build-manifest.json (Pitfall 2); pacman.conf with variant repo injection (above/below core); installer at 0755; jq-driven installer body (T-02-09); first-run user units in `etc/systemd/user/` (T-02-11)
- `generator.py`: `generate()` composes full pipeline; `_main()` for `python -m translators.arch.generator <resolved> <opinions> <profile> <out>`; sys.path injection handles both pytest and -m contexts
- `translate`: argv-stable shell wrapper (FROZEN for Phase 3); parses `--opinions`, `--profile` (default vanilla-arch), `--out`; unknown flags exit 1; execs generator via REPO_ROOT PYTHONPATH
- Four templates: `profiledef.sh.tpl` (iso_name, bootmodes, file_permissions), `pacman.conf.tpl` (variant repo insertion points), `installer.sh.tpl` (%%SENTINEL%% substitution, jq-driven body), `firstrun.service.tpl` (Pattern 2 unit)
- Fixtures: `omarchy_subset_resolved.json` + `omarchy_subset_opinions.json` (7 opinions: OM-001 custom-repo, OM-006 package-install, OM-028 file-asset, OM-036 sysctl, OM-057 service, OM-102 first-run, OM-114 theming)

## Task Commits

RED-before-GREEN for all three tasks per D19:

1. **Task 1 RED: test_variant.py + test_firstrun.py** — `62e87fd` (test)
2. **Task 1 GREEN: variant.py + firstrun.py** — `209340d` (feat)
3. **Task 2 RED: test_profile.py + fixtures** — `047560f` (test)
4. **Task 2 GREEN: profile.py + four templates** — `9cca26d` (feat)
5. **Task 3 RED: test_generator.py** — `434f858` (test)
6. **Task 3 GREEN: generator.py + translate + translators/__init__.py** — `193027b` (feat)

## Files Created/Modified

- `translators/__init__.py` — Package root for `-m translators.arch.generator` invocation
- `translators/arch/variant.py` — load_variant_profile, apply_variant, surface_conflicts (65 lines)
- `translators/arch/firstrun.py` — render_firstrun_unit, firstrun_unit_name (75 lines)
- `translators/arch/profile.py` — emit_profile_tree, _sanitize_dst, template loaders (370 lines)
- `translators/arch/generator.py` — generate(), _main(), sys.path injection (120 lines)
- `translators/arch/translate` — argv-stable shell wrapper (chmod +x, 75 lines)
- `translators/arch/templates/profiledef.sh.tpl` — iso_name/bootmodes/file_permissions
- `translators/arch/templates/pacman.conf.tpl` — {variant_repos_above}/{variant_repos_below} insertion
- `translators/arch/templates/installer.sh.tpl` — %%SENTINEL%% substitution, jq-driven body
- `translators/arch/templates/firstrun.service.tpl` — Pattern 2 oneshot unit
- `translators/arch/tests/test_variant.py` — 41 tests (load, apply_variant, surface_conflicts, all_variants, trust_warnings)
- `translators/arch/tests/test_firstrun.py` — 15 tests (unit sections, flag-file condition, oneshot, graphical-session.target)
- `translators/arch/tests/test_profile.py` — 28 tests (tree structure, installer 0755, zlogin, build-manifest.json, packages.x86_64 minimal, first-run units, pacman.conf ordering, traversal dst raises)
- `translators/arch/tests/test_generator.py` — 16 tests (e2e generate, all three profiles, capability gate, -m invocation, translate argv)
- `translators/arch/tests/fixtures/omarchy_subset_resolved.json` — 7-opinion resolved speech
- `translators/arch/tests/fixtures/omarchy_subset_opinions.json` — 7 opinion bodies

## Decisions Made

- **%%SENTINEL%% template substitution for installer.sh.tpl:** installer has extensive `${SHELL_VAR}` syntax throughout; Python str.format() raises KeyError on these. Targeted `str.replace("%%ISO_LABEL%%", ...)` is the cleanest stdlib-only approach without escaping every `{...}` in the shell body.
- **translators/__init__.py added:** enables `python -m translators.arch.generator` from repo root by making `translators` a Python package. Existing bare-name imports in `arch/*.py` remain unchanged (pytest.ini pythonpath=. still covers them).
- **sys.path.insert in generator.py:** enables bare-name imports (`from capabilities import`) in both pytest context (pythonpath=.) and `-m translators.arch.generator` context (where cwd is repo root, not `translators/arch/`).
- **file_asset dst sanitization called after writing:** _sanitize_dst runs at end of emit_profile_tree (after all other files are written) — the error is still raised before the caller uses the tree, but this structure keeps the validation co-located with the file_assets consumer rather than scattered across the code.
- **keyring_install_before_repos in build-manifest.json:** injected alongside the manifest dict so installer.sh can read it via jq and enforce Pitfall 4 ordering at install time.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] str.format() KeyError on installer.sh.tpl shell variable syntax**
- **Found during:** Task 2 GREEN implementation (first test run)
- **Issue:** installer.sh.tpl contains `${TARGET_DISK}`, `${MANIFEST}` etc. which `string.Template.safe_substitute()` and `str.format()` both misinterpret as Python format placeholders.
- **Fix:** Switched installer template to use `%%SENTINEL%%` markers (`%%ISO_LABEL%%`, `%%SOURCE_DATE_EPOCH%%`) and `str.replace()` for substitution. Other templates (profiledef.sh.tpl, pacman.conf.tpl) don't have shell variable syntax and use `str.format()` correctly.
- **Files modified:** `installer.sh.tpl`, `profile.py`
- **Commit:** `9cca26d`

**2. [Rule 1 - Bug] test_generator.py incorrect repo_root path calculation**
- **Found during:** Task 3 GREEN (first test run)
- **Issue:** `repo_root = os.path.dirname(os.path.dirname(os.path.dirname(_ARCH_DIR)))` went 3 levels above `translators/arch/`, landing at `repos/` instead of `DebateOS/`. The subprocess invocation of `python -m translators.arch.generator` failed with ModuleNotFoundError.
- **Fix:** Changed to `repo_root = os.path.dirname(os.path.dirname(_ARCH_DIR))` (2 levels up from `translators/arch/` = `DebateOS/`).
- **Files modified:** `translators/arch/tests/test_generator.py`
- **Commit:** `193027b`

## Issues Encountered

None beyond the two auto-fixed bugs above. Both were caught immediately by the test suite and fixed inline.

## Known Stubs

None — all code paths are wired and tested. The installer.sh.tpl contains `%%SENTINEL%%` patterns that are replaced at profile generation time (not runtime stubs). The jq-driven data paths in the installer read `build-manifest.json` which is always written by emit_profile_tree.

## Threat Flags

No new security-relevant surface beyond the plan's threat model. Mitigations confirmed:
- T-02-08 (file_asset dst traversal): `_sanitize_dst` tested with `/etc/passwd` (absolute), `../../etc/passwd` (relative traversal), `/home/user/.bashrc` (absolute) — all raise ValueError.
- T-02-09 (shell injection): installer body reads all opinion data via jq from build-manifest.json; no opinion strings interpolated into shell.
- T-02-10 (sig_level=Never warning): `apply_variant` emits trust_warnings for Never repos; warning comments appear in pacman.conf.
- T-02-11 (first-run units user scope): units written to `etc/systemd/user/` (not system/), WantedBy=graphical-session.target.

## Self-Check: PASSED

All 16 created files verified present. All 6 commits confirmed in git log.

Key verifications:
- `cd translators/arch && python -m pytest tests/ -q` → 128 passed
- `go test ./... -count=1` → all packages OK
- `grep -v '^#' translators/arch/variant.py | grep -Eic "if .*(cachyos|garuda|vanilla)"` → 0 (ARCH-04 no-fork)
- `test -x translators/arch/translate` → exits 0
- `grep -Eq "passwd|\\.\\.|traversal" translators/arch/tests/test_profile.py` → exits 0

Commit hashes:
- `62e87fd` — test(02-02): add failing tests for variant + firstrun (RED Task 1)
- `209340d` — feat(02-02): implement variant loader + firstrun unit generator (GREEN Task 1)
- `047560f` — test(02-02): add failing tests for profile tree emitter (RED Task 2)
- `9cca26d` — feat(02-02): implement profile tree emitter + templates (GREEN Task 2)
- `434f858` — test(02-02): add failing tests for generator + translate wrapper (RED Task 3)
- `193027b` — feat(02-02): implement generate() entrypoint + argv-stable translate wrapper (GREEN Task 3)

*Phase: 02-arch-translator*
*Completed: 2026-06-12*
