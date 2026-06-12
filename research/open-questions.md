# Open Questions

**Omarchy pin:** `9cf1852525a5f7de26d3162db9d61e2f5c1d5523` (version 4.0.0.alpha, no git tags)
**Produced:** 2026-06-12
**Phase:** 0 — Omarchy Research & Arch-Variant Study

This document captures surprises, ambiguities, and deferred schema questions discovered
during Phase 0 research. Questions are not resolved here — each is tagged with where it
is resolved (Phase 1 schema design vs ongoing). Do NOT resolve deferred items in this document.

---

## Mandatory Questions (PLAN.md requirements)

### OQ-001: Does the schema need a migrations primitive, or is opinion version-pinning sufficient?

**Question:** Omarchy contains 313 timestamped migration scripts in `migrations/` (UNIX-timestamp
named `.sh` files). Each migration is a numbered idempotent catch-up script that updates an
existing Omarchy installation — adding packages, removing packages, modifying config files,
restarting services. They represent the opinion-evolution history of a running system.
Does the DebateOS schema need a first-class `migration` concept (a migration opinion that
carries a timestamp, an idempotency check, and a catch-up action), or is version-pinning of
opinions sufficient (i.e., pinning a specific opinion version in the speech achieves the same
upgrade guarantee)?

**Evidence:** 313 migration scripts sampled in `research/omarchy-opinion-inventory.md`
(Coverage Notes / Migrations section). Pattern: timestamped idempotent catch-up scripts
encoding the delta between opinion states over time, e.g. `omarchy-pkg-add`, `omarchy-pkg-drop`,
config file patches, service restarts.

**Lean (from RESEARCH.md):** Record as open question; defer to Phase 1. Version-pinning may be
sufficient for v1 (install-time only per D2 — no post-install reconciliation); migrations are
more relevant if DebateOS later supports live-system updates. The 313 count is large enough to
warrant a dedicated migration concept if post-install reconciliation is ever added.

**Resolved:** Phase 1 schema design (or deferred to post-v1 if install-time-only is maintained)

---

### OQ-002: What is the correct execution-phase field for opinions that require a live Wayland session?

**Question:** Omarchy's `install/first-run/` scripts (OM-101 through OM-113, 13 opinions) must
execute after the user's first login, in a live Wayland display session. On a headless CI/VM
build, these opinions cannot run because there is no display server. The schema needs an
`execution-phase` or `execution-context` field to express this distinction.
What are the valid values? Is `install-time` vs `first-run` sufficient, or are more phases
needed (e.g., `post-reboot`, `on-connect`, `manual`)?

**Evidence:** OM-101 (WiFi/notification setup, needs `notify-send`), OM-102 (GTK theme via
`gsettings`, needs running GNOME settings daemon), OM-103 (firewall setup via `ufw`, needs
passwordless sudo from OM-004), OM-104 (DNS symlink — actually could be install-time), OM-105
(GDK scale detection, needs `hyprctl monitors`). All are tagged `execution-phase: first-run`
in the inventory.

**Lean (from RESEARCH.md):** Add `execution-phase: first-run` as a metadata field. Two-value
enum (`install-time` | `first-run`) may be sufficient for v1. Record as schema surprise SR-011.

**Resolved:** Phase 1 schema design (SR-011 partially addresses this; full enum definition
belongs in Phase 1 YAML schema draft)

---

### OQ-003: Is npm-global-install a distinct schema category from OS package install, and does the schema need a generic runtime-tool-install category?

**Question:** Omarchy installs AI coding assistant tools — `codex`, `gemini`, `copilot`,
`opencode`, `playwright`, `pi`, `ghui`, `hunk` — via `npm install -g` (OM-023). These are
entirely separate from the OS package manager (`pacman`). On a Debian foundation, `npm` itself
must first be installed via `apt`. Does the schema need a `runtime-tool-install` category
that is distinct from OS packages, capturing: tool manager (npm, pip, cargo, gem), tool name,
version constraint, and global vs local scope?

**Evidence:** OM-023 (`install/packaging/npm.sh`) installs 8 AI tools via npm. These tools
have no Arch package equivalents and must be installed after `mise` configures Node.js
(OM-041 must precede OM-023 in the pipeline). The category `npm-global-install` is already
flagged as a schema surprise in the inventory.

**Lean (from RESEARCH.md):** Yes — record as schema surprise SR-NNN. A `runtime-tool-install`
category with a `manager:` field (npm | pip | cargo | gem) is cleaner than creating a separate
category for each package manager. This also handles future non-npm tools.

**Resolved:** Phase 1 schema design (SR-010 — Runtime Tool Install — is the schema category
for runtime-tool-install; maps to SR-010 in schema-requirements.md, not SR-007
translator-capability field)

---

## Additional Open Questions

### OQ-004: Tag pinning — no git tags exist in the Omarchy repo; how should version pinning work?

**Question:** The Omarchy repository has no git tags (`git describe --tags` fails; the version
file reads `4.0.0.alpha`). Pinning to a commit hash (`9cf1852525a5f7de26d3162db9d61e2f5c1d5523`)
is the only reliable method at this time.
Should DebateOS opinion manifests cite commit hash only, or also track a branch/tag once
stable tagging resumes? What is the canonical version field format — semver (4.0.0.alpha),
commit hash, or both?

**Evidence:** All Phase 0 deliverables pin to the commit hash. The `version` file contains
`4.0.0.alpha`. The repo is under active development (alpha version, no tags).

**Lean (from RESEARCH.md):** Record commit hash as the canonical pin for now. When the repo
adopts semver tags (post-alpha), the opinion manifest should carry both tag and commit hash
for redundancy. Schema should support a `version_pin` field with both `tag` (nullable) and
`commit` subfields.

**Resolved:** Phase 1 schema design (opinion manifest versioning field)

---

### OQ-005: Is the Omarchy custom repo itself an opinion, or a foundation prerequisite?

**Question:** Installing Omarchy requires adding `pkgs.omarchy.org` (the `[omarchy]` repo) to
pacman.conf and importing the GPG key `40DFB630FF42BCFFB047046CF0134EE680CAC571`. Without
this, no Omarchy-specific packages (omarchy-walker, omarchy-nvim-setup, etc.) can be installed.
Is "add the Omarchy repo" an opinion (OM-001, category: `custom-repo`) that belongs in the
speech? Or is it a foundation prerequisite that the Arch translator automatically applies
whenever an Omarchy speech is targeted at Arch?

**Evidence:** OM-001 is already cataloged as a `custom-repo` opinion and is listed as a
dependency of all subsequent Omarchy package opinions. It is called as the first step in
`install/preflight/pacman.sh`.

**Lean (from RESEARCH.md):** Treat it as an opinion (OM-001). This preserves composability —
a speech that doesn't use any Omarchy-specific packages should not be required to add the repo.
The translator enforces OM-001 as a prerequisite only when other OM-NNN opinions depend on it
(via the standard dependency chain).

**Resolved:** Phase 1 schema design (custom-repo opinion category definition and ordering)

---

### OQ-006: CachyOS CPU arch level — static declaration in variant profile vs runtime detection?

**Question:** CachyOS ships three pacman.conf variants based on CPU ISA level: `x86_64`,
`x86_64-v3` (AVX2+), and `x86_64-v4` (AVX-512). Selecting the correct variant requires
CPU capability detection at install time. The Phase 0 variant profile sketch declares this
as a static field (`cpu_arch_level: v3`).
Should CachyOS CPU arch level be a static declaration in the variant profile (user declares it
at speech composition time) or a dynamic hardware condition evaluated at translator run time
(the translator detects AVX2/AVX-512 support via cpuid and selects the appropriate repo)?

**Evidence:** CachyOS `script-v3-v4.sh` in `linux-cachyos` repo shows Docker-based v3/v4 build
selection; CachyOS `pacman-v3.conf` and `pacman-v4.conf` differ only in repo names and
mirrorlist paths. The Phase 0 variant profile declares `cpu_arch_level: v3` as static.

**Lean (from RESEARCH.md):** Static declaration is simpler for v1 (matches the docs/08
"variant profile" approach). Dynamic detection is more user-friendly but adds complexity.
Defer to Phase 1 for the final decision; record both approaches.

**Resolved:** Phase 1 schema design / Phase 2 Arch translator design

---

### OQ-007: Variant-profile conflict semantics — foundation pre-seeded opinion vs user opinion

**Question (deferred from CONTEXT.md):** When a variant profile (e.g., Garuda) pre-seeds an
opinion (e.g., `grub-btrfs` + snapper root config) and the user's speech also includes a
conflicting opinion (e.g., Omarchy limine + snapper config), what is the semantic of the
conflict?

Options explored:
1. Pre-seeded opinions are always required — any conflicting user opinion is a hard conflict
   that requires explicit resolution (drop one or add a patch)
2. Pre-seeded opinions are nice-to-have by default — a required user opinion wins automatically
3. Pre-seeded opinions have a special `foundation-default` status distinct from required/nice-to-have
   — the resolver prompts the user to confirm replacement

**Evidence:** EC-001 (Garuda snapper vs Omarchy snapper), EC-002 (Garuda GRUB vs Omarchy limine),
EC-003 (Garuda SDDM vs Omarchy SDDM) — all in `research/resolver-edge-cases.md`. All three are
evidence-backed hard conflicts from real variant data.

**Lean:** Pre-seeded opinions should be required by default (option 1); the variant profile
author declares the required set. This keeps the resolution model simple and matches the
docs/04 hierarchy. A `foundation-default: true` metadata flag could mark opinions as
"from the variant, overridable with explicit declaration" for a UX improvement.

**Resolved:** Phase 1 schema design — DEFERRED from Phase 0 per CONTEXT.md deferred items.
Do NOT resolve here.

---

### OQ-008: Unmapped scripts — 306 runtime bin/ helpers and .desktop files

**Question (from inventory Coverage Notes):** The `bin/` directory contains 306 `omarchy-*`
utility scripts (e.g., `omarchy-theme-set`, `omarchy-pkg-add`, `omarchy-hibernation-setup`,
`omarchy-hw-*` hardware detection helpers). These scripts are invoked BY opinion scripts
during install but do not themselves represent post-install decisions. They are runtime
infrastructure that the translator must deploy.

Does the schema need a way to express "this opinion requires a set of runtime helper scripts
to be present on the target system"? And are the 306 helper scripts themselves opinions
(category: `arbitrary-script`, nice-to-have, runtime-only), or are they translator
infrastructure that all Arch-targeted speeches automatically receive?

**Evidence:** `omarchy-hw-asus-rog`, `omarchy-battery-present`, `omarchy-hw-framework16` —
hardware detection helpers called from conditional opinions. `omarchy-theme-set`,
`omarchy-pkg-add` — runtime management tools. `bin/` is NOT referenced in any `all.sh`
install pipeline; these are post-install helper commands, not install-time opinions.

**Lean:** Treat as translator infrastructure: a `runtime-helpers` translator capability
requirement in opinion metadata, not a separate opinion category. Phase 1 must decide
whether the schema carries a `translator-runtime-deps` field for opinions that depend on
specific runtime helpers being deployed.

**Resolved:** Phase 1 schema design (translator capability field extension)

---

### OQ-009: CachyOS default filesystem and bootloader — installer-menu dependent, not PKGBUILD-determinable

**Question (from arch-variants-delta.md Open Questions):** The CachyOS calamares installer
presents a menu for filesystem (btrfs, ext4, xfs, f2fs) and bootloader (systemd-boot, GRUB,
limine) choices. Without running the installer, the exact installed default cannot be confirmed
from PKGBUILD inspection alone. The Phase 0 variant profile declares `filesystem: null` and
`bootloader: null` (user choice).

How should the CachyOS variant profile express this ambiguity? Options:
1. Leave as `null` (user declares at speech composition time)
2. Document all supported options as an enum, requiring the user to pick one
3. Mark as [UNVERIFIED] and note the Phase 2 stretch validation will resolve it

**Evidence:** `cachyos-snapper-support`, `grub-btrfs-support`, `systemd-boot-manager`,
`dracut-cachyos` — all optional CachyOS packages implying menu choice. The `pacman.conf`
confirms the repo structure but not the installer defaults.

**Lean:** Option 2 — document the confirmed menu options as an enum. The variant profile
declares `filesystem: {options: [btrfs, ext4, xfs, f2fs], default: null}` and
`bootloader: {options: [systemd-boot, grub, limine], default: null}`. The speech author
must declare a choice.

**Resolved:** Phase 2 Arch translator design (variant profile schema) / Phase 0 research
marks it as [UNVERIFIED]

---

### OQ-010: Garuda [garuda] repo URL is unverified — calamares config requires auth to clone

**Question (from arch-variants-delta.md Open Questions):** The Garuda variant profile
references `[garuda]` repo at `https://geo-mirror.chaotic.cx/garuda/$arch` [UNVERIFIED].
The GitLab ISO profile configs (primary source) require auth to clone. The GitHub mirrors
only contain PKGBUILDs, not the ISO profile that adds `[garuda]` to the installed
`pacman.conf`.

How should the DebateOS variant profile handle [UNVERIFIED] claims? Flag with a warning
that Phase 2 (when an actual Garuda ISO is inspected) will correct the value?

**Evidence:** `garuda-update/main-update` script uses `https://secret-mirror.chaotic.cx/`
as a fallback; standard CDN mirror is inferred. The `[garuda]` URL is not in any cloned
source.

**Lean:** Flag with `[UNVERIFIED]` annotation and a note that Phase 2 stretch validation
will confirm. Phase 1 schema should support a `verification_status` field on variant profile
repo entries.

**Resolved:** Phase 2 stretch validation (Garuda retarget study)

---

## Summary

| ID | Question | Resolution Phase | Mandatory? |
|----|----------|-----------------|------------|
| OQ-001 | Migrations-as-schema-concept vs version-pinning | Phase 1 / post-v1 | YES (PLAN.md) |
| OQ-002 | execution-phase field (first-run vs install-time) | Phase 1 schema | YES (PLAN.md) |
| OQ-003 | runtime-tool-install category (npm vs OS packages) | Phase 1 schema | YES (PLAN.md) |
| OQ-004 | Tag pinning — no git tags, version format | Phase 1 schema | No |
| OQ-005 | Omarchy repo as opinion vs foundation prerequisite | Phase 1 schema | No |
| OQ-006 | CachyOS CPU arch level — static vs dynamic selection | Phase 1/Phase 2 | No |
| OQ-007 | Variant-profile conflict semantics (deferred from CONTEXT.md) | Phase 1 schema | No (deferred) |
| OQ-008 | 306 runtime bin/ helpers — opinion or translator infra? | Phase 1 schema | No |
| OQ-009 | CachyOS default filesystem/bootloader (installer-menu) | Phase 2 | No |
| OQ-010 | Garuda [garuda] repo URL unverified | Phase 2 | No |
