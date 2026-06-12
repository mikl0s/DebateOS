# Decisions Intel

Synthesized from classified ADR sources. All entries below are LOCKED (source ADR declares them fixed for the autonomous v1.0 run; do not re-open or re-litigate). Echoes of these decisions in `docs/00-START-HERE.md` and `docs/03-architecture.md` are non-authoritative restatements; this register traces to the ADR only.

---

## D1 — Product name is DebateOS
- source: docs/09-decisions.md
- status: locked
- decision: Product name is `DebateOS`. "Speech" is a domain term (a user's composition), not the product name. The early `Speech OS` title was wrong.
- scope: product naming

## D2 — v1.0 = Phases 0–5
- source: docs/09-decisions.md
- status: locked
- decision: v1.0 spans Phases 0 through 5. Phase 6 (hardware-scanning installer) is post-v1.0 and must not be built.
- scope: v1.0 scope

## D3 — Licensing: AGPLv3 code, CC0 schemas/content
- source: docs/09-decisions.md
- status: locked
- decision: AGPLv3 for all code (resolver, CLI, translators, web, Forum); CC0 for schemas and opinion/point/speech content. `LICENSE` (AGPL-3.0) at root; `schemas/LICENSE` + `examples/LICENSE` (CC0-1.0).
- scope: licensing

## D4 — Monorepo
- source: docs/09-decisions.md
- status: locked
- decision: Single monorepo per `docs/11-repo-layout.md`. Community point repos live separately later.
- scope: repository layout

## D5 — Resolver in Go, compiled both native and WASM
- source: docs/09-decisions.md
- status: locked
- decision: One Go resolver library compiled to native (CLI/service) and WASM (browser); identical results in both targets.
- scope: resolver implementation

## D6 — Resolver is rule-based, NOT a SAT/constraint solver
- source: docs/09-decisions.md
- status: locked
- decision: Topological sort for ordering + direct conflict detection + the docs/04 resolution hierarchy + patch lookup. SAT is explicitly out of scope (preserves human-readability invariant; prior attempt confirmed SAT unnecessary for MVP).
- scope: resolver algorithm

## D7 — CLI in Go
- source: docs/09-decisions.md
- status: locked
- decision: Go CLI wraps the native resolver; manages the private pane in `$HOME`; can serve the embedded Debate UI on localhost.
- scope: CLI

## D8 — Translators in shell/Python, per foundation
- source: docs/09-decisions.md
- status: locked
- decision: Arch translator wraps `mkarchiso`; Debian wraps `live-build`/preseed. Go orchestrates; translators emit.
- scope: translators

## D9 — v1.0 foundations: Arch (+ structure for 1–2 Arch variants) and Debian
- source: docs/09-decisions.md
- status: locked
- decision: Phase 2 (Arch) + Phase 4 (Debian). Two foundations prove the abstraction. (Other foundations, e.g. Fedora, are post-v1.0 / community-owned.)
- scope: supported foundations

## D10 — Debate UI: SvelteKit + adapter-static + Tailwind
- source: docs/09-decisions.md
- status: locked
- decision: Static output; runs the Go-WASM resolver client-side; delivered BOTH via GitHub Pages AND embedded in the CLI.
- scope: web UI

## D11 — Builds: local Docker image + GitHub Actions reusable workflow (same image), deterministic
- source: docs/09-decisions.md
- status: locked
- decision: One Docker build image used by both channels; deterministic via `SOURCE_DATE_EPOCH`.
- scope: build channels

## D12 — Registry: plain YAML in GitHub repos + static index on GitHub Pages
- source: docs/09-decisions.md
- status: locked
- decision: GitHub is the v1.0 bootstrap target; GitLab parity desired, not required.
- scope: registry

## D13 — The Forum: Go (chi) + embedded SQLite, owner-hosted VM, optional
- source: docs/09-decisions.md
- status: locked
- decision: Read-mostly index over GitHub-hosted content. GitHub OAuth only (no passwords/email/2FA). No untrusted code execution. No secrets at rest. DB is a rebuildable cache/index.
- scope: Forum service

## D13a — Forum storage: thin `store` interface + sqlc, SQLite default, Postgres optional
- source: docs/09-decisions.md
- status: locked
- decision: Pure-Go `modernc.org/sqlite` (libSQL-compatible) for dev and v1.0 production; Postgres as optional drop-in backend behind the same interface; no heavyweight ORM; search via SQLite FTS5 in v1.0, abstracted for later Postgres `tsvector`.
- scope: Forum storage layer

## D14 — The Forum is OPTIONAL and ADDITIVE
- source: docs/09-decisions.md
- status: locked
- decision: The compose→resolve→build path must work fully without The Forum. It is the only permitted central service and only for discovery.
- scope: decentralization

## D15 — Forum hosting: Oracle Cloud Always Free (Ampere A1, ARM), EU/APAC region; owner server fallback
- source: docs/09-decisions.md
- status: locked
- decision: Single static `linux/arm64` Go binary + one SQLite file; no separate DB server. Free-tier reclaim/sleep tolerated by rebuildable design.
- scope: Forum hosting

## D16 — Secrets / private pane: never in shared artifacts; first-boot injection
- source: docs/09-decisions.md
- status: locked
- decision: Private pane stays local (optionally synced to the user's own private Git). Key-management details finalized in Phase 3; the invariant is fixed now.
- scope: secrets model

## D17 — Schema derived from Phase 0 Omarchy research, drafted in Phase 1
- source: docs/09-decisions.md
- status: locked
- decision: Do not finalize the schema before Phase 0. Empirical floor beats theoretical schema.
- scope: schema process

## D18 — Prior-attempt carry-forward is scoped
- source: docs/09-decisions.md
- status: locked
- decision: Reuse build/resolver/UI patterns (see docs/10); DROP the entire central backend (auth/email/Postgres/sessions/avatar). The Forum is a new, lean service, not a port of the old backend.
- scope: prior-art reuse

## D19 — Development methodology: TDD with very high coverage
- source: docs/09-decisions.md (owner directive added during ingest, 2026-06-12)
- status: locked
- decision: Tests are written before implementation across all code (resolver, CLI, Forum, translators where testable, UI logic). The resolver targets near-total coverage — every resolution rule, conflict scenario, and edge case (including the Phase 0 variant-study corpus) exists as a test before the behavior is implemented. Deterministic builds and WASM/native parity are verified by automated tests, not inspection.
- scope: development methodology

## D20 — Phase 0 includes Arch-variant substitution study (CachyOS, Garuda)
- source: docs/09-decisions.md (owner directive added during ingest, 2026-06-12)
- status: locked
- decision: Alongside the Omarchy deep-dive, Phase 0 runs a targeted delta study of CachyOS and Garuda Linux: validate same-base foundation swaps (e.g. Omarchy speech on CachyOS), inform the Arch translator's variant-profile/multi-repo/keyring/kernel handling without bloat, and harvest real-world resolver edge cases (foundation-default vs opinion collisions, repo-priority conflicts) as Phase 1 test scenarios. Adds deliverables research/arch-variants-delta.md and research/resolver-edge-cases.md. Same-base substitution is a Phase 2 stretch validation, not a gate.
- scope: Phase 0 research scope / translator design input

---

## Invariants (locked, respect at every step)
- source: docs/09-decisions.md

1. Opinions are OS-agnostic intent; translators own all distro mechanics. No Arch/Debian specifics may leak into opinions, points, speeches, or the schema.
2. Conflict resolution: required > nice-to-have; required-vs-required is a hard conflict unless a patch opinion exists; patches are first-class.
3. Everything must stay human-readable. YAML must remain comprehensible on its own; prefer explainable over clever; never decay into a byzantine solver.
4. No central service in the critical build path. Registry = GitHub + Pages; builds = user CI + local Docker. The Forum is the only central service, optional, executes no untrusted code, holds no secrets.
5. Zero cost & non-commercial. Free public infrastructure + user-owned compute. No monetization. No required paid dependency.
6. North star: Omarchy must be reproducible as a speech on vanilla Arch (Phase 2 milestone).
7. Privacy by construction: the private pane never leaves the user's machine/repo; public sharing includes only public panes; no secrets in shared artifacts.

---

## Process notes (locked operating directives)
- source: docs/09-decisions.md

- The owner's earlier `START_PROMPT` preference for iterative, paused, non-autonomous work is superseded: this is a fully autonomous run to v1.0. Do not pause for owner review between phases unless hitting a true blocker (irresolvable ambiguity, unobtainable external dependency, or destructive/irreversible action).
- New forks not covered by these decisions: choose the option most consistent with the locked decisions and invariants, record it in planning notes, and continue — do not pause to ask.
- Documents in Markdown only. Direct, concise communication in status output.
