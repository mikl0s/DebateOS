# Requirements Intel

Extracted from PRD sources (`docs/01-vision.md`, `docs/06-social-layer.md`). Each requirement carries source attribution and acceptance criteria as stated. No competing acceptance variants were found between the two PRDs (they cover complementary scopes).

---

## REQ-compose-build-zero-cost
- source: docs/01-vision.md
- description: A user composes a speech from several curators' points, resolves conflicts visually, and produces a bootable, fully-unattended installer for the foundation of their choice.
- acceptance: End-to-end compose → resolve → build works at zero hosting cost (free public tooling + user-owned compute).
- scope: core user journey

## REQ-omarchy-north-star
- source: docs/01-vision.md
- description: Omarchy is reproducible as a speech on vanilla Arch.
- acceptance: Building the Omarchy speech produces an installed system equivalent to Omarchy (the concrete validation north star; Phase 2 milestone).
- scope: model validation

## REQ-dual-foundation-proof
- source: docs/01-vision.md
- description: Two foundations (Arch + Debian) prove the opinion/translator abstraction is real, not an Arch-shaped illusion.
- acceptance: A representative speech builds installers for both Arch and Debian from the same resolved input.
- scope: abstraction validation

## REQ-curator-ecosystem
- source: docs/01-vision.md
- description: Curators accumulate subscribers; popular speeches become de-facto "distributions" maintained purely as configuration.
- acceptance: Speeches/points are shareable, forkable, subscribable artifacts that build curator reputation (qualitative adoption outcome).
- scope: ecosystem outcome

## REQ-human-readable-yaml
- source: docs/01-vision.md
- description: The visual Debate is the eventual primary interface, but the YAML underneath must remain comprehensible on its own; the system must never decay into a byzantine dependency resolver.
- acceptance: Every resolution is explainable from the YAML alone.
- scope: usability / philosophy (reinforces locked invariant 3)

## REQ-anti-dogmatic-brand
- source: docs/01-vision.md, docs/06-social-layer.md
- description: Playful, tongue-in-cheek rhetoric metaphor throughout (opinions, points, speeches, debates). Tagline: "That's just your opinion, man." There are no conclusions, only debates.
- acceptance: Brand voice applied across UI/docs, softened only where it would obscure meaning; core to identity.
- scope: brand / tone

## REQ-forum-search-discovery
- source: docs/06-social-layer.md
- description: Search and discovery across indexed points and speeches — by curator, tag, popularity, freshness, and foundation compatibility.
- acceptance: The Forum surfaces popularity, freshness, and compatibility information; underlying data lives in Git.
- scope: The Forum (v1.0)

## REQ-forum-subscriptions
- source: docs/06-social-layer.md
- description: Subscriptions — follow curators and see their updates; subscribe to whole point sets or individual points.
- acceptance: A user can subscribe to multiple curators' points and the system merges them into one coherent speech, resolves overlaps/conflicts, and generates a single installer.
- scope: The Forum (v1.0)

## REQ-forum-ratings-reputation
- source: docs/06-social-layer.md
- description: Ratings / reputation — lightweight, GitHub-identity-backed.
- acceptance: Ratings tied to GitHub OAuth identity; no DebateOS-native account system.
- scope: The Forum (v1.0)

## REQ-forum-collab-conflict-resolution
- source: docs/06-social-layer.md
- description: Collaborative conflict resolution — The Forum hosts conflict threads from the docs/04 workflow: a registry of known conflicts, disposable-environment repros, discussion, and links to the patch opinions (GitHub PRs) that resolve them.
- acceptance: Conflict resolution becomes extractable, reusable patches instead of forum-thread folklore; patch opinions themselves live in Git and survive The Forum.
- scope: The Forum (v1.0)

## REQ-forum-boundaries
- source: docs/06-social-layer.md
- description: The Forum is NOT the system of record (Git/GitHub is), NOT a build service (no code execution), NOT an account system (GitHub OAuth only), and NOT required to compose or build a speech.
- acceptance: Losing The Forum means re-indexing, never data loss; the core path works with The Forum offline.
- scope: The Forum boundaries (reinforces locked D13/D14)

## REQ-registry-authoritative
- source: docs/06-social-layer.md
- description: The Git-backed registry is the decentralized source of truth; The Forum is an index and a meeting place on top of it.
- acceptance: Registry authoritative; Forum additive.
- scope: social layer architecture

## REQ-translator-ownership-model
- source: docs/06-social-layer.md, docs/01-vision.md
- description: Distributions are invited to own their translators (Ubuntu controls Ubuntu's); curators own points and speeches; community PRs welcome.
- acceptance: Project bootstraps Arch (+ 1–2 Arch variants) and Debian translators; ownership model documented.
- scope: community model

---

## Scope exclusions (v1.0 non-goals)
- source: docs/01-vision.md

- NO monetization: non-commercial, no paid tiers, no central SaaS dependency.
- NO post-install reconciliation: applying speech changes to an already-installed running system is out of scope (install-time only).
- NO hardware-scanning installer: deferred to Phase 6 (post-v1.0); v1.0 targets ISO/USB installer output.
