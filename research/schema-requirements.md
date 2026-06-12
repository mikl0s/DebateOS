# Omarchy Schema Requirements Floor

**Source inventory:** research/omarchy-opinion-inventory.md
**Source points:** research/omarchy-points.md
**Variant evidence:** research/arch-variants-delta.md
**Conflict resolution floor:** docs/04-conflict-resolution.md
**Pinned commit:** 9cf1852525a5f7de26d3162db9d61e2f5c1d5523
**Produced:** 2026-06-12

---

## Purpose

This document defines the minimum expressive surface an Opinion/Point/Speech schema MUST
support, derived entirely from evidence in the Omarchy inventory and variant delta study.
Every requirement cites at least one OM-NNN or variant-evidence reference. Requirements are
OS-agnostic (invariant 1): they describe what the schema must EXPRESS, not translator
mechanics.

This is the **schema floor** for Phase 1 (D17). Phase 1 drafts the actual YAML schema; this
document bounds it from below with real evidence. The Phase 1 author must satisfy all SR-NNN
requirements herein.

**SR-NNN ID scheme:** `SR-NNN` — sequential three-digit integer (zero-padded), assigned
in logical grouping order. Scope: opinion-level and point-level fields only.

---

## Baseline Metadata Floor (from docs/04)

The following fields are established by the validated floor in `docs/04-conflict-resolution.md`.
The SR-NNN requirements below expand or refine each field based on Omarchy evidence.

| Field | docs/04 baseline | SR-NNN evidence extension |
|-------|-----------------|--------------------------|
| status | required / nice-to-have | SR-001 |
| dependencies | opinions/capabilities | SR-002 |
| conflicts | opinions/capabilities | SR-003 |
| hardware-conditions | requires: hw-type | SR-004, SR-005 |
| ordering-constraints | must-install-before/after | SR-006 |
| known-patches | patch opinion references | SR-003 |
| translator-capability | what translator must support | SR-007 |

---

## SR-001 — Required / Nice-to-Have Status Field

Every opinion MUST declare its status within its point as exactly one of `required` or
`nice-to-have`. This is the primary resolution discriminator in the docs/04 hierarchy.

**Evidence:** OM-003 (run pending migrations — required for install correctness), OM-006
(compositor stack — required for the Hyprland desktop point), OM-095 (Plymouth theme —
nice-to-have; the system boots without it), OM-113 (Voxtype notification — nice-to-have,
explicitly deferred optional).

**Invariant check:** No Arch-specific mechanics in this field. Status is independent of
foundation.

---

## SR-002 — Opinion Dependency References

Every opinion MUST express zero or more typed dependency references to other opinions by
stable OM-NNN-style ID. The schema MUST express: "this opinion requires that opinion to be
present and applied first."

**Evidence:** OM-023 depends on OM-009 and OM-041 (mise + Node.js runtime must be available
before npm-global installs). OM-097 depends on OM-096 (keyring must be configured before
the display manager sets up auto-login). OM-044 depends on OM-008, OM-012, OM-013, OM-015,
OM-021 (all target applications must be installed before MIME type associations are set).
OM-107 depends on OM-004 and all other first-run opinions (cleanup must be last).

---

## SR-003 — Conflict Declarations and Patch References

An opinion MUST be able to declare conflict with one or more other opinions or capabilities.
The schema MUST allow a known-patches list referencing patch opinions that can make the
conflict resolvable. This supports the docs/04 hard-conflict + patch-override model.

**Evidence (direct Omarchy):** OM-002 (mkinitcpio hooks disabled in preflight; this creates
a transient state that conflicts with any opinion attempting a kernel rebuild before OM-099
restores hooks). OM-099 depends on OM-002 being in the disabled state; the pair has an
implicit ordering-conflict if another opinion triggers mkinitcpio between them.

**Evidence (variant):** Garuda's `garuda-dracut-support` conflicts `mkinitcpio` — a
required-vs-required hard conflict with Omarchy's OM-002/OM-099 login opinions.
CachyOS `70-cachyos-settings.conf` sets `fs.file-max` (a global sysctl kernel parameter).
Omarchy OM-038 sets `DefaultLimitNOFILE` via systemd drop-ins in `system.conf.d/` and
`user.conf.d/` — this is a per-process RLIMIT_NOFILE, not a sysctl key. The two are distinct
mechanisms in different namespaces and do NOT collide on a shared key. A real sysctl-param
collision scenario exists between opinions writing the same sysctl key (e.g. two opinions both
writing `fs.inotify.max_user_watches`), which is exercised by EC-005 (synthesized) to confirm
that conflict declarations at the opinion level enable per-key detection.

---

## SR-004 — Single Hardware-Condition Predicate

An opinion MUST be able to declare a hardware condition: a named predicate (e.g.
`omarchy-hw-asus-rog`, `omarchy-battery-present`, `lspci nvidia`) that gates whether the
opinion is applied. When the condition evaluates false, the opinion is silently skipped; when
true, it is applied in normal pipeline order.

**Evidence:** OM-024 (ASUS ROG daemon, gate: `omarchy-hw-asus-rog`), OM-025 (Framework 16 QMK
HID, gate: `omarchy-hw-framework16`), OM-026 (Dell XPS haptic touchpad, gate:
`omarchy-hw-dell-xps-haptic-touchpad`), OM-027 (Surface Marvell WiFi, gate:
`omarchy-hw-surface`), OM-058 (power profile udev rules, gate: `omarchy-battery-present`),
OM-059 (WiFi power-save rules, gate: `omarchy-battery-present`), OM-106 (battery monitor or
performance mode, gate: `omarchy-battery-present`).

---

## SR-005 — Compound Hardware-Condition Predicates

**Schema Surprise #5.** The schema MUST express compound hardware conditions composed of
two or more named predicates joined with AND, OR, or NOT logic. A simple boolean flag is
insufficient — Omarchy uses 18 `omarchy-hw-*` helpers whose results are combined at runtime.

**Evidence:** OM-071 — condition is `omarchy-hw-intel AND omarchy-battery-present AND cpu
model in [151,154,170,172,183,186,189,191,204]` (three-predicate AND with a set membership
check). OM-072 — `omarchy-hw-intel AND omarchy-battery-present AND cpu model >= 42`.
OM-074 — `omarchy-hw-match "XPS" AND omarchy-hw-intel-ptl` (DMI string match AND chipset
check). OM-077 — `omarchy-hw-intel-ptl AND NOT omarchy-hw-match "XPS"` (requires logical
NOT). OM-082 — `omarchy-hw-asus-rog AND ALC285 codec present in /proc/asound` (vendor AND
audio codec). OM-083 — `omarchy-hw-asus-rog AND omarchy-hw-match "GZ302"` (vendor AND model).
OM-086 — DMI product_name matches one of MacBook[89],1|MacBook1[02],1|MacBookPro13,[123]|MacBookPro14,[123]
(multi-pattern OR). OM-088 — PCI ID 106b:1801 OR 106b:1802 (Apple T2 chip OR).

**Implication:** The schema must express at minimum: AND, OR, NOT, set-membership, and
string-match combinators over named hardware predicates. These are NOT translator mechanics;
they are the schema's condition language.

---

## SR-006 — Phase-Level Ordering Constraint (Load-Bearing)

**Schema Surprise #4.** Opinions MUST be able to declare an install-time pipeline phase
(preflight / packaging / config / login / post-install / first-run). Phase ordering is
load-bearing: an opinion in the packaging phase must not execute before all preflight
opinions complete; a login-phase opinion must not execute before all packaging and config
opinions complete. Opinions within a phase MUST additionally support relative ordering
constraints (before/after specific named opinions or capability names).

**Evidence:** OM-001 (preflight phase; must execute before all packaging opinions — enforced
by pipeline structure). OM-002 (preflight; hooks must be disabled before any kernel-triggering
package install). OM-023 (packaging by source location; but must execute AFTER config/OM-041 —
a cross-phase ordering exception: npm.sh is in the packaging/ directory but depends on
mise-work.sh which is in the config/ phase; the pipeline must support cross-phase before/after
constraints, not just within-phase ordering). OM-041 (config phase; before npm global installs).
OM-074 (config/hardware/intel; before OM-099 which rebuilds the UKI). OM-098 (login; before
OM-099, defers initramfs rebuild). OM-099 (login; last login step; must run after all kernel
parameter drop-ins).

**Cross-phase ordering exception (OM-023):** OM-023 is the primary evidence that phase-level
ordering is not sufficient alone — an opinion's declared phase may not match its effective
execution order when cross-phase opinion dependencies exist. The schema must support an
explicit `after: [OM-041]` constraint that overrides the phase sequence when necessary.

**Implication:** A flat "order: integer" field is insufficient. The schema must express both
a discrete phase enum and within-phase before/after relationships by opinion ID, with support
for cross-phase override constraints.

---

## SR-007 — Translator Capability Declaration

Every opinion MUST declare what translator capability it requires. This is an OS-agnostic
statement of abstract capability (e.g. "install named logical packages", "write a sysctl.d
drop-in", "enable a systemd service unit") that binds no foundation-specific mechanics.
Translators advertise their supported capabilities; the resolver uses this to verify that a
target foundation can express all opinions in a speech.

**Evidence:** OM-001 (translator-capability: add a signed external repository with Optional
TrustAll signature policy; import a GPG key by fingerprint). OM-003 (translator-capability:
maintain a persistent state store of applied scripts; enumerate pending migration scripts in
timestamp order). OM-068 (translator-capability: detect GPU model; conditionally install DKMS
driver packages; write modprobe options; write mkinitcpio config drop-in). OM-088 (translator-
capability: detect T2 chip by PCI ID; install multiple hardware support packages; enable
multiple services; write module-load configurations; write mkinitcpio modules; write bootloader
kernel parameters; write hardware daemon configuration — the largest translator capability
footprint in the inventory).

---

## SR-008 — File Asset Payload

**Schema Surprise #1.** An opinion MUST be able to carry or reference file asset payloads:
structured sets of files (images, config snippets, color palette files, Neovim theme files)
that are deployed to target paths as part of the opinion's effect. A package-name or
command-name reference is insufficient for this category.

**Evidence:** OM-114 through OM-134 (all 21 theme bundles) — each carries desktop backgrounds,
`btop` color theme, `colors.toml`, icon theme link, Neovim color scheme, lock screen image,
preview images, VS Code color theme JSON. These cannot be expressed as a package install; they
are file deployments from the opinion's source directory. OM-018 (font file omarchy.ttf
deployed to user font directory). OM-020 (application icon files deployed to icon directory).
OM-095 (Plymouth theme directory deployed to system Plymouth themes location).

**Implication:** The schema must express: a list of file assets (path within opinion bundle →
install path), or a reference to a bundle directory structure. OS-agnostic; the translator
handles the actual copy mechanics.

---

## SR-009 — Custom Package Repository Registration

**Schema Surprise #2.** An opinion MUST be able to register a custom package repository
as part of its effect. The schema MUST express at minimum: repository name, mirror URL or
mirrorlist path, GPG keyring package dependency, per-repository signature verification
level (from enumerated trust levels), and priority relative to other repositories (above
or below the foundation's standard repositories).

**Evidence:** OM-001 (registers the Omarchy custom repo with `Optional TrustAll` SigLevel,
imports GPG key by fingerprint, configures mirror channel). OM-100 (post-install, restores
final package manager config; conditionally appends arch-mact2 repo with `SigLevel = Never`
for Apple T2 hardware — the only `SigLevel=Never` repo in the inventory; annotated as
security-relevant with threat T-00-SIG).

**Variant evidence:** CachyOS registers `[cachyos]`, `[cachyos-v3]`, `[cachyos-v4]` repos
with `Required DatabaseOptional` SigLevel and priority above standard Arch repos (variant
delta study, CachyOS Custom Repos section). Garuda registers `[chaotic-aur]` and `[garuda]`
repos (variant delta study, Garuda Custom Repos section).

**Trust level enumeration required:** The schema MUST enumerate at minimum:
- `Required` — all packages must be signed; no exceptions
- `Required DatabaseOptional` — package signatures required, database signature optional
- `Optional TrustAll` — signature verification optional; any key accepted (OM-001)
- `Never` — no signature verification; explicitly unsigned (OM-100, arch-mact2)

This enumeration is a security-critical schema requirement (threat T-00-SIG2): unsigned
repos must be explicit in the schema floor so translators cannot silently ignore trust levels.

---

## SR-010 — Runtime Tool Install (npm-global and Equivalents)

**Schema Surprise #3.** An opinion MUST be able to express a runtime-tool install that uses
a package manager OTHER than the OS package manager (e.g. npm global install, pip install,
cargo install). These installs are categorically distinct from OS package installs and require
a separate schema field identifying the runtime tool manager and the tool names.

**Evidence:** OM-023 — installs 8 AI development tools via `npm install -g` (Codex CLI,
Gemini CLI, Copilot CLI, OpenCode, Playwright CLI, Pi agent, ghui, Hunk). This opinion's
category is explicitly `npm-global-install` — a distinct category added to the taxonomy
because package-install cannot express it. The translator must manage cross-package-manager
installs separately from the OS package manager.

**Implication:** The schema must express: `runtime_tool_manager` (e.g. `npm`, `pip`, `cargo`,
`gem`) and `tool_packages` (list of tool identifiers). Versions may also need to be
expressible (e.g. `@latest` vs pinned version). A single `packages:` field shared with OS
packages would conflate two distinct install mechanisms with different resolution semantics.

---

## SR-011 — Execution Phase: First-Run vs Install-Time

**Schema Surprise #7.** An opinion MUST declare its execution phase: `install-time` (the
default, executed during the main install pipeline) or `first-run` (executed only after the
user's first login into a live display session). First-run opinions require an active Wayland
session and cannot execute in a chroot or installer context.

**Evidence:** OM-101 through OM-113 (all 13 first-run opinions) carry `execution-phase:
first-run`. Examples: OM-103 (firewall configuration requires `ufw` commands that need a live
network stack and passwordless sudo via OM-004). OM-102 (GTK theme requires `gsettings` which
needs a running GNOME settings daemon — impossible in a chroot). OM-105 (GDK scale detection
queries the running Hyprland compositor for monitor scale — requires live compositor session).
OM-107 (sudoers cleanup is intentionally last in first-run; removes the rule created in OM-004).

**Implication:** Without this field, a translator cannot distinguish opinions that should run
at install time from those deferred to first login. The resolver must also enforce that
first-run opinions are not scheduled before the first-run execution context is active.

---

## SR-012 — Arbitrary Script Payload with Declared Capabilities

**Schema Surprise #6.** An opinion MUST be able to carry an arbitrary script payload that
is not expressible as a combination of package installs, file assets, service enables, or
configuration writes. The opinion MUST accompany any script payload with an explicit
`translator-capability` declaration enumerating what the script does at the abstract level.

**Evidence:** OM-054 — deploys symlinks to four different AI agent skill directories
(`~/.agents/skills`, `~/.claude/skills`, `~/.codex/skills`, `~/.pi/agent/skills`). This
cannot be expressed as a single package install or a config file write. The translator
capability is: "create symbolic links across multiple agent tool skill directories; handle
multiple AI tool ecosystems in parallel." OM-041 — configures mise with Node.js runtime,
handles both chroot (tarball install) and online (latest version) install modes; the branching
logic is script-level, not expressible as a simple package or config opinion.

OM-003 (migration runner), OM-004 (first-run marker + sudoers setup), OM-033 (timezone
sudoers), OM-042 (shebang fix), OM-047 (Nautilus extension deployment), OM-048 (updatedb),
OM-049 (Walker service setup), OM-051 (FUSE sleep hook), OM-055 (Pi agent extension) — all
carry script-level logic that requires explicit capability declaration to allow translator
capability-checking.

**Implication:** The `translator-capability` field is not optional metadata — it is the
mechanism by which the resolver verifies that a target foundation's translator can execute
the opinion. An opinion with an arbitrary script payload and no capability declaration is
unresolvable.

---

## SR-013 — Display Manager Configuration

Every opinion that configures the session display manager MUST be able to express: the
target display manager name, the session type (Wayland/X11), the compositor command or
session file name, the auto-login user, and any PAM configuration requirements.

**Evidence:** OM-097 (SDDM display manager — configures Wayland session with custom Hyprland
compositor command, auto-login for current user, PAM sddm-autologin configuration to disable
encrypted keyring creation).

---

## SR-014 — Bootloader Configuration Opinion

An opinion that configures the bootloader MUST be able to express: bootloader name, UKI
generation hooks, mkinitcpio hook lists, kernel parameter drop-in references, snapshot
menu integration, and EFI boot entry management — all as abstract capabilities, not as
specific tool invocations.

**Evidence:** OM-099 (Limine bootloader — sets mkinitcpio hooks for Plymouth, btrfs-overlayfs,
and encrypt; configures UKI generation; integrates limine-snapper-sync for btrfs snapshot
boot menu; triggers full UKI rebuild; registers EFI boot entry; removes archinstall entry).
OM-002 and OM-099 together demonstrate that bootloader configuration depends on hook state
managed by other opinions — the schema must express these inter-phase dependencies.

---

## SR-015 — Service Enable / Disable with Chroot Compatibility

An opinion that enables or disables a systemd service MUST declare the service unit name(s)
and whether the enable is deferred (safe in chroot, actual start deferred to first boot) or
immediate (requires a running init system). Some service enables appear in chroot contexts
where `systemctl start` cannot run but `systemctl enable` works.

**Evidence:** OM-057 (enable kernel-modules-hook service — comment in source notes chroot
context where `systemctl enable` runs but the service cannot be started). OM-064 (enable
Bluetooth — chroot-compatible enable). OM-097 (enable SDDM — must enable, not just start, in
installer context). OM-103 (enable ufw — first-run; requires live system). OM-108 (start
Elephant service — first-run; requires user session).

---

## SR-016 — Sysctl Parameter Drop-In Opinion

An opinion that tunes kernel parameters via sysctl MUST express the parameter key-value
pairs and the target drop-in file name. Multiple opinions may write to different drop-in
files (different parameter sets); the resolver must detect collisions when two opinions write
the same sysctl key.

**Evidence:** OM-036 (net.ipv4.tcp_mtu_probing=1), OM-037 (fs.inotify.max_user_watches).
Note: OM-038 sets `DefaultLimitNOFILE` via systemd `system.conf.d/` and `user.conf.d/`
drop-ins — this is a per-process RLIMIT_NOFILE managed by the systemd service manager, NOT
a sysctl kernel parameter. OM-038 belongs in a `systemd-limit` category, not `sysctl-param`.

**Variant evidence:** CachyOS `70-cachyos-settings.conf` sets `fs.file-max` (a true sysctl
kernel parameter). This does not collide with OM-038 (different mechanism, different namespace).
The schema must allow per-key conflict detection across sysctl-param opinions so that sysctl
collisions between a speech opinion and a variant's pre-seeded opinion — for example, two
opinions both writing `fs.inotify.max_user_watches` — are surfaced at composition time
(see EC-005 synthesized scenario).

---

## SR-017 — Kernel Boot Parameter Opinion

An opinion that adds or modifies kernel boot parameters MUST express the parameter key-value
pairs and the mechanism as an abstract capability ("append to kernel command line") without
specifying the bootloader tool. Multiple opinions may add parameters; the resolver must
merge them correctly and detect contradictions (e.g. two opinions setting the same parameter
to different values).

**Evidence:** OM-075 (fred=on for Intel PTL). OM-078 (xe.enable_dpcd_backlight=1 for ASUS
PTL display). OM-079 (xe.enable_panel_replay=0 for ASUS ExpertBook B9406). These three
opinions each add exactly one kernel parameter via bootloader entry-tool drop-in files and
must accumulate without collision in the final kernel command line.

---

## SR-018 — User and Group Membership Opinion

An opinion that adds a user to a system group MUST express the group name as an abstract
capability reference. The translator maps the abstract group name to the foundation-specific
group.

**Evidence:** OM-043 (add user to docker group for privileged container access). OM-053 (add
user to input group for /dev/input device access). OM-088 (add user to video group for Touch
Bar access).

---

## SR-019 — MIME Type and Application Association Opinion

An opinion that configures default application associations MUST express the MIME type
pattern(s) and the target application identifier (abstract application name, not binary
path) as a list of associations. A single opinion may register multiple MIME types.

**Evidence:** OM-044 — registers associations for directories, images (png/jpeg/gif/webp/bmp/
tiff), PDF, web URLs (http/https), all video formats, mailto links, text and source file
types, and rebuilds the desktop application database. The translator maps application names
to desktop file IDs for the target foundation.

---

## SR-020 — Theming System Opinion

An opinion in the theming category MUST express: a theme bundle directory reference (relative
path within the opinion's file assets), a list of per-application symlink targets, and a
mechanism for activating one theme as default at install time. Multiple theme opinions may
coexist in a speech; only one may be marked as the install-time default.

**Evidence:** OM-029 (theme system initialization — creates user theme directory, activates
Tokyo Night as default, creates btop and mako symlinks, sets Chromium system color scheme
policy). OM-114 through OM-134 (21 theme bundles each providing the assets that OM-029's
symlink mechanism points to). OM-132 is explicitly marked as the default-activated theme in
OM-029.

---

## SR-021 — Point Metadata: Name, Intent, and Membership

Every Point MUST carry: a stable human-readable name, a one-line OS-agnostic intent
statement, and an ordered list of member opinion IDs. The intent statement must be
comprehensible without reading the individual opinions.

**Evidence:** All 32 points in `research/omarchy-points.md` — each demonstrates the need for
a name (e.g. "Hyprland Desktop Stack"), a one-liner intent, and an explicit OM-NNN member
list. Representative examples: the Hyprland Desktop Stack point groups OM-006, OM-015,
OM-039, OM-040, OM-046, OM-049, OM-056 under a single coherent intent statement. The Visual
Themes point groups OM-114 through OM-134 under a runtime-selectable bundle intent. The
points document serves as the specification input for this requirement.

---

## SR-022 — Speech Metadata: Foundation Target and Point Composition

A Speech MUST declare: the target foundation (e.g. vanilla-arch, cachyos-v3, garuda) and an
ordered list of included points with their conflict-resolution decisions. The foundation
declaration gates which variant-profile pre-seeded opinions are active, enabling the resolver
to detect variant-vs-speech opinion conflicts.

**Variant evidence:** CachyOS pre-seeds `fs.file-max` via sysctl (no collision with OM-038,
which uses a different mechanism: systemd `DefaultLimitNOFILE` drop-ins). CachyOS pre-seeds
`linux-cachyos` kernel (conflicts Omarchy `linux` kernel opinion). Garuda pre-seeds
`mkinitcpio` conflicts (hard conflict with OM-002/OM-099 login phase opinions). These
conflicts are only detectable if the speech declares its foundation target and the resolver
can access the variant profile's pre-seeded opinion list.

---

## Coverage Matrix: Schema Surprises

| Schema Surprise | SR-NNN | Evidence |
|----------------|--------|---------|
| 1. File-asset payloads | SR-008 | OM-114..OM-134, OM-018, OM-020, OM-095 |
| 2. Custom package repository with URL, SigLevel, keyring dep | SR-009 | OM-001, OM-100, variant CachyOS/Garuda repos |
| 3. npm-global / runtime-tool installs distinct from OS pkg mgr | SR-010 | OM-023 |
| 4. Phase-level ordering as load-bearing constraint | SR-006 | OM-001..OM-005, OM-023, OM-074, OM-099 |
| 5. Compound hardware-condition predicates | SR-005 | OM-071, OM-072, OM-074, OM-077, OM-082, OM-083, OM-086, OM-088 |
| 6. Arbitrary script payloads with declared translator capabilities | SR-012 | OM-054, OM-041, OM-003, OM-004, OM-049 |
| 7. First-run vs install-time execution-phase field | SR-011 | OM-101..OM-113 |

All 7 confirmed schema surprises from the inventory are represented. All SR-NNN cite at least
one OM-NNN or variant-evidence reference. No SR intent text contains Arch-specific mechanics
(`pacman`, `AUR`, `mkarchiso`).
