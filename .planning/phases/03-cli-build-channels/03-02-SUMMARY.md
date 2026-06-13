---
phase: 03-cli-build-channels
plan: "02"
subsystem: cli/pane
tags: [age, encryption, pane, secrets, tdd, cli]
dependency_graph:
  requires: ["03-01"]
  provides: ["cli/pane/age.go", "cli/pane/pane.go", "cmd/debateos (pane dispatch)"]
  affects: ["cmd/debateos/main.go"]
tech_stack:
  added: ["filippo.io/age v1.3.1 (direct dep, now actively used for X25519 identity + encrypt/decrypt)"]
  patterns: ["Runner interface for git subprocess isolation", "0600 file permissions for all secrets"]
key_files:
  created:
    - cli/pane/age.go
    - cli/pane/pane.go
  modified:
    - cli/pane/pane_test.go
    - cmd/debateos/main.go
    - .gitignore
decisions:
  - "age X25519 identity stored at identity.age (0600) in config dir; no escrow, no central service (PRIV-01/D16)"
  - "Only pane.yaml.age (ciphertext) ever staged/committed — plaintext pane.yaml never in git (T-03-PLAINTEXT)"
  - "git add/commit/push routed through Runner interface (FakeRunner in tests — zero network calls)"
  - "Runner.Run uses variadic args, never sh -c string (T-03-GITARG)"
metrics:
  duration: "~35 minutes (continued from previous session)"
  completed: "2026-06-13T02:21:01Z"
  tasks_completed: 2
  files_changed: 5
---

# Phase 03 Plan 02: Private Pane + Age Backup/Restore Summary

**One-liner:** Age X25519 identity local-only secrets management with 0600-gated pane.yaml, FakeRunner-isolated git backup/restore, and lossless age encrypt/decrypt round-trip.

## What Was Built

### Task 1: age identity + encrypt/decrypt (age.go)

`cli/pane/age.go` implements three functions:

- `LoadOrCreateIdentity(dir)` — reads `identity.age` (0600) or generates a new X25519 identity via `filippo.io/age` if absent. Private key written with `O_CREATE|O_WRONLY|O_TRUNC, 0600` to prevent wider permissions on first create.
- `EncryptFile(id, src, dst)` — age-encrypts src to dst (0600) using the identity's public key.
- `DecryptFile(id, src, dst)` — age-decrypts src to dst (0600) using the identity.

All secret files use `os.OpenFile(..., 0600)` — never `os.WriteFile` (which doesn't guarantee mode on pre-existing files).

### Task 2: pane verb dispatcher (pane.go)

`cli/pane/pane.go` implements `Run(args, stdout, stderr, runner.Runner) int` with verbs:

- `set <key> <value>` — writes to `pane.yaml` (0600 via OpenFile), YAML-marshaled via `go.yaml.in/yaml/v3`
- `get <key>` — reads and prints value; returns non-zero on missing key
- `list` — prints all keys sorted alphabetically
- `backup` — encrypts `pane.yaml` → `pane.yaml.age` (0600), then calls `Runner.Run("git", "add", ...)`, `Runner.Run("git", "commit", ...)`, `Runner.Run("git", "push")` — only the `.age` file is staged (T-03-PLAINTEXT)
- `restore` — calls `Runner.Run("git", "pull")` then decrypts `pane.yaml.age` → `pane.yaml` (0600)

`cmd/debateos/main.go` was updated with the `case "pane"` dispatch.

## Test Results

All 9 tests GREEN:

```
=== RUN   TestIdentityCreation    --- PASS (0.00s)
=== RUN   TestAgeRoundTrip        --- PASS (0.00s)
=== RUN   TestAgeRoundTripWrongIdentity --- PASS (0.00s)
=== RUN   TestPaneSetGet          --- PASS (0.00s)
=== RUN   TestPaneGetMissing      --- PASS (0.00s)
=== RUN   TestPaneList            --- PASS (0.00s)
=== RUN   TestPanePermissions     --- PASS (0.00s)
=== RUN   TestPaneBackup          --- PASS (0.00s)
=== RUN   TestPaneRestore         --- PASS (0.00s)
PASS    ok  github.com/mikl0s/debateos/cli/pane  0.008s
```

Full regression suite: all packages pass.

## Commits

| Hash | Type | Description |
|------|------|-------------|
| f56ed89 | test(03-02) | RED — add 9 failing tests for pane + age |
| f0b6e9e | feat(03-02) | GREEN — implement age.go + pane.go + fix imports + pane dispatch |

## TDD Gate Compliance

- RED gate: `test(03-02)` commit f56ed89 exists
- GREEN gate: `feat(03-02)` commit f0b6e9e exists
- REFACTOR gate: not needed (code clean on first pass)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Removed self-referential dependency from go.mod**
- **Found during:** GREEN implementation — `go build ./cli/pane/` failed
- **Issue:** go.mod contained `github.com/mikkel0s/debateos v0.0.0-20260613004740-4b5579b94015` as an explicit require entry — the module requiring itself as an external dependency at a stale GitHub commit that predated all implementation packages. go resolved internal imports by fetching that old commit, which lacked the packages, causing "no required module provides package" errors.
- **Fix:** Removed the self-referential line from go.mod and the corresponding stale entries from go.sum (`github.com/mikl0s/debateos v0.0.0-...`)
- **Files modified:** go.mod, go.sum (already committed at HEAD from pre-session state; no change needed after fix)
- **Commit:** f0b6e9e (included in GREEN commit)

**2. [Rule 1 - Bug] Fixed import paths using wrong module owner spelling**
- **Found during:** GREEN implementation — `go build ./cli/pane/` still failed after removing self-reference
- **Issue:** The module name in go.mod is `github.com/mikl0s/debateos` (m-i-k-l-0-s) but pane.go, pane_test.go, and main.go all imported from `github.com/mikkel0s/debateos` (m-i-k-k-e-l-0-s — extra 'k' and 'e'). The wrong name was also present in the RED commit (pane_test.go).
- **Fix:** Rewrote all three files with correct `github.com/mikl0s/debateos` import paths. Used raw byte inspection to confirm correctness.
- **Files modified:** cli/pane/pane.go, cli/pane/pane_test.go, cmd/debateos/main.go
- **Commit:** f0b6e9e

**3. [Rule 3 - Blocking] Restored accidentally deleted cli/compose/compose_test.go**
- **Found during:** `git status` before commit showed compose_test.go as deleted
- **Issue:** The file was deleted during this session's debugging attempts
- **Fix:** `git checkout -- cli/compose/compose_test.go`
- **Commit:** Not separately committed — file restored before staging

**4. [Rule 2 - Missing] Added built binary entries to .gitignore**
- **Found during:** `git status` showed `/debateos` and `/resolve-json` as untracked
- **Fix:** Added `/debateos` and `/resolve-json` to .gitignore
- **Files modified:** .gitignore
- **Commit:** f0b6e9e

## Security Properties Verified

| Control | Implementation | Test |
|---------|---------------|------|
| T-03-PERM | All secrets use `os.OpenFile(..., 0600)` | TestIdentityCreation, TestAgeRoundTrip, TestPaneSetGet, TestPaneBackup, TestPaneRestore |
| T-03-PLAINTEXT | Only `pane.yaml.age` staged; pane.yaml never in git | TestPaneBackup asserts git add path |
| T-03-GITARG | `Runner.Run("git", variadic...)` — no sh -c | Code review + FakeRunner records exact args |
| PRIV-01/D16 | identity.age local-only, no escrow code path | Design only — no network in implementation |
| V6 | filippo.io/age used — no hand-rolled crypto | age.go |

## Known Stubs

None. All functionality is fully wired.

## Threat Flags

No new security surface introduced beyond what the threat model anticipated. The pane backup/restore touches git via Runner interface (isolation guaranteed), and all file I/O respects 0600.

## Self-Check: PASSED

- [x] cli/pane/age.go exists
- [x] cli/pane/pane.go exists
- [x] All 9 tests pass: `go test ./cli/pane/... -count=1` → PASS
- [x] Full regression: `go test ./...` → all pass
- [x] GREEN commit f0b6e9e exists in git log
- [x] RED commit f56ed89 exists in git log
