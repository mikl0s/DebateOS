# DebateOS

## What This Is

DebateOS is a decentralized, zero-cost system for composing Linux installations as "speeches" — YAML compositions of OS-agnostic "opinions" curated into "points" — that resolve conflicts visually (the Debate) and compile into bootable, fully-unattended installer ISOs for interchangeable foundations (Arch, Debian). Curators publish points; users compose speeches; translators effectuate intent per foundation. Popular speeches become de-facto distributions maintained purely as configuration.

## Core Value

A user composes a speech from multiple curators' points, resolves conflicts explainably, and produces a bootable unattended installer for their chosen foundation — entirely on free public tooling and user-owned compute, with no central service in the critical path.

## Requirements

### Validated

(None yet — ship to validate)

### Active

See `.planning/REQUIREMENTS.md` for the full checkable list (34 v1 requirements). Headlines:

- [ ] Phase 0 research deliverables (Omarchy deep-dive + CachyOS/Garuda variant study) gate all schema/code work
- [ ] Opinion/Point/Speech schemas derived from Phase 0 evidence; human-readable YAML throughout
- [ ] Rule-based Go resolver (native + WASM, identical results) implementing the docs/04 hierarchy with explanations
- [ ] NORTH STAR: Omarchy reproducible as a speech on vanilla Arch (Phase 2)
- [ ] Go CLI + two deterministic build channels (local Docker, user-owned GitHub Actions) at zero cost
- [ ] Dual-foundation proof: same resolved speech builds Arch AND Debian installers (Phase 4)
- [ ] Git-backed registry (authoritative) + optional Forum (discovery only) + visual Debate UI (Pages + CLI-embedded)

### Out of Scope

- Monetization, paid tiers, central SaaS dependency — invariant 5: zero cost & non-commercial
- Post-install reconciliation (applying speech changes to a running system) — v1.0 is install-time only
- Hardware-scanning installer — Phase 6, post-v1.0; v1.0 = declared hardware + basic install-time resolution, ISO/USB output only
- Fedora/other translators, direct-to-disk install, full GitLab parity — post-v1.0 / community-owned
- Porting the prior attempt's central backend (auth stack, Postgres, build queue, server-side ISO builds) — D18: eliminated by the no-central-services invariant
- SAT/constraint solver — D6: rule-based resolution only, preserves explainability

## Context

- Complete founding context lives in `docs/` (01–11); `docs/09-decisions.md` is the sole decision authority. Ingest intel synthesized in `.planning/intel/`.
- A prior attempt (~76% of roadmap, strategic restart) validated: layered composition with precedence, toposort + direct conflict detection (no SAT), mkarchiso wrapping, Docker build isolation, SOURCE_DATE_EPOCH determinism, YAML+semver manifests, Svelte debate-UI patterns, /api/v1 versioning. Old code is not in this repo; reuse patterns, not code.
- Anti-patterns to avoid (docs/10): deferring the Debate UI as "polish", unresolved terminology, lingering test bypasses, retrofitting security/determinism, scope creep per phase.
- Build host: Linux VM with sudo + Docker; privileged ISO builds (mkarchiso, live-build, loop devices) are expected and not blockers. Privileged actions confined to build tooling/CI containers; untrusted payloads only in isolated containers; no host secrets in artifacts.
- Operating mandate: fully autonomous run to v1.0. Do not pause between phases except for true blockers. At uncovered forks, pick the option most consistent with locked decisions, record it, continue.
- Brand: playful rhetoric metaphor (opinions, points, speeches, debates; "That's just your opinion, man") applied across UI/docs, softened only where it obscures meaning.

## Constraints

- **Methodology**: TDD with very high coverage (D19) — test scenarios specified before implementation in every phase plan; resolver near-total coverage; determinism and WASM/native parity verified by automated tests
- **Gating**: Phase 0 gates everything — no schema drafting before the six research deliverables exist (D17, D20)
- **Tech stack**: Go resolver (native + WASM), Go CLI, shell/Python translators (mkarchiso, live-build), SvelteKit + adapter-static + Tailwind UI, Docker + GitHub Actions builds, Go chi + SQLite (sqlc, swappable store) Forum — all ADR-locked
- **Architecture**: no central service in the critical compose→resolve→build path; Forum optional/additive/rebuildable; registry = GitHub repos + static Pages index
- **Determinism**: identical inputs → identical ISO; SOURCE_DATE_EPOCH from resolved-speech hash; one Docker image for both build channels
- **Privacy**: private pane never leaves the user's machine/repo; secrets never in shared artifacts, injected at first boot
- **Readability**: YAML comprehensible on its own; every resolution explainable; explainable beats clever
- **Repo**: single monorepo per docs/11; AGPLv3 code, CC0 schemas/examples/content
- **Cost**: zero hosting cost; free public infra + user-owned compute; Forum on Oracle Always Free A1 (EU/APAC) or owner server

## Key Decisions

<decisions>

All decisions below are LOCKED by the ADR (`docs/09-decisions.md`). Do not re-open, re-litigate, or pause to ask. At new forks not covered here, choose the option most consistent with these decisions and invariants, record it, and continue.

| ID | Decision | Status |
|----|----------|--------|
| D1 | Product name is DebateOS; "speech" is a domain term, not the product name | LOCKED |
| D2 | v1.0 = Phases 0–5 only; Phase 6 (hardware-scanning installer) is post-v1.0, do not build | LOCKED |
| D3 | AGPLv3 for all code; CC0 for schemas and opinion/point/speech content | LOCKED |
| D4 | Single monorepo per docs/11; community point repos live separately later | LOCKED |
| D5 | One Go resolver library compiled native + WASM; identical results in both targets | LOCKED |
| D6 | Resolver is rule-based (toposort + direct conflict detection + docs/04 hierarchy + patch lookup); SAT explicitly out of scope | LOCKED |
| D7 | CLI in Go; wraps native resolver; manages private pane in $HOME; serves embedded Debate UI on localhost | LOCKED |
| D8 | Translators in shell/Python per foundation: Arch wraps mkarchiso, Debian wraps live-build/preseed; Go orchestrates, translators emit | LOCKED |
| D9 | v1.0 foundations: Arch (+ structure for 1–2 Arch variants) and Debian; others post-v1.0/community | LOCKED |
| D10 | Debate UI: SvelteKit + adapter-static + Tailwind; Go-WASM resolver client-side; delivered via GitHub Pages AND embedded in CLI | LOCKED |
| D11 | Builds: one Docker image used by local Docker AND a reusable GitHub Actions workflow; deterministic via SOURCE_DATE_EPOCH | LOCKED |
| D12 | Registry: plain YAML in GitHub repos + static index on GitHub Pages; GitLab parity desired, not required | LOCKED |
| D13 | Forum: Go (chi) + embedded SQLite on owner-hosted VM; GitHub OAuth only; no untrusted code execution; no secrets at rest; DB is a rebuildable index | LOCKED |
| D13a | Forum storage: thin `store` interface + sqlc; pure-Go `modernc.org/sqlite` default, Postgres optional drop-in; FTS5 search abstracted for later tsvector; no heavyweight ORM | LOCKED |
| D14 | Forum is OPTIONAL and ADDITIVE; compose→resolve→build works fully without it; only permitted central service, discovery only | LOCKED |
| D15 | Forum hosting: Oracle Cloud Always Free Ampere A1 (ARM), EU/APAC region; owner server fallback; single static linux/arm64 binary + one SQLite file | LOCKED |
| D16 | Secrets/private pane: never in shared artifacts; first-boot injection; key-management details finalized in Phase 3, invariant fixed now | LOCKED |
| D17 | Schema derived from Phase 0 Omarchy research, drafted in Phase 1; empirical floor beats theoretical schema | LOCKED |
| D18 | Prior-attempt carry-forward scoped: reuse build/resolver/UI patterns; DROP the entire central backend; Forum is a new lean service, not a port | LOCKED |
| D19 | TDD with very high coverage: tests before implementation everywhere; resolver near-total coverage incl. the Phase 0 edge-case corpus; determinism and WASM/native parity verified by automated tests | LOCKED |
| D20 | Phase 0 includes targeted CachyOS/Garuda Arch-variant substitution study; adds research/arch-variants-delta.md and research/resolver-edge-cases.md; same-base substitution is a Phase 2 stretch validation, not a gate | LOCKED |

**Invariants (respect at every step):**

1. Opinions are OS-agnostic intent; translators own all distro mechanics — no Arch/Debian specifics in opinions, points, speeches, or the schema
2. Conflict resolution: required > nice-to-have; required-vs-required is a hard conflict unless a patch opinion exists; patches are first-class
3. Everything stays human-readable; YAML comprehensible on its own; explainable over clever; never a byzantine solver
4. No central service in the critical build path; registry = GitHub + Pages; builds = user CI + local Docker; Forum is the only central service — optional, no untrusted code, no secrets
5. Zero cost & non-commercial: free public infrastructure + user-owned compute; no monetization; no required paid dependency
6. North star: Omarchy reproducible as a speech on vanilla Arch (Phase 2 milestone)
7. Privacy by construction: private pane never leaves the user's machine/repo; public sharing includes only public panes; no secrets in shared artifacts

**Process notes (locked):** fully autonomous run to v1.0; no pausing between phases except true blockers; Markdown-only documents; direct, concise status output.

</decisions>

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Phase structure locked to ADR Phases 0–5 (research → schema/resolver → Arch → CLI/builds → Debian → registry/Forum/UI) | docs/07 sequencing is part of the locked spec; Phase 0 gates all design | — Pending |
| Roadmap phases numbered 0–5 to match ADR/SPEC naming exactly | Avoids a confusing off-by-one between planning artifacts and founding docs | — Pending |
| Omarchy-on-variant retarget (CachyOS/Garuda) recorded as non-gating stretch criterion in Phase 2 | D20 explicitly: stretch validation, not a gate | — Pending |

---
*Last updated: 2026-06-12 after doc ingest + project initialization*
