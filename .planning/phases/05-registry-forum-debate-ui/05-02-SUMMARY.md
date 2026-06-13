---
phase: 05-registry-forum-debate-ui
plan: "02"
subsystem: web
tags: [sveltekit, tailwind-v4, wasm, adapter-static, ui, brnd]
dependency_graph:
  requires:
    - resolver/wasm/main.go (window.debateosResolve contract)
    - resolver/resolve/explanation.go (ResolvedSpeech/Explanation JSON tags)
    - resolver/types.go (Speech/Opinion/HardwareProfile JSON tags)
  provides:
    - web/src/lib/types.ts (TS mirror of resolver types)
    - web/src/lib/wasm.ts (loadDebateosWasm + debateosResolve + pure helpers)
    - web/src/app.css (Tailwind v4 @theme token contract)
    - scripts/build-wasm.sh (WASM build + wasm_exec.js copy)
  affects:
    - web/** (entire SvelteKit app scaffold)
    - scripts/build-wasm.sh
tech_stack:
  added:
    - svelte@5.25.3
    - "@sveltejs/kit@2.16.0"
    - "@sveltejs/adapter-static@3.0.10"
    - tailwindcss@4.1.4 (v4 — @theme CSS, no tailwind.config.js)
    - "@tailwindcss/vite@4.1.4"
    - "@sveltejs/vite-plugin-svelte@5.0.3"
    - vite@6.2.3
    - "@lucide/svelte@1.18.0 (replacement for deprecated lucide-svelte)"
    - "@fontsource-variable/inter@5.2.8 (replacement for @fontsource/inter/variable.css)"
    - vitest@3.1.1
    - "@playwright/test@1.50.1"
  patterns:
    - SvelteKit adapter-static with BASE_PATH env dual-delivery (UI-02)
    - Tailwind v4 @theme tokens (no tailwind.config.js — v4 standard)
    - Go WASM loader pattern (loadDebateosWasm + wasm_exec.js from GOROOT)
    - Pure helper extraction (parseResolveOutput/buildResolveInput) for Vitest unit tests
    - SSR disabled for /debate route (ssr=false in +page.ts — Pitfall 3 guard)
key_files:
  created:
    - web/package.json
    - web/svelte.config.js
    - web/vite.config.ts
    - web/vitest.config.ts
    - web/playwright.config.ts
    - web/tsconfig.json
    - web/.gitignore
    - web/src/app.css
    - web/src/app.html
    - web/static/.nojekyll
    - web/src/routes/+layout.ts
    - web/src/routes/+layout.svelte
    - web/src/routes/+page.svelte
    - web/src/routes/debate/+page.ts
    - web/src/routes/debate/+page.svelte
    - web/src/routes/browse/+page.svelte
    - web/src/lib/types.ts
    - web/src/lib/wasm.ts
    - web/src/lib/wasm.test.ts
    - scripts/build-wasm.sh
decisions:
  - "[05-02]: @lucide/svelte replaces lucide-svelte (all versions deprecated; official replacement)"
  - "[05-02]: @fontsource-variable/inter replaces @fontsource/inter for variable font CSS"
  - "[05-02]: adapter-static fallback='404.html' enables CSR-only /debate route (ssr=false)"
  - "[05-02]: svelte.config.js handleHttpError ignores favicon.png 404 (not a route; base-path aware)"
  - "[05-02]: base-prefixed hrefs via $app/paths throughout all routes (UI-02 dual-delivery correctness)"
  - "[05-02]: Tailwind v4 @theme in CSS with CSS custom properties; dark variant via @custom-variant"
metrics:
  duration: ~30 min
  completed_date: "2026-06-13T15:44:00Z"
  tasks_completed: 2
  files_created: 20
  commits:
    - "288cbdc: feat(05-02): Task 1 — scaffold"
    - "f6c2dbb: test(05-02): TDD RED — wasm.test.ts"
    - "1f1fd59: feat(05-02): Task 2 — WASM loader + landing page GREEN"
---

# Phase 5 Plan 02: SvelteKit + Tailwind v4 Scaffold Summary

**One-liner:** SvelteKit 5 + adapter-static + Tailwind v4 @theme token contract, typed Go-WASM loader with invariant-3 guard, dual-delivery BASE_PATH seam, BRND-01 landing page, and build-wasm.sh.

---

## What Was Built

### Task 1: SvelteKit + Tailwind v4 Scaffold + UI-SPEC @theme Tokens + WASM Build Script

**Commit:** `288cbdc`

Created the complete `web/` SvelteKit scaffold:

- **web/package.json** — SvelteKit 5.25.3 + adapter-static 3.0.10 + Tailwind v4.1.4 + vitest 3.1.1 + @playwright/test 1.50.1 + @lucide/svelte 1.18.0 + @fontsource-variable/inter
- **web/svelte.config.js** — adapter-static with `pages/assets: 'build'`, `fallback: '404.html'` (for CSR /debate route), `paths.base: process.env.BASE_PATH ?? ''` (UI-02 dual-delivery seam)
- **web/vite.config.ts** — `plugins: [tailwindcss(), sveltekit()]` (Tailwind v4 Vite plugin — no PostCSS)
- **web/src/app.css** — `@import "tailwindcss"` + `@custom-variant dark (&:where(.dark, .dark *))` + `@theme {}` block with every UI-SPEC token: all 11 color tokens (dark canonical; light media-query overrides in `:root:not(.dark)`), 2 font stacks (sans/mono), 5 font sizes (13/14/16/20/28px), 5 line heights, 7 spacing steps (xs..3xl = 4/8/16/24/32/48/64px), foundation-bar height (56px), touch min-height (44px)
- **web/.gitignore** — WASM artifacts `static/debateos.wasm` + `static/wasm_exec.js` gitignored (T-05-03)
- **web/static/.nojekyll** — prevents GitHub Pages Jekyll processing
- **scripts/build-wasm.sh** — `GOOS=js GOARCH=wasm go build -o web/static/debateos.wasm ./resolver/wasm/` + `cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/static/wasm_exec.js` (T-05-03 mitigation)

### Task 2: Typed WASM Loader/Wrapper + Resolver TS Types + Landing Page (BRND-01)

**Commits:** `f6c2dbb` (TDD RED), `1f1fd59` (TDD GREEN + BRND-01 fix)

- **web/src/lib/types.ts** — complete TS interfaces mirroring all resolver JSON tags: Speech, Opinion (with all 20+ optional fields), Point, PointRef, HardwareProfile (`predicates: string[]` — Pitfall 6 guard), ResolvedSpeech, Explanation, ResolveInput/Output/ParsedResolveOutput
- **web/src/lib/wasm.ts** — `loadDebateosWasm(base)` (dynamic import of wasm_exec.js + WebAssembly.instantiateStreaming with base-prefixed URL); `buildResolveInput()` (forces predicates ?? [] — Pitfall 6); `parseResolveOutput()` (parses outer/inner JSON; throws if no result; returns {resolved, error?} — result present even on hard conflict); `debateosResolve()` (calls window.debateosResolve, invariant 3 enforced)
- **web/src/lib/wasm.test.ts** — 10 Vitest unit tests: parseResolveOutput (result+error; error-only throws; no result throws; clean resolve; explanations propagated), buildResolveInput (predicates array; defaults to []; undefined → []; speech/opinions passed through; facts/pci_ids forwarded)
- **web/src/routes/+layout.ts** — `prerender=true`, `trailingSlash='always'`
- **web/src/routes/+layout.svelte** — AppShell: TopNav with DebateOS wordmark + debate/browse/GitHub links (all hrefs base-prefixed via `$app/paths`); footer with "No conclusions — only debates."
- **web/src/routes/+page.svelte** — BRND-01 landing: display tagline "That's just your opinion, man.", sub-tagline "No conclusions — only debates.", "Start Debating" primary CTA, "Browse Points" secondary CTA; cards explaining Opinions/Points/Speech; zero forbidden terms (A6 clean)
- **web/src/routes/debate/+page.ts** — `ssr=false; prerender=false` (Pitfall 3 guard)
- **web/src/routes/debate/+page.svelte** — WasmLoadGate pattern with UI-SPEC copy ("Loading the resolver…", "The resolver failed to load…"); base-prefixed WASM asset URL passed to loadDebateosWasm()

---

## Verification Results

| Check | Result |
|-------|--------|
| `npm run build` (BASE_PATH=) | PASS — build/browse/index.html produced |
| `npm run build` (BASE_PATH=/debateos) | PASS — dual-delivery works |
| `grep -c '@theme' src/app.css` | 2 (one @theme block, one in comment) |
| `grep -q 'process.env.BASE_PATH' svelte.config.js` | PASS |
| `bash scripts/build-wasm.sh` | PASS — produces debateos.wasm + wasm_exec.js |
| WASM artifacts gitignored | PASS |
| `npx vitest run` | PASS — 10/10 tests |
| forbidden terms in +page.svelte | clean |
| `debateosResolve` in wasm.ts | PASS |
| `predicates` in types.ts | PASS |

---

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] lucide-svelte@* deprecated → @lucide/svelte**
- **Found during:** Task 1 (npm install deprecation warning)
- **Issue:** All versions of `lucide-svelte` carry an npm deprecation notice pointing to `@lucide/svelte` as the official replacement package
- **Fix:** Uninstalled `lucide-svelte`, installed `@lucide/svelte@1.18.0` (latest official). The API is import-compatible.
- **Files modified:** web/package.json, web/package-lock.json
- **Commit:** `288cbdc`

**2. [Rule 1 - Bug] @fontsource/inter/variable.css unavailable → @fontsource-variable/inter**
- **Found during:** Task 1 (build failure on first attempt)
- **Issue:** `@fontsource/inter@5.2.5` does not export a `variable.css` subpath. Variable fonts in fontsource use a separate package `@fontsource-variable/inter`.
- **Fix:** Changed import to `@import "@fontsource-variable/inter"` + installed the package.
- **Files modified:** web/src/app.css, web/package.json, web/package-lock.json
- **Commit:** `288cbdc`

**3. [Rule 3 - Blocking] adapter-static strict + ssr=false debate route collision**
- **Found during:** Task 1 (build failure)
- **Issue:** `adapter-static` with `strict: true` fails when a route has `ssr=false; prerender=false` (the debate page needs CSR-only for WASM). The adapter sees it as a "dynamic route".
- **Fix:** Added `fallback: '404.html'` to adapter-static config, which is the standard SvelteKit SPA fallback pattern for CSR routes. Both GitHub Pages and go:embed net/http support the 404.html fallback pattern.
- **Files modified:** web/svelte.config.js
- **Commit:** `288cbdc`, `1f1fd59`

**4. [Rule 3 - Blocking] Links not base-prefixed → BASE_PATH=/debateos build failure**
- **Found during:** Task 2 verification of dual-delivery
- **Issue:** Hard-coded `href="/"` and `href="/debate/"` in layout/pages caused SvelteKit to error during `BASE_PATH=/debateos` build ("404 / does not begin with `base`")
- **Fix:** Imported `base` from `$app/paths` and prefixed all internal hrefs with `{base}/`. This is the standard SvelteKit dual-delivery pattern.
- **Files modified:** web/src/routes/+layout.svelte, web/src/routes/+page.svelte, web/src/routes/debate/+page.svelte
- **Commit:** `1f1fd59`

**5. [Rule 1 - Bug] Comment in +page.svelte contained forbidden terms**
- **Found during:** Task 2 verification (forbidden-term grep hit the comment `// Forbidden: config, preset, package set, distro`)
- **Issue:** The grep check `grep -Eiq '\b(config|preset|distro|package set)\b'` is file-wide and hit the documentation comment listing what NOT to use
- **Fix:** Rewrote comment to not include the forbidden words themselves
- **Files modified:** web/src/routes/+page.svelte
- **Commit:** `1f1fd59`

### TDD Gate Compliance

- RED commit: `f6c2dbb` — `test(05-02)` — wasm.test.ts added (10 tests)
- GREEN commit: `1f1fd59` — `feat(05-02)` — implementation present (was in Task 1 commit `288cbdc`)

Note: The pure helpers (`parseResolveOutput`, `buildResolveInput`) were written in Task 1 as part of the wasm.ts scaffold. The TDD test file was written immediately after (RED commit `f6c2dbb`). Tests passed immediately on first run because the implementation was already in place. This is a minor TDD ordering deviation — the test file creation was delayed by one commit relative to the implementation. All 10 tests are GREEN.

---

## Known Stubs

The following pages are minimal scaffolds (intentional — Wave 2 will implement full components):

| File | Stub Description | Resolving Plan |
|------|------------------|----------------|
| `web/src/routes/debate/+page.svelte` | WasmLoadGate + empty debate state only; no DebateStage component | Phase 5 Plan 03+ (Wave 2) |
| `web/src/routes/browse/+page.svelte` | Static heading + description only; no PointBrowser component | Phase 5 Plan 03+ (Wave 2) |

These stubs do NOT prevent this plan's goals (UI-01 foundation, UI-02 seam, BRND-01 copy) from being achieved. The WASM loader is fully wired; the design system is complete; the type contract is established. Wave 2 builds debate components against these foundations.

---

## Threat Surface Scan

No new threat surface introduced beyond what the plan's threat model covers:

| Flag | File | Description |
|------|------|-------------|
| T-05-05 | web/src/app.css | Verified: @fontsource-variable/inter is self-hosted (no CDN); fonts bundled into build output |
| T-05-03 | scripts/build-wasm.sh | Verified: wasm_exec.js copied from $(go env GOROOT) at build time; gitignored |
| T-05-04 | web/src/lib/wasm.ts | Verified: debateosResolve calls window.debateosResolve only; parseResolveOutput only parses — never decides |

---

## Self-Check: PASSED

Files verified:
- `web/src/lib/types.ts` — FOUND
- `web/src/lib/wasm.ts` — FOUND
- `web/src/lib/wasm.test.ts` — FOUND
- `web/src/app.css` — FOUND
- `web/src/routes/+layout.ts` — FOUND
- `web/src/routes/+layout.svelte` — FOUND
- `web/src/routes/+page.svelte` — FOUND
- `web/src/routes/debate/+page.ts` — FOUND
- `web/src/routes/debate/+page.svelte` — FOUND
- `web/src/routes/browse/+page.svelte` — FOUND
- `scripts/build-wasm.sh` — FOUND
- `web/static/.nojekyll` — FOUND

Commits verified:
- `288cbdc` (Task 1) — FOUND
- `f6c2dbb` (TDD RED) — FOUND
- `1f1fd59` (Task 2 GREEN) — FOUND
