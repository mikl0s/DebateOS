# Resolver Edge-Case Corpus

**Omarchy pin:** `9cf1852525a5f7de26d3162db9d61e2f5c1d5523` (version 4.0.0.alpha, no git tags)
**Variant evidence pins:**
- CachyOS-PKGBUILDS: `860f2283198059a05d7aa56fe434d80300ee9c56`
- CachyOS-Settings: `b1aedc79d4f5edab86eb9d22a972ba6994c49b26`
- Garuda pkgbuilds: `1dc0c910447288b16021ebd94ec75c53b8255499`

**Purpose:** EC-NNN Given/When/Then conflict scenarios seeding the Phase 1 resolver TDD harness
(D19). Each EC-NNN maps 1:1 to a future resolver test. All scenarios are either evidence-backed
(traceable to a real collision discovered in the cloned sources above) or synthesized (covering
docs/04 resolution rules not yet observed in the variant study, explicitly tagged).

**Provenance tags:**
- `provenance: evidence-backed` — traceable to a real collision in the cloned Omarchy/variant sources
- `provenance: synthesized` — scenario constructed to satisfy docs/04 rule coverage; rationale given

**EC-NNN ID scheme:** sequential three-digit integer, zero-padded (EC-001 through EC-100).
Assignment order: Class 1 (foundation pre-seeded vs user), Class 2 (repo priority),
Class 3 (cross-variant effectuation), Class 4 (docs/04 rule coverage), Class 5 (CachyOS kernel),
Class 6 (Garuda theming).

---

## Class 1: Foundation Pre-seeded Opinion vs User Opinion

### EC-001

**Title:** Garuda pre-seeded snapper root config vs Omarchy snapper root config

**Given:** A speech targets Garuda as its foundation. Garuda's `snapper-support` package
pre-seeds a snapper root configuration at `/etc/snapper/configs/root` using the
`snapper-template-garuda` template (NUMBER_LIMIT=10, NUMBER_LIMIT_IMPORTANT=5,
TIMELINE_CREATE=no). The speech includes Omarchy's `snapshot-policy/omarchy-snapper-root`
opinion (OM-099), which also configures `/etc/snapper/configs/root` with the Omarchy btrfs
subvolume layout and disables btrfs quotas.

**When:** The resolver evaluates the speech on the Garuda foundation.

**Then:** The resolver detects a required-vs-required conflict between the foundation's
pre-seeded snapper config and the speech opinion's snapper config (both write to
`/etc/snapper/configs/root`). The resolver emits:
> "Hard conflict: both the Garuda foundation and your speech include a required snapper root
> configuration. They cannot coexist because both write `/etc/snapper/configs/root`. You must
> either drop one opinion or provide a patch opinion that merges the two configurations."
The resolver halts and requires user action (drop one, or add a patch).

provenance: evidence-backed (Garuda `snapper-support/PKGBUILD`: `depends=(snapper snap-pac grub-btrfs)`;
Omarchy `install/login/limine-snapper.sh` creates snapper root config)

---

### EC-002

**Title:** Garuda pre-seeded GRUB bootloader vs Omarchy Limine bootloader

**Given:** A speech targets Garuda as its foundation. Garuda's foundation pre-seeds GRUB as
a required bootloader via `garuda-hooks` (which installs `99-update-grub.hook`, `grub-install.hook`,
and depends on `grub` + `update-grub`). The speech includes Omarchy's `bootloader-config/limine`
opinion (OM-099), which installs limine, registers a Limine EFI boot entry, and removes
the archinstall-created GRUB entry.

**When:** The resolver evaluates the speech on the Garuda foundation.

**Then:** The resolver detects a required-vs-required hard conflict between the foundation's
GRUB bootloader opinion and the speech's Limine bootloader opinion. The resolver emits:
> "Hard conflict: the Garuda foundation requires GRUB (garuda-hooks depends on grub; grub-install
> and update-grub hooks run on every kernel upgrade). Your speech also includes a required Limine
> bootloader opinion. A system can only have one active bootloader. You must drop one bootloader
> opinion. No known patch opinion exists for this conflict; it requires a manual choice."
The resolver halts.

provenance: evidence-backed (Garuda `garuda-hooks/PKGBUILD`: `depends=('grub' 'update-grub')`;
Omarchy `install/login/limine-snapper.sh` installs limine and removes archinstall boot entry)

---

### EC-003

**Title:** Garuda Dr460nized SDDM theme vs Omarchy SDDM theme

**Given:** A speech targets Garuda as its foundation. Garuda's `garuda-dr460nized` package
pre-seeds the Dr460nized SDDM theme as a required theming opinion. The speech includes
Omarchy's `theming/sddm-omarchy` opinion (OM-097), which deploys the Omarchy SDDM theme
to `/usr/share/sddm/themes/omarchy` and sets it as the active theme in
`/etc/sddm.conf.d/omarchy.conf`.

**When:** The resolver evaluates the speech on the Garuda foundation.

**Then:** The resolver detects a conflict: two theming opinions both set the SDDM active theme.
Omarchy's SDDM theming is required (it controls auto-login and the Wayland session config).
Garuda's SDDM theming is also a required pre-seeded opinion. The resolver emits:
> "Conflict: the Garuda foundation includes a required SDDM theme (Dr460nized) and your speech
> includes a required SDDM theme (Omarchy). Only one SDDM theme can be active. This is a
> required-vs-required conflict. You must drop one or provide a patch opinion that bridges them."

provenance: evidence-backed (Garuda `garuda-dr460nized/PKGBUILD`: KDE + SDDM theming;
Omarchy `install/login/sddm.sh` deploys SDDM theme and config drop-ins)

---

### EC-004

**Title:** CachyOS pre-seeded linux-cachyos kernel vs Omarchy vanilla linux kernel

**Given:** A speech targets CachyOS as its foundation. CachyOS pre-seeds `linux-cachyos`
(EEVDF scheduler, CPU-optimized) as its required default kernel. The speech includes
Omarchy's `package-install/linux-kernel` opinion (OM-006 range), which declares `linux`
(vanilla Arch kernel) as a required package.

**When:** The resolver evaluates the speech on the CachyOS foundation.

**Then:** The resolver detects a required-vs-required conflict between the CachyOS
pre-seeded `linux-cachyos` kernel and the speech's `linux` kernel opinion. The resolver emits:
> "Hard conflict: the CachyOS foundation requires the linux-cachyos kernel (pre-seeded;
> EEVDF scheduler, x86-64 optimized). Your speech also declares linux (vanilla Arch kernel)
> as required. A system can have multiple kernels installed, but requiring both as the primary
> boot kernel is a conflict. You must either drop the linux opinion (and rely on linux-cachyos)
> or provide a patch opinion that installs linux alongside linux-cachyos for dual-boot selection."

provenance: evidence-backed (CachyOS `linux-cachyos` PKGBUILD; CachyOS pacman.conf shows
linux-cachyos as default; Omarchy base packages include `linux`)

---

### EC-005

**Title:** Synthesized: sysctl key collision between two opinions writing the same kernel parameter

**Note (correction):** The original version of this scenario incorrectly claimed that
CachyOS `fs.file-max` (a sysctl kernel parameter) and Omarchy OM-038 `DefaultLimitNOFILE`
(a systemd per-process RLIMIT set via `system.conf.d/` and `user.conf.d/` drop-ins) collide
on the same sysctl key. They do NOT — they are distinct mechanisms operating in different
namespaces (`/usr/lib/sysctl.d/` vs `/etc/systemd/system.conf.d/`; kernel parameter vs
per-process resource limit). CachyOS and Omarchy OM-038 are additive, not conflicting:
CachyOS raises the global fd ceiling (`fs.file-max=2097152`) and Omarchy sets the per-process
limit (`DefaultLimitNOFILE=65536:524288`) — both can coexist.

This scenario has been reclassified as synthesized to demonstrate the general per-key sysctl
collision detection capability required by SR-016. The synthesized scenario uses a hypothetical
second sysctl opinion that DOES write `fs.inotify.max_user_watches` — the same key as OM-037.

**Given:** A speech includes Omarchy's `sysctl-param/increase-file-watchers` opinion (OM-037),
which writes `fs.inotify.max_user_watches = 524288` to `/etc/sysctl.d/99-omarchy-watchers.conf`.
The speech also includes a hypothetical additional opinion `sysctl-param/extra-inotify-tuning`
that writes `fs.inotify.max_user_watches = 1048576` to `/etc/sysctl.d/90-extra-tuning.conf`.
Both opinions claim the same sysctl key.

**When:** The resolver evaluates the speech.

**Then:** The resolver detects a sysctl key collision: two opinions in the same speech both
declare ownership of `fs.inotify.max_user_watches` with different values. The resolver emits:
> "Sysctl key collision: 'fs.inotify.max_user_watches' is written by two opinions in this
> speech — increase-file-watchers (OM-037, value=524288) and extra-inotify-tuning
> (value=1048576). The effective value depends on drop-in numeric prefix order, which is
> non-deterministic across opinion composition. A patch opinion is needed to merge these into
> a single authoritative drop-in, or one must be dropped."

provenance: synthesized (demonstrates per-key sysctl collision detection required by SR-016;
OM-037 is evidence-backed from Omarchy `install/config/increase-file-watchers.sh`;
the second opinion is synthesized to exercise the collision rule. CachyOS `fs.file-max` vs
Omarchy OM-038 `DefaultLimitNOFILE` are NOT a sysctl collision — they are different
mechanisms; see Note above.)

---

## Class 2: Repo Priority Conflicts

### EC-010

**Title:** CachyOS repo priority order requires custom-repo opinions to declare ordering

**Given:** A speech targets CachyOS (x86-64-v3) as its foundation. The speech includes
Omarchy's `custom-repo/omarchy-repo` opinion (OM-001), which adds the `[omarchy]` repo
to pacman.conf. CachyOS pre-seeds `[cachyos-v3]`, `[cachyos-core-v3]`, `[cachyos-extra-v3]`,
and `[cachyos]` repos, all of which must appear above `[core]`/`[extra]`/`[multilib]` in
pacman.conf for the CPU-optimized packages to shadow the standard Arch packages.

**When:** The resolver assembles the final pacman.conf repo list.

**Then:** The resolver detects an ordering ambiguity: both the foundation's custom repos
(CachyOS v3 repos, priority: above-arch) and the speech's custom repo (`[omarchy]`,
priority: after-arch per its declaration) need to be placed relative to `[core]`/`[extra]`.
The resolver emits:
> "Repo ordering decision needed: the CachyOS foundation's repos must appear before [core]/
> [extra] to shadow packages with optimized builds. Your speech adds [omarchy] (after standard
> Arch repos). This ordering can be resolved automatically: [cachyos-v3] > [cachyos-core-v3] >
> [cachyos-extra-v3] > [cachyos] > [core] > [extra] > [multilib] > [omarchy]. Confirm or adjust."

provenance: evidence-backed (CachyOS docker/pacman-v3.conf verified in commit 2f032fd;
Omarchy `default/pacman/pacman-stable.conf` shows `[omarchy]` placed after standard repos)

---

### EC-011

**Title:** Garuda Chaotic-AUR repo vs Omarchy repo — relative priority

**Given:** A speech targets Garuda as its foundation. Garuda pre-seeds `[garuda]` and
`[chaotic-aur]` (above standard Arch repos). The speech includes Omarchy's
`custom-repo/omarchy-repo` opinion (OM-001) adding `[omarchy]`. No explicit priority
declaration is given for `[omarchy]` relative to Garuda's repos.

**When:** The resolver assembles the final repo list for the Garuda foundation.

**Then:** The resolver detects that the speech's custom-repo opinion lacks an explicit
priority relative to the foundation's repos. The resolver emits:
> "Repo priority undeclared: your speech adds [omarchy] but does not declare its priority
> relative to the Garuda foundation's [garuda] and [chaotic-aur] repos. Current resolved
> order (applying nice-to-have default): [garuda] > [chaotic-aur] > [core] > [extra] >
> [multilib] > [omarchy]. If any package exists in both [omarchy] and [chaotic-aur], the
> Chaotic-AUR version will be used. Accept this order or add a priority: above-chaotic-aur
> declaration to the [omarchy] opinion."

provenance: evidence-backed (Garuda `data/pacman-default.conf` shows [chaotic-aur] below standard
Arch repos; Omarchy pacman-stable.conf positions [omarchy] after standard repos)

---

### EC-012

**Title:** Nice-to-have custom repo opinion dropped in favour of required repo

**Given:** A speech includes two custom-repo opinions: a required `custom-repo/omarchy-repo`
opinion (OM-001, required for all subsequent Omarchy package installs) and a nice-to-have
`custom-repo/extra-gaming-repo` opinion. Both are present in the same speech. The `[extra-gaming-repo]`
has the same priority slot as `[omarchy]`, creating a conflict on the priority list.

**When:** The resolver evaluates the speech.

**Then:** The required custom-repo opinion for `[omarchy]` wins. The nice-to-have
`[extra-gaming-repo]` opinion is dropped visibly. The resolver emits:
> "Required beats nice-to-have: [omarchy] repo is required by this speech (all Omarchy
> package opinions depend on it). [extra-gaming-repo] is nice-to-have and would occupy the
> same priority slot. [extra-gaming-repo] has been dropped from this speech. You can add it
> back as an additional slot if needed."

provenance: synthesized (demonstrates required-beats-nice-to-have rule for custom-repo category;
based on OM-001 dependency chain requiring the omarchy repo)

---

## Class 3: Cross-Variant Effectuation Differences

### EC-020

**Title:** Same mesa package opinion effectuated differently on CachyOS vs vanilla Arch

**Given:** A speech includes `package-install/graphics-stack` (OM-012 range), which
requires the `mesa` graphics library package. The speech targets CachyOS x86-64-v3.
CachyOS's `[cachyos-extra-v3]` repo provides a `mesa` package built with `-O3 -march=x86-64-v3`
LTO/PGO/BOLT optimizations. On vanilla Arch, `[extra]` provides the standard `mesa` build.

**When:** The resolver translates the speech for the CachyOS foundation.

**Then:** The resolver recognizes the mesa opinion is satisfied by the CachyOS-optimized
package (same logical intent, different binary). No conflict. The resolver emits:
> "Package mesa: satisfied by CachyOS x86-64-v3 optimized build (cachyos-extra-v3/mesa).
> This is the preferred build for your CachyOS x86-64-v3 target. No action needed."

provenance: evidence-backed (CachyOS docker/pacman-v3.conf confirmed; CachyOS optimized
packages include mesa per CachyOS documentation)

---

### EC-021

**Title:** Omarchy linux-headers opinion requires variant-aware package name resolution

**Given:** A speech includes Omarchy's Intel PTL kernel opinion (OM-074), which installs
`linux-ptl` and `linux-ptl-headers` for Dell XPS Panther Lake systems. On vanilla Arch
the headers package is `linux-headers`. On CachyOS, the CachyOS kernel headers are
`linux-cachyos-headers`. The speech is evaluated against a CachyOS foundation.

**When:** The resolver attempts to resolve the `linux-headers` dependency for a Hyprland
opinion on CachyOS.

**Then:** The resolver detects that `linux-headers` (vanilla name) does not exist in
CachyOS repos; `linux-cachyos-headers` must be used instead. The resolver emits:
> "Package name translation required: your speech declares a dependency on linux-headers,
> but the CachyOS foundation provides linux-cachyos-headers. The CachyOS translator will
> use linux-cachyos-headers to satisfy this dependency. If you specifically need vanilla
> linux-headers, add a required kernel opinion for linux (vanilla) alongside this speech."

provenance: evidence-backed (CachyOS linux-cachyos PKGBUILD verified; Omarchy
`install/config/hardware/intel/ptl-kernel.sh` installs linux-ptl-headers)

---

### EC-022

**Title:** Snapper configuration idempotency on Garuda (pre-seeded vs speech opinion)

**Given:** A speech targets Garuda and includes Omarchy's `snapshot-policy/snapper-root`
opinion. Garuda's foundation already ran snapper init for the root configuration during
its own install (via `snapper-support`). When the speech translator runs, it attempts
to run `snapper -c root create-config /` again.

**When:** The translator executes the speech on Garuda.

**Then:** This is NOT a resolver conflict (the resolver already flagged EC-001); this is
a translation-time idempotency note. The resolver, during pre-check, emits:
> "Warning: the Garuda foundation may have already created a snapper root config.
> If applying this speech over an existing Garuda system, the snapper-root opinion
> (OM-099) is likely a no-op or will fail with 'Config already exists'. A patch
> opinion for Garuda+snapper should use `snapper set-config` instead of `create-config`."

provenance: evidence-backed (Garuda `snapper-support/PKGBUILD` creates root config;
Omarchy `install/login/limine-snapper.sh` also creates root config)

---

### EC-023

**Title:** Bluetooth service enable idempotent across variants

**Given:** A speech includes Omarchy's `service-enable/bluetooth` opinion (OM-064), which
enables `bluetooth.service`. On CachyOS, `cachyos-settings` also enables `systemd-resolved`;
neither CachyOS nor Garuda pre-seeds a bluetooth enable. The speech targets Garuda.

**When:** The resolver evaluates the bluetooth service-enable opinion on the Garuda foundation.

**Then:** No conflict — neither Garuda nor CachyOS pre-seeds bluetooth service enablement.
The opinion is applied. The resolver emits:
> "Service bluetooth: enabled. No foundation pre-seeded opinion conflicts with this.
> Service will be enabled by the Arch translator."

provenance: evidence-backed (CachyOS-Settings `cachyos-settings.install` services do not include
bluetooth; Garuda `garuda-common-settings.install` does not include bluetooth)

---

## Class 4: docs/04 Rule Coverage

### EC-030

**Title:** Required kernel opinion drops nice-to-have audio codec that depends on same kernel

**Given:** A speech includes a required `kernel-install/linux-ptl` opinion (OM-074)
for Intel PTL hardware, which installs `linux-ptl` and removes `linux`. A nice-to-have
`package-install/sof-firmware` opinion (OM-077) exists in the speech. OM-077 has a
condition: `omarchy-hw-intel-ptl AND NOT omarchy-hw-match "XPS"`. On a non-XPS PTL system
OM-077 is nice-to-have (its condition is met) but depends on the PTL kernel being present.
A second nice-to-have `package-install/broadcom-wl-dkms` opinion also depends on `linux-headers`.
After the PTL kernel install removes `linux`, `linux-headers` is no longer installed.
`broadcom-wl-dkms` requires `linux-headers` (the vanilla package).

**When:** The resolver evaluates the speech with `linux-ptl` required and `broadcom-wl-dkms` nice-to-have.

**Then:** The required `linux-ptl` opinion wins. The nice-to-have `broadcom-wl-dkms` opinion
is automatically dropped because its `linux-headers` dependency is no longer satisfiable
after `linux` is removed. The resolver emits:
> "Required beats nice-to-have: linux-ptl (required, Intel PTL hardware) removes linux and
> linux-headers. The nice-to-have broadcom-wl-dkms opinion depends on linux-headers
> (which is no longer available after linux removal). broadcom-wl-dkms has been dropped
> from this speech. If you have Broadcom WiFi on PTL hardware, add linux-ptl-headers
> to your speech instead."

provenance: evidence-backed (Omarchy `install/config/hardware/intel/ptl-kernel.sh` removes
`linux`; `install/config/hardware/fix-bcm43xx.sh` (OM-090) depends on DKMS kernel headers)

---

### EC-031

**Title:** Required vanilla kernel vs required CachyOS kernel — hard conflict, no patch

**Given:** A speech includes two required kernel opinions: `kernel-install/linux-vanilla`
(requires `linux`, `linux-headers`) and `kernel-install/linux-cachyos` (requires
`linux-cachyos`, `linux-cachyos-headers`). Both are marked required. No patch opinion
exists that makes both coexist as the primary boot kernel.

**When:** The resolver evaluates the speech.

**Then:** The resolver detects a required-vs-required hard conflict on the primary boot
kernel role. Multiple kernels can be installed, but both opinions assert they control the
primary boot kernel slot. The resolver emits:
> "Hard conflict: linux-vanilla (required) and linux-cachyos (required) both claim the
> primary boot kernel role. A system can have multiple kernels installed but only one
> can be the primary default. No patch opinion exists for this conflict. You must drop
> one kernel opinion. Suggested: on CachyOS, drop linux-vanilla and keep linux-cachyos
> (the foundation's optimized kernel). On vanilla Arch, drop linux-cachyos."

provenance: evidence-backed (CachyOS linux-cachyos PKGBUILD; Omarchy base packages include
linux; two required kernel opinions is a real scenario when composing Omarchy speech on CachyOS)

---

### EC-032

**Title:** Required initramfs method conflict resolved by patch opinion

**Given:** A speech targets Garuda and includes Omarchy's `initramfs-method/mkinitcpio` opinion
(OM-002 + OM-099 combined, required — Omarchy's entire login phase depends on mkinitcpio).
Garuda's foundation pre-seeds `initramfs-method/dracut` as required
(`garuda-dracut-support` with `conflicts=('mkinitcpio' ...)`). A community-contributed
patch opinion `patch/dracut-omarchy-bridge` exists, which rewrites Omarchy's UKI generation
steps to use dracut hooks instead of mkinitcpio hooks and removes the mkinitcpio dependency.

**When:** The resolver evaluates the speech on the Garuda foundation.

**Then:** The resolver detects a required-vs-required conflict on initramfs method (mkinitcpio
vs dracut). It finds the patch opinion `patch/dracut-omarchy-bridge`. The resolver emits:
> "Hard conflict: your speech requires mkinitcpio (Omarchy login phase), but Garuda requires
> dracut (garuda-dracut-support conflicts mkinitcpio). A patch opinion is available:
> 'dracut-omarchy-bridge' rewrites Omarchy's initramfs steps for dracut compatibility.
> Apply the patch? [Accept / Decline / View details]"
If accepted, the patch opinion is added and the conflict is resolved.

provenance: synthesized (demonstrates required-vs-required-with-patch rule; the dracut/mkinitcpio
conflict is evidence-backed from Garuda `garuda-dracut-support/PKGBUILD` with `conflicts=('mkinitcpio')`;
the patch opinion itself is synthesized to show the resolution path)

---

### EC-033

**Title:** Two nice-to-have terminal emulator opinions — resolver picks sensible default

**Given:** A speech includes two nice-to-have terminal emulator opinions:
`package-install/terminal-foot` (installs `foot`, nice-to-have) and
`package-install/terminal-ghostty` (installs `ghostty`, nice-to-have). Neither is required.
Both declare `conflicts: [other-terminal]`. No patch opinion exists.

**When:** The resolver evaluates the speech.

**Then:** The resolver applies the nice-vs-nice default rule. It picks the first-listed
opinion as the default (foot, which is OM-007's terminal — the Omarchy default). The resolver emits:
> "Nice-to-have conflict: foot and ghostty are both nice-to-have terminal emulators and
> declare mutual conflicts. Default selected: foot (Omarchy's curated default terminal).
> If you prefer ghostty, drop the foot opinion from your speech. Both can be kept
> if you remove the conflict declarations."

provenance: synthesized (demonstrates nice-vs-nice default rule; foot is evidence-backed
as Omarchy's terminal in OM-007; ghostty is a plausible alternative terminal opinion
a curator could add)

---

### EC-034

**Title:** Patch opinion overrides required-vs-required conflict hierarchy

**Given:** A speech includes two required opinions: `custom-repo/omarchy-repo` (OM-001,
required — all Omarchy packages depend on it) and `custom-repo/cachyos-repo` (required
by the CachyOS foundation). Both add repos to pacman.conf. Normally a required-vs-required
conflict would require user action. A patch opinion `patch/merged-pacman-conf` exists that
produces a single pacman.conf drop-in combining both repo blocks in the correct priority order
(CachyOS repos above Arch, Omarchy repo below Arch).

**When:** The resolver evaluates the speech with both required custom-repo opinions.

**Then:** The resolver finds the patch opinion. Instead of halting with a hard conflict,
it applies the patch. The resolver emits:
> "Required-vs-required conflict on pacman.conf resolved by patch: 'merged-pacman-conf'
> combines [cachyos-v3], [cachyos], [core], [extra], [multilib], [omarchy] into a single
> consistent configuration. Patch applied automatically. Review the merged configuration
> in the resolved speech before building."

provenance: synthesized (demonstrates patch-overrides-hierarchy rule; custom-repo ordering
conflict is evidence-backed from CachyOS/Omarchy co-existence requirements)

---

### EC-035

**Title:** Three-hop dependency chain produces correct toposort install order

**Given:** A speech includes three opinions with ordering constraints:
- `npm-global-install/ai-tools` (OM-023): requires mise+Node.js to be installed first
- `config-dotfile/mise-work` (OM-041): requires mise to be installed (OM-009) first
- `package-install/dev-tools` (OM-009): includes mise; no before-dependency

Ordering constraints: OM-023 after OM-041, OM-041 after OM-009.

**When:** The resolver runs the topological sort on the speech.

**Then:** The resolver produces the correct install order: OM-009 → OM-041 → OM-023.
The resolver emits:
> "Install order (topological sort): [1] dev-tools (OM-009, mise installed), [2] mise-work
> config (OM-041, Node.js installed via mise), [3] AI tools npm install (OM-023, npm
> available after mise). Order is deterministic and reproducible."

provenance: evidence-backed (Omarchy `install/config/mise-work.sh` must run before
`install/packaging/npm.sh`; OM-023 depends on OM-009 + OM-041 explicitly)

---

### EC-036

**Title:** Circular ordering constraint — hard cycle error naming both opinions

**Given:** A speech includes two opinions with a circular ordering constraint:
- `service-enable/docker` (OM-043): declares `must-install-after: config/docker-dns`
- `config-dotfile/docker-dns` (synthesized): declares `must-install-after: service-enable/docker`

This creates a cycle: docker-service → docker-dns → docker-service.

**When:** The resolver runs the topological sort.

**Then:** The resolver detects the cycle and emits a hard error. The resolver emits:
> "Cycle detected in install ordering: docker-service (OM-043) declares must-install-after
> docker-dns-config, but docker-dns-config declares must-install-after docker-service.
> This cycle has no valid topological ordering. Remove one of the ordering constraints
> or introduce an intermediate opinion to break the cycle."
The resolver halts; the speech cannot be compiled.

provenance: synthesized (demonstrates cycle-detection rule; docker service + DNS config
are evidence-backed from Omarchy OM-043 `install/config/docker.sh`; the circular constraint
is synthesized to exercise this rule)

---

### EC-037

**Title:** NVIDIA driver opinion silently skipped — hardware condition false

**Given:** A speech includes Omarchy's `hardware-conditional/nvidia-driver` opinion (OM-068),
which installs NVIDIA drivers. The opinion declares condition: `lspci | grep -qi 'nvidia'`.
The speech is composed for a machine that has declared hardware: `gpu: amd-radeon-rx-7600`.
No NVIDIA GPU is declared in the hardware profile.

**When:** The resolver evaluates the hardware-conditional opinion against the declared hardware.

**Then:** The resolver evaluates the condition as false (no NVIDIA GPU in declared hardware).
The NVIDIA opinion is silently skipped. The resolver emits:
> "Skipped (hardware condition false): nvidia-driver (OM-068) — condition requires an NVIDIA
> GPU (lspci nvidia match). Declared hardware has no NVIDIA GPU (amd-radeon-rx-7600 detected).
> This opinion is not applied in your resolved speech. Install the speech on hardware
> with an NVIDIA GPU to activate this opinion."

provenance: evidence-backed (Omarchy `install/config/hardware/nvidia.sh` (OM-068) uses
`lspci | grep -qi 'nvidia'` condition; hardware-conditional skip is a required resolver behavior
per docs/04)

---

### EC-038

**Title:** Apple T2 opinion block applied — hardware condition matches declared hardware

**Given:** A speech includes Omarchy's `hardware-conditional/apple-t2` opinion (OM-088),
which installs the T2 kernel (`linux-t2`), audio firmware, Broadcom WiFi, fan control daemon,
and Touch Bar driver. The opinion declares condition: `lspci -nn | grep -q "106b:1801"`.
The speech is composed for hardware declaring `pci: [106b:1801]` (Apple T2 security chip).

**When:** The resolver evaluates the hardware-conditional opinion against the declared hardware.

**Then:** The resolver evaluates the condition as true (T2 PCI ID present in declared hardware).
The entire T2 opinion block is applied. The resolver emits:
> "Applied (hardware condition true): apple-t2 (OM-088) — Apple T2 security chip detected
> (PCI 106b:1801). Installing: linux-t2, apple-t2-audio-config, apple-bcm-firmware, t2fanrd,
> tiny-dfr. Note: this opinion also adds the arch-mact2 repo (SigLevel=Never) for linux-t2
> updates. Review this trust level before building."

provenance: evidence-backed (Omarchy `install/config/hardware/apple/fix-t2.sh` (OM-088);
PCI ID `106b:1801` is the actual Apple T2 PCI identifier in the script; SigLevel=Never is
verified in Omarchy `install/post-install/pacman.sh`)

---

## Class 5: CachyOS Kernel Collision

### EC-040

**Title:** User speech specifies vanilla linux kernel; CachyOS variant profile pre-seeds linux-cachyos

**Given:** A user composes a speech that includes `package-install/linux-vanilla` (required,
installs `linux` and `linux-headers`) and targets the `cachyos` variant profile. The CachyOS
variant profile pre-seeds `linux-cachyos` (EEVDF scheduler, x86-64 optimized) as a required
kernel opinion. Both opinions claim the primary boot kernel role.

**When:** The resolver evaluates the speech against the CachyOS variant profile.

**Then:** A required-vs-required conflict is raised. The resolver emits:
> "Hard conflict: CachyOS variant pre-seeds linux-cachyos (required: EEVDF scheduler,
> CPU-optimized) and your speech includes linux-vanilla (required). On CachyOS, linux-cachyos
> replaces linux for all officially supported builds. Recommendation: drop linux-vanilla
> and use linux-cachyos. If you need linux alongside linux-cachyos (dual-kernel boot),
> change linux-vanilla to nice-to-have and add a bootloader entry for it."

provenance: evidence-backed (CachyOS linux-cachyos PKGBUILD at commit 39d9d12; CachyOS
pacman.conf positions linux-cachyos as the default kernel; Omarchy uses linux in base packages)

---

### EC-041

**Title:** CachyOS CPU arch level mismatch — speech targets v3, hardware declares v4 capable

**Given:** A speech targets the `cachyos-v3` variant profile (x86-64-v3 optimized repos).
The hardware profile declares `cpu_features: [avx512f, avx512bw, avx512vl]` (AVX-512 capable,
qualifying for x86-64-v4). The v3 variant profile's repos provide x86-64-v3 optimized builds.

**When:** The resolver evaluates the speech CPU arch level against the declared hardware capabilities.

**Then:** The resolver detects the arch level mismatch: the hardware can run v4-optimized code
but the speech only targets v3. The resolver emits:
> "Note: your hardware supports x86-64-v4 (AVX-512 capable) but your speech targets
> CachyOS x86-64-v3. You can upgrade to cachyos-v4 variant profile for better performance.
> This is not a conflict — v3 will run correctly on v4 hardware — but v4 would provide
> further optimization. Switch to cachyos-v4 variant profile? [Upgrade / Keep v3]"

provenance: evidence-backed (CachyOS `script-v3-v4.sh` in linux-cachyos repo shows v3/v4
selection logic; CachyOS pacman-v4.conf verified at commit 2f032fd; CPU arch feature
detection is a known CachyOS install-time decision)

---

### EC-042

**Title:** CachyOS kernel opinion + Omarchy Intel PTL kernel — multi-kernel ordering

**Given:** A speech includes both CachyOS's pre-seeded `linux-cachyos` (required, from
variant profile) and Omarchy's hardware-conditional `linux-ptl` opinion (OM-074, applied
because the hardware declares a Dell XPS PTL system with `omarchy-hw-intel-ptl=true`).
OM-074 removes `linux` but not `linux-cachyos`.

**When:** The resolver evaluates the speech on a CachyOS variant with PTL hardware.

**Then:** The resolver determines that `linux-ptl` targets PTL-specific audio patches
not in `linux-cachyos`. Both kernels serve different needs (PTL driver patches vs EEVDF
scheduler). No hard conflict — multiple kernels are allowed. The resolver emits:
> "Multiple kernel notice: linux-cachyos (CachyOS default, EEVDF scheduler) and linux-ptl
> (hardware-conditional, Dell XPS PTL audio patches) will both be installed. The bootloader
> will present both at boot. Default boot kernel: linux-ptl (per OM-074 priority declaration)."

provenance: evidence-backed (Omarchy `install/config/hardware/intel/ptl-kernel.sh` (OM-074)
installs linux-ptl and removes linux but not other kernels; CachyOS linux-cachyos is a
separate kernel from linux)

---

## Class 6: Garuda Theming Collision

### EC-050

**Title:** Omarchy SDDM theme vs Garuda Dr460nized SDDM theme — active theme slot collision

**Given:** A speech targets Garuda. The speech includes Omarchy's `theming/sddm-omarchy`
opinion (OM-097), which deploys the Omarchy SDDM theme and writes
`/etc/sddm.conf.d/omarchy-theme.conf` setting `Theme=omarchy`. Garuda's foundation
pre-seeds `garuda-dr460nized` which deploys its SDDM theme and sets `Theme=garuda-dr460nized`
via a `/etc/sddm.conf.d/garuda.conf` drop-in.

**When:** The resolver evaluates the speech on Garuda.

**Then:** Both SDDM theme opinions write to the SDDM active theme slot (a single-value
config key). Required-vs-required conflict (OM-097 is required for Omarchy's Wayland
auto-login session; Garuda's theming is pre-seeded as required). The resolver emits:
> "Hard conflict: Omarchy SDDM theme (required, sets Theme=omarchy + Wayland session config)
> vs Garuda Dr460nized SDDM theme (foundation pre-seeded, sets Theme=garuda-dr460nized).
> Only one SDDM theme can be active at a time. Drop one or provide a patch opinion that
> preserves the Omarchy Wayland session config while using the Dr460nized visual theme."

provenance: evidence-backed (Garuda `garuda-dr460nized/PKGBUILD` confirmed; Omarchy
`install/login/sddm.sh` (OM-097) deploys SDDM theme and writes sddm.conf.d drop-in)

---

### EC-051

**Title:** Omarchy Plymouth theme vs Garuda Dr460nized Plymouth theme — same theme path

**Given:** A speech targets Garuda. The speech includes Omarchy's `theming/plymouth-omarchy`
opinion (OM-095), which copies the Omarchy Plymouth theme to
`/usr/share/plymouth/themes/omarchy/` and sets it as active via `plymouth-set-default-theme`.
Garuda's foundation pre-seeds `plymouth-theme-dr460nized` (from `garuda-dr460nized`),
which sets `Theme=dr460nized` as the active Plymouth theme.

**When:** The resolver evaluates the speech on Garuda.

**Then:** Both opinions call `plymouth-set-default-theme`, writing different values to the
Plymouth theme config. Two required opinions conflict on the same config key. The resolver emits:
> "Hard conflict: Omarchy Plymouth theme (required, sets omarchy as active theme) vs
> Garuda Dr460nized Plymouth theme (foundation pre-seeded, sets dr460nized as active theme).
> The active Plymouth theme is a single-value config key. Drop one or provide a patch
> opinion that selects a single theme. Note: Plymouth themes are embedded in the initramfs
> (via mkinitcpio); the last `plymouth-set-default-theme` call before initramfs rebuild wins,
> but this is non-deterministic and must be resolved explicitly."

provenance: evidence-backed (Garuda `plymouth-theme-dr460nized/PKGBUILD` confirmed;
Omarchy `install/login/plymouth.sh` (OM-095) calls `plymouth-set-default-theme omarchy`;
both opinions write to same Plymouth active-theme slot)

---

### EC-052

**Title:** Garuda GRUB theme vs Omarchy (no GRUB opinion) — no collision

**Given:** A speech targets Garuda. Garuda pre-seeds `grub-theme-garuda-dr460nized` (the
GRUB boot splash theme). The speech includes Omarchy's full opinion set, which does NOT
include any GRUB theming opinion (Omarchy uses limine, not GRUB; it has no GRUB theme
opinions). No Omarchy opinion writes to GRUB theme paths.

**When:** The resolver evaluates the GRUB theme opinion on Garuda.

**Then:** No conflict — the GRUB theme opinion from Garuda exists without any competing
Omarchy opinion. (Note: EC-002 already flagged the bootloader conflict itself; this
scenario tests that the GRUB THEME opinion specifically has no Omarchy counterpart.) The
resolver emits:
> "No conflict: Garuda GRUB theme (grub-theme-garuda-dr460nized) has no competing opinion
> in your speech. Applied by default (pre-seeded). Note: the bootloader itself (GRUB vs
> limine) is a separate hard conflict flagged in EC-002."

provenance: evidence-backed (Garuda `grub-theme-garuda-dr460nized` confirmed in pkgbuilds;
Omarchy has no GRUB theming opinion; Omarchy `install/login/limine-snapper.sh` uses limine only)

---

## Coverage Matrix

The following table maps the 4 numbered resolution rules from `docs/04-conflict-resolution.md`
and the additional behavioral sections (hardware-conditional, ordering, cycle detection) to
the EC-NNN scenarios above. Every rule and behavioral section has at least one scenario.

(Note: docs/04 contains exactly 4 numbered resolution rules in its "Resolution hierarchy
(precise rules)" section. The additional rows below map named behavioral sections from
docs/04 — `## Ordering`, `## Hardware-aware resolution`, `## Cycle detection` — which are
covered behaviors but are not numbered rules in docs/04.)

| docs/04 Resolution Rule / Behavior | Relevant EC-NNN | Notes |
|------------------------------------|----------------|-------|
| **Rule 1: Required beats nice-to-have** | EC-012, EC-030 | EC-030: required kernel drops nice-to-have DKMS; EC-012: required repo drops nice-to-have repo |
| **Rule 2: Required-vs-required hard conflict** | EC-001, EC-002, EC-003, EC-004, EC-031, EC-050, EC-051 | Multiple real evidence-backed conflicts; EC-031 synthesized for pure-rule demonstration |
| **Rule 3: Required-vs-required with patch** | EC-032, EC-034 | EC-032: dracut/mkinitcpio patch; EC-034: merged pacman.conf patch |
| **Rule 4: Nice-vs-nice sensible default** | EC-033 | Two nice-to-have terminal emulators; resolver picks Omarchy default (foot) |
| **Patch overrides hierarchy (behavioral section)** | EC-032, EC-034 | Patch opinion resolves what would otherwise be a hard-stop conflict |
| **Ordering/toposort (behavioral section)** | EC-035, EC-010, EC-011 | EC-035: three-hop dependency chain; EC-010/EC-011: repo ordering within pacman.conf |
| **Cycle detection (behavioral section)** | EC-036 | Circular must-install-after between two config opinions |
| **Hardware-conditional skip (behavioral section)** | EC-037 | NVIDIA driver skipped — no NVIDIA GPU in declared hardware |
| **Hardware-conditional apply (behavioral section)** | EC-038, EC-040, EC-041, EC-042 | T2 block applied; CachyOS kernel collision with hardware context |
| **Per-key sysctl collision detection (SR-016 capability)** | EC-005 | Synthesized: two opinions writing same sysctl key in same speech |

---

## Summary Statistics

| Class | EC-IDs | Count | Provenance |
|-------|--------|-------|------------|
| Class 1: Foundation pre-seeded vs user | EC-001 to EC-005 | 5 | EC-001, EC-002, EC-003, EC-004 evidence-backed; EC-005 synthesized (corrected — see EC-005 note) |
| Class 2: Repo priority | EC-010 to EC-012 | 3 | EC-010, EC-011 evidence-backed; EC-012 synthesized |
| Class 3: Cross-variant effectuation | EC-020 to EC-023 | 4 | All evidence-backed |
| Class 4: docs/04 rule coverage | EC-030 to EC-038 | 9 | EC-030, EC-031, EC-035, EC-037, EC-038 evidence-backed; EC-032, EC-033, EC-034, EC-036 synthesized |
| Class 5: CachyOS kernel collision | EC-040 to EC-042 | 3 | All evidence-backed |
| Class 6: Garuda theming | EC-050 to EC-052 | 3 | All evidence-backed |
| **Total** | **EC-001 to EC-052** | **27** | **21 evidence-backed, 6 synthesized** |

---

## Synthesized Scenarios — Rationale

| EC-NNN | Synthesized Reason |
|--------|-------------------|
| EC-005 | Original evidence-backed claim was factually wrong (OM-038 writes DefaultLimitNOFILE via systemd drop-in, NOT fs.file-max via sysctl.d; these are different mechanisms). Reclassified as synthesized to demonstrate per-key sysctl collision detection capability (SR-016) using OM-037 as the evidence-backed base |
| EC-012 | Required-beats-nice-to-have for custom-repo category needed; no natural evidence-backed example found |
| EC-032 | Required-vs-required-with-patch rule needs a concrete example; dracut/mkinitcpio conflict is real, patch is synthesized |
| EC-033 | Nice-vs-nice sensible default rule needs a terminal emulator example; Omarchy uses foot but no real two-terminal conflict in the speech |
| EC-034 | Patch-overrides-hierarchy rule needs an explicit example separate from EC-032 |
| EC-036 | Cycle detection must be exercised; no real cycle exists in Omarchy but docker/docker-dns pair is plausible |
