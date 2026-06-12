# 09 — Locked Decisions & Invariants

**This is the "do not re-ask" document.** Every decision below was made deliberately with the owner during a brainstorming session (2026-06-12). During the autonomous run, treat these as fixed. If you reach a genuinely new fork not covered here, choose the option most consistent with these decisions and the invariants, record it in your planning notes, and continue — do not pause to ask.

## Locked decisions

| # | Decision | Rationale |
|---|----------|-----------|
| D1 | **Product name is `DebateOS`.** "Speech" is a domain term (a user's composition), not the product name. | Owner correction. (The early `Speech OS` title was wrong.) |
| D2 | **v1.0 = Phases 0–5.** Phase 6 (hardware-scanning installer) is post-v1.0 and must not be built. | Owner scope decision. |
| D3 | **License: AGPLv3** for all code (resolver, CLI, translators, web, Forum); **CC0** for schemas and opinion/point/speech content. | Network-copyleft protects against closed hosted forks of the service; CC0 keeps the opinion ecosystem maximally remixable. Add `LICENSE` (AGPL-3.0) at root and `schemas/LICENSE` + `examples/LICENSE` (CC0-1.0). |
| D4 | **Monorepo.** See `11-repo-layout.md`. | Simplest to coordinate an autonomous v1.0; community point repos live separately later. |
| D5 | **Resolver in Go, compiled BOTH native and to WASM.** One resolver, identical results in CLI/service (native) and browser (WASM). | The resolver is the heart; duplicate logic would drift. |
| D6 | **Resolver is rule-based, NOT a SAT/constraint solver.** Topological sort for ordering + direct conflict detection + the `04` hierarchy + patch lookup. | Preserves the human-readability invariant; the prior attempt confirmed SAT is unnecessary for MVP. |
| D7 | **CLI in Go.** Wraps the native resolver; manages the private pane in `$HOME`; can serve the embedded Debate UI on localhost. | Single-binary, cross-platform, matches resolver language. |
| D8 | **Translators in shell/Python, per foundation.** Arch wraps `mkarchiso`; Debian wraps `live-build`/preseed. | Each distro's tooling differs; Go orchestrates, translators emit. |
| D9 | **Foundations for v1.0: Arch (+ structure for 1–2 Arch variants) and Debian.** | Phase 2 + Phase 4. Two foundations prove the abstraction. |
| D10 | **Debate UI: SvelteKit + `adapter-static` + Tailwind.** Static output. Runs the Go-WASM resolver client-side. Delivered BOTH via GitHub Pages AND embedded in the CLI. | Static output satisfies both delivery modes; validated in this domain by the prior attempt; interops with Go-WASM. |
| D11 | **Builds: local Docker image + GitHub Actions reusable workflow (same image). Deterministic via `SOURCE_DATE_EPOCH`.** | Zero-cost, distributed, reproducible. |
| D12 | **Registry: plain YAML in GitHub repos + static index on GitHub Pages.** GitHub is the v1.0 bootstrap target; GitLab parity desired, not required. | Zero hosting cost; Git provides versioning/forking/PRs free. |
| D13 | **The Forum (optional discovery service): Go (chi) + embedded SQLite, owner-hosted VM.** Read-mostly index over GitHub-hosted content. **GitHub OAuth only** (no passwords/email/2FA). **No untrusted code execution.** **No secrets at rest.** DB is a rebuildable cache/index. | Enables the social layer while staying secure-by-design and disposable. See `05`/`06`. |
| D13a | **Forum storage: a thin `store` repository interface + sqlc-generated queries.** Default backend **SQLite** (pure-Go `modernc.org/sqlite`, libSQL-compatible) for dev and v1.0 production; **Postgres** is an optional drop-in backend behind the same interface for future scale. No heavyweight ORM. Search via **SQLite FTS5** in v1.0, abstracted for a later Postgres `tsvector` impl. | Single static binary + one DB file = leanest possible deploy; swappable without over-engineering; in-memory SQLite for tests. Licensing is a non-factor: SQLite (public domain), libSQL (MIT), and `modernc.org/sqlite` (BSD-3) are all AGPL-compatible — the choice is operational (pure-Go, cgo-free, static arm64 builds), not legal. |
| D14 | **The Forum is OPTIONAL and ADDITIVE.** The compose→resolve→build path must work fully without it. | Preserves the decentralization thesis; The Forum is the only permitted central service and only for discovery. |
| D15 | **The Forum hosting target: Oracle Cloud Always Free (Ampere A1, ARM), EU/APAC region; owner's own server as fallback.** Free indefinitely. Deploy is a single static `linux/arm64` Go binary + one SQLite file — no separate DB server. | Keeps the project at $0; rebuildable design tolerates free-tier reclaim/sleep. |
| D16 | **Secrets / private pane: never baked into shared artifacts; first-boot injection on the target machine; private pane stays local (optionally synced to the user's own private Git).** | Security-by-construction; safe public sharing. Key-management details finalized in Phase 3. |
| D17 | **Schema is derived from Phase 0 Omarchy research, then drafted in Phase 1.** Do not finalize the schema before Phase 0. | Empirical floor beats theoretical schema. |
| D18 | **Prior-attempt carry-forward is scoped** (see `10`): reuse build/resolver/UI *patterns*; DROP the entire central backend (auth/email/Postgres/sessions/avatar) — the no-central-services-in-critical-path invariant eliminates it. The Forum is a *new, lean* service, not a port of the old backend. | The old backend assumed a central app; the new architecture does not. |

## Invariants (respect at every step)

1. **Opinions are OS-agnostic intent; translators own all distro mechanics.** No Arch/Debian specifics may leak into opinions, points, speeches, or the schema.
2. **Conflict resolution:** required > nice-to-have; required-vs-required is a hard conflict unless a patch opinion exists; **patches are first-class.**
3. **Everything must stay human-readable.** The visual Debate UI is the eventual primary interface, but the YAML must remain comprehensible on its own. Never let resolution decay into a byzantine solver — prefer *explainable* over *clever*.
4. **No central service in the critical build path.** Registry = GitHub + Pages; builds = user CI + local Docker. The Forum is the only central service, it is optional, and it executes no untrusted code and holds no secrets.
5. **Zero cost & non-commercial.** Free public infrastructure + user-owned compute. No monetization. No required paid dependency.
6. **North star:** *Omarchy must be reproducible as a speech on vanilla Arch* (Phase 2 milestone).
7. **Privacy by construction:** the private pane never leaves the user's machine/repo; public sharing includes only public panes; no secrets in shared artifacts.

## Process notes

- The owner's earlier `START_PROMPT` preference for *iterative, paused, non-autonomous* work is **superseded** for this run: the owner has explicitly chosen a **fully autonomous run to v1.0**. Do not pause for owner review between phases unless you hit a true blocker (ambiguity not resolvable from these docs + invariants, an external dependency that cannot be obtained, or a destructive/irreversible action).
- Documents in Markdown only. Direct, concise communication in any status output.
