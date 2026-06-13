---
phase: 03-cli-build-channels
plan: "04"
subsystem: build/docker, build/actions, scripts, docs
tags: [bld-01, bld-02, bld-03, bld-04, priv-01, docker, gha, determinism, coverage, docs]
dependency_graph:
  requires: ["03-01", "03-02", "03-03"]
  provides:
    - "build/docker/Dockerfile (multi-stage golang builder + digest-pinned archlinux runtime)"
    - "build/docker/entrypoint.sh (container entrypoint: /speech → /out)"
    - ".dockerignore (T-03-CTX: excludes pane.yaml/*.age/private-injection.tar/.config/)"
    - "build/actions/build-speech.yml (workflow_call reusable workflow; same image as BLD-01)"
    - "build/actions/README.md (fork-and-build guide; deferred live-CI note)"
    - ".github/workflows/build-speech.yml (thin caller; same-image proof)"
    - "scripts/determinism-test.sh (double-run sha256 gate; PASSES)"
    - "scripts/secret-free-check.sh (profile tree grep gate; PASSES)"
    - "scripts/check-coverage.sh (two-gate: resolver >=90%, cli >=85%; PASSES 93.5%/85.6%)"
    - "cli/internal/loader/loader_test.go (new test file; loader 88.4%)"
    - "docs/cli-build-channels.md (346-line end-to-end guide)"
    - ".planning/REQUIREMENTS.md (CLI-01/02 + BLD-01..04 + PRIV-01 marked Complete)"
    - ".planning/ROADMAP.md (Phase 3 finalized: 4/4 plans Complete)"
  affects: ["Phase 4 (Debian translator now unblocked)"]
tech_stack:
  added: []
  patterns:
    - "Multi-stage Docker: golang:1.24 builder CGO_ENABLED=0 → archlinux:base-devel runtime (same digest as translators/arch/Dockerfile)"
    - "GHA workflow_call reusable workflow with container: directive (same image)"
    - "Deterministic tar: --sort=name --mtime=@EPOCH --owner=0 --group=0 --numeric-owner --pax-option=delete=atime,delete=ctime"
    - "SOURCE_DATE_EPOCH derived from sha256(resolved.json): first 4 bytes BE uint32 → epochMin + (raw % range)"
    - "FakeRunnerFunc: per-call error control for pane/build test coverage"
    - "GHA input injection prevention: all ${{ inputs.* }} via env vars in run: (T-03-CIWF)"
key_files:
  created:
    - build/docker/Dockerfile
    - build/docker/entrypoint.sh
    - .dockerignore
    - build/actions/build-speech.yml
    - build/actions/README.md
    - .github/workflows/build-speech.yml
    - scripts/determinism-test.sh
    - scripts/secret-free-check.sh
    - cli/internal/loader/loader_test.go
    - docs/cli-build-channels.md
  modified:
    - scripts/check-coverage.sh
    - cli/build/build_test.go
    - cli/pane/pane_test.go
    - cli/validate/validate_test.go
    - cli/runner/fake.go
    - cli/runner/runner_test.go
    - .planning/REQUIREMENTS.md
    - .planning/ROADMAP.md
decisions:
  - "build/docker/Dockerfile digest IDENTICAL to translators/arch/Dockerfile (sha256:dd60dfcca90f1ee6c2dd265ed27062070a1fb2e3b307723838a9d97741284722) — single pin source of truth"
  - "GHA workflow inputs via env vars in run: steps (not inline ${{ }} interpolation) — T-03-CIWF injection prevention"
  - "FakeRunnerFunc added to cli/runner to enable per-call error control needed for pane backup/restore error path coverage"
  - "cli/internal/loader/loader_test.go added as new file (not in plan) — Rule 2: required for 85% coverage gate"
  - "validate_test.go extended with TestValidate_MinimalSpeech + NoHomeOrDebateosDIR — Rule 2: coverage"
  - "Determinism script uses python3 to derive epoch (same algorithm as manifest.py) — consistent with DeriveEpoch in cli/build"
metrics:
  duration: "~17 min"
  completed: "2026-06-13T12:28:45Z"
  tasks: 3
  commits: 3
  files: 18
---

# Phase 3 Plan 04: Docker Image + Actions Workflow + Gates Summary

**One-liner:** Multi-stage Docker image (CGO_ENABLED=0 + digest-pinned archlinux) + GHA reusable workflow + determinism/secret-free/coverage gates all passing; Phase 3 finalized in REQUIREMENTS.md and ROADMAP.md.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Multi-stage Docker image + entrypoint + .dockerignore | 0220973 | build/docker/Dockerfile, build/docker/entrypoint.sh, .dockerignore |
| 2 | GHA reusable workflow + thin caller + scripts + coverage | 1044817 | build/actions/*, .github/workflows/*, scripts/*, cli/internal/loader/loader_test.go, extended test files |
| 3 | Zero-cost docs + REQUIREMENTS/ROADMAP status | 03bc03e | docs/cli-build-channels.md, .planning/REQUIREMENTS.md, .planning/ROADMAP.md |

## What Was Built

### Task 1: Multi-stage Docker image (BLD-01)

`build/docker/Dockerfile` — two-stage build:
- **Stage 1 (builder):** `FROM golang:1.24` — copies go.mod/go.sum first (layer cache), then full source; builds `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/debateos ./cmd/debateos`
- **Stage 2 (runtime):** `FROM archlinux:base-devel@sha256:dd60dfcca90f1ee6c2dd265ed27062070a1fb2e3b307723838a9d97741284722` — installs archiso + python-yaml; copies binary + translators/ + schemas/ from builder; runs entrypoint.sh
- Digest IDENTICAL to `translators/arch/Dockerfile` (single pin; quarterly-reverify comment included)
- `ENTRYPOINT ["/debateos/entrypoint.sh"]`

`build/docker/entrypoint.sh` — mounts /speech + /out, reads `SKIP_ISO` env, invokes `debateos build --dir /speech --out /out [--skip-iso]`; `bash -n` validated.

`.dockerignore` — excludes `pane.yaml`, `*.age`, `identity.age`, `private-injection.tar`, `.config/`, `.planning/`, `.git/` — repo root build context never leaks HOME secrets (T-03-CTX).

### Task 2: GHA workflow + gates (BLD-02, BLD-03, PRIV-01, D19)

**`build/actions/build-speech.yml`** — reusable `workflow_call` with inputs `speech-dir` (required), `profile` (default: vanilla-arch), `skip-iso` (bool); job runs `container: image: ghcr.io/mikl0s/debateos:latest` — the SAME image as BLD-01. All `${{ inputs.* }}` values passed via `env:` in `run:` steps (T-03-CIWF injection guard). Uploads build artifact via `actions/upload-artifact@v4`.

**`.github/workflows/build-speech.yml`** — thin caller triggering on `push` to `examples/**` or `build/**`, calling `uses: ./.github/workflows/build-speech.yml` with `examples/omarchy` / `vanilla-arch` / `skip-iso: true`. Proves same-image reuse.

**`build/actions/README.md`** — fork-and-build guide covering cross-repo `uses:` syntax, local Docker channel, deferred live-CI note, input table, security notes.

**`scripts/determinism-test.sh`** — double-run resolve+translate → deterministic tar (verified flags) → sha256 compare. PASSES: identical sha256 `99db28900f05b1230baaec16b3a729144fe8582f9a39b50c1835f2cc6d73b069` across both runs. Epoch derived in-script via same python3 algorithm as `manifest.py`.

**`scripts/secret-free-check.sh`** — builds profile (`--skip-iso`) and `find`s arch-profile/ for `pane.yaml`, `identity.age`, `private-injection.tar`. PASSES: all three absent.

**`scripts/check-coverage.sh`** — extended to two-gate: `resolver/ >=90%` (existing) + `cli/ >=85%` (new). PASSES: resolver 93.5%, cli 85.6%.

**Coverage extension required adding tests:**
- `cli/internal/loader/loader_test.go` (new file, 12 tests): covers ResolveDir happy path (omarchy + minimal), missing speech.yaml, malformed YAML, missing points/, missing opinions/, unknown point ref, bad point YAML, unknown opinion in point, bad opinion YAML, direct opinion ref. Loader coverage: 88.4%.
- Extended `cli/build/build_test.go`: bad speech dir, unknown flag, Runner error paths, no config dir, empty assets, sanitizeDst empty, DeriveEpoch boundaries. Build coverage: 82.8%.
- Extended `cli/pane/pane_test.go`: 14 additional tests covering Run error paths, set/get/list errors, backup git errors, restore errors, identity corruption. Pane coverage: 83.4%.
- Extended `cli/validate/validate_test.go`: no config dir, minimal speech. Validate coverage: 90.5%.
- Extended `cli/runner/runner_test.go`: FakeRunnerFunc with RunFn, nil RunFn, OutputFn, nil OutputFn. Runner coverage: 94.1%.
- `cli/runner/fake.go`: added `FakeRunnerFunc` for per-call error control.

Both workflow YAMLs validated via `python3 -c "import yaml; yaml.safe_load(open(...))"` — OK.

### Task 3: Documentation + status (BLD-04)

**`docs/cli-build-channels.md`** (346 lines) covers:
- Step-by-step compose → validate → pane → build workflow
- Channel 1 (local Docker) and Channel 2 (GitHub Actions fork-and-build)
- Docker image architecture (multi-stage, digest-pin, build-yourself instructions)
- Deterministic builds (how SOURCE_DATE_EPOCH is derived; tar flags explanation)
- Privacy & secrets (pane.yaml local-only; age key-management; lose-identity = lose-backup)
- Flash-time injection (private-injection.tar → USB → first-boot unit applies)
- Zero-cost no-central-service guarantee table
- Deferred verifications (full ISO on capable host; live cross-repo Actions run)

**REQUIREMENTS.md:** CLI-01, CLI-02, BLD-01, BLD-02, BLD-03, BLD-04, PRIV-01 all marked `[x]` Complete in section checkboxes and traceability table with evidence notes and deferred caveats.

**ROADMAP.md:** 03-04-PLAN.md checked; Phase 3 progress table updated to `4/4 Complete 2026-06-13`.

## Verification Results

```
bash -n build/docker/entrypoint.sh                  — PASS
grep -q 'CGO_ENABLED=0' build/docker/Dockerfile     — PASS
grep -q 'sha256:' build/docker/Dockerfile            — PASS
grep -q 'FROM golang:1.24' build/docker/Dockerfile  — PASS
grep -Eq 'pane.yaml|[*].age' .dockerignore           — PASS
Dockerfile sha256 == translators/arch/Dockerfile sha256  — MATCH

python3 yaml.safe_load both workflow YAMLs           — PASS (yaml-ok)
bash -n scripts/determinism-test.sh                  — PASS
bash -n scripts/secret-free-check.sh                 — PASS
bash -n scripts/check-coverage.sh                    — PASS
grep -q 'workflow_call' build/actions/build-speech.yml  — PASS
grep -q 'sha256sum' scripts/determinism-test.sh         — PASS
grep -q 'pane.yaml' scripts/secret-free-check.sh        — PASS
grep -Eq '85' scripts/check-coverage.sh                 — PASS

bash scripts/determinism-test.sh:
  Run 1 sha256: 99db28900f05b1230baaec16b3a729144fe8582f9a39b50c1835f2cc6d73b069
  Run 2 sha256: 99db28900f05b1230baaec16b3a729144fe8582f9a39b50c1835f2cc6d73b069
  DETERMINISM OK

bash scripts/secret-free-check.sh:
  pane.yaml absent, identity.age absent, private-injection.tar absent
  SECRET-FREE CHECK OK

bash scripts/check-coverage.sh:
  resolver: 93.5% >= 90%  — PASS
  cli:      85.6% >= 85%  — PASS
  ALL COVERAGE GATES PASSED

go test ./... -count=1: all packages PASS (no regressions)
```

## Deviations from Plan

### Auto-added functionality (Rule 2)

**1. [Rule 2 - Missing Coverage] Added cli/internal/loader/loader_test.go**
- **Found during:** Task 2 — running `bash scripts/check-coverage.sh` showed cli/ at 58.9% (loader at 0%)
- **Issue:** The plan required extending the coverage gate to cli >=85%, but the existing cli packages had no loader tests, pulling the aggregate far below 85%
- **Fix:** Created `cli/internal/loader/loader_test.go` with 12 tests covering all four loader functions across success and error paths
- **Files:** cli/internal/loader/loader_test.go
- **Commit:** 1044817

**2. [Rule 2 - Missing Coverage] Extended test files for pane, build, validate, runner**
- **Found during:** Task 2 — after adding loader tests, total cli coverage was 75.1%, then 81.8%, then 83.4%; needed additional error-path tests to reach 85%
- **Fix:** Added error-path and edge-case tests across cli/pane, cli/build, cli/validate, cli/runner packages; added `FakeRunnerFunc` to runner package for per-call error control
- **Files:** cli/pane/pane_test.go, cli/build/build_test.go, cli/validate/validate_test.go, cli/runner/fake.go, cli/runner/runner_test.go
- **Commit:** 1044817

**3. [Rule 2 - Security] GHA workflow input injection guard**
- **Found during:** Task 2 — security plugin flagged `${{ inputs.speech-dir }}` interpolated directly in `run:` shell
- **Fix:** Moved all `${{ inputs.* }}` and `${{ github.workspace }}` values to `env:` block; `run:` step uses only `${ENVVAR}` shell references
- **Files:** build/actions/build-speech.yml
- **Commit:** 1044817

## Security Properties Verified

| Control | Implementation | Test |
|---------|---------------|------|
| T-03-CTX | .dockerignore excludes pane.yaml/*.age/private-injection.tar/.config/ | grep -Eq 'pane.yaml' .dockerignore PASS |
| T-03-IMG | Dockerfile digest matches translators/arch/Dockerfile | sha256 equality check PASS |
| T-03-SECRETFREE | scripts/secret-free-check.sh greps arch-profile/ for secret files | PASSES: all 3 absent |
| T-03-DET | scripts/determinism-test.sh double-run sha256 compare | PASSES: identical hash |
| T-03-CIWF | GHA inputs via env vars in run:; read-only default permissions | code review + yaml syntax check |
| PRIV-01 | pane.yaml 0600; injection tar in outDir not profileDir; secret-free check | existing pane + build tests + secret-free-check.sh |

## Known Stubs

**1. Full mkarchiso ISO build (both channels):**
- `debateos build` (without `--skip-iso`) and the Docker entrypoint (without `SKIP_ISO=1`) invoke `mkarchiso` which requires `devtmpfs` access unavailable on this Proxmox VE host.
- The profile emission path (`--skip-iso`) is fully tested and all code paths up to the docker `r.Run("docker", ...)` call are covered.
- This is the same documented limitation as Phase 2 (ARCH-01/02 evidence notes).
- Resolution: verify on a bare-metal Linux host or a VM with devtmpfs.

**2. Live cross-repo GitHub Actions run:**
- The workflow YAML is syntactically valid (PyYAML validated) and follows official GHA workflow_call documentation syntax.
- A live run requires a fork of mikl0s/DebateOS with Actions enabled and available CI minutes.
- Steps documented in build/actions/README.md and docs/cli-build-channels.md.

## Threat Flags

No new security-relevant surface beyond the plan's threat model.

All T-03-* threat mitigations are in place:
- T-03-CTX: .dockerignore (implemented + verified by grep)
- T-03-IMG: digest pin identical to translators/arch/Dockerfile (verified by equality check)
- T-03-SECRETFREE: secret-free-check.sh (passes)
- T-03-CIWF: GHA inputs via env vars (implemented per security plugin guidance)
- T-03-DET: determinism-test.sh (passes)

## Self-Check: PASSED

- [x] build/docker/Dockerfile exists: /home/mikkel/repos/DebateOS/build/docker/Dockerfile
- [x] build/docker/Dockerfile contains CGO_ENABLED=0: grep PASS
- [x] build/docker/Dockerfile digest matches translators/arch/Dockerfile: sha256 MATCH
- [x] build/docker/entrypoint.sh exists and passes bash -n
- [x] .dockerignore exists and excludes pane.yaml/*.age
- [x] build/actions/build-speech.yml exists and YAML-parses cleanly
- [x] .github/workflows/build-speech.yml exists and YAML-parses cleanly
- [x] both workflows contain workflow_call / uses: respectively
- [x] scripts/determinism-test.sh exits 0 (sha256 identical)
- [x] scripts/secret-free-check.sh exits 0 (no secret files in profile)
- [x] scripts/check-coverage.sh exits 0 (resolver 93.5% ≥90%, cli 85.6% ≥85%)
- [x] go test ./... -count=1 — all packages PASS
- [x] docs/cli-build-channels.md >= 40 lines (346 lines)
- [x] docs/cli-build-channels.md contains 'private-injection.tar': PASS
- [x] docs/cli-build-channels.md contains 'no central service': PASS
- [x] REQUIREMENTS.md: CLI-01/02 + BLD-01..04 + PRIV-01 marked Complete
- [x] ROADMAP.md: Phase 3 marked 4/4 Complete 2026-06-13
- [x] commit 0220973 (Task 1) exists in git log
- [x] commit 1044817 (Task 2) exists in git log
- [x] commit 03bc03e (Task 3) exists in git log
