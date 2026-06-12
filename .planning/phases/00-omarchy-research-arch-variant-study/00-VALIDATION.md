---
phase: 0
slug: omarchy-research-arch-variant-study
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-06-12
---

# Phase 0 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | None — documentation phase; validation via shell assertions (grep/diff/wc) |
| **Config file** | none — Wave 0 not needed |
| **Quick run command** | per-deliverable grep/diff checks (see map below) |
| **Full suite command** | all completeness checks below run in sequence |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run the completeness check for the deliverable just written
- **After every plan wave:** Run all checks for deliverables written so far
- **Before `/gsd-verify-work`:** All six deliverable checks must pass
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 00-01-XX | 01 | 1 | RSCH-01 | — | N/A | shell | Every script in omarchy `install/*/all.sh` run order has ≥1 `OM-NNN` entry citing it: diff of expected-scripts list vs cited-scripts list is empty | ✅ | ⬜ pending |
| 00-01-XX | 01 | 1 | RSCH-01 | — | N/A | shell | `grep -c '^### OM-' research/omarchy-opinion-inventory.md` ≥ 100; every entry has Category/Intent/Source fields | ✅ | ⬜ pending |
| 00-02-XX | 02 | 2 | RSCH-01 | — | N/A | shell | Every OM-NNN ID in omarchy-points.md exists in the inventory; no orphan opinions (every inventory ID appears in exactly one point or an explicit "unassigned" list) | ✅ | ⬜ pending |
| 00-02-XX | 02 | 2 | RSCH-01 | — | N/A | shell | `research/schema-requirements.md`: every requirement row cites ≥1 OM-NNN or variant evidence ID | ✅ | ⬜ pending |
| 00-03-XX | 03 | 2 | RSCH-02 | — | N/A | shell | `research/arch-variants-delta.md` contains sections for CachyOS and Garuda, a variant-profile YAML sketch, and pinned source refs (repo URLs + commits/dates) | ✅ | ⬜ pending |
| 00-04-XX | 04 | 2 | RSCH-03 | — | N/A | shell | `grep -c '^### EC-' research/resolver-edge-cases.md` ≥ 15; coverage matrix maps all 8 docs/04 rules to ≥1 EC-NNN; every EC has Given/When/Then + provenance tag | ✅ | ⬜ pending |
| 00-04-XX | 04 | 2 | RSCH-03 | — | N/A | shell | `research/open-questions.md` lists ≥5 questions incl. migrations primitive, execution-phase field, runtime-tool-install category | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements — documentation phase, shell built-ins only.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Opinion intents are genuinely OS-agnostic (invariant 1) | RSCH-01 | Semantic judgment — grep cannot detect leaked Arch mechanics in intent prose | Spot-check 10 random OM entries: intent field must not mention pacman, AUR, mkarchiso, or Arch paths |
| Variant deltas reflect current reality (researcher confidence LOW on variants) | RSCH-02 | Requires checking live CachyOS/Garuda repos during execution | Executor must verify claims against actual cloned variant repos, not just the RESEARCH.md summary |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references (none missing)
- [x] No watch-mode flags
- [x] Feedback latency < 10s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-06-12 (autonomous run)
