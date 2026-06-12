# Omarchy Candidate Point Groupings

**Source inventory:** research/omarchy-opinion-inventory.md
**Pinned commit:** 9cf1852525a5f7de26d3162db9d61e2f5c1d5523
**Version:** 4.0.0.alpha
**Produced:** 2026-06-12

---

## Purpose

This document groups the 134 OM-NNN opinions from the Omarchy inventory into natural
candidate points — the units a curator would compose and a user would select as a bundle.
Grouping is evidence-driven (not quota-driven). Each point captures a coherent functional
or conceptual area; the OM-NNN membership list is the authoritative set.

**Coverage rule:** Every OM-NNN from the inventory (OM-001 through OM-134) appears in
exactly one point below.

---

## Point: Repository Bootstrap and Preflight

**Intent:** Establish the package manager configuration, apply any pending migrations, and
set the install-time execution context before any packages are installed.

**Member opinions:**
- OM-001 — Register the Omarchy custom package repository and import its signing key
- OM-002 — Disable automatic initramfs regeneration hooks during the install run
- OM-003 — Run pending system-migration scripts to bring an existing installation current
- OM-004 — Set a first-run marker and grant scoped passwordless sudo for first-run scripts
- OM-005 — Log the installation environment variables for debugging and reproducibility

---

## Point: Hyprland Desktop Stack

**Intent:** Install and configure the complete Wayland compositor session stack including
the compositor itself, session launcher, display portal, idle daemon, screen lock, color
temperature control, status bar, notification daemon, application launcher, background setter,
on-screen display, polkit agent, and the display manager (SDDM) with auto-login.

**Member opinions:**
- OM-006 — Install the Wayland compositor and session management stack (Hyprland, UWSM, portals, idle, lock)
- OM-015 — Install desktop shell components (waybar, quickshell, mako, omarchy-walker, swaybg, swayosd, sddm, polkit, GTK themes, input method, portals)
- OM-039 — Detect keyboard layout and propagate it to the compositor input configuration
- OM-040 — Configure XCompose for emoji input and personal identity shortcuts
- OM-046 — Initialize the Omarchy feature toggle state directory and toggle configuration files
- OM-049 — Set up the Walker application launcher autostart, systemd service, and post-upgrade hooks
- OM-056 — Initialize the Omarchy Hyprland feature toggle flags file

---

## Point: Terminal and Shell Toolchain

**Intent:** Install and configure the primary terminal emulator, multiplexer, and shell
enhancement tools that provide the daily command-line working environment.

**Member opinions:**
- OM-007 — Install terminal emulator (foot), multiplexer (tmux), and shell tools (starship, zoxide, bat, eza, fd, ripgrep, fzf, tldr, fastfetch)

---

## Point: Browser and Web Access

**Intent:** Install the primary web browser and its supporting libraries.

**Member opinions:**
- OM-008 — Install Chromium and supporting web content rendering and network libraries

---

## Point: Developer Toolchain

**Intent:** Install the core development tooling: language runtimes and version manager,
build tools, database libraries, the Neovim editor with curated plugin configuration,
version control tooling, and the mise work-directory setup.

**Member opinions:**
- OM-009 — Install developer tools: mise, language runtimes (Ruby, Rust, Lua, Python), build tools, database libraries
- OM-017 — Install development support packages (github-cli, jq, xmlstarlet, 1password, wl-clipboard, imagemagick, OCR, Plymouth, GObject Python)
- OM-019 — Install Neovim with LazyVim configuration and color scheme themes
- OM-041 — Set up default work directory, configure mise with Node.js, handle chroot vs online install modes
- OM-042 — Fix the powerprofilesctl script shebang to use system Python instead of mise-managed Python

---

## Point: AI Tooling

**Intent:** Install and configure all AI coding assistant and agent tooling, including native
packages, npm-global installs, agent skill links, and the Pi agent extension.

**Member opinions:**
- OM-010 — Install AI assistant native packages (claude-code and supporting libraries)
- OM-023 — Install AI development tools as npm global packages (Codex CLI, Gemini CLI, Copilot CLI, OpenCode, Playwright, Pi, ghui, Hunk)
- OM-054 — Deploy the Omarchy AI skill definition as a symlink into each AI agent's skills directory
- OM-055 — Install a custom Pi coding agent extension for Omarchy theme awareness

---

## Point: Containerization and Virtualization

**Intent:** Install container and virtualization tooling and configure the Docker daemon
with production-ready settings including log limits, DNS, network configuration, and
service activation.

**Member opinions:**
- OM-011 — Install containerization tools (Docker engine, Compose, Buildx, qemu-user-static-binfmt)
- OM-043 — Configure Docker daemon (log limits, bridge IP, DNS, socket activation, user group, boot dependency)

---

## Point: Media and Creative Applications

**Intent:** Install the full suite of media consumption and creation tools: video player,
audio tools, screen recorder, image viewer, image editor, PDF viewer, screenshot tools,
and screen annotation.

**Member opinions:**
- OM-012 — Install media tools (mpv, alsa-utils, pamixer, playerctl, wireplumber, wiremix, gpu-screen-recorder, obs-studio, imv, pinta, evince, grim, slurp, satty, xournalpp)

---

## Point: Productivity and Communication Applications

**Intent:** Install personal productivity and communication applications: note-taking,
secure messaging, office suite, markdown editor, audio streaming, and local file sharing.

**Member opinions:**
- OM-013 — Install productivity apps (Obsidian, Signal, LibreOffice, Spotify, Typora, Localsend)

---

## Point: System Utilities and Hardware Tools

**Intent:** Install system utilities covering Bluetooth management, disk tools, file manager
and extensions, display brightness control, system monitoring, file search, and network tools.

**Member opinions:**
- OM-014 — Install system utilities (bluetui, bolt, gnome-disk-utility, dosfstools, exfatprogs, nautilus, brightnessctl, asdcontrol, btop, inxi, impala, usage, dua-cli, plocate, iwd, inetutils, wireless-regdb, whois)
- OM-047 — Install Nautilus Python extensions for LocalSend and Transcode
- OM-053 — Add the current user to the input group for device access

---

## Point: Security and Network Services

**Intent:** Install and configure system security tools including firewall, printing support,
network service discovery, cryptographic keyring, and power management daemon.

**Member opinions:**
- OM-016 — Install security and administration tools (ufw, ufw-docker, cups, avahi, gnome-keyring, libsecret, power-profiles-daemon, watchdog/sensor libraries)

---

## Point: Fonts and Visual Assets

**Intent:** Install and register fonts used by the Omarchy desktop environment.

**Member opinions:**
- OM-018 — Deploy the Omarchy brand font (omarchy.ttf) and rebuild font cache

---

## Point: Web Applications and Launcher Entries

**Intent:** Install web applications as Chromium PWAs and configure TUI applications as
windowed launcher entries with custom icons and compositor layout hints.

**Member opinions:**
- OM-020 — Deploy application icon assets to the user icon directory
- OM-021 — Install curated web applications as Chromium PWAs with launcher entries and URL-scheme handlers
- OM-022 — Install TUI applications (dua, lazydocker) as windowed launcher entries with compositor layout hints

---

## Point: Application Configuration and Defaults

**Intent:** Deploy the baseline configuration file tree, configure MIME type associations,
set XDG user directories, configure GPG, write sudoers rules, initialize locate database,
and establish default user preferences.

**Member opinions:**
- OM-028 — Deploy the complete application configuration tree and shell startup file
- OM-030 — Deploy default ASCII branding assets to the user branding directory
- OM-031 — Set global git identity from install-time inputs
- OM-032 — Configure GPG daemon keyservers and restart the daemon
- OM-033 — Grant passwordless sudo access for timezone management commands
- OM-034 — Increase sudo and screen-lock password attempt limit to 10
- OM-035 — Set account lockout threshold to 10 failed attempts with 2-minute unlock timeout
- OM-044 — Configure system default applications for all major MIME types
- OM-045 — Configure XDG user directories and create GTK bookmarks
- OM-048 — Update the system file-locate database after all packaging

---

## Point: System Performance Tuning

**Intent:** Tune system kernel and resource parameters for a developer-oriented desktop:
inotify watchers, file descriptor limits, TCP MTU probing, and fast shutdown timeouts.

**Member opinions:**
- OM-036 — Enable TCP MTU probing to reduce SSH connection flakiness
- OM-037 — Increase inotify file watcher limit from 8192 to 524288
- OM-038 — Raise system-wide and per-user file descriptor limits
- OM-050 — Configure systemd for fast shutdown (reduce service stop timeout)

---

## Point: System Infrastructure and Hooks

**Intent:** Enable and configure supporting system services and infrastructure hooks that
underpin the desktop: kernel modules cleanup, Docker group membership, FUSE unmount hook,
brightness control sudo access, and miscellaneous infrastructure.

**Member opinions:**
- OM-051 — Install a system-sleep hook that unmounts FUSE filesystems before suspend
- OM-052 — Grant passwordless sudo access for the asdcontrol brightness-control tool
- OM-057 — Enable the kernel modules cleanup service

---

## Point: Networking and Connectivity Configuration

**Intent:** Enable and configure network management, wireless regulatory domain, Bluetooth,
printing services, and DNS resolution.

**Member opinions:**
- OM-061 — Enable the wireless networking daemon (iwd) and disable systemd-networkd-wait-online
- OM-062 — Detect wireless regulatory domain from timezone and write to wireless-regdom configuration
- OM-064 — Enable Bluetooth service, configure power-state persistence, and configure WirePlumber A2DP
- OM-065 — Enable CUPS, Avahi, and cups-browsed; configure mDNS and nsswitch for .local resolution

---

## Point: Hardware Peripheral Configuration

**Intent:** Configure input and peripheral hardware behavior applicable to many systems:
Apple-style function key mode, USB autosuspend policy, power button behavior, and ASUS
ROG laptop audio and microphone fixes.

**Member opinions:**
- OM-063 — Configure Apple-style keyboards to use F-keys as standard F-keys by default
- OM-066 — Disable USB autosuspend globally
- OM-067 — Configure the system to ignore the hardware power button press
- OM-081 — Fix audio volume control on ASUS ROG laptops (ALC285 codec WirePlumber profile)
- OM-082 — Fix internal microphone gain on ASUS ROG laptops with ALC285 codec

---

## Point: GPU and Display Drivers

**Intent:** Detect and install GPU drivers and video acceleration packages for NVIDIA,
AMD, Intel, and Apple Silicon GPUs, including Vulkan drivers.

**Member opinions:**
- OM-068 — Detect NVIDIA GPUs and install appropriate open or legacy DKMS drivers with Hyprland env vars
- OM-069 — Install Vulkan drivers matching detected GPU hardware (Intel/AMD/Apple)
- OM-070 — Install Intel hardware video acceleration drivers based on GPU generation

---

## Point: Intel Hardware Optimizations

**Intent:** Configure Intel-specific power management, thermal control, camera support,
kernel selection, and hardware workarounds for supported Intel CPU and chipset generations.

**Member opinions:**
- OM-071 — Install Intel Low Power Mode Daemon for supported hybrid Intel CPUs on battery-equipped systems
- OM-072 — Install thermald for Intel laptops with Sandy Bridge or newer CPUs
- OM-073 — Install MIPI camera support for Intel IPU7 hardware (OVTI08F4 ACPI HID)
- OM-074 — Install Panther Lake kernel (linux-ptl) on Dell XPS PTL systems; remove standard linux kernel
- OM-075 — Add fred=on kernel parameter for Intel Panther Lake systems
- OM-076 — Disable WiFi 7 EHT on Intel BE200/BE211 cards to work around broken EHT RX path
- OM-077 — Install Sound Open Firmware for non-XPS Intel Panther Lake systems

---

## Point: ASUS Hardware Support

**Intent:** Install ASUS-specific hardware control daemons and fix device-specific
hardware issues for ASUS ROG and ExpertBook/Zenbook variants.

**Member opinions:**
- OM-024 — Install ASUS ROG control daemon (asusctl) for fan curve and RGB control
- OM-078 — Fix display backlight on ASUS ExpertBook B9406 and Zenbook UX5406AA via DPCD AUX mode
- OM-079 — Disable Panel Replay on ASUS ExpertBook B9406 to fix display wake/freeze issue
- OM-080 — Fix ASUS ExpertBook B9406 touchpad motion event drop via libinput quirks override
- OM-083 — Fix ASUS ROG Flow Z13 detachable keyboard touchpad device classification

---

## Point: Framework Hardware Support

**Intent:** Install and configure Framework-specific hardware: QMK HID tool for Framework 16
RGB control, audio profile correction for Framework 13 AMD, and udev access rules.

**Member opinions:**
- OM-025 — Install QMK HID tool for Framework 16 keyboard RGB configuration
- OM-084 — Set HiFi audio card profile on Framework 13 AMD laptops
- OM-085 — Install udev rule for unprivileged access to Framework 16 keyboard HID device

---

## Point: Apple Hardware Support

**Intent:** Install and configure Apple-specific hardware support: SPI keyboard driver,
NVMe suspend fix, and full T2 chip support stack.

**Member opinions:**
- OM-086 — Install SPI keyboard driver and configure mkinitcpio for MacBook SPI models
- OM-087 — Fix NVMe wake-from-suspend on affected MacBook models via D3cold disable service
- OM-088 — Install full Apple T2 support (linux-t2, audio firmware, WiFi/BT firmware, fan control, Touch Bar)

---

## Point: Other Vendor Hardware Support

**Intent:** Install hardware support for Dell, Microsoft Surface, Lenovo, Broadcom WiFi,
Motorcomm Ethernet, Synaptics touchpad, and Tuxedo/Clevo keyboard backlight hardware.

**Member opinions:**
- OM-026 — Install haptic touchpad driver for Dell XPS laptops
- OM-027 — Install Marvell WiFi firmware for Microsoft Surface devices
- OM-089 — Fix bass speaker output on Lenovo Yoga Pro 7 14IAH10 via ALC287 pin quirk
- OM-090 — Install Broadcom WiFi DKMS drivers for BCM4360/BCM4331 chipsets
- OM-091 — Configure mkinitcpio for Surface-specific keyboard input modules
- OM-092 — Install DKMS driver for Motorcomm YT6801 Ethernet adapter (Slimbook Executive)
- OM-093 — Enable Synaptics InterTouch for touchpad devices where psmouse is not yet loaded
- OM-094 — Install Tuxedo/Clevo keyboard backlight DKMS drivers; blacklist conflicting legacy module

---

## Point: Battery and Power Management

**Intent:** Configure power profile switching and WiFi power-save mode based on AC adapter
state, restrict the locate database update to AC power, and manage battery monitoring.

**Member opinions:**
- OM-058 — Configure automatic power profile switching on AC adapter plug/unplug via udev rules
- OM-059 — Configure automatic WiFi power-save mode switching via udev rules on battery state change
- OM-060 — Configure plocate database update service to run only on AC power

---

## Point: Boot, Login, and Snapshot Configuration

**Intent:** Configure the complete boot and login stack: Plymouth boot splash, GNOME keyring
for auto-login, SDDM display manager with auto-login, hibernation support, and the Limine
bootloader with mkinitcpio hooks, UKI generation, btrfs snapshot menu, and EFI entry registration.

**Member opinions:**
- OM-095 — Install Omarchy Plymouth boot splash theme and activate it as default
- OM-096 — Create pre-configured passwordless GNOME keyring for auto-login setups
- OM-097 — Install Omarchy SDDM theme; configure Wayland auto-login; enable display manager service
- OM-098 — Set up hibernation: configure swap device, resume kernel parameters, mkinitcpio hook
- OM-099 — Configure Limine bootloader, mkinitcpio hooks, UKI generation, btrfs snapshot menu, EFI registration

---

## Point: Package Manager Finalization and Post-Install Repositories

**Intent:** Write the final package manager configuration for the installed system,
including the conditional addition of the arch-mact2 unsigned repository for Apple T2
hardware post-install updates.

**Member opinions:**
- OM-100 — Restore final package manager configuration; conditionally add arch-mact2 repo (SigLevel=Never) for T2 systems

---

## Point: First-Run System Setup

**Intent:** Perform one-time setup steps that require a live Wayland display session after
the user's first login: network connectivity check, WiFi prompt, system update notification,
firewall configuration, DNS resolver symlink, HiDPI GDK scale detection, battery monitor
setup, and post-first-run sudoers cleanup.

**Member opinions:**
- OM-101 — Check network connectivity on first login; send WiFi setup and system update prompts
- OM-103 — Configure firewall (default deny-inbound, allow outbound, LocalSend, Docker DNS); enable ufw
- OM-104 — Configure systemd-resolved stub DNS resolver by symlinking /etc/resolv.conf
- OM-105 — Detect monitor scale factor and write GDK_SCALE to compositor monitors configuration
- OM-106 — Enable battery monitor timer or set performance mode on AC-only systems
- OM-107 — Remove temporary first-run sudoers rule file to restore normal sudo behavior

---

## Point: First-Run Desktop Services

**Intent:** Start and enable desktop services that require an active user session on first
login: the Elephant application launcher, SwayOSD on-screen display server, and internal
monitor recovery service.

**Member opinions:**
- OM-108 — Enable and start the Elephant application launcher as a systemd user service
- OM-110 — Enable the omarchy-recover-internal-monitor user service
- OM-111 — Enable and start the SwayOSD on-screen display service

---

## Point: First-Run GTK and User Preferences

**Intent:** Apply GTK/GNOME theme settings, enable primary paste, and deliver onboarding
notifications — all of which require a running display session and GNOME settings backend.

**Member opinions:**
- OM-102 — Apply GTK dark mode and icon theme via gsettings; update icon cache
- OM-109 — Enable GTK primary paste (middle-click paste from selection clipboard) via gsettings
- OM-112 — Send welcome desktop notification with key Omarchy keyboard shortcut reference
- OM-113 — Send optional Voxtype voice dictation install notification (self-deletes after running)

---

## Point: Visual Themes

**Intent:** Provide the complete set of 21 selectable visual theme bundles — each bundle
delivers backgrounds, color palettes, per-application color configs, icon theme links, and
screen-lock images that a user can switch between at runtime without reinstalling the system.

**Member opinions:**
- OM-114 — Catppuccin dark (Mocha) theme bundle
- OM-115 — Catppuccin Latte (light) theme bundle
- OM-116 — Ethereal theme bundle
- OM-117 — Everforest theme bundle
- OM-118 — Flexoki Light theme bundle
- OM-119 — Gruvbox theme bundle
- OM-120 — Hackerman theme bundle
- OM-121 — Kanagawa theme bundle
- OM-122 — Last Horizon theme bundle
- OM-123 — Lumon theme bundle
- OM-124 — Matte Black theme bundle
- OM-125 — Miasma theme bundle
- OM-126 — Nord theme bundle
- OM-127 — Osaka Jade theme bundle
- OM-128 — Retro 82 theme bundle
- OM-129 — Ristretto theme bundle
- OM-130 — Rose Pine theme bundle
- OM-131 — Solitude theme bundle
- OM-132 — Tokyo Night theme bundle (default active theme)
- OM-133 — Vantablack theme bundle
- OM-134 — White (light) theme bundle

---

## Point: Desktop Theming Configuration

**Intent:** Configure the active theme system including the Chromium policy integration,
application-specific symlinks, and Plymouth boot splash branding — separate from the
theme asset bundles themselves.

**Member opinions:**
- OM-029 — Configure the desktop theme system: create user theme directory, activate Tokyo Night default, set up per-app symlinks and Chromium color scheme policy

---

## Summary Statistics

| Point Name | OM-NNN Members | Count |
|-----------|----------------|-------|
| Repository Bootstrap and Preflight | OM-001 to OM-005 | 5 |
| Hyprland Desktop Stack | OM-006, OM-015, OM-039, OM-040, OM-046, OM-049, OM-056 | 7 |
| Terminal and Shell Toolchain | OM-007 | 1 |
| Browser and Web Access | OM-008 | 1 |
| Developer Toolchain | OM-009, OM-017, OM-019, OM-041, OM-042 | 5 |
| AI Tooling | OM-010, OM-023, OM-054, OM-055 | 4 |
| Containerization and Virtualization | OM-011, OM-043 | 2 |
| Media and Creative Applications | OM-012 | 1 |
| Productivity and Communication | OM-013 | 1 |
| System Utilities and Hardware Tools | OM-014, OM-047, OM-053 | 3 |
| Security and Network Services | OM-016 | 1 |
| Fonts and Visual Assets | OM-018 | 1 |
| Web Applications and Launcher Entries | OM-020, OM-021, OM-022 | 3 |
| Application Configuration and Defaults | OM-028, OM-030 to OM-035, OM-044, OM-045, OM-048 | 10 |
| System Performance Tuning | OM-036, OM-037, OM-038, OM-050 | 4 |
| System Infrastructure and Hooks | OM-051, OM-052, OM-057 | 3 |
| Networking and Connectivity Configuration | OM-061, OM-062, OM-064, OM-065 | 4 |
| Hardware Peripheral Configuration | OM-063, OM-066, OM-067, OM-081, OM-082 | 5 |
| GPU and Display Drivers | OM-068, OM-069, OM-070 | 3 |
| Intel Hardware Optimizations | OM-071 to OM-077 | 7 |
| ASUS Hardware Support | OM-024, OM-078 to OM-080, OM-083 | 5 |
| Framework Hardware Support | OM-025, OM-084, OM-085 | 3 |
| Apple Hardware Support | OM-086, OM-087, OM-088 | 3 |
| Other Vendor Hardware Support | OM-026, OM-027, OM-089 to OM-094 | 8 |
| Battery and Power Management | OM-058, OM-059, OM-060 | 3 |
| Boot, Login, and Snapshot Configuration | OM-095 to OM-099 | 5 |
| Package Manager Finalization | OM-100 | 1 |
| First-Run System Setup | OM-101, OM-103 to OM-107 | 6 |
| First-Run Desktop Services | OM-108, OM-110, OM-111 | 3 |
| First-Run GTK and User Preferences | OM-102, OM-109, OM-112, OM-113 | 4 |
| Visual Themes | OM-114 to OM-134 | 21 |
| Desktop Theming Configuration | OM-029 | 1 |
| **Total** | **OM-001 to OM-134** | **134** |

---

## Unassigned

None — all 134 OM-NNN opinions from the inventory have been assigned to exactly one point above.
