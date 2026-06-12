---
phase: 00-omarchy-research-arch-variant-study
plan: 01
subsystem: research
tags: [omarchy, opinion-inventory, arch-linux, install-pipeline, hardware-conditional, theming]

requires: []
provides:
  - "research/omarchy-opinion-inventory.md — 134 atomic OM-NNN opinion entries covering the full Omarchy install pipeline"
  - "OM-NNN ID scheme (OM-001..OM-134, walk-order assigned)"
  - "Fixed per-entry field set: category, intent, source, dependencies, ordering, translator-capability (+ condition for hardware, execution-phase for first-run)"
  - "Coverage notes / unmapped scripts section feeding Plan 04 open-questions"
  - "7 schema surprises identified with evidence"
affects:
  - 00-02 (omarchy-points.md groups these OM-NNN entries into candidate points)
  - 00-03 (schema-requirements.md derives SR-NNN requirements from OM-NNN evidence)
  - 00-04 (open-questions.md uses coverage notes and schema surprises from this inventory)
  - 01-schema (Phase 1 schema must express every category type and field found here)
  - 02-arch-translator (translator-capability fields drive Arch translator requirements)

tech-stack:
  added: []
  patterns:
    - "OM-NNN stable ID scheme: zero-padded 3-digit, sequential in pipeline walk order"
    - "Per-entry field set: category/intent/source/dependencies/ordering/translator-capability"
    - "intent: must be OS-agnostic — no distro mechanics (invariant 1 enforced)"
    - "Hardware-conditional entries carry condition: field with DMI/PCI predicate"
    - "First-run deferred entries carry execution-phase: first-run"

key-files:
  created:
    - research/omarchy-opinion-inventory.md
    - research/ (directory)
  modified: []

key-decisions:
  - "Base package list (155 packages) split into 12 logical package-install opinions rather than one atomic entry — maximizes composability for Plan 02 point grouping"
  - "21 themes cataloged as individual theming opinions (OM-114..OM-134) since themes are independently selectable by users"
  - "Migrations (313 total) sampled but not inventoried — pattern documented in Coverage Notes for Plan 04 open-questions per Pitfall 5"
  - "First-run scripts (13) inventoried as OM-101..OM-113 with execution-phase: first-run even though they have no all.sh orchestrator"
  - "install-voxtype.hook included as OM-113 despite .hook extension — it is an executable post-install notification script"

patterns-established:
  - "Every OM-NNN intent must pass the OS-agnostic check: no pacman, AUR, mkarchiso, or Arch path references"
  - "Hardware conditions use compound predicates (omarchy-hw-* AND battery-present AND cpu model range)"
  - "Custom repos are first-class opinions with SigLevel metadata (trust level is schema-relevant)"
  - "npm-global-install is a distinct category from package-install — cross-tool-manager installs are separate concerns"

requirements-completed: [RSCH-01]

duration: 10min
completed: 2026-06-12
---

# Phase 0 Plan 01: Omarchy Opinion Inventory Summary

**134 atomic OM-NNN opinions covering all 6 Omarchy install phases at commit 9cf1852, with category, OS-agnostic intent, source citations, hardware conditions, and 7 identified schema surprises**

## Performance

- **Duration:** ~10 min
- **Started:** 2026-06-12T18:45:05Z
- **Completed:** 2026-06-12T18:54:58Z
- **Tasks:** 2
- **Files created:** 1 (research/omarchy-opinion-inventory.md, 1736 lines)

## Accomplishments

- Cloned and verified Omarchy at pinned commit 9cf1852525a5f7de26d3162db9d61e2f5c1d5523 (version 4.0.0.alpha)
- Created research/ directory and seeded omarchy-opinion-inventory.md with reproducible header + ID scheme + field set definition
- Walked the full install pipeline in the prescribed 10-step order, reading every referenced script's actual body
- Produced 134 atomic OM-NNN entries: 5 preflight, 22 packaging, 42 config general, 26 config hardware, 5 login, 1 post-install, 13 first-run, 21 themes
- All 89 scripts cited in all.sh files have source: citations; all 13 first-run scripts (no all.sh) also inventoried
- 38 hardware-conditional entries with condition: predicates; 15 first-run entries with execution-phase: first-run
- Both custom-repo opinions documented: [omarchy] Optional TrustAll and arch-mact2 SigLevel = Never
- Zero invariant-1 violations: no pacman/AUR/mkarchiso in any intent: field
- Identified 7 schema surprises requiring schema extension beyond docs/08 floor

## Task Commits

1. **Task 1 + Task 2: Re-clone pinned Omarchy source and walk full install pipeline** — `2693211` (docs)

## Files Created/Modified

- `/home/mikkel/repos/DebateOS/research/omarchy-opinion-inventory.md` — 134 atomic OM-NNN opinion entries, pipeline-order walk, hardware conditions, first-run metadata, schema surprises, coverage notes

## Decisions Made

- **Base package grouping:** 155 packages split into 12 logical groups (compositor, terminal, browser, dev tools, AI tools, Docker, media, productivity, desktop shell, security, development support, system utilities) rather than a single OM entry. Rationale: composability — a curator should be able to assemble a minimal Hyprland speech without pulling in spotify/signal/libreoffice.
- **Themes as individual opinions:** Each of the 21 themes is a separate theming opinion (OM-114..OM-134) because users can select any theme at runtime; they are independently selectable atomic opinions.
- **Migrations not inventoried:** 313 migration scripts follow a timestamped-delta pattern representing opinion evolution. They are schema-surprising enough to deserve their own open-question in Plan 04, not OM entries.
- **install-voxtype.hook as OM-113:** Despite the .hook extension, this file is an executable notification script that runs in the first-run phase; it is an opinion (a deferred install-offer pattern).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed invariant-1 violation in OM-049 intent field**
- **Found during:** Post-write acceptance criteria check
- **Issue:** intent: for OM-049 (walker-elephant.sh) mentioned "pacman hook" violating the OS-agnostic invariant
- **Fix:** Replaced "pacman hook" with "package-manager post-upgrade hook" in the intent field
- **Files modified:** research/omarchy-opinion-inventory.md
- **Committed in:** 2693211 (same task commit — fix applied before final commit)

**2. [Rule 1 - Bug] Fixed invariant-1 violation in OM-100 intent field**
- **Found during:** Post-write acceptance criteria check
- **Issue:** intent: for OM-100 (post-install/pacman.sh) mentioned "pacman.conf" violating the OS-agnostic invariant
- **Fix:** Replaced "pacman.conf" with "package manager configuration" in the intent field
- **Files modified:** research/omarchy-opinion-inventory.md
- **Committed in:** 2693211 (same task commit)

**3. [Rule 2 - Missing Critical] Added Optional TrustAll to OM-001**
- **Found during:** Acceptance criteria check — grep for "Optional TrustAll" returned no results
- **Issue:** OM-001 documented the [omarchy] repo but did not record the SigLevel = Optional TrustAll trust metadata; the plan requires both custom-repo opinions present with their trust level captured (Security Domain findings, threat T-00-SIG)
- **Fix:** Added trust level to OM-001 intent and translator-capability fields
- **Files modified:** research/omarchy-opinion-inventory.md
- **Committed in:** 2693211 (same task commit)

---

**Total deviations:** 3 auto-fixed (2 invariant-1 bugs, 1 missing critical security metadata)
**Impact on plan:** All fixes required for correctness and compliance with CONTEXT.md invariant 1 and threat model T-00-SIG. No scope creep.

## Issues Encountered

- Coverage diff appeared to fail initially because expected-scripts.txt used `$OMARCHY_INSTALL/` prefix while the inventory used `install/` path prefix. After normalizing the prefix, all 89 all.sh-referenced scripts confirmed covered. The additional 13 entries in the covered set are the first-run scripts (expected by design).

## Threat Flags

| Flag | File | Description |
|------|------|-------------|
| threat_flag: unsigned-repo | research/omarchy-opinion-inventory.md (OM-100) | arch-mact2 repo with SigLevel = Never documented as opinion metadata; no repo added to this machine |
| threat_flag: optional-trust-repo | research/omarchy-opinion-inventory.md (OM-001) | [omarchy] repo with SigLevel = Optional TrustAll documented as opinion metadata |

## User Setup Required

None — Phase 0 is a read-only research phase. No external services, no credentials, no deployments.

## Next Phase Readiness

- `research/omarchy-opinion-inventory.md` is ready to be consumed by Plan 02 (omarchy-points.md grouping) and Plan 03 (schema-requirements.md derivation)
- OM-NNN IDs OM-001..OM-134 are stable and can be referenced by downstream plans
- Coverage notes section is ready to seed Plan 04 open-questions deliverable
- 7 schema surprises are documented for Plan 03 to formalize as SR-NNN requirements

---
*Phase: 00-omarchy-research-arch-variant-study*
*Completed: 2026-06-12*
