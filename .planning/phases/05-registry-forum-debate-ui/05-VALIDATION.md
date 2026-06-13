---
phase: 5
slug: registry-forum-debate-ui
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-06-13
---

# Phase 5 — Validation Strategy

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (registry/forum); vitest (UI unit); playwright headless (WASM render — confirmed working on host) + Node WASM harness (secondary); bash gate scripts |
| **Config file** | web/vitest.config.ts, web/playwright.config.ts; sqlc.yaml (forum) |
| **Quick run command** | `go test ./registry/... ./forum/... -count=1 && (cd web && npm run test:unit)` |
| **Full suite command** | `go test ./... -count=1 && (cd web && npm run test:unit && npm run test:e2e) && bash scripts/forum-offline-check.sh && bash scripts/check-coverage.sh` |
| **Estimated runtime** | quick < 30s; full < 5 min (WASM build + playwright) |

## Sampling Rate

- **After every task commit:** package-scoped go test / vitest for the touched area
- **After every plan wave:** full suite incl. forum-offline-check + playwright WASM render
- **Before `/gsd-verify-work`:** full suite green; invariant-4 offline gate green
- **Max feedback latency:** 60s fast loop (WASM build + e2e exempt, wave-level)

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 05-XX | TBD | 1+ | REG-01 | — | index validates every doc via resolver/parse | unit | `go test ./registry/... -count=1` | ❌ W0 | ⬜ pending |
| 05-XX | TBD | 1+ | UI-01 | — | UI calls WASM, never reimplements resolve | unit | `cd web && npm run test:unit` | ❌ W0 | ⬜ pending |
| 05-XX | TBD | 2+ | UI-02 | — | live conflict render matches resolver Explanation | e2e | `cd web && npm run test:e2e` (playwright headless) | ❌ W0 | ⬜ pending |
| 05-XX | TBD | 2+ | UI-02 | — | same build → Pages + cli/embed (go:embed serves) | unit | `go test ./cli/... -run TestEmbedServe -count=1` | ❌ W0 | ⬜ pending |
| 05-XX | TBD | 1+ | FORM-01 | — | FTS5 search; read-only over GitHub | unit | `go test ./forum/... -run TestSearch -count=1` | ❌ W0 | ⬜ pending |
| 05-XX | TBD | 1+ | FORM-02/03 | — | OAuth-only identity (fake provider in tests); no native accounts | unit | `go test ./forum/... -run 'TestSubscribe\|TestRating\|TestOAuth' -count=1` | ❌ W0 | ⬜ pending |
| 05-XX | TBD | 1+ | FORM-04/05 | — | conflict threads link PRs; no code exec; no secrets at rest; rebuildable | unit | `go test ./forum/... -run 'TestConflictThread\|TestReindex\|TestBoundaries' -count=1` | ❌ W0 | ⬜ pending |
| 05-XX | TBD | final | invariant-4 | — | compose→resolve→build works with Forum DOWN | script | `bash scripts/forum-offline-check.sh` | ❌ W0 | ⬜ pending |
| 05-XX | TBD | final | BRND-01 | — | N/A | e2e | playwright assertion A6 (no forbidden terms config/preset/distro in visible text) | ❌ W0 | ⬜ pending |
| 05-XX | TBD | all | regression | — | N/A | unit | `go test ./... -count=1` (resolver/cli/translators stay green) | ✅ | ⬜ pending |

## Wave 0 Requirements

- [ ] web/ SvelteKit + adapter-static + Tailwind v4 scaffold; vitest + playwright config
- [ ] registry/ + forum/ Go package skeletons with RED tests; sqlc.yaml + migrations
- [ ] scripts/forum-offline-check.sh; WASM build step (GOOS=js GOARCH=wasm) wired into web build
- [ ] registry fixtures + UI test data

## Manual-Only / Deferred-to-host

| Behavior | Requirement | Why | Instructions |
|----------|-------------|-----|--------------|
| Live GitHub OAuth flow | FORM-03 | needs a registered GitHub OAuth app | document app-registration; tests use fake provider |
| Live GitHub Pages deploy | UI-02/REG-01 | needs repo Pages enablement | document; build output verified locally |
| Live Oracle A1 (arm64) deploy | FORM hosting (D15) | needs cloud account | document GOARCH=arm64 build + deploy steps |
| Live Actions index rebuild | REG-01 | needs CI minutes | workflow authored; live run deferred |

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity OK; no watch-mode; latency < 60s fast loop
- [x] `nyquist_compliant: true`

**Approval:** approved 2026-06-13 (autonomous run)
