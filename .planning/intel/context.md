# Context Intel

Background notes from DOC sources (`docs/00-START-HERE.md`, `docs/10-prior-art-and-lessons.md`), keyed by topic with source attribution. Where these docs restate decisions, the authoritative versions live in `decisions.md` (sourced from docs/09-decisions.md).

---

## Topic: Founding context & doc map
- source: docs/00-START-HERE.md

The `docs/` folder (01–11, read in order) is the complete and self-contained founding context for DebateOS v1.0 (Phases 0–5). The earlier `NewDocs/` and `OldDocs/` folders were removed on purpose; everything relevant was distilled into these files — do not look for them. Doc map: 01 vision; 02 terminology; 03 architecture; 04 conflict resolution; 05 distribution/infra; 06 social layer; 07 roadmap; 08 Phase 0 research brief; 09 locked decision log + invariants; 10 prior-art lessons; 11 monorepo layout.

## Topic: Operating mandate (autonomous run)
- source: docs/00-START-HERE.md

1. Run autonomously to v1.0 (Phases 0–5); Phase 6 is post-v1.0, do not build it.
2. Decisions in docs/09-decisions.md are LOCKED — do not re-open, re-litigate, or pause to ask. At a genuinely new fork, pick the option most consistent with the locked decisions and invariants, record it, keep going.
3. Respect the docs/09 invariants at every step.
4. Phase 0 gates everything: do the Omarchy research before designing the schema; the schema is derived from real Omarchy data, not invented.
5. North-star validation: Omarchy must be reproducible as a speech on vanilla Arch (Phase 2 milestone).

## Topic: Build & dev environment
- source: docs/00-START-HERE.md

The development/build host is a Linux VM with root via `sudo`; Docker expected available. Phases 2–4 produce real installer ISOs — `mkarchiso`/`archiso` (Arch), `live-build`/preseed (Debian), loop devices, Docker build isolation — all requiring privileged operations. Use `sudo`/Docker as needed; do not treat privileged build steps as blockers. Constraints: keep privileged actions confined to build tooling and CI containers; never run untrusted opinion/translator payloads outside an isolated container; never bake host secrets into build artifacts.

## Topic: Prior attempt — framing & old stack
- source: docs/10-prior-art-and-lessons.md

A previous attempt at the same idea (visual Linux distro builder, debate/opinion metaphors) reached ~76% of its roadmap before a strategic restart (not a technical failure). Old code is not part of this repo. Old stack for reference: Python 3.12 + FastAPI + PostgreSQL 18 (SQLAlchemy 2.0 async) + Alembic + ARQ/Redis + DragonflyDB; JWT/Argon2/OAuth2/TOTP-2FA auth; SvelteKit + TS + Tailwind 4 frontend (Svelte 5 runes, SSE build progress); Caddy + Docker Compose + archiso + SOURCE_DATE_EPOCH determinism.

## Topic: Prior attempt — REUSE (validated patterns)
- source: docs/10-prior-art-and-lessons.md

1. Layer-based composition with enforced precedence + merge strategies → maps onto Opinion → Point → Speech layering.
2. Conflict detection + topological sort (stdlib graph), not SAT → reimplement in Go (matches D6).
3. `archiso`/`mkarchiso` wrapping for unattended Arch ISO output → Arch translator follows this.
4. Docker container build isolation (pivot from systemd-nspawn was correct) → build path 2.
5. `SOURCE_DATE_EPOCH` determinism keyed off a config hash → D11.
6. YAML manifests with semantic-version validation → opinion/point schema versioning.
7. Svelte stores + component patterns (LayerStack, Objection/conflict modal, build-progress UI) → lift the patterns into the new Debate UI.
8. Debate-themed UX naming (build stages like "Preparing the Floor", "Gathering Arguments") — keep where it fits.
9. API/data versioning from day one (`/api/v1`) → apply to the Forum API and the static index format.

## Topic: Prior attempt — DROP (eliminated by new architecture)
- source: docs/10-prior-art-and-lessons.md

The no-central-services-in-critical-path invariant eliminates the old central backend. Do not port: the full auth stack (passwords/Argon2, email verification, password reset, TOTP 2FA, avatar upload, multi-provider OAuth) → replaced by GitHub OAuth only (D13); central user accounts as system-of-record, server-side speech storage as primary, ARQ build queue, DragonflyDB/Redis, server-side ISO building → builds move to user CI/Docker, content lives in Git; Caddy/multi-service Compose topology → The Forum is a single Go binary + embedded SQLite file. The Forum reuses the design of the old community/ratings/speeches/subscriptions endpoints, reimplemented lean in Go, read-mostly, over GitHub-indexed data.

## Topic: Prior attempt — AVOID (anti-patterns)
- source: docs/10-prior-art-and-lessons.md

1. Don't defer major architectural/UX features as late "polish" — the Visual Debate UI is adoption-critical: first-class Phase 5 deliverable; research anything algorithmically heavy up front (Phase 0).
2. Don't leave terminology unresolved — DebateOS terminology is locked in docs/02; use it consistently from Phase 1.
3. Don't leave testing shortcuts in code — no bypasses or stubs left enabled across phase boundaries.
4. Nail security/determinism early, not as retrofit — build isolation (Docker), determinism (SOURCE_DATE_EPOCH), and the Forum's security posture are fixed now.
5. Keep "should-have" out of "must-have" per phase — scope each phase's MVP explicitly (docs/07 does this); resist creep during the autonomous run.
