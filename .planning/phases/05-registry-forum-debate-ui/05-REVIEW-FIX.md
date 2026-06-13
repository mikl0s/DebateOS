---
phase: 05-registry-forum-debate-ui
fixed_at: 2026-06-13T17:20:00Z
review_path: .planning/phases/05-registry-forum-debate-ui/05-REVIEW.md
iteration: 1
findings_in_scope: 14
fixed: 14
skipped: 0
status: all_fixed
---

# Phase 05: Code Review Fix Report

**Fixed at:** 2026-06-13
**Source review:** .planning/phases/05-registry-forum-debate-ui/05-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 14
- Fixed: 14
- Skipped: 0

## Fixed Issues

### CR-01: Session and State Cookies Missing `Secure` Flag

**Files modified:** `forum/api/oauth.go`, `forum/api/internal_test.go`, `forum/api/oauth_test.go`
**Commit:** c13c8a0 (shared with CR-02)
**Applied fix:** Added `Secure: true` to all three `http.SetCookie` calls in oauth.go (loginHandler state cookie, callbackHandler clear cookie, callbackHandler session cookie). Test `TestOAuthCookiesHaveSecureFlag` added to oauth_test.go asserting both `Secure=true` and `HttpOnly=true` on `oauth_state` and `forum_session` cookies.

---

### CR-02: SessionStore Never Evicts Expired Entries

**Files modified:** `forum/api/oauth.go`, `forum/api/internal_test.go`
**Commit:** c13c8a0 (shared with CR-01)
**Applied fix:** `NewSessionStore` now launches a background goroutine (`sweepLoop`) that calls `sweepExpired()` every 5 minutes. `sweepExpired()` is also callable directly from tests for deterministic verification. Test `TestSweepExpiredRemovesExpiredEntries` injects already-expired entries directly into the maps and calls `sweepExpired()` synchronously, asserting they are removed and valid entries are preserved.

---

### CR-03: `getPoint` Path-Mangling Fallback Bypasses chi Router

**Files modified:** `forum/api/search.go`, `forum/api/coverage_test.go`
**Commit:** 129ad0b
**Applied fix:** Replaced `r.PathValue("id")` + path-slice fallback with `chi.URLParam(r, "id")`, consistent with `getRatings`. Returns 400 when id is empty. Added `chi/v5` import. Test `TestGetPointUsesChi` verifies the chi param extraction works end-to-end.

---

### CR-04: Foundation Filter Substring Injection Bypasses Scoping

**Files modified:** `forum/store/sqlite.go`, `forum/store/store_test.go`
**Commit:** c07a309
**Applied fix:** Replaced `strings.Contains(p.FoundationCompat, needle)` with `json.Unmarshal` + exact `id == foundation` comparison. Removed `strings` import, added `encoding/json`. Test `TestFoundationFilterInjectionRejected` verifies that foundation=`arch","debian` does NOT match an arch-only record, and that plain `arch` does match.

---

### WR-01: registry-index Workflow Silently Succeeds When Generator Is Missing

**Files modified:** `.github/workflows/registry-index.yml`
**Commit:** e5df4ac
**Applied fix:** Changed `exit 0` to `exit 1` when `scripts/generate-index.sh` does not exist. Changed message from WARNING to ERROR, directed to stderr. This makes CI fail loudly rather than silently pass.

---

### WR-02: `handleApplyPatch` Is a No-Op Stub That Misleads Users

**Files modified:** `web/src/routes/debate/+page.svelte`
**Commit:** 58c992f (shared with WR-04)
**Applied fix:** `handleApplyPatch` now sets `resolveError` to an explanatory message ("Patch application is not yet supported in v1. Add the patch opinion manually.") instead of silently calling `scheduleResolve()`. The error message surfaces in the UI where `resolveError` is displayed, giving users honest feedback.

---

### WR-03: `TruncateAll` Misleadingly Named

**Files modified:** `forum/query.sql`, `forum/store/generated/query.sql.go`, `forum/store/generated/querier.go`, `forum/store/generated/queries_test.go`, `forum/store/sqlite.go`
**Commit:** 5253169
**Applied fix:** Renamed `TruncateAll` to `TruncateConflictThreads` throughout — SQL source, generated Go code, Querier interface, and the test. Updated `SQLiteStore.Truncate` to call `TruncateConflictThreads` and added a comment documenting exactly what each call covers to prevent future maintainer confusion.

---

### WR-04: `debateAddTestPane`/`debateGetResolved` Globals Exposed in Production Builds

**Files modified:** `web/src/routes/debate/+page.svelte`
**Commit:** 58c992f (shared with WR-02)
**Applied fix:** Changed `if (typeof window !== 'undefined')` to `if (typeof window !== 'undefined' && import.meta.env.DEV)` so these test globals are only attached in Vite development/test builds. They are stripped in production builds by Vite's tree-shaking. E2e tests run in dev mode where they remain available.

---

### WR-05: `embed.go` Serves `http.FileServer` Without SPA Fallback

**Files modified:** `cli/embed/embed.go`, `cli/embed/embed_test.go`
**Commit:** fff2bca
**Applied fix:** Introduced `bufferingRecorder` that buffers the file server response. `NewUIHandler` now intercepts 404 responses from the file server and re-serves `404.html` (the SvelteKit SPA shell) using `http.ServeFileFS`. Buffering prevents partial 404-body bytes from reaching the connection before the SPA shell is written. Test `TestSPAFallback` verifies `/debate/`, `/export/`, `/browse/`, and an arbitrary unknown path all return 200+HTML (not a raw 404). Updated `TestWasmContentType` to handle the new SPA fallback behavior for missing `.wasm` files in placeholder builds.

---

### WR-06: `postConflict` Accepts Empty String `id`

**Files modified:** `forum/api/conflicts.go`, `forum/api/coverage_test.go`
**Commit:** 504da1e
**Applied fix:** Added `if req.ID == ""` check before the store call, returning 400. Tests `TestPostConflictEmptyIDRejected` and `TestPostConflictMissingIDRejected` verify both cases.

---

### IN-01: Reindex Calls `RebuildFTS` N+1 Times

**Files modified:** `forum/store/store.go`, `forum/store/sqlite.go`, `forum/reindex.go`, `forum/reindex_test.go`, `forum/api/coverage_test.go`
**Commit:** 0fea11e
**Applied fix:** Added `UpsertPointBatch` to the `Store` interface and `SQLiteStore`. It calls the raw SQL upsert via `upsertPointRaw` without rebuilding FTS. `Reindex()` now uses `UpsertPointBatch` in the loop and calls `s.Reindex(ctx)` once at the end: O(1) FTS rebuilds for N points. All mock `Store` implementations updated to implement the new method.

---

### IN-02: `TestTokenNotPersisted` Queries Only Hardcoded Tables

**Files modified:** `forum/api/oauth_test.go`
**Commit:** 41ee202
**Applied fix:** Replaced the hardcoded `[]string{"points","subscriptions","ratings","conflict_threads"}` with a `SELECT name FROM sqlite_master WHERE type='table' ...` query that enumerates all user tables dynamically, excluding FTS virtual tables. New tables are automatically covered without test changes.

---

### IN-03: `wasm.ts` `wasmReady` Boolean Has No Concurrency Guard

**Files modified:** `web/src/lib/wasm.ts`
**Commit:** aac6af5
**Applied fix:** Replaced `let wasmReady = false` with `let wasmPromise: Promise<void> | null = null`. `loadDebateosWasm` is now synchronous (non-async): it creates the Promise on first call and returns the same Promise on all subsequent calls. The actual async load logic moved to `_loadDebateosWasm`. All concurrent callers share the single Promise and cannot trigger duplicate `WebAssembly.instantiateStreaming` calls.

---

### IN-04: `export/+page.svelte` Shows Hardcoded Example YAML

**Files modified:** `web/src/lib/stores/speech.ts`, `web/src/routes/debate/+page.svelte`, `web/src/routes/export/+page.svelte`
**Commit:** 1651cd2
**Applied fix:** Added `resolvedSpeechStore` (`writable<ResolvedSpeech|null>`) to `speech.ts`. The debate page writes to it after every successful WASM resolve and clears it when panes are empty. The export page reads `$resolvedSpeechStore` reactively. When null (no speech resolved), it shows "No resolved speech yet." with a link to `/debate/` — honest copy instead of false "Your speech is ready". When resolved, `resolvedSpeechToYaml()` generates YAML from the actual resolver output (schema, foundation, applied, dropped, explanations) and the Download button downloads the real file.

---

## Skipped Issues

None — all 14 findings were fixed.

---

## Gate Results

All gates passed after fixes:

```
go test ./... -count=1                    PASS (all 25 packages)
npm run test:unit (vitest)                PASS (31 tests)
npm run test:e2e (playwright)             PASS (13 tests)
bash scripts/forum-offline-check.sh       PASS (Invariant-4)
bash scripts/dual-foundation-check.sh     PASS (DEB-02, 20/20 checks)
bash scripts/check-coverage.sh            PASS
  resolver/ 93.5% >= 90%
  cli/      86.3% >= 85%
  registry/ 85.4% >= 85%
  forum/    85.5% >= 85%
```

---

_Fixed: 2026-06-13_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
