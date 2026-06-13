---
phase: 05-registry-forum-debate-ui
verified: 2026-06-13T17:45:00Z
status: human_needed
score: 5/5
overrides_applied: 0
human_verification:
  - test: "Forum OAuth live round-trip (GitHub OAuth app required)"
    expected: "User navigates to /oauth/login, is redirected to GitHub, approves, returns to Forum with a session cookie set (HttpOnly + Secure + SameSite=Lax)"
    why_human: "Requires a registered GitHub OAuth app (client ID + secret) and a live HTTPS host. Fake provider covers code paths in tests; browser redirect flow is environment-blocked."
  - test: "GitHub Pages live deploy of the BASE_PATH=/debateos build"
    expected: "Pushing to gh-pages branch serves the debate UI at https://<org>.github.io/debateos/ with WASM loading and all routes reachable"
    why_human: "Requires GitHub Pages enablement on the repository — account/repo setting blocked in this environment."
  - test: "registry-index GitHub Actions workflow live run"
    expected: "A push to main that touches registry/** triggers the workflow, generate-index.sh runs, regenerated index.json is committed back via github-actions[bot]"
    why_human: "Requires CI runner minutes on GitHub Actions — environment-blocked. Workflow file is authored and correct (WR-01 fixed to exit 1 when script missing)."
  - test: "Oracle A1 live deployment (forum/deploy/oracle-a1.md)"
    expected: "Cross-compiled arm64 forumctl binary runs on an Oracle A1 instance, accepts HTTPS traffic, OAuth callback URL matches GITHUB_REDIRECT_URL, systemd unit survives restart"
    why_human: "Requires an Oracle Cloud account and A1 instance — environment-blocked. Deployment docs in forum/deploy/oracle-a1.md are present and correct (D15)."
deferred: []
---

# Phase 5: Registry, Forum & Debate UI — Verification Report

**Phase Goal:** Users can discover points, compose visually with live conflict resolution, and proceed to build — with Git authoritative and the Forum strictly optional

**Verified:** 2026-06-13T17:45:00Z
**Status:** human_needed (automated checks all passed; 4 live-deploy items require human/environment)
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Registry generator produces deterministic JSON index validated via resolver/parse; rebuilds on commit (Git authoritative) | VERIFIED | `go test ./registry/... -count=1` — 8 tests pass (TestGenerateIndex, TestDeterminism, TestGoldenIndex, TestEmitHTML, TestLoadCapabilities, TestComputeCompat/4 sub-tests). Golden file committed. WR-01 fix: exit 1 when generate-index.sh missing. |
| 2 | Debate UI composes with live WASM conflict viz; proceeds to build instructions; Forum-offline compose works (invariant 4) | VERIFIED | 13/13 Playwright e2e pass: A1 (triple encoding), A3 (WASM-only compose Forum-blocked), A6 (no forbidden terms), A9 (verbatim explanation text). `bash scripts/forum-offline-check.sh` exits 0: Applied=5 with Forum process absent. |
| 3 | Dual delivery: `debateos compose --serve` serves embedded UI offline AND BASE_PATH=/debateos build is Pages artifact | VERIFIED | `cli/compose/serve.go` wires `serveUI = embeddedui.ServeUI`. `cli/embed/web/` contains real SvelteKit build + `debateos.wasm`. `TestComposeServeFlag`, `TestSPAFallback` (4 deep-link paths), `TestWasmContentType` pass. `scripts/build-ui-dual.sh` authors both builds deterministically. |
| 4 | Forum search/subscribe/ratings/conflict-threads work; OAuth-only (fake provider in tests); re-indexable | VERIFIED | `go test ./forum/... -count=1` — all 25 test packages pass. TestSearchEndpoint, TestSubscriptionRoundTrip, TestRatingRequiresIdentity, TestConflictEndpoint (6 sub-tests), TestReindex, TestTokenNotPersisted. `forum/reindex.go` + `forumctl serve` provide rebuildable path. |
| 5 | Debate-themed brand voice consistent; forbidden terms absent from all visible UI text | VERIFIED | A6 e2e: 3 routes (/, /debate/, /export/) checked for "config", "preset", "distro", "package set" — 0 occurrences. `web/src/routes/export/+page.svelte` uses "Your speech is ready", "Settled the Debate", "Proceed to Build" — no forbidden terms. |

**Score:** 5/5 truths verified (automated)

---

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `registry/generator.go` | GenerateIndex entrypoint + HTML emit | VERIFIED | 115 lines; GenerateIndex, LoadCapabilities, EmitHTML all present and tested |
| `registry/index/index.go` | RegistryIndex, PointEntry structs + deterministic marshaling | VERIFIED | RegistryIndex{Schema, GeneratedAt, Points}, SchemaVersion = 1 |
| `registry/index/compat.go` | ComputeCompat from capabilities.json | VERIFIED | ComputeCompat + FoundationCompat exported; tested for exact + missing + sort |
| `go.mod` | All phase Go deps (chi v5.3.0, sqlite v1.46.1 PINNED, oauth2 v0.34.0) | VERIFIED | go 1.24.0 preserved; all three deps present as direct requires |
| `forum/api/oauth.go` | OAuth login/callback with Secure cookies, sweep loop | VERIFIED | CR-01: Secure=true on all 3 SetCookie calls (lines 264-272, 299-307, 342-350). CR-02: sweepLoop goroutine in NewSessionStore. |
| `forum/store/sqlite.go` | SearchPoints with exact JSON membership for foundation filter | VERIFIED | CR-04: json.Unmarshal + `id == foundation` exact match (lines 97-109). TestFoundationFilterInjectionRejected passes. |
| `forum/api/search.go` | getPoint uses chi.URLParam, no path-slice fallback | VERIFIED | CR-03: chi.URLParam(r, "id") with 400 on empty; TestGetPointUsesChi passes. |
| `web/src/routes/debate/+page.svelte` | WR-04: test globals gated on import.meta.env.DEV | VERIFIED | Line 144: `if (typeof window !== 'undefined' && import.meta.env.DEV)` — globals absent in production builds. |
| `web/src/routes/export/+page.svelte` | IN-04: export uses actual resolved speech from store | VERIFIED | Reads `$resolvedSpeechStore`; `resolvedSpeechToYaml()` serialises real WASM output; `data-testid="resolved-yaml"` renders real content. |
| `web/src/lib/stores/speech.ts` | resolvedSpeechStore shared between debate + export | VERIFIED | `export const resolvedSpeechStore = writable<ResolvedSpeech | null>(null)` at line end; debate page writes it; export page reads it. |
| `cli/embed/embed.go` | NewUIHandler with SPA fallback (WR-05) | VERIFIED | bufferingRecorder intercepts 404 → serves 404.html SPA shell; TestSPAFallback confirms /debate/, /export/, /browse/ all return 200. |
| `cli/embed/web/` | WASM + SvelteKit build committed | VERIFIED | `cli/embed/web/debateos.wasm` present; `_app/`, `404.html`, `index.html`, `wasm_exec.js` all present. |
| `cli/compose/compose.go` | `--serve` flag wired to serveUI | VERIFIED | `serve.go`: `var serveUI = embeddedui.ServeUI`; TestComposeServeNoListen passes. |
| `.github/workflows/registry-index.yml` | Rebuild workflow triggers on registry/** push | VERIFIED | WR-01 fix applied: exit 1 when generate-index.sh missing. Workflow triggers on push to main, paths: registry/**, supports workflow_dispatch. |
| `forum/deploy/oracle-a1.md` | D15 hosting docs present | VERIFIED | File exists with GOOS=linux GOARCH=arm64 build instructions, environment variable table, systemd unit, nginx TLS config. |
| `scripts/forum-offline-check.sh` | Invariant-4 gate | VERIFIED | Exits 0: compose --dir Applied=5; resolve-json Applied=5. No forum process required. |
| `scripts/build-ui-dual.sh` | Dual-build script (Pages + embed) | VERIFIED | Two-pass build: BASE_PATH=/debateos → dist/pages/; BASE_PATH= → cli/embed/web/. WASM built first. |

---

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `registry/generator.go` | `resolver/parse` | `parse.ParsePoint/ParseOpinion` | VERIFIED | Every YAML doc validated; malformed input → wrapped error naming file (T-05-01). TestGenerateIndexValidation passes. |
| `registry/index/compat.go` | `translators/*/capabilities.json` | `LoadCapabilities` → capability token membership | VERIFIED | ComputeCompat union algorithm; arch-only token → debian Missing. |
| `web/src/routes/debate/+page.svelte` | WASM resolver | `debateosResolve(JSON.stringify(input))` | VERIFIED | `runResolve()` calls WASM on every pane-change with 150ms debounce; A3 e2e proves Forum-blocked path works. |
| `web/src/routes/debate/+page.svelte` | `resolvedSpeechStore` | `resolvedSpeechStore.set(resolved)` | VERIFIED | Line 99 in debate page writes to store after every successful WASM resolve. |
| `web/src/routes/export/+page.svelte` | `resolvedSpeechStore` | `$resolvedSpeechStore` reactive | VERIFIED | Export page reads store reactively; `resolvedSpeechToYaml()` generates YAML from real WASM output (IN-04). |
| `cli/compose/serve.go` | `cli/embed/embeddedui.ServeUI` | `var serveUI = embeddedui.ServeUI` | VERIFIED | Direct function-variable assignment; TestComposeServeFlag injects no-listen seam. |
| `forum/api/oauth.go` | `SessionStore` | `createSession` → cookie | VERIFIED | Session cookie Secure=true; access token zeroed after GetUserID; token not stored (T-05-14). |
| `forum/reindex.go` | `forum/store.UpsertPointBatch` | IN-01 fix: batch upsert + single RebuildFTS | VERIFIED | `UpsertPointBatch` skips per-insert FTS; `s.Reindex(ctx)` called once at end. O(1) FTS rebuilds. |

---

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|--------------|--------|--------------------|--------|
| `web/src/routes/export/+page.svelte` | `resolved` / `yamlContent` | `$resolvedSpeechStore` ← `resolvedSpeechStore.set(resolved)` in debate page ← `debateosResolve(input)` WASM | Yes — WASM returns actual ResolvedSpeech from resolver engine | FLOWING |
| `web/src/routes/debate/+page.svelte` | `resolvedSpeech` / `conflictViews` | `debateosResolve(input)` WASM call | Yes — real WASM resolver (A3 e2e confirmed: `Applied.length > 0` with Forum blocked) | FLOWING |
| `forum/api/search.go` | search results | `store.SearchPoints` → SQLite FTS5 MATCH | Yes — parameterized FTS5 query against real SQLite; TestSearchEndpoint inserts + queries | FLOWING |
| `registry/generator.go` | `RegistryIndex` | `parse.ParsePoint/ParseOpinion` over YAML files | Yes — validated YAML → real PointEntry structs; golden test proves byte-identical output | FLOWING |

---

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| All Go packages compile and test GREEN | `go test ./... -count=1` | 25 packages pass (0 failures) | PASS |
| Vitest unit tests (wasm.ts, conflict.ts) | `cd web && npm run test:unit` | 31/31 pass in 254ms | PASS |
| Playwright e2e (A1/A3/A6/A9 + A2/A7) | `cd web && npm run test:e2e` | 13/13 pass in 4.1s | PASS |
| Invariant-4 offline gate | `bash scripts/forum-offline-check.sh` | INVARIANT-4 PASS: Applied=5 both steps | PASS |
| Coverage thresholds | `bash scripts/check-coverage.sh` | resolver/: 93.5% ≥ 90%; cli/: 86.3% ≥ 85%; registry/: 85.4% ≥ 85%; forum/: 85.5% ≥ 85% | PASS |
| CR-01: Secure cookies | `go test ./forum/api/... -run TestOAuthCookiesHaveSecureFlag` | PASS | PASS |
| CR-02: Sweep loop | `go test ./forum/api/... -run TestSweepExpiredRemovesExpiredEntries` | PASS | PASS |
| CR-03: chi URLParam | `go test ./forum/api/... -run TestGetPointUsesChi` | PASS | PASS |
| CR-04: Foundation exact match | `go test ./forum/store/... -run TestFoundationFilterInjectionRejected` | PASS | PASS |
| WR-04: DEV-gated globals | `grep "import.meta.env.DEV" web/src/routes/debate/+page.svelte` | Line 144 matches | PASS |
| IN-04: Export uses real speech | `grep "resolvedSpeechStore" web/src/routes/export/+page.svelte` | Imports and reads `$resolvedSpeechStore` reactively | PASS |

---

### Probe Execution

No `scripts/tests/probe-*.sh` conventional probes declared. Invariant-4 script (`scripts/forum-offline-check.sh`) run above: PASS.

---

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|---------|
| REG-01 | 05-01-PLAN.md | Registry static index: Git YAML → deterministic JSON, validates via resolver/parse, Pages-compatible HTML | SATISFIED | TestGenerateIndex, TestDeterminism, TestGoldenIndex, TestEmitHTML all GREEN; WR-01 CI fix applied |
| UI-01 | 05-02-PLAN.md, 05-04-PLAN.md | UI calls WASM, never reimplements resolution (invariant 3) | SATISFIED | A3 e2e: WASM resolves with Forum blocked; `conflict.ts` maps only; `wasm.ts` calls only; 31 unit tests GREEN |
| UI-02 | 05-02-PLAN.md, 05-04-PLAN.md | Live WASM conflict viz; dual delivery (Pages + offline serve) | SATISFIED | A1 triple-encoding e2e PASS; 13 Playwright tests pass; `debateos compose --serve` wired; embed assets committed including `.wasm` |
| BRND-01 | 05-04-PLAN.md | Debate-themed brand voice; no forbidden terms | SATISFIED | A6 e2e: 3 routes checked, 0 forbidden term occurrences; build stage labels use rhetoric names |
| FORM-01 | 05-03-PLAN.md, 05-05-PLAN.md | FTS5 search by curator/tag/popularity/freshness/foundation-compat | SATISFIED | TestSearchEndpoint, TestSearchWithFoundationFilter, TestSearchWithLimit, TestSearchEmptyResult pass; foundation exact-match filter (CR-04) |
| FORM-02 | 05-03-PLAN.md | Subscribe to curators/points | SATISFIED | TestSubscriptionRoundTrip, TestSubscriptionRequiresIdentity pass; AddSubscription/RemoveSubscription/GetSubscriptions wired |
| FORM-03 | 05-05-PLAN.md | Rate via GitHub OAuth identity (fake provider in tests) | SATISFIED | TestRatingRequiresIdentity, TestOAuthCookiesHaveSecureFlag, TestOAuthCallbackValidatesState, TestTokenNotPersisted all pass |
| FORM-04 | 05-05-PLAN.md | Conflict threads link patch-opinion PRs | SATISFIED | TestConflictEndpoint (6 sub-tests) pass; thread stores patchPRURL; GET/POST endpoints working |
| FORM-05 | 05-05-PLAN.md | Forum optional; DB loss → re-index recovers; D15 hosting docs | SATISFIED | `scripts/forum-offline-check.sh` PASS; TestReindex (7 sub-tests) pass; `forum/deploy/oracle-a1.md` present |

---

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| `web/src/routes/debate/+page.svelte:134` | handleApplyPatch | Shows "not yet supported" error message (WR-02 fix) | Info | User-visible but honest feedback; not a silent stub. Acceptable v1 limitation. |
| None | — | No TBD/FIXME/XXX in production code | — | Clean |
| None | — | No hardcoded empty returns in data paths | — | Clean |

No BLOCKER anti-patterns found. The `handleApplyPatch` function intentionally shows an error message (WR-02 fix applied) rather than silently re-resolving — this is the correct v1 behavior per the review fix.

---

### Human Verification Required

The four items below are environment-blocked live deployments explicitly documented as deferrals in `05-VALIDATION.md` (§Manual-Only / Deferred-to-host). All code-level logic is verified by tests. Human sign-off is required before milestone closure.

#### 1. GitHub OAuth Live Round-Trip

**Test:** Register a GitHub OAuth app, set `GITHUB_CLIENT_ID`/`GITHUB_CLIENT_SECRET`/`GITHUB_REDIRECT_URL`, run `forumctl serve`, navigate to `/oauth/login`
**Expected:** Browser redirects to GitHub authorization page; after approval, callback URL receives code + state; session cookie `forum_session` set with `HttpOnly; Secure; SameSite=Lax`; subsequent POST to `/api/ratings` succeeds with the session
**Why human:** Requires a GitHub OAuth app registration and HTTPS host. Fake `OAuthProvider` interface covers all code paths in tests (TestOAuthLoginRedirect, TestOAuthCallbackValidatesState, TestOAuthSessionHasUserID).

#### 2. GitHub Pages Live Deploy

**Test:** Push the `dist/pages/` output (BASE_PATH=/debateos build) to `gh-pages` branch; enable Pages on repo
**Expected:** `https://<org>.github.io/debateos/` serves the debate UI; WASM loads at `/debateos/debateos.wasm`; `/debateos/debate/` and `/debateos/export/` route correctly via 404.html SPA shell
**Why human:** Requires GitHub repository Pages enablement — account/repository setting. Build output correctness is verified locally by the dual-build script.

#### 3. registry-index Actions Workflow Live Run

**Test:** Push a commit that modifies `registry/` on `main` branch
**Expected:** `Rebuild Registry Index` workflow triggers; `scripts/generate-index.sh` runs; regenerated `registry/index.json` is committed back via `github-actions[bot]` with message `chore(registry): regenerate index.json [skip ci]`; workflow returns green
**Why human:** Requires GitHub Actions CI runner. The workflow file is correct (WR-01 fixed: `exit 1` when script missing) and has been authored at `.github/workflows/registry-index.yml`.

#### 4. Oracle A1 Live Deployment (D15)

**Test:** Cross-compile `GOOS=linux GOARCH=arm64 go build -o forumctl ./forum/cmd/forumctl`, SCP to A1 instance, deploy with systemd unit from `forum/deploy/oracle-a1.md`, verify service starts and OAuth callback URL is reachable
**Expected:** `forumctl serve` runs on ARM64, accepts HTTPS via nginx reverse proxy, `/health` returns 200, OAuth flow completes end-to-end
**Why human:** Requires Oracle Cloud account and A1 instance. Deployment docs are present and correct. Binary cross-compiles cleanly (`go build` verified locally).

---

### Gaps Summary

No automated gaps found. All 5 must-have truths verified. All 14 code review findings (CR-01..04, WR-01..06, IN-01..04) confirmed fixed in code by tests and direct code inspection. The 4 human verification items are live-deploy deferrals explicitly scoped as environment-blocked in the verification policy — not gaps in implementation.

---

_Verified: 2026-06-13T17:45:00Z_
_Verifier: Claude (gsd-verifier)_
_Mode: initial_
