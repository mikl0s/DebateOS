---
phase: 00-omarchy-research-arch-variant-study
plan: "04"
subsystem: research
tags: [edge-cases, resolver, tdd-corpus, conflict-resolution, open-questions, ec-nnn, omarchy, cachyos, garuda]
dependency_graph:
  requires: ["00-01", "00-03"]
  provides: [research/resolver-edge-cases.md, research/open-questions.md]
  affects: [Phase 1 resolver TDD harness (D19), Phase 1 schema design, Phase 2 Arch translator]
tech_stack:
  added: []
  patterns: [EC-NNN Given/When/Then corpus, docs/04 coverage matrix, provenance-tagged scenarios]
key_files:
  created:
    - research/resolver-edge-cases.md
    - research/open-questions.md
  modified: []
decisions:
  - "27 EC-NNN scenarios produced (19 evidence-backed, 8 synthesized); all 8 docs/04 rules covered"
  - "CachyOS fs.file-max sysctl collision with Omarchy OM-038 is evidence-backed hard conflict (EC-005)"
  - "Garuda dracut/mkinitcpio conflict (EC-032) is evidence-backed; patch path documented as synthesized example"
  - "variant-profile conflict semantics deferred to Phase 1 schema design (OQ-007)"
  - "migrations-as-schema-concept recorded as open question OQ-001, deferred to Phase 1/post-v1"
  - "306 runtime bin/ helpers classified as translator infrastructure, not opinions (OQ-008)"
metrics:
  duration: "~10 min"
  completed: "2026-06-12T19:25:57Z"
  tasks_completed: 2
  tasks_total: 2
  files_created: 2
  files_modified: 0
---

# Phase 0 Plan 04: Resolver Edge-Case Corpus and Open Questions Summary

**One-liner:** 27 EC-NNN Given/When/Then resolver scenarios (full docs/04 coverage, all 6 collision classes) plus 10 open questions seeding the Phase 1 TDD harness and schema design.

---

## What Was Built

### Task 1: research/resolver-edge-cases.md

The EC-NNN resolver edge-case corpus seeding the Phase 1 TDD harness (D19). Contains 27 scenarios
across 6 collision classes derived from real Omarchy/CachyOS/Garuda source evidence:

**Class 1 — Foundation pre-seeded vs user opinion (EC-001..EC-005):**
- EC-001: Garuda snapper root config vs Omarchy snapper root config (required-vs-required hard conflict)
- EC-002: Garuda GRUB bootloader vs Omarchy Limine bootloader (required-vs-required hard conflict)
- EC-003: Garuda Dr460nized SDDM theme vs Omarchy SDDM theme (required-vs-required conflict)
- EC-004: CachyOS linux-cachyos kernel vs Omarchy linux kernel (required-vs-required conflict)
- EC-005: CachyOS fs.file-max sysctl collision with Omarchy OM-038 (same key, drop-in ordering)

**Class 2 — Repo priority conflicts (EC-010..EC-012):**
- EC-010: CachyOS v3 repos must precede standard Arch repos; [omarchy] ordering
- EC-011: Garuda [garuda]+[chaotic-aur] vs [omarchy] relative priority
- EC-012: Nice-to-have custom repo dropped by required repo (synthesized)

**Class 3 — Cross-variant effectuation (EC-020..EC-023):**
- EC-020: mesa satisfied by CachyOS x86-64-v3 optimized build (no conflict)
- EC-021: linux-headers requires linux-cachyos-headers name translation on CachyOS
- EC-022: snapper idempotency on Garuda (both pre-seed and speech try to create root config)
- EC-023: Bluetooth service-enable idempotent across variants (no conflict baseline)

**Class 4 — docs/04 rule coverage (EC-030..EC-038):**
- EC-030: Required kernel drops nice-to-have DKMS opinion (required beats nice-to-have)
- EC-031: Two required kernel opinions (linux-vanilla vs linux-cachyos) — hard conflict, no patch
- EC-032: Required initramfs conflict (mkinitcpio vs dracut) resolved by patch opinion (synthesized patch)
- EC-033: Two nice-to-have terminals (foot vs ghostty) — resolver picks foot (synthesized)
- EC-034: Patch opinion overrides required-vs-required pacman.conf conflict (synthesized)
- EC-035: Three-hop toposort (OM-009 → OM-041 → OM-023) (evidence-backed)
- EC-036: Circular ordering constraint — hard cycle error (synthesized)
- EC-037: NVIDIA hardware-conditional skipped (no NVIDIA GPU declared) (evidence-backed)
- EC-038: Apple T2 hardware-conditional block applied (T2 PCI ID matches) (evidence-backed)

**Class 5 — CachyOS kernel collision (EC-040..EC-042):**
- EC-040: Vanilla linux vs pre-seeded linux-cachyos — hard conflict on CachyOS
- EC-041: Speech targets CachyOS v3; hardware is v4-capable — resolver suggests upgrade
- EC-042: CachyOS kernel + PTL kernel — multi-kernel install, no conflict

**Class 6 — Garuda theming (EC-050..EC-052):**
- EC-050: Omarchy SDDM theme vs Garuda Dr460nized SDDM theme — active theme slot collision
- EC-051: Omarchy Plymouth theme vs Garuda Dr460nized Plymouth theme — same config path
- EC-052: Garuda GRUB theme — no Omarchy counterpart, no collision (baseline)

### Task 2: research/open-questions.md

10 OQ-NNN open questions capturing surprises, ambiguities, and deferred schema questions:

- OQ-001: Migrations primitive — 313 timestamped scripts; version-pinning vs dedicated migration concept
- OQ-002: execution-phase field — first-run vs install-time; headless CI cannot run first-run opinions
- OQ-003: runtime-tool-install category — npm AI tools distinct from OS package manager
- OQ-004: Tag pinning — no git tags; commit-hash-only vs semver+commit hybrid version format
- OQ-005: Omarchy custom repo as opinion vs foundation prerequisite
- OQ-006: CachyOS CPU arch level — static declaration vs runtime hardware detection
- OQ-007: Variant-profile conflict semantics — foundation pre-seeded vs user opinion (DEFERRED to Phase 1)
- OQ-008: 306 runtime bin/ helpers — unmapped scripts from inventory Coverage Notes
- OQ-009: CachyOS default filesystem/bootloader — installer-menu dependent, not PKGBUILD-determinable
- OQ-010: Garuda [garuda] repo URL unverified — calamares config requires auth

---

## Coverage Matrix Verification

| docs/04 Resolution Rule | EC-NNN | Status |
|------------------------|--------|--------|
| Rule 1: Required beats nice-to-have | EC-012, EC-030 | Covered |
| Rule 2: Required-vs-required hard conflict | EC-001..EC-005, EC-031, EC-050, EC-051 | Covered (8 scenarios) |
| Rule 3: Required-vs-required with patch | EC-032, EC-034 | Covered |
| Rule 4: Nice-vs-nice sensible default | EC-033 | Covered |
| Rule 5: Patch overrides hierarchy | EC-032, EC-034 | Covered |
| Rule 6: Ordering/toposort | EC-035, EC-010, EC-011 | Covered |
| Rule 7: Cycle detection | EC-036 | Covered |
| Rule 8: Hardware-conditional (skip) | EC-037 | Covered |
| Rule 8: Hardware-conditional (apply) | EC-038, EC-040..EC-042 | Covered |

---

## Deviations from Plan

None — plan executed exactly as written.

---

## Commits

| Task | Commit | Files |
|------|--------|-------|
| Task 1: EC-NNN corpus | 5a62f31 | research/resolver-edge-cases.md |
| Task 2: open-questions | 196755a | research/open-questions.md |

---

## Self-Check

Self-check results:

- FOUND: research/resolver-edge-cases.md (737 lines, 27 EC-NNN scenarios)
- FOUND: research/open-questions.md (271 lines, 10 OQ-NNN questions)
- FOUND: commit 5a62f31 (Task 1)
- FOUND: commit 196755a (Task 2)

## Self-Check: PASSED
