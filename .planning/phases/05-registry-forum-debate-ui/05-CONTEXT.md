# Phase 5: Registry, Forum & Debate UI - Context

**Gathered:** 2026-06-13
**Status:** Ready for planning
**Mode:** Smart discuss (autonomous — recommended answers auto-accepted per owner directive + ADR process notes)

<domain>
## Phase Boundary

Phase 5 ships the discovery + composition layer: the static Git-backed registry index (`registry/`, REG-01), the SvelteKit static Debate UI running the Go-WASM resolver client-side (`web/`, UI-01/UI-02, delivered via GitHub Pages AND embedded in `debateos compose` per cli/embed/), the optional Forum service (`forum/`, Go chi + SQLite, FORM-01..05), and the debate-themed brand voice (BRND-01). This is the final v1.0 phase.

Out of scope: anything post-v1.0 (Phase 6 hardware-scanning installer, Fedora translator, direct-to-disk, full GitLab parity, post-install reconciliation); making the Forum required for any compose→resolve→build step (invariant 4 — it is strictly optional/additive).

</domain>

<decisions>
## Implementation Decisions

### Registry Index (REG-01, D12)
- `registry/` is a Go static-site generator: reads point/speech YAML from configured GitHub repo(s) (for v1 + tests, read from local fixture repos + examples/; live multi-repo crawl documented), validates each via resolver/parse, emits a static JSON index (curator, tags, version, foundation-compatibility, freshness/commit-date) + minimal static HTML browse pages consumable by GitHub Pages.
- Git remains authoritative (invariant): the index is a derived cache; regeneration is idempotent and deterministic. A GitHub Actions workflow rebuilds the index on commit (mirrors Phase 3 channel pattern; live run deferred-to-CI like prior phases).
- Foundation-compatibility per point/speech is computed from the translators' capabilities.json (which foundations can effectuate the required opinions) — reuse the capability data, do not duplicate.

### Debate UI (UI-01/UI-02, D10, BRND-01)
- SvelteKit + adapter-static + Tailwind; static output. Runs the Go-WASM resolver (resolver/wasm from Phase 1) client-side — the UI NEVER reimplements resolution (invariant 3); it calls the WASM build and renders Explanation output.
- Core flow: load points (from the static registry index / local), compose a speech across visual "panes" (glass panes over a foundation), live conflict visualization (red = hard conflict, green = compatible overlap) driven by resolve() results, then produce build instructions (the `debateos build` command + downloadable resolved speech YAML). v1 UI does NOT itself build ISOs (no central build service — invariant 4); it hands off to the CLI/Actions.
- Dual delivery: the same build output deploys to GitHub Pages AND is embedded under cli/embed/ so `debateos compose` serves it on localhost offline (wire the Phase 3 serve hook point now). Byte-identical UI both places.
- Brand voice (BRND-01): debate/rhetoric metaphor (opinions/points/speeches/debates; "That's just your opinion, man"; "no conclusions, only debates") applied across UI copy + docs, softened only where it would obscure meaning. Apply, don't overdo.
- WASM/native parity already guaranteed by Phase 1 tests; the UI consuming WASM is validated by a Playwright (or minimal headless) test that loads a known speech and asserts the rendered conflict matches the resolver's expected Explanation.

### Forum (FORM-01..05, D13/D13a/D14/D15)
- `forum/` Go service: chi router + embedded SQLite via the thin `store` interface + sqlc-generated queries (modernc.org/sqlite pure-Go, libSQL-compatible); in-memory SQLite for tests. FTS5 search abstracted behind the store for a later Postgres tsvector swap. Layout per docs/11: forum/{api,index,store,migrations,deploy}.
- FORM-01 search/discovery: search indexed points/speeches by curator, tag, popularity, freshness, foundation compatibility (FTS5). FORM-02 subscriptions: follow curators / subscribe to point sets or individual points (subscription edges in SQLite). FORM-03 ratings/reputation: GitHub-OAuth-identity-backed ratings; NO native accounts (D13). FORM-04 conflict threads: registry of known conflicts + links to resolving patch-opinion PRs (GitHub URLs); the patch opinions themselves live in Git (survive Forum loss). FORM-05 boundaries: Forum is an index over GitHub — NOT system-of-record, NOT a build service (no code execution ever), NOT an account system (OAuth only), NOT required to compose/build.
- Security (D13, mandatory): read-mostly index; no arbitrary uploads; no untrusted code execution; GitHub OAuth only (no passwords/email/2FA); no secrets at rest; rebuildable (total DB loss → re-index GitHub); single static linux/arm64 binary + one SQLite file. The re-index path is a tested command.
- Hosting (D15): deploy notes for Oracle Cloud Always Free Ampere A1 (ARM, EU/APAC region), owner-server fallback. Deploy = single static binary + SQLite file; docs only (no live deploy in this phase).
- GitHub OAuth: implement the OAuth web flow with a Runner/HTTP-client interface so tests use a fake OAuth provider (no live GitHub calls in tests); document the real app-registration steps.

### Invariant 4 Gate (the phase's hardest guarantee)
- An automated test/script proves the compose→resolve→build path works with the Forum process DOWN: the Debate UI (static, WASM resolver) + registry index (static) + CLI build need zero Forum availability. Forum-offline test = bring nothing up, run the core path, assert success.

### TDD (D19)
- Go (registry generator, forum api/store/index) → table-driven TDD, RED before GREEN; sqlc-generated queries tested against in-memory SQLite. UI logic → Vitest unit tests + one Playwright/headless integration test for the WASM-conflict render; RED before GREEN where the harness supports it.
- New deps justified: SvelteKit/Tailwind/adapter-static (UI, D10-locked), modernc.org/sqlite + sqlc (Forum storage, D13a-locked), chi (Forum router, D13-locked), an OAuth2 lib (golang.org/x/oauth2). All ADR-sanctioned. Coverage gates extended (forum/registry Go ≥85%).
- Slow/host-limited paths (live OAuth, live Pages deploy, live Oracle deploy, live Actions index rebuild) = deferred-to-host/CI, documented (consistent policy with Phases 2-4). The WASM-in-browser Playwright test runs headless on this host if Chromium is installable; if not, fall back to a Node-based WASM harness test + document the browser smoke as deferred.

### Claude's Discretion
- web/ component structure, Tailwind theme tokens, sqlc query organization, exact JSON index schema, FTS5 schema details, brand-copy wording.

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- resolver/wasm (Phase 1 — the WASM resolver the UI calls; debateosResolve js.Func) + scripts/wasm-parity-test.sh.
- resolver/parse (registry validates YAML via it); resolver/resolve Explanation (UI renders it).
- translators/*/capabilities.json (foundation-compatibility source for the index).
- cli/ (Phase 3 — `debateos compose` serve hook point for embedding the UI; cli/embed/ target per docs/11).
- examples/ (omarchy, dual-foundation) + registry fixtures for index generation + UI test data.
- cmd/resolve-json (resolved-speech helper) + the build command (UI's "proceed to build" hands off to it).

### Established Patterns
- TDD RED/GREEN; coverage gates; minimal/ADR-sanctioned deps; deterministic outputs; environment-blocked paths documented-deferred; Runner/interface seams for external calls (mirror for OAuth/HTTP); invariant 1 (no distro mechanics) — UI/registry/Forum are foundation-agnostic.
- Go module github.com/mikl0s/debateos (one k, one l). Licensing: code AGPL (root LICENSE); schemas/examples CC0.

### Integration Points
- web/ build output consumed twice: GitHub Pages + go:embed under cli/embed/ (the dual-delivery contract, docs/11).
- Forum reads ONLY already-public GitHub content; never in the build critical path (invariant 4).
- registry index feeds both the UI (point discovery) and the Forum (what it indexes).

</code_context>

<specifics>
## Specific Ideas

- The phase's defining guarantee: with the Forum offline the whole compose→resolve→build path still works (invariant 4) — make this an explicit automated check.
- UI calls the WASM resolver, never reimplements resolution (invariant 3).
- Forum executes NO untrusted code, holds NO secrets, GitHub OAuth only, rebuildable from GitHub (D13).
- Brand voice applied but never at the cost of clarity (BRND-01).
- Host: node 24 + npm 11 + Go available, Docker yes, no QEMU/devtmpfs ISO builds, no live GitHub OAuth/Pages/Oracle in tests (fakes + documented deferrals).
</specifics>

<deferred>
## Deferred Ideas

- Live GitHub OAuth app, live Pages deploy, live Oracle A1 deploy, live Actions index rebuild → deferred-to-host/CI (documented, fakes/tests cover logic).
- Postgres tsvector backend → post-v1.0 (store interface accommodates it; SQLite FTS5 in v1).
- GitLab registry parity → post-v1.0 (GitHub is the v1 bootstrap target, D12).
- Browser-based Playwright smoke if Chromium uninstallable → deferred; Node WASM harness substitutes.

</deferred>
