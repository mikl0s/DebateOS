# DebateOS Translator Ownership Model (COMM-01)

**Status:** v1.0  
**Defined:** Phase 4 (Debian Translator)  
**Source:** 04-CONTEXT.md §Ownership Model Docs, 04-05-PLAN.md Task 3

---

## Core Principle

> **Distributions own their translators. Curators own their points and speeches.**

The schema and resolver are project-owned and foundation-agnostic — they describe *what* users want (opinions, points, speeches) without embedding assumptions about *how* any distribution implements those wants. This separation is Invariant 1 of the DebateOS design (see `docs/09-decisions.md`).

---

## Ownership Tiers

### 1. Project-Owned (Foundation-Neutral)

The following components are owned by the DebateOS project and must never carry distribution-specific assumptions:

| Component | Description |
|-----------|-------------|
| `schemas/` | Opinion/Point/Speech YAML schemas — describe intent, never mechanics |
| `resolver/` | Go resolver — conflict resolution, ordering, hardware evaluation |
| `translators/common/` | Shared Python utilities (manifest building, contract loading, first-run units) |
| `cli/` | Go CLI — compose/validate/build/pane subcommands; dispatch is data-driven via `foundationRegistry` |
| `examples/dual-foundation/` | Foundation-neutral representative speech proving DEB-02 |

**Invariant:** A schema change that embeds Arch, Debian, or any foundation-specific assumption is a regression. The DEB-03 audit (`docs/arch-leak-audit.md`) documents the single genuine leak that was found and fixed.

### 2. Distributions Own Their Translators

Distributions (Ubuntu, Arch Linux, Fedora, etc.) are invited to **own their translators** directly:

- **The Ubuntu community owns `translators/ubuntu/`** (post-v1.0 stretch; profile structure is ready in `translators/debian/profiles/`)
- **The Arch Linux community owns `translators/arch/`** (Arch-specific mechanics: pacman, mkinitcpio, limine, AUR)
- **Debian stable is maintained by the DebateOS project** as the reference second foundation (`translators/debian/`)

Distribution ownership means:
- The distribution decides which capability tokens to declare in `capabilities.json`
- The distribution decides how to implement each token (apt vs pacman vs dnf)
- The distribution may add variant profiles (ubuntu.yaml, cachyos.yaml) without forking the generator
- The distribution has final say on the translator's mechanics (Invariant 1)

### 3. Curators Own Their Points and Speeches

Opinion curators (individuals, teams, or communities) own their points and speeches:

- Points and speeches are plain YAML in the curator's Git repository
- Versioning, forking, PRs, and attribution come free from Git/GitHub
- A curator may publish a point set for any foundation — the translator's capability gate ensures composition-time clarity when a point targets only some foundations
- The project owns `examples/omarchy/` and `examples/dual-foundation/` as reference content (CC0-1.0)

---

## Community Model

**Community PRs are welcome** in the following areas:

| Area | Examples |
|------|---------|
| New translator | Fedora (`translators/fedora/`), NixOS (`translators/nixos/`) |
| Variant profiles | Ubuntu on the Debian translator (`profiles/ubuntu.yaml`) |
| New capability tokens | After implementing the handler and adding tests |
| New example speeches | Foundation-neutral content in `examples/` (CC0 preferred) |
| Shared core improvements | `translators/common/` utilities; must stay foundation-neutral |

**The bar for merging a new translator:** Implement the entrypoint contract (below) + declare capabilities + provide variant profile(s) + pass the full test suite.

---

## How to Add a New Translator

Adding support for a new foundation (e.g., Fedora) requires four steps:

### Step 1: Implement the Argv-Stable Entrypoint Contract

```bash
translate <resolved.json> --opinions <dir> [--profile <name>] [--out <dir>]
```

This argv is **FROZEN** and must be satisfied by every translator. The Phase 3 CLI invokes translators as subprocesses using exactly this contract.

| Argument | Required | Default | Description |
|----------|----------|---------|-------------|
| `<resolved.json>` | yes | — | Path to the ResolvedSpeech JSON (Phase 1 resolver output) |
| `--opinions <dir>` | yes | — | Directory of YAML opinion files |
| `--profile <name>` | no | translator-default | Variant profile name |
| `--out <dir>` | no | `./<foundation>-profile` | Output directory for the generated profile tree |

**Exit codes:** 0 on success, 1 on error (capability gate failure, missing file, etc.).

Reference implementations:
- `translators/arch/translate` (Arch Linux; wraps mkarchiso)
- `translators/debian/translate` (Debian; wraps live-build)

### Step 2: Declare `capabilities.json`

Create `translators/<foundation>/capabilities.json` listing every `translator_capability` token your foundation can effectuate. The capability gate fires **before any file I/O** — undeclared required tokens produce a `CapabilityError` at composition time, never a silent install-time failure (SC-3 / Invariant 2).

```json
[
  "install-packages",
  "deploy-config-file-tree",
  "enable-systemd-service",
  "write-sysctl-drop-in",
  "add-user-to-group"
]
```

- Include only tokens you have implemented.
- Omit Arch-specific tokens (mkinitcpio, limine, pacman-AUR) if your foundation doesn't use them — this is correct and expected, not a deficiency.
- See `docs/arch-leak-audit.md` for the full list of which tokens are translator-owned vs. schema-owned.
- The dual-foundation proof tokens (`install-packages`, `deploy-config-file-tree`, `enable-systemd-service`, `write-sysctl-drop-in`, `add-user-to-group`) should be declared by every translator that can implement them — they are the foundation-neutral core.

### Step 3: Provide Variant Profiles

Create at least one YAML profile in `translators/<foundation>/profiles/<name>.yaml`. Profiles parameterize the generator without forking code (ARCH-04 / Invariant applied across all foundations):

```yaml
# translators/fedora/profiles/fedora.yaml
variant: fedora
description: "Fedora Workstation — dnf, systemd-boot, btrfs default."
repos: []
kernel:
  package: kernel
  headers: kernel-headers
defaults:
  initramfs: dracut
  bootloader: systemd-boot
  filesystem: btrfs
pre_seeded_opinions: []
```

The generator must be data-driven by the profile YAML — **no per-variant code branches**. Variant differences live in the YAML; the generator code must not branch on variant name.

### Step 4: Register the Foundation in `cli/build/build.go`

Add an entry to the `foundationRegistry` map in `cli/build/build.go`:

```go
var foundationRegistry = map[string]foundationConfig{
    "arch":   {"translators/arch/translate",    "arch-profile",   "vanilla-arch"},
    "debian": {"translators/debian/translate",  "debian-profile", "debian"},
    // Add here:
    "fedora": {"translators/fedora/translate",  "fedora-profile", "fedora"},
}
```

The registry maps a speech's `foundation:` field to the correct translator binary. This is the only place in the CLI that needs updating for a new foundation.

### Step 5: Reuse `translators/common/`

The `translators/common/` package provides foundation-neutral utilities:

| Module | Provides |
|--------|---------|
| `common/contract.py` | `load_resolved_speech()`, `load_opinion_bodies()` |
| `common/manifest.py` | `BuildManifest`, `derive_source_date_epoch()` |
| `common/firstrun.py` | `render_firstrun_unit()` — systemd user oneshot pattern |

Import via thin shims in your translator directory (e.g. `translators/fedora/manifest.py` re-exports from `translators.common.manifest`). This keeps bare-name imports working in tests while sharing a single source of truth.

---

## Boundary: What Translators Own vs. What the Schema Owns

This boundary is Invariant 1. When in doubt, ask: "Does this belong in the intent (schema) or the mechanic (translator)?"

| Translator Owns | Schema Owns |
|----------------|-------------|
| `apt install` vs `pacman -S` vs `dnf install` | `packages: [git, curl, vim]` (upstream upstream names) |
| `lb build` vs `mkarchiso` vs `lorax` | Output: a bootable installer |
| `grub2` vs `limine` vs `systemd-boot` config | `bootloader.name` + `bootloader.timeout` (intent) |
| `initramfs-tools` vs `mkinitcpio` | `initramfs: <tool>` in profile YAML (translator-local) |
| `[trusted=yes]` vs `[signed-by=...]` | `sig_level` enum: Required/OptionalTrustAll/Never (intent) |
| preseed vs kickstart vs ignition | Installer automation is translator-owned |
| Capability token set | None — capability tokens are translator-declared |

**Corollary:** A schema PR that adds a new field referencing a specific package manager, bootloader, or init system is a bug. Surface it in a translator's `capabilities.json` and `profile.py` instead.

---

## Reference Implementations

| Foundation | Translator | Status | Reference Speech |
|------------|-----------|--------|-----------------|
| Arch Linux | `translators/arch/` | Complete (Phase 2) | `examples/omarchy/` (Omarchy north star) |
| Debian stable | `translators/debian/` | Complete (Phase 4) | `examples/dual-foundation/` |
| Ubuntu (Debian family) | `translators/debian/profiles/ubuntu.yaml` | Post-v1.0 (profile shape ready) | — |
| Fedora | `translators/fedora/` | Post-v1.0 community contribution | — |

**Cross-reference:** `docs/arch-leak-audit.md` documents which capabilities are translator-owned (Findings 3 & 4) vs. schema-owned, and the single genuine infrastructure leak that was fixed (Finding 5: `build.go` dispatch).

---

*COMM-01 requirement: docs/ownership-model.md documents the translator ownership model.*  
*Phase 4 deliverable. Updated: 2026-06-13.*
