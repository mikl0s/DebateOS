# DebateOS — START HERE (GSD Kickoff Context)

> *"That's just your opinion, man."*

**Read this first, then read `docs/` 01 → 11 in order.** This `docs/` folder is the **complete and self-contained** founding context for DebateOS. It is everything you need to plan and build **v1.0 (Phases 0–5)** autonomously. The earlier `NewDocs/` and `OldDocs/` folders have been removed on purpose — do not look for them; everything relevant from both has been distilled into these files.

## What you are building

DebateOS is a **declarative configuration abstraction layer for Linux** that decouples the base distribution from personalization and stack choices. Post-install decisions become first-class, composable objects:

- **Opinion** — one atomic, OS-agnostic configuration decision.
- **Point** — a curated bundle of opinions, maintained by a person/org.
- **Speech** — a user's full composition of points + private customization.
- **Debate** — the visual conflict-resolution process that turns a composition into a coherent speech.
- **Translator** — the per-OS layer that turns an abstract speech into a concrete, fully-unattended installer (Arch, Debian, …).

Full definitions in `02-concepts.md`. The product name is **DebateOS**; "Speech" is a *domain term* (a user's composition), not the product name.

## Operating mandate for this autonomous run

1. **Run autonomously to v1.0.** v1.0 = Phases 0 through 5 (see `07-roadmap.md`). Phase 6 (hardware-scanning installer) is explicitly **post-v1.0** — do not build it.
2. **The decisions in `09-decisions.md` are LOCKED.** They were made deliberately during a brainstorming session with the owner. Do **not** re-open, re-litigate, or pause to ask about them. If you hit a genuinely new fork not covered there, pick the option most consistent with the locked decisions and the invariants below, record it in your planning notes, and keep going.
3. **Respect the invariants** in `09-decisions.md` §Invariants at every step.
4. **Phase 0 gates everything.** Do the Omarchy research (`08-omarchy-research.md`) before designing the schema. The schema is derived from real Omarchy data, not invented.
5. **North-star validation:** *Omarchy must be reproducible as a speech on vanilla Arch* (the Phase 2 milestone).

## Build order (high level)

| Phase | Delivers | Gates |
|---|---|---|
| 0 | Omarchy decomposition → opinion inventory + schema requirements | everything |
| 1 | YAML schema + Go resolver core (rule-based) + conflict test harness | translators |
| 2 | Arch translator (`mkarchiso`) → **reproduce Omarchy** | proof of concept |
| 3 | Go CLI + Docker build image + GitHub Actions build workflow | user-facing builds |
| 4 | Debian translator (`live-build`/preseed) | proves abstraction |
| 5 | Static registry index + **The Forum** discovery service (Go + embedded SQLite, optional) + Visual Debate UI (SvelteKit static + Go→WASM resolver) | adoption |

## Tech at a glance (all locked — see `09-decisions.md`)

- **Resolver + CLI:** Go. Resolver is a library, compiled **both** native (CLI) and to **WASM** (browser).
- **Translators:** shell/Python per distro.
- **Debate UI:** SvelteKit + `adapter-static` + Tailwind; runs the Go-WASM resolver in-browser; served **both** from GitHub Pages and embedded in the CLI (`debateos compose`).
- **The Forum (optional discovery service):** Go (chi) + embedded SQLite (swappable `store`; Postgres optional later) on an owner-hosted VM. GitHub OAuth only. Read-mostly index over GitHub-hosted content. No untrusted code execution. DB is a single rebuildable file.
- **Builds:** local Docker + GitHub Actions reusable workflow. Deterministic via `SOURCE_DATE_EPOCH`.
- **Registry:** plain YAML in GitHub repos + a static index on GitHub Pages.
- **License:** AGPLv3 for code; CC0 for schemas + opinion/point/speech content.
- **Layout:** monorepo (`11-repo-layout.md`).

## Build & dev environment

The development/build host is a Linux VM with **root via `sudo`** (and Docker expected to be available). Phases 2–4 produce **real installer ISOs** — `mkarchiso`/`archiso` (Arch), `live-build`/preseed (Debian), loop devices, and Docker-based build isolation — all of which require privileged operations. **Use `sudo`/Docker as needed; do not treat privileged build steps as blockers.** Constraints: keep privileged actions confined to build tooling and CI containers; never run untrusted opinion/translator payloads outside an isolated container; and never bake host secrets into build artifacts (see `05`, invariant 7).

## Document map

| File | Contents |
|---|---|
| `01-vision.md` | Mission, philosophy, the problem |
| `02-concepts.md` | Core terminology, precisely defined |
| `03-architecture.md` | Components, data flow, all technical decisions |
| `04-conflict-resolution.md` | Metadata model, resolution rules, patches |
| `05-distribution-and-infra.md` | Zero-cost build/registry + The Forum + secrets model |
| `06-social-layer.md` | Curators, discovery, reputation, collaborative conflict resolution |
| `07-roadmap.md` | Phases 0–5 with goals, deliverables, success criteria |
| `08-omarchy-research.md` | Phase 0 research brief (gates the schema) |
| `09-decisions.md` | **The locked decision log + invariants** |
| `10-prior-art-and-lessons.md` | What to reuse / drop / avoid from the prior attempt |
| `11-repo-layout.md` | Concrete monorepo directory tree to create |
