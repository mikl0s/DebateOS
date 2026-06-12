# 03 — Architecture

All technical choices here are **locked** (see `09-decisions.md` for rationale). This document is the authoritative component and data-flow map.

## Component stack

```
┌─────────────────────────────────────────────────────────────────────┐
│  Visual Debate UI  (SvelteKit static + Tailwind)                      │
│   • glass-pane composition, red/green conflict overlays               │
│   • runs the resolver as Go→WASM, fully client-side                   │
│   • delivered TWO ways: GitHub Pages  AND  embedded in the CLI        │
│   • discovery features call The Forum (optional)                      │
├─────────────────────────────────────────────────────────────────────┤
│  CLI  (Go)         debateos compose | validate | build | pane         │
│   • wraps the resolver (native), manages the private pane in $HOME    │
│   • can serve the embedded Debate UI on localhost                     │
├─────────────────────────────────────────────────────────────────────┤
│  Resolver  (Go library)                                               │
│   • parse + validate speeches/points/opinions                         │
│   • dependency graph, conflict detection, rule-based resolution       │
│   • patch-opinion application, hardware-aware checks                   │
│   • compiled BOTH native (CLI/service) and WASM (browser)             │
├─────────────────────────────────────────────────────────────────────┤
│  Translators  (shell/Python, one per foundation)                      │
│   • Arch (wraps mkarchiso) [+ Arch variants]   • Debian (live-build)  │
│   • consume a RESOLVED speech, emit concrete build instructions       │
├─────────────────────────────────────────────────────────────────────┤
│  Build backends                                                       │
│   • local Docker image (resolver + translators + ISO tooling)         │
│   • GitHub Actions reusable workflow (same Docker image)              │
│   • deterministic via SOURCE_DATE_EPOCH                               │
├─────────────────────────────────────────────────────────────────────┤
│  Registry           plain YAML in GitHub repos + static Pages index   │
│  The Forum (opt.)   Go (chi) + SQLite (swappable store), owner VM      │
└─────────────────────────────────────────────────────────────────────┘
```

## Why one resolver, two compile targets

The resolver is the heart of the system and must produce **identical results** whether it runs in the browser (live conflict visualization during a debate) or on the build machine (final validation before translation). Writing it once in Go and compiling to **both native and WASM** guarantees that. The Debate UI never reimplements resolution logic; it calls the WASM build. The CLI and The Forum call the native build.

## Data flow (compose → installed system)

1. **Compose.** User assembles a speech in the Visual Debate UI (or hand-edits YAML). Discovery of points/speeches is served by the static registry index and, if available, The Forum.
2. **Resolve.** The resolver pulls referenced points (from GitHub), builds the dependency/conflict graph, applies the resolution hierarchy (`04`), confirms the chosen translator supports every **required** opinion, and emits a **resolved speech** — fully concrete and ordered. In the UI this runs continuously as WASM, so conflicts surface live.
3. **Translate.** The chosen translator converts the resolved speech into OS-specific build instructions (native package manager calls, config placement, service enablement, kernel params, script payloads).
4. **Build.** A build backend (user's GitHub Actions *or* local Docker) executes the translator output and emits a **bootable, fully-unattended installer image** (ISO for USB; direct-to-disk later). Deterministic inputs → deterministic output (cacheable).
5. **Install.** The installer runs, applies everything with zero questions, and deploys the private pane → a finished system boots.

## Resolver design (rule-based, not SAT)

Conflict resolution is **rule-based**, not a SAT/constraint solver. This is a deliberate invariant: the conflict graph must stay **human-readable** and every resolution must be explainable to an average user. The model:

- Topological sort over ordering constraints to compute install order.
- Direct pairwise conflict detection between opinions/capabilities.
- The resolution hierarchy in `04` decides outcomes (required > nice-to-have; required-vs-required = hard conflict unless a patch exists; patches first-class).
- Hardware-conditional opinions evaluated against declared/scanned hardware.

A full SAT solver is explicitly **out of scope** — the prior attempt confirmed it is unnecessary for the MVP and it would violate the readability invariant.

## Hardware detection

A small program in the installer scans hardware at install time to resolve hardware-conditional opinions (NVIDIA vs AMD drivers, kernel choice). The Debate UI also uses declared/scanned hardware at composition time so hardware conflicts surface during the debate, not at install. (The full hardware-scanning *installer* is Phase 6 / post-v1.0; v1.0 supports declared hardware + basic install-time resolution.)

## Image building

Each foundation has native tooling for custom images — `mkarchiso` (Arch), `live-build`/preseed (Debian). Translators wrap that tooling. Output target for v1.0 is a bootable ISO; the stock distro installer flow is bypassed entirely.

## The Forum (optional discovery service)

A lean Go service (chi router) backed by an **embedded SQLite database** (pure-Go driver, libSQL-compatible) behind a thin swappable `store` interface, hosted on an owner-provided VM. It is **optional and additive**: the compose→build path works fully without it. Security-by-design properties (see `05`/`06`):

- **Read-mostly index over GitHub.** Points/speeches live as YAML in users' GitHub repos; The Forum ingests and indexes them. No arbitrary file uploads.
- **No untrusted code execution.** Builds never run on the VM.
- **No passwords, no secrets.** GitHub OAuth only.
- **Rebuildable.** The DB is a cache/index of public GitHub data + lightweight social state (ratings, subscriptions, conflict threads); a total wipe means re-index, not data loss.

## Open questions deferred *by design* (not blockers)

These are intentionally deferred and must NOT block the autonomous run; pick the simplest option consistent with the invariants when reached:

- Exact registry static-index format and search UX (refine in Phase 5).
- Direct-to-disk install target (post-v1.0; ISO is the v1.0 target).
- Full GitLab parity (GitHub is the v1.0 bootstrap target).
- Post-install reconciliation (explicitly out of v1.0 scope).
