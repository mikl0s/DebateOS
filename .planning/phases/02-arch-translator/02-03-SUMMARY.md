---
phase: 02-arch-translator
plan: 03
subsystem: translator
tags: [yaml, declarative, variant-profiles, archiso, arch-translator, debateos]

# Dependency graph
requires:
  - phase: 00-omarchy-research-arch-variant-study
    provides: research/arch-variants-delta.md — verified CachyOS/Garuda repo/keyring/kernel/conflict data

provides:
  - translators/arch/profiles/vanilla-arch.yaml: Baseline north-star profile (no custom repos, linux kernel)
  - translators/arch/profiles/cachyos.yaml: CachyOS profile (repos + keyring + kernel + 9 pre-seeded opinions)
  - translators/arch/profiles/garuda.yaml: Garuda profile with 4 hard Omarchy conflicts as structured data
  - translators/arch/profiles/README.md: Full variant-profile schema documentation + ARCH-04 no-fork invariant

affects:
  - 02-arch-translator-plan-02  # variant.py in Plan 02 consumes these profiles
  - 02-arch-translator-plan-04  # Omarchy speech (examples/omarchy/) targets vanilla-arch

# Tech tracking
tech-stack:
  added: []  # Pure YAML/markdown authoring — no new code or dependencies
  patterns:
    - "ARCH-04 invariant: one generator, no per-variant code forks — variant differences are data only"
    - "keyring_install_before_repos ordering: keyring packages must be installed via standard Arch repos before custom repos are activated (Pitfall 4 mitigation)"
    - "[UNVERIFIED] tag convention: unverified data preserved in-line from delta study; consumers see the tag"
    - "conflicts_with_omarchy list: each entry names Garuda mechanism, Omarchy mechanism, and affected OM-NNN opinions"

key-files:
  created:
    - translators/arch/profiles/vanilla-arch.yaml
    - translators/arch/profiles/cachyos.yaml
    - translators/arch/profiles/garuda.yaml
    - translators/arch/profiles/README.md
  modified: []

key-decisions:
  - "vanilla-arch uses null for bootloader and filesystem — translator/speech choice, not profile-forced"
  - "cachyos.yaml includes repos_by_arch_level for v3/v4 ISA-optimised tiers as an optional extension key"
  - "garuda.yaml above_core: false for both repos — Garuda adds custom repos BELOW standard Arch repos (unlike CachyOS which adds them ABOVE)"
  - "4 hard Omarchy conflicts captured as structured data in conflicts_with_omarchy list: dracut/mkinitcpio, GRUB/limine, snapper/snapper, SDDM theme"
  - "Garuda non-gating stretch note placed at file top-level (as comment) and in open_questions — not a separate field"

requirements-completed: [ARCH-04]

# Metrics
duration: 4min
completed: 2026-06-12
---

# Phase 02 Plan 03: Declarative Variant Profiles Summary

**Three declarative variant profile YAML files (vanilla-arch, cachyos, garuda) authored from the verified delta study — ARCH-04 satisfied: one generator, no per-variant code forks; garuda's four hard Omarchy conflicts captured as structured data; all [UNVERIFIED] items tagged in-line; README documents schema and invariant**

## Performance

- **Duration:** ~4 min
- **Started:** 2026-06-12T22:37:56Z
- **Completed:** 2026-06-12T22:42:35Z
- **Tasks:** 2
- **Files created:** 4

## Accomplishments

- `vanilla-arch.yaml`: Baseline north-star profile — empty repos list, empty keyring_install_before_repos, linux/linux-headers kernel, mkinitcpio default, null bootloader/filesystem (translator/speech choice), empty pre_seeded_opinions. This is the Omarchy ARCH-02 target.
- `cachyos.yaml`: Full CachyOS profile from VERIFIED delta-study facts: [cachyos] repo with CDN77 mirror URL (VERIFIED), v3/v4 ISA-optimised tiers in repos_by_arch_level (VERIFIED), cachyos-keyring in keyring_install_before_repos (VERIFIED), linux-cachyos kernel 7.0.12 EEVDF (VERIFIED), 9 pre_seeded_opinions including all 12 sysctl params from 70-cachyos-settings.conf (VERIFIED). [UNVERIFIED] items tagged: FS/bootloader defaults, cpu_arch_level detection, cachyos-snapper-support default, NetworkManager dns.conf.
- `garuda.yaml`: Garuda profile with chaotic-aur (VERIFIED) + [garuda] ([UNVERIFIED] URL), chaotic-keyring, linux kernel (no custom kernel, VERIFIED), dracut/grub/btrfs mandatory defaults (VERIFIED), 4 hard conflicts as structured conflicts_with_omarchy entries. [UNVERIFIED] items tagged: garuda repo URL/SigLevel, garuda-keyring, @log/@pkg subvols, SDDM package name.
- `profiles/README.md`: Documents full schema (variant, repos, keyring_install_before_repos, kernel, defaults, pre_seeded_opinions, conflicts_with_omarchy), trust/security handling, [UNVERIFIED] tag convention, and ARCH-04 no-fork invariant.

## Task Commits

1. **Task 1: vanilla-arch + cachyos + README** — `ce532cb` (feat)
2. **Task 2: garuda profile with marked Omarchy conflicts** — `dfa6779` (feat)

## Files Created

- `translators/arch/profiles/vanilla-arch.yaml` — Baseline profile, 32 lines
- `translators/arch/profiles/cachyos.yaml` — CachyOS profile, 165 lines
- `translators/arch/profiles/garuda.yaml` — Garuda profile with 4 hard conflicts, 275 lines
- `translators/arch/profiles/README.md` — Schema documentation, ~120 lines

## Decisions Made

- **vanilla-arch bootloader/filesystem null:** Both are set to null with a comment explaining the translator/speech choice, not forced by the profile. Omarchy OM-099 handles limine at speech time.
- **repos_by_arch_level extension key:** Added to cachyos.yaml as an optional key for v3/v4 ISA tiers; the generator in Plan 02 will use cpu_arch_level to select the right repo set. This makes the v3/v4 data expressible in the profile without requiring any code fork.
- **Garuda above_core: false:** Unlike CachyOS (which puts custom repos ABOVE [core]), Garuda's pacman-default.conf places [chaotic-aur] AFTER standard Arch repos. This is a meaningful data difference captured per the delta study.
- **Garuda top-level note as comment:** The non-gating stretch note ("Omarchy-on-Garuda is post-v1.0") is placed as a YAML comment at the top of the file and as an open_questions entry — readable by humans without adding a mandatory schema key.

## Deviations from Plan

None — plan executed exactly as written. All four files listed in files_modified frontmatter were created. The plan specifies ≥ 4 conflicts_with_omarchy entries; garuda.yaml has exactly 4 (dracut/mkinitcpio, GRUB/limine, snapper/snapper, SDDM theme). All [UNVERIFIED] markers from the delta study are preserved in-line.

## Known Stubs

None — these are pure data files. All VERIFIED facts are copied faithfully from `research/arch-variants-delta.md`. Unverified items are clearly tagged [UNVERIFIED] rather than being stubbed with placeholder values.

## Threat Flags

| Flag | File | Description |
|------|------|-------------|
| threat_flag: sig_level_unverified | garuda.yaml | [garuda] repo SigLevel tagged [UNVERIFIED] — generator must not assume Required DatabaseOptional for this repo (T-02-04 mitigation: surface as trust_warning if sig_level unknown) |
| threat_flag: chaotic_aur_scope | garuda.yaml | [chaotic-aur] is a large AUR mirror with pre-built packages — broader attack surface than official repos; documented in README Security section |

The threat model entries T-02-04 (sig_level per repo), T-02-05 ([UNVERIFIED] data tagged), and T-02-SC (no installs in this plan) are all addressed in the profiles.

## Self-Check: PASSED

Files verified present:
- `translators/arch/profiles/vanilla-arch.yaml` — FOUND
- `translators/arch/profiles/cachyos.yaml` — FOUND
- `translators/arch/profiles/garuda.yaml` — FOUND
- `translators/arch/profiles/README.md` — FOUND

Commits verified:
- `ce532cb` — feat(02-03): add vanilla-arch and cachyos variant profiles + schema README — FOUND
- `dfa6779` — feat(02-03): add garuda variant profile with marked Omarchy hard conflicts — FOUND

Verification results:
- All 3 profiles parse with `yaml.safe_load` — OK (all 3 parse cleanly)
- `variant: vanilla-arch` present in vanilla-arch.yaml — OK
- `cachyos-keyring` present in cachyos.yaml — OK
- `cdn77.cachyos.org` URL in cachyos.yaml — OK
- `[UNVERIFIED]` tags in cachyos.yaml — OK
- `[UNVERIFIED]` tags in garuda.yaml — OK
- `conflicts_with_omarchy` list in garuda.yaml: 4 entries — OK (assert passes)
- `chaotic-keyring` in garuda.yaml — OK
- `dracut|grub|snapper|sddm` in garuda.yaml — OK (73 matches, ≥ 4 requirement satisfied)
- `keyring_install_before_repos|pre_seeded_opinions|conflicts_with_omarchy` in README — OK (7 matches, ≥ 3 required)
- `go test ./... -count=1` — all 6 packages OK (no Go regression)
