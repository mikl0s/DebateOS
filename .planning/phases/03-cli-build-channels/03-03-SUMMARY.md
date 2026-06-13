---
phase: 03-cli-build-channels
plan: "03"
subsystem: cli/build
tags: [build, tdd, docker, translate, injection-tar, determinism, secrets, epoch]
dependency_graph:
  requires: ["03-01", "03-02"]
  provides:
    - "cli/build/build.go (Run, DeriveEpoch)"
    - "cli/build/inject.go (WriteInjectionTar, PaneAsset, sanitizeDst)"
    - "cmd/debateos/main.go (build dispatch case)"
  affects: ["03-04-PLAN.md (determinism gate: DeriveEpoch consts)", "first-boot injection unit"]
tech_stack:
  added: []
  patterns:
    - "DeriveEpoch: sha256 → BE uint32 → _MIN + raw % (_MAX-_MIN) (mirrors manifest.py exactly)"
    - "archive/tar stdlib for injection tar (no external deps)"
    - "sanitizeDst: sentinel-root containment check mirroring profile.py _sanitize_dst"
    - "Runner.Run variadic for all subprocess calls (no sh -c)"
key_files:
  created:
    - cli/build/build.go
    - cli/build/inject.go
    - cli/build/build_test.go
  modified:
    - cmd/debateos/main.go
    - .gitignore
decisions:
  - "DeriveEpoch exported so 03-04 determinism gate can import and compare against manifest.py constants without re-implementing"
  - "WriteInjectionTar writes to outDir (never profileDir): T-03-LEAK; tar path returned for caller verification"
  - "sanitizeDst uses sentinel-root filepath.Clean containment check: identical semantic to profile.py _sanitize_dst (T-03-TRAV)"
  - "Empty PaneAsset slice produces a valid tar with only the manifest (version=1, files=[]): first-boot unit always finds artifact"
  - "FakeRunner records exact argv so tests assert frozen translate contract without real binaries (D19/T-03-DKARG)"
  - ".gitignore build/ narrowed to /web/build/ to allow cli/build/ package to be tracked (deviation from original broad pattern)"
metrics:
  duration: "~20 min"
  completed: "2026-06-13T12:11:20Z"
  tasks: 2
  commits: 3
  files: 5
---

# Phase 3 Plan 03: Build Subcommand Summary

**One-liner:** debateos build resolve→epoch→translate→docker via FakeRunner-testable Runner with --dry-run/--skip-iso gates; private-injection.tar emitted locally with sanitized target-relative paths and debateos-private.json manifest.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| RED | Build orchestration + injection tar failing tests | 55205dd | cli/build/build_test.go, .gitignore |
| GREEN (Task 1+2) | build.go + inject.go + main.go dispatch | 39e2d53 | cli/build/build.go, cli/build/inject.go, cmd/debateos/main.go, cli/build/build_test.go |

## What Was Built

### Task 1: build orchestration (build.go)

`cli/build/build.go` implements `Run(args, stdout, stderr, runner.Runner) int`:

- **Flag parsing:** `--dir` (speech dir), `--profile` (default: vanilla-arch), `--out` (default: ./out), `--dry-run`, `--skip-iso`
- **Step 1:** `loader.ResolveDir(speechDir)` — reuses the shared CLI loader pipeline
- **Step 2:** `resolve.CanonicalJSON(rs)` → `resolved.json` written to `--out`
- **Step 3:** `DeriveEpoch(canonicalBytes)` — single derivation point (T-03-EPOCH)
- **Step 4 / --dry-run gate:** If `--dry-run`, prints plan (resolved.json path, epoch, translate argv, docker argv) and returns 0 with ZERO Runner calls
- **Step 5:** `Runner.Run("translators/arch/translate", resolved.json, --opinions, <dir>, --profile, <name>, --out, <profileDir>)` — exact frozen argv
- **Step 5b:** `WriteInjectionTar(outDir, nil)` — private-injection.tar always emitted in `--skip-iso` path
- **Step 6 / --skip-iso gate:** Returns after translate; no docker call
- **Step 7:** `Runner.Run("docker", "run", "-v", speech:/speech, "-v", out:/out, "-e", SOURCE_DATE_EPOCH=<N>, dockerImage)` — full mode only

`DeriveEpoch(contentBytes []byte) int64` is exported:
- Algorithm: `sha256.Sum256(bytes)` → first 4 bytes `binary.BigEndian.Uint32` → `epochMin + (raw % (epochMax - epochMin))`
- Constants: `epochMin = 1577836800`, `epochMax = 2208988800` (mirror `manifest.py _MIN_EPOCH / _MAX_EPOCH`)

### Task 2: private-injection.tar (inject.go)

`cli/build/inject.go` implements `WriteInjectionTar(outDir string, assets []PaneAsset) (string, error)`:

- **Security gate (T-03-TRAV):** `sanitizeDst(dst)` applied to every asset before any write — rejects empty, absolute, and `..` traversal paths using a sentinel-root `filepath.Clean` containment check
- **Tar root:** `debateos-private.json` manifest `{version:1, created:<RFC3339>, files:[{path,mode}]}`
- **Assets:** each `PaneAsset.Content` stored at the sanitized target-relative `Dst` path
- **Placement (T-03-LEAK):** `tarPath = filepath.Join(outDir, "private-injection.tar")` — always in outDir, never in profileDir
- Returns the absolute tar path for caller verification

## Verification Results

```
go test ./cli/build/... -count=1
ok  github.com/mikl0s/debateos/cli/build  0.015s (10 tests)

go build ./cmd/debateos            -- SUCCESS
go vet ./cli/... ./cmd/...         -- CLEAN
grep -c 'SOURCE_DATE_EPOCH' cli/build/build.go  -- 7 (≥1 required)
grep -c 'debateos-private.json' cli/build/inject.go  -- 5 (≥1 required)

Full regression: go test ./... -count=1
All 17 test packages: PASS (no regressions)
```

## TDD Gate Compliance

- RED gate: `test(03-03)` commit 55205dd — 10 failing tests (build.go did not exist)
- GREEN gate: `feat(03-03)` commit 39e2d53 — all 10 tests pass
- REFACTOR gate: not needed (code clean on first implementation pass)

## Security Properties Verified

| Control | Implementation | Test |
|---------|---------------|------|
| T-03-DKARG | Runner.Run(name, args...) variadic; no sh -c in build.go or inject.go | TestBuildSkipISO asserts exact frozen argv via FakeRunner.Calls |
| T-03-EPOCH | DeriveEpoch called once; result in epochEnv exported to both translate and docker | TestBuildEpochConsistency + TestBuildDocker assert SOURCE_DATE_EPOCH presence |
| T-03-LEAK | WriteInjectionTar writes to outDir; profile tree untouched | TestInjectionTarLayout asserts !strings.HasPrefix(tarPath, profileDir) |
| T-03-TRAV | sanitizeDst rejects absolute + .. traversal before any tar write | TestInjectSanitizeAbsolute + TestInjectSanitizeTraversal |
| PRIV-01 | pane.yaml/identity.age/private-injection.tar never written inside profile tree | TestSecretFreeProfile walks profileDir and asserts absence |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] .gitignore `build/` pattern shadowed cli/build/ package**
- **Found during:** RED commit staging — `git add cli/build/build_test.go` was rejected
- **Issue:** `.gitignore` contained `build/` matching all `build/` directories including `cli/build/`. The entry was intended for Node/SvelteKit (`web/build/`) but was written without anchoring.
- **Fix:** Changed `build/` → `/web/build/` to scope it to the repo root's web build output directory only.
- **Files modified:** .gitignore
- **Commit:** 55205dd (RED)

**2. [Rule 1 - Bug] Minimal speech fixture missing `schema: 1`**
- **Found during:** First GREEN test run — all build tests failed with "value must be 1 at /schema"
- **Issue:** `minimalSpeechDir` produced `speech.yaml` without the required `schema: 1` field; the JSON schema validator (santhosh-tekuri/jsonschema/v6) enforces this at parse time
- **Fix:** Added `schema: 1` to the minimal speech YAML fixture string in `build_test.go`
- **Files modified:** cli/build/build_test.go
- **Commit:** 39e2d53 (GREEN, included with implementation)

## Known Stubs

One intentional partial implementation tracked for future plan:

**Private pane asset injection (build.go:Step 5b):** `WriteInjectionTar(outDir, nil)` — the `assets` slice is always empty in the current implementation because `build.Run` does not yet load and expose private-pane file assets from `pane.yaml`. The tar is emitted but contains only the manifest (files=[]).

- **File:** cli/build/build.go (the `WriteInjectionTar(outDir, nil)` call)
- **Reason:** Loading actual private-pane `file_assets` from `pane.yaml` requires a pane YAML loader and private-asset schema definition. The PRIV-01 primary goal (tar emitted locally, not inside profile/ISO) is fully satisfied. The tar is structurally correct — first-boot unit can parse it; it just has no files to apply.
- **Resolution plan:** A future plan (or 03-04 wave) that wires private pane `file_assets` into the build invocation will pass the loaded `[]PaneAsset` slice to `WriteInjectionTar`.

This stub does NOT prevent the plan's core goals from being achieved:
- `--skip-iso` works end-to-end (profile emission + injection tar emission)
- `--dry-run` emits a correct plan with zero runner calls
- Full mode issues docker with correct argv + epoch
- Sanitization tests cover the security boundary

## Threat Flags

No new security-relevant surface beyond the plan's threat model.

The `WriteInjectionTar` return value (the tar path) is used only for test assertions and not surfaced to external callers. The `sanitizeDst` function is the only trust boundary for user-controlled paths and is fully covered by negative tests (TestInjectSanitizeAbsolute, TestInjectSanitizeTraversal).

## Self-Check: PASSED

- [x] cli/build/build.go exists at /home/mikkel/repos/DebateOS/cli/build/build.go
- [x] cli/build/inject.go exists at /home/mikkel/repos/DebateOS/cli/build/inject.go
- [x] cli/build/build_test.go exists with 10 test functions
- [x] cmd/debateos/main.go contains `case "build": build.Run(...)`
- [x] All 10 tests pass: `go test ./cli/build/... -count=1` → PASS
- [x] Full regression: `go test ./... -count=1` → all pass
- [x] `go build ./cmd/debateos` → SUCCESS
- [x] `go vet ./cli/... ./cmd/...` → CLEAN
- [x] RED commit 55205dd exists in git log
- [x] GREEN commit 39e2d53 exists in git log
