---
phase: 00-omarchy-research-arch-variant-study
reviewed: 2026-06-12T00:00:00Z
depth: standard
files_reviewed: 6
files_reviewed_list:
  - research/omarchy-opinion-inventory.md
  - research/omarchy-points.md
  - research/schema-requirements.md
  - research/arch-variants-delta.md
  - research/resolver-edge-cases.md
  - research/open-questions.md
findings:
  critical: 2
  warning: 6
  info: 2
  total: 10
status: issues_found
fixes_applied: true
fixes_applied_at: 2026-06-12T00:00:00Z
resolved_findings:
  - CR-01
  - CR-02
  - WR-01
  - WR-02
  - WR-03
  - WR-04
  - WR-05
  - WR-06
open_findings:
  - IN-01
  - IN-02
---

# Phase 00: Code Review Report

**Reviewed:** 2026-06-12
**Depth:** standard
**Files Reviewed:** 6
**Status:** issues_found

## Summary

Reviewed six Phase 0 research deliverables against the referential integrity, contract
violation, evidence quality, and cross-file consistency dimensions specified in the review
scope. Reference contracts were drawn from `docs/04-conflict-resolution.md`,
`docs/09-decisions.md`, and `.planning/phases/00-omarchy-research-arch-variant-study/00-VALIDATION.md`.

The bulk of the deliverable is sound: 134 OM-NNN IDs are sequential and contiguous
(OM-001 through OM-134, verified), all 134 are assigned to exactly one point in
`omarchy-points.md` (no duplicates, no orphans, verified), all 27 EC-NNN entries carry
provenance tags, all 22 SR-NNN cite at least one OM-NNN or variant-evidence reference,
and no `intent:` field in the inventory contains the forbidden strings `pacman`, `AUR`,
or `mkarchiso` (Invariant 1 verified). The variant-profile YAML sketches are correctly
labeled as candidates in their section header (D17 compliant).

Two critical defects were found: a factually incorrect collision claim that conflates two
distinct Linux parameter mechanisms (`DefaultLimitNOFILE` vs `fs.file-max`), propagated
across three deliverables; and wrong SR cross-references in `open-questions.md` that will
misdirect Phase 1 schema work. Six warnings cover wrong OM-NNN ID cross-references in the
inventory (four of them), an incorrect EC count in the summary statistics, and two unverified
claims in the cross-variant comparison table that lack the required `[UNVERIFIED]` tag.

---

## Critical Issues

### CR-01: False sysctl collision — `DefaultLimitNOFILE` conflated with `fs.file-max` (propagated across 3 files)

**File:** `research/resolver-edge-cases.md`, EC-005 (lines ~128–148); `research/schema-requirements.md`, SR-003 and SR-016; `research/arch-variants-delta.md`, CachyOS pre-seeded opinions section and delta table

**Issue:** OM-038 (`install/config/increase-fd-limit.sh`) sets `DefaultLimitNOFILE` via
systemd drop-in files in `system.conf.d/` and `user.conf.d/`. This is a per-process POSIX
resource limit (`RLIMIT_NOFILE`) managed by the systemd service manager — not a sysctl
kernel parameter. CachyOS `70-cachyos-settings.conf` sets `fs.file-max = 2097152`, which
is a system-wide sysctl kernel parameter in a different namespace. These are two distinct
mechanisms that do not conflict on a shared key.

EC-005 asserts: `"Omarchy's sysctl-param/increase-fd-limit opinion (OM-038), which writes
fs.file-max to a drop-in file in /etc/sysctl.d/"` — this is factually wrong. OM-038's
`translator-capability` field explicitly states: `"write systemd system.conf.d and user.conf.d
drop-in files to set DefaultLimitNOFILE"`. There is no `/etc/sysctl.d/` write; there is no
`fs.file-max` key written by OM-038.

The same error appears in:
- `schema-requirements.md` SR-003 Evidence: `"CachyOS 70-cachyos-settings.conf sets fs.file-max;
  Omarchy OM-038 also sets fs.file-max via increase-fd-limit.sh"` (wrong)
- `schema-requirements.md` SR-016 Evidence: `"OM-038 (DefaultLimitNOFILE via systemd drop-in,
  logically a sysctl-class param)"` — the "logically a sysctl-class param" reasoning is the
  source of the conflation
- `arch-variants-delta.md` delta table: `"fs.file-max direct collision with Omarchy"` and
  variant-profile YAML `conflict_with_omarchy: "fs.file-max (direct collision with increase-fd-limit.sh)"`

**Impact:** EC-005 is an evidence-backed scenario whose provenance claim is false. If used as a
TDD test case (D19), the resolver would be built to detect a collision that does not exist, or
would fail to detect a real systemd/sysctl layering issue that might actually matter.

**Fix:**
1. Correct EC-005 `Given`/`Then`/provenance to reflect the actual difference: CachyOS writes
   `fs.file-max` (global sysctl), Omarchy writes `DefaultLimitNOFILE` (per-process systemd
   ulimit). State accurately that these are **different mechanisms** that operate in different
   namespaces and do NOT produce a sysctl key collision.
2. Determine whether a real interaction exists (e.g., if a user needs both a high global
   `fs.file-max` and a high per-process `RLIMIT_NOFILE`, the two opinions are additive, not
   conflicting). If a real interaction exists, describe it accurately; if not, remove EC-005
   as a false positive or reclassify its provenance to `synthesized` with a corrected scenario.
3. Correct SR-003 evidence: remove the `"Omarchy OM-038 also sets fs.file-max"` sentence.
4. Correct SR-016 evidence: remove `"logically a sysctl-class param"` from OM-038's description.
5. Correct `arch-variants-delta.md` delta table and variant-profile YAML to remove the false
   `fs.file-max` collision claim for OM-038.

---

### CR-02: Wrong SR cross-references in `open-questions.md` — OQ-002 and OQ-003 both cite SR-007 for unrelated requirements

**File:** `research/open-questions.md`, OQ-002 (line 56, 58) and OQ-003 (line 81)

**Issue:** Two mandatory open questions cite the wrong schema requirement ID.

- **OQ-002** (execution-phase field): `"Record as schema surprise SR-007"` and `"Resolved:
  Phase 1 schema design (SR-007 partially addresses this)"`. SR-007 in `schema-requirements.md`
  is `"Translator Capability Declaration"`. The execution-phase requirement is **SR-011**
  (`"Execution Phase: First-Run vs Install-Time"`). The link from OQ-002 to SR-007 is incorrect.

- **OQ-003** (npm-global-install / runtime-tool-install): `"Resolved: Phase 1 schema design
  (SR category for runtime-tool-install; maps to SR-007 translator-capability field)"`. SR-007
  is translator capability. The runtime-tool-install requirement is **SR-010** (`"Runtime Tool
  Install (npm-global and Equivalents)"`). Claiming it "maps to SR-007" obscures the actual
  category that was created for exactly this purpose.

**Impact:** Both OQ items are marked as mandatory (PLAN.md requirements). When Phase 1 schema
authors consult open-questions.md to understand what schema requirements apply to these open
questions, the wrong SR numbers will send them to the wrong requirement. OQ-002's link to SR-007
provides no guidance on execution-phase; SR-011 contains the full requirement. OQ-003's link
similarly obscures SR-010.

**Fix:**
- OQ-002: Replace both occurrences of `SR-007` with `SR-011`. The lean sentence should read:
  `"Record as schema surprise SR-011."` and the Resolved note should reference `SR-011`.
- OQ-003: Replace `"maps to SR-007 translator-capability field"` with `"maps to SR-010
  (Runtime Tool Install) in schema-requirements.md."` The translator-capability field (SR-007)
  is how the translator announces its capability to handle the runtime tool install, but the
  schema category for the opinion itself is SR-010.

---

## Warnings

### WR-01: Four wrong OM-NNN ID cross-references in `omarchy-opinion-inventory.md`

**File:** `research/omarchy-opinion-inventory.md`

**Issue:** Four `ordering:` fields in the inventory cite incorrect OM-NNN IDs.

| Entry | Field | Claimed | Actual | Correct ID |
|-------|-------|---------|--------|------------|
| OM-010 | `ordering:` | "npm-based AI tools are separate opinions **(OM-024)**" | OM-024 is ASUS ROG daemon (`asusctl`) | Should be **OM-023** |
| OM-011 | `ordering:` | "docker service configuration is a separate opinion **(OM-058)**" | OM-058 is `powerprofilesctl-rules.sh` (power profile switching) | Should be **OM-043** |
| OM-023 | `ordering:` | "after mise-work **(OM-054)** which installs Node via mise" | OM-054 is `omarchy-ai-skill.sh` (AI skill symlinks) | Should be **OM-041** |
| OM-088 | `dependencies:` | "also adds arch-mact2 repo (see post-install/pacman.sh **(OM-098)**)" | OM-098 is `hibernation.sh` (login phase) | Should be **OM-100** |

The OM-041 / OM-054 confusion in OM-023 is the most consequential: EC-035 in
`resolver-edge-cases.md` correctly cites OM-041 for the three-hop dependency chain, meaning
the EC corpus has the right ID even though the inventory has the wrong one. Future readers
diffing the inventory against the EC tests will find an apparent discrepancy.

**Fix:** In `omarchy-opinion-inventory.md`, correct:
- OM-010 `ordering:` → change `OM-024` to `OM-023`
- OM-011 `ordering:` → change `OM-058` to `OM-043`
- OM-023 `ordering:` → change `OM-054` to `OM-041`
- OM-088 `dependencies:` → change `OM-098` to `OM-100`

---

### WR-02: EC corpus summary statistics overstate synthesized count (claims 8, actual is 5)

**File:** `research/resolver-edge-cases.md`, Summary Statistics table (line ~722)

**Issue:** The summary table footer row states `"19 evidence-backed, 8 synthesized"` but the
actual count derived from the per-class provenance column and the synthesized rationale table
is **22 evidence-backed, 5 synthesized**:

| Class | EB | Synth |
|-------|----|-------|
| Class 1 (5 total) | 5 | 0 |
| Class 2 (3 total) | 2 (EC-010, EC-011) | 1 (EC-012) |
| Class 3 (4 total) | 4 | 0 |
| Class 4 (9 total) | 5 (EC-030,031,035,037,038) | 4 (EC-032,033,034,036) |
| Class 5 (3 total) | 3 | 0 |
| Class 6 (3 total) | 3 | 0 |
| **Total** | **22** | **5** |

The synthesized rationale table at the bottom of the file lists exactly 5 synthesized ECs
(EC-012, EC-032, EC-033, EC-034, EC-036), confirming 5 synthesized and 22 evidence-backed.
The summary footer appears to have miscounted.

**Impact:** The TDD harness described in D19 will consume this corpus. If the provenance
classification is wrong, test organization may be incorrect. Synthetic tests require
different validation criteria than evidence-backed ones (the patch opinion in EC-032 is
synthesized; the conflict it exercises is real).

**Fix:** Correct the summary table total row to: `"22 evidence-backed, 5 synthesized"`.

---

### WR-03: EC coverage matrix mischaracterizes docs/04 as having "8 resolution rules"

**File:** `research/resolver-edge-cases.md`, Coverage Matrix section (line ~700)

**Issue:** The coverage matrix header states: `"maps all 8 resolution rules from
docs/04-conflict-resolution.md"`. `docs/04` contains exactly **4** numbered resolution rules
in its "Resolution hierarchy (precise rules)" section. The coverage matrix has constructed 8
rows by splitting Rule 2 into sub-cases and adding ordering, cycle-detection, and
hardware-conditional as numbered "rules" — none of which are numbered rules in `docs/04`
(they appear as separate named sections: `## Ordering`, `## Hardware-aware resolution`).

The EC file's coverage is substantively good — the relevant docs/04 behaviors ARE covered.
The problem is the framing: saying "all 8 rules" when docs/04 has 4 numbered rules is a
misstatement that could cause a Phase 1 reader to believe docs/04 has been revised.

**Fix:** Correct the coverage matrix header to: `"maps the 4 numbered resolution rules
from docs/04 and the additional behavioral sections (hardware-conditional, ordering,
cycle detection)"`. Optionally renumber the matrix rows to distinguish the 4 core rules
from the 4 derived behaviors.

---

### WR-04: `cachy-update` claim in cross-variant comparison table lacks `[UNVERIFIED]` tag

**File:** `research/arch-variants-delta.md`, Cross-Variant Comparison table, "Update wrapper" row

**Issue:** The cross-variant table lists `"cachy-update (optional notifier)"` for the CachyOS
update wrapper column. This term (`cachy-update`) does not appear anywhere in the detailed
CachyOS section above, is not cited to any cloned source, and carries no `[UNVERIFIED]` tag.
The detailed CachyOS section (which does not mention a `cachy-update` tool) establishes that
CachyOS uses standard `pacman` for updates by default, with optional tools. The cross-comparison
table introduces a specific tool name without evidence.

**Fix:** Either:
1. Add `[UNVERIFIED]` to the CachyOS "Update wrapper" cell: `"cachy-update (optional notifier) [UNVERIFIED]"`, or
2. Replace with `"pacman (standard); optional update notifier [UNVERIFIED]"` to reflect what
   was actually confirmed in the CachyOS section.

---

### WR-05: `cachyos-themes-sddm` in cross-variant comparison table lacks `[UNVERIFIED]` tag

**File:** `research/arch-variants-delta.md`, Cross-Variant Comparison table, "Theming" row

**Issue:** The cross-variant table lists `"SDDM: cachyos-themes-sddm"` as CachyOS theming.
This package name does not appear in the detailed CachyOS section above the table, is not
cited to any cloned source (`CachyOS-PKGBUILDS`, `linux-cachyos`, `CachyOS-Settings`, or
`docker`), and carries no `[UNVERIFIED]` tag. The detailed verification methodology
(verification note at top of file) explicitly requires unverified claims to be tagged.

**Fix:** Add `[UNVERIFIED]` to the CachyOS theming cell or replace with `"[UNVERIFIED — SDDM
theming not confirmed from cloned CachyOS sources]"`.

---

### WR-06: OM-023 packaging-phase placement vs config-phase dependency creates undocumented cross-phase ordering

**File:** `research/omarchy-opinion-inventory.md`, OM-023; `research/schema-requirements.md`, SR-006

**Issue:** OM-023 is classified in `PHASE 2: Packaging — Specialized` (source:
`install/packaging/npm.sh`) but its corrected dependency is OM-041 (`config/mise-work.sh`,
explicitly in the Config phase). The install pipeline order is:
`preflight → packaging → config → login → post-install → first-run`. A packaging-phase
opinion cannot execute after a config-phase opinion without a cross-phase ordering exception.

SR-006 acknowledges this anomaly by citing `"OM-023 (packaging; after OM-041/mise, not just
after all packaging)"` as evidence for the load-bearing phase-level ordering constraint
requirement. However, the inventory itself does not flag this as anomalous — it simply lists
OM-023 as `"ordering: packaging phase"` without noting the cross-phase dependency that
violates the standard phase sequence.

This is consistent with the Omarchy source having `npm.sh` in the `packaging/` directory but
actually being invoked after `mise-work.sh` in practice. The inventory should document this
exception explicitly rather than silently listing a phase that conflicts with the dependency.

**Fix:** In OM-023's `ordering:` field, note the exception explicitly:
```
ordering: packaging phase by source location; but executes after config/mise-work.sh (OM-041)
  — this is a cross-phase dependency that requires the schema to support inter-phase
  ordering constraints (see SR-006)
```
This makes the cross-phase anomaly visible to Phase 1 schema authors consulting the inventory.

---

## Info

### IN-01: Garuda `[garuda]` repo URL and keyring — `[UNVERIFIED]` correctly tagged but variant profile uses the claim in SigLevel field

**File:** `research/arch-variants-delta.md`, Garuda variant-profile YAML (lines ~577–582)

**Issue:** The Garuda variant profile YAML declares `sig_level: Required DatabaseOptional` for
the `[garuda]` repo with the inline comment `[UNVERIFIED — exact SigLevel not in cloned source]`.
This correctly uses the `[UNVERIFIED]` tag. However, OQ-010 records this as a question for
Phase 2 but does not note that the variant profile already incorporates the unverified value
as a field value rather than a placeholder. A Phase 1 schema consumer reading the YAML directly
might use `Required DatabaseOptional` as confirmed fact.

**Fix:** Consider changing the variant profile YAML value to a placeholder string rather than
an assumed value: `sig_level: "UNVERIFIED"` or `sig_level: null  # [UNVERIFIED]`.

---

### IN-02: EC-023 title claims "idempotent across variants" but scenario only tests Garuda

**File:** `research/resolver-edge-cases.md`, EC-023 (line ~299)

**Issue:** EC-023 is titled "Bluetooth service enable idempotent across variants" but the
`Given`/`When`/`Then` scenario only targets Garuda: `"The speech targets Garuda."` The title
implies cross-variant coverage (CachyOS and Garuda) but the test case only exercises one
variant. This does not affect the test's validity — the provenance is sound — but the title
is misleading for the Phase 1 test harness author who will implement the actual resolver test.

**Fix:** Rename EC-023 to `"Bluetooth service enable: no conflict on Garuda foundation"` or
expand the scenario to explicitly state that the same result holds on CachyOS (the provenance
citation already confirms `cachyos-settings.install` does not include bluetooth).

---

_Reviewed: 2026-06-12_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
