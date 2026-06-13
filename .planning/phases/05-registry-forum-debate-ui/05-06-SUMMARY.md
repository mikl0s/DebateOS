---
phase: "05"
plan: "06"
subsystem: registry-forum-ui
tags: [embed, dual-delivery, coverage, offline-guard, ci, docs, requirements]
dependency_graph:
  requires: [05-01, 05-02, 05-03, 05-04, 05-05]
  provides: [milestone-v1.0-complete, registry-ci-workflow, forum-offline-guard, all-phase5-reqs-complete]
  affects: [cli/embed, cli/compose, scripts/check-coverage.sh, forum/api, forum/store, forum/migrations, forum/reindex, registry]
tech_stack:
  added:
    - "go:embed all:web directive for SvelteKit build in cli/embed"
    - "http.FileServer(http.FS(sub)) for embedded static serving"
    - "SvelteKit dual-delivery: BASE_PATH= (embed) + BASE_PATH=/debateos (Pages)"
    - "GitHub Actions workflow for registry index rebuild on commit"
    - "forum/api/internal_test.go: white-box package api tests for unexported symbols"
    - "failingStore mock (store.Store interface) for API 500 error path coverage"
    - "errStore mock (store.Store interface) for forum.Reindex error path coverage"
  patterns:
    - "Dual-delivery SvelteKit: same source, two BASE_PATH builds, two commit targets"
    - "Test seam: --no-listen flag on compose --serve prevents port binding in tests"
    - "Invalid port :99999 causes immediate http.ListenAndServe failure (sync, not async)"
    - "Closed-DB pattern: Open(:memory:) → New(db) → db.Close() → all store ops fail immediately"
    - "check_package_coverage now accepts variadic package args via array expansion"
key_files:
  created:
    - cli/embed/embed.go
    - cli/embed/embed_test.go
    - cli/embed/web/index.html  # placeholder + real SvelteKit build
    - cli/compose/serve.go
    - cli/compose/serve_test.go
    - scripts/build-ui-dual.sh
    - scripts/forum-offline-check.sh
    - .github/workflows/registry-index.yml
    - forum/deploy/oracle-a1.md
    - docs/registry-forum-ui.md
    - forum/api/coverage_test.go
    - forum/api/internal_test.go
    - forum/cmd/forumctl/forumctl_test.go
    - forum/deploy/oracle-a1.md
    - forum/migrations/migrate_test.go
    - forum/store/generated/queries_test.go
  modified:
    - cli/compose/compose.go  # --serve, --addr, --no-listen flags
    - cli/compose/serve_test.go  # error path tests
    - cli/embed/embed_test.go  # ServeUI invalid-port test
    - cmd/debateos/main.go  # updated usage string
    - forum/api/oauth_test.go  # callback branches + RealGitHubOAuth constructor
    - forum/reindex_test.go  # errStore mock + error path tests
    - forum/store/store_test.go  # closed-DB error path tests
    - registry/generator_test.go  # LoadCapabilities + GenerateIndexEmptyDirs
    - scripts/check-coverage.sh  # fixed forum gate array expansion bug + 4-gate threshold
    - .planning/REQUIREMENTS.md  # REG-01 Complete
decisions:
  - "Dual-delivery SvelteKit: BASE_PATH= for CLI embed so assets resolve at localhost root"
  - "go:embed all:web includes dotfiles (.nojekyll) for GitHub Pages compatibility"
  - "compose --serve uses package-level var serveUI as test seam to avoid port binding"
  - "Coverage gate fix: check_package_coverage now uses array expansion not string split"
  - "RealGitHubOAuth: NewRealGitHubOAuth+AuthCodeURL are testable (no network); Exchange+GetUserID are not (live GitHub API) — documented structural constraint"
  - "Closed-DB test pattern: Open+New+Close produces store that fails synchronously on all ops"
  - "Forum coverage threshold at 85% applied to testable core packages (excluding forumctl binary entrypoint)"
  - "REG-01 marked Complete: registry/index.json + CI workflow + docs satisfies the requirement"
metrics:
  duration: "~4h (continuation from prior session)"
  completed: "2026-06-13"
  tasks_completed: 2
  tasks_total: 2
  files_changed: 26
---

# Phase 05 Plan 06: go:embed + Dual-Delivery + Coverage Gates + Milestone Close Summary

**One-liner:** go:embed dual-delivery SvelteKit (BASE_PATH= embed, BASE_PATH=/debateos Pages), compose --serve offline UI, four-gate coverage all passing (resolver 93.5%, cli 85.8%, registry 85.4%, forum 85.3%), invariant-4 offline guard, GitHub Actions registry CI, Oracle A1 deployment notes, end-to-end docs, REG-01 Complete — v1.0 milestone closed.

## Tasks Completed

### Task 1: go:embed Dual-Delivery + compose --serve

**Commit:** `3586746`

- **`cli/embed/embed.go`**: Package `embeddedui` with `//go:embed all:web`, `NewUIHandler()` returning `http.FileServer(http.FS(sub))` via `fs.Sub(WebFS,"web")`, and `ServeUI(addr) error` calling `http.ListenAndServe`.
- **`cli/embed/web/`**: Real SvelteKit build committed (BASE_PATH=); contains `debateos.wasm` (4.5 MB), `wasm_exec.js`, `index.html`, `_app/`, `browse/`, `export/`, `.nojekyll`. NOT gitignored — these ARE the embed artifacts.
- **`cli/compose/compose.go`**: Added `--serve bool`, `--addr string` (default `:8080`), `--no-listen bool` flags. When `--serve`: serves embedded UI (blocking). `--no-listen` is the test seam that avoids port binding.
- **`cli/compose/serve.go`**: `var serveUI = embeddedui.ServeUI` package-level var enables test injection.
- **`scripts/build-ui-dual.sh`**: Two SvelteKit builds — Pages (BASE_PATH=/debateos → dist/pages/) and embed (BASE_PATH= → cli/embed/web/). `SKIP_PAGES=1` skips the Pages build. T-05-18 mitigation documented.
- **TDD**: RED commit `d098e1b` (failing tests) before GREEN commit `3586746` (implementation).

### Task 2: Coverage Gates + Offline Guard + CI + Docs + REG-01

**Commit:** `9816f89`

**Coverage (all four gates now passing):**
- `resolver/`: 93.5% ≥ 90% ✓ (unchanged)
- `cli/`: 85.8% ≥ 85% ✓ (was 84.6%; fixed by ServeUI invalid-port test + compose --serve error path test)
- `registry/`: 85.4% ≥ 85% ✓ (TestLoadCapabilities + TestGenerateIndexEmptyDirs added)
- `forum/`: 85.3% ≥ 85% ✓ (was 0% due to script bug; fixed + 60+ new tests added)

**Coverage script fix**: `check-coverage.sh` passed the forum package list as a single space-separated string, which `go test` rejected as "malformed import path". Fixed by changing `check_package_coverage` to accept variadic package args and expanding the array with `"${PKGS[@]}"`.

**New test coverage patterns added:**
- `forum/api/internal_test.go`: white-box tests for `jsonDecodeBody` (unexported), `SessionStore.createState/validateState/createSession`, `identityFnFromSessions`
- `forum/api/oauth_test.go`: `NewRealGitHubOAuth`+`AuthCodeURL` (no network needed), `callbackHandler` branches (missing cookie, missing code, exchange error, getUserID error)
- `forum/api/coverage_test.go`: `failingStore` mock implementing `store.Store` for all 500 error paths in handlers
- `forum/store/store_test.go`: closed-DB pattern for all store error paths
- `forum/reindex_test.go`: `errStore` mock for `Reindex` UpsertPoint error + FTS rebuild error
- `forum/migrations/migrate_test.go`: `TestApplyClosedDB` for error return path
- `cli/embed/embed_test.go`: `TestServeUIInvalidAddr` — port 99999 fails synchronously
- `cli/compose/serve_test.go`: `TestComposeServeErrorPath` — covers `serveUI` error branch

**Structural constraint documented:** `RealGitHubOAuth.Exchange` and `GetUserID` remain at 0% — they require live GitHub API calls (`r.cfg.Exchange` calls the hardcoded `github.Endpoint.TokenURL`). `NewRealGitHubOAuth` and `AuthCodeURL` ARE covered (constructor + URL formatting, no network).

**Forum offline invariant guard (`scripts/forum-offline-check.sh`):**
- Kills any `forumctl` processes
- Runs `debateos compose` on `examples/dual-foundation/`
- Runs `resolve-json` on the same directory
- Asserts Applied≥1 in JSON output
- PASSES: Applied=5 for both compose and resolve-json with no Forum running
- T-05-19 (Invariant-4) mitigated and verified

**GitHub Actions (`/.github/workflows/registry-index.yml`):**
- Triggers on push to `main` touching `registry/**`
- Runs `scripts/generate-index.sh` to rebuild `registry/index.json`
- Commits back with `[skip ci]` to avoid loop
- `workflow_dispatch` for manual trigger
- Live run deferred to CI host (requires runner with Go 1.24+)

**Deployment (`forum/deploy/oracle-a1.md`):**
- Cross-compile: `GOOS=linux GOARCH=arm64 go build -o forumctl ./forum/cmd/forumctl`
- Environment variables: `FORUM_DSN`, `FORUM_ADDR`, `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `GITHUB_REDIRECT_URL`
- systemd unit with `EnvironmentFile=/etc/forum/env` (0600, no secrets in unit file)
- nginx reverse proxy with TLS via Certbot
- OCI Security List and iptables firewall rules
- DB recovery: `forumctl reindex --registry registry/index.json`

**End-to-end documentation (`docs/registry-forum-ui.md`):**
- Core vocabulary table (Opinion/Point/Speech/Debate/Foundation/Curator/Translator)
- Registry index structure and adding new points
- Forum optional status (Invariant-4), GitHub OAuth setup, security properties
- Embedded UI: `debateos compose --serve`, dual-delivery strategy, rebuild instructions
- Operator journey table (local preview → offline browse → social Forum → recovery → CI)

**REQUIREMENTS.md:** REG-01 flipped to Complete with evidence. All 9 Phase-5 requirements now Complete: UI-01, UI-02, BRND-01, FORM-01, FORM-02, FORM-03, FORM-04, FORM-05, REG-01.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed check-coverage.sh forum gate (malformed import path)**
- **Found during:** Task 2 coverage verification
- **Issue:** `FORUM_TESTABLE_PKGS` array was joined into a single space-separated string and passed as one argument to `go test`, producing "malformed import path" error (packages were rejected as `github.com/.../forum forum/api ...` with spaces — invalid)
- **Fix:** Changed `check_package_coverage` signature to accept variadic package args; passes array with `"${PKGS[@]}"` so `go test` receives each as a separate argument
- **Files modified:** `scripts/check-coverage.sh`
- **Commit:** `9816f89`

**2. [Rule 2 - Missing Coverage] Added tests for jsonDecodeBody, SessionStore internals**
- **Found during:** Task 2 coverage analysis
- **Issue:** `forum/api/util.go:jsonDecodeBody` was at 0% because it is only called from `oauth.go:GetUserID` (live GitHub API path, untestable). Coverage tests using POST /api/ratings exercised `json.NewDecoder` directly, not `jsonDecodeBody`.
- **Fix:** Added `forum/api/internal_test.go` (package api, not api_test) to test `jsonDecodeBody` and other unexported SessionStore helpers directly
- **Files modified:** `forum/api/internal_test.go` (new)
- **Commit:** `9816f89`

**3. [Rule 2 - Missing Coverage] Added failingStore + errStore mocks for error paths**
- **Found during:** Task 2 — HTTP 500 paths in API handlers and error paths in forum.Reindex were unreachable via real SQLiteStore in normal operation
- **Fix:** `failingStore` (in coverage_test.go) implements `store.Store` and returns configured errors; `errStore` (in reindex_test.go) same pattern for `forum.Reindex`
- **Files modified:** `forum/api/coverage_test.go`, `forum/reindex_test.go`
- **Commit:** `9816f89`

### Structural Constraints (not deviations — documented design limits)

- **`RealGitHubOAuth.Exchange` and `GetUserID` at 0%:** These methods call `r.cfg.Exchange` (uses `github.Endpoint.TokenURL` hardcoded to `https://github.com/login/oauth/access_token`) and `httpClient.Do(req)` to `https://api.github.com/user`. No mocking is possible without changing the production code to accept an HTTP client parameter. Coverage is excluded from the threshold by keeping `NewRealGitHubOAuth` + `AuthCodeURL` covered (URL-formatting only, no network), achieving 85.8% api package coverage despite the two 0% methods.
- **Live deployments deferred to CI host:** GitHub Pages deploy, GitHub Actions live run, Oracle A1 SSH deploy — all require infrastructure not present in the dev environment. All are documented in the plan as "deferred-to-host/CI" per the environment_notes.

## Known Stubs

None. All plan deliverables are implemented or documented as deferred-to-CI (per environment_notes which explicitly lists live GitHub OAuth / Pages deploy / Oracle deploy / Actions run as host-only).

## Threat Flags

No new threat surface introduced in this plan. All new HTTP handlers are read-only (serve static files). The `scripts/forum-offline-check.sh` runs only `go run` + `python3 -c` with no untrusted input. The GitHub Actions workflow does not use any `${{ github.event.* }}` expressions in `run:` steps (all actions use hardcoded commands), consistent with the security guidance.

## Self-Check: PASSED

All key files found on disk:
- cli/embed/embed.go ✓
- cli/embed/web/index.html ✓
- cli/compose/serve.go ✓
- scripts/build-ui-dual.sh ✓
- scripts/forum-offline-check.sh ✓
- .github/workflows/registry-index.yml ✓
- forum/deploy/oracle-a1.md ✓
- docs/registry-forum-ui.md ✓
- forum/api/internal_test.go ✓
- forum/api/coverage_test.go ✓

Commits verified:
- `d098e1b`: TDD RED (failing tests for embed + compose --serve)
- `3586746`: Task 1 (go:embed dual-delivery + compose --serve + dual-build script)
- `9816f89`: Task 2 (coverage gates + offline guard + CI + docs + REG-01)

All 22 Go test packages pass. All four coverage gates pass (resolver 93.5%, cli 85.8%, registry 85.4%, forum 85.3%). Invariant-4 offline guard passes (Applied=5 with Forum offline).
