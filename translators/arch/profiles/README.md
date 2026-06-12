# Arch Translator Variant Profiles

This directory contains declarative YAML profiles that parameterise the Arch translator for
different Arch-based distributions. Each profile describes the deltas between a base Arch
installation and the target distribution.

## ARCH-04 Invariant

**One generator. No per-variant code forks. Differences are data only.**

The generator (`translators/arch/generator.py`) loads a profile from this directory and uses
its data to adjust pacman.conf, keyring installation, kernel package selection, and conflict
reporting. There is no per-distribution branching in the generator code.

## Available Profiles

| File | Variant | Description |
|------|---------|-------------|
| `vanilla-arch.yaml` | `vanilla-arch` | Baseline — standard Arch Linux, no custom repos. North-star target for the Omarchy speech. |
| `cachyos.yaml` | `cachyos` | CachyOS — CPU-architecture-optimised packages (x86-64-v3/v4), custom kernel family (EEVDF), CachyOS repos + keyring. |
| `garuda.yaml` | `garuda` | Garuda Linux — btrfs-first defaults, GRUB + grub-btrfs, dracut initramfs, chaotic-aur repo. Hard conflicts with Omarchy documented. |

## Profile Schema

Each profile is a YAML file with the following top-level keys:

### Required Keys

```yaml
variant: string
  # Unique identifier for this variant. Matches the filename stem.
  # Example: "vanilla-arch", "cachyos", "garuda"

description: string
  # Human-readable one-line summary of the distribution.

repos: list[RepoEntry]
  # Ordered list of custom repositories to add to pacman.conf.
  # Repos that must appear ABOVE [core]/[extra]/[multilib] set above_core: true.
  # Empty list for vanilla-arch (baseline).

keyring_install_before_repos: list[string]
  # Package names that must be installed using the standard Arch repos BEFORE
  # any custom repo is activated. Typically the distribution's keyring package.
  # The generator enforces this ordering — see research/arch-variants-delta.md Pitfall 4.
  # Empty list for vanilla-arch.

kernel:
  package: string          # Primary kernel package (e.g. "linux", "linux-cachyos")
  headers: string          # Corresponding headers package (e.g. "linux-headers")

defaults:
  initramfs: string|null   # "mkinitcpio" or "dracut" or null (translator choice)
  bootloader: string|null  # "grub", "limine", "systemd-boot", or null (translator/speech choice)
  filesystem: string|null  # "btrfs", "ext4", or null (installer/speech choice)

pre_seeded_opinions: list[PreSeededOpinion]
  # Capabilities/configurations that the distribution's base packages already express.
  # The generator uses these to detect conflicts with opinions in the resolved speech.
  # Empty list for vanilla-arch.
```

### Optional Keys

```yaml
version_at_research: string
  # Research date and source commit references.

cpu_arch_level: string
  # CachyOS only: "x86_64", "v3", or "v4" ISA level for optimised repo selection.

repos_by_arch_level: map[string, list[RepoEntry]]
  # CachyOS only: additional repos enabled for specific ISA levels.

conflicts_with_omarchy: list[OmarchyConflict]
  # Garuda only: structured list of hard conflicts between this distribution's
  # base configuration and Omarchy opinions. See Omarchy Conflict Schema below.

open_questions: list[string]
  # Notes on items that could not be verified from cloned source.

btrfs_subvolumes: list[string]
  # Garuda only: standard btrfs subvolume layout created by the installer.
```

### RepoEntry Schema

```yaml
name: string          # pacman repo name (e.g. "cachyos", "chaotic-aur")
url: string           # Server URL (with $arch/$repo placeholders where applicable)
sig_level: string     # pacman SigLevel value (e.g. "Required DatabaseOptional")
above_core: bool      # If true, repo must appear before [core] in pacman.conf
keyring: string       # (optional) keyring package name that authenticates this repo
```

### PreSeededOpinion Schema

```yaml
category: string            # Opinion category (e.g. "sysctl-param", "service-enable")
id: string                  # Unique identifier: "variant/opinion-name"
description: string         # What this pre-seeded opinion does
conflict_with_omarchy: string  # Conflict analysis (or "None identified")
keys: list[string]          # (sysctl-param only) sysctl key=value pairs
packages: list[string]      # (optional) packages delivering this pre-seeded opinion
source_file: string         # (optional) source file in the distribution's package
source_commit: string       # (optional) repo/commit reference for the verified data
```

### OmarchyConflict Schema (garuda only)

```yaml
conflicts_with_omarchy:
  - mechanism: string         # The Garuda mechanism that causes the conflict
    omarchy_mechanism: string # The Omarchy mechanism it collides with
    omarchy_opinions: list[string]  # Affected Omarchy opinion IDs (e.g. ["OM-099"])
    conflict_type: string     # "hard" (package conflict) or "direct" (config collision)
    description: string       # Human-readable explanation of the conflict
```

## [UNVERIFIED] Tag Convention

Items marked `[UNVERIFIED]` in a profile were not directly verified from the cloned upstream
source repositories. They are derived from secondary sources (documentation, build scripts,
community sources) or flagged as assumptions that require installer execution to confirm.

Consumers of these profiles (the generator, plan reviewers) must treat `[UNVERIFIED]` items
as candidates, not confirmed facts. See `research/arch-variants-delta.md` for the full
verification record.

## Trust and Security

Custom repos introduce trust boundaries. The generator handles them as follows:

1. **Keyring-first ordering** (`keyring_install_before_repos`): The generator installs all
   listed keyring packages using standard Arch repos before any custom repo is activated.
   This prevents `key "..." could not be looked up remotely` errors (Pitfall 4).

2. **sig_level surfacing**: If a repo has `sig_level: Never` or `sig_level: Optional TrustAll`,
   the generator adds a `TrustWarning` to the build manifest (see `manifest.py`). This is
   surfaced as a comment in the generated `pacman.conf` and logged at composition time (T-02-04).

3. **[UNVERIFIED] sig_levels** are noted in the profile. The generator must not assume
   `Required DatabaseOptional` for repos where the sig_level is unverified.

## Adding a New Profile

1. Create `translators/arch/profiles/<variant-name>.yaml`.
2. Populate all required keys. Mark any unverified items `[UNVERIFIED]`.
3. Add the variant to the table in this README.
4. Add tests in `translators/arch/tests/test_variant.py` covering:
   - Profile loads with `yaml.safe_load` (no parse errors).
   - `variant` field matches the filename stem.
   - `keyring_install_before_repos` ordering is enforced by the generator.
   - Any `conflicts_with_omarchy` entries are correctly structured.
5. Run `python3 -m pytest translators/arch/tests/ -x -q` to verify.

## Omarchy-on-Variant Stretch Criterion

Full ISO validation for CachyOS and Garuda (SC-5) is a **non-gating stretch** criterion,
deferred to post-v1.0. The profiles and their conflict markers are Phase 2 deliverables;
full variant ISO boot validation is not. Omarchy-on-vanilla-arch (ARCH-02 north-star) is
the Phase 2 gate.
