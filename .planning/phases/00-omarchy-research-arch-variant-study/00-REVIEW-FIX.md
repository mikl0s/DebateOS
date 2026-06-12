---
phase: 00-omarchy-research-arch-variant-study
fixed_at: 2026-06-12T00:00:00Z
review_path: .planning/phases/00-omarchy-research-arch-variant-study/00-REVIEW.md
iteration: 1
findings_in_scope: 8
fixed: 8
skipped: 0
status: all_fixed
---

# Phase 00: Code Review Fix Report

**Fixed at:** 2026-06-12
**Source review:** `.planning/phases/00-omarchy-research-arch-variant-study/00-REVIEW.md`
**Iteration:** 1

**Summary:**
- Findings in scope: 8 (CR-01, CR-02, WR-01 through WR-06)
- Fixed: 8
- Skipped: 0
- Info findings out of scope: IN-01, IN-02

## Fixed Issues

### CR-01: False sysctl collision — DefaultLimitNOFILE conflated with fs.file-max

**Files modified:** `research/resolver-edge-cases.md`, `research/schema-requirements.md`, `research/arch-variants-delta.md`
**Commit:** 4c51d97
**Applied fix:**

EC-005 was reclassified from `evidence-backed` to `synthesized` with a corrective note explaining
that OM-038 sets `DefaultLimitNOFILE` via systemd `system.conf.d/`/`user.conf.d/` drop-ins
(per-process RLIMIT_NOFILE), NOT `fs.file-max` via sysctl.d. The scenario was rewritten to
demonstrate a real per-key sysctl collision (two opinions writing `fs.inotify.max_user_watches`
— OM-037 as the evidence-backed base and a synthesized second opinion).

Additional corrections in the same commit:
- SR-003: Replaced false "OM-038 also sets fs.file-max" with accurate description of distinct mechanisms
- SR-016: Removed "logically a sysctl-class param" mischaracterization; added note that OM-038 belongs in `systemd-limit` category
- SR-022: Corrected "conflicts OM-038" to accurate statement about kernel vs per-process mechanisms
- arch-variants-delta.md: Corrected delta summary table sysctl row and pre-seeded opinions YAML `conflict_with_omarchy` field
- WR-02: Summary statistics updated to 21 evidence-backed / 6 synthesized (EC-005 now synthesized, added to synthesized rationale table)
- WR-03: Coverage matrix header corrected from "all 8 resolution rules" to "4 numbered resolution rules + behavioral sections"
- WR-04: Added `[UNVERIFIED]` to `cachy-update (optional notifier)` in cross-variant comparison table
- WR-05: Added `[UNVERIFIED]` to `cachyos-themes-sddm` in cross-variant comparison table

### CR-02: Wrong SR cross-references in open-questions.md — OQ-002 and OQ-003

**Files modified:** `research/open-questions.md`
**Commit:** dd55562
**Applied fix:**

- OQ-002: Both occurrences of SR-007 replaced with SR-011 (Execution Phase: First-Run vs Install-Time). The Lean sentence now reads "Record as schema surprise SR-011" and the Resolved note references SR-011.
- OQ-003: "maps to SR-007 translator-capability field" replaced with "SR-010 — Runtime Tool Install — is the schema category for runtime-tool-install; maps to SR-010 in schema-requirements.md".

### WR-01: Four wrong OM-NNN ID cross-references in omarchy-opinion-inventory.md

**Files modified:** `research/omarchy-opinion-inventory.md`, `research/schema-requirements.md`
**Commit:** 356a9e5
**Applied fix:**

All four wrong cross-references corrected:
- OM-010 `ordering:` field: `OM-024` → `OM-023` (npm-global-install AI tools, not ASUS ROG daemon)
- OM-011 `ordering:` field: `OM-058` → `OM-043` (docker service configuration, not power profile switching rules)
- OM-023 `ordering:` field: `OM-054` → `OM-041` (mise-work config, not AI skill symlinks)
- OM-088 `dependencies:` field: `OM-098` → `OM-100` (post-install pacman.sh adding arch-mact2 repo, not hibernation.sh)

### WR-02: EC corpus summary statistics overstate synthesized count

**Files modified:** `research/resolver-edge-cases.md`
**Commit:** 4c51d97 (combined with CR-01)
**Applied fix:**

Summary statistics table corrected. Due to EC-005 being reclassified as synthesized as part of CR-01,
the final counts are 21 evidence-backed / 6 synthesized (not 22/5 as projected in the review — the
review's count of 22/5 assumed EC-005 would remain evidence-backed with a corrected scenario; instead
it was reclassified). The synthesized rationale table now lists all 6 synthesized ECs: EC-005, EC-012,
EC-032, EC-033, EC-034, EC-036.

### WR-03: EC coverage matrix mischaracterizes docs/04 as having "8 resolution rules"

**Files modified:** `research/resolver-edge-cases.md`
**Commit:** 4c51d97 (combined with CR-01)
**Applied fix:**

Coverage matrix header corrected from "maps all 8 resolution rules from docs/04-conflict-resolution.md"
to "maps the 4 numbered resolution rules from docs/04 and the additional behavioral sections
(hardware-conditional, ordering, cycle detection)". Added explanatory note distinguishing the 4
numbered rules from the named behavioral sections (Ordering, Hardware-aware resolution, Cycle detection).
Column header row updated to "docs/04 Resolution Rule / Behavior".

### WR-04: cachy-update claim lacks [UNVERIFIED] tag

**Files modified:** `research/arch-variants-delta.md`
**Commit:** 4c51d97 (combined with CR-01)
**Applied fix:**

Cross-variant comparison table "Update wrapper" row: `cachy-update (optional notifier)` →
`cachy-update (optional notifier) [UNVERIFIED]`.

### WR-05: cachyos-themes-sddm claim lacks [UNVERIFIED] tag

**Files modified:** `research/arch-variants-delta.md`
**Commit:** 4c51d97 (combined with CR-01)
**Applied fix:**

Cross-variant comparison table "Theming" row: `SDDM: cachyos-themes-sddm` →
`SDDM: cachyos-themes-sddm [UNVERIFIED]`.

### WR-06: OM-023 cross-phase dependency undocumented in inventory and SR-006

**Files modified:** `research/omarchy-opinion-inventory.md`, `research/schema-requirements.md`
**Commit:** 356a9e5 (combined with WR-01)
**Applied fix:**

OM-023 `ordering:` field updated to explicitly document the cross-phase ordering exception:
"packaging phase by source location; after mise-work (OM-041) which configures Node via mise —
this is a cross-phase dependency that requires the schema to support inter-phase ordering
constraints (see SR-006)".

SR-006 updated with a dedicated "Cross-phase ordering exception (OM-023)" section explaining
this as the primary evidence that cross-phase override constraints are a load-bearing schema
requirement, and updating the Implication to note that cross-phase override constraints must
be supported in addition to within-phase ordering.

## Skipped Issues

None — all in-scope findings were fixed.

## Mechanical Gate Verification (Post-Fix)

All mechanical gates passed after fixes:

1. **OM-NNN referential integrity:** All OM-NNN IDs cited in any non-inventory file exist in
   `omarchy-opinion-inventory.md`. Verified by diff of cited vs defined ID sets — no orphan citations.

2. **EC count >= 15 with Given/When/Then + provenance:** 27 EC-NNN entries, all 27 have Given/When/Then
   triplets, all 27 have provenance tags. Verified by grep counts.

3. **Summary counts match actual counts:** 21 evidence-backed + 6 synthesized = 27 total. Verified by
   counting `provenance: evidence-backed` (22 lines; 1 is the header definition — 21 scenarios) and
   `provenance: synthesized` (7 lines; 1 is the header definition — 6 scenarios). Synthesized rationale
   table lists exactly the 6 synthesized ECs (EC-005, EC-012, EC-032, EC-033, EC-034, EC-036).

---

_Fixed: 2026-06-12_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
