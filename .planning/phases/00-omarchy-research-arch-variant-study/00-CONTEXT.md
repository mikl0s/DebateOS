# Phase 0: Omarchy Research & Arch-Variant Study - Context

**Gathered:** 2026-06-12
**Status:** Ready for planning
**Mode:** Smart discuss (autonomous — recommended answers auto-accepted per owner directive + ADR process notes)

<domain>
## Phase Boundary

Phase 0 produces the evidence base that gates everything downstream: six research deliverables under `research/` derived from the cloned `basecamp/omarchy` source (exhaustive) plus a targeted CachyOS/Garuda delta study (per D20). The phase delivers documents only — no schemas, no code, no resolver. The schema floor and the Phase 1 TDD edge-case corpus come out of this phase as traceable evidence, not theory (D17).

Out of scope: drafting the actual Opinion/Point/Speech schema (Phase 1), any translator code (Phase 2+), installing CachyOS/Garuda in VMs (delta catalog is derived from their public repos and docs).

</domain>

<decisions>
## Implementation Decisions

### Research Method & Evidence Standards
- Pin Omarchy to the latest stable release tag at clone time; record tag + commit hash in every deliverable so all evidence is reproducible.
- Every opinion-inventory entry cites its source path (install script, config file, dotfile) in the cloned repo at the pinned commit.
- CachyOS/Garuda evidence comes from their public git repos (repo definitions, calamares configs, kernel PKGBUILDs) and official docs — no full ISO installs required for a delta catalog.
- Depth boundary: exhaustive for Omarchy (every post-base-Arch decision, per docs/08); targeted delta-only for the variants (repos, keyrings, kernels, defaults, pre-seeded configs/tooling).

### Opinion Inventory Structure
- Markdown inventory with fixed per-opinion fields: stable ID (`OM-NNN`), category, OS-agnostic intent, dependencies, ordering constraints, un-agnostic flags (→ translator capability requirements).
- Category taxonomy starts from the docs/08 list (package install/removal, config/dotfile, service, kernel/boot param, theming, keybinding, hardware-conditional, arbitrary script); new categories are added when evidence demands and each is flagged as a schema surprise.
- Atomic granularity: one opinion per post-install decision. Grouping happens only in `omarchy-points.md`.
- Hardware-conditional behaviors (e.g. Apple display bindings) are recorded as hardware-conditional opinions with explicit condition metadata.

### Resolver Edge-Case Corpus (feeds Phase 1 TDD per D19)
- Structured Given/When/Then scenarios with stable IDs (`EC-NNN`) and expected explanation text, so Phase 1 tests trace 1:1 to corpus entries.
- Primarily evidence-backed scenarios; clearly-labeled synthesized cases are allowed where the docs/04 hierarchy needs coverage (e.g. required-vs-required with patch), with provenance marked.
- Include a coverage matrix mapping every docs/04 resolution rule (required>nice-to-have, required-vs-required hard conflict, patch override, nice-vs-nice default, hardware-conditional, ordering/toposort, cycle error) to at least one scenario.
- Variant collision scenarios enumerated per variant at minimum: kernel-default vs kernel-opinion (CachyOS), theming collision (Garuda), repo-priority conflict (CachyOS/Chaotic-AUR).

### Variant-Profile Proposal Shape
- `arch-variants-delta.md` proposes a declarative YAML variant-profile sketch (repos, keyrings, kernel, defaults, pre-seeded opinions) as a candidate only — the final schema is drafted in Phase 1 (D17).
- "Foundation already has an opinion about this" is modeled as a pre-seeded opinions list in the profile; conflict semantics are deferred to Phase 1 with the open question recorded in `open-questions.md`.
- Variant versions: current stable CachyOS and Garuda releases at research time, versions recorded in the deliverable.

### Claude's Discretion
- Exact clone/inspection tooling and working directory layout for the research.
- Internal section ordering of each deliverable, as long as the required fields and IDs are present.
- How many candidate points to propose in omarchy-points.md (driven by evidence, not a quota).

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- None — the repo contains founding docs only (`docs/00`–`docs/11`); no code exists yet.

### Established Patterns
- Monorepo layout per docs/11: research deliverables belong under `research/` at the repo root.
- Documents in Markdown only (locked process note).

### Integration Points
- `research/` output gates Phase 1 (schema + resolver TDD harness); `research/resolver-edge-cases.md` seeds the Phase 1 test scenarios directly; `research/arch-variants-delta.md` feeds the Phase 2 Arch translator variant-profile design.

</code_context>

<specifics>
## Specific Ideas

- Source of truth: `https://github.com/basecamp/omarchy` — clone and analyze actual source; blog summaries explicitly forbidden (docs/08).
- The variant study answers: "can a speech targeting Arch be retargeted to an Arch variant unchanged?" — concretely, what would break running the Omarchy speech on CachyOS or Garuda.
- Identify variant behaviors that are pre-installed opinions in disguise (Garuda theming, CachyOS kernel choice).
- North-star framing: every cataloged Omarchy decision is something the schema must express and the Arch translator must build (stretch: via variant profile on CachyOS/Garuda with only declared, explainable differences).

</specifics>

<deferred>
## Deferred Ideas

- Phase 2 stretch validation (Omarchy speech retargeted to a variant) — recorded in docs/08; not a Phase 0 deliverable beyond knowing what it would take.
- Variant-profile conflict semantics ("foundation pre-seeded opinion vs user opinion") — decided in Phase 1 schema design, question recorded in open-questions.md.

</deferred>
