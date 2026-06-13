# Phase 5: Registry, Forum & Debate UI — Research

**Researched:** 2026-06-13
**Domain:** SvelteKit + Go-WASM UI / Go chi+SQLite Forum / Go static registry generator
**Confidence:** HIGH (stack confirmed via live registry checks + empirical tests on this host)

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Registry: Go static-site generator; reads point/speech YAML from local fixtures + examples/; validates via resolver/parse; emits static JSON index + minimal HTML; GitHub Actions rebuilds on commit (deferred live run).
- Foundation-compatibility computed from translators/*/capabilities.json — reuse, do not duplicate.
- Debate UI: SvelteKit + adapter-static + Tailwind; static output. Calls Go-WASM resolver (resolver/wasm) client-side. NEVER reimplements resolution (invariant 3). Core flow: load points → compose panes → live conflict visualization → build instructions.
- Dual delivery: same build → GitHub Pages AND go:embed under cli/embed/ for `debateos compose` offline serve. Byte-identical.
- Brand voice (BRND-01): debate/rhetoric metaphor applied; "That's just your opinion, man"; "no conclusions, only debates".
- Forum: Go chi router + modernc.org/sqlite + sqlc + store interface; in-memory SQLite for tests; GitHub OAuth only (golang.org/x/oauth2); FTS5 search; no native accounts; no code execution; rebuildable.
- Security mandatory: read-mostly; no arbitrary uploads; no secrets at rest; OAuth only; single static binary + one SQLite file.
- Hosting: Oracle Cloud Always Free Ampere A1 (ARM, EU/APAC); deploy notes only (no live deploy this phase).
- Invariant 4 gate: automated test proves compose→resolve→build with Forum DOWN.
- TDD: Go table-driven RED/GREEN; sqlc queries tested against in-memory SQLite; Vitest unit + Playwright/headless WASM-conflict test; coverage gates forum/registry ≥85%.

### Claude's Discretion
- web/ component structure, Tailwind theme tokens, sqlc query organization, exact JSON index schema, FTS5 schema details, brand-copy wording.

### Deferred Ideas (OUT OF SCOPE)
- Live GitHub OAuth app, live Pages deploy, live Oracle A1 deploy, live Actions index rebuild — deferred-to-host/CI.
- Postgres tsvector backend — post-v1.0.
- GitLab registry parity — post-v1.0.
- Browser-based Playwright smoke if Chromium uninstallable — RESOLVED: Chromium IS installable and works on this host (tested below).
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| REG-01 | Static registry index from point/speech YAML in GitHub repos, hosted on Pages, rebuilt on commit | §Registry Generator section; deterministic output pattern; resolver/parse reuse |
| UI-01 | SvelteKit + adapter-static + Tailwind visual Debate UI with live WASM conflict visualization | §SvelteKit Stack; §WASM Integration; §Conflict Visualization; UI-SPEC contract |
| UI-02 | Same UI build deployed to GitHub Pages AND embedded via go:embed in CLI for offline serve | §Dual Delivery Architecture; base-path handling; §go:embed serve pattern |
| BRND-01 | Debate-themed brand voice across UI and docs | §Brand Voice section; UI-SPEC copywriting contract |
| FORM-01 | Forum search/discovery over indexed points/speeches (FTS5) | §Forum Stack; FTS5 VERIFIED functional in modernc.org/sqlite v1.46.1 |
| FORM-02 | Subscriptions: follow curators, subscribe to point sets | §Forum Store Schema; subscription edges in SQLite |
| FORM-03 | Ratings/reputation tied to GitHub OAuth identity | §OAuth Integration; GitHub OAuth web flow pattern |
| FORM-04 | Conflict threads: known-conflict registry + links to patch PRs | §Forum Store Schema; conflict_threads table |
| FORM-05 | Forum optional/additive; core path works Forum-offline; rebuildable; no untrusted code | §Invariant-4 Gate test; §Forum Security constraints |
</phase_requirements>

---

## Summary

Phase 5 delivers four cooperating artifacts: a Go static registry index generator, a SvelteKit/Tailwind/WASM Debate UI with dual-delivery (GitHub Pages + go:embed), an optional Go chi+SQLite Forum service, and debate-themed brand voice. All four are greenfield within the existing monorepo (module github.com/mikl0s/debateos, go 1.24.0 / toolchain go1.24.1).

The most significant technical fact discovered this session: **Playwright + Chromium headless is installable and fully functional on this host** (system deps installed, smoke-tested, 1 passed in 864ms). The CONTEXT.md fallback to a Node WASM harness is available but not needed — the Playwright integration test IS the primary path. Additionally, the **Node.js WASM harness also works** (tested with debateos.wasm directly, `debateosResolve()` called and returned a valid `ResolvedSpeech`) so it serves as a secondary assertion harness.

The Forum's SQLite version requires attention: **modernc.org/sqlite v1.52.0 requires Go 1.25**, but v1.46.1 (the last go-1.24-compatible release) passes a full FTS5 functional test. The plan must pin to v1.46.1 OR accept that adding v1.52.0 will auto-upgrade the toolchain directive to go1.25 (downloads ~200MB toolchain on first run). Recommendation: pin v1.46.1 to keep the existing toolchain intact.

**Primary recommendation:** Implement in four waves: Wave 0 (WASM build + test infra scaffolding), Wave 1 (Registry generator + registry tests), Wave 2 (SvelteKit UI scaffold + WASM integration + core Debate components), Wave 3 (Forum Go service), Wave 4 (Integration, go:embed serve, Playwright test, invariant-4 gate).

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Point/speech YAML validation | Go (registry generator) | — | resolver/parse already handles this; registry reuses it |
| Static registry JSON index | Go static generator | GitHub Actions (trigger) | Deterministic, committed output; no server needed |
| Conflict resolution logic | Go WASM (browser) | Go native (CLI) | Invariant 3: UI never reimplements; calls debateosResolve() |
| Conflict visualization | SvelteKit component (browser) | — | Pure UI rendering of ResolvedSpeech returned from WASM |
| Static UI serving (Pages) | GitHub Pages CDN | — | adapter-static build output |
| Static UI serving (offline) | Go net/http + embed.FS | CLI compose subcommand | go:embed cli/embed/; http.FileServer over embed.FS |
| Forum search/discovery | Go chi API (server) + SQLite FTS5 | Client-side JSON filter (offline fallback) | Server-side FTS5 for online; static registry JSON for offline |
| GitHub OAuth identity | Go chi + golang.org/x/oauth2 | — | Server-side flow; no client secrets in browser |
| Ratings/subscriptions/threads | Go chi API + SQLite | — | Server-side store; Forum-offline degrades gracefully |
| Foundation-compat computation | Registry generator (Go) | — | Reads capabilities.json, computes per-point compatibility |
| Brand voice / copy | SvelteKit templates + docs | — | BRND-01 copywriting contract from UI-SPEC |

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| svelte | 5.56.3 | UI framework | [VERIFIED: npm registry] Current Svelte 5 with runes |
| @sveltejs/kit | 2.65.0 | App framework | [VERIFIED: npm registry] SvelteKit 2 stable |
| @sveltejs/adapter-static | 3.0.10 | Static/SSG output | [VERIFIED: npm registry] Produces Pages-ready static build |
| tailwindcss | 4.3.1 | Utility CSS | [VERIFIED: npm registry] v4 uses Vite plugin, not PostCSS |
| @tailwindcss/vite | 4.3.1 | Vite integration for Tailwind v4 | [VERIFIED: npm registry] Replaces @tailwindcss/postcss in v4 |
| @sveltejs/vite-plugin-svelte | 7.1.2 | Vite/Svelte integration | [VERIFIED: npm registry] |
| vite | 8.0.16 | Build bundler | [VERIFIED: npm registry] |
| lucide-svelte | 1.0.1 | Icon components | [VERIFIED: npm registry] Tree-shakeable MIT icons (UI-SPEC mandated) |
| @fontsource/inter | 5.2.8 | Self-hosted Inter font | [VERIFIED: npm registry] Satisfies offline/embed invariant (no CDN) |
| github.com/go-chi/chi/v5 | v5.3.0 | Forum HTTP router | [VERIFIED: Go proxy] Minimal, idiomatic; go 1.23+ compat |
| modernc.org/sqlite | v1.46.1 | Pure-Go SQLite (Forum) | [VERIFIED: empirically tested] Last go-1.24-compatible release; FTS5 confirmed functional |
| golang.org/x/oauth2 | v0.36.0 | GitHub OAuth web flow | [VERIFIED: Go proxy] Official Go extended library |

### Supporting (dev/test)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| vitest | 4.1.8 | UI unit tests | SvelteKit logic, store functions, WASM wrapper — no browser needed |
| @playwright/test | 1.60.0 | Browser integration test | WASM-render test, accessibility assertions, invariant-4 browser path |
| sqlc | v1.30.0 (installed), v1.31.1 (latest) | SQL→Go code generation | Run `sqlc generate` in forum/ at dev time; generated code committed |

Note: sqlc is a codegen tool, not a runtime library. The locally installed binary is v1.30.0; v1.31.1 is latest. Either works. Use the installed binary (`/home/mikkel/go/bin/sqlc version` = v1.30.0).

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| modernc.org/sqlite v1.46.1 | v1.52.0 | v1.52.0 requires Go 1.25 (auto-downloads toolchain); pin v1.46.1 to keep go 1.24 |
| Playwright browser test | Node.js WASM harness | Both work on this host; Playwright is preferred (tests real browser rendering) |
| @tailwindcss/vite (v4) | @tailwindcss/postcss | v4 standard is Vite plugin; postcss approach is v3 legacy |
| lucide-svelte | Heroicons, Phosphor | lucide is specified in UI-SPEC; do not substitute |

**Installation:**

```bash
# web/ frontend
cd web/
npm install svelte @sveltejs/kit @sveltejs/adapter-static tailwindcss @tailwindcss/vite @sveltejs/vite-plugin-svelte vite lucide-svelte @fontsource/inter
npm install -D vitest @playwright/test
npx playwright install chromium   # system deps already installed on this host

# Go modules (add to go.mod via go get)
go get github.com/go-chi/chi/v5@v5.3.0
go get modernc.org/sqlite@v1.46.1
go get golang.org/x/oauth2@v0.36.0
```

---

## Package Legitimacy Audit

| Package | Registry | Source Repo | Verdict | Disposition |
|---------|----------|-------------|---------|-------------|
| svelte@5.56.3 | npm | github.com/sveltejs/svelte | OK | Approved |
| @sveltejs/kit@2.65.0 | npm | github.com/sveltejs/kit | OK | Approved |
| @sveltejs/adapter-static@3.0.10 | npm | github.com/sveltejs/kit | OK | Approved |
| tailwindcss@4.3.1 | npm | github.com/tailwindlabs/tailwindcss | OK | Approved |
| @tailwindcss/vite@4.3.1 | npm | github.com/tailwindlabs/tailwindcss | OK | Approved |
| @sveltejs/vite-plugin-svelte@7.1.2 | npm | github.com/sveltejs/vite-plugin-svelte | OK | Approved |
| vite@8.0.16 | npm | github.com/vitejs/vite | OK | Approved |
| lucide-svelte@1.0.1 | npm | github.com/lucide-icons/lucide | OK | Approved |
| @fontsource/inter@5.2.8 | npm | github.com/fontsource/font-files | OK | Approved |
| vitest@4.1.8 | npm | github.com/vitest-dev/vitest | OK | Approved |
| @playwright/test@1.60.0 | npm | github.com/microsoft/playwright | OK | Approved |
| github.com/go-chi/chi/v5@v5.3.0 | Go proxy | github.com/go-chi/chi | OK | Approved |
| modernc.org/sqlite@v1.46.1 | Go proxy | gitlab.com/cznic/sqlite | OK | Approved |
| golang.org/x/oauth2@v0.36.0 | Go proxy | github.com/golang/oauth2 | OK | Approved |

**Packages removed due to SLOP verdict:** none
**Packages flagged as suspicious (SUS):** none

---

## Architecture Patterns

### System Architecture Diagram

```
                         ┌─── BUILD TIME ───────────────────────────────┐
  examples/ + fixtures   │                                               │
  (YAML points/opinions) │  registry/generator.go                       │
         │               │  ├── resolver/parse (validate YAML)           │
         └───────────────►  ├── translators/*/capabilities.json          │
                         │  │   (foundation-compat computation)          │
                         │  └──► registry/index.json + static HTML      │
                         │                      │                        │
  resolver/wasm/main.go  │  GOOS=js GOARCH=wasm │                       │
         │               │  go build → debateos.wasm                    │
         └───────────────►                       │                       │
                         │  web/ (SvelteKit build)                       │
                         │  ├── copy wasm_exec.js from $(GOROOT)/lib/wasm│
                         │  ├── include debateos.wasm in static/         │
                         │  └──► web/build/ (static HTML/JS/CSS/WASM)   │
                         │             │                │               │
                         │     GitHub Pages         cli/embed/          │
                         │     (Pages deploy)       (go:embed target)   │
                         └─────────────────────────────────────────────-┘

                         ┌─── RUNTIME: BROWSER ─────────────────────────┐
  User browser           │                                               │
  GET /debate ───────────► SvelteKit SPA (static JS)                    │
                         │  ├── WasmLoadGate: fetch /debateos.wasm       │
                         │  │   WebAssembly.instantiateStreaming()       │
                         │  │   go.run(instance) → init() registers     │
                         │  │   window.debateosResolve                  │
                         │  │                                            │
                         │  ├── DebateStage component                    │
                         │  │   ├── User adds points → build input JSON  │
                         │  │   ├── 150ms debounce →                    │
                         │  │   │   window.debateosResolve(JSON.stringify│
                         │  │   │   {speech, opinions, hardware}))       │
                         │  │   ├── Parse ResolvedSpeech from result     │
                         │  │   └── ConflictOverlay + ResolutionPanel    │
                         │  │                                            │
                         │  └── PointBrowser (HEAD Forum /health, 3s)   │
                         │      ├── Forum online: FTS5 API search        │
                         │      └── Forum offline: client filter over    │
                         │          static registry/index.json           │
                         └─────────────────────────────────────────────-┘

                         ┌─── RUNTIME: FORUM SERVICE (optional) ────────┐
  Browser ──────────────► chi Router (Go)                               │
  GET /api/search        │  ├── /api/search      (FTS5 SQLite query)    │
  GET /api/points/:id    │  ├── /api/points/:id  (point detail)         │
  POST /api/ratings      │  ├── /api/ratings     (OAuth-gated write)    │
  GET /api/conflicts     │  ├── /api/conflicts   (conflict threads)      │
  GET /oauth/callback    │  └── /oauth/...       (GitHub OAuth flow)    │
                         │       │                                       │
                         │  store.Store interface                        │
                         │  └── sqlc-generated queries                   │
                         │       └── modernc.org/sqlite (FTS5)           │
                         └─────────────────────────────────────────────-┘

  debateos compose ──────► cli/embed/ (embed.FS)                        
                           http.FileServer(http.FS(sub(embedFS,"embed")))
                           → serves web/build/ at localhost:PORT/        
```

### Recommended Project Structure

```
web/                          # SvelteKit app (Phase 5 new)
├── package.json
├── svelte.config.js          # adapter-static; paths.base via env
├── vite.config.ts            # @tailwindcss/vite plugin + sveltekit()
├── playwright.config.ts      # Playwright test config
├── src/
│   ├── app.css               # @import "tailwindcss"; @theme {...}; @custom-variant dark
│   ├── app.html
│   ├── lib/
│   │   ├── wasm.ts           # WASM loader: loadDebateosWasm(), debateosResolve()
│   │   ├── stores/           # Svelte stores: speech.ts, forum.ts, wasm.ts
│   │   └── types.ts          # TypeScript types mirroring resolver structs
│   └── routes/
│       ├── +layout.svelte    # AppShell, TopNav, ToastStack
│       ├── +layout.ts        # prerender=true, trailingSlash='always'
│       ├── +page.svelte      # Landing / empty debate
│       ├── debate/
│       │   ├── +page.svelte  # DebateStage (ssr=false — WASM only)
│       │   └── +page.ts      # export const ssr = false
│       ├── browse/
│       │   ├── +page.svelte  # PointBrowser
│       │   └── [id]/+page.svelte  # PointDetail
│       ├── curator/[id]/+page.svelte
│       └── export/+page.svelte
├── static/
│   ├── .nojekyll             # prevents Jekyll processing on GitHub Pages
│   └── debateos.wasm         # built at build time (not committed; see build script)
└── tests/
    └── wasm-render.spec.ts   # Playwright WASM integration test (assertion A3 + A9)

registry/                     # Go static index generator (Phase 5 new)
├── generator.go              # main: scan fixtures, parse, emit index.json + HTML
├── generator_test.go         # table-driven: golden index.json, determinism
└── index/
    ├── index.go              # RegistryIndex, PointEntry structs
    └── compat.go             # foundation-compat from capabilities.json

forum/                        # Go Forum service (Phase 5 new)
├── api/
│   ├── router.go             # chi.NewRouter(), mount routes
│   ├── search.go             # GET /api/search (FTS5)
│   ├── points.go             # GET /api/points, /api/points/:id
│   ├── ratings.go            # POST/GET /api/ratings (OAuth-gated)
│   ├── conflicts.go          # GET/POST /api/conflicts
│   └── oauth.go              # /oauth/login, /oauth/callback (GitHub)
├── store/
│   ├── store.go              # Store interface (domain methods)
│   ├── sqlite.go             # sqlc-backed SQLite implementation
│   └── inmem.go              # in-memory SQLite for tests (":memory:")
├── migrations/
│   ├── 001_init.sql          # schema + FTS5 virtual tables
│   └── migrate.go            # embed .sql; apply on start
├── sqlc.yaml                 # sqlc config (engine: sqlite)
├── query.sql                 # Named sqlc queries
└── deploy/
    └── oracle-a1.md          # ARM binary + SQLite deployment notes

cli/embed/                    # go:embed target (created this phase)
└── (web/build/ contents copied here at build time — see Makefile/script)
```

### Pattern 1: SvelteKit adapter-static + Tailwind v4

**What:** Fully static SvelteKit output suitable for GitHub Pages and go:embed. Tailwind v4 uses a Vite plugin instead of PostCSS config.

**When to use:** Any SvelteKit route that doesn't need server-side data (all routes in this phase are prerendered or CSR-only).

```typescript
// svelte.config.js
// Source: https://svelte.dev/docs/kit/adapter-static [CITED]
import adapter from '@sveltejs/adapter-static';

export default {
  kit: {
    adapter: adapter({
      pages: 'build',
      assets: 'build',
      fallback: undefined,
      precompress: false,
      strict: true
    }),
    paths: {
      // Set BASE_PATH=/debateos for Pages deploy; empty for localhost
      base: process.env.BASE_PATH ?? ''
    }
  }
};
```

```typescript
// vite.config.ts
// Source: https://teta.so/blog/sveltekit-tailwind-css-setup-best-practices [CITED]
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [tailwindcss(), sveltekit()]
});
```

```css
/* src/app.css */
@import "tailwindcss";

/* Custom dark variant (class-based, dark is default) */
@custom-variant dark (&:where(.dark, .dark *));

/* Design tokens from UI-SPEC §Tailwind Theme Extension */
@theme {
  --color-surface-base-dark: #0f1117;
  --color-surface-base-light: #f9fafb;
  --color-surface-card-dark: #1a1f2e;
  --color-accent-brand-dark: #6366f1;
  --color-conflict-hard: #ef4444;
  --color-conflict-compat: #22c55e;
  --color-conflict-warn: #f59e0b;
  /* ... full token list per UI-SPEC */
  --font-sans: 'Inter var', 'Inter', system-ui, sans-serif;
  --font-mono: ui-monospace, 'Cascadia Code', monospace;
}
```

```typescript
// src/routes/+layout.ts — prerender entire app
// Source: https://svelte.dev/docs/kit/adapter-static [CITED]
export const prerender = true;
export const trailingSlash = 'always';
```

### Pattern 2: Go WASM Loading from SvelteKit (CSR-only page)

**What:** The /debate route must be CSR-only (WASM cannot run in Node/SSR). Load wasm_exec.js and debateos.wasm on mount. Register `window.debateosResolve` then use it synchronously.

**Key facts (verified this session):**
- `wasm_exec.js` path: `$(go env GOROOT)/lib/wasm/wasm_exec.js` = `/usr/local/go/lib/wasm/wasm_exec.js` [VERIFIED: filesystem]
- Built WASM size: **4.3 MB** for the full resolver WASM [VERIFIED: `GOOS=js GOARCH=wasm go build ./resolver/wasm/ -o debateos_test.wasm`]
- Do NOT commit wasm_exec.js — copy at build time from GOROOT [CITED: resolver/wasm/main.go T-01-15]
- Build command: `GOOS=js GOARCH=wasm go build -o web/static/debateos.wasm ./resolver/wasm/` [VERIFIED: build works]
- Copy wasm_exec.js: `cp $(go env GOROOT)/lib/wasm/wasm_exec.js web/static/wasm_exec.js`

```typescript
// src/routes/debate/+page.ts — disable SSR for WASM page
// Source: https://svelte.dev/docs/kit/page-options [CITED]
export const ssr = false;
```

```typescript
// src/lib/wasm.ts — WASM loader module
// Source: resolver/wasm/main.go [VERIFIED: codebase read]
let wasmReady = false;

export async function loadDebateosWasm(): Promise<void> {
  if (wasmReady) return;

  // wasm_exec.js adds globalThis.Go
  await import('/wasm_exec.js');
  const go = new (globalThis as any).Go();

  const result = await WebAssembly.instantiateStreaming(
    fetch('/debateos.wasm'),  // served from static/
    go.importObject
  );
  go.run(result.instance);
  // init() in main.go registers window.debateosResolve synchronously
  wasmReady = true;
}

// Matches debateosResolveFunc contract from resolver/wasm/main.go
export interface ResolveInput {
  speech: Speech;
  opinions: Opinion[];
  hardware: HardwareProfile;
}

export interface ResolveOutput {
  result?: string;   // JSON-encoded ResolvedSpeech (present even on hard conflict)
  error?: string;    // non-empty = at least one hard conflict
}

export function debateosResolve(input: ResolveInput): { resolved: ResolvedSpeech; error?: string } {
  const raw: string = (globalThis as any).debateosResolve(JSON.stringify(input));
  const out: ResolveOutput = JSON.parse(raw);
  if (!out.result) throw new Error(out.error ?? 'resolver returned no result');
  return { resolved: JSON.parse(out.result), error: out.error };
}
```

```svelte
<!-- src/routes/debate/+page.svelte (excerpt) -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { loadDebateosWasm, debateosResolve } from '$lib/wasm.js';
  import WasmLoadGate from '$lib/components/WasmLoadGate.svelte';

  let wasmLoaded = false;
  let wasmError = false;

  onMount(async () => {
    try {
      await loadDebateosWasm();
      wasmLoaded = true;
    } catch (e) {
      wasmError = true;
    }
  });
</script>

{#if !wasmLoaded && !wasmError}
  <WasmLoadGate />     <!-- role="status" aria-live="polite" "Loading the resolver…" -->
{:else if wasmError}
  <p>The resolver failed to load. Refresh to try again.</p>
{:else}
  <!-- DebateStage with debateosResolve available -->
{/if}
```

**WASM call contract (from resolver/wasm/main.go — VERIFIED):**
- Input: `JSON.stringify({ speech: Speech, opinions: Opinion[], hardware: HardwareProfile })`
- `speech.foundation` required; `opinions` must include all opinions referenced by speech.points
- `hardware.predicates` is `string[]` (NOT an object — a common mistake)
- Output: `{ result?: string, error?: string }` — `result` is a JSON-encoded `ResolvedSpeech`; `error` is non-empty on hard conflict but `result` is STILL present (partial resolution with explanations)
- The function is synchronous (Go WASM is single-threaded); wrap in 150ms debounce

**Base-path for dual delivery:** Set `BASE_PATH=/debateos` (or the actual repo name) at build time for Pages; omit (empty string) for localhost embed. The SvelteKit `$app/paths` base import prefixes all asset URLs automatically. The go:embed FileServer serves at `/` regardless, so the CLI-served version must be built with `BASE_PATH=` (empty).

**Resolution:** Build the UI twice: once with `BASE_PATH=/debateos` for Pages deployment, once with `BASE_PATH=` for cli/embed/. Both builds are identical except for the base path. The embedded build is the localhost build.

### Pattern 3: go:embed FileServer for CLI serve

**What:** Embed the SvelteKit `build/` directory into the Go binary and serve it with `net/http`. [VERIFIED: codebase pattern + search]

```go
// cli/embed/embed.go
package embed

import "embed"

//go:embed all:web
var WebFS embed.FS
```

```go
// cli/compose/serve.go (new, Phase 5 extension)
import (
  "io/fs"
  "net/http"
  debateosweb "github.com/mikl0s/debateos/cli/embed"
)

func ServeUI(addr string) error {
  sub, err := fs.Sub(debateosweb.WebFS, "web")
  if err != nil {
    return err
  }
  http.Handle("/", http.FileServer(http.FS(sub)))
  return http.ListenAndServe(addr, nil)
}
```

The `cli/embed/` directory contains the SvelteKit `build/` output (copied at build time). The `//go:embed all:web` directive embeds all files under `cli/embed/web/`. The `fs.Sub` strips the `web/` prefix so files are served at root.

**Content-Type for WASM:** `net/http` auto-detects `.wasm` → `application/wasm` from Go 1.17+. No extra config needed. [VERIFIED: Go standard library behavior]

### Pattern 4: Forum Go Service (chi + sqlc + modernc sqlite)

**What:** A read-mostly Go HTTP service with SQLite FTS5 search, GitHub OAuth, and a `Store` interface that abstracts all DB operations.

```go
// forum/store/store.go — the Store interface
type Store interface {
  SearchPoints(ctx context.Context, q string, foundation string, limit int) ([]PointEntry, error)
  GetPoint(ctx context.Context, id string) (*PointEntry, error)
  ListPoints(ctx context.Context, limit, offset int) ([]PointEntry, error)
  UpsertPoint(ctx context.Context, p PointEntry) error

  AddSubscription(ctx context.Context, userID, pointID string) error
  RemoveSubscription(ctx context.Context, userID, pointID string) error
  GetSubscriptions(ctx context.Context, userID string) ([]PointEntry, error)

  SetRating(ctx context.Context, userID, pointID string, stars int) error
  GetRatings(ctx context.Context, pointID string) (RatingSummary, error)

  GetConflicts(ctx context.Context, pointA, pointB string) ([]ConflictThread, error)
  UpsertConflictThread(ctx context.Context, t ConflictThread) error

  // For tests: wipe state
  Truncate(ctx context.Context) error
}
```

```yaml
# forum/sqlc.yaml
# Source: https://docs.sqlc.dev/en/stable/tutorials/getting-started-sqlite.html [CITED]
version: "2"
sql:
  - engine: "sqlite"
    queries: "query.sql"
    schema: "migrations/001_init.sql"
    gen:
      go:
        package: "store"
        out: "store/generated"
        emit_json_tags: true
        emit_interface: true
```

```sql
-- forum/migrations/001_init.sql (excerpt)
CREATE TABLE IF NOT EXISTS points (
  id          TEXT PRIMARY KEY,
  name        TEXT NOT NULL,
  intent      TEXT,
  curator     TEXT,
  foundation_compat TEXT,  -- JSON array of foundation IDs
  commit_date TEXT,         -- ISO 8601
  subscribers INTEGER DEFAULT 0,
  avg_rating  REAL DEFAULT 0,
  rating_count INTEGER DEFAULT 0
);

-- FTS5 virtual table for full-text search
-- VERIFIED: modernc.org/sqlite v1.46.1 has FTS5 compiled in (tested empirically)
CREATE VIRTUAL TABLE IF NOT EXISTS points_fts USING fts5(
  name, intent, curator, id UNINDEXED,
  content='points', content_rowid='rowid'
);

CREATE TABLE IF NOT EXISTS subscriptions (
  user_id   TEXT NOT NULL,
  point_id  TEXT NOT NULL REFERENCES points(id),
  PRIMARY KEY (user_id, point_id)
);

CREATE TABLE IF NOT EXISTS ratings (
  user_id   TEXT NOT NULL,
  point_id  TEXT NOT NULL REFERENCES points(id),
  stars     INTEGER NOT NULL CHECK(stars BETWEEN 1 AND 5),
  PRIMARY KEY (user_id, point_id)
);

CREATE TABLE IF NOT EXISTS conflict_threads (
  id         TEXT PRIMARY KEY,
  point_a    TEXT NOT NULL,
  point_b    TEXT NOT NULL,
  status     TEXT DEFAULT 'open',
  patch_pr_url TEXT,
  created_at TEXT DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);
```

```go
// forum/store/sqlite.go — open with modernc.org/sqlite driver
import (
  "database/sql"
  _ "modernc.org/sqlite"
)

func Open(dsn string) (*sql.DB, error) {
  // dsn = ":memory:" for tests, "forum.db" for production
  db, err := sql.Open("sqlite", dsn)
  if err != nil {
    return nil, err
  }
  // SQLite WAL mode for concurrent reads
  db.Exec("PRAGMA journal_mode=WAL")
  db.Exec("PRAGMA foreign_keys=ON")
  return db, nil
}
```

### Pattern 5: GitHub OAuth (golang.org/x/oauth2 + fake provider for tests)

**What:** Server-side GitHub OAuth web flow. Tests use a custom `oauth2.Endpoint` pointing to an `httptest.Server` fake provider.

```go
// forum/api/oauth.go
import "golang.org/x/oauth2"
import "golang.org/x/oauth2/github"

type OAuthProvider interface {
  AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
  Exchange(ctx context.Context, code string) (*oauth2.Token, error)
  GetUserID(ctx context.Context, token *oauth2.Token) (string, error)
}

// RealGitHubOAuth wraps golang.org/x/oauth2 for production
type RealGitHubOAuth struct {
  cfg *oauth2.Config
}

func NewRealGitHubOAuth(clientID, clientSecret, callbackURL string) *RealGitHubOAuth {
  return &RealGitHubOAuth{cfg: &oauth2.Config{
    ClientID:     clientID,
    ClientSecret: clientSecret,
    RedirectURL:  callbackURL,
    Scopes:       []string{"read:user"},
    Endpoint:     github.Endpoint,
  }}
}
```

```go
// forum/api/oauth_fake_test.go — fake provider
import "net/http/httptest"

type FakeOAuthProvider struct {
  UserID string
}

func (f *FakeOAuthProvider) AuthCodeURL(state string, _ ...oauth2.AuthCodeOption) string {
  return "/fake-oauth?state=" + state
}
func (f *FakeOAuthProvider) Exchange(_ context.Context, _ string) (*oauth2.Token, error) {
  return &oauth2.Token{AccessToken: "fake-token"}, nil
}
func (f *FakeOAuthProvider) GetUserID(_ context.Context, _ *oauth2.Token) (string, error) {
  return f.UserID, nil
}
```

### Pattern 6: Registry Index Generator

**What:** Go command that scans fixture/examples directories, parses each point/speech via `resolver/parse`, computes foundation-compat from `capabilities.json`, and emits deterministic JSON + HTML.

**Foundation-compat algorithm:**
1. Load each translator's `capabilities.json` → `map[string][]string` (foundation → capability set)
2. For each opinion in the point's members, collect `translator_capabilities` tokens
3. For each foundation: count how many required tokens appear in that foundation's capability set
4. A point is "compatible" with a foundation if all required opinion tokens are in the capability set
5. Emit per-point `foundation_compat: ["arch", "debian"]` or `["arch"]` etc.

```go
// registry/index/compat.go
type FoundationCompat struct {
  Foundation string   `json:"foundation"`
  Compatible bool     `json:"compatible"`
  Missing    []string `json:"missing_capabilities,omitempty"`
}

func ComputeCompat(point *resolver.Point, opinions []resolver.Opinion,
  caps map[string][]string) []FoundationCompat {
  // caps: map[foundationID] → []capability_token
  // For each foundation, check if ALL required opinion translator_capabilities
  // are in the capability set.
}
```

**Determinism:** Sort all output slices by ID before marshaling. Use `encoding/json` with sorted keys (the default for struct fields). Use `time.Time` from git commit metadata for `commit_date`. Identical inputs produce byte-identical JSON. [CITED: existing determinism pattern in BLD-03]

**Index JSON schema (Claude's Discretion — recommended):**
```json
{
  "schema": 1,
  "generated_at": "2026-06-13T00:00:00Z",
  "points": [
    {
      "id": "omarchy/ai-tooling",
      "name": "AI Tooling",
      "intent": "...",
      "curator": "omarchy@basecamp.com",
      "members": ["OM-010", "OM-023"],
      "foundation_compat": [
        {"foundation": "arch", "compatible": true},
        {"foundation": "debian", "compatible": false, "missing_capabilities": ["install-npm-global-packages"]}
      ],
      "commit_date": "2026-06-13T00:00:00Z",
      "tags": []
    }
  ]
}
```

### Pattern 7: Invariant-4 Gate (Forum-Offline Test)

**What:** Prove compose→resolve→build works with no Forum process running.

```bash
#!/bin/bash
# scripts/invariant4-check.sh
# Brings nothing up. Runs the CLI compose path. Asserts exit 0.
set -euo pipefail

# 1. Ensure no forum process is running
pkill -f "forum" 2>/dev/null || true

# 2. Run compose on the omarchy example (uses WASM path via go run)
# The CLI compose command resolves locally; no network needed.
go run ./cmd/debateos compose --dir examples/omarchy/

# 3. Run resolve-json to assert resolver produces output
go run ./cmd/resolve-json examples/omarchy/speech.yaml > /tmp/resolved.json
echo "Resolved: $(python3 -c "import json,sys; d=json.load(open('/tmp/resolved.json')); print(f\"Applied={len(d.get('applied',[]))} Hard-conflicts={sum(1 for e in d.get('explanations',[]) if e.get('rule')=='rule2')}\")")"

# 4. (Optional) run debateos build --skip-iso if available
# go run ./cmd/debateos build --dir examples/omarchy/ --skip-iso

echo "INVARIANT-4 CHECK PASS: compose→resolve works with Forum offline"
```

The Playwright-based invariant-4 test:
```typescript
// web/tests/invariant4.spec.ts
test('WASM compose path works with Forum offline', async ({ page }) => {
  // Block Forum domain entirely
  await page.route('**/api/**', route => route.abort());
  await page.route('**/health**', route => route.abort());

  await page.goto('/debate');
  // Wait for WASM to load
  await page.waitForSelector('[data-wasm-ready="true"]', { timeout: 10000 });

  // Inject known speech + opinions directly into the WASM
  const result = await page.evaluate(() => {
    const input = JSON.stringify({
      speech: { schema: 1, foundation: 'arch', points: [] },
      opinions: [],
      hardware: { predicates: [], facts: {}, pci_ids: [] }
    });
    const raw = (window as any).debateosResolve(input);
    return JSON.parse(JSON.parse(raw).result);
  });

  expect(result.schema).toBe(1);
  expect(result.explanations).toBeDefined();
});
```

### Anti-Patterns to Avoid

- **Committing wasm_exec.js:** Do NOT commit `/usr/local/go/lib/wasm/wasm_exec.js`. Always copy from `$(go env GOROOT)/lib/wasm/wasm_exec.js` at build time. The file changes between Go versions. [CITED: resolver/wasm/main.go T-01-15 comment]
- **SSR on the /debate route:** SvelteKit will try to run WASM during SSR (Node.js), which fails. Mark `/debate/+page.ts` with `export const ssr = false`. [CITED: SvelteKit page-options docs]
- **Using modernc.org/sqlite v1.47.0–v1.52.0 without updating go.mod:** These versions require Go 1.25. If go.mod says `go 1.24.0`, the toolchain will auto-download go1.25.x (safe but slow on first run). Pin v1.46.1 to stay on the current toolchain. [VERIFIED: go proxy version scan]
- **Making Forum a dependency of compose/resolve/build:** Invariant 4. The UI loads points from the static registry JSON when Forum is offline; the WASM resolver requires zero network. [CITED: CONTEXT.md]
- **Reimplementing resolution in JavaScript:** Invariant 3. All resolution goes through `window.debateosResolve()`. [CITED: CONTEXT.md]
- **Using PostCSS for Tailwind v4:** Tailwind v4 uses the Vite plugin (`@tailwindcss/vite`). The `postcss.config.js` approach is v3 legacy. [CITED: WebSearch teta.so]
- **Relying on color alone for conflict state:** The UI-SPEC mandates triple encoding (color + icon + text label) for accessibility (deuteranopia/protanopia). [CITED: UI-SPEC §Conflict color rules]
- **Using tailwind.config.js for v4 theme tokens:** v4 declares `@theme { }` in CSS, not in a JS config file. [CITED: WebSearch tailwindcss v4 docs]

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| SQLite FTS5 full-text search | Custom inverted index | modernc.org/sqlite FTS5 virtual table | FTS5 is compiled in (verified empirically); handles tokenization, ranking, partial match |
| GitHub OAuth web flow | Custom OAuth handshake | golang.org/x/oauth2 + github.Endpoint | State validation, token exchange, PKCE — all handled |
| Static file serving with embed | Custom file server | net/http.FileServer + embed.FS + fs.Sub | Standard library; Content-Type auto-detection including application/wasm |
| SQL type-safe queries | String-interpolated SQL | sqlc code generation | Prevents SQL injection; type-safe at compile time; compile-errors on schema change |
| SvelteKit base-path prefixing | Manual URL manipulation | `$app/paths` base import | SvelteKit automatically prefixes all asset paths when `paths.base` is set |
| Conflict resolution logic in JS | Re-implement docs/04 rules | window.debateosResolve() (WASM) | Invariant 3; JS implementation would drift from Go truth |
| Drag-and-drop library | Native HTML5 DnD | Native HTML5 drag API + Svelte actions | UI-SPEC specifies no DnD library; keeps bundle lean for WASM co-bundle |

**Key insight:** The hardest problem in this phase is WASM + SSR interaction. The solution (per-route `ssr=false`) is one line but must be applied before any SSR hydration error appears — it cannot be fixed after-the-fact without a refactor.

---

## Common Pitfalls

### Pitfall 1: wasm_exec.js version skew
**What goes wrong:** wasm_exec.js is copied from GOROOT at build time. If the file in the web bundle was built with a different Go version than the one that compiled debateos.wasm, the WASM module fails to instantiate with a cryptic error.
**Why it happens:** wasm_exec.js is internal Go runtime glue; it changes between Go releases.
**How to avoid:** Always copy from `$(go env GOROOT)/lib/wasm/wasm_exec.js` in the same build script that runs `GOOS=js GOARCH=wasm go build`. Never commit wasm_exec.js.
**Warning signs:** `TypeError: go.run is not a function` or `WebAssembly.RuntimeError: memory access out of bounds`.

### Pitfall 2: modernc.org/sqlite Go version mismatch
**What goes wrong:** Adding modernc.org/sqlite v1.47.0+ to go.mod (go 1.24.0) triggers an auto-toolchain download (go1.25.x) on first build, adding ~2 minutes and 200MB.
**Why it happens:** v1.47.0+ upgraded their module's go directive to 1.25.0.
**How to avoid:** Pin `modernc.org/sqlite@v1.46.1` (last go-1.24-compatible version; FTS5 confirmed functional).
**Warning signs:** `go: downloading go1.25.11 (linux/amd64)` in build output.

### Pitfall 3: SvelteKit SSR + WASM
**What goes wrong:** SvelteKit pre-renders pages server-side by default. Any import of WASM-related code at module level crashes the SSR render.
**Why it happens:** Node.js doesn't have `WebAssembly.instantiateStreaming` in the same form as browsers; `window` is undefined server-side.
**How to avoid:** Two guards: (1) export `ssr = false` in `/debate/+page.ts`; (2) all WASM calls go inside `onMount()`. [CITED: SvelteKit page-options docs]
**Warning signs:** `ReferenceError: window is not defined` during build or first SSR render.

### Pitfall 4: GitHub Pages subpath vs localhost root
**What goes wrong:** The GitHub Pages URL is `https://mikl0s.github.io/debateos/` (subpath `/debateos/`). The cli/embed serve is at `http://localhost:PORT/` (root). A single build won't work for both if base-path is hardcoded.
**Why it happens:** SvelteKit embeds the base path into all script/link/asset URLs at build time.
**How to avoid:** Build twice: `BASE_PATH=/debateos npm run build` for Pages; `BASE_PATH= npm run build` for the embedded CLI version. The embedded CLI version is the one committed to cli/embed/.
**Warning signs:** 404 on JS/CSS/WASM assets when served from wrong base.

### Pitfall 5: FTS5 content table sync
**What goes wrong:** The FTS5 virtual table with `content='points'` (external content) doesn't auto-update when the `points` table changes.
**Why it happens:** External content FTS5 tables require explicit sync triggers or manual ROWID insert/delete/update signals.
**How to avoid:** Either use triggers (add in migration) OR do `INSERT INTO points_fts(points_fts) VALUES('rebuild')` after bulk index. For the Forum's append-heavy re-index pattern, prefer rebuilding the FTS index explicitly after each batch upsert.
**Warning signs:** FTS5 search returns stale results after index rebuild.

### Pitfall 6: Hardware.predicates type
**What goes wrong:** JS callers pass `hardware: { predicates: {} }` (object) instead of `hardware: { predicates: [] }` (array). The WASM resolver returns a parse error JSON.
**Why it happens:** `HardwareProfile.Predicates` is `[]string` in Go; JS callers might assume it's a map.
**How to avoid:** TypeScript type for `HardwareProfile` declares `predicates: string[]`. The `debateosResolve()` wrapper function enforces the type.
**Warning signs:** `parse error (json: cannot unmarshal object into Go struct field HardwareProfile.hardware.predicates of type []string)` in the WASM output.

---

## Runtime State Inventory

Phase 5 is greenfield (new directories: web/, registry/, forum/, cli/embed/). No rename/refactor. No runtime state to migrate.

- **Stored data:** None — Forum SQLite is created fresh (schema applied on start); registry index generated fresh at build.
- **Live service config:** None — no live Forum running yet; Pages not yet deployed.
- **OS-registered state:** None.
- **Secrets/env vars:** `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET` — documented for production; NOT used in tests (FakeOAuthProvider substitutes). Must be in deployment environment, not in code.
- **Build artifacts:** `cli/embed/` directory does not exist yet — will be created this phase.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Node.js | web/ build, Vitest, Playwright | ✓ | v24.12.0 | — |
| npm | Package management | ✓ | 11.17.0 | — |
| Go | forum/, registry/, go:embed | ✓ | go1.24.1 linux/amd64 | — |
| Docker | (no ISO build in Phase 5) | ✓ | 29.5.3 | not needed |
| sqlc binary | Forum code generation | ✓ | v1.30.0 at /home/mikkel/go/bin/sqlc | Install v1.31.1 via `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` |
| Playwright chromium | WASM integration test | ✓ | 148.0.7778.96 (headless shell downloaded) | Node WASM harness (secondary) |
| $(GOROOT)/lib/wasm/wasm_exec.js | WASM build | ✓ | /usr/local/go/lib/wasm/wasm_exec.js | — |
| GitHub OAuth (live) | FORM-03 production | ✗ | — | FakeOAuthProvider in tests; deferred-live |
| GitHub Pages (live deploy) | UI-02 production | ✗ | — | Deferred-to-CI; local build verifiable |
| Oracle Cloud A1 (live) | FORM-05 production | ✗ | — | Deferred; deploy notes in forum/deploy/ |

**Missing dependencies with no fallback:** none — all blockers have fallbacks or are production-only deferrals.

**Missing dependencies with fallback:**
- Live GitHub OAuth: FakeOAuthProvider covers all test paths.
- Live Pages deploy: `BASE_PATH=/debateos npm run build` produces Pages-ready output; push step deferred.
- Live Oracle A1: `forum/deploy/oracle-a1.md` documents steps; binary cross-compiles with `GOOS=linux GOARCH=arm64`.

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Go test framework | `go test ./...` (standard library) |
| UI unit test framework | Vitest 4.1.8 |
| UI integration test | Playwright 1.60.0 + Chromium 148.0.7778.96 (headless) |
| Node WASM harness | Node.js + wasm_exec.js (secondary WASM assertion) |
| Coverage gate (Go) | forum/ ≥85%, registry/ ≥85% (mirror Phase 3–4 policy) |
| Coverage gate (UI) | Vitest —coverage (no hard gate for Phase 5; all critical logic in Go) |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| REG-01 | Registry index generator reads YAML, validates, emits deterministic JSON | Unit (Go) | `go test ./registry/... -run TestGenerateIndex -v` | ❌ Wave 0 |
| REG-01 | Foundation-compat computed from capabilities.json | Unit (Go) | `go test ./registry/... -run TestFoundationCompat -v` | ❌ Wave 0 |
| REG-01 | Registry index is deterministic (identical inputs → identical JSON) | Integration (Go) | `go test ./registry/... -run TestDeterminism -v` | ❌ Wave 0 |
| UI-01 | WASM loads and debateosResolve callable from browser | Integration (Playwright) | `cd web && npx playwright test tests/wasm-render.spec.ts` | ❌ Wave 2 |
| UI-01 | WASM resolve returns ResolvedSpeech with expected conflict state | Integration (Playwright) | `cd web && npx playwright test tests/wasm-render.spec.ts -g "conflict"` | ❌ Wave 2 |
| UI-01 | ConflictOverlay triple encoding (A1): color + icon + text | Playwright + CSS | `cd web && npx playwright test tests/a11y.spec.ts -g "A1"` | ❌ Wave 2 |
| UI-01 | Explanation.Text rendered verbatim (A9) | Playwright | `cd web && npx playwright test tests/wasm-render.spec.ts -g "A9"` | ❌ Wave 2 |
| UI-01 | Touch target minimum 44px (A2) | Playwright | `cd web && npx playwright test tests/a11y.spec.ts -g "A2"` | ❌ Wave 2 |
| UI-02 | go:embed serve: static files served at localhost root | Integration (Go) | `go test ./cli/embed/... -run TestServeUI -v` | ❌ Wave 3 |
| UI-02 | WASM binary served with correct Content-Type: application/wasm | Integration (Go) | `go test ./cli/embed/... -run TestWasmContentType -v` | ❌ Wave 3 |
| FORM-01 | FTS5 search returns matching points | Unit (Go) | `go test ./forum/... -run TestFTS5Search -v` | ❌ Wave 3 |
| FORM-01 | Client-side fallback filter works over static JSON (Forum offline) | Unit (Vitest) | `cd web && npx vitest run src/lib/filter.test.ts` | ❌ Wave 3 |
| FORM-02 | Subscribe/unsubscribe round-trips in SQLite | Unit (Go) | `go test ./forum/... -run TestSubscriptions -v` | ❌ Wave 3 |
| FORM-03 | Rating write requires OAuth token; fake provider gates access | Unit (Go) | `go test ./forum/... -run TestRatings -v` | ❌ Wave 3 |
| FORM-04 | Conflict thread create/retrieve with patch PR URL | Unit (Go) | `go test ./forum/... -run TestConflictThreads -v` | ❌ Wave 3 |
| FORM-05 | Invariant-4: compose→resolve with Forum DOWN | Script + Playwright | `bash scripts/invariant4-check.sh` | ❌ Wave 4 |
| FORM-05 | Forum re-index command runs to completion (no crash) | Unit (Go) | `go test ./forum/... -run TestReindex -v` | ❌ Wave 3 |
| BRND-01 | No forbidden terms in UI copy (A6) | Playwright | `cd web && npx playwright test tests/brand.spec.ts -g "A6"` | ❌ Wave 2 |
| BRND-01 | Typography scale (A7): only {13,14,16,20,28}px | Playwright + CSS | `cd web && npx playwright test tests/a11y.spec.ts -g "A7"` | ❌ Wave 2 |

### Deferred / Host-CI Items (not run locally)

| Test | Why Deferred | Where It Runs |
|------|-------------|---------------|
| Live GitHub OAuth callback | Requires live GitHub app registration | CI + documented-deferred |
| Live Pages deploy check | Requires GitHub Pages activation | CI workflow (BLD-02 pattern) |
| Live Oracle A1 binary test | No Oracle instance; ARM cross-compile only locally | Owner deploy + smoke |
| Live Actions index rebuild | Requires GitHub repo + push | CI (mirrors Phase 3 BLD-02 policy) |
| Forum load test | Out of scope for v1 | Post-v1.0 |

### Sampling Rate

- **Per task commit:** `go test ./... -count=1` (Go) + `cd web && npx vitest run` (UI unit)
- **Per wave merge:** full suite: `go test ./... -count=1 -coverprofile=c.out && go tool cover -func=c.out` + `cd web && npx playwright test`
- **Phase gate:** Full suite green before `/gsd-verify-work`. Coverage: forum/ ≥85%, registry/ ≥85%.

### Wave 0 Gaps

- [ ] `web/package.json` — SvelteKit + Tailwind v4 + Vitest + Playwright deps
- [ ] `web/svelte.config.js` — adapter-static + paths.base
- [ ] `web/vite.config.ts` — @tailwindcss/vite + sveltekit
- [ ] `web/playwright.config.ts` — test dir, chromium browser, baseURL
- [ ] `web/src/app.css` — @import tailwindcss + @theme tokens + dark variant
- [ ] `web/static/.nojekyll` — GitHub Pages Jekyll bypass
- [ ] `Makefile` or `scripts/build-wasm.sh` — GOOS=js GOARCH=wasm build + wasm_exec.js copy
- [ ] `registry/` Go package scaffolding
- [ ] `forum/` Go package scaffolding + `sqlc.yaml`
- [ ] `cli/embed/` directory creation + `embed.go`
- [ ] `go test ./forum/... -run TestFTS5Smoke` — confirms FTS5 functional before writing schema

---

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | yes (Forum OAuth) | golang.org/x/oauth2 + GitHub OAuth only; no passwords |
| V3 Session Management | yes (Forum sessions) | HTTPS-only session cookie; httpOnly; SameSite=Lax |
| V4 Access Control | yes (write operations) | All writes gated on OAuth session; reads are public |
| V5 Input Validation | yes (search query, rating value) | chi middleware; sqlc parameterized queries (no raw SQL interpolation) |
| V6 Cryptography | minimal | HTTPS termination at reverse proxy; no secrets stored; OAuth token not persisted |

### Known Threat Patterns for this Stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| SQL injection via search query | Tampering | sqlc parameterized queries (`?` placeholders — never string interpolation) |
| CSRF on rating/subscribe writes | Tampering | CSRF token (double-submit cookie or SameSite=Strict session cookie) |
| OAuth state parameter forgery | Spoofing | Validate `state` parameter in /oauth/callback; use `crypto/rand` for state generation |
| Stored XSS in point names/intent | Tampering | Content served as JSON; UI renders as text nodes (Svelte auto-escapes) |
| Arbitrary file upload | Tampering | Forum has no upload endpoint; registry reads only from local fixtures |
| WASM supply chain | Tampering | WASM built from monorepo source; not fetched from CDN |
| Secrets in shared artifacts | Information Disclosure | Forum stores no private pane data; OAuth token used then discarded (not stored) |

**Security invariants (D13, mandatory):**
1. No untrusted code execution on Forum server.
2. No passwords, email, or 2FA stored — GitHub OAuth only.
3. No secrets at rest — OAuth access tokens are used for user ID lookup then discarded.
4. Total DB loss → re-index from GitHub → no data loss beyond ephemeral social state.

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Tailwind PostCSS config | @tailwindcss/vite Vite plugin | Tailwind v4 (2024–2025) | No postcss.config.js; CSS @theme replaces tailwind.config.js |
| tailwind.config.js for tokens | @theme { } in CSS | Tailwind v4 | Design tokens defined in CSS, not JS |
| SvelteKit adapter-static v2 | adapter-static v3 | SvelteKit 2 era | API compatible; same configuration |
| Svelte 4 Options API | Svelte 5 Runes | 2024 | `$props()`, `$state()`, `$derived()` replace `export let`, stores in components |
| SQLite WAL not default | WAL mode strongly recommended | SQLite 3.x | `PRAGMA journal_mode=WAL` for concurrent read performance |

**Deprecated/outdated:**
- `svelte:options` with `accessors` — replaced by Svelte 5 runes
- `tailwind.config.js` JavaScript config — v4 uses CSS `@theme` directive
- `@tailwindcss/postcss` — v4 uses `@tailwindcss/vite` for Vite projects

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `paths.base` in SvelteKit svelte.config.js uses `process.env.BASE_PATH` pattern for dual-delivery | Dual delivery, Pattern 1 | If SvelteKit doesn't support env-var base at build time, need a different dual-build strategy |
| A2 | FTS5 external content table sync requires explicit rebuild after upsert | Pitfall 5 | If modernc FTS5 auto-syncs, the explicit rebuild is redundant (harmless) |
| A3 | Playwright chromium-headless-shell does not need a full X11 display server | Environment Availability | Confirmed working (headless mode); no Xvfb needed |
| A4 | Inter variable font via @fontsource/inter self-hosts all weights in one npm package | Standard Stack | If package only includes specific weights, may need additional weight variants |
| A5 | `net/http` auto-detects `.wasm` content-type as `application/wasm` | go:embed pattern | If wrong, wasm loading fails; fix: set explicit content-type header in FileServer wrapper |

**Confirmed / not assumed:**
- FTS5 works in modernc.org/sqlite v1.46.1: VERIFIED empirically (test ran, FTS5 search returned results)
- Playwright Chromium works on this host: VERIFIED empirically (1 test passed in 864ms)
- Node.js WASM harness works: VERIFIED empirically (`debateosResolve()` called successfully)
- debateos.wasm builds to 4.3MB: VERIFIED (`GOOS=js GOARCH=wasm go build ./resolver/wasm/`)
- modernc.org/sqlite v1.52.0 requires go 1.25: VERIFIED (go.mod scan, empirical test)
- wasm_exec.js path: VERIFIED (`/usr/local/go/lib/wasm/wasm_exec.js` exists)

---

## Open Questions (RESOLVED)

> RESOLVED 2026-06-13 by the orchestrator under the autonomous-run mandate: (1) use the explicit TWO-build approach (BASE_PATH=/debateos for Pages, BASE_PATH= for cli/embed) — safe and explicit beats a relative-path gamble; revisit single-build post-v1.0; (2) modernc.org/sqlite v1.46.1 ARM cross-compile to linux/arm64 is CGo-free so assumed to work — the Oracle A1 deploy is itself a deferred-to-host item, so confirm during that deploy (document the GOARCH=arm64 build command in forum/deploy notes). Originals retained below.

1. **Base path for dual delivery: one build or two?**
   - What we know: `paths.base` is baked into the build output; GitHub Pages subpath differs from localhost root.
   - What's unclear: Whether a single build with relative paths (`./`) could work at both subpath and root, avoiding two builds.
   - Recommendation: Build twice (separate outputs for Pages vs cli/embed/); the embedded CLI build uses `BASE_PATH=` (empty = root). Two builds adds ~30s to the workflow but is explicit and correct.

2. **sqlc version: 1.30.0 (installed) vs 1.31.1 (latest)?**
   - What we know: v1.30.0 is installed locally; v1.31.1 is the latest stable.
   - What's unclear: Any sqlite-specific fixes in 1.31.1 that affect our schema.
   - Recommendation: Use installed v1.30.0 for development (avoid toolchain drift). Document upgrade path in forum/README.

3. **Forum ARM cross-compilation for Oracle A1?**
   - What we know: Oracle A1 is ARM (linux/arm64); Go supports cross-compile with `GOOS=linux GOARCH=arm64`.
   - What's unclear: Whether modernc.org/sqlite v1.46.1 cross-compiles cleanly to arm64 (it's CGo-free so it should).
   - Recommendation: Add `GOOS=linux GOARCH=arm64 go build ./forum/cmd/forumctl` to the deploy docs; note it as untested until CI runs on the target.

---

## Sources

### Primary (HIGH confidence — verified/cited from authoritative sources)

- `resolver/wasm/main.go` — WASM contract, js.Func registration, input/output types [VERIFIED: codebase read]
- `resolver/resolve/explanation.go` — `ResolvedSpeech`, `Explanation` struct definitions [VERIFIED: codebase read]
- `translators/arch/capabilities.json` + `translators/debian/capabilities.json` — foundation capability sets [VERIFIED: codebase read]
- Go WASM build: `GOOS=js GOARCH=wasm go build ./resolver/wasm/ -o /tmp/debateos_test.wasm` → 4.3MB [VERIFIED: run on host]
- FTS5 in modernc.org/sqlite v1.46.1: functional test passed (search result returned) [VERIFIED: empirical]
- Playwright chromium headless: `1 passed (864ms)` smoke test [VERIFIED: empirical]
- Node.js WASM harness: `debateosResolve()` called, valid `ResolvedSpeech` returned [VERIFIED: empirical]
- npm registry: svelte@5.56.3, @sveltejs/kit@2.65.0, adapter-static@3.0.10, tailwindcss@4.3.1, vitest@4.1.8, @playwright/test@1.60.0 [VERIFIED: `npm view`]
- Go proxy: chi/v5@v5.3.0, modernc.org/sqlite version list, oauth2@v0.36.0 [VERIFIED: Go proxy API]
- wasm_exec.js: `/usr/local/go/lib/wasm/wasm_exec.js` exists [VERIFIED: filesystem]

### Secondary (MEDIUM confidence — cited from official documentation)

- SvelteKit adapter-static config: https://svelte.dev/docs/kit/adapter-static [CITED]
- SvelteKit page options (ssr=false): https://svelte.dev/docs/kit/page-options [CITED]
- sqlc SQLite config: https://docs.sqlc.dev/en/stable/tutorials/getting-started-sqlite.html [CITED]
- Tailwind v4 + SvelteKit Vite setup: https://teta.so/blog/sveltekit-tailwind-css-setup-best-practices [CITED]
- go:embed FileServer pattern: https://eli.thegreenplace.net/2022/serving-static-files-and-web-apps-in-go/ [CITED]
- docs/11-repo-layout.md, docs/06-social-layer.md, docs/05-distribution-and-infra.md [CITED: project docs]
- UI-SPEC (05-UI-SPEC.md): design contract, component inventory, copywriting [CITED: project docs]

### Tertiary (LOW confidence — training knowledge, noted)

- Oracle Cloud Always Free Ampere A1 ARM binary deployment pattern [ASSUMED — not tested this session; well-documented externally]

---

## Metadata

**Confidence breakdown:**
- WASM integration: HIGH — empirically tested on this host
- SvelteKit + Tailwind v4 stack: HIGH — versions confirmed via npm registry; docs verified
- Forum stack (chi + sqlite + sqlc + oauth2): HIGH — versions confirmed via Go proxy; FTS5 empirically verified
- Registry generator: HIGH — reuses existing resolver/parse package (tested in prior phases)
- Playwright on this host: HIGH — installed and smoke-tested this session
- Dual-delivery base-path: MEDIUM — one assumption about env-var base approach; fallback is two builds

**Research date:** 2026-06-13
**Valid until:** 2026-07-13 (stable stack; Tailwind v4 and SvelteKit 2 are not fast-moving at this point)
