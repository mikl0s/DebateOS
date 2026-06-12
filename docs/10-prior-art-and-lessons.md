# 10 — Prior Art & Lessons (from the abandoned attempt)

A previous attempt at this same idea was built and then paused (not for technical reasons — a strategic restart). It reached ~76% of its roadmap: a working **visual Linux distro builder** using the same debate/opinion metaphors. This document distills what to **reuse**, what to **drop**, and what to **avoid**. It is self-contained — the old code is not part of this repo.

> **Important framing:** the old and new projects are the **same domain** (visual Linux distro composition with a debate metaphor). The new vision is a cleaner restart: Go-first resolver, OS-agnostic opinions + per-OS translators, and zero-cost distributed infra — replacing the old central Python/SvelteKit app.

## Old stack (for reference)

- Backend: Python 3.12 + FastAPI + PostgreSQL 18 (SQLAlchemy 2.0 async / asyncpg) + Alembic + ARQ (Redis-backed queue, chosen over Celery) + DragonflyDB.
- Auth: JWT + Argon2 + OAuth2 (GitHub, Google) + TOTP 2FA + email verification.
- Frontend: SvelteKit + TypeScript + Vite + TailwindCSS 4 (Svelte 5 runes; no component library; native fetch; SSE for build progress).
- Infra: Caddy reverse proxy, Docker Compose, `archiso` ISO builder, Docker container build isolation, `SOURCE_DATE_EPOCH` determinism, GitHub Actions overlay validation.

## REUSE — validated patterns & ideas

1. **Layer-based composition with enforced precedence + merge strategies** worked well. Maps directly onto Opinion → Point → Speech layering.
2. **Conflict detection + topological sort (stdlib graph), NOT a SAT solver.** The old project deferred SAT correctly and shipped value with simple, explainable resolution. → Reimplement equivalently in Go (matches D6).
3. **`archiso`/`mkarchiso` wrapping** for unattended Arch ISO output — the Arch translator follows this.
4. **Docker container build isolation** (they pivoted from `systemd-nspawn` to Docker for portability — correct). → Build path 2.
5. **`SOURCE_DATE_EPOCH` determinism** keyed off a config hash → reproducible builds + caching/dedup. → D11.
6. **YAML manifests with semantic-version validation** for overlays/dependencies. → Opinion/Point schema versioning.
7. **Svelte stores + component patterns** (e.g. a `LayerStack`, an `Objection`/conflict modal, build-progress UI). → Lift the *patterns* into the new SvelteKit Debate UI.
8. **Debate-themed UX naming** (build stages like "Preparing the Floor", "Gathering Arguments") — on-brand and worth keeping where it fits.
9. **API/data versioning from day one** (`/api/v1`) — apply to the Forum's API and the static index format.

## DROP — does not fit the new architecture

The new "no central service in the critical path" invariant **eliminates the old central backend**. Do **not** port these:

- The full auth stack: passwords/Argon2, email verification, password reset, **TOTP 2FA**, avatar upload, multi-provider OAuth. → Replaced by **GitHub OAuth only** in The Forum (D13).
- Central user accounts as system-of-record, server-side speech storage as primary, the ARQ build queue, DragonflyDB/Redis, server-side ISO building. → Builds move to user CI/Docker; content lives in Git.
- Caddy/multi-service Compose topology for a central app. → The Forum is a single Go binary + an embedded SQLite file.

The Forum reuses the *design* of the old community/ratings/speeches/subscriptions endpoints, reimplemented lean in Go, read-mostly, over GitHub-indexed data.

## AVOID — anti-patterns that bit the old project

1. **Don't defer major architectural/UX features as late "polish."** The old project pushed its core visual differentiator (3D viz) and advanced resolution to "Phase 7–8, TBD" with no upfront research, creating risk they'd never land. → For DebateOS, the Visual Debate UI is adoption-critical: treat it as a first-class Phase 5 deliverable, and research anything algorithmically heavy up front (Phase 0), not after.
2. **Don't leave terminology unresolved.** The old project carried a "Platform vs Podium" naming confusion from early phases to late ones. → DebateOS terminology is locked in `02`; use it consistently from Phase 1.
3. **Don't leave testing shortcuts in code.** The old project had auth-gate bypass `TODO`s left in the build flow when it paused. → No bypasses or stubs left enabled across phase boundaries; clean or integrate before moving on.
4. **Nail security/determinism early, not as retrofit.** The old project flagged unsandboxed execution as its #1 pitfall yet pivoted isolation mid-project. → Build isolation (Docker), determinism (`SOURCE_DATE_EPOCH`), and the Forum's security-by-design posture are fixed now (`05`, `09`).
5. **Keep "should-have" out of "must-have" per phase.** Scope each phase's MVP explicitly (the roadmap `07` does this) and resist creep during the autonomous run.
