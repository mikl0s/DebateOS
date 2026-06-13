---
phase: 05-registry-forum-debate-ui
plan: "04"
subsystem: web
tags: [sveltekit, wasm, playwright, e2e, conflict-viz, accessibility, brnd]
dependency_graph:
  requires:
    - web/src/lib/wasm.ts (debateosResolve call surface — 05-02)
    - web/src/lib/types.ts (Explanation/ResolvedSpeech TS types — 05-02)
    - resolver/resolve/explanation.go (Explanation.Rule values + JSON tags)
    - resolver/wasm/main.go (window.debateosResolve JS contract)
    - scripts/build-wasm.sh (WASM binary for e2e runtime)
  provides:
    - web/src/lib/conflict.ts (pure mapExplanation() — A1/A9 mapping)
    - web/src/lib/components/ConflictOverlay.svelte (triple-encoded conflict viz — A1)
    - web/src/lib/components/ExplanationCard.svelte (verbatim Explanation.text — A9)
    - web/src/routes/debate/+page.svelte (WASM-driven compose, 150ms debounce — A3)
    - web/tests/wasm-render.spec.ts (Playwright WASM-render e2e — UI-02 gate)
    - web/tests/a11y.spec.ts (A2/A6/A7 assertions)
  affects:
    - web/** (debate compose flow, export screen, all conflict components)
    - web/package.json (test:unit + test:e2e scripts)
tech_stack:
  added:
    - "@lucide/svelte icons: AlertTriangle, CheckCircle2, Cpu, Puzzle, Info, XCircle"
    - "Playwright 1.50.1 Chromium headless (confirmed working on this host)"
  patterns:
    - TDD RED/GREEN for conflict.ts (conflict.test.ts committed before implementation)
    - Pure mapExplanation() — A9 text verbatim guarantee enforced at data layer
    - data-conflict-state attr on ConflictOverlay for Playwright DOM assertions
    - window.debateAddTestPane / window.debateGetResolved seams for e2e injection
    - page.route() Forum abort — A3 Forum-offline compose invariant 4 gate
key_files:
  created:
    - web/src/lib/conflict.ts
    - web/src/lib/conflict.test.ts
    - web/src/lib/stores/speech.ts
    - web/src/lib/components/WasmLoadGate.svelte
    - web/src/lib/components/FoundationBar.svelte
    - web/src/lib/components/ConflictBadge.svelte
    - web/src/lib/components/ConflictOverlay.svelte
    - web/src/lib/components/ExplanationCard.svelte
    - web/src/lib/components/ResolutionPanel.svelte
    - web/src/lib/components/PaneCard.svelte
    - web/src/lib/components/DebateStage.svelte
    - web/src/routes/export/+page.svelte
    - web/tests/wasm-render.spec.ts
    - web/tests/a11y.spec.ts
    - web/tests/fixtures/conflicting-speech.json
  modified:
    - web/src/routes/debate/+page.svelte (stub → full WASM-driven compose)
    - web/package.json (added test:unit + test:e2e scripts)
decisions:
  - "[05-04]: window.debateAddTestPane / window.debateGetResolved seams exposed by +page.svelte for Playwright injection — no external API needed"
  - "[05-04]: JSON import attribute `with { type: 'json' }` required for Playwright test ESM fixture import (Node 22 module spec)"
  - "[05-04]: conflict.ts mapExplanation() returns state='info' fallback for unknown rules — label=rule name for debug visibility"
  - "[05-04]: data-conflict-bg-rgb attr on ConflictOverlay exposes base RGB separately from opacity-blended background-color for Playwright color assertions"
  - "[05-04]: Triple encoding: color (background + border) + icon (Lucide shape) + text label — neither color alone nor icon alone satisfies A1"
  - "[05-04]: export/+page.svelte shows static YAML example; full debate-store wiring deferred to debate-store integration in 05-05/05-06"
metrics:
  duration: ~40 min
  completed_date: "2026-06-13T16:06:00Z"
  tasks_completed: 2
  files_created: 15
  files_modified: 2
  commits:
    - "35124f6: test(05-04): TDD RED — conflict.test.ts (A1/A9 assertions)"
    - "313f5d7: feat(05-04): Task 1 — debate components + conflict mapping + export screen"
    - "08d7e36: feat(05-04): Task 2 — Playwright WASM-render e2e (A1/A2/A3/A6/A7/A9) GREEN"
---

# Phase 5 Plan 04: Debate Compose UI + WASM-render e2e Summary

**One-liner:** WASM-driven conflict visualization (triple-encoded ConflictOverlay), pure mapExplanation() with verbatim A9 text, ResolutionPanel + export screen, 13 Playwright e2e tests all GREEN (A1/A2/A3/A6/A7/A9).

---

## What Was Built

### Task 1: Debate components + conflict mapping + export screen (TDD GREEN)

**Commits:** `35124f6` (RED), `313f5d7` (GREEN)

**conflict.ts** — Pure `mapExplanation(e: Explanation) → ConflictView`:
- Maps every `Explanation.Rule` to `{ state, icon, label }` triple per UI-SPEC:
  - `rule2` → `{ state:'hard', icon:'AlertTriangle', label:'Hard conflict' }`
  - `rule1`/`rule3` → `{ state:'warn', icon:'Info', label:'Will be dropped' }`
  - `hardware-skip` → `{ state:'hardware', icon:'Cpu', label:'Hardware mismatch' }`
  - `no-conflict` → `{ state:'compat', icon:'CheckCircle2', label:'Compatible' }`
  - `rule4` → `{ state:'patch', icon:'Puzzle', label:'Patch applied' }`
  - `cycle` → `{ state:'hard', icon:'AlertTriangle', label:'Circular dependency' }`
  - fallback → `{ state:'info', icon:'Info', label: rule }` (unknown rules visible in debug)
- A9 contract: `view.text = e.text` verbatim — guarded by 3 unit tests + Playwright assertion
- `hasPatch`/`patchOffered` derived from `e.patch_offered` non-empty check (TestPatchBadge)

**stores/speech.ts** — Svelte store with `{ foundation, panes, hardware }`:
- `addPane(pointId, name, opinions)` / `removePane(paneId)` / `resetDebate()` / `setFoundation()`
- `derived` stores: `allOpinions` (flat opinion list), `paneCount`

**Components** (all per UI-SPEC):
- `WasmLoadGate.svelte`: `role="status"` `aria-live="polite"` + `data-wasm-ready` attr; UI-SPEC copy verbatim
- `FoundationBar.svelte`: 56px strip (`--height-foundation-bar`), canonical vocabulary ("Foundation", not "distro")
- `ConflictBadge.svelte`: triple-encoded badge (color via CSS var + `@lucide/svelte` icon + text label)
- `ConflictOverlay.svelte`: `data-conflict-state` attr, `role="status"` `aria-label`, `data-conflict-bg-rgb`, rgba background at specified opacities
- `ExplanationCard.svelte`: rule badge pill + verbatim `.explanation-text` paragraph + kept/dropped rows + patch row + trust-warning banner
- `ResolutionPanel.svelte`: right sidebar, `ExplanationCard[]` + `BuildReadyBanner` (accent-brand CTA to `/export/`)
- `PaneCard.svelte`: `role="region"` `aria-label="X pane"`, `.pane-header` min-height 44px, `.conflict-action-btn` 44px, inline destructive confirm (no modal)
- `DebateStage.svelte`: `FoundationBar` + `PaneStack` with empty-state copy verbatim from UI-SPEC

**debate/+page.svelte** (rebuilt from stub):
- `onMount` dynamic import of `wasm.ts` (Pitfall 3 guard — no SSR)
- 150ms debounce → `debateosResolve()` → `mapExplanation()` → `conflictViews`
- `window.debateAddTestPane` / `window.debateGetResolved` test seams for Playwright injection
- Forum-offline: no network calls for compose path (invariant 4 / A3)

**export/+page.svelte**: resolved YAML preview + `debateos build` command + BRND-01 stage names ("Settling the Debate", "Finding Your Foundation's Voice", etc.)

### Task 2: Playwright WASM-render e2e + accessibility/brand assertions

**Commit:** `08d7e36`

**tests/fixtures/conflicting-speech.json**: SDDM (`OM-015`, `status:required`) vs greetd (`OM-015-greetd`, `status:required`) with mutual `conflicts:` declarations — produces `rule2` hard conflict from WASM.

**tests/wasm-render.spec.ts** (6 tests, all GREEN):
- **A3** — `page.route()` aborts all `/api/**`, `/health**`, `/forum/**`, Forum ports; WASM loads, `debateosResolve` returns `Applied.length > 0`; conflicting fixture produces `rule2` with `out.error` set
- **A9** — WASM `Explanation[0].text` extracted; panes injected via `window.debateAddTestPane`; `.explanation-text` content verified byte-for-byte
- **A1** — `[data-conflict-state="hard"]` has red `background-color` (239 RGB), `[data-icon="AlertTriangle"]` present, visible text matches `/hard conflict/i`

**tests/a11y.spec.ts** (7 tests, all GREEN):
- **A6** — `/`, `/debate/`, `/export/` checked for `config`/`preset`/`distro`/`package set` in visible text (excluding `<pre>`/`<code>` blocks); all clean
- **A2** — `.pane-header` and `.conflict-action-btn` `getBoundingClientRect().height >= 44` on all rendered elements
- **A7** — All computed font-sizes in `{13,14,16,20,28}` px (±0.5 tolerance); all computed font-weights in `{400,500,600}`

---

## Verification Results

| Check | Result |
|-------|--------|
| `npx vitest run src/lib/conflict.test.ts` (21 tests) | PASS |
| `npx vitest run` (31 tests total) | PASS |
| `npm run build` | PASS |
| `grep -q 'ssr = false' web/src/routes/debate/+page.ts` | PASS |
| `grep -q 'data-conflict-state' web/src/lib/components/ConflictOverlay.svelte` | PASS |
| `npx playwright test tests/wasm-render.spec.ts` (6 tests) | PASS |
| `npx playwright test tests/a11y.spec.ts` (7 tests) | PASS |
| `npx playwright test` (13 tests total) | PASS — 4.0s headless |
| A3: Forum-blocked compose | PASS — WASM resolves with all Forum routes aborted |
| A9: Verbatim Explanation text | PASS — byte-identical text from WASM to `.explanation-text` |
| A1: Triple encoding | PASS — red bg + AlertTriangle icon + "Hard conflict" text all present |
| A6: No forbidden terms | PASS — all 3 routes clean |
| A2: 44px touch targets | PASS — pane-header and conflict-action-btn verified |
| A7: Typography scale | PASS — {13,14,16,20,28}px only; {400,500,600} weights only |
| bash scripts/build-wasm.sh | PASS — debateos.wasm + wasm_exec.js produced |

---

## TDD Gate Compliance

- RED commit: `35124f6` — `test(05-04)` — conflict.test.ts written (21 tests), file failed with `Failed to load url ./conflict.js`
- GREEN commit: `313f5d7` — `feat(05-04)` — conflict.ts implemented, 21/21 tests pass

---

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] JSON import attribute required for Playwright ESM**
- **Found during:** Task 2 (Playwright run failed with `TypeError: Module needs an import attribute of "type: json"`)
- **Issue:** Node 22 ESM spec requires `with { type: 'json' }` for JSON module imports in `.ts` Playwright tests
- **Fix:** Added `with { type: 'json' }` to the `conflicting-speech.json` import in `wasm-render.spec.ts`
- **Files modified:** `web/tests/wasm-render.spec.ts`
- **Commit:** `08d7e36`

### Intentional Deviations

**1. export/+page.svelte shows static YAML example**
- **Context:** The plan specifies "resolved speech YAML preview" — the debate store integration that would pass real resolved-speech data to the export page was not planned for 05-04 (it depends on the debate → export navigation flow, which is a 05-05/05-06 concern).
- **Decision:** Static example YAML displayed as preview; `downloadYaml()` uses the example. Full wiring with live `resolvedSpeech` from the debate store is the logical next step in 05-06 (the Forum-offline gate plan).
- **Impact:** A6/BRND-01 copy is correct; stage names are correct; structure is complete. Not a stub that blocks this plan's goals.

---

## Known Stubs

| File | Stub Description | Resolving Plan |
|------|------------------|----------------|
| `web/src/routes/export/+page.svelte` | Shows static example YAML; full debate-store wiring pending | Phase 5 Plan 05 or 06 |
| `handleApplyPatch()` in debate/+page.svelte | Triggers re-resolve only; full patch opinion injection requires Forum/registry integration | Phase 5 Plan 05 |

---

## Threat Surface Scan

No new threat surface beyond the plan's threat model:

| Flag | File | Description |
|------|------|-------------|
| T-05-10 | web/src/lib/conflict.ts | VERIFIED: mapExplanation only maps output; never decides; A9 text verbatim |
| T-05-11 | web/src/routes/debate/+page.svelte | VERIFIED: 150ms debounce in scheduleResolve(); WASM call is synchronous |
| T-05-12 | web/src/lib/components/ConflictOverlay.svelte | VERIFIED: Triple encoding (color+icon+text); amber for warn states; A1 e2e confirms |

---

## Self-Check: PASSED

Files verified:
- `web/src/lib/conflict.ts` — FOUND
- `web/src/lib/conflict.test.ts` — FOUND
- `web/src/lib/stores/speech.ts` — FOUND
- `web/src/lib/components/WasmLoadGate.svelte` — FOUND
- `web/src/lib/components/FoundationBar.svelte` — FOUND
- `web/src/lib/components/ConflictBadge.svelte` — FOUND
- `web/src/lib/components/ConflictOverlay.svelte` — FOUND
- `web/src/lib/components/ExplanationCard.svelte` — FOUND
- `web/src/lib/components/ResolutionPanel.svelte` — FOUND
- `web/src/lib/components/PaneCard.svelte` — FOUND
- `web/src/lib/components/DebateStage.svelte` — FOUND
- `web/src/routes/export/+page.svelte` — FOUND
- `web/tests/wasm-render.spec.ts` — FOUND
- `web/tests/a11y.spec.ts` — FOUND
- `web/tests/fixtures/conflicting-speech.json` — FOUND

Commits verified:
- `35124f6` (TDD RED) — FOUND
- `313f5d7` (Task 1 GREEN) — FOUND
- `08d7e36` (Task 2 GREEN) — FOUND
