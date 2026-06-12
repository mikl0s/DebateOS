---
phase: 02-arch-translator
plan: 01
subsystem: translator
tags: [python, pytest, archiso, capability-gate, build-manifest, tdd, debateos]

# Dependency graph
requires:
  - phase: 01-schema-resolver-core
    provides: ResolvedSpeech canonical JSON contract (applied/skipped/dropped/install_order/explanations); Opinion struct with all payload fields (translator_capabilities, packages, services, etc.)

provides:
  - translators/arch/ Python package skeleton with pytest.ini and requirements-dev.txt
  - capabilities.json: 29 declared Arch translator capability tokens
  - capabilities.py: CapabilityError + load_capabilities() + check_capabilities() (SC-3 / ARCH-03 gate)
  - contract.py: load_resolved_speech() + load_opinion_bodies() (JSON file or YAML directory)
  - manifest.py: BuildManifest @dataclass + from_resolved() + derive_source_date_epoch() + to_dict()
  - Full pytest test suite: 43 tests GREEN (14 capability gate + 29 manifest)

affects: [02-arch-translator-plan-02, 02-arch-translator-plan-03, 02-arch-translator-plan-05, 03-cli-builds]

# Tech tracking
tech-stack:
  added: [pytest>=9.0.3 (installed via pip --break-system-packages), PyYAML>=6.0, hashlib+struct (stdlib SHA-256 epoch derivation)]
  patterns:
    - SC-3 capability gate fires before any assembly (check_capabilities raises CapabilityError naming opinion + token + "composition time")
    - TDD RED/GREEN commit pair per task (D19)
    - BuildManifest.from_resolved() aggregates payloads in install_order; deduplicates packages by first-occurrence
    - derive_source_date_epoch: SHA-256 first 4 bytes mod [MIN,MAX) range for deterministic epoch (BLD-03 groundwork)
    - to_dict() for JSON-serializable build-manifest.json (T-02-01 data-as-JSON, Pitfall 6 mitigation)
    - trust_warnings list captures sig_level=Never repos before pacman.conf emission (T-02-02)

key-files:
  created:
    - translators/arch/__init__.py
    - translators/arch/pytest.ini
    - translators/arch/requirements-dev.txt
    - translators/arch/capabilities.json
    - translators/arch/capabilities.py
    - translators/arch/contract.py
    - translators/arch/manifest.py
    - translators/arch/tests/__init__.py
    - translators/arch/tests/test_capability_gate.py
    - translators/arch/tests/test_manifest.py
    - translators/arch/tests/fixtures/minimal_resolved.json
    - translators/arch/tests/fixtures/minimal_opinions.json
    - translators/arch/tests/fixtures/unsupported_required_resolved.json
    - translators/arch/tests/fixtures/unsupported_required_opinions.json
  modified: []

key-decisions:
  - "pytest installed via pip --break-system-packages (host Debian restrictions); documented in SUMMARY; version 9.0.3 matches Arch official python-pytest 1:9.0.3-1"
  - "install-npm-global-packages intentionally absent from capabilities.json so the gate has a token to exercise on nice-to-have drop tests (can be added when implemented)"
  - "first_run opinions do NOT contribute install-time packages/services; they are collected as {id, script_payload} for systemd oneshot unit generation in Plan 02"
  - "check_capabilities() returns list of (opinion_id, reason) tuples for dropped nice-to-haves; empty list on clean pass"
  - "load_opinion_bodies accepts JSON array file OR directory of *.yaml/*.json opinion files for composability with examples/omarchy/"

patterns-established:
  - "Pattern: capability-gate-first — check_capabilities() must be called before any manifest assembly begins (ARCH-03 invariant enforced in from_resolved)"
  - "Pattern: install-order-authoritative — target_packages list follows resolved.install_order; packages from later entries come after earlier; dedup preserves first occurrence"
  - "Pattern: execution_phase-split — opinions with execution_phase=='first-run' are collected into manifest.first_run and excluded from install-time package/service aggregation"
  - "Pattern: trust-warning-capture — custom_repos with sig_level='Never' emit human-readable trust_warnings for surfacing in pacman.conf comments (Plan 02)"
  - "Pattern: data-as-JSON — to_dict() produces fully serializable output; installer reads build-manifest.json, never shell-interpolates raw strings (T-02-01)"

requirements-completed: [ARCH-01, ARCH-03]

# Metrics
duration: 6min
completed: 2026-06-12
---

# Phase 02 Plan 01: Arch Translator Data Layer Summary

**Python package skeleton with SC-3 capability gate (CapabilityError + 29-token capabilities.json), contract loaders (ResolvedSpeech JSON + opinion bodies), and BuildManifest dataclass aggregating the full payload in install_order with deterministic SOURCE_DATE_EPOCH — 43 pytest GREEN, RED-before-GREEN commits for both TDD tasks (D19)**

## Performance

- **Duration:** ~6 min
- **Started:** 2026-06-12T22:28:16Z
- **Completed:** 2026-06-12T22:34:xx Z
- **Tasks:** 2 (each with RED + GREEN commits)
- **Files created:** 14

## Accomplishments

- `capabilities.json` declares 29 Arch translator capability tokens; `check_capabilities()` enforces ARCH-03 gate: required opinion + unsupported token raises `CapabilityError` naming opinion + token + "composition time"; nice-to-have opinions return `(id, reason)` drop tuples
- `contract.py` provides `load_resolved_speech()` returning the full ResolvedSpeech dict and `load_opinion_bodies()` accepting JSON array file or directory of YAML/JSON opinion files — both stdlib-only
- `manifest.py` provides `BuildManifest.from_resolved()` aggregating all 10+ payload fields (packages, services, first_run, sysctl, kernel, groups, mime, themes, repos, trust_warnings) in install_order; `derive_source_date_epoch()` for BLD-03 determinism groundwork; `to_dict()` for JSON serialization (Pitfall 6 mitigation)
- Full TDD: 14 RED → GREEN for capability gate (Task 1), 29 RED → GREEN for BuildManifest (Task 2); `go test ./...` regression still green

## Task Commits

Each task was committed atomically per D19 RED-before-GREEN discipline:

1. **Task 1 RED: Failing capability gate tests** - `d21db85` (test)
2. **Task 1 GREEN: capabilities.json + capabilities.py + contract.py** - `1bb3c6d` (feat)
3. **Task 2 RED: Failing BuildManifest tests** - `866f859` (test)
4. **Task 2 GREEN: manifest.py** - `347aa0c` (feat)

_Note: TDD tasks have two commits each (test commit → feat commit)_

## Files Created/Modified

- `translators/arch/__init__.py` — Package root
- `translators/arch/pytest.ini` — `testpaths = tests`, `pythonpath = .`
- `translators/arch/requirements-dev.txt` — `pytest>=9.0.3`, `PyYAML>=6.0`
- `translators/arch/capabilities.json` — 29 declared capability tokens
- `translators/arch/capabilities.py` — `CapabilityError`, `load_capabilities()`, `check_capabilities()` (SC-3 / ARCH-03)
- `translators/arch/contract.py` — `load_resolved_speech()`, `load_opinion_bodies()` (JSON file or YAML directory)
- `translators/arch/manifest.py` — `BuildManifest @dataclass`, `from_resolved()`, `derive_source_date_epoch()`, `to_dict()`
- `translators/arch/tests/__init__.py` — Tests package
- `translators/arch/tests/test_capability_gate.py` — 14 tests: load_capabilities, check_capabilities (raise + drop paths), contract loaders
- `translators/arch/tests/test_manifest.py` — 29 tests: package dedup/order, first_run split, services, trust_warnings, epoch determinism, to_dict serialization
- `translators/arch/tests/fixtures/minimal_resolved.json` — Valid ResolvedSpeech with OM-001, OM-006
- `translators/arch/tests/fixtures/minimal_opinions.json` — OM-001 (custom-repo, required), OM-006 (package-install, required)
- `translators/arch/tests/fixtures/unsupported_required_resolved.json` — ResolvedSpeech with OM-006 + OM-023 (npm, required)
- `translators/arch/tests/fixtures/unsupported_required_opinions.json` — OM-023 with `install-npm-global-packages` (undeclared token)

## Decisions Made

- **pytest host install:** pytest was not on the host (Debian-managed Python); installed via `pip --break-system-packages` (version 9.0.3, matching Arch official); documented here per RESEARCH.md §Environment Availability
- **`install-npm-global-packages` absent from capabilities.json:** Intentional — token left undeclared so the gate has a real "unsupported" token for test fixtures and future nice-to-have drop behavior; add it when the npm-global install handler is implemented in a later plan
- **`first_run` isolation:** opinions with `execution_phase=="first-run"` skip ALL install-time aggregation (no packages, no services) — they produce only `{id, script_payload}` entries for the systemd oneshot unit generator in Plan 02
- **`check_capabilities` return value:** returns `list[(opinion_id, reason)]` for dropped nice-to-haves (empty list on clean pass) rather than raising; makes callers able to log/display drops without try/except

## Deviations from Plan

None — plan executed exactly as written. All files listed in `files_modified` frontmatter were created. Line minimums exceeded (capabilities.py: 123 lines vs min 40; contract.py: 145 vs min 30; manifest.py: 311 vs min 60). TDD RED-before-GREEN strictly followed.

## Issues Encountered

- **pytest not installed on host:** Host uses Debian-managed Python which blocks `pip install --user` without `--break-system-packages`. Used `python3 -m pip install --user pytest --break-system-packages`. Version 9.0.3 installed, matching the research-verified Arch official package. Not a plan deviation — RESEARCH.md §Environment explicitly listed "pytest (host): not installed — fallback: pip install". Documented here.

## User Setup Required

None — no external service configuration required. pytest was installed as part of execution (host toolchain gap).

## Known Stubs

None — all contract loaders, capability gate, and manifest builder are fully wired and tested. No placeholder data or hardcoded empty values in the data flow.

## Threat Flags

No new security-relevant surface introduced beyond what the plan's `<threat_model>` covers:
- T-02-01 (Tampering via opinion payload strings): mitigated — `to_dict()` / JSON serialization enforced in `manifest.py`
- T-02-02 (sig_level=Never repos): mitigated — `trust_warnings` captured in `BuildManifest.from_resolved()`
- T-02-03 (required opinion + undeclared capability silently passing): mitigated — `check_capabilities()` raises before assembly in `from_resolved()`

## Next Phase Readiness

- Plan 02 (profile emitter) can immediately consume `BuildManifest` via `from_resolved()` and `to_dict()`; the data layer contract is stable and tested
- Go resolver regression suite (`go test ./... -count=1`) stays GREEN — no Go changes in this plan
- `examples/omarchy/` opinion YAML files can use `load_opinion_bodies()` directory mode when authored

---
*Phase: 02-arch-translator*
*Completed: 2026-06-12*

## Self-Check: PASSED

All 14 created files verified present. All 4 commits confirmed in git log:
- `d21db85` — test(02-01): add failing tests for capability gate (RED - Task 1)
- `1bb3c6d` — feat(02-01): implement capability gate + contract loaders (GREEN - Task 1)
- `866f859` — test(02-01): add failing tests for BuildManifest builder (RED - Task 2)
- `347aa0c` — feat(02-01): implement BuildManifest builder + deterministic epoch (GREEN - Task 2)

Verification:
- `cd translators/arch && python3 -m pytest tests/ -x -q` → 43 passed
- `go test ./... -count=1` → all packages OK
- `python3 -c "from manifest import derive_source_date_epoch as d; assert d(b'abc')==d(b'abc')"` → exits 0
- `python3 -c "import json; d=json.load(open('translators/arch/capabilities.json')); assert 'install-named-packages' in d['capabilities']"` → exits 0
