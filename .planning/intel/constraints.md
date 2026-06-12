# Constraints Intel

Extracted from SPEC sources (`docs/02`, `docs/03`, `docs/04`, `docs/05`, `docs/07`, `docs/08`, `docs/11`). Manifest precedence among SPECs: 07 (10) > 03 (11) > 04 (12) > 08 (13) > 02 (14) > 11 (15) > 05 (16). No contradictions among SPECs or against the ADR were found; precedence was not needed to discard any content.

---

## Terminology contract
- source: docs/02-concepts.md
- type: schema
- content: Load-bearing terms, used consistently across code, schemas, docs, UI:
  - **Opinion** — atomic, OS-agnostic configuration decision (intent only, never distro mechanics). May encompass package install/removal, script payloads, config/dotfile deployment, service enablement, kernel/boot params, hardware-conditional logic, theming, keybindings. Carries metadata: dependencies, conflicts, hardware conditions, ordering constraints, known patches, required-vs-nice-to-have, translator capability requirements.
  - **Point** — curated, coherent bundle of opinions; each opinion marked required or nice-to-have; versioned, forkable, subscribable; foundation-agnostic.
  - **Speech** — user's complete composition (points + individually selected opinions + private customization), expressed as YAML; public panes + ONE private pane that never leaves the user's machine; compiles into a fully-unattended installer.
  - **Debate** — the composition and conflict-resolution process (visual: foundation + glass panes, red/green overlaps); compresses into the final speech.
  - **Translator** — OS-specific effectuation layer, one per foundation; declares supported opinions/capabilities; unsupported opinions break visibly at composition time, never silently at install time.
  - **Foundation** — base OS + bootloader + installer; deliberately boring, interchangeable.
  - **Registry** — public Git-backed index of points and public speeches (GitHub repos + static Pages index).
  - **Patch Opinion** — first-class community opinion making two conflicting opinions coexist; discoverable, versioned, attributable.
  - **Forum** — optional discovery service on top of the registry.

## Component stack & data flow
- source: docs/03-architecture.md
- type: api-contract
- content: Layers: Visual Debate UI (SvelteKit static + Tailwind, Go-WASM resolver client-side, delivered via GitHub Pages AND embedded in CLI) → CLI (Go: `debateos compose | validate | build | pane`; serves embedded UI on localhost) → Resolver (Go library: parse/validate, dependency graph, conflict detection, rule-based resolution, patch application, hardware-aware checks; compiled native + WASM) → Translators (shell/Python; Arch wraps mkarchiso, Debian wraps live-build; consume a RESOLVED speech, emit concrete build instructions) → Build backends (local Docker image; GitHub Actions reusable workflow using the same image; deterministic via SOURCE_DATE_EPOCH). Registry = YAML in GitHub repos + static Pages index. The Forum = Go (chi) + SQLite (swappable store) on owner VM, optional.
  Data flow: Compose → Resolve (resolver pulls referenced points, builds dependency/conflict graph, applies docs/04 hierarchy, confirms translator supports every required opinion, emits a resolved speech) → Translate → Build (bootable, fully-unattended installer ISO; deterministic inputs → deterministic output) → Install (zero questions; private pane deployed).

## Single resolver, dual compile targets
- source: docs/03-architecture.md
- type: nfr
- content: The resolver must produce identical results in browser (WASM) and on the build machine (native). The Debate UI never reimplements resolution logic; it calls the WASM build. The CLI and The Forum call the native build.

## Rule-based resolution, no SAT solver
- source: docs/03-architecture.md, docs/04-conflict-resolution.md
- type: nfr
- content: Conflict resolution is rule-based: topological sort over ordering constraints, direct pairwise conflict detection, the docs/04 hierarchy, hardware-conditional evaluation. A full SAT solver is explicitly out of scope; the conflict graph must stay human-readable and every resolution explainable to an average user.

## Hardware detection scope (v1.0)
- source: docs/03-architecture.md
- type: nfr
- content: A small program in the installer scans hardware at install time to resolve hardware-conditional opinions (NVIDIA vs AMD drivers, kernel choice). The Debate UI uses declared/scanned hardware at composition time so hardware conflicts surface during the debate. The full hardware-scanning installer is Phase 6 / post-v1.0; v1.0 supports declared hardware + basic install-time resolution only.

## Opinion metadata minimum set (schema floor)
- source: docs/04-conflict-resolution.md
- type: schema
- content: Every opinion declares: status within its point (`required` | `nice-to-have`); dependencies (opinions/capabilities required); conflicts (cannot coexist); hardware conditions (e.g. `requires: nvidia-gpu`, `requires: uefi`); ordering constraints (must-install-before / must-install-after); known patches (references to patch opinions); translator capability requirements. This set is the validated floor; Phase 0 is expected to expand it (e.g. arbitrary script payloads, theming assets). The exact schema is derived from Phase 0 research, drafted in Phase 1.

## Resolution hierarchy (precise rules)
- source: docs/04-conflict-resolution.md
- type: protocol
- content: When two opinions conflict, apply in order: (1) required beats nice-to-have — nice-to-have dropped visibly, with an explanation; (2) required vs required — hard conflict; user must drop or replace a point, UNLESS a patch opinion exists, in which case it is offered as the resolution; (3) nice-to-have vs nice-to-have — system picks a sensible default or asks the user; (4) patch opinions can override any of the above when one exists for the pair.

## Patch opinions (first-class)
- source: docs/04-conflict-resolution.md
- type: protocol
- content: Discoverable (attached to the conflict pair in metadata; offered automatically by the resolver), not hacks (versioned, maintained, attributable), cumulative (the opinion graph grows more connected over time, resolving more conflicts automatically).

## Hardware-aware resolution
- source: docs/04-conflict-resolution.md
- type: protocol
- content: The composer cross-checks opinions against declared/scanned hardware and suggests swaps (e.g. "You only have NVIDIA — you probably want the NVIDIA gaming point instead") with one-click apply. Version-level incompatibilities use the same machinery: declare, detect, suggest, patch.

## Ordering / topological sort
- source: docs/04-conflict-resolution.md
- type: protocol
- content: Ordering constraints feed a topological sort producing the concrete install order in the resolved speech. Cycles are a hard error surfaced at composition time with the offending opinions named.

## Community conflict-resolution workflow
- source: docs/04-conflict-resolution.md
- type: protocol
- content: New conflict with no known resolution → (1) spin up disposable VM/container reproducing the composition; (2) community works the problem there; (3) solution extracted into a patch opinion + metadata update via PR; (4) resolver thereafter knows the resolution. The Forum hosts the conflict threads and links to resolving patch PRs; patch opinions live in Git so the knowledge survives The Forum.

## Readability invariant
- source: docs/04-conflict-resolution.md
- type: nfr
- content: The conflict graph must remain human-readable; the system must be understandable from the YAML alone; when choosing between smarter automatic resolution and explainable resolution, choose explainable.

## Zero-cost critical path (core constraint)
- source: docs/05-distribution-and-infra.md
- type: nfr
- content: The critical compose→resolve→build path must run on free public tooling and user-owned compute, no central bottleneck. Possible with nothing but a browser (or CLI) plus a free CI tier or local Docker. No central service may be required for this path. The Forum is the single deliberate exception — optional, additive, rebuildable; if offline, the core path still works.

## Registry on GitHub
- source: docs/05-distribution-and-infra.md
- type: api-contract
- content: Points and public speeches are plain YAML files in Git repos; versioning/forking/PRs/attribution come free from GitHub. A static registry index is generated from those repos and hosted on GitHub Pages, rebuilt on commit. GitLab parity desirable, not required for v1.0.

## Build path 1: distributed CI (user's own credits)
- source: docs/05-distribution-and-infra.md
- type: protocol
- content: Project publishes a reusable GitHub Actions workflow/action. User forks a template repo, commits their speech YAML; build runs in THEIR CI on THEIR free-tier minutes: parse speech → resolve against the public registry → run translator → emit installer ISO artifact. The action must be dead-simple with excellent docs. Private pane never leaves the user's repo/CI.

## Build path 2: local Docker (full privacy)
- source: docs/05-distribution-and-infra.md
- type: protocol
- content: A published Docker image bundles resolver + translators + ISO builders. `docker run debateos:latest` with the speech YAML mounted → installer ISO out, locally. The SAME image is used by the GitHub Action internally — one build environment, two delivery channels.

## Deterministic builds
- source: docs/05-distribution-and-infra.md
- type: nfr
- content: Identical inputs → identical output (enables caching/dedup). `SOURCE_DATE_EPOCH` derived from the resolved-speech hash; pinned package snapshots where the foundation allows. Validated in the prior attempt.

## The Forum: service & storage contract
- source: docs/05-distribution-and-infra.md
- type: api-contract
- content: Lean Go service (chi router) on owner-hosted VM, embedded SQLite (pure-Go `modernc.org/sqlite`, libSQL-compatible) behind a thin swappable `store` repository interface with sqlc-generated type-safe queries. SQLite default for dev + v1.0 production (in-memory for tests); Postgres optional drop-in for future scale; no heavyweight ORM. Full-text search via SQLite FTS5, abstracted for a later Postgres `tsvector` impl. Stores: indexed point/speech metadata (repo, version, tags, curator, popularity/freshness), subscription edges, ratings, conflict threads with pointers to GitHub PRs/patch opinions. Never stores: private panes, secrets, credentials, or any non-public primary content.

## The Forum: security properties (all mandatory)
- source: docs/05-distribution-and-infra.md
- type: nfr
- content: (1) Read-mostly index over GitHub; no arbitrary file uploads, no primary content hosting. (2) No untrusted code execution, ever; builds never run on the VM (resolver may run server-side only over already-public, already-indexed GitHub content for indexing). (3) No passwords, no email, no 2FA — GitHub OAuth only; no credentials stored. (4) No secrets at rest. (5) Rebuildable — total DB loss recoverable by re-indexing GitHub. (6) Minimal surface — single static Go binary + one SQLite file.

## The Forum: hosting target
- source: docs/05-distribution-and-infra.md
- type: nfr
- content: Primary: Oracle Cloud Always Free Ampere A1 (ARM) — provision in EU/APAC region (Frankfurt/Singapore/Tokyo) to avoid US capacity constraints. Fallback: owner-provided server. Single static `linux/arm64` Go binary + one SQLite file; no separate DB server. Free-tier reclaim/sleep is a tolerable failure mode by design.

## Private pane storage & secrets model
- source: docs/05-distribution-and-infra.md
- type: nfr
- content: The user's speech (incl. private pane) lives in the home folder, managed by the CLI; optionally backed up to the user's own private Git repo. Public sharing is opt-in and includes only public panes. Secrets are never baked into shared/public artifacts; injected at first boot on the target machine. Detailed key management finalized during Phase 3; the invariant is fixed now.

## Phase sequencing & gating (v1.0 roadmap)
- source: docs/07-roadmap.md
- type: protocol
- content: Phases sequential; Phase 0 gates everything; Phase 6 is post-v1.0 and must not be built.
  - **Phase 0 — Omarchy Research (+ Arch-variant study, per D20):** deliver `research/omarchy-opinion-inventory.md`, `research/omarchy-points.md`, `research/schema-requirements.md`, `research/open-questions.md`, `research/arch-variants-delta.md`, `research/resolver-edge-cases.md`. Success: complete evidence-backed inventory; metadata surface justified by real Omarchy decisions; CachyOS/Garuda deltas cataloged with a proposed variant-profile shape; edge-case corpus ready to seed the Phase 1 test harness. Gates Phase 1.
  - **Phase 1 — Schema & Resolver Core:** `schemas/` (Opinion/Point/Speech YAML), `resolver/` (Go: parse, validate, graph, conflict detection, hierarchy, patch application, hardware checks; native + WASM wired), conflict test harness (required-vs-required, hardware mismatch, version clash, at least one patchable pair), 3–4 example files incl. one deliberately conflicting. Success: resolver resolves every harness scenario per docs/04 rules and explains each resolution; WASM results identical to native. Gates Phase 2.
  - **Phase 2 — Arch Translator:** `translators/arch/` wrapping mkarchiso (structured for 1–2 Arch variants); the Omarchy-as-a-speech composition. Success (NORTH STAR): Omarchy reproducible as a speech on vanilla Arch. Gates Phases 3–5.
  - **Phase 3 — CLI & Build Channels:** `cli/` (compose/validate/build, private-pane mgmt, embedded UI serving), `build/docker/`, `build/actions/` (same image), secrets model per docs/05. Success: Omarchy speech builds end-to-end via both Docker and GitHub Actions, deterministically.
  - **Phase 4 — Debian Translator:** `translators/debian/` wrapping live-build/preseed; documented schema/capability adjustments. Success: a representative speech builds installers for BOTH Arch and Debian from the same resolved input; leaked Arch assumptions identified and fixed.
  - **Phase 5 — Registry, The Forum & Visual Debate UI:** `registry/` static index generator; `web/` Debate UI (SvelteKit + adapter-static + Tailwind, Go-WASM resolver, Pages + CLI-embeddable); `forum/` per docs/05–06 (FTS5 search, subscriptions, ratings, conflict threads); VM deployment notes. Success: discover via Forum, compose in UI with live conflict visualization, proceed to build — core path still functional with Forum offline.
  - **Post-v1.0:** Phase 6 hardware-scanning installer; additional translators (Fedora etc., community-owned); direct-to-disk; full GitLab parity; post-install reconciliation.

## Phase 0 research method & gating
- source: docs/08-omarchy-research.md (amended 2026-06-12: Arch-variant substitution study added per D20)
- type: protocol
- content: Clone and analyze `https://github.com/basecamp/omarchy` source (not blog summaries). Walk every decision made after a base Arch install; record each as a candidate atomic opinion with: category (package install/removal, config/dotfile, service, kernel/boot param, theming, keybinding, hardware-conditional, arbitrary script, …), OS-agnostic expression of intent, dependencies and ordering, and anything that cannot be made OS-agnostic (→ translator capability requirement). Group opinions into candidate points; note surprises that expand the schema (reordered boot steps → ordering metadata mandatory; custom scripts → arbitrary script payloads; theming assets → file payloads). The schema floor comes from this analysis; do not invent the schema before it is done. Validation inverse: Omarchy reproducible as a speech on vanilla Arch (Phase 2).
  PLUS the Arch-variant substitution study (targeted, not a second deep inventory): catalog CachyOS and Garuda deltas from vanilla Arch (extra/replacement repos incl. Chaotic-AUR, repo priority/mirrors, keyrings/signing, kernel variants, default fs/bootloader, pre-seeded configs/tooling); determine what a declarative translator "variant profile" must absorb (repo list + keyring + kernel + defaults — not a fork per variant); record every case where the same opinion effectuates differently across vanilla Arch/CachyOS/Garuda (each is a system-support requirement AND a resolver test scenario); identify variant behaviors that are pre-installed opinions in disguise (informs schema modeling of "the foundation already has an opinion about this"). Extra deliverables: research/arch-variants-delta.md, research/resolver-edge-cases.md. Stretch validation (Phase 2, not a gate): the Omarchy speech retargeted to a variant via a variant profile builds with only declared, explainable differences.

## Monorepo layout (v1.0 target)
- source: docs/11-repo-layout.md
- type: schema
- content: Single Go module rooted at repo (`module github.com/<owner>/debateos`) covering `resolver/`, `cli/`, `forum/`. Tree: `docs/` (founding context), `research/` (Phase 0), `schemas/` (+ CC0 LICENSE), `resolver/` (`parse/ graph/ resolve/ patch/ hardware/ wasm/`), `cli/` (+ `embed/` for UI assets), `translators/` (`arch/`, `debian/`), `build/` (`docker/`, `actions/`), `web/` (SvelteKit), `registry/`, `forum/` (`api/ index/ store/ migrations/ deploy/`), `examples/` (+ CC0 LICENSE, `omarchy/`), `.github/workflows/`. Directories appear as their phase delivers them.

## Module & build-channel contracts
- source: docs/11-repo-layout.md
- type: api-contract
- content: Resolver builds two ways: native (default) and `GOOS=js GOARCH=wasm` via the `resolver/wasm` entrypoint. `web/` build output consumed twice: GitHub Pages and `go:embed` under `cli/embed/` so `debateos compose` serves it offline. Translators are intentionally outside the Go module (shell/Python); the Docker image and CLI invoke them as subprocesses with a defined input contract — the resolved-speech JSON/YAML. Licensing split: root `LICENSE` AGPL-3.0 (code); `schemas/LICENSE` and `examples/LICENSE` CC0-1.0 (content). Community point/speech repos live as separate GitHub repos outside the monorepo.

## Deferred-by-design open questions (must NOT block)
- source: docs/03-architecture.md
- type: nfr
- content: Intentionally deferred; pick the simplest option consistent with the invariants when reached: exact registry static-index format and search UX (refine in Phase 5); direct-to-disk install target (post-v1.0; ISO is the v1.0 target); full GitLab parity (GitHub is the v1.0 bootstrap target); post-install reconciliation (out of v1.0 scope).
