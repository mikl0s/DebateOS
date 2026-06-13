---
phase: 04-debian-translator
plan: "04"
subsystem: cli
tags: [go, build, foundationRegistry, dispatch, deb-03, arch-leak-audit]

# Dependency graph
requires:
  - phase: 04-debian-translator
    provides: translators/debian/translate entrypoint (04-03) and ResolvedSpeech.Foundation field (01-04)
provides:
  - "foundationRegistry map in cli/build/build.go: arch + debian dispatch, extensible to future foundations"
  - "docs/arch-leak-audit.md: complete DEB-03 audit, 6 findings, 1 genuine leak fixed"
  - "Four Go tests covering foundation dispatch: debian path, arch unchanged, explicit override, unknown foundation error"
affects:
  - 04-05 (dual-foundation-check.sh invokes both translators via the now-wired build paths)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "foundationRegistry pattern: data-driven map[string]foundationConfig replacing hardcoded translator paths"
    - "--profile '' default resolved from registry.DefaultProfile (foundation-specific, backward-compatible)"
    - "Closed-set registry guard: unknown foundation → stderr + exit 1, never falls through to wrong translator (T-04-10)"

key-files:
  created:
    - docs/arch-leak-audit.md
  modified:
    - cli/build/build.go
    - cli/build/build_test.go

key-decisions:
  - "[04-04]: foundationRegistry is a compile-time constant map — TranslateBin values are never derived from speech data (T-04-11)"
  - "[04-04]: --profile default changed to '' resolving from registry; explicit --profile overrides backward-compatibly (Pitfall 6 resolution)"
  - "[04-04]: build.go (Finding 5) is the only genuine DEB-03 leak; no schema field required modification"
  - "[04-04]: foundationRegistry ubuntu entry commented as forward hint but not activated for v1.0"

patterns-established:
  - "foundationRegistry pattern: add one entry + translator dir to support a new foundation; no if/else chains"

requirements-completed: [DEB-01, DEB-03]

# Metrics
duration: 18min
completed: 2026-06-13
---

# Phase 4 Plan 04: Foundation-Aware CLI Build Refactor + DEB-03 Arch-Leak Audit Summary

**`foundationRegistry` in `cli/build/build.go` dispatches `translators/arch/translate` or `translators/debian/translate` by `rs.Foundation`, plus complete DEB-03 6-finding audit confirming `build.go` is the only genuine leak**

## Performance

- **Duration:** 18 min
- **Started:** 2026-06-13T14:30:00Z
- **Completed:** 2026-06-13T14:48:00Z
- **Tasks:** 2 (1 TDD + 1 auto)
- **Files modified:** 3 (build.go, build_test.go, docs/arch-leak-audit.md)

## Accomplishments

- Replaced hardcoded `translators/arch/translate`, `arch-profile`, and `vanilla-arch` defaults in `build.go` with a data-driven `foundationRegistry` map keyed on `rs.Foundation`
- Wired `translators/debian/translate` path via registry; arch path unchanged (backward compatible)
- Added TDD tests: `TestBuildFoundationDispatch`, `TestBuildArchUnchanged`, `TestBuildExplicitProfileOverride`, `TestBuildUnknownFoundation` — all pass; full `go test ./... -count=1` green
- Authored `docs/arch-leak-audit.md`: all 6 DEB-03 findings with type and action, confirming `build.go` (Finding 5) is the only genuine leak

## Task Commits

Each task was committed atomically:

1. **Task 1 RED: foundation dispatch tests** - `babb554` (test)
2. **Task 1 GREEN: foundationRegistry in build.go** - `a0dc4c6` (feat)
3. **Task 2: docs/arch-leak-audit.md** - `dd3b340` (docs)

_Note: TDD task has separate RED and GREEN commits per TDD protocol_

## Files Created/Modified

- `/home/mikkel/repos/DebateOS/cli/build/build.go` — Added `foundationConfig` struct, `foundationRegistry` map, registry lookup after `ResolveDir`, `--profile ""` default with foundation resolution
- `/home/mikkel/repos/DebateOS/cli/build/build_test.go` — Added `minimalSpeechDirFoundation` helper + 4 new dispatch tests
- `/home/mikkel/repos/DebateOS/docs/arch-leak-audit.md` — DEB-03 complete audit: 6 findings, dispositions, cross-references

## foundationRegistry Values (for 04-05 reference)

```go
var foundationRegistry = map[string]foundationConfig{
    "arch":   {"translators/arch/translate",   "arch-profile",   "vanilla-arch"},
    "debian": {"translators/debian/translate", "debian-profile", "debian"},
    // Future: "ubuntu": {"translators/debian/translate", "debian-profile", "ubuntu"},
}
```

`dual-foundation-check.sh` (Plan 04-05) can invoke both translators consistently via:
- Arch: `translators/arch/translate <resolved.json> --opinions <dir> --profile vanilla-arch --out <out>/arch-profile`
- Debian: `translators/debian/translate <resolved.json> --opinions <dir> --profile debian --out <out>/debian-profile`

## DEB-03 Audit Summary

| Finding | Type | Action |
|---------|------|--------|
| `sig_level` enum | Intentional abstraction | apt mapping in `variant.py`; no schema change |
| `install_phase` enum | Foundation-neutral | No change |
| mkinitcpio capability tokens | Correctly isolated | Debian `capabilities.json` omits them |
| limine bootloader tokens | Correctly isolated | Debian uses GRUB2; abstract `bootloader` schema field |
| `build.go` hardcoded to Arch | **Genuine leak — FIXED** | `foundationRegistry` refactor (this plan) |
| `keyring` field interpretation | Minor asymmetry | Documented as translator-interpreted; no v1.0 change |

**DEB-03 conclusion:** Only `build.go` required a code change. No schema field was modified.

## Decisions Made

- `[04-04]`: `foundationRegistry` uses compile-time constants for `TranslateBin` — no injection surface from speech data (T-04-11)
- `[04-04]`: `--profile` default `""` resolves from `foundationRegistry[rs.Foundation].DefaultProfile` — backward-compatible for callers who pass `--profile vanilla-arch` explicitly
- `[04-04]`: Ubuntu entry commented as forward hint in registry; not activated (post-v1.0)
- `[04-04]`: `build.go` is the only genuine DEB-03 leak; no schema field changed

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered

None. The RED phase tests failed as expected (arch hardcoding confirmed). GREEN implementation straightforward. Full `go test ./... -count=1` green after both tasks.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- Plan 04-05 (dual-foundation-check.sh) can now invoke both translators via registry-consistent argv
- `debateos build` is foundation-agnostic; Debian speeches dispatch correctly
- All DEB-03 findings documented; invariant 1 proven through audit

## Self-Check: PASSED

- FOUND: cli/build/build.go (foundationRegistry present, debian dispatch wired)
- FOUND: cli/build/build_test.go (4 new dispatch tests)
- FOUND: docs/arch-leak-audit.md (6 findings, build.go confirmed only genuine leak)
- FOUND: .planning/phases/04-debian-translator/04-04-SUMMARY.md
- FOUND commit babb554 (RED: foundation dispatch tests)
- FOUND commit a0dc4c6 (GREEN: foundationRegistry in build.go)
- FOUND commit dd3b340 (docs: arch-leak-audit.md)
- FOUND commit d7da436 (docs: plan metadata)
- go test ./... -count=1: ALL PASS (no regressions)

---
*Phase: 04-debian-translator*
*Completed: 2026-06-13*
