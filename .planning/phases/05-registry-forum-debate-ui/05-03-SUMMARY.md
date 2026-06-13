---
phase: 05-registry-forum-debate-ui
plan: "03"
subsystem: forum
tags: [forum, sqlite, fts5, sqlc, chi, store, tdd, subscriptions, ratings, search]

requires:
  - phase: 05-01
    provides: "go.mod with chi/v5 v5.3.0 + modernc.org/sqlite v1.46.1 + oauth2 v0.34.0; deps_guard.go single-owner invariant"
  - phase: 05-01
    provides: "registry/index.PointEntry type that forum indexes parallel"

provides:
  - "forum/store.Store interface (SearchPoints/GetPoint/ListPoints/UpsertPoint/Subscriptions/Ratings/Conflicts/Reindex/Truncate)"
  - "forum/store.SQLiteStore backed by sqlc-generated queries over modernc.org/sqlite FTS5"
  - "forum/store.NewInMemory() in-memory constructor for tests"
  - "forum/migrations/001_init.sql: points + points_fts(FTS5 external-content) + subscriptions + ratings + conflict_threads"
  - "forum/api.NewRouter(store, identityFn) chi mux with IdentityFn seam for OAuth/tests"
  - "forum/api: GET /api/search, GET /api/points, GET /api/points/{id}, GET /api/ratings/{pointId}"
  - "forum/api: POST /api/ratings (identity-gated 401), POST/DELETE /api/subscriptions (identity-gated)"
  - "forum/sqlc.yaml + forum/query.sql: parameterized sqlc queries (no interpolation)"
  - "forum/store/generated/: sqlc v1.30.0 generated Go code committed"

affects:
  - "05-05 (OAuth flow + threads + reindex + cmd) — consumes Store interface, chi router seam, IdentityFn"
  - "web/ (05-04) — calls /api/search, /api/points, /api/ratings endpoints"

tech-stack:
  added:
    - "modernc.org/sqlite v1.46.1 (FTS5 confirmed functional, consumed from go.mod via 05-01 dep anchor)"
    - "github.com/go-chi/chi/v5 v5.3.0 (Forum HTTP router)"
    - "sqlc v1.30.0 (installed binary, generated code committed)"
  patterns:
    - "TDD RED/GREEN per task (2 RED commits, 2 GREEN commits)"
    - "FTS5 external-content table with explicit rebuild after UpsertPoint (Pitfall 5 mitigated)"
    - "IdentityFn seam: func(r) (string,bool) injected at NewRouter — no OAuth coupling in this plan"
    - "sqlc parameterized queries only (never fmt.Sprintf for SQL); FTS5 MATCH via raw sql.DB (sqlc can't parse virtual tables)"
    - "Stars validation in Go + SQLite CHECK — defense-in-depth (T-05-06/V5)"
    - "Foundation filter: JSON substring match on stored foundation_compat JSON array"

key-files:
  created:
    - forum/migrations/001_init.sql
    - forum/migrations/migrate.go
    - forum/store/store.go
    - forum/store/sqlite.go
    - forum/store/inmem.go
    - forum/store/store_test.go
    - forum/store/generated/db.go
    - forum/store/generated/models.go
    - forum/store/generated/querier.go
    - forum/store/generated/query.sql.go
    - forum/sqlc.yaml
    - forum/query.sql
    - forum/api/router.go
    - forum/api/search.go
    - forum/api/ratings.go
    - forum/api/points.go
    - forum/api/api_test.go
  modified: []

key-decisions:
  - "[05-03] FTS5 MATCH query implemented via raw sql.DB (not sqlc) — sqlc cannot parse virtual table columns in schema; all other queries remain sqlc-generated (parameterized)"
  - "[05-03] Foundation filter applied post-FTS5 in Go (substring match on JSON array) — avoids SQL injection surface; acceptable for read-mostly small result sets"
  - "[05-03] IdentityFn seam: func(r *http.Request) (string, bool) injected at NewRouter — decouples OAuth session from this plan; 05-05 wires the real session; tests pass fakeIdentity"
  - "[05-03] Stub methods (GetConflicts, UpsertConflictThread, Reindex) implemented as thin wrappers that compile and pass interface check — 05-05 will complete their behavior"
  - "[05-03] sqlc null_style option not available in v1.30.0 — removed; AvgRating returned as interface{} from COALESCE query; cast handled in GetRatings via type switch"

patterns-established:
  - "Pattern: FTS5 external-content sync — always call RebuildFTS after UpsertPoint batch"
  - "Pattern: IdentityFn seam for OAuth-gated endpoints — inject at NewRouter; tests use closures"
  - "Pattern: chi middleware.Recoverer at router level; per-route identity check inline"

requirements-completed: [FORM-01, FORM-02, FORM-03]

duration: ~35min
completed: 2026-06-13
---

# Phase 5 Plan 03: Forum Store + FTS5 + sqlc + chi Read API Summary

**Forum storage core: sqlc-generated queries over modernc.org/sqlite FTS5 with subscription/rating store and chi read API gated on injected IdentityFn — all tested against in-memory SQLite.**

## Performance

- **Duration:** ~35 min
- **Started:** 2026-06-13
- **Completed:** 2026-06-13
- **Tasks:** 2
- **Files modified:** 17

## Accomplishments

- FTS5 full-text search confirmed functional in modernc.org/sqlite v1.46.1; SearchPoints with text + foundation filter works
- Subscription edges (AddSubscription/RemoveSubscription/GetSubscriptions) round-trip idempotently in SQLite
- Ratings (SetRating/GetRatings) aggregate correctly; out-of-range stars rejected both in Go and by SQLite CHECK constraint
- chi router with IdentityFn seam: POST /api/ratings returns 401 without identity, 200 with injected fake userID
- All 7 store tests + 2 API tests pass against in-memory SQLite; full `go test ./...` green

## Task Commits

1. **Task 1 RED: FTS5Smoke/SearchPoints/ListGetUpsert tests** - `4bd29d9` (test)
2. **Task 1 GREEN: migrations+FTS5+sqlc+Store interface** - `186d238` (feat)
3. **Task 2 RED: subscriptions/ratings/identity-gate/search-endpoint tests** - `1917233` (test)
4. **Task 2 GREEN: subscriptions+ratings+chi read API** - `5a3c012` (feat)

## Files Created/Modified

- `forum/migrations/001_init.sql` — Schema: points, points_fts (FTS5 external-content), subscriptions, ratings, conflict_threads
- `forum/migrations/migrate.go` — embed+apply SQL via //go:embed; Apply(db) error
- `forum/store/store.go` — Store interface with all domain methods + stub declarations for 05-05
- `forum/store/sqlite.go` — SQLiteStore: Open(dsn) with WAL+FK pragmas, sqlc-backed impl; FTS5 via raw sql.DB
- `forum/store/inmem.go` — NewInMemory() for tests
- `forum/store/store_test.go` — 7 tests: FTS5Smoke, SearchPoints, ListGetUpsert, Subscriptions, Ratings
- `forum/store/generated/` — sqlc v1.30.0 generated code (db.go, models.go, querier.go, query.sql.go)
- `forum/sqlc.yaml` — sqlc config (engine:sqlite, package:generated, emit_json_tags)
- `forum/query.sql` — Parameterized named queries (no FTS5 SELECT — handled raw in sqlite.go)
- `forum/api/router.go` — NewRouter(store, identityFn) chi mux; IdentityFn type; requireIdentity helper
- `forum/api/search.go` — GET /api/search, GET /api/points, GET /api/points/{id}
- `forum/api/ratings.go` — POST /api/ratings (identity-gated), GET /api/ratings/{pointId}
- `forum/api/points.go` — POST /api/subscriptions, DELETE /api/subscriptions (identity-gated)
- `forum/api/api_test.go` — 2 tests: TestRatingRequiresIdentity, TestSearchEndpoint

## Decisions Made

- FTS5 MATCH query written as raw SQL (sqlc cannot parse virtual table columns — generates column-not-found error). All non-FTS5 queries remain sqlc-generated and fully parameterized.
- Foundation filter applied in Go via JSON substring match rather than SQL `json_each()` — simpler, no injection surface, adequate for small result sets.
- IdentityFn as `func(*http.Request) (string, bool)` — flat closure, no interface overhead; tests inject `fakeIdentity("github-user-1")`, production (05-05) will inject session reader.
- GetConflicts/UpsertConflictThread/Reindex implemented as thin wrappers now; 05-05 fills their behavior (conflict thread management + bulk reindex from registry).
- sqlc v1.30.0 lacks `null_style` option — removed from config; COALESCE AVG returned as `interface{}` and cast via type switch in GetRatings.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Removed FTS5 query from sqlc schema — sqlc can't parse virtual table columns**
- **Found during:** Task 1 (sqlc generate)
- **Issue:** `sqlc generate` failed with `column "points_fts" does not exist` — the FTS5 virtual table's column reference is not recognized by sqlc's SQLite schema parser
- **Fix:** Moved FTS5 MATCH query to raw `sql.DB.QueryContext` in sqlite.go; all non-FTS5 queries remain in sqlc. Commented the query.sql entry to explain the omission.
- **Files modified:** forum/query.sql, forum/store/sqlite.go
- **Verification:** `sqlc generate` clean; FTS5 search tests pass; no string interpolation in SQL
- **Committed in:** `186d238` (Task 1 GREEN)

**2. [Rule 3 - Blocking] Removed unsupported null_style option from sqlc.yaml**
- **Found during:** Task 1 (sqlc generate)
- **Issue:** `null_style` field not found in opts.Options for sqlc v1.30.0
- **Fix:** Removed `null_style: "option"` from sqlc.yaml; handled nullable AVG return via type switch
- **Files modified:** forum/sqlc.yaml, forum/store/sqlite.go
- **Committed in:** `186d238` (Task 1 GREEN)

---

**Total deviations:** 2 auto-fixed (both Rule 3 — blocking issues from tool version constraints)
**Impact on plan:** Functionally equivalent — FTS5 search works as specified; parameterization maintained via raw sql.DB; no security surface added.

## Issues Encountered

None beyond the two auto-fixed blocking issues above.

## Known Stubs

The following Store methods are declared and compile but are intentionally minimal for 05-05:
- `GetConflicts` / `UpsertConflictThread` — thin sqlc wrappers; 05-05 adds conflict thread management UI
- `Reindex` — calls RebuildFTS only; 05-05 adds full re-index from registry index.json

These stubs do NOT block the plan's goal (FORM-01/02/03). They are the stable interface contract for 05-05.

## Threat Flags

No new security surface beyond the plan's threat model:
- T-05-06: Parameterized queries throughout; FTS5 MATCH uses `?` placeholder (raw sql.DB, not interpolated)
- T-05-07: All write endpoints gated on IdentityFn; 401 returned if identity absent
- T-05-08: All data served as JSON; UI renders as text nodes (Svelte auto-escape)
- T-05-09: No OAuth tokens, passwords, or secrets stored in SQLite

## Next Phase Readiness

- Store interface is stable; 05-05 can implement OAuth flow + conflict threads by extending without modifying the interface
- chi router accepts additional routes from 05-05 via the same NewRouter pattern
- All `go test ./...` green; no regressions in existing packages

---
*Phase: 05-registry-forum-debate-ui*
*Completed: 2026-06-13*

## Self-Check: PASSED

- All 17 files created and exist on disk
- All 4 task commits found in git log (4bd29d9, 186d238, 1917233, 5a3c012)
- `go test ./... -count=1` all green (17 packages, 0 failures)
