# Phase 4: Debian Translator - Research

**Researched:** 2026-06-13
**Domain:** Debian live-build / preseed / apt, Python translator pattern, Go CLI refactor, Arch-leak audit
**Confidence:** HIGH (codebase), MEDIUM (live-build), MEDIUM (reproducibility)

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Same Python-generator + thin-shell pattern as translators/arch. Extract foundation-neutral
  shared code into translators/common/ if the refactor is clean; otherwise document duplication.
- Same argv-stable entrypoint: `translators/debian/translate <resolved.json> --opinions <dir> --profile <name> --out <dir>`
- Wraps live-build (lb config/lb build) with preseed for fully-unattended install.
  `execution_phase: first-run` → systemd first-boot units (same flag-file pattern as Arch).
- translators/debian/capabilities.json declares supported capability tokens; required+unsupported
  → loud CapabilityError at composition time (SC-2); nice-to-have+unsupported → visible drop.
- profiles/debian.yaml minimum (stable). Profile structure must accommodate Ubuntu post-v1.0.
- Dual-foundation proof: examples/dual-foundation/ — representative foundation-neutral speech →
  both translators → valid profile trees; ISO build is the slow gate (deferred if devtmpfs-blocked).
- Leaked-Arch-assumption audit: schema fields, resolver logic, shared generator. Fix genuine
  leaks; document correctly-isolated ones.
- docs/ownership-model.md: distributions own translators, curators own points/speeches.
- TDD RED before GREEN. Full Go + Arch pytest + examples stay green after schema/shared changes.

### Claude's Discretion
- Whether to extract translators/common/ now or document duplication.
- live-build profile internals; preseed template structure.
- Exact representative-speech contents for dual-foundation proof.

### Deferred Ideas (OUT OF SCOPE)
- Ubuntu + other Debian-family variant profiles (post-v1.0).
- Full Debian ISO build on this host if devtmpfs-blocked (deferred-to-capable-host).
- translators/common/ extraction if not cleanly doable in-phase (document duplication, defer).
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID     | Description | Research Support |
|--------|-------------|------------------|
| DEB-01 | Debian translator (translators/debian/) wraps live-build/preseed and emits a bootable, fully-unattended Debian installer from a resolved speech, declaring its capabilities like Arch's | live-build config tree layout, preseed template pattern, capabilities.json gate |
| DEB-02 | DUAL-FOUNDATION PROOF — representative speech builds installers for BOTH Arch and Debian from the same resolved input | examples/dual-foundation/ structure, dual-foundation-check.sh script, equivalence checks |
| DEB-03 | Arch assumptions that leaked into schema/resolver/opinions are identified and fixed; schema/capability adjustments documented | sig_level enum is the primary concrete leak; install_phase enum is foundation-neutral; mkinitcpio tokens are Arch-specific capabilities (correctly isolated) |
| COMM-01 | Translator ownership model documented: distributions own translators, curators own points/speeches, community PRs welcome | docs/ownership-model.md structure and content |
</phase_requirements>

---

## Summary

Phase 4 proves invariant 1 ("translators own mechanics") is real by implementing a second foundation.
The Debian translator mirrors the Arch translator structurally — same Python-generator pattern, same
argv-stable shell entrypoint, same capability gate — but wraps Debian's live-build and preseed toolchain
instead of mkarchiso.

**Host environment confirmation:** Docker 29.5.3 is available and functional. `debootstrap` works correctly
inside a privileged Debian container (tested — base system installs successfully). However, `lb build`
fails at the chroot mount step on this Proxmox host — the same devtmpfs/loop-device restriction that
blocks mkarchiso also blocks live-build's chroot bind mounts. The gate for Phase 4 is
**profile-emission + structural validation** on this host; full ISO build is deferred-to-capable-host,
matching the Phase 2/3 policy. [VERIFIED: live Docker test on this host]

**Primary recommendation:** Use live-build's config tree (lb config/lb build) as the ISO assembly
mechanism, with the Debian translator's Python generator emitting the entire `config/` tree.
Use `preseed.cfg` in `config/includes.installer/` for fully-unattended d-i. Use a chroot hook
(`config/hooks/live/9000-debateos-apply.hook.chroot`) to run install-time opinion application
(packages via apt, file assets, services, sysctl, kernel params) at lb-build chroot time — this
is cleaner than preseed `late_command` for a large payload and avoids network dependency at install time.
Reserve `late_command` for the minimal unattended d-i boot flow only. First-run opinions follow the
same flag-file systemd oneshot pattern as Arch.

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Capability gate (required opinions) | translators/debian/ | — | Translator owns what it supports; resolver only sees capabilities as tokens |
| Custom apt repos → sources.list.d | translators/debian/ | — | apt-specific mechanic; schema carries foundation-neutral RepoDecl |
| Preseed (d-i automation) | translators/debian/ | — | Installer automation is translator-owned; schema knows nothing about d-i |
| First-run systemd units | translators/debian/ (shares pattern with arch) | translators/common/ candidate | Flag-file guard pattern is genuinely shared |
| Package name resolution | Opinion author | translators/debian/ (apt install) | Schema carries upstream names; translator invokes apt |
| Kernel param injection | translators/debian/ | — | grub/syslinux vs limine is translator-specific; schema has KernelParam {key,value} |
| Foundation field routing | cli/build | — | build.go reads speech Foundation and dispatches to correct translator dir |
| Dual-foundation equivalence gate | scripts/dual-foundation-check.sh | — | One script resolves once, calls both translators, runs structural checks |
| Ownership model docs | docs/ownership-model.md | translator READMEs | Human + machine readable; consumed by Forum (Phase 5) |

---

## Standard Stack

### Core
| Library / Tool | Version | Purpose | Why Standard |
|----------------|---------|---------|--------------|
| live-build | 1:20250505+deb13u1 (Debian stable) | lb config + lb build — generates Debian live ISO with embedded d-i | The official Debian live system toolchain; ships in Debian stable repos |
| debootstrap | 1.0.141 (Debian stable) | Bootstrap minimal Debian rootfs in Docker | Required by live-build; also usable standalone; ships in Debian stable |
| preseed (d-i) | bundled with Debian Installer | Fully-unattended d-i answers | The official Debian automated-install mechanism |
| PyYAML | ≥6.0.1 (already in repo) | Load variant profiles and opinion YAML | Already a project dependency; stdlib alternative requires extra parsing |
| pytest | 9.0.3 (already installed) | TDD test suite for Debian generator | Already used by Arch translator; matches installed version |

[VERIFIED: live Docker test confirms live-build 1:20250505+deb13u1 in debian:stable]
[VERIFIED: debootstrap 1.0.141 installs base system successfully in privileged Docker on this host]
[VERIFIED: pytest 9.0.3 already installed on this host via pip --break-system-packages]

### Supporting
| Library / Tool | Version | Purpose | When to Use |
|----------------|---------|---------|-------------|
| squashfs-tools | in debian:stable | Compress chroot into squashfs for ISO | Used by lb build internally; install in Docker image |
| xorriso | in debian:stable | ISO creation | Used by lb build internally |
| python3 | 3.12.3 (host) / in debian:stable | Generator runtime | Same as Arch translator |

[ASSUMED] squashfs-tools and xorriso version numbers in current debian:stable (not verified via apt-cache)

### Package Legitimacy Audit

> This phase installs no new Python or npm packages. All new dependencies are system tools
> (live-build, debootstrap) pulled from the official Debian apt repo inside Docker.
> No external package registry verification needed for this phase.

| Package | Registry | Source | Verdict | Disposition |
|---------|----------|--------|---------|-------------|
| live-build | Debian apt (official) | packages.debian.org | OK | Approved — official Debian toolchain |
| debootstrap | Debian apt (official) | packages.debian.org | OK | Approved — official Debian toolchain |
| PyYAML | PyPI (already in use) | pypi.org/project/PyYAML | OK | Already approved (Arch translator) |
| pytest | PyPI (already installed) | pypi.org/project/pytest | OK | Already approved (Arch translator) |

**Packages removed due to SLOP verdict:** none
**Packages flagged as suspicious SUS:** none

---

## Architecture Patterns

### System Architecture Diagram

```
examples/dual-foundation/
  speech.yaml (foundation: debian, foundation-neutral opinions)
  opinions/DF-*.yaml
        |
        v
go run ./cmd/resolve-json  -->  resolved.json  (Foundation: "debian")
        |
        +--------> translators/arch/translate  -->  out/arch-profile/
        |                                            (Arch gate, unchanged)
        v
translators/debian/translate  <resolved.json> --opinions ... --profile debian --out out/debian-profile/
        |
        v
translators/debian/generator.py
        |
   +---------+----------+
   |         |          |
capabilities manifest  variant
gate        assembly   (debian.yaml)
(DEB-03)   BuildManifest
        |
        v
translators/debian/profile.py  -->  out/debian-profile/
        |                             config/
        |                               archives/       (apt sources + keyrings)
        |                               hooks/live/     (chroot install hook)
        |                               includes.installer/
        |                                 preseed.cfg   (d-i answers)
        |                               package-lists/
        |                                 debateos.list.chroot_install
        |                             airootfs-equiv/
        |                               etc/systemd/user/  (first-run units)
        |                               usr/lib/debateos/  (scripts)
        |                             build-manifest.json  (runtime data)
        v
[lb build inside Docker] --> debian-live.iso  (deferred-to-capable-host)
        |
        v
[scripts/debian-validate-iso.sh]  (structural validation, host-runnable)
```

### Recommended Project Structure

```
translators/debian/
├── translate                  # argv-stable shell entrypoint (mirrors arch/translate)
├── generator.py               # generate() entrypoint, composes pipeline
├── capabilities.py            # load_capabilities / check_capabilities / CapabilityError
├── capabilities.json          # declared capability tokens (Debian-specific subset)
├── contract.py                # SHARED: load_resolved_speech / load_opinion_bodies
│                              #   (identical to arch/contract.py — candidate for common/)
├── manifest.py                # SHARED: BuildManifest + derive_source_date_epoch
│                              #   (identical to arch/manifest.py — candidate for common/)
├── firstrun.py                # SHARED: render_firstrun_unit (same flag-file pattern)
│                              #   (nearly identical to arch/firstrun.py)
├── variant.py                 # Debian-specific: load_variant_profile, apply_variant
│                              #   (sig_level → apt [trusted=yes]/[signed-by=] mapping)
├── profile.py                 # Debian-specific: emit_profile_tree → lb config/ tree
├── profiles/
│   └── debian.yaml            # Variant profile: stable, grub2, initramfs-tools
├── templates/
│   ├── preseed.cfg.tpl        # d-i preseed template (locale/disk/user/pkgsel/late_command)
│   ├── chroot-install.hook.tpl # chroot hook: apply packages, files, services, sysctl
│   └── firstrun.service.tpl   # Identical to arch template (shared candidate)
├── tests/
│   ├── __init__.py
│   ├── fixtures/
│   ├── test_capability_gate.py
│   ├── test_manifest.py
│   ├── test_profile.py
│   ├── test_variant.py
│   └── test_generator.py
├── __init__.py
├── pytest.ini
└── requirements-dev.txt

examples/
└── dual-foundation/
    ├── speech.yaml            # foundation: debian (neutral opinions only)
    ├── opinions/
    │   └── DF-*.yaml          # 4-6 foundation-neutral opinions
    └── README.md

scripts/
├── dual-foundation-check.sh   # resolve once → both translators → equivalence
├── debian-build-iso.sh        # lb build inside Docker (deferred-to-capable-host gate)
└── debian-validate-iso.sh     # structural validation of emitted config tree

docs/
└── ownership-model.md         # COMM-01: translator ownership model

translators/common/            # OPTIONAL: extract if refactor is clean in-phase
├── __init__.py
├── contract.py                # load_resolved_speech, load_opinion_bodies
├── manifest.py                # BuildManifest, derive_source_date_epoch
└── firstrun.py                # render_firstrun_unit, firstrun_unit_name, flag_file_path
```

### Pattern 1: Chroot Hook for Opinion Application (Preferred over late_command)

**What:** A `*.hook.chroot` script placed in `config/hooks/live/` runs inside the
live-build chroot at build time (not at install time). This is the recommended approach
for applying packages, file assets, services, and sysctl — it produces a pre-configured
chroot that is then squashed into the live system / installer filesystem.

**When to use:** All install-time opinion application (packages, file_assets, services,
sysctl_params, kernel_params, group_memberships). This avoids needing network access at
d-i time and is faster (packages already installed in the chroot).

**Example:** [CITED: live-team.pages.debian.net/live-manual]
```bash
#!/bin/bash
# config/hooks/live/9000-debateos-apply.hook.chroot
# Generated by translators/debian/profile.py from build-manifest.json

set -euo pipefail

# Install target packages (aggregated from applied opinions)
DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
  bash \
  curl \
  git
# ... (generated from manifest.target_packages)

# Deploy file assets
install -Dm644 /debateos-assets/etc/some-config /etc/some-config

# Enable services
systemctl enable some-service.service

# Apply sysctl params
install -Dm644 /dev/stdin /etc/sysctl.d/50-debateos.conf <<'EOF'
net.ipv4.tcp_fastopen = 3
EOF
```

### Pattern 2: Preseed for D-I Automation (Minimal)

**What:** `config/includes.installer/preseed.cfg` answers all debian-installer questions
automatically. The preseed drives d-i-level concerns only (partitioning, locale, user
account, base package selection). Opinion application belongs in the chroot hook (Pattern 1).

**When to use:** The preseed handles d-i automation; it does NOT need to install opinion
packages (those are already in the chroot from Pattern 1 for the live path). For the
installer path, `pkgsel/include` and `late_command` can apply a lightweight script.

[CITED: www.debian.org/releases/stable/example-preseed.txt]
```
# config/includes.installer/preseed.cfg — generated by translators/debian/profile.py
_preseed_V1

# Locale
d-i debian-installer/locale string en_US.UTF-8
d-i keyboard-configuration/xkb-keymap select us

# Network (DHCP auto-configure)
d-i netcfg/choose_interface select auto
d-i netcfg/get_hostname string debian

# Disk partitioning (auto — uses entire first disk)
d-i partman-auto/method string lvm
d-i partman-auto-lvm/guided_size string max
d-i partman-auto/choose_recipe select atomic
d-i partman/confirm_write_new_label boolean true
d-i partman/confirm boolean true
d-i partman/confirm_nooverwrite boolean true
d-i partman-lvm/confirm boolean true
d-i partman-lvm/confirm_nooverwrite boolean true

# User account (%%USERNAME%%, %%USER_FULLNAME%%, %%HASHED_PASSWORD%% sentinels)
d-i passwd/root-login boolean false
d-i passwd/user-fullname string %%USER_FULLNAME%%
d-i passwd/username string %%USERNAME%%
d-i passwd/user-password-crypted password %%HASHED_PASSWORD%%

# Package selection
d-i pkgsel/include string %%PKGSEL_PACKAGES%%
tasksel tasksel/first multiselect standard

# Late command: apply debateos-install.sh from the chroot (if available)
d-i preseed/late_command string \
  in-target /usr/lib/debateos/debateos-install.sh 2>&1 | tee /var/log/debateos-install.log || true

# Reboot after install
d-i finish-install/reboot_in_progress note
```

### Pattern 3: Apt Repo → sources.list.d Mapping (DEB-03 Core)

**What:** The schema `custom_repos` field uses `RepoDecl` with `sig_level` enum
(Required/RequiredDatabaseOptional/OptionalTrustAll/Never). The Debian translator maps
these to apt `sources.list.d` stanzas with `signed-by` or `trusted=yes` options.

[CITED: manpages.debian.org/testing/apt/sources.list.5.en.html]
[CITED: wiki.debian.org/DebianRepository/UseThirdParty]

**SigLevel enum → apt mapping:**

| Schema sig_level | apt sources.list option | Notes |
|-----------------|------------------------|-------|
| `Required` | `[signed-by=/etc/apt/trusted.gpg.d/NAME.asc]` | Keyring must be provided in `keyring` field |
| `RequiredDatabaseOptional` | `[signed-by=/etc/apt/trusted.gpg.d/NAME.asc]` | Same as Required in apt (no direct equiv) |
| `OptionalTrustAll` | `[trusted=yes]` | Trust warning in output (same as Arch) |
| `Never` | `[trusted=yes]` | Trust warning + LOUD WARNING comment; sig_level=Never means "accept unsigned" |

**Note:** `sig_level` enum values were designed for pacman. `OptionalTrustAll` and `Never`
both map to apt `[trusted=yes]` but with different warning levels. The schema enum is
**correctly shared** because it expresses the INTENT (how much to trust) — the translator
converts it to the distribution-specific mechanism. This is NOT an Arch leak; it is
intentional abstraction. [ASSUMED — requires confirmation with project owner on intent]

**Archive file pattern:**
```bash
# config/archives/debateos-REPONAME.list.chroot_install
deb [signed-by=/etc/apt/trusted.gpg.d/debateos-REPONAME.asc] https://repo.example.com/debian stable main
```
The keyring file from `RepoDecl.keyring` URL is fetched and written to
`config/archives/debateos-REPONAME.key.chroot` (live-build auto-installs it). [CITED: live-team.pages.debian.net/live-manual]

### Pattern 4: First-Run Systemd Units (Shared with Arch)

**What:** `execution_phase: first-run` opinions become systemd user oneshot units with
flag-file guards. The pattern is identical to Arch — same template, same flag-file dir,
same ConditionPathExists sentinel.

**Key difference from Arch:** Units land in `config/includes.chroot/etc/systemd/user/`
instead of `airootfs/etc/systemd/user/`. The live-build `includes.chroot` dir overlays
files directly into the chroot filesystem.

[ASSUMED] includes.chroot layout verified from search results, not live-build source inspection.

### Pattern 5: Foundation-Aware build.go Dispatch (CLI-01 refactor)

**What:** Currently `build.go` hardcodes `translateBin = "translators/arch/translate"` and
`profileDir = filepath.Join(outDir, "arch-profile")`. The refactor makes this data-driven
from `rs.Foundation` (the ResolvedSpeech Foundation field, which comes from speech.yaml).

[VERIFIED: resolver/resolve/explanation.go — ResolvedSpeech.Foundation is a string field]
[VERIFIED: cli/build/build.go line 141 — "arch-profile" is hardcoded]
[VERIFIED: cli/build/build.go line 146 — translateBin is hardcoded to "translators/arch/translate"]

**Refactor pattern:**
```go
// Foundation → translator config (data-driven, no if/else chains)
type foundationConfig struct {
    TranslateBin string // e.g. "translators/arch/translate"
    ProfileDir   string // subdirectory name under outDir, e.g. "arch-profile"
    DefaultProfile string // e.g. "vanilla-arch"
}

var foundationRegistry = map[string]foundationConfig{
    "arch":   {"translators/arch/translate",   "arch-profile",   "vanilla-arch"},
    "debian": {"translators/debian/translate", "debian-profile", "debian"},
    // Future: "ubuntu": {"translators/debian/translate", "debian-profile", "ubuntu"},
}

// rs.Foundation is set by the resolver from speech.yaml's foundation field.
fc, ok := foundationRegistry[rs.Foundation]
if !ok {
    return fmt.Errorf("build: unknown foundation %q — no translator registered", rs.Foundation)
}
```

The `--profile` flag default must also become foundation-aware, or the user overrides it.
Resolution: if `--profile` is empty string (not the current default "vanilla-arch"),
use the foundation's default profile; otherwise use the user's value.

### Anti-Patterns to Avoid

- **Using preseed late_command for large package installs:** late_command runs at d-i time,
  requiring network access at install time. Put packages in the chroot hook instead.
- **Hardcoding Debian package names in schema:** Schema carries upstream/common names;
  the Arch translator already uses `packages: [neovim]` and calls `pacman -S neovim`.
  Debian does the same with `apt-get install neovim`. Package name mapping (if the names
  differ) is a translator responsibility, not a schema field.
- **Per-variant code branches in generator.py:** ARCH-04 invariant applies equally to Debian.
  All variant differences must live in `profiles/debian.yaml`, `profiles/ubuntu.yaml` etc.
  The generator code must not branch on variant name.
- **Skipping the capability gate for nice-to-have opinions:** The Debian capability set
  will be smaller than Arch's (no pacman-specific capabilities, no mkinitcpio tokens).
  Required opinions that need missing capabilities must fail loudly at composition time.
- **Putting private assets in the lb config tree:** Same policy as Arch — private-injection.tar
  lives next to the ISO, not inside the chroot or the ISO itself.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Debian rootfs bootstrap | Custom debootstrap wrapper | live-build lb bootstrap stage | live-build handles debootstrap + apt config + hook ordering correctly |
| Preseed file parsing | Custom preseed reader | Use as a template output only — d-i parses it | Preseed is consumed by debian-installer at runtime; generator only writes the file |
| GPG keyring handling | Fetch + trust + verify pipeline | live-build config/archives/*.key.chroot pattern | live-build installs the keyring at the correct stage before package fetch |
| APT pinning for priorities | Custom priority resolver | `config/archives/REPO.pref.chroot` (APT pinning) | live-build picks up preference files automatically; maps from RepoDecl.priority |
| SquashFS compression | Custom mksquashfs invocation | lb build handles it internally | live-build calls mksquashfs with correct flags; SOURCE_DATE_EPOCH is inherited |
| ISO assembly | Custom xorriso call | lb binary stage | live-build handles hybrid ISO creation with correct bootloader integration |
| Flag-file guard pattern | Custom first-boot mechanism | Same template as Arch (ConditionPathExists) | Already proven in Arch; genuinely shared pattern |

**Key insight:** live-build is a high-level orchestrator. The translator generates the `config/` tree;
live-build handles all the low-level ISO assembly. Do not try to drive xorriso or mksquashfs directly.

---

## Arch-Leak Audit (DEB-03)

This is the most important research output for the planner. Each finding must become a plan task.

### Finding 1: `sig_level` enum — INTENTIONAL ABSTRACTION (not a leak)

**Schema field:** `repoDecl.sig_level` enum: Required/RequiredDatabaseOptional/OptionalTrustAll/Never

**Assessment:** These values were named after pacman's SigLevel but express a TRUST INTENT
that maps cleanly to both pacman and apt. The values mean: "require signatures" / "require
signatures on db only" / "trust even unsigned" / "never check signatures". The Debian translator
maps them to apt `[signed-by=...]` and `[trusted=yes]` as described in Pattern 3.

**Action:** No schema change needed. Document the mapping in the Debian translator's
variant.py `apply_variant()` function. Record as "correctly abstracted" in DEB-03 audit.

[ASSUMED — needs confirmation that RequiredDatabaseOptional maps cleanly to apt behavior]

### Finding 2: `install_phase` enum — FOUNDATION-NEUTRAL (not a leak)

**Schema field:** `install_phase` enum: preflight/packaging/config/login/post-install/first-run

**Assessment:** These phase names describe WHEN in the lifecycle an opinion effectuates.
They are pipeline-neutral: "preflight" = before anything, "packaging" = package install time,
"config" = configuration time, "login" = first user session, etc. The Debian translator maps
them the same way Arch does: packaging → chroot hook, first-run → systemd unit.

**Action:** No schema change needed. Document in DEB-03 audit as "foundation-neutral by design."

### Finding 3: `mkinitcpio`-specific capability tokens — CORRECTLY ISOLATED (translator-owned)

**Arch capabilities.json tokens:** `configure-mkinitcpio-hooks-and-modules`,
`write-mkinitcpio-config-drop-in`, `write-mkinitcpio-module-configuration`,
`write-mkinitcpio-module-list`, `write-mkinitcpio-module-list-configuration`,
`install-initramfs-hooks`

**Assessment:** These tokens appear in Arch opinions that configure mkinitcpio. The Debian
translator simply DOES NOT declare these capabilities. Any opinion that requires them will
fire a CapabilityError at composition time on the Debian foundation — which is correct behavior.
The dual-foundation proof speech must not include opinions with these tokens.

**Action:** No schema change. These tokens are correctly translator-owned. The Debian translator
declares equivalent capabilities for initramfs-tools (if needed): e.g., `configure-initramfs-tools-hooks`.
Record as "correctly isolated" in DEB-03 audit.

### Finding 4: `limine` bootloader tokens — CORRECTLY ISOLATED (Arch translator-owned)

**Arch capabilities.json tokens:** `manage-limine-bootloader-installation`,
`write-bootloader-entry-tool-drop-in`, `manage-efi-boot-entries`

**Assessment:** Limine is used by Omarchy on Arch. Debian uses GRUB2. These tokens are
translator-specific; Debian capabilities.json will declare `configure-grub2-bootloader`
or similar. Opinions that require limine capabilities will CapabilityError on Debian.

**Action:** No schema change. The `bootloader` schema field (`{name, timeout, snapshot}`)
is correctly abstract — it names the desired bootloader, and the translator maps it to
the distribution mechanism. Record as "correctly isolated."

### Finding 5: `build.go` hardcoded to Arch — GENUINE INFRASTRUCTURE LEAK

**Location:** `cli/build/build.go`
**Leak:** `translateBin = "translators/arch/translate"` (line ~146), `profileDir = "arch-profile"`
(line ~141), `--profile` default `"vanilla-arch"` (line ~81).

**Assessment:** The CLI layer is supposed to be foundation-agnostic (the user specifies
the foundation in their speech). Hardcoding Arch means `debateos build` fails for any
non-Arch speech.

**Action:** Refactor build.go to use `foundationRegistry` map (Pattern 5 above). This IS
a DEB-03 fix — it's infrastructure-level Arch assumption. Any schema/resolver change that
touches `ResolvedSpeech.Foundation` must not break this dispatch.

[VERIFIED: cli/build/build.go — lines 81, 141, 146-151 confirm the hardcoding]

### Finding 6: `custom_repos` keyring field — NAMING ASYMMETRY (minor, document only)

**Schema field:** `repoDecl.keyring` is a string (package name in Arch, URL/path in Debian).

**Assessment:** In Arch, `keyring` is a package name (e.g., `cachyos-keyring`). In Debian,
the equivalent is a `.asc` file downloaded from a URL. The current schema allows keyring
to be a string — both use cases fit if the translator interprets it as appropriate.

**Action:** Document in DEB-03 that the keyring field is "translator-interpreted" — Arch
translators treat it as a pacman package name, Debian translators treat it as a URL or
file path. No schema change required for v1.0. If it becomes confusing, add a `keyring_url`
field in a future schema version.

### Summary DEB-03 Findings

| Finding | Type | Action Required |
|---------|------|----------------|
| sig_level enum | Intentional abstraction | Document mapping; no schema change |
| install_phase enum | Foundation-neutral | No action needed |
| mkinitcpio capability tokens | Correctly isolated (Arch-owned) | Debian caps.json omits them |
| limine capability tokens | Correctly isolated (Arch-owned) | Debian caps.json omits them |
| build.go hardcoded to Arch | Genuine leak — infrastructure | Fix: foundationRegistry map |
| keyring field interpretation | Minor asymmetry | Document "translator-interpreted" |

---

## Host Environment: ISO Build Gate

**CONFIRMED via live Docker testing on this Proxmox host:**

| Test | Result | Notes |
|------|--------|-------|
| `docker run --privileged debian:stable` | Works | Full Docker available |
| `debootstrap` inside Docker | SUCCESS | Base system installs cleanly [VERIFIED] |
| Loop device availability | FAILS | `/dev/loop0` lost — `losetup` returns "No such file or directory" [VERIFIED] |
| `lb build` chroot stage | FAILS | `mount: permission denied` at chroot/test-dev-null [VERIFIED] |
| `lb config` (config tree generation) | SUCCESS | Config tree emitted cleanly [VERIFIED] |
| Profile emission (Python generator → config/) | Expected: SUCCESS | No loop/mount needed |

**Gate policy:** Same as Phases 2/3 (mkarchiso). Profile emission + structural validation
on this host. Full `lb build` ISO gated to capable host. Document in `scripts/debian-build-iso.sh`.

---

## Common Pitfalls

### Pitfall 1: chroot_install vs chroot suffix confusion

**What goes wrong:** Using `.list.chroot` instead of `.list.chroot_install` for the target package
list. Packages with `.list.chroot` are installed ONLY in the live environment (not the installed
system). Packages with `.list.chroot_install` go into both.

**Why it happens:** The suffix difference is subtle and not obvious.

**How to avoid:** The Debian translator should emit package lists with `.list.chroot_install`
for packages that must be in the installed system. Only live-env-specific packages use `.list.chroot`.

**Warning signs:** Installed system is missing expected packages from opinions.

### Pitfall 2: lb build requires loop devices (Proxmox restriction)

**What goes wrong:** `lb build` fails with `mount: permission denied` or `losetup` errors
when run on Proxmox VE hosts due to devtmpfs blocking privileged loop mounts.

**Why it happens:** Proxmox's security policy prevents `/dev/loop*` device creation even
in privileged containers. This is the same restriction that blocks mkarchiso.

**How to avoid:** Run `lb build` on a standard Linux host (KVM VM, bare metal, or
non-Proxmox CI). On this host, the gate is `--skip-iso` (profile-emission only).

**Warning signs:** `losetup` returns "device node is lost" or "No such file or directory."

### Pitfall 3: Preseed vs chroot hook confusion for package install

**What goes wrong:** Using preseed `late_command` with `in-target apt-get install` for all
opinion packages. This requires network access at d-i time and is slow (downloads at install time).

**Why it happens:** Preseed `late_command` is the obvious hook for "do something after install."

**How to avoid:** Put package installation in the chroot hook (Pattern 1) so packages are
already in the squashfs. Use `late_command` only for minimal d-i-level setup.

### Pitfall 4: Keyring timing (must install before custom repos)

**What goes wrong:** apt fails to fetch packages from custom repos because the signing key
isn't installed yet when the repo is first accessed.

**Why it happens:** In live-build, `config/archives/REPO.key.chroot` is installed by
live-build before packages are fetched — but the ordering can be fragile if the chroot hook
also tries to use the key.

**How to avoid:** Use live-build's `config/archives/` mechanism for keyrings (live-build
handles ordering). In the chroot hook, assume the keyring is already installed.
Mirror the Arch translator's `keyring_install_before_repos` logic in the chroot hook preamble.

### Pitfall 5: Dual-foundation speech using Arch-specific opinions

**What goes wrong:** The representative speech in `examples/dual-foundation/` includes opinions
with Arch-specific capability tokens (e.g., `configure-mkinitcpio-hooks-and-modules`). The
Debian translator CapabilityErrors on these, making the dual-foundation proof fail.

**Why it happens:** It is easy to accidentally include opinions that work on Arch but
require Arch-specific mechanics (mkinitcpio, limine, pacman AUR helpers).

**How to avoid:** The dual-foundation speech must include ONLY opinions whose
`translator_capabilities` tokens are declared by BOTH translators' `capabilities.json`.
Verify by running the capability gate for both translators before writing the proof script.

### Pitfall 6: `--profile` default breaks after build.go refactor

**What goes wrong:** After the foundation-aware refactor, `debateos build` with an Arch
speech uses the wrong default profile because the `--profile` flag still defaults to `"vanilla-arch"`.

**Why it happens:** The profile default is currently hardcoded in the flag definition.

**How to avoid:** Make `--profile ""` the default and resolve to `foundationRegistry[foundation].DefaultProfile`
when the flag is empty. This is backward-compatible — existing users who pass `--profile vanilla-arch`
explicitly are unaffected.

---

## Code Examples

### Verified: ResolvedSpeech.Foundation field

```go
// resolver/resolve/explanation.go — ResolvedSpeech (VERIFIED by codebase read)
type ResolvedSpeech struct {
    Schema       int                `json:"schema"`
    Foundation   string             `json:"foundation"`   // ← dispatches translator choice
    InstallOrder []resolver.OpinionID `json:"install_order,omitempty"`
    Applied      []resolver.OpinionID `json:"applied,omitempty"`
    // ...
}
```

### Verified: Arch translator generate() pipeline (mirror for Debian)

```python
# translators/arch/generator.py — generate() (VERIFIED by codebase read)
def generate(resolved_path, opinions_path, profile_name, out_dir):
    with open(resolved_path, "rb") as fh:
        resolved_bytes = fh.read()
    resolved = json.loads(resolved_bytes.decode("utf-8"))
    # default missing keys
    opinions_index = load_opinion_bodies(opinions_path)
    capabilities = load_capabilities()
    manifest = BuildManifest.from_resolved(resolved, opinions_index, capabilities, resolved_bytes)
    variant = load_variant_profile(profile_name)
    emit_profile_tree(out_dir=out_dir, manifest=manifest, variant=variant)
    return out_dir
```

The Debian generator mirrors this exactly, replacing `emit_profile_tree` with the Debian
profile emitter that writes the `config/` tree instead of the archiso tree.

### Verified: Arch profile.py _sanitize_dst (reuse in Debian)

```python
# translators/arch/profile.py — _sanitize_dst (VERIFIED by codebase read)
# This function is foundation-neutral and should be copied/shared.
sentinel = "/debateos-airootfs-root"
joined = os.path.normpath(os.path.join(sentinel, dst))
if not joined.startswith(sentinel + "/") and joined != sentinel:
    raise ValueError(f"file_asset dst '{dst}' traverses outside the target root")
```

### Verified: build.go hardcoded Arch references (to refactor)

```go
// cli/build/build.go — lines 81, 141, 146-151 (VERIFIED by codebase read)
profileFlag := fs.String("profile", "vanilla-arch", "translator profile name")  // ← hardcoded
// ...
profileDir := filepath.Join(outDir, "arch-profile")  // ← hardcoded
translateBin := "translators/arch/translate"          // ← hardcoded
```

### Pattern: Debian variant profile YAML (to implement)

```yaml
# translators/debian/profiles/debian.yaml
---
variant: debian
description: "Debian stable — standard apt repos, GRUB2 bootloader, initramfs-tools."
version_at_research: "2026-06-13"

repos: []  # Uses standard Debian mirrors; custom repos come from opinions' custom_repos

keyring_install_before_repos: []  # debian-archive-keyring is pre-installed

kernel:
  package: linux-image-amd64
  headers: linux-headers-amd64

defaults:
  initramfs: initramfs-tools  # Debian default (NOT mkinitcpio)
  bootloader: grub2           # Debian default (NOT limine)
  filesystem: ext4            # Debian default

# No pre-seeded opinions for vanilla Debian stable.
pre_seeded_opinions: []
```

### Pattern: Chroot hook template (core install-time application)

```bash
#!/bin/bash
# config/hooks/live/9000-debateos-apply.hook.chroot
# Generated by translators/debian/profile.py from build-manifest.json
# Source: build-manifest.json is placed in / by includes.chroot before this hook runs.
set -euo pipefail
export DEBIAN_FRONTEND=noninteractive

# Apply target packages (from manifest.target_packages)
apt-get install -y --no-install-recommends %%PACKAGES%%

# Remove packages (from manifest.remove_packages)
%%REMOVE_PACKAGES_STANZA%%

# Deploy file assets (from manifest.file_assets)
%%FILE_ASSET_STANZA%%

# Enable services (from manifest.system_services where enable=true)
%%SERVICE_ENABLE_STANZA%%

# Apply sysctl params (from manifest.sysctl_params)
%%SYSCTL_STANZA%%

# Apply kernel params (from manifest.kernel_params — GRUB_CMDLINE_LINUX)
%%KERNEL_PARAM_STANZA%%

# Add user to groups (from manifest.group_memberships)
%%GROUP_MEMBERSHIP_STANZA%%
```

The template uses `%%SENTINEL%%` replacement (same approach as Arch's installer.sh.tpl)
to avoid conflicts with shell variable syntax.

---

## Common/Shared Code Recommendation

**Recommendation: Extract `translators/common/` in Phase 4 if the refactor is clean.**

The following modules are byte-for-byte identical or near-identical between Arch and Debian:

| Module | Status | Notes |
|--------|--------|-------|
| `contract.py` | Identical | `load_resolved_speech`, `load_opinion_bodies` — no Arch-specific code |
| `manifest.py` | Identical | `BuildManifest`, `derive_source_date_epoch` — purely foundation-neutral |
| `firstrun.py` | Near-identical | Same flag-file pattern; only the template path differs |

**Extraction approach:**
```
translators/common/__init__.py
translators/common/contract.py   # from arch/contract.py unchanged
translators/common/manifest.py   # from arch/manifest.py unchanged
translators/common/firstrun.py   # parameterized template path
```

Both `translators/arch/` and `translators/debian/` then import:
```python
from translators.common.contract import load_resolved_speech, load_opinion_bodies
from translators.common.manifest import BuildManifest, derive_source_date_epoch
from translators.common.firstrun import render_firstrun_unit, firstrun_unit_name
```

**Risk:** Any test that imports from `translators.arch.contract` directly breaks and must
be updated. The Arch translator's `generate()` function imports `from contract import ...`
(bare name) — PYTHONPATH must include the translator dir, not just the common dir. Safest
approach: keep bare imports in arch/ working by symlinking or re-exporting from common/.

**If extraction is not clean:** Document the duplication, add `# SHARED: see translators/common` comments,
and plan a post-v1.0 refactor. Do NOT let extraction risk block the translator implementation.

---

## Validation Architecture

> `workflow.nyquist_validation` is not set in `.planning/config.json`; treated as enabled.

### Test Framework

| Property | Value |
|----------|-------|
| Framework | pytest 9.0.3 |
| Config file | `translators/debian/pytest.ini` (Wave 0 gap — create) |
| Quick run command | `pytest translators/debian/tests/ -x -q` |
| Full suite command | `pytest translators/debian/tests/ -v` |
| Regression suite | `pytest translators/arch/tests/ -q && go test ./... -count=1` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| DEB-01 | capability gate fires CapabilityError for required+unsupported | unit | `pytest translators/debian/tests/test_capability_gate.py -x` | Wave 0 gap |
| DEB-01 | nice-to-have+unsupported → visible drop (not error) | unit | `pytest translators/debian/tests/test_capability_gate.py::test_nice_to_have_dropped -x` | Wave 0 gap |
| DEB-01 | emit_profile_tree writes config/ tree (preseed.cfg, hook, package-lists) | unit | `pytest translators/debian/tests/test_profile.py -x` | Wave 0 gap |
| DEB-01 | variant profile debian.yaml loads and applies variant | unit | `pytest translators/debian/tests/test_variant.py -x` | Wave 0 gap |
| DEB-01 | BuildManifest from_resolved aggregates packages/services/sysctl | unit | `pytest translators/debian/tests/test_manifest.py -x` | Wave 0 gap |
| DEB-01 | generate() end-to-end emits valid config/ tree | integration | `pytest translators/debian/tests/test_generator.py -x` | Wave 0 gap |
| DEB-02 | dual-foundation-check.sh: resolve once → both translators → valid trees | integration | `bash scripts/dual-foundation-check.sh --skip-iso` | Wave 0 gap |
| DEB-02 | examples/dual-foundation speech resolves clean (Go test) | integration | `go test ./examples/ -run TestExampleDualFoundation -v` | Wave 0 gap |
| DEB-03 | build.go: foundation=debian dispatches translators/debian/translate | unit | `go test ./cli/build/ -run TestBuildFoundationDispatch -v` | Wave 0 gap |
| DEB-03 | build.go: unknown foundation returns error | unit | `go test ./cli/build/ -run TestBuildUnknownFoundation -v` | Wave 0 gap |
| DEB-03 | Arch pytest still green after schema/shared-code changes | regression | `pytest translators/arch/tests/ -q` | Exists |
| DEB-03 | Go test suite still green after build.go refactor | regression | `go test ./... -count=1` | Exists |
| COMM-01 | docs/ownership-model.md exists and contains required sections | smoke | manual / `test -f docs/ownership-model.md` | Wave 0 gap |

### Sampling Rate

- **Per task commit:** `pytest translators/debian/tests/ -x -q && pytest translators/arch/tests/ -q`
- **Per wave merge:** `pytest translators/debian/tests/ -v && go test ./... -count=1`
- **Phase gate:** Full suite green + `bash scripts/dual-foundation-check.sh --skip-iso` green

### Wave 0 Gaps

- [ ] `translators/debian/pytest.ini` — test configuration
- [ ] `translators/debian/tests/__init__.py` — package init
- [ ] `translators/debian/tests/test_capability_gate.py` — covers DEB-01 capability gate
- [ ] `translators/debian/tests/test_manifest.py` — covers BuildManifest assembly
- [ ] `translators/debian/tests/test_profile.py` — covers config/ tree emission
- [ ] `translators/debian/tests/test_variant.py` — covers debian.yaml loading
- [ ] `translators/debian/tests/test_generator.py` — covers generate() integration
- [ ] `translators/debian/tests/fixtures/` — resolved.json and opinion fixtures
- [ ] `examples/dual-foundation/` — speech.yaml + opinions/ (4-6 neutral opinions)
- [ ] `scripts/dual-foundation-check.sh` — dual-foundation proof gate
- [ ] `scripts/debian-build-iso.sh` — lb build inside Docker (deferred-to-capable-host)
- [ ] `scripts/debian-validate-iso.sh` — structural config/ tree validation (host-runnable)

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| preseed + late_command for all packages | chroot hooks for packages, late_command for d-i only | live-build >= 3.x | Faster installs; no network needed at d-i time |
| Global apt keyrings (apt-key add) | Per-repo signed-by= in sources.list | Debian bullseye (2021) | Per-repo trust; apt-key deprecated |
| Monolithic package list | .list.chroot_install / .list.chroot suffix split | live-build 3.0+ | Live env vs installed system package separation |
| CONFIG_SQUASH_GZIP | squashfs with SOURCE_DATE_EPOCH support | squashfs-tools 4.4+ | Deterministic builds possible |

**Deprecated/outdated:**
- `apt-key add`: deprecated since Debian bullseye; use `signed-by=` in sources.list.d
- `--packages` flag for `lb config`: not valid in current live-build (confirmed via test — use package-lists/)
- `/etc/apt/trusted.gpg`: deprecated in favor of `/etc/apt/trusted.gpg.d/*.gpg`

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | sig_level=Never maps to apt [trusted=yes]; RequiredDatabaseOptional maps same as Required | Arch-Leak Audit Finding 1 | May need a different apt option; low risk since both "trust" variants are explicitly warned |
| A2 | includes.chroot layout (config/includes.chroot/etc/systemd/user/) places files in installed system | Pattern 4 | First-run units land in wrong location; verify with lb config docs or test |
| A3 | config/includes.installer/preseed.cfg is the correct preseed location (not config/binary_debian-installer/preseed.cfg) | Pattern 2 | Preseed not picked up by d-i; low risk (both locations documented) |
| A4 | squashfs-tools and xorriso version numbers in debian:stable | Standard Stack | Minor; only affects version table accuracy |
| A5 | lb build fails due to devtmpfs/chroot mount restriction (not just loop devices) | Host Environment | CONFIRMED via live test — this is verified, not assumed |
| A6 | chroot hooks in config/hooks/live/*.hook.chroot run at lb-build time inside the chroot | Pattern 1 | Packages not installed at build time; verify hook naming convention |
| A7 | keyring field in RepoDecl should be treated as URL/path by Debian translator | Finding 6 | Keyring install fails; requires human decision on interpretation |

---

## Open Questions (RESOLVED)

> RESOLVED 2026-06-13 by the orchestrator under the autonomous-run mandate: (1) attempt translators/common/ extraction in Wave 1, fall back to documented duplication if it exceeds the 30-min/clean-import bar; (2) dual-foundation proof uses 5-6 foundation-neutral opinions across install-packages/deploy-config-file-tree/enable-systemd-service/write-sysctl-drop-in/add-user-to-group (Claude's discretion on exact contents); (3) build.go --profile default becomes "" and resolves from the foundationRegistry by Speech.Foundation (backward-compatible: arch* still resolves vanilla-arch); (4) apply all opinion effectuation in the lb chroot hook, minimize/avoid late_command. Originals retained below for traceability.

1. **Should translators/common/ be extracted in Phase 4 or deferred?**
   - What we know: manifest.py, contract.py, firstrun.py are byte-identical candidates
   - What's unclear: Whether the bare-import pattern in arch/ tests will break cleanly
   - Recommendation: Extract in Phase 4 Wave 1 (first task). If it takes > 30 min, document
     and defer. The planner should gate this in Wave 1 with a `checkpoint:human-verify` if
     extraction is attempted.

2. **What should the dual-foundation proof speech contain?**
   - What we know: Must use only opinions with capability tokens declared by BOTH translators.
     Foundation-neutral categories: package-install, file-assets, services, sysctl, groups.
     Must avoid: mkinitcpio, limine, pacman-AUR, initramfs-tools (Debian-specific)
   - What's unclear: Exact opinions to author (Claude's discretion per CONTEXT.md)
   - Recommendation: 4-6 opinions: base-cli-tools (packages), a config file (file_assets),
     a service (services), a sysctl param (sysctl_params), a group membership (group_memberships).
     These test the full manifest aggregation pipeline without requiring foundation-specific mechanics.

3. **build.go: should --profile default change to "" or stay "vanilla-arch"?**
   - What we know: Current default "vanilla-arch" breaks for non-Arch speeches
   - What's unclear: Whether existing users pass --profile explicitly
   - Recommendation: Change default to `""` and resolve from foundationRegistry. Add a
     `--profile` help string noting the foundation-specific default.

4. **Does `preseed/late_command` with `in-target` run in the installed system or the chroot?**
   - What we know: `in-target` runs commands in `/target` (the installed system during d-i)
   - What's unclear: Whether `/usr/lib/debateos/debateos-install.sh` is available at that
     point if it was only in the chroot (squashed into the live env, not necessarily in /target)
   - Recommendation: The late_command should be minimal. Opinion application belongs in the
     chroot hook. If late_command is needed, copy the install script into the installer
     filesystem via `config/includes.installer/`.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Docker | lb build + ISO gate | ✓ | 29.5.3 | — |
| python3 | Debian generator | ✓ | 3.12.3 | — |
| pytest | Debian tests | ✓ | 9.0.3 | — |
| Go | build.go refactor + Go tests | ✓ | (existing) | — |
| live-build | lb config (profile emission) | ✗ host, ✓ Docker | 1:20250505+deb13u1 in debian:stable | Profile emission only (no lb build) |
| debootstrap | lb bootstrap stage | ✗ host, ✓ Docker | 1.0.141 in debian:stable | n/a (in Docker) |
| squashfs-tools | lb build (ISO assembly) | ✗ host, ✓ Docker | in debian:stable | n/a (deferred to capable host) |
| loop devices | lb build chroot stage | ✗ (Proxmox blocks) | N/A | --skip-iso path (profile emission gate) |

**Missing dependencies with no fallback on this host:**
- Full `lb build` ISO construction (loop devices blocked by Proxmox devtmpfs) — same
  policy as mkarchiso. Gate = profile emission + structural validation.

**Missing dependencies with fallback:**
- live-build not on host → use Docker (`debian:stable`) for `lb config` output verification
  if needed. Profile emission by the Python generator does not require live-build on the host.

---

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | No auth in translator |
| V3 Session Management | No | No sessions in translator |
| V4 Access Control | No | No access control layer |
| V5 Input Validation | Yes | `_sanitize_dst` (T-02-08 pattern, reuse from Arch); preseed template sentinel replacement |
| V6 Cryptography | No (key handled by live-build) | Keyrings via apt signed-by — live-build handles GPG |

### Known Threat Patterns for Debian translator stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Path traversal via file_asset dst | Tampering | `_sanitize_dst()` reuse from arch/profile.py (T-02-08) |
| Shell injection via opinion data in hook template | Tampering | `%%SENTINEL%%` replacement (never str.format with raw opinion data); all opinion data via build-manifest.json |
| Unsigned package installation from custom repos | Tampering | sig_level trust warnings in output (T-02-02 pattern); [trusted=yes] emits LOUD WARNING comment |
| Private pane in ISO/chroot | Information Disclosure | Same policy as Arch: private-injection.tar next to ISO, never in config/ tree |
| preseed password in plain text | Information Disclosure | Use hashed password (crypt) in preseed; never plaintext |

---

## Sources

### Primary (HIGH confidence)
- Codebase: `translators/arch/` (all Python files) — pattern to mirror [VERIFIED]
- Codebase: `cli/build/build.go` — hardcoded Arch references confirmed [VERIFIED]
- Codebase: `resolver/resolve/explanation.go` — ResolvedSpeech.Foundation confirmed [VERIFIED]
- Codebase: `resolver/types.go` — RepoDecl, SigLevel, Opinion types confirmed [VERIFIED]
- Codebase: `schemas/opinion.schema.json` — install_phase enum, sig_level enum confirmed [VERIFIED]
- Live Docker test: debootstrap SUCCESS, lb config SUCCESS, lb build FAIL (loop devices), loop device FAIL [VERIFIED]

### Secondary (MEDIUM confidence)
- [live-team.pages.debian.net/live-manual — customizing-package-installation](https://live-team.pages.debian.net/live-manual/html/live-manual/customizing-package-installation.en.html) — config/archives, package-lists, hook structure
- [live-team.pages.debian.net/live-manual — customizing-installer](https://live-team.pages.debian.net/live-manual/html/live-manual/customizing-installer.en.html) — lb config --debian-installer live, preseed.cfg location
- [www.debian.org/releases/stable/example-preseed.txt](https://www.debian.org/releases/stable/example-preseed.txt) — preseed syntax, pkgsel/include, late_command, partitioning
- [wiki.debian.org/ReproducibleInstalls/LiveImages](https://wiki.debian.org/ReproducibleInstalls/LiveImages) — SOURCE_DATE_EPOCH + live-build
- [manpages.debian.org/testing/apt/sources.list.5.en.html](https://manpages.debian.org/testing/apt/sources.list.5.en.html) — signed-by, trusted=yes options

### Tertiary (LOW confidence)
- WebSearch results on chroot hook naming conventions — marked [ASSUMED] where used

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — verified via live Docker test on this exact host
- Architecture: HIGH — codebase verified; live-build config layout from official docs
- Arch-leak audit: HIGH (build.go hardcoding VERIFIED; schema analysis from source); MEDIUM (sig_level mapping intent ASSUMED)
- Pitfalls: MEDIUM — some from official docs, some from reasoning about the pattern
- Validation architecture: HIGH — mirrors Arch pattern exactly

**Research date:** 2026-06-13
**Valid until:** 2026-07-13 (30 days) — live-build stable channel; Debian stable is slow-moving
