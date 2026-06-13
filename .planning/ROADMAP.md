# Roadmap: DebateOS

## Overview

v1.0 spans Phases 0–5, strictly sequential, exactly as locked by the ADR (docs/07, docs/09). Phase 0 research gates everything: the schema is derived from real Omarchy evidence plus a CachyOS/Garuda variant study, never invented. Phase 1 builds the schema and the rule-based Go resolver (native + WASM) test-first against the Phase 0 edge-case corpus. Phase 2 delivers the Arch translator and the north-star validation: Omarchy reproducible as a speech on vanilla Arch. Phase 3 wraps it all in the Go CLI and two deterministic zero-cost build channels. Phase 4 adds the Debian translator to prove the abstraction across two foundations. Phase 5 ships the discovery layer: Git-backed registry index, optional Forum, and the visual Debate UI. Phase 6 (hardware-scanning installer) and all post-v1.0 items are explicitly out of this roadmap. Per D19, every phase plan specifies test scenarios before implementation tasks.

## Phases

**Phase Numbering:**

- Integer phases (0–5): ADR-locked milestone work (numbering matches docs/07 exactly)
- Decimal phases (e.g. 2.1): urgent insertions only (marked INSERTED)

- [x] **Phase 0: Omarchy Research & Arch-Variant Study** - Evidence-backed opinion inventory + CachyOS/Garuda delta study; six deliverables that gate all design (completed 2026-06-12)
- [x] **Phase 1: Schema & Resolver Core** - Opinion/Point/Speech schemas from Phase 0 data; rule-based Go resolver, native + WASM, built test-first (completed 2026-06-12)
- [x] **Phase 2: Arch Translator** - mkarchiso-wrapping translator with variant-profile structure; NORTH STAR: Omarchy as a speech on vanilla Arch (completed 2026-06-12)
- [x] **Phase 3: CLI & Build Channels** - Go CLI, private-pane/secrets model, deterministic Docker + GitHub Actions builds at zero cost (completed 2026-06-13)
- [x] **Phase 4: Debian Translator** - live-build/preseed translator; dual-foundation proof from one resolved speech; de-Arch the abstraction (completed 2026-06-13)
- [ ] **Phase 5: Registry, Forum & Debate UI** - Static registry index, optional Forum (search/subscriptions/ratings/conflict threads), visual Debate UI on Pages + embedded in CLI

## Phase Details

### Phase 0: Omarchy Research & Arch-Variant Study

**Goal**: The schema floor and resolver test corpus exist as evidence, not theory — every later design decision traces to real Omarchy and Arch-variant data
**Depends on**: Nothing (first phase; gates everything after it)
**Requirements**: RSCH-01, RSCH-02, RSCH-03
**Success Criteria** (what must be TRUE):

  1. All six research deliverables exist in `research/` (omarchy-opinion-inventory, omarchy-points, schema-requirements, open-questions, arch-variants-delta, resolver-edge-cases), built from the cloned `basecamp/omarchy` source, not blog summaries
  2. Every post-base-Arch Omarchy decision is recorded as a candidate atomic opinion with category, OS-agnostic intent, dependencies/ordering, and anything un-agnostic flagged as a translator capability requirement; opinions grouped into candidate points
  3. The proposed opinion metadata surface is justified by real Omarchy decisions (schema surprises like ordering, script payloads, theming assets explicitly captured), and CachyOS/Garuda deltas are cataloged with a proposed declarative variant-profile shape
  4. The resolver edge-case corpus is written as concrete test scenarios ready to seed the Phase 1 TDD harness (foundation-default vs opinion collisions, repo-priority conflicts, cross-variant effectuation differences)

**Plans**: 4 plansPlans:
**Wave 1**

- [x] 00-01-PLAN.md — Exhaustive Omarchy opinion inventory (OM-NNN atomic entries from cloned source)
- [x] 00-03-PLAN.md — CachyOS/Garuda variant delta study + declarative variant-profile sketch

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 00-02-PLAN.md — Point groupings + evidence-backed schema-requirements floor (SR-NNN)
- [x] 00-04-PLAN.md — Resolver edge-case corpus (EC-NNN Given/When/Then) + open-questions

### Phase 1: Schema & Resolver Core

**Goal**: A composition can be parsed, validated, and resolved — every conflict handled per the docs/04 hierarchy with a human-readable explanation, identically in native and WASM
**Depends on**: Phase 0
**Requirements**: SCHM-01, SCHM-02, RSLV-01, RSLV-02, RSLV-03, RSLV-04, RSLV-05, RSLV-06
**Success Criteria** (what must be TRUE):

  1. Opinion/Point/Speech YAML schemas exist in `schemas/` (CC0), cover the full Phase 0-derived metadata floor, and a person can understand any example composition and resolution from the YAML alone
  2. The resolver resolves every harness scenario per the docs/04 rules: nice-to-have drops are visible with explanations, required-vs-required is a hard conflict unless a patch opinion exists (then offered automatically), nice-vs-nice picks a sensible default, ordering cycles fail with the offending opinions named
  3. Hardware-conditional opinions resolve against declared hardware with swap suggestions surfaced at composition time
  4. WASM and native builds produce identical results, proven by automated parity tests; resolver coverage is near-total per D19, with the Phase 0 edge-case corpus encoded as tests before implementation
  5. 3–4 example files exist (including one deliberately conflicting) that exercise the harness end-to-end

**Plans**: 5 plans
Plans:
**Wave 1**

- [x] 01-01-PLAN.md — Module bootstrap + shared types + JSON Schema 2020-12 + SR traceability + parse layer + licenses (SCHM-01, SCHM-02)

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 01-02-PLAN.md — Dependency/ordering graph + deterministic Kahn toposort with cycle detection (RSLV-03)
- [x] 01-03-PLAN.md — Compound hardware predicate evaluation + first-class patch discovery (RSLV-02, RSLV-04)

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 01-04-PLAN.md — Resolve engine: docs/04 hierarchy + Explanation + canonical JSON + 27 EC corpus (RSLV-01, RSLV-06)

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 01-05-PLAN.md — WASM entrypoint + native/WASM parity + coverage gate + 4 example compositions (RSLV-05, RSLV-06)

### Phase 2: Arch Translator

**Goal**: A resolved speech becomes a bootable, fully-unattended Arch installer — and Omarchy is reproducible as a speech on vanilla Arch (the north star)
**Depends on**: Phase 1
**Requirements**: ARCH-01, ARCH-02, ARCH-03, ARCH-04
**Success Criteria** (what must be TRUE):

  1. NORTH STAR (invariant 6): building the Omarchy speech in `examples/omarchy/` produces an installed system equivalent to Omarchy on vanilla Arch, with zero install-time questions
  2. The Arch translator consumes a resolved speech via the defined input contract, wraps mkarchiso, and emits a bootable unattended ISO from inside the isolated build environment
  3. The translator declares its supported opinions/capabilities, and a speech containing an unsupported required opinion breaks visibly at composition time, never silently at install time
  4. The translator is structured around declarative variant profiles (repo list + keyring + kernel + defaults) per the Phase 0 delta study, with no per-variant fork
  5. STRETCH (non-gating, per D20): the Omarchy speech retargeted to CachyOS or Garuda via a variant profile builds with only declared, explainable differences

**Plans**: 5 plans
Plans:
**Wave 1**

- [x] 02-01-PLAN.md — Translator package + input-contract loader + capability gate (SC-3) + BuildManifest builder (ARCH-01, ARCH-03) (completed 2026-06-12)
- [x] 02-03-PLAN.md — Declarative variant profiles: vanilla-arch / cachyos / garuda YAML with marked Omarchy conflicts (ARCH-04) (completed 2026-06-12)
- [x] 02-04-PLAN.md — examples/omarchy authoring: 134 opinions + 32 points + speech + clean-resolve Go harness (ARCH-02) (completed 2026-06-12)

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 02-02-PLAN.md — archiso profile emitter + variant application (no fork) + first-run units + generate()/translate entrypoint (ARCH-01, ARCH-04) (completed 2026-06-12)

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 02-05-PLAN.md — Slow gates (Docker mkarchiso build, ISO structural validation, north-star equivalence) + README + status (ARCH-01..04) (completed 2026-06-12)

### Phase 3: CLI & Build Channels

**Goal**: Anyone can go compose → resolve → build to an ISO at zero cost, via local Docker or their own GitHub Actions minutes, deterministically, with their private pane never leaving their control
**Depends on**: Phase 2
**Requirements**: CLI-01, CLI-02, BLD-01, BLD-02, BLD-03, BLD-04, PRIV-01
**Success Criteria** (what must be TRUE):

  1. `debateos compose | validate | build | pane` work against the native resolver, with the speech (including private pane) managed in `$HOME` and optionally backed up to the user's own private Git repo
  2. The Omarchy speech builds end-to-end via local Docker (`docker run` with speech mounted → ISO out) AND via the published reusable GitHub Actions workflow on a forked template repo — the same image powering both channels
  3. Builds are deterministic — identical inputs produce identical ISOs, `SOURCE_DATE_EPOCH` keyed to the resolved-speech hash — verified by automated tests, not inspection
  4. The full path runs at zero hosting cost with no central service involved; secrets/private pane never appear in shared artifacts and inject at first boot, with the key-management design finalized and documented

**Plans**: 4 plans
Plans:
**Wave 1**

- [x] 03-01-PLAN.md — CLI foundation: config-dir resolver + Runner interface + compose/validate subcommands + dispatch entrypoint (CLI-01) (completed 2026-06-13)

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 03-02-PLAN.md — Private pane + age X25519 backup/restore via Runner; 0600 enforcement; key-management finalized (CLI-02, PRIV-01)

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 03-03-PLAN.md — build subcommand: resolve→epoch→translate→docker via Runner; --dry-run/--skip-iso; private-injection.tar (CLI-01, BLD-01, BLD-03, PRIV-01) (completed 2026-06-13)

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 03-04-PLAN.md — Multi-stage Docker image + reusable Actions workflow + thin caller + determinism/secret-free/coverage gates + docs + status (BLD-01..04, PRIV-01) (completed 2026-06-13)

### Phase 4: Debian Translator

**Goal**: The opinion/translator abstraction is proven real, not Arch-shaped — one resolved speech yields installers for two foundations
**Depends on**: Phase 3
**Requirements**: DEB-01, DEB-02, DEB-03, COMM-01
**Success Criteria** (what must be TRUE):

  1. DUAL-FOUNDATION PROOF: a representative speech builds bootable, fully-unattended installers for BOTH Arch and Debian from the same resolved input
  2. The Debian translator wraps live-build/preseed, declares its capabilities, and unsupported required opinions break visibly at composition time
  3. Arch assumptions that leaked into the schema, resolver, or example opinions are identified and fixed, with schema/capability adjustments documented
  4. The translator ownership model is documented: distributions own their translators, curators own points/speeches, community PRs welcome

**Plans**: 5 plans
Plans:
**Wave 1**

- [x] 04-01-PLAN.md — Extract translators/common/ shared core (contract/manifest/firstrun) + Arch regression gate (DEB-03)
- [x] 04-02-PLAN.md — Foundation-neutral dual-foundation example speech + clean-resolve Go gate (DEB-02)

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 04-03-PLAN.md — Debian translator core: capability gate + apt sig_level mapping + preseed/chroot-hook emitter + translate entrypoint (DEB-01, DEB-03)
- [x] 04-04-PLAN.md — Foundation-aware build.go dispatch (foundationRegistry) + DEB-03 Arch-leak audit doc (DEB-01, DEB-03)

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 04-05-PLAN.md — dual-foundation-check gate + Debian build/validate scripts + ownership-model.md + REQUIREMENTS status (DEB-01, DEB-02, COMM-01) (completed 2026-06-13)

### Phase 5: Registry, Forum & Debate UI

**Goal**: Users can discover points, compose visually with live conflict resolution, and proceed to build — with Git authoritative and the Forum strictly optional
**Depends on**: Phase 4
**Requirements**: REG-01, UI-01, UI-02, BRND-01, FORM-01, FORM-02, FORM-03, FORM-04, FORM-05
**Success Criteria** (what must be TRUE):

  1. A user can discover points/speeches via the Forum (search by curator, tag, popularity, freshness, foundation compatibility), subscribe to curators or individual points, rate via GitHub OAuth identity, and follow conflict threads that link to resolving patch-opinion PRs
  2. A user can compose a speech in the Debate UI with live conflict visualization (panes, red/green overlaps) powered by the client-side WASM resolver, and proceed to build instructions — in the GitHub Pages deployment AND the identical UI served offline by `debateos compose`
  3. The static registry index generates from the GitHub YAML repos, deploys to GitHub Pages, and rebuilds on commit — Git remains the authoritative source of record
  4. With the Forum offline, the entire compose → resolve → build path still works (invariant 4); total Forum DB loss is recoverable by re-indexing GitHub; deployment notes cover the D15 hosting target
  5. The debate-themed brand voice is applied consistently across the UI and docs without obscuring meaning

**Plans**: TBD
**UI hint**: yes

## Progress

**Execution Order:**
Phases execute in numeric order: 0 → 1 → 2 → 3 → 4 → 5 (strictly sequential; each phase gates the next per docs/07)

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 0. Omarchy Research & Arch-Variant Study | 4/4 | Complete   | 2026-06-12 |
| 1. Schema & Resolver Core | 5/5 | Complete   | 2026-06-12 |
| 2. Arch Translator | 5/5 | Complete   | 2026-06-13 |
| 3. CLI & Build Channels | 4/4 | Complete   | 2026-06-13 |
| 4. Debian Translator | 5/5 | Complete   | 2026-06-13 |
| 5. Registry, Forum & Debate UI | 0/TBD | Not started | - |

---
*Roadmap created: 2026-06-12 from ADR-locked phase structure (docs/07, docs/09 via .planning/intel/)*
*Post-v1.0 (NOT in this roadmap): Phase 6 hardware-scanning installer, Fedora/other translators, direct-to-disk, full GitLab parity, post-install reconciliation*
