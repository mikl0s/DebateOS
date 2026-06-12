# Synthesis Summary — DebateOS Doc Ingest

Mode: new (fresh .planning bootstrap) | Generated: 2026-06-12
Classifications consumed: 12/12 from `.planning/intel/classifications/` (all high confidence; no UNKNOWN)

## Doc counts by type

- ADR: 1 — docs/09-decisions.md (locked, precedence 0)
- SPEC: 7 — docs/07-roadmap.md (10), docs/03-architecture.md (11), docs/04-conflict-resolution.md (12), docs/08-omarchy-research.md (13), docs/02-concepts.md (14), docs/11-repo-layout.md (15), docs/05-distribution-and-infra.md (16)
- PRD: 2 — docs/01-vision.md (20), docs/06-social-layer.md (21)
- DOC: 2 — docs/00-START-HERE.md (30), docs/10-prior-art-and-lessons.md (31)

(Total: 1 ADR + 7 SPEC + 2 PRD + 2 DOC = 12.)

## Decisions (locked)

21 locked decisions (D1–D20 + D13a) + 7 invariants + locked process notes, from docs/09-decisions.md. D19 (TDD with very high coverage — tests before implementation everywhere; resolver near-total; determinism and WASM/native parity test-verified) and D20 (Phase 0 Arch-variant substitution study: CachyOS + Garuda deltas, translator variant profiles, resolver edge-case corpus) were added by owner directive during the ingest session (2026-06-12) and mirrored into docs/09 + docs/08.
→ `.planning/intel/decisions.md`

Headline: product = DebateOS; v1.0 = Phases 0–5 only; AGPLv3 code / CC0 schemas+content; monorepo; Go resolver (rule-based, native + WASM); Go CLI; shell/Python translators (Arch mkarchiso, Debian live-build); SvelteKit static Debate UI (Pages + CLI-embedded); deterministic Docker/Actions builds; GitHub YAML registry + Pages index; optional Forum (Go chi + SQLite via sqlc store, GitHub OAuth only, Oracle A1 free-tier hosting); secrets never in shared artifacts (first-boot injection); schema gated by Phase 0 Omarchy research; prior backend dropped.

## Requirements

13 requirements (REQ-compose-build-zero-cost, REQ-omarchy-north-star, REQ-dual-foundation-proof, REQ-curator-ecosystem, REQ-human-readable-yaml, REQ-anti-dogmatic-brand, REQ-forum-search-discovery, REQ-forum-subscriptions, REQ-forum-ratings-reputation, REQ-forum-collab-conflict-resolution, REQ-forum-boundaries, REQ-registry-authoritative, REQ-translator-ownership-model) + 3 explicit v1.0 scope exclusions (no monetization, no post-install reconciliation, no hardware-scanning installer).
→ `.planning/intel/requirements.md`

## Constraints

24 constraints from the 7 SPECs. Type breakdown: schema 4 (terminology contract, opinion metadata floor, monorepo layout, plus opinion schema floor), api-contract 4 (component stack/data flow, registry-on-GitHub, Forum service/storage, module/build-channel contracts), protocol 8 (resolution hierarchy, patch opinions, hardware-aware resolution, ordering/toposort, community conflict workflow, CI build path, Docker build path, phase sequencing + Phase 0 method), nfr 8 (dual compile targets, rule-based-not-SAT, hardware scope, readability, zero-cost path, determinism, Forum security/hosting, private pane/secrets, deferred open questions).
→ `.planning/intel/constraints.md`

## Context

6 topics from the 2 DOCs: founding-context/doc map, autonomous operating mandate, build/dev environment (sudo + Docker available, privileged ISO builds expected), prior-attempt old stack, REUSE list (9 patterns), DROP list, AVOID list (5 anti-patterns).
→ `.planning/intel/context.md`

## Conflicts

0 blockers, 0 competing variants, 5 auto-resolved/informational.
- Cycle detection found two 2-cycles in the cross-ref graph ({04,06} and {09,10}); both are benign see-also references with no contradicting content, neutralized by manifest precedence ordering — recorded as INFO, not gated.
- Three precedence applications logged as INFO (ADR over DOC restatements, ADR over PRD foundation scope, ADR/SPEC scope clarification on install-time hardware detection).
→ `.planning/INGEST-CONFLICTS.md`

## Status

READY — no blockers, no competing variants. Safe for `gsd-roadmapper` to consume this directory.

## Notes for downstream

- docs/09-decisions.md is the sole decision authority; restatements in docs/00 and docs/03 are summaries.
- Phase 0 (Omarchy research + CachyOS/Garuda variant study, per D20) gates the schema and everything after it; do not draft schemas before `research/` deliverables exist (now six files, incl. arch-variants-delta.md and resolver-edge-cases.md).
- D19: every phase plan must be TDD-shaped — test harness/scenarios specified before implementation tasks; resolver coverage near-total; deterministic-build and WASM/native-parity checks are automated tests.
- Phase 6 (hardware-scanning installer) and the post-v1.0 list (Fedora translator, direct-to-disk, GitLab parity, post-install reconciliation) must not enter v1.0 scope.
- The Forum is optional/additive: any roadmap item that puts it in the critical compose→resolve→build path violates locked invariant 4.
