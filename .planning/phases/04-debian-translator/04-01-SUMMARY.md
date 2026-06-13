---
phase: "04-debian-translator"
plan: "01"
subsystem: "translators/common"
tags: [tdd, extraction, shared-core, foundation-neutral, regression-gate]
dependency_graph:
  requires: []
  provides:
    - translators/common (contract, manifest, firstrun)
  affects:
    - translators/arch (re-exports from common)
    - translators/debian (will import from common)
tech_stack:
  added:
    - "translators/common Python package (stdlib + PyYAML, no new deps)"
  patterns:
    - "Shim re-export: translator-local modules delegate to common"
    - "Capability gate moved to caller: foundation-neutral manifest"
    - "Parameterized template_dir in firstrun for future translator divergence"
key_files:
  created:
    - translators/common/__init__.py
    - translators/common/contract.py
    - translators/common/manifest.py
    - translators/common/firstrun.py
    - translators/common/templates/firstrun.service.tpl
    - translators/common/pytest.ini
    - translators/common/tests/__init__.py
    - translators/common/tests/test_common_shared.py
  modified:
    - translators/arch/contract.py (replaced with shim)
    - translators/arch/manifest.py (replaced with shim)
    - translators/arch/firstrun.py (replaced with shim)
    - translators/arch/generator.py (explicit capability gate; sys.path update)
    - translators/arch/pytest.ini (added .. to pythonpath)
    - translators/arch/tests/test_manifest.py (capability gate tests updated)
decisions:
  - "[04-01]: check_capabilities removed from common/manifest.py — caller-responsibility per translator; generator.py calls it before BuildManifest.from_resolved() (SC-3 / ARCH-03 gate preserved)"
  - "[04-01]: Shim re-export pattern used for arch/*.py — bare-name imports continue working with no arch-test changes except capability gate test update"
  - "[04-01]: Parameterized template_dir in render_firstrun_unit — Debian can pass its own template path if it ever diverges; common default used otherwise"
  - "[04-01]: Clean extraction path taken — no 30-min fallback needed; all gates green"
metrics:
  duration: "~20 min"
  completed: "2026-06-13"
  tasks: 3
  files: 13
---

# Phase 4 Plan 01: Extract translators/common Shared Core Summary

**One-liner:** Foundation-neutral shared translator core (contract loaders, BuildManifest, first-run renderer) extracted to translators/common/ with arch re-exporting via shims; all regression gates green.

## What Was Built

Created `translators/common/` as the single source of truth for the three byte-identical/near-identical modules shared between the Arch and Debian translators:

- **common/contract.py**: `load_resolved_speech`, `load_opinion_bodies` — verbatim from arch/contract.py; no Arch-specific code.
- **common/manifest.py**: `BuildManifest`, `derive_source_date_epoch` — refactored to remove the `check_capabilities` import (foundation-neutral; see Deviations).
- **common/firstrun.py**: `render_firstrun_unit`, `firstrun_unit_name` — parameterized `template_dir` parameter; defaults to `translators/common/templates/`.
- **common/templates/firstrun.service.tpl**: Single source of truth for the unit template (identical to former arch template).
- **36 tests** in `translators/common/tests/test_common_shared.py` covering all public surface with TDD RED/GREEN cycle.

Arch translator re-points via shim files: `arch/contract.py`, `arch/manifest.py`, `arch/firstrun.py` each contain a re-export shim (`from common.X import ...`) that makes bare-name imports (`from contract import load_opinion_bodies`) continue to resolve to the common implementation.

## TDD Gate Compliance

| Gate | Commit | Status |
|------|--------|--------|
| RED — `test(04-01): RED shared common-module tests` | 1eca01f | 36 tests failing (import errors; modules did not exist) |
| GREEN — `feat(04-01): extract translators/common shared core; arch re-exports it` | 5db62b6 | 36 common tests + 135 arch tests passing |
| Regression — `test(04-01): regression gate green after common extraction` | f623bca | All gates green (see below) |

## Regression Gate Results

| Gate | Command | Result |
|------|---------|--------|
| Python common suite | `python3 -m pytest translators/common/tests/ -q` | 36 passed |
| Python arch suite | `python3 -m pytest translators/arch/tests/ -q` | 135 passed |
| Go suite | `go test ./... -count=1` | all packages GREEN |
| North-star (--skip-build) | `bash scripts/arch-northstar-check.sh --skip-build` | 16/16 PASSED |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Updated arch/tests/test_manifest.py capability gate tests**

- **Found during:** Task 2 GREEN
- **Issue:** Two tests (`test_from_resolved_raises_on_unsupported_required_capability`, `test_from_resolved_capability_gate_runs_before_assembly`) called `BuildManifest.from_resolved()` directly and expected `CapabilityError`. After moving the gate to the caller (generator.py), these tests failed — `from_resolved` no longer calls `check_capabilities`.
- **Fix:** Updated both tests to call `check_capabilities()` directly (the new correct call site). Renamed tests to `test_capability_gate_raises_on_unsupported_required_capability` and `test_capability_gate_runs_before_assembly` to reflect the new design. The behavior (gate fires before assembly) is identical; the call site moved.
- **Files modified:** `translators/arch/tests/test_manifest.py`
- **Commit:** 5db62b6

**2. [Rule 1 - Bug] Fixed test_from_resolved_no_check_capabilities_import test**

- **Found during:** Task 2 GREEN (first common test run)
- **Issue:** The test used `inspect.getsource()` to check the source string for `"from capabilities import"`. This matched the docstring in common/manifest.py that says "removed the `from capabilities import check_capabilities` import", causing a false positive.
- **Fix:** Changed the test to check `hasattr(cm, "check_capabilities")` on the loaded module instead of inspecting source text — more precise and immune to documentation text.
- **Files modified:** `translators/common/tests/test_common_shared.py`
- **Commit:** 5db62b6

### Design Note: Capability Gate Architecture

The plan stated: "Remove the `check_capabilities` import and the in-method `check_capabilities(...)` call from the common copy; ... Then update arch/generator.py to call check_capabilities() explicitly before BuildManifest.from_resolved."

This was implemented exactly. The two arch tests that previously verified the gate fires from `from_resolved()` had to be updated because the gate no longer lives in `from_resolved()` — it's now in the caller. This is the intended design change (and was expected by the plan).

## Fallback Decision

**Clean extraction path taken.** The shim approach worked cleanly:
- `translators/arch/pytest.ini` got `..` added to `pythonpath` so `common` is importable in arch test context.
- `translators/arch/generator.py` inserts `_TRANSLATORS_DIR` (parent of arch/) into `sys.path` so `import common` resolves at runtime.
- No import conflicts; no 30-min bar exceeded.
- The 30-min fallback (documented duplication with `# SHARED:` comments) was NOT needed.

## Known Stubs

None. All common modules are fully implemented with real logic. No hardcoded empty values, placeholder text, or unwired data sources.

## Threat Flags

| Flag | File | Description |
|------|------|-------------|
| T-04-01 mitigated | common/manifest.py | `sig_level=Never/OptionalTrustAll` trust_warning emission verified by 2 tests in `test_common_shared.py` |

No new attack surface beyond what already existed in arch/manifest.py. The extraction is a pure refactor — the same logic now runs from common/.

## Self-Check: PASSED

All created files exist on disk. All task commits present in git log:
- 1eca01f test(04-01): RED shared common-module tests
- 5db62b6 feat(04-01): extract translators/common shared core; arch re-exports it
- f623bca test(04-01): regression gate green after common extraction
