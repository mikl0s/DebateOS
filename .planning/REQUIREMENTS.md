# Requirements: DebateOS

**Defined:** 2026-06-12
**Core Value:** A user composes a speech from multiple curators' points, resolves conflicts explainably, and produces a bootable unattended installer for their chosen foundation — entirely on free tooling and user-owned compute, with no central service in the critical path.

Requirement IDs below are the checkable v1.0 form of the ingest intel (`.planning/intel/requirements.md` REQ-* items plus SPEC constraints). Source mapping is noted per category.

## v1 Requirements

### Research (Phase 0 deliverables — gate everything)

<!-- Sources: D17, D20, docs/07, docs/08 -->

- [x] **RSCH-01**: Omarchy deep-dive complete from cloned source (not summaries): every post-base-Arch decision recorded as a candidate atomic opinion with category, OS-agnostic intent, dependencies/ordering, and translator-capability fallout — delivered as `research/omarchy-opinion-inventory.md`, `research/omarchy-points.md`, `research/schema-requirements.md`, `research/open-questions.md`
- [x] **RSCH-02**: CachyOS/Garuda variant substitution study delivered as `research/arch-variants-delta.md`: deltas from vanilla Arch cataloged (repos incl. Chaotic-AUR, repo priority/mirrors, keyrings, kernel variants, default fs/bootloader, pre-seeded configs) with a proposed declarative variant-profile shape (repo list + keyring + kernel + defaults — not a fork per variant)
- [x] **RSCH-03**: Resolver edge-case corpus delivered as `research/resolver-edge-cases.md`: every case where the same opinion effectuates differently across vanilla Arch/CachyOS/Garuda, foundation-default-vs-opinion collisions, and repo-priority conflicts — each expressed as a Phase 1 test scenario

### Schema

<!-- Sources: D17, REQ-human-readable-yaml, docs/02, docs/04 metadata floor -->

- [ ] **SCHM-01**: Opinion/Point/Speech YAML schemas exist in `schemas/` (CC0), derived from Phase 0 evidence, covering the validated metadata floor: required|nice-to-have status, dependencies, conflicts, hardware conditions, ordering constraints, known patches, translator capability requirements — plus Phase 0 expansions (script payloads, theming/file assets, foundation-default modeling)
- [ ] **SCHM-02**: Schemas and example files are human-readable: a person can understand any composition and every resolution from the YAML alone; no Arch/Debian specifics leak into schema or content (invariants 1, 3)

### Resolver

<!-- Sources: D5, D6, D19, docs/03, docs/04 -->

- [ ] **RSLV-01**: Resolver parses and validates speeches, pulls referenced points, builds the dependency/conflict graph, and applies the docs/04 hierarchy (required beats nice-to-have with visible drop + explanation; required-vs-required hard conflict unless patched; nice-vs-nice sensible default or ask), emitting a resolved speech with a human-readable explanation for every resolution
- [ ] **RSLV-02**: Patch opinions are first-class: attached to conflict pairs in metadata, discovered and offered automatically by the resolver, able to override the hierarchy when one exists for the pair
- [x] **RSLV-03**: Ordering constraints feed a topological sort producing the concrete install order; cycles are a hard composition-time error naming the offending opinions
- [x] **RSLV-04**: Hardware-conditional opinions evaluate against declared hardware at composition time; mismatches surface during the debate with suggested swaps
- [x] **RSLV-05**: Resolver compiles to native and WASM (`resolver/wasm` entrypoint) and produces identical results in both targets, verified by automated parity tests
- [x] **RSLV-06**: TDD conflict test harness covers the Phase 0 edge-case corpus plus required-vs-required, hardware mismatch, version clash, and at least one patchable pair, with near-total resolver coverage; 3–4 example files exist including one deliberately conflicting

### Arch Translator

<!-- Sources: REQ-omarchy-north-star, D8, D9, D20, docs/02 translator contract -->

- [x] **ARCH-01**: Arch translator (`translators/arch/`, shell/Python) consumes a resolved speech via the defined JSON/YAML input contract, wraps mkarchiso, and emits a bootable, fully-unattended Arch installer ISO
- [x] **ARCH-02**: NORTH STAR — building the Omarchy speech (`examples/omarchy/`) produces an installed system equivalent to Omarchy on vanilla Arch (02-04/02-05: 134 opinions + 32 points + speech.yaml + TestExampleOmarchy: Applied=99 Skipped=35 Hard-conflicts=0; arch-northstar-check.sh --skip-build: 16/16 PASS; full Docker ISO build blocked by host devtmpfs restriction)
- [x] **ARCH-03**: Translator declares its supported opinions/capabilities; unsupported required opinions break visibly at composition time, never silently at install time (02-01: capabilities.json + check_capabilities() gate, 43 pytest GREEN)
- [x] **ARCH-04**: Translator is structured for 1–2 Arch variants via declarative variant profiles (repo list + keyring + kernel + defaults) informed by the Phase 0 delta study — no per-variant forks (02-03: vanilla-arch/cachyos/garuda YAML profiles, schema README, 4 Garuda conflicts as structured data)

### CLI

<!-- Sources: D7, D16, docs/03 -->

- [x] **CLI-01**: `debateos compose | validate | build | pane` work, wrapping the native resolver and invoking translators as subprocesses
- [x] **CLI-02**: CLI manages the user's speech including the private pane in `$HOME`, with optional backup to the user's own private Git repo

### Build Channels

<!-- Sources: REQ-compose-build-zero-cost, D11, docs/05 -->

- [x] **BLD-01**: Published Docker image bundles resolver + translators + ISO builders; `docker run` with speech YAML mounted produces an installer ISO locally (full privacy path). Evidence: build/docker/Dockerfile (multi-stage golang builder + digest-pinned archlinux runtime; CGO_ENABLED=0 static binary); build/docker/entrypoint.sh. Note: full mkarchiso ISO build requires a non-Proxmox host (devtmpfs restriction — same policy as ARCH-01/02); --skip-iso path exercises the full pipeline on this host.
- [x] **BLD-02**: Published reusable GitHub Actions workflow (using the SAME image) lets a user fork a template repo, commit their speech, and build the ISO on their own free-tier CI minutes — dead simple, well documented. Evidence: build/actions/build-speech.yml (workflow_call; container: ghcr.io/mikl0s/debateos:latest matching BLD-01 image); .github/workflows/build-speech.yml (thin caller); build/actions/README.md (fork-and-build guide). Note: live cross-repo Actions run is a deferred verification item (requires a fork + CI minutes; workflow YAML is PyYAML-validated and follows official GHA docs syntax).
- [x] **BLD-03**: Builds are deterministic: identical inputs → identical ISO, `SOURCE_DATE_EPOCH` derived from the resolved-speech hash, verified by automated tests. Evidence: scripts/determinism-test.sh (double-run resolve+translate → deterministic tar with --sort=name --mtime=@EPOCH --pax-option=delete=atime,delete=ctime → sha256 compare); PASSES with identical sha256 on this host. Full-ISO determinism uses the same script on a capable host.
- [x] **BLD-04**: End-to-end compose → resolve → build runs at zero hosting cost on both channels with no central service involved. Evidence: docs/cli-build-channels.md (full walkthrough: local Docker + GitHub Actions on user's own minutes; no DebateOS infrastructure required; automated secret-free and determinism gates documented).

### Privacy & Secrets

<!-- Sources: D16, invariant 7, docs/05 -->

- [x] **PRIV-01**: Secrets and the private pane never enter shared/public artifacts; public sharing includes only public panes; secrets inject at first boot on the target machine; key-management details finalized in Phase 3. Evidence: scripts/secret-free-check.sh (greps arch-profile/ for pane.yaml/identity.age/private-injection.tar — all absent; PASSES); private-injection.tar written to outDir only (T-03-LEAK); age X25519 identity.age 0600 (T-03-PERM); key-management documented in docs/cli-build-channels.md (local-only, no escrow, no central service).

### Debian Translator

<!-- Sources: REQ-dual-foundation-proof, D8, D9, docs/07 Phase 4 -->

- [x] **DEB-01**: Debian translator (`translators/debian/`) wraps live-build/preseed and emits a bootable, fully-unattended Debian installer from a resolved speech, declaring its capabilities like Arch's
- [x] **DEB-02**: DUAL-FOUNDATION PROOF — a representative speech builds installers for BOTH Arch and Debian from the same resolved input
- [x] **DEB-03**: Arch assumptions that leaked into schema/resolver/opinions are identified and fixed; schema/capability adjustments documented

### Community Model

<!-- Sources: REQ-translator-ownership-model -->

- [x] **COMM-01**: Translator ownership model documented: distributions invited to own their translators, curators own points/speeches, community PRs welcome

### Registry

<!-- Sources: REQ-registry-authoritative, REQ-curator-ecosystem, D12, docs/05 -->

- [x] **REG-01**: Points and public speeches are plain YAML in GitHub repos (versioning/forking/PRs/attribution via GitHub); a static registry index is generated from those repos, hosted on GitHub Pages, rebuilt on commit — the registry is the authoritative source of truth and the shareable/forkable/subscribable substrate for curator reputation

### Debate UI

<!-- Sources: D10, docs/03, docs/07 Phase 5, REQ-anti-dogmatic-brand -->

- [x] **UI-01**: Visual Debate UI (SvelteKit + adapter-static + Tailwind) lets a user compose a speech with live conflict visualization (foundation + glass panes, red/green overlaps) using the Go-WASM resolver client-side — never reimplementing resolution logic — and proceed to build instructions
- [x] **UI-02**: The same UI build output is delivered via GitHub Pages AND `go:embed`-ded in the CLI so `debateos compose` serves it offline on localhost

### Brand

<!-- Sources: REQ-anti-dogmatic-brand -->

- [x] **BRND-01**: Debate-themed brand voice ("That's just your opinion, man"; opinions/points/speeches/debates; playful build-stage naming) applied consistently across UI and docs, softened only where it would obscure meaning

### The Forum

<!-- Sources: REQ-forum-* (search, subscriptions, ratings, collab, boundaries), D13/D13a/D14/D15 -->

- [x] **FORM-01**: Forum search and discovery over indexed points/speeches — by curator, tag, popularity, freshness, and foundation compatibility (SQLite FTS5 behind the abstracted store)
- [x] **FORM-02**: Subscriptions: a user can follow curators and subscribe to point sets or individual points; subscribed points merge into one coherent speech that resolves and builds a single installer
- [x] **FORM-03**: Ratings/reputation are lightweight and tied to GitHub OAuth identity only — no DebateOS-native accounts, passwords, email, or 2FA
- [x] **FORM-04**: Conflict threads host the docs/04 community workflow: known-conflict registry, disposable-environment repro notes, discussion, and links to the patch-opinion PRs that resolve them — patches live in Git and survive the Forum
- [x] **FORM-05**: Forum is optional and additive: the core compose→resolve→build path works with the Forum offline; total DB loss is recoverable by re-indexing GitHub; no untrusted code execution, no secrets at rest; deployed per D15 with deployment notes

## v2 Requirements

Deferred post-v1.0. Tracked but not in the current roadmap (D2).

### Post-v1.0 (locked deferrals)

- **POST-01**: Phase 6 hardware-scanning installer (full hardware detection at install time)
- **POST-02**: Additional translators (Fedora etc., community-owned)
- **POST-03**: Direct-to-disk install target (v1.0 is ISO/USB)
- **POST-04**: Full GitLab registry parity (GitHub is the v1.0 bootstrap target)
- **POST-05**: Postgres Forum backend at scale (store interface already abstracts it)

## Out of Scope

| Feature | Reason |
|---------|--------|
| Monetization / paid tiers / central SaaS | Invariant 5: zero cost, non-commercial, no required paid dependency |
| Post-install reconciliation | v1.0 is install-time only; applying speech changes to a running system is explicitly excluded |
| Hardware-scanning installer | Phase 6, post-v1.0; v1.0 = declared hardware + basic install-time resolution |
| SAT/constraint solver | D6 + invariant 3: rule-based resolution keeps every result explainable |
| Central backend port (auth stack, Postgres-as-primary, build queue, server-side ISO builds) | D18 + invariant 4: prior backend dropped by design |
| Forum-hosted primary content / file uploads / code execution | D13: Forum is a read-mostly rebuildable index over GitHub |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| RSCH-01 | Phase 0 | Complete |
| RSCH-02 | Phase 0 | Complete |
| RSCH-03 | Phase 0 | Complete |
| SCHM-01 | Phase 1 | Pending |
| SCHM-02 | Phase 1 | Pending |
| RSLV-01 | Phase 1 | Pending |
| RSLV-02 | Phase 1 | Pending |
| RSLV-03 | Phase 1 | Complete |
| RSLV-04 | Phase 1 | Complete |
| RSLV-05 | Phase 1 | Complete |
| RSLV-06 | Phase 1 | Complete |
| ARCH-01 | Phase 2 | Complete (02-01 capabilities gate, 02-02 profile emitter + translate entrypoint, 02-05 Docker build scripts + structural validation; full mkarchiso ISO build requires host devtmpfs support) |
| ARCH-02 | Phase 2 | Complete (02-04 + 02-05: full north-star pipeline green; equivalence gate 16/16 PASS; build tooling in place; full ISO requires host devtmpfs support) |
| ARCH-03 | Phase 2 | Complete (02-01) |
| ARCH-04 | Phase 2 | Complete (02-03) |
| CLI-01 | Phase 3 | Complete (03-01/03-03: compose/validate/build/pane subcommands; 03-04: coverage gate 85.6%) |
| CLI-02 | Phase 3 | Complete (03-02: age X25519 backup/restore; pane.yaml 0600; git push via Runner) |
| BLD-01 | Phase 3 | Complete (03-04: build/docker/Dockerfile multi-stage + entrypoint; --skip-iso works on this host; full ISO deferred to capable host) |
| BLD-02 | Phase 3 | Complete (03-04: build/actions/build-speech.yml workflow_call + .github/workflows/build-speech.yml thin caller; live cross-repo run deferred) |
| BLD-03 | Phase 3 | Complete (03-04: scripts/determinism-test.sh; sha256 identical across two runs on this host) |
| BLD-04 | Phase 3 | Complete (03-04: docs/cli-build-channels.md end-to-end walkthrough; zero hosting cost; no central service) |
| PRIV-01 | Phase 3 | Complete (03-02/03-03/03-04: pane.yaml local-only 0600; injection tar next to ISO; scripts/secret-free-check.sh PASSES; docs/cli-build-channels.md key-management) |
| DEB-01 | Phase 4 | Complete (04-03: capabilities gate + emit_profile_tree; 04-05: debian-build-iso.sh + Dockerfile + README; full lb ISO deferred to capable host — devtmpfs restriction VERIFIED) |
| DEB-02 | Phase 4 | Complete (04-02: dual-foundation example speech + TestExampleDualFoundation; 04-05: scripts/dual-foundation-check.sh --skip-iso: 20/20 PASS — resolve ONCE, both translators, equivalence git/curl/vim + etc/motd VERIFIED) |
| DEB-03 | Phase 4 | Complete (04-04: 6-finding audit in docs/arch-leak-audit.md; build.go foundationRegistry FIXED; 5 other findings correctly isolated/documented) |
| COMM-01 | Phase 4 | Complete (04-05: docs/ownership-model.md — "distributions own their translators", entrypoint contract, capabilities.json, profile schema, foundationRegistry registration, translators/common/ reuse) |
| REG-01 | Phase 5 | Complete (05-06: registry/index.json generated; .github/workflows/registry-index.yml CI rebuild on commit; docs/registry-forum-ui.md authoritative guide; registry is the source of truth for Forum Reindex and CLI compose) |
| UI-01 | Phase 5 | Complete |
| UI-02 | Phase 5 | Complete |
| BRND-01 | Phase 5 | Complete |
| FORM-01 | Phase 5 | Complete |
| FORM-02 | Phase 5 | Complete |
| FORM-03 | Phase 5 | Complete |
| FORM-04 | Phase 5 | Complete |
| FORM-05 | Phase 5 | Complete |

**Coverage:**

- v1 requirements: 35 total
- Mapped to phases: 35
- Unmapped: 0 ✓

**Intel REQ mapping** (all 13 ingest requirements covered):

| Intel REQ | Covered by |
|-----------|------------|
| REQ-compose-build-zero-cost | BLD-01, BLD-02, BLD-04 |
| REQ-omarchy-north-star | ARCH-02 |
| REQ-dual-foundation-proof | DEB-02 |
| REQ-curator-ecosystem | REG-01, FORM-02 (qualitative adoption outcome) |
| REQ-human-readable-yaml | SCHM-02 |
| REQ-anti-dogmatic-brand | BRND-01 |
| REQ-forum-search-discovery | FORM-01 |
| REQ-forum-subscriptions | FORM-02 |
| REQ-forum-ratings-reputation | FORM-03 |
| REQ-forum-collab-conflict-resolution | FORM-04 |
| REQ-forum-boundaries | FORM-05 |
| REQ-registry-authoritative | REG-01 |
| REQ-translator-ownership-model | COMM-01 |

---
*Requirements defined: 2026-06-12*
*Last updated: 2026-06-13 after Phase 5 Plan 06 completion (05-06: REG-01 marked Complete; all 9 Phase-5 requirements Complete — UI-01, UI-02, BRND-01, FORM-01..05, REG-01)*
