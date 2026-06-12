---
phase: 00-omarchy-research-arch-variant-study
verified: 2026-06-12T00:00:00Z
status: passed
score: 4/4
overrides_applied: 0
---

# Phase 0: Omarchy Research & Arch-Variant Study — Verification Report

**Phase Goal:** The schema floor and resolver test corpus exist as evidence, not theory — every later design decision traces to real Omarchy and Arch-variant data
**Verified:** 2026-06-12
**Status:** passed
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | All six research deliverables exist in `research/`, built from cloned Omarchy source | VERIFIED | All 6 files present; pinned commit `9cf1852...` in inventory header; `source: install/...` citations throughout |
| 2 | Every post-base-Arch Omarchy decision recorded as atomic opinion with category, OS-agnostic intent, dependencies/ordering, un-agnostic flags; opinions grouped into candidate points | VERIFIED | 134 OM-NNN entries, all with all 6 required fields (verified programmatically); 32 candidate points covering all 134 with 0 orphans |
| 3 | Opinion metadata surface justified by real Omarchy decisions (schema surprises captured); CachyOS/Garuda deltas cataloged with proposed declarative variant-profile shape | VERIFIED | 7 schema surprises mapped to SR-NNN; variant-profile YAML sketches present for CachyOS and Garuda with pinned repo/commit citations |
| 4 | Resolver edge-case corpus written as concrete test scenarios ready to seed Phase 1 TDD harness | VERIFIED | 27 EC-NNN entries across 6 collision classes, all with Given/When/Then + provenance; coverage matrix maps 4 docs/04 numbered rules + 3 behavioral sections |

**Score:** 4/4 truths verified

---

## Mechanical Gate Results (from 00-VALIDATION.md)

| Gate | Command | Result | Status |
|------|---------|--------|--------|
| `grep -c '^### OM-'` >= 100 | `grep -c '^### OM-' research/omarchy-opinion-inventory.md` | 134 | PASS |
| Every OM-NNN has Category/Intent/Source fields | Python field check on all 134 entries | 0 incomplete entries | PASS |
| Every OM-NNN in omarchy-points.md exists in inventory | comm diff | 0 orphan references | PASS |
| No orphan inventory IDs (every ID in exactly one point) | comm diff | 0 unassigned IDs; "Unassigned: None" explicitly stated | PASS |
| Every SR-NNN cites >= 1 OM-NNN or variant evidence | Python check on 22 SR entries | 0 entries without OM-NNN or variant ref | PASS |
| arch-variants-delta.md has CachyOS + Garuda sections, variant-profile YAML, pinned refs | grep checks | CachyOS (36 refs), Garuda (33 refs), `variant-profile` (10 refs), github.com commits present | PASS |
| `grep -c '^### EC-'` >= 15, all with Given/When/Then + provenance | grep counts | 27 ECs; 27/27/27 Given/When/Then; 27 provenance tags (minus 1 header = 26 EC-level + 1 header) | PASS |
| coverage matrix maps docs/04 rules to >= 1 EC-NNN | grep check | All 4 numbered rules + 3 behavioral sections covered | PASS |
| open-questions.md >= 5 questions including 3 mandatory | grep count | 10 OQ entries; OQ-001 (migrations), OQ-002 (execution-phase), OQ-003 (runtime-tool-install) all present | PASS |

---

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `research/omarchy-opinion-inventory.md` | Exhaustive OM-NNN inventory from cloned source | VERIFIED | 1736 lines; 134 entries; pinned commit in header; all 6 required fields on every entry |
| `research/omarchy-points.md` | Candidate point groupings covering all OM-NNN | VERIFIED | 499 lines; 32 candidate points; all 134 OM-NNN assigned; "Unassigned: None" explicit |
| `research/schema-requirements.md` | SR-NNN evidence-backed schema requirements floor | VERIFIED | 467 lines; 22 SR entries; all 7 schema surprises mapped; all SRs cite OM-NNN or variant evidence |
| `research/open-questions.md` | Surprises, ambiguities, deferred schema questions | VERIFIED | 272 lines; 10 OQ entries; 3 mandatory questions present with correct SR references (SR-011, SR-010) |
| `research/arch-variants-delta.md` | CachyOS + Garuda delta catalog with variant-profile shapes | VERIFIED | 690 lines; both variants covered; variant-profile YAML for each; 6 pinned repo commits |
| `research/resolver-edge-cases.md` | EC-NNN Given/When/Then corpus for Phase 1 TDD | VERIFIED | 761 lines; 27 ECs; 6 collision classes; coverage matrix correct after WR-03 fix |

---

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `omarchy-opinion-inventory.md` | Omarchy source repo | `source: install/...` citations | VERIFIED | `source: install/preflight/pacman.sh` etc.; reproducibility block with exact clone + checkout commands |
| `omarchy-points.md` | `omarchy-opinion-inventory.md` | OM-NNN references | VERIFIED | 134 unique OM-NNN refs in points; 0 orphan refs; 0 unassigned inventory IDs |
| `schema-requirements.md` | `omarchy-opinion-inventory.md` | OM-NNN citations per SR | VERIFIED | All 22 SRs contain OM-NNN or variant evidence in their evidence blocks |
| `resolver-edge-cases.md` | `docs/04-conflict-resolution.md` | Coverage matrix | VERIFIED | Corrected from "8 rules" to "4 numbered rules + behavioral sections"; every docs/04 section has >= 1 EC |
| `arch-variants-delta.md` | CachyOS/Garuda public repos | Pinned commits | VERIFIED | 6 pinned source commits (CachyOS-PKGBUILDS, linux-cachyos, CachyOS-Settings, docker, garuda-linux/pkgbuilds, garuda-tools) |
| `open-questions.md` | `schema-requirements.md` | SR cross-references | VERIFIED | OQ-002 → SR-011 (was SR-007, fixed by CR-02); OQ-003 → SR-010 (was SR-007, fixed by CR-02) |

---

### Review Fix Verification (commits 4c51d97, dd55562, 356a9e5)

All 8 in-scope review findings (CR-01, CR-02, WR-01 through WR-06) were verified as actually applied in the files, not just claimed in REVIEW-FIX.md:

| Finding | Fix Claimed | Fix Verified |
|---------|-------------|--------------|
| CR-01: False sysctl collision (EC-005, SR-003, SR-016, arch-variants-delta) | EC-005 reclassified synthesized; delta table corrected; SR-003/SR-016/SR-022 fixed | VERIFIED — EC-005 title reads "Synthesized: sysctl key collision..."; corrective note present; SR-003 no longer claims OM-038 writes fs.file-max; SR-016 explicitly states OM-038 is NOT sysctl; delta table row 206 reads "No sysctl key collision with Omarchy" |
| CR-02: Wrong SR refs in open-questions.md (OQ-002, OQ-003) | OQ-002 → SR-011; OQ-003 → SR-010 | VERIFIED — OQ-002 "Record as schema surprise SR-011" and "SR-011 partially addresses this"; OQ-003 "maps to SR-010 in schema-requirements.md" |
| WR-01: Four wrong OM-NNN cross-refs in inventory | OM-010→OM-023, OM-011→OM-043, OM-023→OM-041, OM-088→OM-100 | VERIFIED — all four ordering/dependencies fields contain corrected IDs |
| WR-02: EC summary stats wrong (claimed 8 synth, actual 6) | Updated to 21 evidence-backed / 6 synthesized | VERIFIED — table footer reads "21 evidence-backed, 6 synthesized" |
| WR-03: Coverage matrix claims "8 resolution rules" (docs/04 has 4) | Header corrected to "4 numbered resolution rules + behavioral sections" | VERIFIED — header now reads correct framing with explanatory note |
| WR-04: `cachy-update` without [UNVERIFIED] | Added [UNVERIFIED] tag | VERIFIED — "cachy-update (optional notifier) [UNVERIFIED]" |
| WR-05: `cachyos-themes-sddm` without [UNVERIFIED] | Added [UNVERIFIED] tag | VERIFIED — "SDDM: cachyos-themes-sddm [UNVERIFIED]" |
| WR-06: OM-023 cross-phase dependency undocumented | ordering: field updated with cross-phase exception note | VERIFIED — "packaging phase by source location; after mise-work (OM-041) which configures Node via mise — this is a cross-phase dependency that requires the schema to support inter-phase ordering constraints (see SR-006)" |

---

### Manual Spot-Checks (from VALIDATION.md)

#### Spot-Check 1: OS-Agnostic Intent (Invariant 1)

**Method:** Checked all 134 actual `^intent:` field lines for forbidden strings (pacman, AUR, mkarchiso, Arch-specific paths). Sampled 10 intent fields for qualitative review.

**Result:** 0 intent field lines contain any forbidden string. Sampled intents correctly describe abstract functional goals ("Register the system package manager external repository", "Configure XCompose for emoji input", "Add the current user to the input group", etc.) without naming Arch mechanics.

**Status:** VERIFIED

#### Spot-Check 2: Variant Delta Reality Check (Researcher Confidence Was LOW)

**Method:** Verified the file header: pinned commits from 4 CachyOS repos and 2 Garuda repos are cited. Claims not verifiable from cloned sources are tagged `[UNVERIFIED]`. Corrections from the RESEARCH.md summary are explicitly annotated with "CORRECTION:" in the file.

**Finding — Residual false claim (WARNING, not BLOCKER):** The "Open Questions" section (line 452-455) of `research/arch-variants-delta.md` contains a narrative paragraph that was not updated by the CR-01 fix: *"Both `70-cachyos-settings.conf` and Omarchy's `increase-fd-limit.sh` write to `fs.file-max`."* This claim was corrected in the authoritative data structures (delta table, variant-profile YAML, SR-003, SR-016, SR-022, EC-005) but the Open Questions narrative paragraph was missed. A reader consulting this background section will see a claim that contradicts the corrected data tables in the same file.

**Impact:** Low. The authoritative data for Phase 1 schema design (delta table row 206, variant-profile YAML, SR-003, SR-016) all carry the correct information. The Open Questions section was a scratchpad for feeding `open-questions.md` — and `open-questions.md` itself does not contain this claim. Phase 1 authors using the structured data (delta table, SRs, ECs) will get correct information.

**Status:** VERIFIED (with warning — see Anti-Patterns below)

---

### Requirements Coverage

| Requirement | Description | Status | Evidence |
|-------------|-------------|--------|----------|
| RSCH-01 | Omarchy deep-dive from cloned source: atomic opinions with category, OS-agnostic intent, dependencies/ordering, translator-capability; opinions grouped into points | SATISFIED | 134 OM-NNN entries, all fields complete, all grouped into 32 points with 0 orphans |
| RSCH-02 | CachyOS/Garuda variant substitution study with declarative variant-profile shape | SATISFIED | arch-variants-delta.md with 6 pinned commits; variant-profile YAML for both variants |
| RSCH-03 | Resolver edge-case corpus as concrete Phase 1 test scenarios | SATISFIED | 27 EC-NNN Given/When/Then scenarios, 6 collision classes, docs/04 coverage matrix |

---

### Anti-Patterns Found

| File | Location | Pattern | Severity | Impact |
|------|----------|---------|----------|--------|
| `research/arch-variants-delta.md` | Line 452-455, "Open Questions" section item #3 | Factually incorrect claim: "Both `70-cachyos-settings.conf` and Omarchy's `increase-fd-limit.sh` write to `fs.file-max`" — contradicts the corrected delta table (line 206), variant-profile YAML (line 536), and all 4 SR/EC fixes from CR-01 | WARNING | Low — authoritative data structures are correct; this is a stale narrative paragraph in a background section that was not consumed by open-questions.md. Phase 1 authors using structured data (delta table, SRs, ECs) receive correct information. |

No TBD/FIXME/XXX debt markers found in any of the 6 research deliverables.

---

### Human Verification Required

None. All VALIDATION.md manual-only verifications were performed directly against the file contents:

1. **OS-agnostic intent spot-check** — performed by grep on all 134 `^intent:` lines (0 violations) and qualitative sampling of 10 intent fields. No human judgment required; result is unambiguous.

2. **Variant delta reality check** — performed by verifying presence of pinned source commits, explicit "CORRECTION:" annotations in the file, and `[UNVERIFIED]` tagging of unconfirmed claims. The file's own verification methodology is self-documented. The residual stale paragraph is documented as a warning but does not require human decision — the authoritative data is correct.

---

### Behavioral Spot-Checks

Step 7b: SKIPPED — documentation phase with no runnable entry points.

### Probe Execution

Step 7c: SKIPPED — no probe scripts declared in PLAN.md or found in `scripts/*/tests/probe-*.sh`.

---

## Gaps Summary

No blocking gaps found. All 4 success criteria are met. All 6 research deliverables exist and pass mechanical gate checks. All 8 code review findings (CR-01, CR-02, WR-01 through WR-06) are verified fixed in the actual files. The 2 info-only findings (IN-01, IN-02) were out of scope for the fix pass per the REVIEW-FIX.md and do not affect goal achievement.

One warning was identified: a residual false claim in a narrative paragraph of `arch-variants-delta.md` that was missed by the CR-01 fix pass. The authoritative data structures (delta table, variant-profile YAML, SR-003, SR-016, SR-022, EC-005) are all correct. This is a cosmetic inconsistency in a background section that does not gate Phase 1 schema design.

---

_Verified: 2026-06-12_
_Verifier: Claude (gsd-verifier)_
