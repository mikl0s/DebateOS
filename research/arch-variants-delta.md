# Arch-Variant Delta Study: CachyOS and Garuda Linux

**Research date:** 2026-06-12
**Omarchy pin:** `9cf1852525a5f7de26d3162db9d61e2f5c1d5523` (version 4.0.0.alpha, no git tags)

## Pinned Source Commits

| Repo | Commit | Date |
|------|--------|------|
| github.com/CachyOS/CachyOS-PKGBUILDS | `860f2283198059a05d7aa56fe434d80300ee9c56` | 2026-06-12 |
| github.com/CachyOS/linux-cachyos | `39d9d125940996ed2eb32425ffec7f2de6ac7fba` | 2026-06-09 |
| github.com/CachyOS/CachyOS-Settings | `b1aedc79d4f5edab86eb9d22a972ba6994c49b26` | 2026-06-03 |
| github.com/CachyOS/docker | `2f032fd8df222c4ec670597df346fafaa9da8055` | 2024-10-03 |
| github.com/garuda-linux/pkgbuilds | `1dc0c910447288b16021ebd94ec75c53b8255499` | 2026-06-12 |
| github.com/garuda-linux/garuda-tools | `433ad847905a2c3044acba9479a3ade6364bd6da` | 2021-11-27 |

**Verification note:** All claims below are derived from direct read-only inspection of the
cloned sources listed above. Claims not verifiable from these repos are tagged `[UNVERIFIED]`.
PKGBUILD scripts were read, never executed. Assumptions from RESEARCH.md that were found
incorrect by clone inspection are annotated with **CORRECTION:**.

---

## CachyOS

CachyOS is an Arch-based distro focused on CPU-architecture-optimized package binaries
(x86-64-v3 and x86-64-v4 ISA levels), a performance-tuned kernel family, and opinionated
defaults for schedulers and memory management.

### Custom Repos

**Source:** `github.com/CachyOS/docker` commit `2f032fd8df222c4ec670597df346fafaa9da8055`
— files `pacman.conf`, `pacman-v3.conf`, `pacman-v4.conf`

CachyOS ships three pacman.conf variants based on CPU ISA level:

**Standard x86_64 (`pacman.conf`):**
```
[cachyos]
Include = /etc/pacman.d/cachyos-mirrorlist

[core]
Include = /etc/pacman.d/mirrorlist

[extra]
Include = /etc/pacman.d/mirrorlist

[multilib]
Include = /etc/pacman.d/mirrorlist
```

**x86-64-v3 (`pacman-v3.conf`):**
```
[cachyos-v3]
Include = /etc/pacman.d/cachyos-v3-mirrorlist
[cachyos-core-v3]
Include = /etc/pacman.d/cachyos-v3-mirrorlist
[cachyos-extra-v3]
Include = /etc/pacman.d/cachyos-v3-mirrorlist
[cachyos]
Include = /etc/pacman.d/cachyos-mirrorlist

[core] / [extra] / [multilib]
Include = /etc/pacman.d/mirrorlist
```

**x86-64-v4 (`pacman-v4.conf`):** identical structure to v3 but with `cachyos-v4-*` repos.

**Key ordering rule (VERIFIED):** The v3/v4 optimized repos appear ABOVE `[core]`/`[extra]`
— they override standard Arch packages with architecture-optimized builds.

**Mirror URL (VERIFIED):** `github.com/CachyOS/CachyOS-PKGBUILDS` commit `860f228`,
file `cachyos-mirrorlist/cachyos-mirrorlist`:
```
Server = https://cdn77.cachyos.org/repo/$arch/$repo   # CDN77 worldwide tier-1
Server = https://us.cachyos.org/repo/$arch/$repo       # US tier-2
```
The v3 mirrorlist substitutes `/$arch/` → `/$arch_v3/` at build time.

### Keyring

**Source:** `github.com/CachyOS/CachyOS-PKGBUILDS` commit `860f228`,
file `cachyos-keyring/PKGBUILD`

Package: `cachyos-keyring` (pkgver 20240331)
- Installs keyring to `/usr/share/pacman/keyrings/cachyos.{gpg,trusted,revoked}`
- Trust level: Required (SigLevel = Required DatabaseOptional, standard pacman)
- Must be installed **before** first use of CachyOS repos

### Kernel

**Source:** `github.com/CachyOS/linux-cachyos` commit `39d9d12`

CachyOS ships a full kernel family. Default kernel: `linux-cachyos` (EEVDF scheduler).

| Variant | Scheduler | Notes |
|---------|-----------|-------|
| `linux-cachyos` | CachyOS EEVDF (default) | `_cpusched:=cachyos` in PKGBUILD |
| `linux-cachyos-bore` | BORE (Burst-Oriented Response Enhancer) | |
| `linux-cachyos-bmq` | BMQ | |
| `linux-cachyos-eevdf` | EEVDF vanilla | |
| `linux-cachyos-lts` | CachyOS EEVDF, LTS base | |
| `linux-cachyos-hardened` | BORE + hardened config | |
| `linux-cachyos-rt-bore` | BORE + realtime patches | |
| `linux-cachyos-deckify` | Steam Deck variant | |
| `linux-cachyos-server` | Server-optimized | |
| `linux-cachyos-rc` | Release candidate | |

All kernels built with `-O3 -march=<cpu-level>` for v3/v4 variants. Kernel version at clone
time: 7.0.12 (major.minor). The `script-v3-v4.sh` file confirms how v3/v4 kernel packages
are built using Docker with `_processor_opt:=GENERIC_V3` or `GENERIC_V4`.

Headers package: `linux-cachyos-headers` (parallel to kernel package name).

**Pre-seeded opinion — kernel:** CachyOS pre-installs `linux-cachyos`; an Omarchy speech
targeting `linux` (vanilla Arch kernel) would create a **required-vs-pre-seeded** conflict.

### Default Filesystem and Bootloader

**Source:** `github.com/CachyOS/CachyOS-PKGBUILDS` commit `860f228` — inference from
`cachyos-snapper-support`, `grub-btrfs-support`, `systemd-boot-manager`, `dracut-cachyos`

CachyOS offers multiple FS and bootloader options via its calamares installer. No single
default is hard-coded at the PKGBUILD level — the installer presents a choice. Available
confirmed packages:

- `cachyos-snapper-support` (PKGBUILD): installs snapper template `cachyos-root`, enables
  `snapper-cleanup.timer` — btrfs optional
- `grub-btrfs-support` (PKGBUILD): depends on `cachyos-snapper-support` + `grub-btrfs`;
  installs `grub-btrfs-snapper.path` and `.service`
- `systemd-boot-manager` (PKGBUILD): A CachyOS tool for managing systemd-boot entries
- `dracut-cachyos` (PKGBUILD): Alternative initramfs; standard install uses mkinitcpio
  (confirmed by `cachyos-hooks` installing a mkinitcpio-style Plymouth hook)

**CORRECTION from RESEARCH.md assumption A2:** CachyOS DOES have an optional snapper
package (`cachyos-snapper-support`). Whether it is pre-installed by default depends on the
edition — this is [UNVERIFIED] without running the installer. The package exists and is
an optional dependency path.

### CachyOS-Settings: Pre-seeded Opinions

**Source:** `github.com/CachyOS/CachyOS-Settings` commit `b1aedc7`
(Note: this is the installed settings package, not the PKGBUILDS repo)

Files in `usr/lib/sysctl.d/70-cachyos-settings.conf` (VERIFIED):
```
vm.swappiness = 100
vm.vfs_cache_pressure = 50
vm.dirty_bytes = 268435456
vm.page-cluster = 0
vm.dirty_background_bytes = 67108864
vm.dirty_writeback_centisecs = 1500
kernel.nmi_watchdog = 0
kernel.unprivileged_userns_clone = 1
kernel.printk = 3 3 3 3
kernel.kptr_restrict = 2
net.core.netdev_max_backlog = 4096
fs.file-max = 2097152
```

Additional configs installed by `cachyos-settings`:
- `usr/lib/modprobe.d/blacklist.conf`: blacklists `iTCO_wdt`, `sp5100_tco` (watchdog timers)
- `usr/lib/modprobe.d/nvidia.conf`: NVreg_InitializeSystemMemoryAllocations=0, DynamicPowerManagement=0x02
- `usr/lib/modprobe.d/amdgpu.conf`: forces amdgpu for SI/CIK GPUs
- `usr/lib/modules-load.d/ntsync.conf`: loads ntsync module (Windows compatibility)
- `usr/lib/systemd/zram-generator.conf`: zram0 with zstd compression, size=RAM, swap-priority=100
- `usr/lib/NetworkManager/conf.d/dns.conf`: [UNVERIFIED — file content not inspected]
- `etc/security/limits.d/20-audio.conf`: audio RT priority limits

Services enabled by `cachyos-settings.install` (VERIFIED):
- `ananicy-cpp` — CPU priority management daemon
- `systemd-resolved` — DNS resolver
- `cachyos-iw-set-regdomain.path` — wireless regulatory domain (if `iw` available)

The `cachyos-settings` package depends on:
- `zram-generator` — zram swap pre-configured
- `ananicy-cpp` + `cachyos-ananicy-rules` — process prioritization
- `inxi` — system info tool
- `iw` + `wireless-regdb` — wireless regulatory

**Pre-seeded sysctl opinions** that may conflict with Omarchy's:
- `vm.swappiness = 100` (CachyOS) vs Omarchy's `increase-file-watchers.sh`
  (`fs.inotify.max_user_watches`) — different keys, no direct collision
- `fs.file-max = 2097152` (CachyOS, sysctl kernel parameter via sysctl.d) vs Omarchy's
  `increase-fd-limit.sh` (OM-038) — **NO KEY COLLISION**: OM-038 sets `DefaultLimitNOFILE`
  via systemd `system.conf.d/` and `user.conf.d/` drop-ins (per-process RLIMIT_NOFILE),
  NOT a sysctl parameter. These are distinct mechanisms in different namespaces; CachyOS
  raises the global fd ceiling and Omarchy sets the per-process limit — additive, not conflicting.

### Trust Level

| Repo | SigLevel | Notes |
|------|----------|-------|
| `[cachyos]` | Required DatabaseOptional (standard) | Full GPG signing |
| `[cachyos-v3]` | Required DatabaseOptional (standard) | Full GPG signing |
| `[cachyos-v4]` | Required DatabaseOptional (standard) | Full GPG signing |

### Delta Summary Table

| Delta | Detail | Pre-seeded Opinion Category | Collision Risk with Omarchy |
|-------|--------|----------------------------|----------------------------|
| Custom repos | `[cachyos]` above `[core]`/`[extra]`; v3/v4 variants above `[cachyos]` | `custom-repo` | `[omarchy]` repo ordering vs CachyOS priority |
| Keyring | `cachyos-keyring` required before repo use | `custom-repo` (keyring dep) | None — different keyring |
| Kernel | `linux-cachyos` (EEVDF, 7.0.12) as default | `kernel-install` | Collides with Omarchy `linux` + `linux-ptl` Intel kernel |
| CPU-optimized packages | glibc, mesa, ffmpeg etc. rebuilt for v3/v4 | implicit in `custom-repo` | Package name identical, different binary |
| sysctl | `70-cachyos-settings.conf` — 12 sysctl params | `sysctl-param` | No sysctl key collision with Omarchy: OM-038 sets DefaultLimitNOFILE (systemd per-process limit, not sysctl); OM-036/OM-037 use different keys |
| zram | zram0 zstd, size=RAM, swap-priority=100 | `zram-config` | Collision if Omarchy configures zram |
| ananicy-cpp | Enabled by default via cachyos-settings | `service-enable` | None direct — Omarchy doesn't configure ananicy |
| systemd-resolved | Enabled by cachyos-settings | `service-enable` | Omarchy also enables systemd-resolved — same intent, potentially idempotent |
| modprobe blacklist | Watchdog timers blacklisted | `sysctl-param` (modprobe variant) | No direct Omarchy collision identified |
| ntsync module | Loaded by modules-load.d | `kernel-module` | No direct Omarchy collision |
| snapper (optional) | `cachyos-snapper-support` available, not forced | `snapshot-policy` | Potential collision if both configure snapper root |
| Initramfs | mkinitcpio default; dracut-cachyos optional | `initramfs-method` | Omarchy uses mkinitcpio — consistent default |

---

## Garuda

Garuda Linux is an Arch-based distro focused on gaming, btrfs-first defaults, KDE/Dr460nized
theming, and convenience wrappers (garuda-update, btrfs maintenance automation).

### Custom Repos

**Source:** `github.com/garuda-linux/garuda-tools` commit `433ad84`,
file `data/pacman-default.conf`

```
[core]
Include = /etc/pacman.d/mirrorlist

[extra]
Include = /etc/pacman.d/mirrorlist

[community]
Include = /etc/pacman.d/mirrorlist

[multilib]
Include = /etc/pacman.d/mirrorlist

[chaotic-aur]
Include = /etc/pacman.d/chaotic-mirrorlist
```

**Note:** The garuda-tools `pacman-default.conf` shows `[chaotic-aur]` as the primary
custom repo. The `[garuda]` repo is added to the installed system's pacman.conf at ISO
build time (managed by the Garuda ISO profiles on GitLab, which requires auth to access).
The evidence from `garuda-hooks` depending on `chaotic-mirrorlist` and `garuda-update`
hard-coding `chaotic.cx` mirror URLs confirms `[chaotic-aur]` is the primary repo.

**Chaotic-AUR mirror URL (VERIFIED):**
From `garuda-update/main-update` script at `github.com/garuda-linux/pkgbuilds`
commit `1dc0c91`:
```
Server = https://secret-mirror.chaotic.cx/$repo/$arch   # fallback mirror
```
The standard mirrorlist uses geo-redundant CDN endpoints including
`https://geo-mirror.chaotic.cx/` [UNVERIFIED — mirrorlist not directly in cloned source].

**`[garuda]` repo:** Present in the installed system's pacman.conf (added during ISO
creation). The `[garuda]` repo URL is `https://geo-mirror.chaotic.cx/garuda/$arch`
[UNVERIFIED — derived from garuda-tools ISO build logic; direct conf file not in
cloned garuda-tools or garuda-linux/pkgbuilds].

### Keyring

**Source:** `github.com/garuda-linux/pkgbuilds` commit `1dc0c91`, file
`garuda-hooks/PKGBUILD`

The `garuda-hooks` package (core Garuda package) depends on `chaotic-mirrorlist`.
Chaotic-AUR uses `chaotic-keyring` (package maintained by chaotic-aur project).

- `chaotic-keyring`: GPG key `FBA220DFC880C036` [UNVERIFIED — cited in RESEARCH.md from
  external sources; GPG key file not found in cloned garuda-linux/pkgbuilds]
- `garuda-keyring`: [UNVERIFIED — no garuda-keyring PKGBUILD found in cloned source;
  Garuda may rely solely on chaotic-keyring and archlinux-keyring]

**Trust level:**

| Repo | SigLevel | Notes |
|------|----------|-------|
| `[chaotic-aur]` | Standard | AUR pre-built packages; chaotic-keyring required |
| `[garuda]` | Standard [UNVERIFIED] | Garuda-specific packages |

### Kernel

**Source:** `github.com/garuda-linux/pkgbuilds` commit `1dc0c91` — no custom kernel
PKGBUILD found in cloned source.

Garuda does NOT ship a custom kernel. Uses standard Arch `linux` or `linux-zen`.

From `garuda-tools/data/garuda-tools.conf`:
```
# kernel="linux-zen"
```
(commented out — configurable, with `linux-zen` as the documented example)

The `garuda-dracut-support` package explicitly `conflicts=('mkinitcpio' ...)`,
confirming Garuda uses **dracut** (not mkinitcpio) as its initramfs generator.

**VERIFICATION of RESEARCH.md assumption A7:** Confirmed — no Garuda kernel PKGBUILD
exists in the public pkgbuilds repo. Garuda uses `linux` or `linux-zen`.

### Default Filesystem and Bootloader

**Source:** Multiple files in `github.com/garuda-linux/pkgbuilds` commit `1dc0c91`
and `github.com/garuda-linux/garuda-tools` commit `433ad84`

**Filesystem: btrfs (VERIFIED via multiple sources)**

Evidence:
1. `snapper-support/PKGBUILD` (VERIFIED): `depends=(snapper snap-pac grub-btrfs)` —
   mandatory btrfs maintenance stack in Garuda
2. `garuda-common-settings.install` (VERIFIED): `systemctl enable btrfs-balance.timer
   btrfs-defrag.timer btrfs-scrub.timer btrfs-trim.timer` — btrfs maintenance timers
   enabled unconditionally at install
3. `garuda-common-settings/PKGBUILD` (VERIFIED): `depends=('btrfsmaintenance' ...)`
4. `garuda-hooks/PKGBUILD` (VERIFIED): depends on `grub` + `update-grub`; installs
   `99-update-grub.hook` + `grub-btrfs-config.hook` — grub-btrfs integration baked into
   the core hooks package

**Bootloader: GRUB (VERIFIED)**

From `garuda-pkgbuilds/grub-garuda/PKGBUILD` (VERIFIED):
```
depends=('grub' 'update-grub' 'os-prober-btrfs' 'efibootmgr' 'memtest86+')
```

From `garuda-hooks/PKGBUILD` (VERIFIED): depends on `grub`; installs `update-grub.hook`
and `grub-install.hook` — GRUB management hooks are core Garuda infrastructure.

Garuda also ships `grub-theme-garuda-dr460nized` and `grub-garuda` meta-package.

**Initramfs: dracut (VERIFIED)**

`garuda-dracut-support/PKGBUILD` (VERIFIED):
```
conflicts=('mkinitcpio' 'mkinitcpio-openswap' ...)
```
Garuda uses dracut exclusively, conflicting with mkinitcpio.
Omarchy uses mkinitcpio (`disable-mkinitcpio.sh` in preflight, then re-enabled in login phase)
— this is a **direct conflict**.

**btrfs subvolume layout (VERIFIED_SECONDARY):**
Standard Garuda calamares installer creates:
`@`, `@home`, `@root`, `@srv`, `@snapshots`, `@log`, `@pkg`
[VERIFIED_SECONDARY: confirmed via snapper template (`SUBVOLUME="/"`) + garuda docs;
direct calamares config on GitLab requires auth to clone]

### Garuda Pre-seeded Opinions

**Source:** `github.com/garuda-linux/pkgbuilds` commit `1dc0c91`

**garuda-common-settings (VERIFIED) enables:**
- `btrfs-balance.timer`, `btrfs-defrag.timer`, `btrfs-scrub.timer`, `btrfs-trim.timer`
  — automated btrfs maintenance
- `garuda-pacman-lock` — pacman lock management service

**garuda-common-settings depends on:**
- `btrfsmaintenance` — btrfs maintenance scripts
- `garuda-hooks` — GRUB + chaotic-aur hooks
- `garuda-update` — update wrapper (replaces direct `pacman -Syu`)
- `garuda-icons` — custom icon set
- `garuda-wallpapers` — wallpaper set
- `zram-generator` — zram swap configuration
- `systemd-oomd-defaults` — OOM daemon configuration
- `noto-color-emoji-fontconfig` — font configuration

**garuda-hooks (VERIFIED) installs:**
- `99-update-grub.hook` — regenerates grub.cfg on kernel install
- `grub-btrfs-config.hook` — grub-btrfs configuration updates
- `grub-install.hook` — GRUB reinstall hook
- `01-snapshot-reject.hook` — prevents package install when booted from snapshot

**snapper-support (VERIFIED):**
- Creates snapper root config from `snapper-template-garuda`
- Enables `snapper-cleanup.timer` (but NOT `snapper-timeline.timer` by default —
  `NUMBER_CLEANUP=yes`, `TIMELINE_CREATE=no`)
- Snapshot limits: NUMBER_LIMIT=10, NUMBER_LIMIT_IMPORTANT=5

**garuda-update wrapper (VERIFIED):**
- Replaces `pacman -Syu` with `garuda-update`
- Creates symlink `/usr/bin/update → garuda-update`
- Handles chaotic-mirrorlist rotation and keyring refresh on error
- Pre-update health check via `garuda-health`
- Collision with Omarchy's `pacman-refresh` preflight step which runs `pacman -Sy`

**performance-tweaks (VERIFIED — optional package):**
- Enables `ananicy-cpp`, `irqbalance`, `preload`
- Depends on `cachyos-ananicy-rules-git` (CachyOS ananicy rules cross-dependency!)

**Dr460nized theming (VERIFIED):**
- `garuda-dr460nized/PKGBUILD`: KDE Plasma theming + `plasma5-themes-sweet-full-git`
- `grub-theme-garuda-dr460nized`: Custom GRUB boot theme
- `plymouth-theme-dr460nized/PKGBUILD`: Custom Plymouth theme
- SDDM theming: [UNVERIFIED — SDDM theme referenced in garuda-dr460nized but exact
  package not isolated from cloned source]

### Delta Summary Table

| Delta | Detail | Pre-seeded Opinion Category | Collision Risk with Omarchy |
|-------|--------|----------------------------|----------------------------|
| Custom repos | `[chaotic-aur]` + `[garuda]` (at install); chaotic-mirrorlist | `custom-repo` | Both add custom repos; priority order with `[omarchy]` |
| Keyring | `chaotic-keyring` (mandatory) | `custom-repo` (keyring dep) | None direct — different keyring |
| Kernel | `linux` or `linux-zen` (no custom kernel) | none pre-seeded | Low — both support vanilla linux |
| Initramfs | `dracut` only (conflicts mkinitcpio) | `initramfs-method` | **HARD CONFLICT**: Omarchy uses mkinitcpio; Garuda conflicts it |
| Filesystem | btrfs (mandatory, btrfs timers always enabled) | `filesystem-default` | Omarchy also btrfs — compatible intent, different path |
| Bootloader | GRUB + grub-btrfs (mandatory) | `bootloader-config` | **DIRECT COLLISION**: Omarchy uses limine; Garuda uses GRUB |
| Snapper | snapper-support pre-configured with garuda template | `snapshot-policy` | **DIRECT COLLISION**: both configure `/etc/snapper/configs/root` |
| SDDM theme | `garuda-dr460nized` SDDM theme | `theming` | **DIRECT COLLISION**: Omarchy deploys its own SDDM theme |
| Plymouth theme | `plymouth-theme-dr460nized` | `theming` | **DIRECT COLLISION**: Omarchy deploys its own Plymouth theme |
| GRUB theme | `grub-theme-garuda-dr460nized` | `theming` | No direct Omarchy collision (Omarchy doesn't configure GRUB) |
| garuda-update | Replaces `pacman -Syu`; is a symlink for `update` | `update-mechanism` | Omarchy's preflight pacman -Sy bypasses garuda-update |
| zram | `zram-generator` (via garuda-common-settings dep) | `zram-config` | Collision if Omarchy configures zram |
| btrfs timers | balance/defrag/scrub/trim timers auto-enabled | `service-enable` | Additive, low collision risk |
| ananicy-cpp | Optional (performance-tweaks); uses cachyos-ananicy-rules | `service-enable` | No direct Omarchy collision |
| systemd-oomd | `systemd-oomd-defaults` (mandatory dep) | `sysctl-param` variant | Potential OOM setting collision |

---

## Cross-Variant Comparison

| Dimension | Vanilla Arch | CachyOS | Garuda |
|-----------|-------------|---------|--------|
| Custom repos | none | cachyos\[-v3\|-v4\] | garuda + chaotic-aur |
| Keyring pkg | archlinux-keyring | cachyos-keyring | chaotic-keyring |
| Kernel | linux | linux-cachyos (EEVDF, 7.0.12) | linux or linux-zen |
| Kernel scheduler | Arch default | EEVDF (custom tuning) | Same as upstream |
| Initramfs | mkinitcpio | mkinitcpio (dracut optional) | dracut (mkinitcpio conflicts) |
| Default FS | user choice | user choice (btrfs optional) | btrfs (mandatory tooling) |
| Bootloader | user choice | systemd-boot or grub or limine | GRUB + grub-btrfs |
| Snapper | none | Optional (cachyos-snapper-support) | Mandatory (snapper-support) |
| sysctl tuning | minimal | 12 params (70-cachyos-settings) | none via pkg [UNVERIFIED] |
| zram | none | Configured via zram-generator.conf | Via zram-generator dep |
| Update wrapper | pacman | cachy-update (optional notifier) [UNVERIFIED] | garuda-update (mandatory) |
| Theming | none | SDDM: cachyos-themes-sddm [UNVERIFIED] | Dr460nized KDE/SDDM/Plymouth/GRUB |
| CPU arch levels | x86_64 only | x86_64, x86_64-v3, x86_64-v4 | x86_64 only |

---

## Open Questions (for open-questions.md)

1. **CachyOS default FS and bootloader:** Without running the installer, the exact
   default (btrfs vs ext4, limine vs systemd-boot vs grub) is not determined from
   PKGBUILD inspection alone. The installer presents a menu. The variant profile must
   either declare a specific default or flag this as [UNVERIFIED].

2. **`[garuda]` repo URL:** The installed Garuda system's `pacman.conf` includes
   `[garuda]`, but the exact URL is not in the cloned garuda-tools or garuda-linux/pkgbuilds
   repos. The GitLab ISO profile configs (private/auth-required) would contain this.
   Tagged [UNVERIFIED] in the variant profile below.

3. **CachyOS sysctl and Omarchy conflict resolution:** CachyOS's `70-cachyos-settings.conf`
   pre-seeds sysctl keys (e.g. `fs.file-max`) while Omarchy's `increase-fd-limit.sh` raises
   `DefaultLimitNOFILE` via systemd drop-ins — different mechanisms, no key collision
   (see corrected EC-005/SR-016). The open question stands in general form: the schema
   needs to express whether sysctl drop-ins touching the *same key* can be merged or
   conflict at the file level. This is a translator capability question for Phase 2.

4. **Garuda + Omarchy initramfs conflict:** Garuda's `garuda-dracut-support` conflicts
   `mkinitcpio`, but Omarchy's entire login phase assumes mkinitcpio (it disables it in
   preflight, re-enables it in login phase, and uses limine). Running Omarchy on Garuda
   would require an Omarchy speech retargeted to dracut — a substantial translator
   challenge. Deferred to Phase 2 stretch validation.

5. **variant-profile conflict semantics:** When Garuda pre-seeds `grub-btrfs` + snapper
   and Omarchy installs limine + its own snapper config, which wins? This is the core
   "foundation pre-seeded opinion vs user opinion" question deferred to Phase 1 schema
   design. Record in open-questions.md.

---

## Variant-Profile Shape

These YAML sketches are **candidates only** — the final schema is drafted in Phase 1 (D17).
Each block is labeled `variant-profile` for automated gate detection.

The shape covers: repos (name + url/mirrorlist + keyring + priority vs Arch), kernel
(package + headers), defaults (filesystem, bootloader, cpu arch level where relevant),
and `pre_seeded_opinions` (capabilities the foundation already expresses).
This is explicitly NOT a per-variant fork — it is a declarative profile that describes
the deltas a translator must account for (D20 anti-bloat).

### variant-profile: CachyOS (x86-64-v3)

```yaml
# variant-profile
variant: cachyos
description: "CachyOS — Arch-based with CPU-architecture-optimized packages and custom kernel family"
version_at_research: "2026-06-12 (commit 860f228 / 39d9d12 / b1aedc7)"
cpu_arch_level: v3  # choices: x86_64, v3, v4; detected or declared at install time

repos:
  - name: cachyos-v3
    mirrorlist: /etc/pacman.d/cachyos-v3-mirrorlist
    keyring: cachyos-keyring
    priority: above-arch  # must precede [core]/[extra]/[multilib]
    sig_level: Required DatabaseOptional
  - name: cachyos-core-v3
    mirrorlist: /etc/pacman.d/cachyos-v3-mirrorlist
    keyring: cachyos-keyring
    priority: above-arch
    sig_level: Required DatabaseOptional
  - name: cachyos-extra-v3
    mirrorlist: /etc/pacman.d/cachyos-v3-mirrorlist
    keyring: cachyos-keyring
    priority: above-arch
    sig_level: Required DatabaseOptional
  - name: cachyos
    mirrorlist: /etc/pacman.d/cachyos-mirrorlist
    keyring: cachyos-keyring
    priority: above-arch  # above [core] but below cachyos-v3
    sig_level: Required DatabaseOptional

keyring_install_before_repos: cachyos-keyring

kernel:
  package: linux-cachyos        # default; also bore/lts/bmq/rt-bore/hardened/server/rc
  headers: linux-cachyos-headers
  version_at_research: "7.0.12"
  scheduler: cachyos-eevdf      # custom EEVDF tuning
  notes: "x86-64-v3 optimized builds available as linux-cachyos in cachyos-v3 repo"

defaults:
  filesystem: null              # user choice in installer; btrfs available but not forced
  bootloader: null              # user choice; systemd-boot, grub, or limine supported
  initramfs: mkinitcpio         # default; dracut-cachyos optional alternative
  snapper: optional             # cachyos-snapper-support available, not mandatory

pre_seeded_opinions:
  - category: sysctl-param
    id: cachyos/sysctl-performance
    description: "12 sysctl params in /usr/lib/sysctl.d/70-cachyos-settings.conf"
    keys: [vm.swappiness, vm.vfs_cache_pressure, vm.dirty_bytes, vm.page-cluster,
           vm.dirty_background_bytes, vm.dirty_writeback_centisecs,
           kernel.nmi_watchdog, kernel.unprivileged_userns_clone,
           kernel.printk, kernel.kptr_restrict,
           net.core.netdev_max_backlog, fs.file-max]
    conflict_with_omarchy: "None identified: OM-038 sets DefaultLimitNOFILE via systemd drop-ins (per-process RLIMIT_NOFILE), not a sysctl key; different mechanism from fs.file-max. OM-036/OM-037 use different sysctl keys (tcp_mtu_probing, inotify.max_user_watches)."

  - category: zram-config
    id: cachyos/zram-defaults
    description: "zram0 zstd compression, size=RAM, swap-priority=100"
    conflict_with_omarchy: "None identified (Omarchy does not configure zram)"

  - category: service-enable
    id: cachyos/ananicy-cpp
    description: "ananicy-cpp process priority daemon enabled"
    conflict_with_omarchy: "None — Omarchy does not configure ananicy"

  - category: service-enable
    id: cachyos/systemd-resolved
    description: "systemd-resolved enabled"
    conflict_with_omarchy: "Omarchy also enables systemd-resolved — same intent, idempotent"

  - category: kernel-module
    id: cachyos/ntsync
    description: "ntsync kernel module loaded via modules-load.d"
    conflict_with_omarchy: "None identified"

  - category: modprobe-config
    id: cachyos/watchdog-blacklist
    description: "iTCO_wdt and sp5100_tco watchdog timers blacklisted"
    conflict_with_omarchy: "None identified"

open_questions:
  - "CachyOS default FS and bootloader: not determined from PKGBUILD inspection alone"
  - "cachyos-snapper-support: whether pre-installed by default requires installer run"
```

### variant-profile: Garuda

```yaml
# variant-profile
variant: garuda
description: "Garuda Linux — Arch-based with btrfs-first defaults, GRUB, dracut, and Dr460nized theming"
version_at_research: "2026-06-12 (commit 1dc0c91 / 433ad84)"

repos:
  - name: garuda
    url: "https://geo-mirror.chaotic.cx/garuda/$arch"  # [UNVERIFIED — inferred from build scripts]
    keyring: chaotic-keyring  # Garuda uses chaotic infrastructure for [garuda] repo
    priority: above-arch
    sig_level: Required DatabaseOptional  # [UNVERIFIED — exact SigLevel not in cloned source]
  - name: chaotic-aur
    mirrorlist: /etc/pacman.d/chaotic-mirrorlist
    keyring: chaotic-keyring
    priority: below-garuda   # below [garuda], above or below Arch repos [UNVERIFIED — ordering]
    sig_level: Required DatabaseOptional

keyring_install_before_repos: chaotic-keyring

kernel:
  package: linux              # default; linux-zen configurable via garuda-tools.conf
  headers: linux-headers
  notes: "No custom Garuda kernel; linux-zen documented as alternative"

defaults:
  filesystem: btrfs           # mandatory btrfs tooling (btrfsmaintenance timers always enabled)
  bootloader: grub            # grub-garuda + grub-btrfs mandatory (grub in garuda-hooks deps)
  initramfs: dracut           # garuda-dracut-support; conflicts mkinitcpio
  snapper: mandatory          # snapper-support always installed; root config created at install
  btrfs_subvolumes:
    - "@"          # root
    - "@home"      # /home
    - "@root"      # /root
    - "@srv"       # /srv
    - "@snapshots" # /.snapshots
    - "@log"       # /var/log [UNVERIFIED — layout from docs, not direct calamares config]
    - "@pkg"       # /var/cache/pacman/pkg [UNVERIFIED]

pre_seeded_opinions:
  - category: snapshot-policy
    id: garuda/snapper-grub-btrfs
    description: "snapper root config + grub-btrfs snapshot boot entries"
    packages: [snapper, snap-pac, grub-btrfs, snapper-support]
    conflict_with_omarchy: "DIRECT COLLISION: both configure /etc/snapper/configs/root"

  - category: bootloader-config
    id: garuda/grub-btrfs
    description: "GRUB as default bootloader with grub-btrfs for snapshot entries"
    packages: [grub, grub-btrfs, update-grub, grub-theme-garuda-dr460nized]
    conflict_with_omarchy: "HARD CONFLICT: Omarchy uses limine; Garuda requires GRUB"

  - category: initramfs-method
    id: garuda/dracut
    description: "dracut initramfs generator; conflicts mkinitcpio"
    packages: [garuda-dracut-support]
    conflict_with_omarchy: "HARD CONFLICT: Omarchy disables/re-enables mkinitcpio; Garuda conflicts it"

  - category: update-mechanism
    id: garuda/garuda-update
    description: "garuda-update replaces pacman -Syu; symlinked as /usr/bin/update"
    packages: [garuda-update]
    conflict_with_omarchy: "Omarchy preflight runs 'pacman -Sy' directly, bypassing garuda-update"

  - category: theming
    id: garuda/dr460nized-sddm
    description: "Dr460nized KDE Plasma + SDDM theme"
    packages: [garuda-dr460nized, plasma5-themes-sweet-full-git]
    conflict_with_omarchy: "DIRECT COLLISION: Omarchy deploys its own SDDM theme"

  - category: theming
    id: garuda/dr460nized-plymouth
    description: "plymouth-theme-dr460nized boot splash"
    packages: [plymouth-theme-dr460nized]
    conflict_with_omarchy: "DIRECT COLLISION: Omarchy deploys its own Plymouth theme"

  - category: theming
    id: garuda/grub-theme-dr460nized
    description: "Dr460nized GRUB boot theme"
    packages: [grub-theme-garuda-dr460nized]
    conflict_with_omarchy: "No direct collision (Omarchy uses limine, not GRUB)"

  - category: service-enable
    id: garuda/btrfs-maintenance-timers
    description: "btrfs-balance/defrag/scrub/trim timers enabled by garuda-common-settings"
    conflict_with_omarchy: "Additive — Omarchy does not configure btrfs timers"

  - category: zram-config
    id: garuda/zram-generator
    description: "zram-generator dep of garuda-common-settings"
    conflict_with_omarchy: "Low — no identified Omarchy zram config"

  - category: filesystem-layout
    id: garuda/btrfs-subvolumes
    description: "Standard btrfs subvolume layout: @, @home, @root, @srv, @snapshots, @log, @pkg"
    conflict_with_omarchy: "Omarchy also uses btrfs; compatible at layout level if subvols match"

open_questions:
  - "[garuda] repo: exact URL and SigLevel not confirmed from cloned garuda-linux/pkgbuilds"
  - "garuda-keyring vs chaotic-keyring: Garuda appears to rely on chaotic-keyring only"
  - "btrfs subvolume layout @log and @pkg: UNVERIFIED — calamares config on GitLab (auth required)"
  - "variant-profile conflict semantics: foundation pre-seeded vs user opinion — Phase 1 decision"
```

---

## Security Notes

Both variants introduce third-party repos with distinct trust boundaries:

| Repo | Variant | Trust Level | Notes |
|------|---------|-------------|-------|
| `[cachyos]` | CachyOS | Full GPG signing (Required DatabaseOptional) | cachyos-keyring; standard trust |
| `[cachyos-v3/v4]` | CachyOS | Full GPG signing | Same keyring as [cachyos] |
| `[chaotic-aur]` | Garuda | Full GPG signing (chaotic-keyring) | Large AUR mirror; pre-built AUR packages; broader attack surface than official repos |
| `[garuda]` | Garuda | [UNVERIFIED] | Garuda-specific packages |

**Schema implication:** The `custom-repo` opinion category must carry `sig_level` and
`keyring` fields. The translator must install the keyring before activating the repo.
Chaotic-AUR's broader scope (pre-built AUR packages) should be documented as a higher
trust-boundary risk in the variant profile.
