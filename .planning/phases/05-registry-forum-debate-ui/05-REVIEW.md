---
phase: 05-registry-forum-debate-ui
reviewed: 2026-06-13T00:00:00Z
depth: standard
files_reviewed: 34
files_reviewed_list:
  - registry/generator.go
  - registry/generator_test.go
  - registry/index/compat.go
  - registry/index/compat_test.go
  - registry/index/index.go
  - forum/store/store.go
  - forum/store/sqlite.go
  - forum/store/inmem.go
  - forum/store/store_test.go
  - forum/store/generated/query.sql.go
  - forum/migrations/001_init.sql
  - forum/api/router.go
  - forum/api/search.go
  - forum/api/ratings.go
  - forum/api/oauth.go
  - forum/api/conflicts.go
  - forum/api/points.go
  - forum/api/util.go
  - forum/api/api_test.go
  - forum/api/coverage_test.go
  - forum/api/conflicts_test.go
  - forum/api/internal_test.go
  - forum/api/oauth_test.go
  - forum/reindex.go
  - forum/cmd/forumctl/main.go
  - web/src/lib/wasm.ts
  - web/src/lib/types.ts
  - web/src/lib/conflict.ts
  - web/src/lib/stores/speech.ts
  - web/src/lib/wasm.test.ts
  - web/src/lib/conflict.test.ts
  - web/src/routes/debate/+page.svelte
  - web/src/routes/export/+page.svelte
  - web/src/lib/components/ConflictOverlay.svelte
  - cli/embed/embed.go
  - cli/compose/compose.go
  - scripts/forum-offline-check.sh
  - scripts/build-ui-dual.sh
  - scripts/build-wasm.sh
  - scripts/check-coverage.sh
  - .github/workflows/registry-index.yml
findings:
  critical: 4
  warning: 6
  info: 4
  total: 14
status: issues_found
fixes_applied: true
resolved_ids:
  - CR-01
  - CR-02
  - CR-03
  - CR-04
  - WR-01
  - WR-02
  - WR-03
  - WR-04
  - WR-05
  - WR-06
  - IN-01
  - IN-02
  - IN-03
  - IN-04
---

# Phase 05: Code Review Report

**Reviewed:** 2026-06-13
**Depth:** standard
**Files Reviewed:** 34+
**Status:** issues_found

## Summary

This phase delivers the Forum HTTP API (OAuth, search, ratings, conflicts), the static registry index generator, the WASM loader + conflict mapping in the Svelte UI, and the embedded serve plumbing. The core security architecture is sound: OAuth state validation is correctly double-checked (cookie vs. URL parameter), access tokens are genuinely discarded after user-ID lookup, session IDs are stored only in memory, all SQL is parameterized, and write endpoints consistently require identity. Invariant 3 (UI never resolves) is respected — `conflict.ts` only maps, `wasm.ts` only calls. Invariant 4 (Forum-offline compose) is genuinely proven by `forum-offline-check.sh`.

Four BLOCKER-level issues require fixing before ship: missing `Secure` flag on session cookies (HTTPS-critical), an unbounded in-memory session store that never evicts expired entries, a path-mangling fallback in `getPoint` that bypasses chi routing and can expose unintended IDs under path-prefix deployments, and a `foundation` filter parameter that accepts arbitrary user strings and constructs a JSON needle with them — enabling a trivial filter bypass via injection.

---

## Critical Issues

### CR-01: Session and State Cookies Missing `Secure` Flag

**File:** `forum/api/oauth.go:228-235` (state cookie), `forum/api/oauth.go:303-311` (session cookie)

**Issue:** Both the `oauth_state` and `forum_session` cookies are set without `Secure: true`. In the production deployment (Oracle A1 behind HTTPS, `forum/deploy/oracle-a1.md`), the absence of `Secure` means browsers will transmit these cookies over plain HTTP if any non-HTTPS resource is requested or if the load-balancer issues a plain-HTTP redirect. A network-position attacker or a co-located actor can steal the session cookie over HTTP, impersonating any logged-in user. `HttpOnly` protects against XSS but not network interception. `SameSite=Lax` reduces CSRF risk but does not substitute for `Secure`.

**Fix:** Add `Secure: true` to all three `http.SetCookie` calls in `oauth.go`. For localhost dev/test use (where HTTPS is absent) the flag can be gated on an `httpsMode bool` parameter passed at router construction time, or unconditionally set in production (the Go http stack sends `Secure` cookies over HTTPS only; on localhost it is silently ignored by most browsers):

```go
http.SetCookie(w, &http.Cookie{
    Name:     stateCookieName,
    Value:    state,
    Path:     "/",
    MaxAge:   int(stateTTL.Seconds()),
    HttpOnly: true,
    Secure:   true, // ADD THIS
    SameSite: http.SameSiteLaxMode,
})
```

Apply the same change to the session cookie set in `callbackHandler` (line ~303) and the state-clear cookie (line ~261).

---

### CR-02: SessionStore Never Evicts Expired Entries — Unbounded Memory Growth

**File:** `forum/api/oauth.go:126-197`

**Issue:** `SessionStore.sessions` and `SessionStore.states` maps grow indefinitely. Expired entries are checked on read (`validateState` checks `time.Now().Before(exp)`, `GetUserID` checks `expiresAt`) but are never deleted. A state nonce that expires without being consumed stays in `states` forever. A session cookie that expires stays in `sessions` forever. In a long-running production instance (24-hour session TTL), any user who logs in and never returns leaves a `sessionEntry` that is never removed. Under sustained load (many users, service uptime of weeks/months) this constitutes a memory leak that can exhaust process memory.

This is a correctness and reliability defect: the expired state map also leaks information about past OAuth flows for the process lifetime, violating the intent of short-lived nonces.

**Fix:** Add a background goroutine launched from `NewSessionStore` that periodically sweeps both maps and deletes entries past their expiry:

```go
func NewSessionStore() *SessionStore {
    ss := &SessionStore{
        sessions: make(map[string]sessionEntry),
        states:   make(map[string]time.Time),
    }
    go ss.sweepLoop()
    return ss
}

func (ss *SessionStore) sweepLoop() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    for range ticker.C {
        now := time.Now()
        ss.mu.Lock()
        for k, exp := range ss.states {
            if now.After(exp) {
                delete(ss.states, k)
            }
        }
        for k, e := range ss.sessions {
            if now.After(e.expiresAt) {
                delete(ss.sessions, k)
            }
        }
        ss.mu.Unlock()
    }
}
```

---

### CR-03: `getPoint` Path-Mangling Fallback Bypasses chi Router

**File:** `forum/api/search.go:66-70`

**Issue:** When `r.PathValue("id")` returns an empty string, the handler falls back to slicing `r.URL.Path` directly:

```go
id = r.URL.Path[len("/api/points/"):]
```

This fallback has two problems:

1. **Incorrect under any deployment with a path prefix.** If a reverse proxy or the base-path embed wrapper adds a prefix, `r.URL.Path` could be `/prefix/api/points/foo`, making `r.URL.Path[len("/api/points/"):]` produce `prefix/api/points/foo` — a nonsensical ID that will 404 via `GetPoint` instead of finding the correct record.

2. **Dead code that signals a misunderstanding.** With chi registered as `r.Get("/api/points/{id}", h.getPoint)`, chi will always populate the `{id}` URL parameter when the route matches. The fallback can only trigger if the handler is called outside of chi's routing — which should never happen. The slice fallback therefore creates an illusion of safety while hiding routing misconfiguration.

**Fix:** Remove the fallback entirely. If `id` is empty after `chi.URLParam` (which is the correct way to extract chi path params, not `r.PathValue`), return 400:

```go
func (h *handlers) getPoint(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    if id == "" {
        http.Error(w, "point id required", http.StatusBadRequest)
        return
    }
    // ... rest of handler unchanged
}
```

Note: `r.PathValue("id")` is the standard library's Go 1.22+ path param accessor, while chi uses `chi.URLParam(r, "id")`. The current code uses `r.PathValue("id")` first (which returns empty string when used with chi), then falls back to the path slice hack. Use `chi.URLParam` exclusively, consistent with `getRatings` (line 59 of `ratings.go`).

---

### CR-04: Foundation Filter Substring Injection Bypasses Scoping

**File:** `forum/store/sqlite.go:93-100`

**Issue:** The `foundation` query parameter from the HTTP request is used to build a JSON-substring needle without validation or escaping:

```go
needle := `"` + foundation + `"`
for _, p := range results {
    if strings.Contains(p.FoundationCompat, needle) {
```

The stored `FoundationCompat` field is a JSON array like `["arch","debian"]`. The filter is intended to match only exact foundation IDs. However, because the check is a plain `strings.Contains` on a raw JSON string, a caller who passes `foundation=arch","debian` constructs:

```
needle = `"arch","debian"`
```

which matches any record whose compat string contains both `arch` and `debian` adjacently in the JSON serialization — circumventing the intended single-foundation filter. Worse, `foundation="}; DROP TABLE -- ` does not cause SQL injection (the DB query is complete at this point), but `foundation=a` matches `"arch"` (the `a` from `arch`) and `"java"` in any hypothetical entry, since `strings.Contains('"arch"', '"a"')` is false but `strings.Contains('"arcana"', '"a"')` would also be false — still, the exact boundary depends on the JSON encoding of actual data and is fragile.

A more direct bypass: passing `foundation=` (empty string) bypasses the filter entirely (handled), but `foundation=arch` will also match `"arch-extra"` if such a value is ever stored.

The real correctness problem is the substring check treats JSON as an opaque string rather than using proper JSON parsing to check membership.

**Fix:** Parse the `FoundationCompat` JSON field and check for exact membership:

```go
if foundation != "" {
    filtered := results[:0]
    for _, p := range results {
        var ids []string
        if err := json.Unmarshal([]byte(p.FoundationCompat), &ids); err == nil {
            for _, id := range ids {
                if id == foundation {
                    filtered = append(filtered, p)
                    break
                }
            }
        }
    }
    results = filtered
}
```

Alternatively, add a `foundation` column to the `points` table and filter in SQL.

---

## Warnings

### WR-01: registry-index Workflow Silently Succeeds When Generator Is Missing

**File:** `.github/workflows/registry-index.yml:51-57`

**Issue:** When `scripts/generate-index.sh` does not exist, the workflow step prints a warning and then runs `exit 0` — the job succeeds with a green check. This means CI gives a false pass when the index has not actually been regenerated. Pushes that modify registry files will appear to have succeeded the index-rebuild check when in reality nothing ran.

```yaml
echo "WARNING: scripts/generate-index.sh not found — index not regenerated"
exit 0   # <-- silently passes
```

**Fix:** Change to `exit 1` so the workflow fails loudly when the generator is absent, forcing the developer to notice and add the missing script:

```yaml
echo "ERROR: scripts/generate-index.sh not found — index cannot be regenerated" >&2
exit 1
```

---

### WR-02: `handleApplyPatch` in Debate Page Is a No-Op Stub That Misleads Users

**File:** `web/src/routes/debate/+page.svelte:130-136`

**Issue:** The `handleApplyPatch` function does nothing except reschedule a resolve:

```typescript
function handleApplyPatch(_patchId: string) {
    // ...For now: trigger a re-resolve which will show the updated state.
    scheduleResolve();
}
```

The "Apply Patch" button is rendered in `ConflictOverlay.svelte` whenever `view.hasPatch` is true (line 133 of `ConflictOverlay.svelte`). Clicking it calls `handleApplyPatch` → `scheduleResolve()` → the resolver runs again with the identical state → the same conflict overlay reappears. From the user's perspective the button appears to do nothing, which is confusing and could be interpreted as a broken UI.

The comment in the code acknowledges this as a v1 deferral, but there is no UI feedback informing the user that patch application is not yet supported. The button is misleading: it looks actionable but produces no visible change.

**Fix:** Either disable/hide the Apply Patch button in v1 with an explanatory tooltip ("Patch application coming soon"), or show a visible feedback message on click:

```typescript
function handleApplyPatch(_patchId: string) {
    // v1: patch application requires registry lookup (deferred).
    // Show user feedback rather than silently re-resolving.
    resolveError = 'Patch application is not yet supported in v1. Add the patch opinion manually.';
}
```

---

### WR-03: `Truncate` in `sqlite.go` Has Wrong Deletion Order — FK Violation Risk

**File:** `forum/store/sqlite.go:272-284`

**Issue:** The `Truncate` method comment states "Must delete in FK order: ratings, subscriptions, conflict_threads, then points." The implementation calls them in this sequence:

1. `TruncateRatings` — deletes from `ratings` (FK child of `points`) ✓
2. `TruncateSubscriptions` — deletes from `subscriptions` (FK child of `points`) ✓
3. `TruncateAll` — deletes from `conflict_threads` ✓
4. `TruncatePoints` — deletes from `points` ✓

This is the correct order. However, the generated query named `TruncateAll` (`query.sql.go:256`) deletes only from `conflict_threads`, not from all tables as its name implies. The name is misleading: a future developer adding a new FK-child table might assume `TruncateAll` covers it and not add a separate truncation step.

Additionally, if FK enforcement is ON (which it is — `PRAGMA foreign_keys=ON` in `Open`), attempting to delete from `points` while child rows remain will fail. Steps 1–3 prevent this, but only so long as `TruncateAll` correctly covers all children. The current code survives because `conflict_threads` has no FK back to `points`, but the pattern is fragile.

**Fix:** Rename `TruncateAll` in the generated query and `query.sql` source to `TruncateConflictThreads` to make the scope explicit, and add a comment in `Truncate` that lists what each call covers. This is a maintainability fix to prevent future correctness bugs.

---

### WR-04: `debateAddTestPane` and `debateGetResolved` Globals Exposed in Production Builds

**File:** `web/src/routes/debate/+page.svelte:140-147`

**Issue:** Two test-only globals are unconditionally attached to `window` in the production Svelte component:

```typescript
if (typeof window !== 'undefined') {
    (window as any).debateAddTestPane = ...
    (window as any).debateGetResolved = ...
}
```

The `typeof window !== 'undefined'` guard prevents SSR errors but does not guard against production exposure. SvelteKit's static build includes these functions in the shipped JS bundle. Any user with browser console access can call `window.debateAddTestPane(...)` to inject arbitrary opinion data into the debate store and observe the resolver's internal state via `window.debateGetResolved()`.

This violates the principle of least exposure: internal test seams should not be visible to end users. While the resolver itself is WASM-based and these functions cannot execute arbitrary code on the server, they allow users to feed crafted opinion data into the resolver and observe outputs that the UI would otherwise not show them — a privacy-adjacent concern and an undocumented attack surface on the WASM resolver.

**Fix:** Gate these assignments on a Vite dev/test mode check:

```typescript
if (typeof window !== 'undefined' && import.meta.env.DEV) {
    (window as any).debateAddTestPane = ...
    (window as any).debateGetResolved = ...
}
```

For Playwright e2e tests running against a production-mode build, use `page.exposeFunction()` or `page.addInitScript()` in the test setup instead of relying on globally exposed functions.

---

### WR-05: `embed.go` Serves `http.FileServer` Without SPA Fallback — Deep Links Return 404

**File:** `cli/embed/embed.go:33-41`

**Issue:** `NewUIHandler` returns `http.FileServer(http.FS(sub))` directly. The SvelteKit build with `adapter-static` and `fallback: '404.html'` generates a `404.html` in the build root that acts as the SPA entry point for client-side routes (`/debate/`, `/export/`, `/browse/`). However, Go's `net/http.FileServer` does not special-case `404.html`: when a request arrives for `/debate/` (a route with no pre-rendered static file), the file server returns HTTP 404 (with body from the server's default 404 page, not `404.html`).

The practical effect: `debateos compose --serve` opens at `http://localhost:8080/`, which works. But if a user pastes or bookmarks `http://localhost:8080/debate/` directly, they get a 404 from Go's file server instead of the SPA shell.

**Fix:** Wrap `http.FileServer` with a custom handler that serves `404.html` for any file-not-found response:

```go
func NewUIHandler() http.Handler {
    sub, err := fs.Sub(WebFS, "web")
    if err != nil {
        panic("embeddedui: fs.Sub failed: " + err.Error())
    }
    fileServer := http.FileServer(http.FS(sub))
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Try the file server; intercept 404 → serve 404.html (SPA shell).
        rr := &responseRecorder{ResponseWriter: w, code: http.StatusOK}
        fileServer.ServeHTTP(rr, r)
        if rr.code == http.StatusNotFound {
            r2 := r.Clone(r.Context())
            r2.URL.Path = "/404.html"
            http.ServeFileFS(w, r2, sub, "404.html")
        }
    })
}
```

---

### WR-06: `postConflict` Accepts Empty String `id` — Silently Creates Untargetable Thread

**File:** `forum/api/conflicts.go:66-93`

**Issue:** The `postConflictRequest.ID` field is optional with no validation. When `req.ID` is empty string (`""`), the handler passes `thread.ID = ""` to `store.UpsertConflictThread`. In SQLite, `id TEXT PRIMARY KEY` accepts an empty string as a valid row key. A subsequent `ON CONFLICT(id) DO UPDATE` with `id=""` will update that row rather than creating a new one. This means all conflict threads submitted with an empty ID collide into a single row in the database. There is no way for the client to individually target or retrieve a thread with `id=""` (the GET endpoint filters by `(point_a, point_b)` pairs, not by ID, so this is survivable — but the data model is corrupted and the intent of `id` as a stable external reference is violated).

**Fix:** Generate a UUID/random ID server-side when `req.ID` is empty, rather than accepting the empty string:

```go
if req.ID == "" {
    b := make([]byte, 8)
    if _, err := rand.Read(b); err != nil {
        http.Error(w, "failed to generate thread ID", http.StatusInternalServerError)
        return
    }
    req.ID = hex.EncodeToString(b)
}
```

Or reject the request with 400 and require the client to supply a stable ID.

---

## Info

### IN-01: Reindex Calls `RebuildFTS` N+1 Times — Inefficient for Large Indexes

**File:** `forum/reindex.go:35-52`

**Issue:** `forum.Reindex` calls `UpsertPoint` for each point, and `UpsertPoint` calls `RebuildFTS` after every single insert. For N points, this triggers N full FTS5 index rebuilds, each of which scans the entire `points` table. After the loop, `s.Reindex(ctx)` calls `RebuildFTS` one more time, making it N+1 total rebuilds. For the current small fixture size this is harmless, but as the registry grows this becomes quadratic in the worst case.

The comment in `reindex.go` acknowledges the belt-and-suspenders approach but does not address the per-insert redundancy.

**Fix:** Add a `UpsertPointRaw` variant to the store interface that skips the per-insert FTS rebuild, allowing the caller to do one final rebuild after all upserts. Alternatively, suppress the FTS rebuild inside `UpsertPoint` by wrapping the Reindex loop in a transaction that defers the FTS rebuild until commit.

---

### IN-02: `TestTokenNotPersisted` Queries Tables With Raw SQL — Fragile Test

**File:** `forum/api/oauth_test.go:194-228`

**Issue:** The test constructs raw SQL table names via string concatenation:

```go
tables := []string{"points", "subscriptions", "ratings", "conflict_threads"}
for _, tbl := range tables {
    rows, err := db.QueryContext(context.Background(), "SELECT * FROM "+tbl)
```

This hardcodes the schema table list in the test. If a new table is added that could store token material (e.g., an audit log), the test will not catch it. The test verifies the negative ("token does not appear") but only over the tables the test author thought to list.

Additionally `SELECT *` in tests is fragile — it will fail silently if new columns change the expected scan shape (though here it just reads into `[]interface{}` so it's tolerable).

**Fix:** Either query `sqlite_master` to enumerate all user tables dynamically, or add a comment noting that new tables that might store user-identity-adjacent data must be added to this list.

---

### IN-03: `wasm.ts` Module-Level `wasmReady` Singleton Has No Concurrency Guard

**File:** `web/src/lib/wasm.ts:35-64`

**Issue:** `wasmReady` is a module-level boolean. `loadDebateosWasm` checks and sets it without any synchronization:

```typescript
if (wasmReady) return;
// ... async work ...
wasmReady = true;
```

If `loadDebateosWasm` is called a second time before the first `await` chain completes (i.e., while `WebAssembly.instantiateStreaming` is in-flight), `wasmReady` is still `false` and a second load races with the first. Two concurrent `go.run(result.instance)` calls could register `window.debateosResolve` twice or produce a corrupted WASM state.

In the current debate page implementation this is unlikely because `loadDebateosWasm` is called only once in `onMount`. However the function is exported (`export async function`) and callers in other contexts could trigger concurrent calls.

**Fix:** Replace the boolean with a `Promise` singleton to make the idempotency guard race-safe:

```typescript
let wasmPromise: Promise<void> | null = null;

export function loadDebateosWasm(base = ''): Promise<void> {
    if (!wasmPromise) {
        wasmPromise = _load(base);
    }
    return wasmPromise;
}

async function _load(base: string): Promise<void> {
    // ... existing implementation
}
```

---

### IN-04: `export/+page.svelte` Displays a Hardcoded Example YAML Not Wired to Actual Debate State

**File:** `web/src/routes/export/+page.svelte:26-37`

**Issue:** The export page shows a hardcoded `exampleYaml` string regardless of what the user has composed in the debate page. The download button downloads this example YAML. There is no connection to the `debate` store or the WASM resolver output. A user who completes a debate and navigates to `/export/` will download a static example file, not their actual resolved speech.

This is a v1 limitation that the file itself acknowledges ("In v1, this page shows a template; full wiring is from the debate compose flow"), but it is not communicated to the user at all — the heading says "Your speech is ready" and the CTA says "Download Resolved Speech", both of which are false.

**Fix:** Either (a) wire the export page to the `debate` store and generate the YAML from `resolvedSpeech` via a Svelte import, or (b) add a visible disclaimer: "This is a preview template. Return to the Debate page to compose your speech before exporting." Until wired, the "Download Resolved Speech" button should be disabled or the copy adjusted to "Download Example Speech."

---

_Reviewed: 2026-06-13_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
