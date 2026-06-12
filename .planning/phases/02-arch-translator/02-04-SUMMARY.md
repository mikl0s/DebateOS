---
phase: 02-arch-translator
plan: "04"
subsystem: examples/omarchy
tags: [north-star, content, opinions, points, speech, resolver, cc0]
dependency_graph:
  requires: [01-04 (resolve.Resolve), 01-03 (parse package)]
  provides: [examples/omarchy canonical speech, ARCH-02 clean-resolution proof, translator opinion-body source]
  affects: [02-05 (north-star integration plan)]
tech_stack:
  added: [PyYAML (generator), CC0 example content]
  patterns: [glob-load + point-expansion + resolve.Resolve, idempotent Python generator]
key_files:
  created:
    - examples/omarchy/opinions/OM-001.yaml .. OM-134.yaml (134 files)
    - examples/omarchy/points/*.yaml (32 files)
    - examples/omarchy/speech.yaml
    - examples/omarchy/gen/generate.py
    - examples/omarchy/README.md
    - examples/omarchy/LICENSE
    - examples/omarchy_test.go
  modified: []
decisions:
  - "Resolve receives flat []resolver.Opinion — test assembles this by expanding speech.Points through point files (resolve.Resolve does not read point files itself)"
  - "Status policy OQ-1 applied: required=OM-001/006/097/099+hw-conditional; nice-to-have=themes OM-114..134 + optional extras; all others required"
  - "Vanilla-arch hw profile: empty predicates/pci_ids — 35 hw-gated opinions land in Skipped (expected, not a hard conflict)"
  - "generate.py uses stdlib+PyYAML only; idempotent (re-run produces identical YAML)"
  - "sig_level=OptionalTrustAll (no space) used in OM-001 custom_repo — schema-enum-compliant"
metrics:
  duration: "25 min"
  completed: "2026-06-12"
  tasks: 2
  files: 170
---

# Phase 2 Plan 04: Omarchy North-Star Composition Summary

**One-liner:** 134 OS-agnostic OM-NNN opinion YAMLs + 32 curated points + one vanilla-Arch speech that resolves clean (Applied=99, Skipped=35, Hard conflicts=0), proving ARCH-02 with a CC0-licensed, script-generated, schema-valid example corpus.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Author all 134 opinions + 32 points (schema-valid) | b8e007d | 169 files (134 opinions, 32 points, generate.py, LICENSE, README) |
| 2 | speech.yaml + Go harness asserting clean resolution | 1bea96c | 2 files (speech.yaml, omarchy_test.go) |

## Verification Results

```
go test ./examples/ -run TestExampleOmarchy -count=1 -v
=== RUN   TestExampleOmarchy
    omarchy_test.go:240: TestExampleOmarchy: Applied=99 Skipped=35 Dropped=0 InstallOrder=99
--- PASS: TestExampleOmarchy (0.02s)

go test ./... -count=1
ok  github.com/mikl0s/debateos/examples        0.038s
ok  github.com/mikl0s/debateos/resolver/graph  0.002s
ok  github.com/mikl0s/debateos/resolver/hardware 0.004s
ok  github.com/mikl0s/debateos/resolver/parse  0.023s
ok  github.com/mikl0s/debateos/resolver/patch  0.002s
ok  github.com/mikl0s/debateos/resolver/resolve 0.003s
```

Acceptance checks:
- 134 opinion files: PASS
- 32 point files: PASS
- Coverage (each OM exactly once): PASS
- CC0 license: PASS
- speech targets arch: PASS (foundation: arch)
- speech references 32 points: PASS
- test asserts absence of Hard conflict: PASS
- generate.py committed and idempotent: PASS
- Go parse validation (134 opinions, 32 points): 0 errors

## Decisions Made

1. **Flat opinion assembly in test** — `resolve.Resolve` takes `[]resolver.Opinion` directly; the test expands `speech.Points` through point YAML files to build the flat slice. This is the correct caller pattern: resolve.go does not read point files.

2. **Status policy OQ-1** — required: OM-001 (custom-repo), OM-006 (compositor), OM-097 (display-manager), OM-099 (bootloader) + all hardware-conditional opinions (gated by `hardware_condition`, not status alone); nice-to-have: themes OM-114..134 + optional extras (npm tools, PWAs, branding, first-run UX niceties); all others required.

3. **Vanilla-arch baseline** — Empty `predicates`, `facts: {}`, `pci_ids: []` in speech.yaml. The 35 hardware-gated opinions (NVIDIA, Intel PTL, ASUS, Apple T2, etc.) resolve to Skipped — this is correct and expected for a generic machine target. The test explicitly verifies OM-068 (NVIDIA) is not in Applied.

4. **generate.py design** — Reads `research/omarchy-opinion-inventory.md` and `research/omarchy-points.md` as authoritative sources; uses stdlib+PyYAML only; includes coverage assertion (134 members, no duplicates). Re-running from repo root produces identical output.

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None. All opinions carry real intent and payload fields drawn from the Phase 0 inventory. Theme bundles reference `bundle_dir` paths which are data declarations consumed by the translator, not files that must exist in this repo.

## Threat Flags

| Flag | File | Description |
|------|------|-------------|
| threat_flag: trust-optionaltrustall | examples/omarchy/opinions/OM-001.yaml | OM-001 custom_repo sig_level=OptionalTrustAll — documented in plan T-02-06; resolver emits TrustWarning (T-01-10) |
| threat_flag: trust-never | examples/omarchy/opinions/OM-100.yaml | OM-100 arch-mact2 repo sig_level=Never — documented in plan; surface for T2 hardware only |

## Self-Check: PASSED

- examples/omarchy/opinions/ — 134 files: FOUND
- examples/omarchy/points/ — 32 files: FOUND
- examples/omarchy/speech.yaml: FOUND
- examples/omarchy/gen/generate.py: FOUND
- examples/omarchy_test.go: FOUND
- Commit b8e007d: FOUND (git log)
- Commit 1bea96c: FOUND (git log)
