# 05 — Distribution & Infrastructure

## Core constraint (invariant)

The **critical compose→resolve→build path must run on free public tooling and user-owned compute, with no central bottleneck.** It must be possible to compose a speech and build an installer with nothing but a browser (or the CLI) and either a free CI tier or local Docker. No central service may be *required* for this path.

**The Forum** (the discovery/social service) is the single, deliberate exception — and it is **optional, additive, and rebuildable**. If it is offline, the core path still works.

## Registry: GitHub as backend

- Points and public speeches are plain **YAML files in Git repositories**.
- Versioning, forking, PRs, attribution, and review come free from Git/GitHub.
- A **static registry index** is generated from those repos and hosted on **GitHub Pages**, rebuilt on commit → zero cost.
- GitLab parity is desirable but not required for v1.0; GitHub is the bootstrap target.

## Build path 1: Distributed CI (user's own credits)

The core infrastructure bet: **ISO building runs on the user's own CI, not central infrastructure.**

- The project publishes a reusable **GitHub Actions workflow/action**.
- The user forks a template repo, commits their personal speech YAML; the build triggers in **their** CI on **their** free-tier minutes.
- Workflow: parse speech → resolve against the public registry → run the chosen translator → emit the installer ISO as an artifact.
- Result: completely distributed compute, no central bottleneck, no sponsorship required.
- Privacy bonus: the private pane never leaves the user's repo/CI.
- Requirement: the action must be dead-simple with excellent docs.

## Build path 2: Local Docker (full privacy)

- A published Docker image bundles all build tooling: resolver, translators, ISO builders.
- `docker run debateos:latest` with the speech YAML mounted in → installer ISO out, locally.
- For homelab users, offline builds, no-GitHub paths, and anyone without CI credits.
- **Same image** is used by the GitHub Action internally — one build environment, two delivery channels.

## Determinism

Builds are deterministic: identical inputs → identical output, enabling caching/deduplication. Use `SOURCE_DATE_EPOCH` (derived from the resolved-speech hash) and pinned package snapshots where the foundation allows. (Carried forward from the prior attempt, where this was validated.)

## The Forum: optional discovery service

A lean **Go** service (chi router) on an **owner-hosted VM**, backed by an **embedded SQLite database** (pure-Go `modernc.org/sqlite` driver, libSQL-compatible) behind a thin swappable `store` interface. Purpose: the social/discovery features that a static index can't do well — search, reputation, subscriptions, and the collaborative conflict-resolution workflow (`06`).

**Storage layer:** A `store` repository interface (domain methods) with **sqlc**-generated type-safe queries. Default backend is **SQLite** (single file, no separate DB server, in-memory for tests) for dev and v1.0 production. **Postgres** is an optional drop-in backend behind the same interface for future multi-node scale — no heavyweight ORM, no DB-agnostic plumbing beyond the interface. Full-text search uses **SQLite FTS5** in v1.0, abstracted behind the store so a Postgres `tsvector` implementation can be added later.

**Security-by-design properties (all mandatory):**

1. **Read-mostly index over GitHub.** It ingests and indexes points/speeches that live as YAML in users' GitHub repos. It does **not** accept arbitrary file uploads or host primary content.
2. **No untrusted code execution, ever.** Builds and resolution of *untrusted* input never run on the VM. (The resolver may run server-side only over already-public, already-indexed GitHub content for indexing purposes.)
3. **No passwords, no email, no 2FA.** Identity is **GitHub OAuth only** — delegated entirely to GitHub. The service stores no credentials.
4. **No secrets at rest.** No private-pane data, no SSH keys, no user files.
5. **Rebuildable.** The DB holds a cache/index of public GitHub data plus lightweight social state (ratings, subscriptions, conflict threads, patch-pointers). A total DB loss is recoverable by re-indexing GitHub.
6. **Minimal surface.** A single static Go binary plus one SQLite file — no separate database server to run, secure, and trivial to audit, back up, or rebuild.

**What it stores (SQLite):** indexed point/speech metadata (repo, version, tags, curator, popularity/freshness), subscription edges, ratings, and conflict-resolution threads with pointers to the GitHub PRs/patch opinions that resolve them.

**What it never stores:** private panes, secrets, credentials, or any primary content that isn't already public on GitHub.

### Hosting target (zero-cost)

The Forum is sized to run **free, indefinitely**:

- **Primary: Oracle Cloud Always Free — Ampere A1 (ARM).** Always-free (no time limit): up to 4 OCPU / 24 GB RAM / 200 GB storage / 10 TB egress. The service is just a single static `linux/arm64` Go binary + one SQLite file, so it runs with enormous headroom (a fraction of even 1 OCPU / 6 GB). Caveat: the free A1 shape is capacity-constrained in US regions ("out of host capacity"); provision in an EU/APAC region (Frankfurt/Singapore/Tokyo) to avoid it.
- **Fallback: owner-provided server.** Any tiny always-on host works since there is no separate database server to provision. (If Postgres is ever adopted at scale via the swappable store, free managed options like Neon/Aiven exist — but v1.0 needs none of that.)

Because the service is read-mostly, code-free, and reconstructible from GitHub, free-tier reclaim or sleep is a tolerable failure mode by design — never data loss, never a broken core path.

## Personal speech storage & the private pane

- The user's speech (including the private pane) lives in the **home folder**, managed by the CLI — environment variables, dotfiles, keys, home-folder layout.
- Optionally backed up to the user's **own private Git repo**.
- Public sharing is opt-in and includes **only** public panes.

## Secrets model (private pane)

- Secrets (SSH keys, tokens, anything in the private pane) are **never baked into shared/public artifacts.**
- For the GitHub Actions path, the private pane stays within the user's own repo/CI.
- For the Docker path, it never leaves the machine.
- Secrets are **injected at first boot** on the target machine, not embedded in shared images. (Detailed key-management design is finalized during Phase 3 CLI work; the invariant — *no secrets in shared artifacts, first-boot injection* — is fixed now.)

## Optional future: hosted builds

If the project takes off, a sponsor-funded central build service could offer "compose → download ISO" for people without CI or Docker. Strictly optional, sponsor-funded, never required. The default remains free tiers + user compute. **Out of scope for v1.0.**
