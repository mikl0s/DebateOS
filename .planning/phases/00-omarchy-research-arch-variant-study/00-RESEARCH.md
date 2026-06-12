# Phase 0: Omarchy Research & Arch-Variant Study — Research

**Researched:** 2026-06-12
**Domain:** Omarchy source analysis + CachyOS/Garuda delta study (Arch-based distro opinions)
**Confidence:** MEDIUM (Omarchy: HIGH — cloned source analyzed directly; Variants: LOW — web sources only, no local clone)

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- Pin Omarchy to the latest stable release tag at clone time; record tag + commit hash in every deliverable so all evidence is reproducible.
- Every opinion-inventory entry cites its source path (install script, config file, dotfile) in the cloned repo at the pinned commit.
- CachyOS/Garuda evidence comes from their public git repos (repo definitions, calamares configs, kernel PKGBUILDs) and official docs — no full ISO installs required for a delta catalog.
- Depth boundary: exhaustive for Omarchy (every post-base-Arch decision, per docs/08); targeted delta-only for the variants (repos, keyrings, kernels, defaults, pre-seeded configs/tooling).
- Markdown inventory with fixed per-opinion fields: stable ID (`OM-NNN`), category, OS-agnostic intent, dependencies, ordering constraints, un-agnostic flags (→ translator capability requirements).
- Category taxonomy starts from the docs/08 list; new categories added when evidence demands and flagged as schema surprises.
- Atomic granularity: one opinion per post-install decision; grouping only in `omarchy-points.md`.
- Hardware-conditional behaviors recorded as hardware-conditional opinions with explicit condition metadata.
- Structured Given/When/Then scenarios with stable IDs (`EC-NNN`) and expected explanation text for resolver edge cases.
- Include a coverage matrix mapping every docs/04 resolution rule to at least one EC-NNN scenario.
- Variant collision scenarios enumerated per variant at minimum: kernel-default vs kernel-opinion (CachyOS), theming collision (Garuda), repo-priority conflict (CachyOS/Chaotic-AUR).
- `arch-variants-delta.md` proposes a declarative YAML variant-profile sketch as a candidate only — the final schema is drafted in Phase 1.
- "Foundation already has an opinion about this" is modeled as a pre-seeded opinions list in the profile; conflict semantics deferred to Phase 1.

### Claude's Discretion

- Exact clone/inspection tooling and working directory layout for the research.
- Internal section ordering of each deliverable, as long as the required fields and IDs are present.
- How many candidate points to propose in omarchy-points.md (driven by evidence, not a quota).

### Deferred Ideas (OUT OF SCOPE)

- Phase 2 stretch validation (Omarchy speech retargeted to a variant) — recorded in docs/08; not a Phase 0 deliverable beyond knowing what it would take.
- Variant-profile conflict semantics ("foundation pre-seeded opinion vs user opinion") — decided in Phase 1 schema design, question recorded in open-questions.md.
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| RSCH-01 | Omarchy deep-dive from cloned source: every post-base-Arch decision as a candidate atomic opinion with category, OS-agnostic intent, dependencies/ordering, translator-capability fallout — delivered as `research/omarchy-opinion-inventory.md`, `research/omarchy-points.md`, `research/schema-requirements.md`, `research/open-questions.md` | Direct clone analysis of 9cf1852 confirms ~120–150 distinct post-base opinions across 8 install phases; category taxonomy and field set defined below |
| RSCH-02 | CachyOS/Garuda variant substitution study as `research/arch-variants-delta.md` with proposed declarative variant-profile shape | CachyOS and Garuda repo structures confirmed via GitHub; key repos, keyrings, kernel variants, and default filesystem choices cataloged below |
| RSCH-03 | Resolver edge-case corpus as `research/resolver-edge-cases.md`: cross-variant effectuation differences, foundation-default-vs-opinion collisions, repo-priority conflicts — each as a Phase 1 test scenario | Concrete collision scenarios derived from the source analysis and variant delta study; coverage matrix for all docs/04 rules defined below |
</phase_requirements>

---

## Summary

Omarchy is a post-base-Arch installer implemented as a structured set of shell scripts (commit `9cf1852525a5f7de26d3162db9d61e2f5c1d5523`, version `4.0.0.alpha`). The install pipeline runs in six sequential phases — preflight, packaging, config, login, post-install, first-run — each orchestrated by an `all.sh` that calls individual scripts. Direct analysis of the cloned source reveals approximately 120–150 distinct post-base-Arch decisions that map to atomic opinions, organized across ~110 install scripts plus 155 base packages and 73 auxiliary packages. Hardware-conditional logic is pervasive: 34 hardware scripts covering Apple T2, NVIDIA, Intel PTL, ASUS ROG, Framework, Dell XPS, Lenovo, Surface, and more. The theme system ships 21 complete asset bundles, and 313 migrations document the evolution of the system. Schema surprises confirmed from the source: ordering is critical (preflight before packaging before config is enforced by the pipeline structure), arbitrary script payloads are first-class (both inline shell and file deployments), theming carries file assets (PNG backgrounds, theme configs, per-app color maps), and hardware detection gates whole blocks of opinions.

For the variant study, CachyOS and Garuda are both Arch-based but opinionated in fundamentally different dimensions. CachyOS operates a layered repo system (`cachyos`, `cachyos-v3`, `cachyos-v4`) that must sit above standard Arch repos in priority, and ships a custom kernel family (`linux-cachyos`, `linux-cachyos-bore`, `linux-cachyos-lts`) with CPU-architecture-optimized variants. Garuda ships Chaotic-AUR as a first-class repo alongside a `[garuda]` repo, defaults to btrfs + snapper + grub-btrfs, pre-seeds heavy KDE/Dr460nized theming, and replaces `pacman` updates with `garuda-update`. Both variants pre-seed opinions that conflict directly with what the Omarchy speech would install, producing a rich corpus of resolver edge cases.

**Primary recommendation:** Walk the Omarchy install pipeline in phase order (preflight → packaging → config → login → post-install → first-run), treating each script as an opinion candidate. Use `install/config/all.sh`, `install/packaging/all.sh`, and `install/login/all.sh` as the definitive ordered script lists. Enumerate hardware scripts separately as hardware-conditional opinions. Extract variant deltas from public GitHub repos without ISO installs.

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Opinion inventory extraction | Research / document | — | Pure analysis output; no code tier exists yet |
| Opinion ID assignment (OM-NNN) | Research / document | — | Stable IDs must be assigned during research, not later |
| Point grouping | Research / document | — | Grouping reflects natural clusters in Omarchy; informs schema design |
| Schema floor derivation | Research → Phase 1 Schema | — | Phase 0 produces evidence; Phase 1 drafts the schema |
| Variant delta catalog | Research / document | — | Targeted analysis from public repos, no runtime needed |
| Edge-case corpus | Research / document | Phase 1 Resolver | EC-NNN scenarios feed Phase 1 TDD harness directly |
| Variant-profile shape sketch | Research / document | Phase 2 Arch Translator | Candidate YAML shape; final schema in Phase 1 |
| Hardware detection logic | Omarchy source (reference) | Phase 2 Translator | Documented as translator capability requirements, not implemented in Phase 0 |

---

## Omarchy Repository Structure (Verified from Clone)

**Source:** `https://github.com/basecamp/omarchy`
**Commit:** `9cf1852525a5f7de26d3162db9d61e2f5c1d5523`
**Version file:** `4.0.0.alpha`
**Note:** No git tags exist in the shallow clone (`git describe --tags` fails). The version file reads `4.0.0.alpha`. Pin by commit hash.
[VERIFIED: direct git clone + file inspection]

### Top-Level Layout

```
omarchy/
├── boot.sh              # entry point for curl-pipe installs (sets mirror, clones, calls install.sh)
├── install.sh           # main orchestrator: sources helpers + 5 phase all.sh files
├── version              # version string ("4.0.0.alpha")
├── install/
│   ├── helpers/         # logging, error handling, presentation, chroot helpers
│   ├── preflight/       # pre-install guards: pacman config, migrations, guard, begin
│   ├── packaging/       # package installation scripts (base, fonts, nvim, icons, webapps, tuis, npm, hardware-specific)
│   ├── config/          # system configuration scripts (34 general + 34 hardware-specific = 68 total)
│   ├── login/           # boot/login setup (plymouth, sddm, hibernation, limine-snapper)
│   ├── post-install/    # finalization (pacman re-config, allow-reboot, finished)
│   ├── first-run/       # deferred first-login tasks (wifi, gnome-theme, firewall, dns, gdk-scale)
│   ├── omarchy-base.packages   # 155 packages (excluding comments/blanks)
│   └── omarchy-other.packages  # 73 packages (hardware + optional)
├── config/              # dotfiles and config file templates (32 app subdirs)
├── default/             # default config files deployed at install time (25+ subdirs)
├── themes/              # 21 complete theme bundles (each with backgrounds, btop, colors, neovim, etc.)
├── migrations/          # 313 timestamped migration scripts
├── bin/                 # 306 omarchy-* utility scripts
├── applications/        # .desktop files
└── test/                # test infrastructure
```

[VERIFIED: direct directory inspection of cloned repo]

### Install Pipeline Phases (Execution Order)

The pipeline is strictly ordered. `install.sh` sources phases sequentially:

1. **helpers** — logging, error handling, chroot wrappers (not opinions, infrastructure)
2. **preflight** — pacman config (adds `[omarchy]` repo + mirror), runs pending migrations, guards re-runs
3. **packaging** — installs all packages (base list, then specialized scripts)
4. **config** — applies all system configuration (general then hardware-specific)
5. **login** — configures display manager, boot loader, snapper
6. **post-install** — restores pacman config, signals reboot allowed
7. **first-run** — deferred opinions that require a live display session (WiFi UI, gnome theming, firewall, GDK scale)

**Ordering constraint significance:** packaging must run before config (packages must exist before configs apply), login must run after config (bootloader config depends on system state), first-run must run on the target machine after first login. This ordering is mandatory metadata for the opinion schema.

[VERIFIED: direct analysis of install.sh and all.sh scripts]

### Custom Repository

Omarchy maintains its own pacman repository (`[omarchy]`) hosted at `pkgs.omarchy.org`. This repo sits below the standard Arch core/extra/multilib repos in priority. Three mirror channels: `stable`, `rc`, `edge`. The keyring package is `omarchy-keyring` (GPG key `40DFB630FF42BCFFB047046CF0134EE680CAC571`).

```
[omarchy]
SigLevel = Optional TrustAll
Server = https://pkgs.omarchy.org/stable/$arch
```

For Apple T2 hardware, an additional unofficial repo (`arch-mact2`) is added with `SigLevel = Never`. This is a hardware-conditional repo opinion.

[VERIFIED: direct analysis of `default/pacman/pacman-stable.conf`]

---

## Opinion Volume Estimate

| Install Phase | Script Count | Estimated Distinct Opinions | Notes |
|---------------|-------------|----------------------------|-------|
| preflight | 8 | ~6 | pacman repo + mirror config, migrations invocation, guard |
| packaging/base | 1 (+ pkg list) | ~1 | All 155 base packages as a bulk-install opinion or ~10–15 logical groupings |
| packaging/specialized | 11 scripts | ~15 | fonts, nvim, icons, webapps, tuis, npm AI tools, hardware-specific packaging |
| config/general | 34 scripts | ~34 | Each script is typically one focused config opinion |
| config/hardware | 34 scripts | ~34 (conditional) | Each is a hardware-conditional opinion with explicit condition |
| login | 6 scripts | ~6 | plymouth, sddm, sddm autologin, hibernation, limine-snapper |
| post-install | 4 scripts | ~3 | pacman finalization, reboot unlock |
| first-run | 12 scripts | ~12 | deferred display-session opinions |
| **Total** | **~110** | **~111–125 non-hardware + ~34 hardware-conditional** | Plus ~21 theme opinions |

**Practical ID range:** OM-001 through OM-200 provides comfortable headroom.
**Observation:** The base package list (155 packages) could be split into ~10–15 logical groups (compositor, terminal, browser, dev tools, AI tools, media, fonts, etc.) or treated atomically. The inventory should prefer fine-grained groupings to maximize composability.

[VERIFIED: direct script count from cloned repo]

---

## Category Taxonomy

The docs/08 starting taxonomy is confirmed and expanded by direct source analysis:

| Category | Evidence in Omarchy | OS-Agnostic? | Schema Note |
|----------|--------------------|----|-------------|
| `package-install` | Base packages, specialized packaging scripts | YES (list of logical packages) | Package names are Arch-specific; translator maps to distro packages |
| `package-removal` | Migrations drop packages; hardware scripts remove conflicting kernels | YES (logical package intent) | Translator handles `pacman -Rdd` semantics |
| `config-dotfile` | `config/` tree deployed to `~/.config/`; `default/` tree deployed to `/etc/` | YES (intent: deploy file X to path Y) | File payload must be schema-expressible |
| `service-enable` | bluetooth, docker.socket, sddm, intel_lpmd, t2fanrd, ufw, etc. | YES (service name as logical capability) | Translator maps to systemctl/launchd |
| `kernel-boot-param` | `KERNEL_CMDLINE[default]+=" fred=on"`, apple iommu args, limine config | PARTIAL — bootloader-specific | Translator capability: "set kernel cmdline param X" |
| `theming` | 21 theme bundles: PNG backgrounds, btop themes, neovim color schemes, waybar CSS | PARTIAL — per-app asset files | Schema must carry file asset payloads |
| `keybinding` | hypr/keybindings.conf in default/ | YES (logical action → key combo) | Asset-based or structured metadata |
| `hardware-conditional` | 34 hardware scripts with explicit device detection | YES (condition + block of opinions) | Condition metadata: DMI match, PCI ID, lspci pattern |
| `arbitrary-script` | omarchy-ai-skill.sh (symlinks to agent dirs), walker-elephant.sh, pi.sh | PARTIAL | Schema must support opaque script payloads with declared capabilities |
| `user-group` | `usermod -aG docker ${USER}`, `usermod -aG video ${USER}` | YES | Translator maps to OS group management |
| `sysctl-param` | increase-file-watchers, increase-fd-limit, ssh-flakiness (tcp_mtu_probing) | YES | Translator writes sysctl.d drop-in |
| `custom-repo` | `[omarchy]` repo, `arch-mact2` repo (hardware-conditional) | PARTIAL | Translator capability: add signed repo to package manager |
| `npm-global-install` | codex, gemini, copilot, opencode, playwright, pi, ghui, hunk | YES (logical tool) | Translator invokes npm; schema captures tool names + version |
| `display-manager` | sddm config, autologin, wayland session .desktop | YES (intent: auto-login to session X) | Translator handles per-DM specifics |
| `bootloader-config` | limine setup, limine-snapper-sync, UKI, EFI bootmgr | PARTIAL — bootloader-specific | Translator capability: configure bootloader Y |
| `snapshot-policy` | snapper config, btrfs quota disable | YES (intent: snapshot-on-update) | Translator capability: configure snapshot tool |
| `firewall-rule` | ufw rules (deny-in, allow LocalSend, allow Docker DNS) | YES | Translator maps to ufw/firewalld |
| `dns-config` | systemd-resolved stub-resolv symlink | YES | Translator handles resolver configuration |
| `font-install` | fonts packaging script | YES | Package-based; some fonts from AUR |
| `mime-type` | mimetypes.sh (xdg-mime associations) | YES | Translator invokes xdg-mime or mimeapps.list |

**Schema surprises confirmed (expand beyond docs/08 floor):**
1. **File asset payloads** — theming bundles, dotfile templates, plymouth theme files; schema must carry binary/text file content or references.
2. **Custom package repository registration** — adding a signed repo to pacman is itself an opinion with `SigLevel`, URL, and keyring package dependency.
3. **npm global installs** — entirely separate from the distro package manager; the schema must express cross-tool installs.
4. **Ordering is load-bearing at phase level** — preflight before packaging before config is not optional; schema ordering constraints must be expressive enough to capture this.
5. **Multi-output hardware detection** — the 18 `omarchy-hw-*` helpers reveal that hardware conditions are not just boolean flags but compound predicates (PCI ID match + DMI match + battery presence etc.).
6. **Arbitrary script payloads with declared capabilities** — `omarchy-ai-skill.sh` is not expressible as a package or config; it deploys dotfiles to multiple agent directories. The schema needs an opaque script payload type with declared translator capability requirements.
7. **Migrations as opinions** — 313 timestamped migrations represent the opinion evolution history of an installed system. This is not directly in scope for v1 schema, but the migration pattern is worth recording as an open question.

[VERIFIED: direct analysis of cloned source]

---

## Walk Order for Inventory Analysis

The planner should task the analysis to walk the install pipeline in this exact order to ensure completeness:

1. `install/preflight/all.sh` — read each referenced script in order
2. `install/packaging/all.sh` — walk each packaging script; process `omarchy-base.packages` and `omarchy-other.packages` for package opinions
3. `install/config/all.sh` — walk each config script in listed order (both general and hardware subdirs)
4. `install/login/all.sh` — walk login scripts
5. `install/post-install/all.sh` — walk post-install scripts
6. `install/first-run/` — enumerate all `.sh` files (no `all.sh` — check for an orchestrator or enumerate directly)
7. `themes/` — catalog all 21 theme directories as theming opinions
8. `config/` — verify dotfile directories against config scripts (cross-reference)
9. `default/` — verify default config files against what scripts deploy
10. `migrations/` — sample only (not exhaustive; note count + pattern as open question)

**Completeness check:** Every `.sh` file in `install/` that is called from an `all.sh` should produce at least one OM-NNN entry. Files not referenced by any `all.sh` should be flagged in `open-questions.md`.

[VERIFIED: cross-referencing all.sh files with directory contents in cloned repo]

---

## Arch-Variant Delta Study

### CachyOS

**Public repos to analyze:**
- `https://github.com/CachyOS/CachyOS-PKGBUILDS` — PKGBUILDs including `pacman/pacman.conf`
- `https://github.com/CachyOS/linux-cachyos` — Kernel PKGBUILDs
- `https://github.com/CachyOS/CachyOS-Settings` — System defaults (etc/, usr/ trees)
- `https://github.com/CachyOS/docker` — Docker-specific pacman.conf files (authoritative repo ordering examples)

[CITED: https://github.com/CachyOS]

**Key deltas from vanilla Arch:**

| Delta | Detail | Pre-seeded Opinion? | Collision Risk |
|-------|--------|-------------------|---------------|
| Custom repos | `[cachyos]`, `[cachyos-v3]`/`[cachyos-core-v3]`/`[cachyos-extra-v3]` for x86-64-v3 CPUs, `[cachyos-v4]`/`[cachyos-core-v4]`/`[cachyos-extra-v4]` for x86-64-v4 CPUs — placed ABOVE Arch repos | YES — `custom-repo` opinion | Omarchy's `[omarchy]` repo priority vs CachyOS repos |
| Keyring | `cachyos-keyring` package must be installed first | YES | Must precede package installs |
| Kernel | `linux-cachyos` (EEVDF-tuned) as default; also `linux-cachyos-bore`, `linux-cachyos-lts`, `linux-cachyos-bmq`, `linux-cachyos-rt`; all available in x86_64, x86_64-v3, x86_64-v4 variants | YES — `kernel-install` opinion | Collides with Omarchy's Intel PTL kernel override |
| Optimized packages | Core packages (glibc, mesa, ffmpeg, etc.) rebuilt with `-O3 -march=x86-64-v3` or v4 plus LTO/PGO/BOLT | YES (implicit) | Same logical package, different binary — variant profile must declare arch-level |
| sched-ext | `CONFIG_SCHED_CLASS_EXT=y` in cachyos kernels — BPF scheduler framework enabled | YES | No direct Omarchy collision; adds capability |
| CachyOS-Settings | Performance sysctl tuning, ananicy-cpp, zram-generator pre-configured | YES — multiple `sysctl-param` opinions | Could conflict with Omarchy's sysctl opinions |
| Mirror priority | CachyOS repos must come before core/extra/multilib | YES — ordering constraint | `[omarchy]` also has custom ordering; priority resolution needed |

[CITED: https://wiki.cachyos.org/features/optimized_repos/, https://github.com/CachyOS/docker/blob/master/pacman-v3.conf]
[ASSUMED: exact contents of CachyOS-Settings etc/ tree — confirmed via webfetch that etc/ and usr/ trees exist but specific files not extracted]

**CachyOS variant-profile candidate fields:**
```yaml
variant: cachyos
repos:
  - name: cachyos-v3        # conditional: x86-64-v3 CPU
    url: https://mirror.cachyos.org/repo/x86_64_v3/cachyos-v3
    keyring: cachyos-keyring
    priority: above-arch    # must precede core/extra
  - name: cachyos-core-v3
    url: https://mirror.cachyos.org/repo/x86_64_v3/core
    priority: above-arch
  # ... v4 variants analogously
kernel:
  package: linux-cachyos    # default; also bore/lts/bmq/rt available
  headers: linux-cachyos-headers
cpu_arch_level: v3          # or v4; auto-detected at install time
pre_seeded_opinions:
  - kernel-install/cachyos-default-kernel
  - sysctl/cachyos-performance-tuning
  - zram/cachyos-zram-defaults
```

### Garuda Linux

**Public repos to analyze:**
- `https://github.com/garuda-linux/pkgbuilds` — All Garuda PKGBUILDs (GitHub mirror of GitLab)
- `https://github.com/garuda-linux/pkgbuilds-aur` — Curated AUR package list
- `https://github.com/garuda-linux/garuda-tools` — Image build tooling
- `https://github.com/chaotic-aur/pkgbuilds` — Chaotic-AUR package lists

[CITED: https://github.com/garuda-linux]

**Key deltas from vanilla Arch:**

| Delta | Detail | Pre-seeded Opinion? | Collision Risk |
|-------|--------|-------------------|---------------|
| Custom repos | `[garuda]` repo + `[chaotic-aur]` with `Include = /etc/pacman.d/chaotic-mirrorlist`; keyring: `chaotic-keyring` (GPG key `FBA220DFC880C036`) | YES | Priority vs Omarchy's `[omarchy]` repo; package shadowing |
| Default filesystem | btrfs (not ext4) — btrfs maintenance tools pre-installed | YES — `filesystem-default` opinion | Omarchy also uses btrfs+limine-snapper; compatible but via different path |
| Snapper setup | snapper + grub-btrfs pre-configured; automatic snapshot-on-update | YES — `snapshot-policy` opinion | Omarchy has its own snapper config — direct collision on snapper root config |
| Bootloader | GRUB (not limine) as default; grub-btrfs integration for snapshot boot | YES | Direct collision: Omarchy is limine-first |
| Theming (Dr460nized) | KDE Plasma Dr460nized theme: grub theme, SDDM theme, KDE color scheme, icon set, wallpapers | YES — multiple `theming` opinions | Omarchy has its own SDDM theme + Plymouth theme — SDDM collision |
| garuda-update wrapper | Replaces direct `pacman -Syu` with `garuda-update` which handles Chaotic-AUR key refreshes | YES — `update-mechanism` opinion | Omarchy's preflight pacman refresh would bypass garuda-update |
| ZRAM | zram-generator configured by garuda-common-settings | YES | Omarchy may also configure zram (CachyOS-Settings does — check Garuda) |
| systemd-oomd | systemd-oomd-defaults pre-configured | YES | Could conflict with Omarchy's OOM settings |
| btrfs subvolume layout | Garuda's calamares uses a specific subvolume layout (@, @home, @root, @srv, @snapshots, @log, @pkg) | YES — `filesystem-layout` opinion | A new schema category |

[CITED: https://gadgeteer.co.za/garuda-linux-rolling-release-distro-based-arch-linux-btrfs-default-filesystem-easy-rollbacks-grub/]
[ASSUMED: exact garuda-common-settings contents — PKGBUILD fetched but specific config files require repo clone to enumerate fully]

**Garuda variant-profile candidate fields:**
```yaml
variant: garuda
repos:
  - name: garuda
    url: https://geo-mirror.chaotic.cx/garuda/$arch
    keyring: garuda-keyring
    priority: above-arch
  - name: chaotic-aur
    mirrorlist: /etc/pacman.d/chaotic-mirrorlist
    keyring: chaotic-keyring
    priority: below-garuda
kernel:
  package: linux       # or linux-zen; Garuda does not ship a custom kernel
filesystem:
  type: btrfs
  subvolumes: ["@", "@home", "@root", "@srv", "@snapshots", "@log", "@pkg"]
bootloader: grub       # grub-btrfs for snapshot boot
pre_seeded_opinions:
  - snapshot-policy/garuda-snapper-grub
  - theming/garuda-dr460nized
  - update-mechanism/garuda-update
  - filesystem-layout/garuda-btrfs-subvolumes
```

---

## Resolver Edge-Case Corpus: Confirmed Collision Classes

These are the concrete conflict classes that `research/resolver-edge-cases.md` must express as EC-NNN Given/When/Then scenarios. Each class is evidence-backed from the source analysis above.

### Class 1: Foundation Pre-seeded Opinion vs User Opinion (same capability)

| Scenario | Foundation | User Opinion | Collision Type |
|----------|-----------|--------------|---------------|
| EC-001 | Garuda | Omarchy snapper config | Both configure snapper root; the config file is the same path `/etc/snapper/configs/root` |
| EC-002 | Garuda | Omarchy limine bootloader | Garuda pre-seeds GRUB; Omarchy installs limine — required vs required |
| EC-003 | Garuda | Omarchy SDDM theme | Both set SDDM theme; file-level collision |
| EC-004 | CachyOS | Omarchy linux kernel | CachyOS pre-seeds `linux-cachyos`; Omarchy installs `linux` + optionally `linux-ptl` — direct package conflict |
| EC-005 | CachyOS | Omarchy sysctl (file-watchers, fd-limit, ssh) | CachyOS-Settings writes sysctl.d drop-ins; Omarchy writes to same directory — merge or conflict |

### Class 2: Repo Priority Conflicts

| Scenario | Conflict |
|----------|---------|
| EC-010 | Omarchy `[omarchy]` repo vs CachyOS repos — same package name, different binary (e.g. if any Omarchy package also exists in cachyos repo) |
| EC-011 | CachyOS repos must be above Arch repos; Omarchy repo in pacman.conf after Arch repos — priority ordering is an opinion, not data |
| EC-012 | Garuda `[garuda]` + `[chaotic-aur]` vs Omarchy `[omarchy]` — both add custom repos; relative ordering matters |

### Class 3: Cross-Variant Effectuation Differences

| Scenario | Same Opinion | Vanilla Arch | CachyOS | Garuda |
|----------|-------------|-------------|---------|--------|
| EC-020 | Install `mesa` | Arch package | CachyOS v3/v4 optimized build | Standard Arch build via Chaotic-AUR |
| EC-021 | Install `linux-headers` | `linux-headers` | `linux-cachyos-headers` | `linux-headers` (or zen-headers) |
| EC-022 | Configure btrfs snapshots | Manual snapper setup | Manual (no pre-seed) | snapper pre-configured — idempotency concern |
| EC-023 | Enable bluetooth service | `systemctl enable bluetooth.service` | Same | Same, but garuda-common-settings may already do this |

### Class 4: docs/04 Rule Coverage (Required for Complete Test Matrix)

| Rule | EC-IDs | Scenario |
|------|--------|---------|
| required beats nice-to-have | EC-030 | Required kernel opinion drops nice-to-have audio codec opinion that depends on the same kernel |
| required-vs-required hard conflict | EC-031 | Two required kernel opinions (linux vs linux-cachyos) — no patch exists |
| required-vs-required with patch | EC-032 | Two required compositor opinions with a compatibility patch opinion |
| nice-vs-nice sensible default | EC-033 | Two nice-to-have terminal emulators (foot vs ghostty) — resolver picks one with explanation |
| patch overrides hierarchy | EC-034 | Required vs required conflict resolved by patch opinion |
| ordering/toposort | EC-035 | Opinion A depends on B; B depends on C; correct install order derived |
| cycle detection | EC-036 | Opinion A after B, B after A — hard error naming both |
| hardware-conditional skip | EC-037 | NVIDIA driver opinion skipped because hardware condition is false (no NVIDIA GPU declared) |
| hardware-conditional apply | EC-038 | Apple T2 opinion block applied because T2 PCI ID condition matches declared hardware |

### Class 5: Variant-Kernel Collision (CachyOS-specific)

| Scenario | Detail |
|----------|--------|
| EC-040 | User speech includes `kernel-install/linux-vanilla`; variant profile pre-seeds `linux-cachyos` — two kernel opinions, different packages, required vs pre-seeded |
| EC-041 | User speech targets `x86-64-v3` variant; hardware declares `x86-64-v4` capable CPU — should the resolver upgrade to v4? |

### Class 6: Theming Collision (Garuda-specific)

| Scenario | Detail |
|----------|--------|
| EC-050 | Omarchy theming opinion sets SDDM theme to `omarchy`; Garuda pre-seeds SDDM theme as `garuda-dr460nized` — which wins? |
| EC-051 | Omarchy deploys `default/plymouth/omarchy` theme; Garuda pre-seeds its own plymouth theme — same config path, different content |

[ASSUMED: EC-040–EC-051 are synthesized from confirmed variant delta data; provenance to be marked in edge-case corpus]

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Counting scripts in a directory | Custom code | `find install/ -name "*.sh" \| wc -l` | Already shown to work; standard shell tools |
| Walking install pipeline in order | Custom parser | Read `all.sh` files sequentially | `all.sh` files are the authoritative ordered lists |
| Identifying hardware conditions | AST analysis | `grep -r "omarchy-hw-\|if lspci\|if grep.*dmi" install/config/hardware/` | Hardware detection is always via helper or inline grep |
| Package list extraction | Manual transcription | `grep -v "^#" omarchy-base.packages \| grep -v "^$"` | Direct file read; same pattern used by `install/packaging/base.sh` |
| Variant repo extraction | ISO install | Clone public GitHub repos + read PKGBUILD/pacman.conf | No ISO needed; all config is in public repos |

---

## Common Pitfalls

### Pitfall 1: Conflating install phases with opinion categories

**What goes wrong:** Treating all scripts in `install/config/` as "config opinions" and all scripts in `install/packaging/` as "package opinions". Many config scripts actually install packages (e.g. `docker.sh` enables a socket service, `kernel-modules-hook.sh` installs a package).
**Why it happens:** The phase names suggest category but the scripts are multi-category.
**How to avoid:** Classify each opinion by what it *does* (its category), not which install phase directory it lives in.
**Warning signs:** An opinion inventory entry that has a category not matching its source directory.

### Pitfall 2: Treating omarchy-base.packages as a single atomic opinion

**What goes wrong:** Recording the 155-package base list as one OM-NNN entry, making the schema floor appear simpler than it is.
**Why it happens:** The script installs them all at once via `omarchy-pkg-add`.
**How to avoid:** Group packages into logical opinions (compositor, terminal, browser stack, AI tools, media tools, etc.) that curators would actually compose separately.
**Warning signs:** If a proposed point can only be expressed by including the entire base list.

### Pitfall 3: Missing the first-run phase

**What goes wrong:** Omitting the 12 first-run scripts from the inventory because they are deferred and not in `install.sh`'s main pipeline.
**Why it happens:** `first-run/` has no `all.sh` in the standard pipeline; it runs post-first-login.
**How to avoid:** Enumerate `install/first-run/*.sh` explicitly as a separate ordered walk step.
**Warning signs:** Firewall opinion, DNS resolver opinion, and GDK scale opinion missing from inventory.

### Pitfall 4: Assuming CachyOS/Garuda repo configs can be cloned without auth

**What goes wrong:** The primary GitLab source for Garuda is `gitlab.com/garuda-linux/` — the GitHub repos are read-only mirrors.
**Why it happens:** The delta study references both GitHub and GitLab.
**How to avoid:** Use GitHub mirrors (`github.com/garuda-linux/pkgbuilds`) for clone access; they are public and up to date.
**Warning signs:** 403 errors fetching GitLab URLs without auth.

### Pitfall 5: Counting migrations as opinions

**What goes wrong:** Attempting to enumerate all 313 migrations as OM-NNN entries.
**Why it happens:** Migrations represent change over time but most are idempotent catch-up scripts, not stable opinions.
**How to avoid:** Sample 10–15 migrations to understand the pattern; record the count and pattern as an open question about schema evolution, not as inventory entries.
**Warning signs:** Inventory growing unmanageably large due to migrations.

### Pitfall 6: Not recording the commit hash in every deliverable

**What goes wrong:** Evidence becomes non-reproducible; reviewers cannot trace a claim back to source.
**Why it happens:** Easy to forget after the initial clone.
**How to avoid:** The commit hash `9cf1852525a5f7de26d3162db9d61e2f5c1d5523` must appear in the header of every `research/*.md` file.
**Warning signs:** Any deliverable that references "the Omarchy source" without a pinned commit.

---

## Code Examples

### Walking the install pipeline in order

```bash
# Verified pattern: all.sh files are the authoritative ordered script lists
# Source: direct analysis of install/config/all.sh, install/packaging/all.sh

# Extract ordered script list from an all.sh
grep "run_logged" /tmp/omarchy/install/config/all.sh | sed 's/run_logged //'

# Count total scripts per phase
find /tmp/omarchy/install -name "*.sh" -not -name "all.sh" -not -name "helpers*" | sort
```

### Extracting base packages (confirmed pattern from base.sh)

```bash
# Source: install/packaging/base.sh
grep -v '^#' /tmp/omarchy/install/omarchy-base.packages | grep -v '^$'
```

### Identifying hardware-conditional scripts

```bash
# Hardware conditions always use omarchy-hw-* helpers or inline lspci/grep patterns
grep -rn "omarchy-hw-\|lspci\|dmi\|dmidecode\|/sys/class/dmi" \
  /tmp/omarchy/install/config/hardware/ | grep "if "
```

### Identifying services enabled

```bash
# Service enables use systemctl or chrootable_systemctl_enable
grep -rn "systemctl enable\|chrootable_systemctl_enable" /tmp/omarchy/install/ \
  | grep -v "^Binary"
```

### Identifying sysctl parameters

```bash
grep -rn "sysctl\|/etc/sysctl.d\|/etc/modprobe.d\|/etc/modules-load.d" \
  /tmp/omarchy/install/config/
```

---

## Validation Architecture

How to verify each deliverable is mechanically complete:

### omarchy-opinion-inventory.md

```bash
# Every script referenced in an all.sh must have ≥1 OM-NNN entry citing it
# Generate the authoritative reference list:
grep "run_logged" /tmp/omarchy/install/*/all.sh | sed 's/.*run_logged //' | sort > /tmp/expected-scripts.txt

# Cross-check against opinion inventory:
grep -oP 'source: install/[^\s]+' research/omarchy-opinion-inventory.md | sort | uniq > /tmp/covered-scripts.txt
diff /tmp/expected-scripts.txt /tmp/covered-scripts.txt
# Empty diff = complete coverage
```

```bash
# Every OM-NNN entry must have all required fields
grep -c "^## OM-" research/omarchy-opinion-inventory.md   # count entries
grep -c "category:" research/omarchy-opinion-inventory.md  # must equal count
grep -c "intent:" research/omarchy-opinion-inventory.md    # must equal count
grep -c "source:" research/omarchy-opinion-inventory.md    # must equal count
```

### omarchy-points.md

```bash
# Every OM-NNN from the inventory must appear in exactly one point
grep -oP 'OM-\d+' research/omarchy-opinion-inventory.md | sort > /tmp/all-opinions.txt
grep -oP 'OM-\d+' research/omarchy-points.md | sort > /tmp/grouped-opinions.txt
diff /tmp/all-opinions.txt /tmp/grouped-opinions.txt
# Opinions in inventory but not in any point = ungrouped (flag as error)
# Opinions in points but not in inventory = phantom (flag as error)
```

### schema-requirements.md

```bash
# Every schema requirement must cite at least one OM-NNN or EC-NNN
grep -c "SR-" research/schema-requirements.md   # count requirements
grep -c "OM-\|EC-" research/schema-requirements.md  # should be ≥ count
```

### resolver-edge-cases.md

```bash
# Coverage matrix check: each docs/04 rule must map to ≥1 EC-NNN
# Rules to check:
# - required-beats-nice-to-have
# - required-vs-required
# - required-vs-required-with-patch
# - nice-vs-nice-default
# - patch-overrides-hierarchy
# - ordering-toposort
# - cycle-detection
# - hardware-conditional

grep -c "EC-" research/resolver-edge-cases.md    # count scenarios
grep "required beats" research/resolver-edge-cases.md    # must be present
grep "hard conflict" research/resolver-edge-cases.md     # must be present
grep "toposort\|ordering" research/resolver-edge-cases.md # must be present
grep "cycle" research/resolver-edge-cases.md              # must be present
```

### arch-variants-delta.md

```bash
# Both variants must be present with repo, keyring, kernel, default-fs, and pre-seeded opinions
grep "CachyOS\|cachyos" research/arch-variants-delta.md | wc -l    # > 0
grep "Garuda\|garuda" research/arch-variants-delta.md | wc -l      # > 0
grep "keyring" research/arch-variants-delta.md | wc -l             # ≥ 2 (one per variant)
grep "variant-profile" research/arch-variants-delta.md | wc -l     # ≥ 1
```

---

## State of the Art

| Old Approach | Current Approach | Notes |
|--------------|-----------------|-------|
| GRUB only | Limine + GRUB (both supported) | Omarchy 4.x uses Limine; Garuda uses GRUB — both are active |
| mkinitcpio UKI disabled during install | mkinitcpio disabled with `disable-mkinitcpio.sh`, then re-enabled in login phase | Preflight step required before packaging |
| Single Omarchy repo channel | Three channels: stable, rc, edge | Mirror URL embedded in mirrorlist file |
| btrfs quotas enabled | btrfs quotas disabled | Omarchy explicitly disables quota accounting for performance |
| Manual SDDM config | autologin preconfigured, PAM patched | PAM modification to prevent encrypted keyring creation |

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Garuda boots GRUB by default (not Limine) | Variant Delta Study | Low — multiple sources confirm; affects EC-002 (bootloader collision) |
| A2 | CachyOS does not pre-seed a custom snapper config | Variant Delta Study | Low — if wrong, adds another collision class to edge-case corpus |
| A3 | Garuda `garuda-common-settings` configures snapper automatically | Variant Delta Study | Medium — PKGBUILD dependencies confirm snapper/btrfsmaintenance but exact snapper config paths unverified |
| A4 | CachyOS-Settings writes sysctl.d drop-ins that may overlap with Omarchy's | Variant Delta Study | Low — confirmed etc/ tree exists; specific files not enumerated |
| A5 | `install/first-run/` scripts have no `all.sh` orchestrator | Pipeline Walk Order | Low — verified by `ls`; mitigated by explicit enumeration in walk order |
| A6 | Omarchy version `4.0.0.alpha` is the current "stable" state (no tags in repo) | Pinning | Medium — alpha version; may change. Pinned to commit hash which is stable |
| A7 | Garuda does not use a custom kernel (uses standard linux or linux-zen) | Variant Delta Study | Low — confirmed by multiple sources; no Garuda kernel PKGBUILD found |

---

## Open Questions (for open-questions.md)

1. **Tag pinning:** The Omarchy repo has no git tags (git describe fails). The version file reads `4.0.0.alpha`. Should the inventory cite commit hash only, or also attempt to track a branch/tag once stable tagging resumes?

2. **Migrations as schema concept:** 313 timestamped migrations represent an applied-delta system that lets an existing Omarchy install stay current. Does the DebateOS schema need a migration concept, or is version pinning of opinions sufficient? (Recommend: record as open question in open-questions.md, defer to Phase 1.)

3. **Omarchy custom repo as a required opinion:** Installing Omarchy requires adding `pkgs.omarchy.org` to pacman.conf and importing its GPG key. Is "add the Omarchy repo" itself an opinion in the inventory, or is it a foundation prerequisite? (Lean toward: it is an opinion, with category `custom-repo`, and its translator capability requirement is "add signed repo to package manager".)

4. **npm-installed AI tools:** Omarchy installs `codex`, `gemini`, `copilot`, `opencode`, `playwright`, `pi`, `ghui`, `hunk` via global npm. On a Debian foundation, npm is a separate install step. Does the schema need a `runtime-tool-install` category that is distinct from OS packages? (Recommend: yes, record as schema surprise SR-NNN.)

5. **First-run vs install-time ordering:** Some opinions in `install/first-run/` are deferred because they need a live Wayland session. On a headless CI/VM build, these would never run. The schema needs an `execution-context` or `install-phase` field to express this distinction. (Recommend: add `execution-phase: first-run` as a metadata field, flag as schema surprise.)

6. **CachyOS CPU arch detection:** Selecting between v3 and v4 repo variants requires CPU capability detection at install time. The schema must either bake this into the variant profile (static declaration) or support a runtime hardware condition. The docs/08 "variant profile" approach suggests static declaration — record as open question about dynamic vs static variant selection.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| git | Clone Omarchy + variant repos | Yes | 2.43.0 | — |
| curl | Fetch public files if needed | Yes | 8.5.0 | — |
| python3 | Optional analysis scripting | Yes | 3.12.3 | bash |
| jq | JSON processing for any structured output | Yes | 1.7 | — |
| grep | Script analysis patterns | Yes | ugrep 7.5.0 | — |
| Internet access | Clone Omarchy, CachyOS, Garuda repos | Assumed (clone succeeded) | — | — |
| Local disk (/tmp) | Scratch space for cloned repos | Available | ~50MB per repo | /var/tmp |

**Omarchy already cloned:** `/tmp/omarchy` at commit `9cf1852525a5f7de26d3162db9d61e2f5c1d5523`.

**CachyOS repos to clone for variant study:**
```bash
git clone --depth=1 https://github.com/CachyOS/CachyOS-PKGBUILDS /tmp/cachyos-pkgbuilds
git clone --depth=1 https://github.com/CachyOS/linux-cachyos /tmp/cachyos-kernel
git clone --depth=1 https://github.com/CachyOS/CachyOS-Settings /tmp/cachyos-settings
```

**Garuda repos to clone for variant study:**
```bash
git clone --depth=1 https://github.com/garuda-linux/pkgbuilds /tmp/garuda-pkgbuilds
git clone --depth=1 https://github.com/garuda-linux/garuda-tools /tmp/garuda-tools
```

**Missing dependencies with no fallback:** None — all required tools available.

---

## Validation Architecture

> `workflow.nyquist_validation` key absent from config.json — treated as enabled.

### Phase 0 Deliverables as Validation Units

Phase 0 produces documents, not code. "Tests" for this phase are mechanical completeness checks that can be run as shell one-liners after each deliverable is written.

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| RSCH-01 | Every all.sh-referenced script has ≥1 OM-NNN entry | completeness-check | `diff <(grep "run_logged" /tmp/omarchy/install/*/all.sh \| sed 's/.*run_logged //') <(grep -oP 'source: install/[^\s]+' research/omarchy-opinion-inventory.md \| sort -u)` | No — Wave 0 |
| RSCH-01 | Every OM-NNN entry has all 6 required fields | field-check | `for f in category intent source dependencies ordering translator-capability; do echo "$f: $(grep -c "^$f:" research/omarchy-opinion-inventory.md)"; done` | No — Wave 0 |
| RSCH-01 | Every OM-NNN in inventory appears in exactly one point | cross-ref | `diff <(grep -oP 'OM-\d+' research/omarchy-opinion-inventory.md \| sort) <(grep -oP 'OM-\d+' research/omarchy-points.md \| sort)` | No — Wave 0 |
| RSCH-02 | Both CachyOS and Garuda sections present in delta | presence-check | `grep -c "CachyOS" research/arch-variants-delta.md && grep -c "Garuda" research/arch-variants-delta.md` | No — Wave 0 |
| RSCH-02 | Variant-profile YAML sketch present | syntax-check | `grep -A20 "variant-profile" research/arch-variants-delta.md \| grep -c "repos:"` | No — Wave 0 |
| RSCH-03 | All 8 docs/04 rules have ≥1 EC-NNN scenario | coverage-matrix | `for rule in "required beats" "hard conflict" "patch overrides" "nice.*nice" "toposort\|ordering" "cycle" "hardware-conditional"; do echo "$rule: $(grep -Ec "$rule" research/resolver-edge-cases.md)"; done` | No — Wave 0 |
| RSCH-03 | All 3 variant collision classes have ≥1 EC-NNN | coverage-check | `grep -c "CachyOS" research/resolver-edge-cases.md && grep -c "Garuda" research/resolver-edge-cases.md` | No — Wave 0 |

### Wave 0 Gaps

No code infrastructure needed — all checks are shell one-liners. The `research/` directory must exist at repo root before deliverables are written:

- [ ] `research/` directory — create at repo root
- [ ] `research/omarchy-opinion-inventory.md`
- [ ] `research/omarchy-points.md`
- [ ] `research/schema-requirements.md`
- [ ] `research/open-questions.md`
- [ ] `research/arch-variants-delta.md`
- [ ] `research/resolver-edge-cases.md`

---

## Security Domain

> `security_enforcement` not set in config.json — treated as enabled.

Phase 0 is a research-and-document phase with no code, no deployed services, and no external API calls beyond public git clones. Security considerations are minimal and confined to research hygiene:

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | — |
| V3 Session Management | No | — |
| V4 Access Control | No | — |
| V5 Input Validation | No | — |
| V6 Cryptography | No | — |

### Security-Adjacent Research Findings

These are NOT runtime security concerns for Phase 0, but they produce schema requirements documented in `schema-requirements.md` and `open-questions.md`:

- **GPG key handling in package repos:** Omarchy imports its own GPG key (`40DFB630FF42BCFFB047046CF0134EE680CAC571`) during install. CachyOS requires `cachyos-keyring`. Garuda requires `chaotic-keyring`. The schema must express "install this keyring before using this repo" as an ordering constraint on the `custom-repo` opinion category. This is a translator-capability requirement, not a schema security boundary.
- **`SigLevel = Never` for arch-mact2 repo:** Omarchy adds an unsigned repo for Apple T2 hardware. The schema should record this as a per-repo trust level in the variant-profile / `custom-repo` opinion — and the planner/user should be informed that unsigned repos are explicitly flagged.
- **`omarchy-keyring` with `SigLevel = Optional TrustAll`:** The Omarchy repo itself uses `Optional TrustAll` rather than full signature verification. This should be documented in the opinion metadata.

---

## Sources

### Primary (HIGH confidence — direct clone analysis)
- `https://github.com/basecamp/omarchy` commit `9cf1852525a5f7de26d3162db9d61e2f5c1d5523` — full install pipeline, package lists, config scripts, hardware scripts, theme bundles, migrations

### Secondary (LOW confidence — web sources)
- `https://github.com/CachyOS/docker/blob/master/pacman-v3.conf` — CachyOS pacman.conf repo ordering for v3
- `https://github.com/CachyOS/CachyOS-PKGBUILDS` — PKGBUILDs including cachyos-settings structure
- `https://github.com/garuda-linux/pkgbuilds` — Garuda pkgbuilds structure
- `https://wiki.cachyos.org/features/optimized_repos/` — CachyOS repo configuration guidance
- `https://gadgeteer.co.za/garuda-linux-rolling-release-distro-based-arch-linux-btrfs-default-filesystem-easy-rollbacks-grub/` — Garuda default filesystem + bootloader

---

## Metadata

**Confidence breakdown:**
- Omarchy source analysis: HIGH — cloned directly and analyzed; all claims from actual files
- Omarchy opinion count estimate: HIGH — derived from actual script counts
- CachyOS variant delta: LOW — repo structure confirmed, specific config file contents partially inferred
- Garuda variant delta: LOW — repo structure confirmed, specific config file contents partially inferred; PKGBUILD dependency lists give strong signal
- Resolver edge-case classes: MEDIUM — classes 1–4 are evidence-backed; classes 5–6 are synthesized from confirmed variant data

**Research date:** 2026-06-12
**Valid until:** 2026-07-12 (Omarchy is under active development; commit hash pins the evidence; variant repos change at release cadence)
