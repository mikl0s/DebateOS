---
phase: 05-registry-forum-debate-ui
plan: "01"
subsystem: registry
tags: [registry, go-mod, foundation-compat, static-index, tdd]
dependency_graph:
  requires:
    - resolver/parse (ParsePoint, ParseOpinion strict YAML decoding)
    - resolver/types.go (Point, Opinion, PointMember structs)
    - translators/arch/capabilities.json
    - translators/debian/capabilities.json
  provides:
    - registry.GenerateIndex (point/opinion YAML → RegistryIndex)
    - registry.EmitHTML (RegistryIndex → static browse HTML)
    - registry.LoadCapabilities (capabilities.json → map[foundation][]token)
    - registry/index.ComputeCompat (per-point foundation compatibility)
    - registry/index.RegistryIndex, PointEntry, FoundationCompat (index schema)
    - go.mod with all Phase-5 Go deps (single-owner invariant)
  affects:
    - forum/ (05-03, 05-05) — consumes chi, sqlite, oauth2 from go.mod (must NOT modify go.mod)
    - web/ (05-02) — consumes registry/index.json at build time
tech_stack:
  added:
    - github.com/go-chi/chi/v5 v5.3.0 (Forum HTTP router, anchored here)
    - modernc.org/sqlite v1.46.1 (Forum pure-Go SQLite, PINNED, anchored here)
    - golang.org/x/oauth2 v0.34.0 (GitHub OAuth, anchored here; v0.36.0 requires Go 1.25)
  patterns:
    - TDD RED/GREEN per task (2 RED commits, 2 GREEN commits)
    - Deterministic JSON via sorted slices + struct field order + fixed generatedAt
    - resolver/parse strict validation (KnownFields=true) on every YAML doc ingested
    - foundation-compat computed from capabilities.json union-of-required-tokens algorithm
    - go.mod single-owner: deps_guard.go blank-import anchor prevents tidy pruning
key_files:
  created:
    - registry/generator.go
    - registry/index/index.go
    - registry/index/compat.go
    - registry/deps_guard.go
    - registry/generator_test.go
    - registry/index/compat_test.go
    - registry/testdata/fixtures/points/sample-point.yaml
    - registry/testdata/fixtures/opinions/SMP-001.yaml
    - registry/testdata/fixtures/opinions/SMP-002.yaml
    - registry/testdata/fixtures/opinions/SMP-003.yaml
    - registry/testdata/golden/index.json
  modified:
    - go.mod (added chi, sqlite, oauth2; go 1.24.0 preserved)
    - go.sum (updated for new deps)
decisions:
  - "[05-01] golang.org/x/oauth2 downgraded from v0.36.0 → v0.34.0: v0.36.0 requires go 1.25 (breaks go 1.24 directive); v0.34.0 is last go-1.24-compatible release"
  - "[05-01] golang.org/x/sys pinned to v0.38.0 (original): v0.42.0 (pulled by sqlite transitive chain) requires go 1.25; reverted manually"
  - "[05-01] modernc.org/libc pinned to v1.67.6: v1.67.7+ not verified as go-1.24 compatible at this time"
  - "[05-01] deps_guard.go blank-import pattern: three phase-5 deps anchored in registry/ as direct requires so forum/ plans never touch go.mod"
  - "[05-01] ComputeCompat uses union of TranslatorCapabilities across all point members (not per-member per-foundation) — consistent with Pattern 6 in 05-RESEARCH.md"
  - "[05-01] LoadCapabilities reads real capabilities.json at generation time; tests pass inline caps map for isolation (no filesystem side-effect in tests)"
  - "[05-01] Golden file auto-created on first run (os.IsNotExist → write + pass); subsequent runs compare byte-for-byte"
metrics:
  duration: "~6 minutes"
  completed_date: "2026-06-13"
  tasks_completed: 2
  files_changed: 13
---

# Phase 5 Plan 01: Registry Static Index Generator + Single-Owner go.mod Summary

**One-liner:** Go static registry generator (REG-01) — YAML → deterministic JSON index + static browse HTML with foundation-compat from capabilities.json; go.mod owns all Phase-5 deps (chi/sqlite v1.46.1/oauth2).

## What Was Built

### Task 1: Phase-5 Go Deps (single-owner go.mod) + foundation-compat (RED → GREEN)

**RED commit:** `ece27e9` — compat_test.go with 4 TestComputeCompat sub-tests + fixture YAML (build fails, no impl).

**GREEN commit:** `08bf8b8` — Implemented:

- `go.mod`: Added `github.com/go-chi/chi/v5 v5.3.0`, `modernc.org/sqlite v1.46.1` (PINNED), `golang.org/x/oauth2 v0.34.0`. `go 1.24.0` / `toolchain go1.24.1` preserved.
- `registry/deps_guard.go`: Blank-import anchor file — prevents `go mod tidy` from pruning the three deps as indirect before forum/ plans import them. Marked as "phase-5 dep anchor; consumed by forum/ (05-03, 05-05)".
- `registry/index/index.go`: `RegistryIndex{Schema, GeneratedAt, Points}` + `PointEntry{ID, Name, Intent, Curator, Members, FoundationCompat, CommitDate, Tags}` + `SchemaVersion = 1`.
- `registry/index/compat.go`: `FoundationCompat{Foundation, Compatible, Missing}` + `ComputeCompat` — collects union of `TranslatorCapabilities` across member opinions, checks each foundation's capability set, returns sorted-by-Foundation slice with Missing sorted lexically.

### Task 2: Generator — parse+validate fixtures, deterministic index.json + browse HTML (RED → GREEN)

**RED commit:** `3c1e700` — generator_test.go with 5 tests (TestGenerateIndex, TestGenerateIndexValidation, TestDeterminism, TestGoldenIndex, TestEmitHTML) — build fails, no impl.

**GREEN commit:** `fe7a9b7` — Implemented:

- `registry/generator.go`:
  - `GenerateIndex(fixturesDir, caps, generatedAt)`: walks `points/` and `opinions/` sub-dirs, parses every YAML via `parse.ParsePoint` / `parse.ParseOpinion` (KnownFields strict), returns wrapped error naming the offending file on failure (T-05-01 mitigated), sorts Points by ID and Members by ID for determinism.
  - `LoadCapabilities(archPath, debianPath)`: reads `{"capabilities": [...]}` JSON from both translators, returns `map[string][]string{"arch": [...], "debian": [...]}`.
  - `EmitHTML(idx, w)`: writes minimal static browse HTML — one table row per point, text-based `arch ✓ / debian ✗` compat badges, no `<script>` tags, GitHub Pages compatible.
- `registry/testdata/golden/index.json`: committed golden output with `fixedAt = "2026-01-01T00:00:00Z"` showing `sample-point` with `arch: compatible=true`, `debian: compatible=false, missing=["deploy-sddm-theme"]`.

## Verification Results

```
go test ./registry/... -count=1
ok  github.com/mikl0s/debateos/registry         (6 tests PASS)
ok  github.com/mikl0s/debateos/registry/index   (4 tests PASS)

go test ./... -count=1
ok  all packages (no regressions)
```

Tests verified:
- `TestComputeCompat/both_foundations_compatible` — PASS
- `TestComputeCompat/arch_only_token_debian_incompatible` — PASS (debian Missing: ["deploy-sddm-theme"])
- `TestComputeCompat/determinism_sorted_by_foundation` — PASS
- `TestComputeCompat/missing_sorted_lexically` — PASS
- `TestGenerateIndex` — PASS (fixture point indexed, members sorted, foundation_compat populated)
- `TestGenerateIndexValidation` — PASS (malformed YAML → error naming "BAD-001.yaml")
- `TestDeterminism` — PASS (two runs → byte-identical JSON)
- `TestGoldenIndex` — PASS (emitted JSON matches committed golden byte-for-byte)
- `TestEmitHTML` — PASS (JS-free HTML with point name + arch/debian badges)

go.mod invariants:
- `modernc.org/sqlite v1.46.1` — PINNED ✓
- `github.com/go-chi/chi/v5 v5.3.0` ✓
- `golang.org/x/oauth2 v0.34.0` ✓
- `go 1.24.0` — preserved ✓
- `toolchain go1.24.1` ✓

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] golang.org/x/oauth2 downgraded from v0.36.0 → v0.34.0**
- **Found during:** Task 1, running `go get golang.org/x/oauth2@v0.36.0`
- **Issue:** v0.36.0 requires `go >= 1.25.0` — `go get` silently upgraded the `go` directive to 1.25.0, violating the plan invariant "go 1.24.0 directive must be unchanged"
- **Fix:** Pinned `golang.org/x/oauth2@v0.34.0` (latest go-1.24-compatible release). All oauth2 functionality needed by forum/ (05-05) is present in v0.34.0.
- **Files modified:** go.mod, go.sum
- **Impact:** None on forum/ — both v0.34 and v0.36 provide the same oauth2 API surface

**2. [Rule 1 - Bug] golang.org/x/sys pinned to v0.38.0 (original version)**
- **Found during:** Task 1, go mod tidy cascade
- **Issue:** modernc.org/sqlite's transitive deps pulled in `golang.org/x/sys@v0.42.0` which requires go 1.25.0
- **Fix:** Explicitly pinned `golang.org/x/sys@v0.38.0` (original version from go.mod before this plan; requires go 1.24.0). Used GOTOOLCHAIN=local to validate the constraint.
- **Files modified:** go.mod, go.sum

**3. [Rule 1 - Bug] modernc.org/libc pinned to v1.67.6 (not v1.67.7)**
- **Found during:** Task 1, go mod tidy resolution
- **Issue:** tidy wanted v1.67.7; v1.67.6 is the cached version with go 1.24.0 requirement verified
- **Fix:** Explicit `go get modernc.org/libc@v1.67.6` to keep the go-1.24-compatible version

### Scope Note: golang.org/x/exp added as transitive dep

`golang.org/x/exp v0.0.0-20251023183803-a4bb9ffd2546` was added as an indirect dep (pulled in by modernc.org/sqlite). This is expected — verified it requires go 1.24.0. No action needed.

## Known Stubs

None. All data is live: fixture YAML → real parse → real compat computation → committed golden.

## Threat Flags

No new security-relevant surface introduced beyond the plan's `<threat_model>`:
- T-05-01 mitigated: every doc validated via `parse.ParseOpinion`/`parse.ParsePoint` (KnownFields strict); malformed input returns wrapped error naming the file
- T-05-SC mitigated: all three packages verified in 05-RESEARCH.md Package Legitimacy Audit (all OK/Approved); sqlite PINNED to v1.46.1 (downgrade of oauth2 to v0.34.0 is also Approved)

## Self-Check: PASSED
