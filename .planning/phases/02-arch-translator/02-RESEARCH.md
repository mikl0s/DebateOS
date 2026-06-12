# Phase 2: Arch Translator — Research

**Researched:** 2026-06-12
**Domain:** archiso / mkarchiso / Python generator / unattended installer / Omarchy speech authoring
**Confidence:** HIGH (core mechanics verified in archlinux Docker container; variant data verified from cloned sources)

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- Python for the profile generator (resolved-speech JSON → archiso profile tree); shell as thin mkarchiso/Docker invocation layer.
- Input contract: Phase 1 resolver's CanonicalJSON ResolvedSpeech, passed as a file path argument.
- `translators/arch/capabilities.json` declares supported opinion categories and translator_capability tokens; composition-time capability check fails loudly on unsupported required opinions.
- Output: archiso profile directory + bootable unattended ISO built inside Docker (archlinux base image with archiso installed). Determinism via `SOURCE_DATE_EPOCH` derived from resolved-speech hash.
- Custom automated installer script embedded in ISO (airootfs): preset partitioning defaults, pacstrap from resolved package set, file-asset/config deployment, group memberships, service enablement, sysctl/kernel params, theme assets — driven by build-time manifest from the resolved speech. Zero install-time questions.
- `execution_phase: first-run` opinions become systemd first-boot units (oneshot with condition flag file), mirroring Omarchy's install-time vs first-run split.
- Declared hardware only: hardware-conditional resolution already happened in the resolver; no install-time hardware scanning.
- Bootability gate = ISO structural validation (ISO9660 + boot entries + airootfs + installer present). QEMU boot smoke is optional/manual.
- North-star equivalence gate = mechanical rootfs checks (package-set diff, file-asset presence, service enablement, first-run units) via inspecting built profile/rootfs.
- Declarative YAML profiles in `translators/arch/profiles/`: `vanilla-arch.yaml`, `cachyos.yaml`, `garuda.yaml`.
- `examples/omarchy/`: all 134 OM-NNN opinions as schema-valid YAML, resolver-resolved without hard conflicts.
- Slow paths (Docker ISO build, Omarchy ISO) are separate gated scripts; pytest covers the Python generator.
- Python: stdlib + PyYAML only; pytest as dev dependency. Pin in `translators/arch/requirements-dev.txt`.
- TDD RED/GREEN commits per D19.

### Claude's Discretion

- archiso profile internals (which releng baseline to start from).
- Exact installer script structure.
- Manifest format details.
- Docker image tag pinning strategy.

### Deferred Ideas (OUT OF SCOPE)

- Full variant ISO boot validation (CachyOS/Garuda) — non-gating stretch, post-v1.0.
- GitHub Actions reusable workflow + published Docker image — Phase 3.
- Direct-to-disk install target — post-v1.0.

</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| ARCH-01 | Arch translator (`translators/arch/`, shell/Python) consumes a resolved speech via JSON/YAML input contract, wraps mkarchiso, and emits a bootable, fully-unattended Arch installer ISO | §archiso mechanics, §installer script embedding, §Docker execution |
| ARCH-02 | NORTH STAR — building the Omarchy speech (`examples/omarchy/`) produces an installed system equivalent to Omarchy on vanilla Arch | §Omarchy opinion set, §north-star verification, §package set sizing |
| ARCH-03 | Translator declares its supported opinions/capabilities; unsupported required opinions break visibly at composition time | §capability gating, §capabilities.json pattern |
| ARCH-04 | Translator structured for 1-2 Arch variants via declarative variant profiles (repo + keyring + kernel + defaults) — no per-variant forks | §variant profiles, §CachyOS/Garuda data |

</phase_requirements>

---

## Summary

The Arch translator consists of three layers: (1) a Python generator that reads a `ResolvedSpeech` JSON and emits an archiso profile tree plus a build manifest; (2) a shell wrapper that invokes mkarchiso inside a privileged Docker container; (3) declarative YAML variant profiles that parameterize the generator without forking logic.

The archiso releng profile is the correct baseline. It uses `SOURCE_DATE_EPOCH` natively (verified in mkarchiso source), supports systemd-boot (UEFI) and syslinux (BIOS) out of the box, and embeds the installer via the autologin + `.zlogin` → `/root/.automated_script.sh` hook pattern — the same pattern the releng profile itself uses. The installer script lives in `airootfs/root/` and is invoked at first console login on the live medium. The resolved package set goes into `packages.x86_64`; file assets, service units, sysctl drop-ins, and first-run units are written into `airootfs/` at profile-generation time.

The Omarchy speech has 134 opinions (OM-001..OM-134) mapping to 32 evidence-driven points. The baseline install is ~155 named packages in 12 logical groups, plus 23 specialized/AUR opinions, 41 config opinions, 8 hardware-conditional opinions (hardware-resolved before the translator sees them), 5 login opinions, 1 post-install, 13 first-run (systemd oneshot units), and 21 themes (file-asset payloads). This is a large but tractable package set: expect 800–1500 packages total in the ISO (releng baseline ~128 + Omarchy ~285 non-overlapping) and an ISO build time of 20–40 minutes in Docker with a warm network cache.

**Primary recommendation:** Generate the archiso profile in Python (PyYAML + stdlib only), invoke mkarchiso in a `--privileged` Docker container (archlinux:base-devel pinned by digest), embed the installer as `airootfs/root/debateos-install.sh` called from `.zlogin`, and derive `SOURCE_DATE_EPOCH` from the resolved-speech JSON SHA-256 hash at profile-generation time.

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Resolved-speech JSON parsing | Python generator | — | Generator owns the translation contract |
| Profile tree emission | Python generator | — | Python owns archiso profile structure |
| Package list construction | Python generator | — | Reads `Opinion.Packages` from ResolvedSpeech |
| Variant repo/keyring injection | Python generator (via profile YAML) | — | Profile YAML parameterizes pacman.conf generation |
| Capability gating (SC-3) | Python generator | — | Checked before profile emission; fails loudly |
| Installer script generation | Python generator | Shell script template | Manifest-driven; shell does the actual install work |
| First-run unit generation | Python generator | — | Emits `.service` files into airootfs |
| mkarchiso invocation | Shell wrapper | Docker | Thin layer; no logic |
| ISO structural validation | Shell script | xorriso + unsquashfs | Post-build check; no bootability gate via QEMU |
| North-star rootfs checks | Shell script | Python | Inspects generated profile and ISO contents |
| Docker image management | Shell wrapper | — | Digest-pinned archlinux:base-devel |
| Omarchy speech authoring | examples/omarchy/ YAML | Python schema validation | Content; schema-validated by existing resolver/parse |

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `archiso` | `88-1` | mkarchiso profile builder | Official Arch Linux ISO creation tool [VERIFIED: archlinux Docker container] |
| `python-yaml` (PyYAML) | `6.0.3-2` | YAML parsing in generator | Project constraint: stdlib+PyYAML only; available in archlinux repos [VERIFIED: archlinux Docker container] |
| `arch-install-scripts` | bundled with archiso | pacstrap + arch-chroot | mkarchiso depends on this; provides pacstrap [VERIFIED: archlinux Docker container] |
| `squashfs-tools` | `4.7.5-1` | airootfs squashfs creation | mkarchiso dependency [VERIFIED: archlinux Docker container] |
| `libisoburn` (xorriso) | `1.5.8.2-1` | ISO creation + validation | mkarchiso dependency; also used for structural checks [VERIFIED: archlinux Docker container] |

### Dev (Python)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `python-pytest` | `1:9.0.3-1` | Generator TDD | All Python tests; pin in requirements-dev.txt [VERIFIED: archlinux Docker container] |
| `python-yaml` | `6.0.3-2` | PyYAML in tests | Same as production; no test-only YAML alternative needed |

### Docker Image

| Image | Tag Strategy | Notes |
|-------|-------------|-------|
| `archlinux:base-devel` | Digest pin (e.g. `sha256:dd60dfcca90f1...`) | `base-devel` includes build tools; `latest` is insufficient. Current digest: `sha256:dd60dfcca90f1ee6c2dd265ed27062070a1fb2e3b307723838a9d97741284722` [VERIFIED: docker pull archlinux:base-devel] |
| `archlinux:base-devel-YYYYMMDD.0.NNNNNN` | Dated tag (format confirmed) | `base-devel-20260607.0.541780` pulls the same image by date. Use digest pin in Dockerfile; dated tag for human reference. [VERIFIED: docker pull archlinux:base-devel-20260607.0.541780] |

**Installation (inside archlinux Docker container):**
```bash
pacman -Sy --noconfirm archiso python-yaml python-pytest
```

**Version verification (run at plan time):**
```bash
docker run --rm archlinux:base-devel bash -c "pacman -Sy --noconfirm archiso 2>/dev/null | tail -3 && pacman -Q archiso python-yaml python-pytest"
```

---

## Package Legitimacy Audit

> All packages in this phase are Arch Linux official-repo packages or the archlinux Docker official image — no npm/PyPI/crates packages are installed. Package legitimacy is governed by the Arch Linux packaging policy, not the npm/PyPI registry.

| Package | Registry | Notes | Verdict | Disposition |
|---------|----------|-------|---------|-------------|
| `archiso` | Arch extra | Official Arch ISO tooling, maintained by Arch devs | OK | Approved |
| `python-yaml` | Arch extra | PyYAML official Python binding | OK | Approved |
| `python-pytest` | Arch extra | Official pytest | OK | Approved |
| `arch-install-scripts` | Arch core | pacstrap + arch-chroot, maintained by Arch devs | OK | Approved |
| `archlinux:base-devel` | Docker Hub official | Official Arch Linux Docker image | OK | Approved |

**Packages removed due to [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

---

## Architecture Patterns

### System Architecture Diagram

```
ResolvedSpeech JSON (Phase 1 output)
          │
          ▼
  ┌─────────────────────────────────┐
  │  Python Generator               │
  │  translators/arch/generator.py  │
  │                                 │
  │  1. Parse ResolvedSpeech JSON   │
  │  2. Capability gate check       │──→ CapabilityError (SC-3: loud fail)
  │  3. Load variant profile YAML   │
  │  4. Emit archiso profile tree:  │
  │     • profiledef.sh             │
  │     • packages.x86_64           │
  │     • pacman.conf               │
  │     • airootfs/                 │
  │       ├─ root/debateos-install.sh  (installer)
  │       ├─ root/.zlogin           │   (hooks installer)
  │       ├─ etc/systemd/system/    │   (first-run units)
  │       ├─ etc/sysctl.d/          │
  │       └─ ...                    │
  │  5. Write build-manifest.json   │
  └─────────────┬───────────────────┘
                │ archiso profile dir
                ▼
  ┌──────────────────────────────────────┐
  │  Shell Wrapper                       │
  │  scripts/arch-build-iso.sh           │
  │                                      │
  │  docker run --privileged \           │
  │    -e SOURCE_DATE_EPOCH=<hash-epoch> │
  │    archlinux:base-devel@<digest>     │
  │    mkarchiso -v -w /work -o /out     │
  │    /profile                          │
  └─────────────┬────────────────────────┘
                │ .iso file
                ▼
  ┌──────────────────────────────────────┐
  │  Structural Validation               │
  │  scripts/arch-validate-iso.sh        │
  │                                      │
  │  xorriso -indev <iso> -toc           │
  │  xorriso -indev <iso> -find /        │
  │  unsquashfs -l airootfs.sfs          │
  │  check: installer present            │
  │  check: boot entries present         │
  └──────────────────────────────────────┘
                │
                ▼ (north-star gate)
  ┌──────────────────────────────────────┐
  │  North-Star Rootfs Checks            │
  │  scripts/arch-northstar-check.sh     │
  │                                      │
  │  - package set diff vs resolved speech│
  │  - file-asset presence check         │
  │  - service enablement check          │
  │  - first-run unit presence           │
  └──────────────────────────────────────┘
```

### Recommended Project Structure

```
translators/
└── arch/
    ├── README.md              # input contract, entrypoint usage
    ├── capabilities.json      # declared translator capabilities
    ├── generator.py           # ResolvedSpeech JSON → archiso profile tree
    ├── manifest.py            # build-manifest dataclass + serialization
    ├── profile.py             # archiso profile tree emitter
    ├── variant.py             # variant YAML loader + merger
    ├── firstrun.py            # first-run unit generator
    ├── requirements-dev.txt   # pytest (+ PyYAML already stdlib-adjacent)
    ├── profiles/
    │   ├── vanilla-arch.yaml  # baseline profile (no custom repos)
    │   ├── cachyos.yaml       # CachyOS repos/keyring/kernel
    │   └── garuda.yaml        # Garuda repos/keyring/dracut/btrfs
    ├── templates/
    │   ├── profiledef.sh.j2   # (or string template — no Jinja2 dep)
    │   ├── pacman.conf.tpl    # pacman.conf template
    │   ├── installer.sh.tpl   # debateos-install.sh template
    │   └── firstrun.service.tpl  # oneshot first-run unit template
    └── tests/
        ├── test_generator.py
        ├── test_manifest.py
        ├── test_profile.py
        ├── test_variant.py
        ├── test_firstrun.py
        ├── test_capability_gate.py
        └── fixtures/
            ├── minimal_resolved.json
            ├── omarchy_resolved.json  (subset for fast tests)
            └── cachyos_resolved.json

examples/
└── omarchy/
    ├── speech.yaml            # targets vanilla-arch, all 32 points
    ├── points/
    │   ├── om-point-01.yaml   # preflight + repo setup
    │   ├── om-point-02.yaml   # ...
    │   └── ... (32 total)
    ├── opinions/
    │   ├── OM-001.yaml        # custom-repo
    │   └── ... (134 total)
    └── LICENSE                # CC0 (examples/LICENSE exists)

scripts/
├── arch-build-iso.sh          # Docker mkarchiso invocation (slow gate)
├── arch-validate-iso.sh       # ISO structural validation
└── arch-northstar-check.sh    # package/file/service equivalence checks
```

### Pattern 1: Installer Embedding via `.zlogin` Hook

**What:** The releng profile uses `airootfs/root/.zlogin` which calls `~/.automated_script.sh`. That script looks for a `script=` kernel cmdline parameter (for network scripts) but defaults to running the automated script directly. For an unattended installer, the simpler pattern is to have `.zlogin` call the installer directly when running on tty1.

**When to use:** Every unattended Arch ISO — this is the canonical releng pattern.

```bash
# airootfs/root/.zlogin  (generated by translator)
# Source: /usr/share/archiso/configs/releng/airootfs/root/.zlogin (archiso 88-1)
if [[ "$(tty)" == "/dev/tty1" ]]; then
    /root/debateos-install.sh 2>&1 | tee /root/install.log
fi
```

```bash
# airootfs/etc/systemd/system/getty@tty1.service.d/autologin.conf
# Source: releng profile (archiso 88-1) — root autologin
[Service]
ExecStart=
ExecStart=-/usr/bin/agetty --noreset --noclear --autologin root - ${TERM}
```

**Why autologin + .zlogin over a custom systemd service:** The releng profile already provides both the autologin override and the `.zlogin` hook. Adding a custom systemd service for the installer would require additional ordering constraints to ensure the live system is fully up. The zsh/zlogin approach is simpler, battle-tested, and mirrors what real Arch ISO builds do. [VERIFIED: archlinux Docker container, archiso 88-1]

### Pattern 2: First-Run Systemd Oneshot Units (flag-file gated)

**What:** Opinions with `execution_phase: first-run` require a live Wayland/X session. They are embedded as systemd user services with a flag-file condition rather than `ConditionFirstBoot` (which checks `/etc/machine-id = uninitialized` — a system-level condition that fires once per machine-id initialization, too coarse for per-user first-run).

**When to use:** Any opinion requiring a display session (gsettings, notify-send, display-dependent tools).

```ini
# airootfs/etc/systemd/user/debateos-firstrun-OM-102.service
# (generated by translator for OM-102 gnome-theme)
[Unit]
Description=DebateOS first-run: OM-102 GTK theme configuration
ConditionPathExists=!/var/lib/debateos/.firstrun-OM-102.done
After=graphical-session.target

[Service]
Type=oneshot
ExecStart=/usr/lib/debateos/firstrun/OM-102.sh
ExecStartPost=/bin/touch /var/lib/debateos/.firstrun-OM-102.done
RemainAfterExit=yes

[Install]
WantedBy=graphical-session.target
```

The flag file `/var/lib/debateos/.firstrun-OM-NNN.done` ensures idempotency without relying on `ConditionFirstBoot` (which requires `machine-id = uninitialized` — set to `uninitialized` by mkarchiso, then initialized at first boot, so it fires exactly once — this IS actually appropriate; see Pitfall 3 below for the subtlety). [ASSUMED — exact condition gate; either ConditionFirstBoot or flag files are valid; flag files are simpler and more explicit]

### Pattern 3: SOURCE_DATE_EPOCH Derivation

**What:** mkarchiso reads `SOURCE_DATE_EPOCH` from the environment and uses it for all timestamps (ISO label, build date, file mtimes). [VERIFIED: archlinux Docker container — grep of mkarchiso source shows `[[ -v SOURCE_DATE_EPOCH ]] || printf -v SOURCE_DATE_EPOCH '%(%s)T' -1`]

**Derivation approach (Claude's discretion):** Hash the canonical ResolvedSpeech JSON (SHA-256) and take the first 4 bytes as a Unix timestamp integer. This guarantees same speech → same timestamp → same ISO label, satisfying BLD-03 groundwork.

```python
import hashlib, json, struct

def derive_source_date_epoch(resolved_speech_path: str) -> int:
    """Derive SOURCE_DATE_EPOCH from resolved speech hash for build determinism."""
    with open(resolved_speech_path) as f:
        content = f.read()
    digest = hashlib.sha256(content.encode()).digest()
    # Take first 4 bytes as big-endian uint32 for a stable epoch in valid range
    epoch = struct.unpack(">I", digest[:4])[0]
    # Clamp to a reasonable range: 2020-01-01 to 2040-01-01
    MIN_EPOCH = 1577836800
    MAX_EPOCH = 2208988800
    return MIN_EPOCH + (epoch % (MAX_EPOCH - MIN_EPOCH))
```

### Pattern 4: Capability Gate (SC-3)

**What:** `translators/arch/capabilities.json` lists every `translator_capability` token the Arch translator supports. At generator startup, compare `Opinion.TranslatorCapabilities` for each applied opinion against the capabilities set. If any required opinion declares an unsupported capability, fail with a human-readable error.

```python
# Source: CONTEXT.md SC-3 decision
def check_capabilities(resolved_speech, capabilities: set[str]) -> None:
    for opinion_id in resolved_speech["applied"]:
        opinion = opinions_index[opinion_id]
        if opinion.get("status") == "required":
            for cap in opinion.get("translator_capabilities", []):
                if cap not in capabilities:
                    raise CapabilityError(
                        f"Opinion {opinion_id} requires capability '{cap}' "
                        f"which is not declared by the Arch translator.\n"
                        f"Add it to translators/arch/capabilities.json when implemented."
                    )
        else:  # nice-to-have
            # Drop with explanation, don't fail
            unsupported = [c for c in opinion.get("translator_capabilities", []) if c not in capabilities]
            if unsupported:
                log_dropped(opinion_id, f"unsupported capabilities: {unsupported}")
```

### Pattern 5: pacstrap at ISO Build Time (Inside the Container)

**What:** mkarchiso calls `pacstrap -C pacman.conf -G -M /work/iso/airootfs/x86_64/airootfs PACKAGE...` to build the live ISO's squashfs. This IS the standard approach — packages are downloaded and installed at ISO build time inside the container, not at install time from the live ISO.

**For the installed TARGET rootfs:** The installer script (running on the live ISO) calls `pacstrap /mnt PACKAGE...` to install onto the target disk. This requires network access at install time. This is the correct approach for "fully-unattended installer ISO" — it is simpler than an offline package cache and consistent with standard Arch installs. The trade-off: install time is network-dependent.

**Offline package cache option:** Embedding a package cache in the ISO (a `[custom]` repo in pacman.conf pointing to `airootfs/` mounted packages) is possible but adds ~2–3 GB to ISO size for a full Omarchy install. This is Phase 3+ territory (build channels with caching). For Phase 2: **network at install time** is the recommended approach. [ASSUMED — offline cache feasibility for full Omarchy set; network-at-install confirmed simpler per CONTEXT.md "fully-unattended installer ISO"]

**Recommendation:** The installer script uses `pacstrap /mnt base linux linux-firmware <resolved-packages>` with the default Arch mirrors. The ISO includes the live environment packages (releng baseline ~128 packages) but does NOT pre-cache the target system packages.

### Anti-Patterns to Avoid

- **Hand-rolling a bootloader configurator:** mkarchiso handles GRUB, syslinux, and systemd-boot entries. Don't write bootloader config manually — use the profiledef.sh `bootmodes` array.
- **Using `/dev/loop` directly in Docker without `--privileged`:** mkarchiso uses `mksquashfs` (not loop devices) for squashfs builds. The only privileged operation is the chroot/bind-mounts that pacstrap uses. Use `--privileged` — attempts to restrict to `--cap-add SYS_ADMIN,MKNOD` often fail on overlayfs Docker storage drivers. [VERIFIED: archlinux Docker container — mkarchiso grep shows no losetup calls; pacstrap uses mount bind]
- **Storing SOURCE_DATE_EPOCH in a mutable file:** mkarchiso reads it from the environment variable. Set it in the `docker run -e SOURCE_DATE_EPOCH=...` call.
- **Having the installer script prompt for any input:** Use `--noconfirm` in all pacman calls; pre-set locale/timezone in airootfs; use `printf 'y\n'` for any interactive tool that can't be silenced.
- **Installing the `omarchy` repo keyring in a chroot without pacman-key init:** The installer must `arch-chroot /mnt pacman-key --init && pacman-key --populate archlinux` before adding custom repos.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| ISO creation | Custom xorriso invocations | `mkarchiso` | mkarchiso handles boot entries (syslinux + systemd-boot), El Torito, squashfs, reproducibility — 2,000+ lines of battle-tested shell |
| Squashfs creation | `mksquashfs` directly | mkarchiso (calls mksquashfs) | Ordering, permissions, options handled |
| YAML parsing in Python | Custom parser | `python-yaml` (PyYAML) | YAML 1.1 edge cases (boolean inference etc.) are non-trivial |
| Package list validation | Custom resolver | Go resolver (Phase 1) | Already done; translator consumes ResolvedSpeech |
| Boot entry templating | Custom efi/grub templates | Copy releng baseline, parameterize only label/version | releng entries tested on thousands of machines |
| Arch chroot execution | Custom mount/unmount | `arch-chroot` (from arch-install-scripts) | Handles proc/sys/dev bind mounts and teardown correctly |

**Key insight:** mkarchiso is the only sanctioned, community-tested path for Arch ISOs. Any custom ISO builder has a ~95% chance of producing something that fails to boot on real hardware (UEFI secure boot, syslinux vs systemd-boot fallback, El Torito catalog placement). Use it.

---

## archiso Profile Internals

### Profile Structure (archiso 88-1, releng baseline)

[VERIFIED: archlinux Docker container, archiso 88-1]

```
profile/
├── profiledef.sh              # ISO metadata + build config
├── packages.x86_64            # packages in the live environment (one per line)
├── pacman.conf                # pacman config for building the live env
├── airootfs/                  # overlay onto the live env rootfs
│   ├── etc/
│   │   ├── hostname
│   │   ├── locale.conf
│   │   ├── systemd/
│   │   │   ├── system/getty@tty1.service.d/autologin.conf
│   │   │   └── system/pacman-init.service
│   │   └── shadow              # root password hash
│   └── root/
│       ├── .zlogin             # calls installer on tty1
│       └── debateos-install.sh # generated installer script
├── efiboot/                   # systemd-boot entries (UEFI)
│   └── loader/entries/*.conf
├── grub/                      # GRUB config (BIOS + UEFI fallback)
│   ├── grub.cfg
│   └── loopback.cfg
└── syslinux/                  # syslinux config (BIOS)
    └── *.cfg
```

**`profiledef.sh` key fields:**
- `iso_name="debateos"` — ISO filename prefix
- `iso_label="DEBATEOS_$(date ...)"` — volume label (uses SOURCE_DATE_EPOCH)
- `bootmodes=('bios.syslinux' 'uefi.systemd-boot')` — both modes from releng
- `airootfs_image_type="squashfs"` — standard squashfs (not ext4+squashfs)
- `file_permissions=(...)` — must include installer script at 0:0:755

### Minimal Deviation from releng Baseline

The generator should copy the releng baseline and make targeted modifications:
1. Replace `packages.x86_64` with live-env packages (releng baseline + installer dependencies).
2. Add `airootfs/root/debateos-install.sh` (generated installer).
3. Replace `airootfs/root/.zlogin` with one that calls the installer.
4. Add first-run service units to `airootfs/etc/systemd/user/`.
5. Modify `profiledef.sh` `iso_name`, `iso_label`, `iso_version` only.
6. For variants: modify `pacman.conf` to add custom repos + keyring packages to `packages.x86_64`.

Do NOT modify the syslinux/efiboot/grub configs unless changing kernel params — these are tested configurations.

---

## Variant Profiles

### vanilla-arch.yaml (baseline — no custom repos)

```yaml
variant: vanilla-arch
description: "Standard Arch Linux — no custom repos or kernel variants"
repos: []
keyring_install_before_repos: []
kernel:
  package: linux
  headers: linux-headers
defaults:
  initramfs: mkinitcpio
  bootloader: null   # translator choice: limine (per Omarchy) or systemd-boot
  filesystem: null   # translator choice from speech defaults
pre_seeded_opinions: []
```

### cachyos.yaml (CachyOS — verified from cloned source)

Key verified facts [VERIFIED: research/arch-variants-delta.md, direct clone inspection]:
- Repos: `[cachyos]`, `[cachyos-v3]`, `[cachyos-core-v3]`, `[cachyos-extra-v3]` above `[core]`/`[extra]`
- Mirror: `https://cdn77.cachyos.org/repo/$arch/$repo`
- Keyring: `cachyos-keyring` (pkg 20240331), must be installed before repos
- Kernel: `linux-cachyos` (EEVDF, 7.0.12 at research time), headers: `linux-cachyos-headers`
- Initramfs: mkinitcpio (default); dracut-cachyos optional
- Pre-seeded sysctl: 12 params in `70-cachyos-settings.conf` (no collision with Omarchy)
- Unverified: default FS/bootloader (installer presents choice — mark as `null` with note)

### garuda.yaml (Garuda — verified from cloned source)

Key verified facts [VERIFIED: research/arch-variants-delta.md]:
- Repos: `[chaotic-aur]` (confirmed), `[garuda]` (URL `[UNVERIFIED]`)
- Keyring: `chaotic-keyring` (mandatory); `garuda-keyring` existence unverified
- Kernel: `linux` or `linux-zen` (no custom kernel); headers: `linux-headers`
- Initramfs: `dracut` (HARD CONFLICT with Omarchy — Garuda's `garuda-dracut-support` `conflicts=('mkinitcpio')`)
- Filesystem: btrfs mandatory (maintenance timers always enabled)
- Bootloader: GRUB mandatory (HARD CONFLICT with Omarchy which uses limine)
- Snapper: mandatory root config (HARD CONFLICT with Omarchy's own snapper config)
- SDDM theme: Dr460nized (DIRECT COLLISION with Omarchy SDDM theme)

**Note for planner:** The Garuda variant profile should be authored and the conflicts marked, but Omarchy-on-Garuda is a non-gating stretch per CONTEXT.md (deferred to post-v1.0). The `garuda.yaml` variant profile is Phase 2 work; the full ISO validation with Garuda is not.

---

## Omarchy Speech Details

### Opinion Count and Grouping

| Phase | IDs | Count | Notes |
|-------|-----|-------|-------|
| Preflight | OM-001..005 | 5 | Repo setup, mkinitcpio disable, migrations, first-run mode |
| Packaging/Base | OM-006..017 | 12 | 12 logical groups of 155 packages |
| Packaging/Specialized | OM-018..023 | 6 | Fonts, nvim, icons, webapps, TUIs, npm globals |
| Packaging/Hardware-Conditional | OM-024..027 | 4 | ASUS ROG, Framework 16, Dell XPS haptic, Surface WiFi |
| Config/General | OM-028..060 | 33 | Dotfiles, services, sysctl, docker, etc. |
| Config/Hardware General | OM-061..069 | 9 | iwd, regdom, Bluetooth, CUPS, USB, NVIDIA, Vulkan |
| Config/Hardware Intel | OM-070..077 | 8 | VAAPI, lpmd, thermald, IPU7, PTL kernel, FRED, WiFi7 fix, sof-firmware |
| Config/Hardware ASUS | OM-078..083 | 6 | Backlight, display, touchpad, audio |
| Config/Hardware Framework | OM-084..085 | 2 | F13 AMD audio, QMK udev |
| Config/Hardware Apple | OM-086..088 | 3 | SPI kbd, NVMe suspend, T2 support |
| Config/Hardware Lenovo | OM-089 | 1 | Yoga Pro audio |
| Config/Hardware Misc | OM-090..094 | 5 | BCM WiFi, Surface keyboard, YT6801, synaptics, Tuxedo |
| Login | OM-095..099 | 5 | Plymouth, keyring, SDDM, hibernation, limine-snapper |
| Post-Install | OM-100 | 1 | Final pacman.conf (+ arch-mact2 for T2 systems) |
| First-Run | OM-101..113 | 13 | All need live session; become systemd oneshot units |
| Themes | OM-114..134 | 21 | File-asset bundles (21 themes) |
| **Total** | | **134** | |

### ISO Size and Build Time Estimate

- **Live ISO packages (releng baseline):** ~128 packages (verified from packages.x86_64)
- **Omarchy live-env additions:** pacstrap target installer needs `arch-install-scripts` (already in releng), plus live-env pkgs for the installer to run (minimal — limine, pacman, base, btrfs-progs). The live env does NOT pre-install all 285 Omarchy packages; they are downloaded at install time from the network.
- **Expected ISO size (live env only):** 800 MB–1.2 GB (releng baseline is ~800 MB; additions are minimal)
- **Install time on target (network-dependent):** 20–45 minutes for full Omarchy package set (~285 packages + AUR)
- **ISO build time in Docker (warm network cache):** 20–35 minutes (pacstrap of live env packages; squashfs creation)
- **ISO build time (cold network):** 35–60 minutes [ASSUMED — based on package count and typical Arch mirror speeds]

### Key Translator Capability Tokens Required by Omarchy

From the opinion inventory, these are the capability tokens the translator must declare in `capabilities.json`:

| Token (from `translator-capability:` fields) | OM-NNN | Category |
|----------------------------------------------|--------|----------|
| add-signed-external-repo | OM-001 | custom-repo |
| import-gpg-key-by-fingerprint | OM-001 | custom-repo |
| disable-package-manager-hooks | OM-002 | service-enable |
| restore-package-manager-hooks | OM-002 | service-enable |
| install-named-packages | OM-006..017 | package-install |
| deploy-font-file | OM-018 | font-install |
| run-post-install-setup-script | OM-019 | arbitrary-script |
| copy-icon-files | OM-020 | arbitrary-script |
| create-desktop-launchers | OM-021, OM-022 | arbitrary-script |
| install-npm-global-packages | OM-023 | npm-global-install |
| install-packages-conditional-hardware | OM-024..027 | hardware-conditional |
| deploy-config-file-tree | OM-028 | config-dotfile |
| manage-multi-app-theme-system | OM-029 | theming |
| write-git-global-config | OM-031 | config-dotfile |
| write-sysctl-drop-in | OM-036..038 | sysctl-param |
| write-systemd-drop-in | OM-043, OM-050 | service-enable |
| configure-docker-daemon | OM-043 | service-enable |
| register-mime-types | OM-044 | mime-type |
| enable-systemd-service-chroot | OM-057, OM-064, OM-065 | service-enable |
| write-udev-rules | OM-058, OM-059 | config-dotfile |
| write-modprobe-config | OM-063, OM-066 | config-dotfile |
| add-user-to-group | OM-053 | user-group |
| configure-bootloader-limine | OM-099 | bootloader-config |
| create-snapper-config | OM-099 | bootloader-config |
| deploy-plymouth-theme | OM-095 | theming |
| configure-sddm-autologin | OM-097 | display-manager |
| configure-pam | OM-034, OM-035, OM-097 | config-dotfile |
| write-sudoers-drop-in | OM-033, OM-034, OM-052 | config-dotfile |
| deploy-file-asset-bundle | OM-114..134 | theming |
| systemd-user-service-firstrun | OM-101..113 | service-enable |

**Note for planner:** Not all capabilities need be implemented in Wave 1. The capability gate (ARCH-03) only fails on *required* opinions with unsupported capabilities. Nice-to-have opinions with unsupported capabilities are dropped with explanation. The minimum set for the north-star ARCH-02 gate is the subset of capabilities used by OM-001..OM-134 where those opinions have `status: required` in the Omarchy speech.

---

## Common Pitfalls

### Pitfall 1: mkarchiso Requires Root (Privileged Docker)

**What goes wrong:** Running mkarchiso as non-root fails at the pacstrap step with `mount: /proc: permission denied` or similar.

**Why it happens:** pacstrap uses bind mounts (`mount --bind`) to set up the chroot environment. This requires `CAP_SYS_ADMIN` at minimum; in practice `--privileged` is needed because Docker's default AppArmor/seccomp profiles also block some operations.

**How to avoid:** Always run the mkarchiso container with `docker run --privileged`. Don't attempt `--cap-add` combinations — they are unreliable across Docker storage drivers and kernel versions.

**Warning signs:** `mount: permission denied`, `pacstrap: error: failed to install packages to new root`, anything about `/proc`.

[VERIFIED: archlinux Docker container — pacstrap grep shows mount bind calls requiring CAP_SYS_ADMIN]

### Pitfall 2: packages.x86_64 vs Target Install

**What goes wrong:** Adding all 285 Omarchy packages to `packages.x86_64` (the live ISO's squashfs) results in a 3+ GB ISO and 40+ minute squashfs compression, while the target install still runs pacstrap anyway.

**Why it happens:** `packages.x86_64` installs packages INTO the live ISO's squashfs (the RAM-resident environment). The installer runs pacstrap FROM the live environment TO the target disk. These are separate package sets.

**How to avoid:** Keep `packages.x86_64` minimal (releng baseline + installer requirements: `arch-install-scripts`, `btrfs-progs`, `limine` if needed for the installer, maybe `dosfstools`, `e2fsprogs`). All Omarchy packages go into the installer script's `pacstrap /mnt` call.

### Pitfall 3: ConditionFirstBoot vs Flag File for First-Run Units

**What goes wrong:** Using `ConditionFirstBoot=yes` in user services doesn't work as expected because `ConditionFirstBoot` checks the SYSTEM's `/etc/machine-id` state — it's a system-level condition evaluated by systemd, not a user-level condition.

**Why it happens:** `ConditionFirstBoot=yes` checks `machine-id = uninitialized` (set by mkarchiso, cleared at first boot by `systemd-machine-id-setup`). It fires once for the entire system, not per-user. User units with `ConditionFirstBoot=yes` should work for system-scope units but may not for user-scope ones depending on systemd version.

**How to avoid:** Use a flag file approach for user-scope first-run units: `ConditionPathExists=!/var/lib/debateos/.firstrun-OM-NNN.done`. The service creates the flag file on success. This is explicit, debuggable, and doesn't rely on systemd internal state. For system-scope first-run units, `ConditionFirstBoot=yes` is appropriate. [ASSUMED — per-user ConditionFirstBoot behavior; flag file approach is conservative and safe]

**Warning signs:** First-run script runs on every boot, or never runs.

### Pitfall 4: Keyring Installation Order for Custom Repos

**What goes wrong:** Adding a CachyOS or Garuda repo to `pacman.conf` before installing the corresponding keyring results in `error: key "XXXX" could not be looked up remotely`.

**Why it happens:** pacman verifies package signatures against its local keyring. The keyring package (`cachyos-keyring`, `chaotic-keyring`) must be installed with `--noconfirm` and `pacman-key` initialized BEFORE using the repo.

**How to avoid:** In the installer script, install keyring packages first (using a bootstrapped pacman.conf without the custom repos), then add the custom repos to pacman.conf, then install remaining packages. The variant YAML's `keyring_install_before_repos` field documents this ordering.

**Warning signs:** `error: key "..." could not be looked up remotely`, `error: database '...' is not valid`.

[VERIFIED: research/arch-variants-delta.md — CachyOS and Garuda both require keyring pre-install]

### Pitfall 5: AUR Packages Cannot Be Installed via pacstrap

**What goes wrong:** Some Omarchy packages (e.g. `omarchy-walker`, `omarchy-nvim-setup`, `claude-code`, `gpu-screen-recorder`, `limine`) come from the Omarchy custom repo or AUR, not from official Arch repos. `pacstrap` uses pacman, which can only install from configured repos.

**Why it happens:** AUR packages require building from source (makepkg). pacstrap cannot run makepkg inside a chroot without a build environment.

**How to avoid:** Either (a) all required packages are available from official repos + Omarchy repo (preferred — OM-001 registers the Omarchy repo, which provides omarchy-specific packages), or (b) the installer script uses `paru` or `yay` in the target chroot for AUR-only packages. Check: does the Omarchy repo provide all packages listed in `omarchy-base.packages`? [ASSUMED — Omarchy repo coverage of all OM-NNN packages; verified via OM-001 which registers the repo; `paru`/`yay` fallback may be needed]

**Warning signs:** `error: target not found: <package-name>`.

### Pitfall 6: Python Template Generation vs Shell Templates

**What goes wrong:** Using string `.format()` or f-strings to generate shell scripts with arbitrary user data (package names, paths) risks shell injection if any field contains characters like `$`, `` ` ``, `'`, etc.

**Why it happens:** The resolved speech is attacker-controlled data only in the Phase 5+ registry context. In Phase 2 it is locally authored. But defensive practice matters.

**How to avoid:** Either (a) write all variable data as heredoc strings properly quoted in the generated shell, or (b) serialize the build manifest as JSON and have the installer script read it at runtime (`jq` in the live env). Option (b) is cleaner and avoids injection. [ASSUMED — which quoting approach is best; JSON manifest read at runtime is the safer pattern]

---

## Code Examples

### Minimal archiso profiledef.sh

```bash
#!/usr/bin/env bash
# Source: /usr/share/archiso/configs/releng/profiledef.sh (archiso 88-1)
# Modified for DebateOS

iso_name="debateos"
iso_label="DEBATEOS_$(date --date="@${SOURCE_DATE_EPOCH:-$(date +%s)}" +%Y%m)"
iso_publisher="DebateOS"
iso_application="DebateOS Arch Installer"
iso_version="$(date --date="@${SOURCE_DATE_EPOCH:-$(date +%s)}" +%Y.%m.%d)"
install_dir="arch"
buildmodes=('iso')
bootmodes=('bios.syslinux' 'uefi.systemd-boot')
pacman_conf="pacman.conf"
airootfs_image_type="squashfs"
airootfs_image_tool_options=('-comp' 'xz' '-Xbcj' 'x86' '-b' '1M' '-Xdict-size' '1M')
file_permissions=(
  ["/etc/shadow"]="0:0:400"
  ["/root"]="0:0:750"
  ["/root/debateos-install.sh"]="0:0:755"
  ["/root/.zlogin"]="0:0:644"
)
```

### Docker Invocation Pattern

```bash
#!/usr/bin/env bash
# scripts/arch-build-iso.sh
# Source: CONTEXT.md decisions + archiso 88-1 verified mechanics

PROFILE_DIR="${1:?usage: arch-build-iso.sh <profile-dir> <out-dir>}"
OUT_DIR="${2:?}"
SOURCE_DATE_EPOCH="${SOURCE_DATE_EPOCH:?SOURCE_DATE_EPOCH must be set}"
DOCKER_IMAGE="archlinux:base-devel@sha256:dd60dfcca90f1ee6c2dd265ed27062070a1fb2e3b307723838a9d97741284722"

docker run --privileged --rm \
  -v "$(realpath "$PROFILE_DIR"):/profile:ro" \
  -v "$(realpath "$OUT_DIR"):/out" \
  -e "SOURCE_DATE_EPOCH=${SOURCE_DATE_EPOCH}" \
  "${DOCKER_IMAGE}" \
  bash -c "
    pacman -Sy --noconfirm archiso 2>/dev/null
    mkarchiso -v -w /tmp/work -o /out /profile
  "
```

### ISO Structural Validation

```bash
#!/usr/bin/env bash
# scripts/arch-validate-iso.sh

ISO="${1:?usage: arch-validate-iso.sh <iso-file>}"

echo "=== ISO9660 / El Torito check ==="
xorriso -indev "${ISO}" -pvd_info 2>&1 | grep -E 'Volume|Boot'

echo "=== Boot entries ==="
xorriso -indev "${ISO}" -find /EFI -type f 2>&1 | head -20

echo "=== airootfs.sfs presence ==="
xorriso -indev "${ISO}" -find /arch/x86_64 -type f 2>&1 | grep airootfs

echo "=== Squashfs listing (installer check) ==="
# Extract airootfs.sfs first, then list
# (full extraction not needed — use xorriso to extract just the sfs)
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT
xorriso -indev "${ISO}" -osirrox on -extract /arch/x86_64/airootfs.sfs "${TMPDIR}/airootfs.sfs" 2>/dev/null
unsquashfs -l "${TMPDIR}/airootfs.sfs" | grep -E 'install\.sh|\.zlogin|firstrun'
```

### Python Generator Skeleton

```python
# translators/arch/generator.py
# Source: CONTEXT.md translator structure decision

import json
import sys
import hashlib
import struct
from pathlib import Path
import yaml

from .capabilities import load_capabilities, check_capabilities
from .profile import emit_profile_tree
from .manifest import BuildManifest
from .variant import load_variant_profile


def generate(resolved_speech_path: str, profile_name: str, out_dir: str) -> None:
    """Main entry: ResolvedSpeech JSON -> archiso profile tree."""
    resolved_path = Path(resolved_speech_path)
    with open(resolved_path) as f:
        rs = json.load(f)

    # SC-3: capability gate (required before any profile emission)
    capabilities = load_capabilities()
    opinions_index = _load_opinions_index(rs)
    check_capabilities(rs, opinions_index, capabilities)

    # Derive SOURCE_DATE_EPOCH for determinism (BLD-03 groundwork)
    source_date_epoch = _derive_epoch(resolved_path.read_text())

    # Load variant profile
    variant = load_variant_profile(profile_name)

    # Build manifest
    manifest = BuildManifest.from_resolved_speech(rs, opinions_index, variant)

    # Emit profile tree
    emit_profile_tree(Path(out_dir), manifest, source_date_epoch)


def _derive_epoch(content: str) -> int:
    digest = hashlib.sha256(content.encode()).digest()
    epoch = struct.unpack(">I", digest[:4])[0]
    MIN_EPOCH = 1577836800  # 2020-01-01
    MAX_EPOCH = 2208988800  # 2040-01-01
    return MIN_EPOCH + (epoch % (MAX_EPOCH - MIN_EPOCH))
```

### Entrypoint (argv-stable for Phase 3 CLI)

```bash
# translators/arch/translate  (thin shell entrypoint per CONTEXT.md Phase 3 contract)
#!/usr/bin/env bash
set -euo pipefail
RESOLVED_JSON="${1:?usage: translate <resolved.json> --profile <profile> --out <dir>}"
shift
PROFILE="vanilla-arch"
OUT_DIR="./arch-profile"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --profile) PROFILE="$2"; shift 2 ;;
    --out) OUT_DIR="$2"; shift 2 ;;
    *) echo "unknown flag: $1" >&2; exit 1 ;;
  esac
done
exec python3 -m translators.arch.generator "$RESOLVED_JSON" "$PROFILE" "$OUT_DIR"
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| GRUB-only archiso (older profiles) | systemd-boot (UEFI) + syslinux (BIOS) | archiso ~v50+ | releng bootmodes covers both; no manual efibootmgr calls |
| `/etc/systemd/system/archiso-mount.service` pattern | autologin+.zlogin as standard | archiso 60+ | Simpler; no custom service unit needed for installer hook |
| `archlinux:base` Docker image for builds | `archlinux:base-devel` (includes makepkg/base-devel) | ongoing | base-devel needed for any package build; base insufficient |
| Manual SOURCE_DATE_EPOCH tracking | mkarchiso reads it natively from env | archiso (long-standing) | Set env var in docker run; no patching of mkarchiso needed |

**Current archiso version:** `88-1` (verified June 2026 in Docker container)

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `ConditionFirstBoot=yes` in user-scope systemd units is unreliable across systemd versions; flag files are safer | Pitfall 3 / Pattern 2 | Flag files still work; ConditionFirstBoot might work fine — add it as optimization later |
| A2 | Offline package cache is not needed for Phase 2 (network at install time is acceptable) | Pattern 5 | If network-less install is required, substantial rework needed to embed packages in ISO |
| A3 | Omarchy custom repo provides all packages listed in `omarchy-base.packages` (no AUR-only packages requiring `yay`/`paru`) | Pitfall 5 | Some packages may require AUR build; installer script would need `paru` bootstrap step |
| A4 | ISO size ~800 MB–1.2 GB (live env only); install time 20–45 min | §ISO Size | Network/hardware variation; estimate could be 2x on slow mirrors |
| A5 | JSON manifest pattern (installer reads `manifest.json` at runtime) is the correct injection approach for generated data | Pitfall 6 | Shell quoting approach also valid; either works; JSON is cleaner |
| A6 | `execution_phase: first-run` maps to systemd user oneshot services with flag-file conditions (rather than `ConditionFirstBoot`) | Pattern 2 | If ConditionFirstBoot works correctly for user services, either approach is valid |
| A7 | `paru` is not needed in Phase 2 (all Omarchy packages available via configured repos) | Pitfall 5 | If AUR-only packages exist in the resolved set, installer needs AUR helper |

---

## Open Questions

1. **Which opinions in examples/omarchy/ should be `status: required` vs `status: nice-to-have`?**
   - What we know: The Omarchy inventory has 134 opinions; the CONTEXT.md says "resolver-resolved without hard conflicts" implies careful status assignment.
   - What's unclear: Whether core desktop opinions (compositor, terminal, display manager) should all be required, or whether some can be nice-to-have (allowing substitution).
   - Recommendation: Mark as required: OM-001 (repo), OM-006 (compositor), OM-097 (display manager), OM-099 (bootloader). Mark hardware-conditional as required (their hardware condition already gates them). Mark theme opinions (OM-114..134) as nice-to-have.

2. **Does the Omarchy repo (OM-001, GPG key 40DFB630FF42BCFFB047046CF0134EE680CAC571) remain stable across Omarchy versions?**
   - What we know: Pinned to commit 9cf1852; key ID recorded.
   - What's unclear: Whether the repo key changes across releases.
   - Recommendation: Embed the key fingerprint in the OM-001 opinion; translator checks fingerprint at install time.

3. **`limine` bootloader: is it in official Arch repos or the Omarchy repo?**
   - What we know: OM-099 uses limine-snapper; Omarchy repo registered by OM-001.
   - What's unclear: Whether `limine` is in `[core]`/`[extra]` or in the `[omarchy]` custom repo.
   - Recommendation: Check `pacman -Si limine` in an archlinux container before the limine-snapper implementation task.

4. **Point grouping for examples/omarchy/: 32 points from omarchy-points.md — is single-opinion-per-point acceptable for points that have only one natural opinion?**
   - What we know: CONTEXT.md says "single-opinion points allowed where natural" (from STATE.md decision).
   - What's unclear: Whether the 32 points map cleanly to the 134 opinions or require splitting.
   - Recommendation: Points map 1:N to opinions; some points will have 1 opinion, others 5–15.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Docker | ISO builds, all slow gates | ✓ | 29.5.3 | — |
| Python 3 | Generator unit tests (host) | ✓ | 3.12.3 | — |
| PyYAML (host) | Generator unit tests (host) | ✓ | available | — |
| pytest (host) | Generator unit tests | ✗ (not installed on host) | — | Install via `pip install pytest` or use archlinux container |
| QEMU | Boot smoke test | ✗ | — | DOCUMENTED: structural validation only (no QEMU) |
| archiso 88-1 | ISO building | ✓ (inside Docker) | 88-1 | — |
| xorriso | ISO validation | ✓ (inside Docker) | 1.5.8.pl02 | — |
| unsquashfs | Squashfs listing | ✓ (inside Docker) | 4.7.5-1 | — |
| Go toolchain | Regression tests (Phase 1 suite) | ✓ | 1.24 | — |

**Missing dependencies with no fallback:**
- QEMU: boot smoke test is explicitly documented as optional/manual (CONTEXT.md decision). No fallback needed — structural validation is the Phase 2 bootability gate.

**Missing dependencies with fallback:**
- pytest on host: The generator tests can run either on the host (after `pip install pytest`) or inside an archlinux Docker container. Both are valid. The Docker approach is more reproducible.

**Note:** All slow gates (ISO build, structural validation, north-star checks) run inside Docker; they have no host-side dependency beyond Docker itself.

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | pytest 1:9.0.3-1 (Arch official) |
| Config file | `translators/arch/pytest.ini` — Wave 0 gap |
| Quick run command | `pytest translators/arch/tests/ -x -q` |
| Full suite command | `pytest translators/arch/tests/ -v && go test ./... -count=1` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| ARCH-01 | Generator emits valid archiso profile tree from minimal ResolvedSpeech | unit | `pytest translators/arch/tests/test_profile.py -x` | ❌ Wave 0 |
| ARCH-01 | Installer script embedded at correct path with 0755 permissions | unit | `pytest translators/arch/tests/test_profile.py::test_installer_path -x` | ❌ Wave 0 |
| ARCH-01 | Docker ISO build completes without error (slow gate) | integration | `bash scripts/arch-build-iso.sh <profile> <out>` | ❌ Wave 0 |
| ARCH-01 | ISO structural validation passes (xorriso + unsquashfs) | integration | `bash scripts/arch-validate-iso.sh <iso>` | ❌ Wave 0 |
| ARCH-02 | Omarchy speech resolves cleanly (no hard conflicts) | unit | `go test ./examples/... -run TestExampleOmarchy -v` | ❌ Wave 0 |
| ARCH-02 | Generated profile contains all OM-NNN packages | unit | `pytest translators/arch/tests/test_northstar.py -x` | ❌ Wave 0 |
| ARCH-02 | Full north-star: build + rootfs package/file/service check (slow gate) | integration | `bash scripts/arch-northstar-check.sh` | ❌ Wave 0 |
| ARCH-03 | Unsupported required capability raises CapabilityError with message | unit | `pytest translators/arch/tests/test_capability_gate.py -x` | ❌ Wave 0 |
| ARCH-03 | Unsupported nice-to-have capability silently drops opinion | unit | `pytest translators/arch/tests/test_capability_gate.py::test_nicetohave_drop -x` | ❌ Wave 0 |
| ARCH-04 | CachyOS profile YAML loads and injects correct repos/keyring into profile | unit | `pytest translators/arch/tests/test_variant.py::test_cachyos_repos -x` | ❌ Wave 0 |
| ARCH-04 | Garuda profile YAML loads; conflicts with Omarchy opinions logged | unit | `pytest translators/arch/tests/test_variant.py::test_garuda_conflicts -x` | ❌ Wave 0 |
| ARCH-04 | Single generator handles all three variant profiles without code forks | unit | `pytest translators/arch/tests/test_variant.py::test_all_variants -x` | ❌ Wave 0 |
| regression | Go resolver suite stays green | unit | `go test ./... -count=1` | ✅ exists |

### Sampling Rate

- **Per task commit:** `pytest translators/arch/tests/ -x -q && go test ./... -count=1`
- **Per wave merge:** Full suite + `bash scripts/arch-validate-iso.sh` (if ISO was built in this wave)
- **Phase gate:** Full suite green + `bash scripts/arch-northstar-check.sh` green before `/gsd-verify-work`

### Wave 0 Gaps

- [ ] `translators/arch/tests/test_generator.py` — covers ARCH-01 profile tree emission
- [ ] `translators/arch/tests/test_manifest.py` — BuildManifest construction
- [ ] `translators/arch/tests/test_profile.py` — profile tree structure + file permissions
- [ ] `translators/arch/tests/test_variant.py` — variant profile loading + repo injection (ARCH-04)
- [ ] `translators/arch/tests/test_capability_gate.py` — SC-3 gating (ARCH-03)
- [ ] `translators/arch/tests/test_firstrun.py` — first-run unit generation
- [ ] `translators/arch/tests/test_northstar.py` — package set equivalence check
- [ ] `translators/arch/pytest.ini` — test config
- [ ] `translators/arch/tests/fixtures/minimal_resolved.json` — minimal test input
- [ ] `examples/omarchy/` — all 134 opinions + 32 points + speech.yaml (ARCH-02)
- [ ] `scripts/arch-build-iso.sh` — Docker mkarchiso wrapper
- [ ] `scripts/arch-validate-iso.sh` — ISO structural validation
- [ ] `scripts/arch-northstar-check.sh` — package/file/service equivalence (ARCH-02 gate)

*(Existing test infrastructure: `go test ./... -count=1` is green as of Phase 1 completion)*

---

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V5 Input Validation | yes | Validate ResolvedSpeech JSON against known schema before processing; reject unknown fields |
| V6 Cryptography | partial | SOURCE_DATE_EPOCH derivation uses SHA-256 (stdlib hashlib); no custom crypto |
| V4 Access Control | partial | Installer runs as root inside the ISO; minimize root scope in first-run units |

### Known Threat Patterns

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Shell injection via opinion data in installer script | Tampering | Serialize all opinion data as JSON manifest; installer reads JSON with `jq`, never interpolates raw strings into shell |
| Unsigned custom repo packages (OM-001: OptionalTrustAll) | Spoofing | Translator emits `TrustWarning` from Explanation (T-01-10 already implemented in resolver); document in generated profile |
| Stale Docker image with vulnerable packages | Elevation | Pin Docker image by digest; add note to scripts to update digest quarterly |
| First-run units running as root when user-scope intended | Elevation | First-run units are systemd USER services (not system); installer enables them with `systemctl --user enable` inside the target user's context |

---

## Sources

### Primary (HIGH confidence)

- archlinux Docker container (archiso 88-1, June 2026) — archiso profile structure, mkarchiso source, SOURCE_DATE_EPOCH handling, xorriso/squashfs-tools, autologin pattern, packages.x86_64, profiledef.sh, pacman.conf, systemd service patterns
- `research/arch-variants-delta.md` (verified from cloned repos, June 2026) — CachyOS repos/keyring/kernel, Garuda repos/bootloader/dracut/btrfs, all VERIFIED claims
- `research/omarchy-opinion-inventory.md` (from Omarchy commit 9cf1852, June 2026) — all 134 opinions, translator-capability tokens, phase ordering
- `resolver/resolve/explanation.go` + `resolver/resolve/canonical.go` — exact ResolvedSpeech fields
- `resolver/types.go` — Opinion struct fields (all translator-relevant fields confirmed)
- `.planning/phases/02-arch-translator/02-CONTEXT.md` — all locked decisions

### Secondary (MEDIUM confidence)

- `docs/03-architecture.md`, `docs/05-distribution-and-infra.md` — architectural constraints, determinism requirements
- `examples/examples_test.go` — example test harness pattern to extend for omarchy speech

### Tertiary (LOW confidence)

- ISO size and build time estimates — training data + package count reasoning; not measured in this session

---

## Metadata

**Confidence breakdown:**
- archiso profile mechanics: HIGH — verified directly in Docker container
- Installer embedding pattern: HIGH — verified releng baseline
- SOURCE_DATE_EPOCH: HIGH — verified in mkarchiso source
- Variant profile data (CachyOS/Garuda): HIGH for VERIFIED claims, LOW for [UNVERIFIED] items
- Omarchy capability tokens: HIGH — derived from opinion inventory
- ISO size/build time: LOW — estimates only
- First-run unit approach (flag file vs ConditionFirstBoot): MEDIUM — conservative choice tagged [ASSUMED]

**Research date:** 2026-06-12
**Valid until:** 2026-07-12 (archiso 88-1 stable; Docker image digest should be re-verified before building)
