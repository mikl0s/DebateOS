---
phase: 00-omarchy-research-arch-variant-study
plan: 02
subsystem: research
tags: [omarchy, schema, opinions, points, schema-requirements, SR-NNN, OM-NNN]

# Dependency graph
requires:
  - phase: 00-omarchy-research-arch-variant-study plan 01
    provides: "OM-001..OM-134 opinion inventory from cloned Omarchy source at commit 9cf1852"
  - phase: 00-omarchy-research-arch-variant-study plan 03
    provides: "arch-variants-delta.md — CachyOS and Garuda variant evidence used for SR-009, SR-016, SR-022"
provides:
  - "research/omarchy-points.md — 32 evidence-driven candidate point groupings over all 134 OM-NNN opinions"
  - "research/schema-requirements.md — 22 SR-NNN schema floor requirements, every one evidence-backed"
  - "SR-NNN ID scheme definition (## SR-NNN format, sequential zero-padded integers)"
affects:
  - "Phase 1 schema design: every SR-NNN is a hard constraint on the Opinion/Point/Speech YAML schema"
  - "Phase 1 TDD harness: SR-005 compound hardware predicates, SR-011 execution-phase, SR-006 phase ordering define test axes"
  - "Phase 2 Arch translator: SR-007 translator capability declarations, SR-009 repo trust levels define translator interface"

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "SR-NNN requirement format: ## SR-NNN heading, evidence-backed, OS-agnostic requirement body"
    - "OM-NNN grouping convention: bullet lists under ## Point: headings, coverage enforced by check-points.sh"

key-files:
  created:
    - research/omarchy-points.md
    - research/schema-requirements.md
  modified: []

key-decisions:
  - "32 evidence-driven points chosen (not a quota); single-opinion points allowed where natural cluster is a single OM-NNN"
  - "22 SR-NNN requirements defined, exceeding the 7-surprise minimum, to cover docs/04 floor plus Omarchy-specific fields"
  - "SR-009 custom-repo trust level enumerated: Required / Required DatabaseOptional / Optional TrustAll / Never (T-00-SIG2 mitigated)"
  - "SR-005 compound hardware conditions require AND/OR/NOT combinators — simple boolean flag is insufficient"
  - "SR-006 phase-level ordering is a discrete enum plus within-phase before/after references — flat integer order is insufficient"

patterns-established:
  - "SR-NNN: ## SR-NNN heading, bold Evidence: paragraph citing OM-NNN or variant evidence, OS-agnostic requirement body"
  - "Points file: ## Point: name, one-line OS-agnostic intent, bullet member list, summary table"

requirements-completed: [RSCH-01]

# Metrics
duration: 15min
completed: 2026-06-12
---

# Phase 0 Plan 02: Point Groupings and Schema Requirements Floor Summary

**134 OM-NNN opinions grouped into 32 evidence-driven candidate points; 22 OS-agnostic SR-NNN schema requirements derived with full evidence backing including all 7 confirmed schema surprises**

## Performance

- **Duration:** ~15 min
- **Started:** 2026-06-12T19:02:00Z
- **Completed:** 2026-06-12T19:16:37Z
- **Tasks:** 2
- **Files created:** 2

## Accomplishments

- Grouped all 134 Omarchy inventory opinions into 32 natural candidate points (Repository Bootstrap, Hyprland Desktop, Developer Toolchain, AI Tooling, 28 others); check-points.sh gate passes with full OM-ID parity, no duplicates, no phantoms
- Derived 22 SR-NNN schema floor requirements from the Omarchy inventory and variant delta evidence, covering the docs/04 baseline and all 7 confirmed schema surprises; check-schema.sh gate passes with SR count 22 >= 7, 91 evidence citations, all 7 surprises present
- SR-009 explicitly enumerates the custom-repo trust level as Required / Required DatabaseOptional / Optional TrustAll / Never — mitigates threat T-00-SIG2 (unsigned repos explicit in schema floor)

## Task Commits

1. **Task 1: Group the OM-NNN inventory into candidate points** - `331c80d` (feat)
2. **Task 2: Derive the evidence-backed schema-requirements floor** - `fc50a37` (feat)

**Plan metadata:** (committed with docs commit below)

## Files Created

- `research/omarchy-points.md` — 32 candidate point groupings with OS-agnostic intents and OM-NNN member lists; all 134 inventory opinions assigned; summary table; no unassigned
- `research/schema-requirements.md` — SR-001 through SR-022; seven schema surprises as SR-005, SR-006, SR-008 to SR-012; docs/04 floor extended with evidence-backed Omarchy fields

## Decisions Made

- 32 points chosen (evidence-driven, not quota). Several single-opinion points are intentional: Terminal and Shell Toolchain (OM-007), Browser and Web Access (OM-008), Media and Creative Applications (OM-012) — these are naturally atomic bundles a user selects as a unit.
- SR-009 custom-repo trust level enumeration added beyond the plan requirement — mitigates T-00-SIG2 threat directly (trust levels must be explicit in the schema floor, not translator-internal).
- SR-016 sysctl collision detection added: per-key conflict detection across sysctl-param opinions is necessary to surface CachyOS `fs.file-max` vs Omarchy `increase-fd-limit.sh` collision at composition time.
- SR-022 speech-level foundation target declaration added: without this, variant pre-seeded opinion conflicts (Garuda dracut vs Omarchy mkinitcpio) are undetectable.

## Deviations from Plan

None — plan executed exactly as written. The gate scripts check-points.sh and check-schema.sh were written as required and both pass.

Four SR-NNN requirements (SR-013 through SR-016, SR-017 through SR-022) were added beyond the 7-surprise minimum because the inventory evidence clearly demands them (display manager, bootloader, sysctl collisions, point/speech metadata). These are additive requirements, not deviations — the plan says "expand with the seven confirmed schema surprises" as the floor, not as an exhaustive ceiling.

## Known Stubs

None — both deliverables are fully specified research documents, not implementation artifacts.

## Threat Flags

None beyond what was already in the plan's threat model. SR-009 explicitly mitigates T-00-SIG2 by requiring trust level enumeration in the schema floor.

## Next Phase Readiness

- Phase 1 can draft the YAML schema using SR-001..SR-022 as hard constraints
- The 32 candidate points in omarchy-points.md provide the composition examples Phase 1 needs for schema validation
- SR-NNN IDs are stable reference targets for Phase 1 schema fields and Phase 1 TDD test annotations
- research/schema-requirements.md satisfies the RSCH-01 grouping + schema-floor portion; Plans 01 and 02 together complete the OM-NNN deliverables for RSCH-01

---
*Phase: 00-omarchy-research-arch-variant-study*
*Completed: 2026-06-12*
