---
phase: 00-omarchy-research-arch-variant-study
plan: "03"
subsystem: research
tags: [arch-variants, cachyos, garuda, delta-study, variant-profile]
dependency_graph:
  requires: []
  provides: [research/arch-variants-delta.md]
  affects: [Phase 2 Arch translator variant-profile design, Plan 04 edge-case corpus]
tech_stack:
  added: []
  patterns: [variant-profile YAML sketch, repo-list+keyring+kernel+defaults shape]
key_files:
  created:
    - research/arch-variants-delta.md
  modified: []
decisions:
  - "CachyOS assumption A2 corrected: cachyos-snapper-support does exist (optional, not mandatory)"
  - "CachyOS uses mkinitcpio by default (dracut-cachyos is optional alternative)"
  - "Garuda uses dracut exclusively (conflicts mkinitcpio) — hard conflict with Omarchy"
  - "Garuda [garuda] repo URL tagged [UNVERIFIED]: not in cloned public GitHub mirrors"
  - "Garuda does not ship a custom kernel (uses linux or linux-zen) — A7 confirmed"
  - "CachyOS has grub-btrfs-support and cachyos-snapper-support as optional btrfs tooling"
metrics:
  duration: "~10 min"
  completed: "2026-06-12T19:07:20Z"
  tasks_completed: 2
  tasks_total: 2
  files_created: 1
  files_modified: 0
---

# Phase 0 Plan 03: Arch-Variant Delta Study Summary

**One-liner:** CachyOS + Garuda delta catalog from 6 freshly cloned repos with declarative
variant-profile YAML sketches (repos + keyring + kernel + defaults + pre_seeded_opinions),
all claims cited to pinned commits or tagged [UNVERIFIED].

## What Was Built

`research/arch-variants-delta.md` (687 lines) — the CachyOS + Garuda substitution study.

**CachyOS section:** Custom repos (`[cachyos]`, `[cachyos-v3]`, `[cachyos-v4]` — above Arch
repos), `cachyos-keyring`, linux-cachyos kernel family (7.0.12, 10 variants), sysctl/zram/
ananicy-cpp/ntsync pre-seeded opinions from CachyOS-Settings. CPU arch levels (v3/v4) verified
from docker `pacman-v3.conf` and `pacman-v4.conf`. Correction to RESEARCH.md: CachyOS has an
optional snapper support package (`cachyos-snapper-support`) — assumption A2 was incorrect.

**Garuda section:** `[chaotic-aur]` + `[garuda]` repos, `chaotic-keyring` (no garuda-keyring
package found), btrfs-mandatory stack (btrfsmaintenance timers always enabled via
garuda-common-settings), GRUB mandatory (in garuda-hooks deps), dracut mandatory (conflicts
mkinitcpio), Dr460nized theming stack, garuda-update wrapper, snapper pre-configured with
garuda template. Five HARD CONFLICTs with an Omarchy speech identified.

**Variant-profile shapes:** Declarative YAML candidate sketches for both variants covering
repos, keyring, kernel, defaults, btrfs subvolumes, and pre_seeded_opinions. No per-variant
fork proposed (D20 anti-bloat). Labeled `variant-profile` for automated gate detection.

## Tasks Completed

| Task | Description | Commit |
|------|-------------|--------|
| 1 | Clone and verify CachyOS + Garuda repos; catalog deltas from vanilla Arch | f61f821 |
| 2 | Propose declarative variant-profile YAML shape for both variants | f61f821 |

(Tasks 1 and 2 both modified `research/arch-variants-delta.md`; committed together.)

## Sources Cloned and Pinned

| Repo | Commit | Date |
|------|--------|------|
| github.com/CachyOS/CachyOS-PKGBUILDS | 860f2283 | 2026-06-12 |
| github.com/CachyOS/linux-cachyos | 39d9d125 | 2026-06-09 |
| github.com/CachyOS/CachyOS-Settings | b1aedc79 | 2026-06-03 |
| github.com/CachyOS/docker | 2f032fd8 | 2024-10-03 |
| github.com/garuda-linux/pkgbuilds | 1dc0c910 | 2026-06-12 |
| github.com/garuda-linux/garuda-tools | 433ad847 | 2021-11-27 |

## Deviations from Plan

### Research Corrections (Rule 1 — Bugs in RESEARCH.md)

**1. [Rule 1 - Bug] CachyOS snapper assumption corrected**
- **Found during:** Task 1 — reading `cachyos-snapper-support/PKGBUILD`
- **Issue:** RESEARCH.md assumption A2 stated "CachyOS does not pre-seed a custom snapper
  config" as LOW risk
- **Actual:** `cachyos-snapper-support` exists in CachyOS-PKGBUILDS and installs a snapper
  template + enables `snapper-cleanup.timer`. It is optional (not in default install group)
  but the package exists and must be treated as a potential collision
- **Fix:** Correction documented in the deliverable; snapper conflict risk noted

**2. [Rule 2 - Missing detail] CachyOS provides grub-btrfs-support**
- **Found during:** Task 1
- **Issue:** RESEARCH.md did not mention CachyOS's `grub-btrfs-support` package
- **Fix:** Documented in the deliverable; CachyOS supports both btrfs+snapper+GRUB path
  AND limine path depending on installer choice

### Unverifiable Claims Tagged [UNVERIFIED]

- `[garuda]` repo URL: not found in cloned GitHub mirrors (GitLab source requires auth)
- Garuda SDDM theme exact package: referenced but not isolated
- Btrfs subvolume layout `@log`/`@pkg`: inferred from docs, direct calamares config on GitLab
- `garuda-keyring`: no such package found; Garuda appears to rely on `chaotic-keyring` only
- CachyOS default FS/bootloader: installer presents menu, no hard default in PKGBUILDs

## Known Stubs

None — the deliverable is complete research documentation derived from cloned sources.
All placeholder sections are explicitly labeled `[UNVERIFIED]` per the plan requirements.

## Threat Flags

No new network endpoints, auth paths, file access patterns, or schema changes introduced.
This plan is research-only: read-only analysis of public GitHub repos. No cloned scripts
executed.

## Verification Gates

Both verification gates passed:

```
bash /tmp/check-variants-src.sh    → All Task 1 checks PASSED
bash /tmp/check-variants-profile.sh → All Task 2 checks PASSED
```

## Open Questions Recorded

1. CachyOS default FS and bootloader: undetermined from PKGBUILD inspection alone
2. `[garuda]` repo URL: not in cloned public mirrors (GitLab private)
3. CachyOS `fs.file-max` collision with Omarchy's `increase-fd-limit.sh`
4. Garuda + Omarchy initramfs conflict (dracut vs mkinitcpio): substantial translator challenge
5. Variant-profile conflict semantics (foundation pre-seeded vs user opinion): Phase 1 decision

## Self-Check: PASSED

- [x] research/arch-variants-delta.md exists (687 lines, >150 minimum)
- [x] Commit f61f821 exists
- [x] Both ## CachyOS and ## Garuda sections with repos/keyring/kernel/fs/bootloader/pre-seeded opinions
- [x] Omarchy pin 9cf1852525a5f7de26d3162db9d61e2f5c1d5523 in file header
- [x] 8 pinned 40-char commit hashes in the file
- [x] All variant claims cited to cloned sources or tagged [UNVERIFIED]
- [x] Declarative variant-profile YAML sketches for both variants present (no per-variant forks)
- [x] Both gate scripts exit 0
