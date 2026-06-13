---
phase: 03-cli-build-channels
plan: "01"
subsystem: cli-foundation
tags: [cli, tdd, config, runner, compose, validate, age]
dependency_graph:
  requires: [02-05-SUMMARY.md]
  provides: [cli/config.DebateOSDir, cli/runner.Runner, cli/compose.Run, cli/validate.Run, cmd/debateos/main.go, filippo.io/age-direct-dep]
  affects: [03-02-PLAN.md, 03-03-PLAN.md]
tech_stack:
  added: [filippo.io/age v1.3.1 (direct dep)]
  patterns: [flag.FlagSet subcommand dispatch, Runner interface, FakeRunner test double, shared internal loader]
key_files:
  created:
    - cli/config/config.go
    - cli/config/config_test.go
    - cli/runner/runner.go
    - cli/runner/fake.go
    - cli/runner/runner_test.go
    - cli/internal/loader/loader.go
    - cli/compose/compose.go
    - cli/compose/compose_test.go
    - cli/validate/validate.go
    - cli/validate/validate_test.go
    - cmd/debateos/main.go
  modified:
    - go.mod (filippo.io/age promoted to direct)
    - go.sum
decisions:
  - "Runner interface uses variadic args (exec.Command(name, args...)); never sh -c — per T-03-AI threat mitigation"
  - "Shared cli/internal/loader.ResolveDir() extracted so compose and validate reuse identical pipeline without drift from cmd/resolve-json"
  - "FakeRunner join-key is 'name arg1 arg2 ...' (space-joined); Outputs map uses same key — consistent with 03-RESEARCH.md Pattern 2"
  - "compose prints Applied/Skipped/Dropped counts + all Explanation .Text lines; no JSON output (preview format)"
  - "validate prints one-line 'validate: OK — Applied=N Skipped=N Dropped=N' to stdout on success; failure reason to stderr"
  - "os.Exit called exactly once per exit point in main(); zero os.Exit in cli/ packages — testable Run() signature"
  - "filippo.io/age promoted to direct require in go.mod for plan 03-02 pane package to import without go.mod conflict"
metrics:
  duration: "~12 min"
  completed: "2026-06-13"
  tasks: 2
  commits: 4
  files: 12
---

# Phase 3 Plan 01: CLI Foundation Summary

**One-liner:** stdlib flag.FlagSet subcommand dispatch with testable Run() int exit codes, DEBATEOS_DIR-first config resolution, Runner/FakeRunner interface, and compose/validate subcommands backed by the shared resolve pipeline — filippo.io/age v1.3.1 promoted to direct dep for pane plan.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 (RED) | Config/Runner failing tests | c380af2 | cli/config/config_test.go, cli/runner/runner_test.go |
| 1 (GREEN) | Config/Runner implementation + age dep | 620f6a8 | cli/config/config.go, cli/runner/runner.go, cli/runner/fake.go, go.mod, go.sum |
| 2 (RED) | Compose/Validate failing tests | 183b7d4 | cli/compose/compose_test.go, cli/validate/validate_test.go |
| 2 (GREEN) | Compose/Validate + main.go | 4d63d26 | cli/compose/compose.go, cli/validate/validate.go, cli/internal/loader/loader.go, cmd/debateos/main.go |

## Verification Results

```
go test ./cli/config/... ./cli/runner/... ./cli/compose/... ./cli/validate/... -count=1
ok  github.com/mikkel0s/debateos/cli/config    0.002s
ok  github.com/mikkel0s/debateos/cli/runner    0.003s
ok  github.com/mikkel0s/debateos/cli/compose   0.039s
ok  github.com/mikkel0s/debateos/cli/validate  0.043s

go build ./cmd/debateos ./cmd/resolve-json  -- SUCCESS
go vet ./cli/... ./cmd/...                  -- CLEAN
grep -c 'func DebateOSDir' cli/config/config.go  -- 1
grep 'filippo.io/age' go.mod | grep -v '// indirect'  -- filippo.io/age v1.3.1 (direct)
```

## TDD Gate Compliance

- RED commit (test): c380af2 — config/runner failing tests
- GREEN commit (feat): 620f6a8 — config/runner implementation
- RED commit (test): 183b7d4 — compose/validate failing tests
- GREEN commit (feat): 4d63d26 — compose/validate + main.go

Both tasks followed strict RED before GREEN sequence. All tests failed to compile before implementation (correct RED signal).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing critical functionality] Extracted shared loader package**
- **Found during:** Task 2 implementation
- **Issue:** Plan noted "extract [loading logic] into a shared internal helper"; both compose and validate would have duplicated the 6-step load pipeline otherwise
- **Fix:** Created `cli/internal/loader/loader.go` with `ResolveDir()` function extracting the exact pipeline from `cmd/resolve-json/main.go`
- **Files modified:** cli/internal/loader/loader.go (new)
- **Commit:** 4d63d26

**2. [Rule 1 - Promotion approach] go mod tidy pruned filippo.io/age**
- **Found during:** Task 1 go.mod promotion
- **Issue:** Running `go get filippo.io/age@v1.3.1` followed by `go mod tidy` removed the dep since no code in the module yet imports age (pane package is plan 03-02)
- **Fix:** Manually edited go.mod to add `filippo.io/age v1.3.1` as a direct require with all its transitive deps restored; ran `go build ./...` to confirm validity
- **Files modified:** go.mod
- **Commit:** 620f6a8

## Known Stubs

None — all implemented functionality is fully wired. compose and validate both read real speech directories and run the live resolver pipeline.

## Threat Flags

No new security-relevant surface not in the plan's threat model. The `cli/internal/loader` package reuses the existing `resolver/parse` + `resolver/resolve` pipeline which already has schema validation (T-03-IV mitigation). The ExecRunner uses variadic args not `sh -c` (T-03-AI mitigation). Config dir returns paths only (T-03-CFG accepted).

## Self-Check: PASSED

All 11 implementation files verified present on disk. All 4 task commits (c380af2, 620f6a8, 183b7d4, 4d63d26) verified in git log.
