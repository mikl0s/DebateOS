# Omarchy Opinion Inventory

**Source repository:** https://github.com/basecamp/omarchy
**Pinned commit:** 9cf1852525a5f7de26d3162db9d61e2f5c1d5523
**Version:** 4.0.0.alpha (from `version` file; no git tags exist in this repo)
**Research date:** 2026-06-12
**Researcher:** DebateOS Phase 0 automated analysis

---

## Reproducibility

All evidence in this file traces to the Omarchy source at commit `9cf1852525a5f7de26d3162db9d61e2f5c1d5523`.
Every `source:` citation is a path within that clone. To reproduce:

```bash
git clone https://github.com/basecamp/omarchy /tmp/omarchy
git -C /tmp/omarchy checkout 9cf1852525a5f7de26d3162db9d61e2f5c1d5523
git -C /tmp/omarchy rev-parse HEAD  # must equal 9cf1852525a5f7de26d3162db9d61e2f5c1d5523
cat /tmp/omarchy/version             # must read 4.0.0.alpha
```

---

## OM-NNN ID Scheme

- **Format:** `OM-NNN` — zero-padded 3-digit integer, e.g. `OM-001`
- **Range:** OM-001 through OM-200 (comfortable headroom over the ~120–150 distinct opinions found)
- **Assignment order:** sequential in install-pipeline walk order:
  1. preflight
  2. packaging/base (logical groups)
  3. packaging/specialized
  4. packaging/hardware-conditional
  5. config/general
  6. config/hardware (general)
  7. config/hardware/intel
  8. config/hardware/asus
  9. config/hardware/framework
  10. config/hardware/apple
  11. config/hardware/lenovo
  12. config/hardware/misc
  13. login
  14. post-install
  15. first-run
  16. themes

---

## Per-Entry Field Set

Every OM-NNN entry carries:

| Field | Required | Notes |
|-------|----------|-------|
| `category:` | Always | From the taxonomy defined in 00-RESEARCH.md |
| `intent:` | Always | OS-agnostic description of what is achieved; no Arch/pacman/AUR specifics |
| `source:` | Always | Path within the Omarchy clone (install/, config/, default/, themes/) |
| `dependencies:` | Always | Other OM-NNN IDs or "none" |
| `ordering:` | Always | Pipeline phase; "before" / "after" constraints relative to other opinions |
| `translator-capability:` | Always | What the Arch translator must be able to do |
| `condition:` | Hardware-conditional only | DMI / PCI ID / lspci predicate that gates this opinion |
| `execution-phase:` | First-run only | `first-run` — requires a live display session post-first-login |

**Invariant:** No `intent:` field may mention `pacman`, `AUR`, `mkarchiso`, or Arch-specific paths. Arch mechanics belong in `source:` and `translator-capability:` only.

---

## Install Pipeline Walk

### PHASE 1: Preflight

---

### OM-001

category: custom-repo
intent: Register the Omarchy package repository with the system package manager, import its GPG signing key (ID 40DFB630FF42BCFFB047046CF0134EE680CAC571), and set the preferred mirror channel (stable / rc / edge). This makes Omarchy-curated packages available before any other install step. Trust level: Optional TrustAll (SigLevel = Optional TrustAll — signature verification is optional, not enforced).
source: install/preflight/pacman.sh
dependencies: none
ordering: must execute before all packaging opinions; pipeline enforces this (preflight phase)
translator-capability: add a signed external package repository with Optional TrustAll signature policy; import a GPG key by fingerprint; configure a named mirror channel

---

### OM-002

category: service-enable
intent: Disable automatic initramfs regeneration hooks during the installation run to prevent multiple redundant rebuilds. Hooks are restored after the bootloader configuration phase completes.
source: install/preflight/disable-mkinitcpio.sh
dependencies: none
ordering: must execute before packaging (before any kernel or hook-triggering package installs); restored in login/limine-snapper.sh
translator-capability: temporarily disable package-manager post-install hooks; restore them after a designated later phase

---

### OM-003

category: arbitrary-script
intent: Run all pending system-migration scripts so the existing installation is current before new opinions are applied. Mark each migration as applied so it will not re-run on future install invocations.
source: install/preflight/migrations.sh
dependencies: OM-001 (package repo must be configured so migration scripts can install packages)
ordering: preflight phase; after OM-001, before packaging
translator-capability: maintain a persistent state store of applied scripts; enumerate and apply pending migration scripts in timestamp order

---

### OM-004

category: arbitrary-script
intent: Set a marker that tells deferred first-run scripts that they should execute after the user's first login session. Configure passwordless sudo access scoped to the exact set of commands those first-run scripts will need (systemctl, ufw, ufw-docker, gtk-update-icon-cache, resolv symlink, self-cleanup).
source: install/preflight/first-run-mode.sh
dependencies: none
ordering: preflight phase; before login
translator-capability: write a persistent state marker; configure scoped passwordless sudo rules for a defined set of commands

---

### OM-005

category: arbitrary-script
intent: Log the installation environment variables to aid debugging and provide a reproducible record of the install-time inputs (user name, email, repo, ref, paths).
source: install/preflight/show-env.sh
dependencies: none
ordering: preflight phase; informational only
translator-capability: log structured key-value data to the install transcript

---

### PHASE 2: Packaging — Base

---

### OM-006

category: package-install
intent: Install the compositor and Wayland session management stack: tiling compositor (Hyprland), session launcher (UWSM), display portal (hyprland/gtk), idle daemon, screen lock, color temperature tool.
source: install/packaging/base.sh + install/omarchy-base.packages
dependencies: OM-001 (repo must be configured)
ordering: packaging phase; before config (packages must exist before configuration)
translator-capability: install named logical packages; resolve package names to distro equivalents

---

### OM-007

category: package-install
intent: Install the primary terminal emulator (foot) and terminal-multiplexer (tmux) with supporting tools (shell completion, modern shell enhancements via starship, zoxide navigation, bat, eza, fd, ripgrep, fzf, tldr, tree-sitter-cli, fastfetch).
source: install/packaging/base.sh + install/omarchy-base.packages
dependencies: OM-001
ordering: packaging phase
translator-capability: install named logical packages

---

### OM-008

category: package-install
intent: Install the primary web browser (Chromium) and supporting libraries for web content rendering and network stack.
source: install/packaging/base.sh + install/omarchy-base.packages
dependencies: OM-001
ordering: packaging phase
translator-capability: install named logical packages

---

### OM-009

category: package-install
intent: Install developer tooling: version manager (mise), language runtimes (Ruby, Rust, Lua, Python), build tools (clang, llvm, luarocks, python-poetry-core), and database libraries (mariadb-libs, postgresql-libs).
source: install/packaging/base.sh + install/omarchy-base.packages
dependencies: OM-001
ordering: packaging phase
translator-capability: install named logical packages; handle multi-language runtime management tooling

---

### OM-010

category: package-install
intent: Install AI assistant tools that run as native applications or integration layers: claude-code (AI coding assistant), and supporting libraries for AI workflows.
source: install/packaging/base.sh + install/omarchy-base.packages
dependencies: OM-001
ordering: packaging phase; npm-based AI tools are separate opinions (OM-024)
translator-capability: install named logical packages

---

### OM-011

category: package-install
intent: Install containerization and virtualization tools: Docker engine, Docker Compose, Docker Buildx plugin, and cross-architecture emulation (qemu-user-static-binfmt).
source: install/packaging/base.sh + install/omarchy-base.packages
dependencies: OM-001
ordering: packaging phase; docker service configuration is a separate opinion (OM-058)
translator-capability: install named logical packages; handle privilege-requiring container tooling

---

### OM-012

category: package-install
intent: Install media consumption and creation tools: video player (mpv with mpris), audio tools (alsa-utils, pamixer, playerctl, wireplumber, wiremix), video recording (gpu-screen-recorder, obs-studio), image viewer (imv), image editor (pinta), PDF viewer (evince), document viewer, screenshot tools (grim, slurp, satty), screen annotation (xournalpp).
source: install/packaging/base.sh + install/omarchy-base.packages
dependencies: OM-001
ordering: packaging phase
translator-capability: install named logical packages

---

### OM-013

category: package-install
intent: Install productivity and communication applications: Obsidian (note-taking), Signal (messaging), LibreOffice (office suite), Spotify (audio streaming), Typora (markdown editor), Localsend (local file sharing).
source: install/packaging/base.sh + install/omarchy-base.packages
dependencies: OM-001
ordering: packaging phase
translator-capability: install named logical packages

---

### OM-014

category: package-install
intent: Install system utilities and hardware support: Bluetooth manager (bluetui, bolt), disk utilities (gnome-disk-utility, dosfstools, exfatprogs), file manager (nautilus, nautilus-python, sushi), display control (brightnessctl, asdcontrol), system info and monitoring (btop, inxi, impala, usage, dua-cli), file search (plocate), network tools (iwd, inetutils, wireless-regdb, whois).
source: install/packaging/base.sh + install/omarchy-base.packages
dependencies: OM-001
ordering: packaging phase
translator-capability: install named logical packages

---

### OM-015

category: package-install
intent: Install graphical desktop shell components: status bar (waybar, quickshell), notification daemon (mako), application launcher (omarchy-walker), background setter (swaybg), OSD overlay (swayosd), display manager (sddm), polkit agent (polkit-gnome), GTK themes (gnome-themes-extra, kvantum-qt5), file manager extensions (gvfs-mtp, gvfs-nfs, gvfs-smb), input method framework (fcitx5, fcitx5-gtk, fcitx5-qt), screen color picker (hyprpicker), share picker (hyprland-preview-share-picker), desktop portal (xdg-desktop-portal-gtk, xdg-desktop-portal-hyprland), terminal executor (xdg-terminal-exec).
source: install/packaging/base.sh + install/omarchy-base.packages
dependencies: OM-001
ordering: packaging phase
translator-capability: install named logical packages

---

### OM-016

category: package-install
intent: Install system security and administration tools: firewall (ufw, ufw-docker), printer support (cups, cups-browsed, cups-filters, cups-pdf), network service discovery (avahi, nss-mdns), cryptography (gnome-keyring, libsecret, gnome-calculator), power management (power-profiles-daemon), hardware watchdog and sensors libraries.
source: install/packaging/base.sh + install/omarchy-base.packages
dependencies: OM-001
ordering: packaging phase
translator-capability: install named logical packages

---

### OM-017

category: package-install
intent: Install development support packages: version control (github-cli), JSON tool (jq), XML tool (xmlstarlet), web archive tool (woff2-font-awesome), session tools (1password-beta, 1password-cli), screen recording (gpu-screen-recorder), Wayland clipboard (wl-clipboard), image processing (imagemagick, ffmpegthumbnailer), OCR (tesseract, tesseract-data-eng), Plymouth boot splash, GObject Python bindings, YAML library.
source: install/packaging/base.sh + install/omarchy-base.packages
dependencies: OM-001
ordering: packaging phase
translator-capability: install named logical packages

---

### PHASE 2: Packaging — Specialized

---

### OM-018

category: font-install
intent: Deploy the custom Omarchy brand font (omarchy.ttf) to the user's local font directory and rebuild the font cache so the font is immediately available to desktop applications.
source: install/packaging/fonts.sh
dependencies: OM-001
ordering: packaging phase; after base packaging
translator-capability: deploy a font file to the user's local font directory; trigger font cache rebuild

---

### OM-019

category: package-install
intent: Install the Neovim editor with the Omarchy-curated LazyVim configuration and color scheme themes via the omarchy-nvim-setup script.
source: install/packaging/nvim.sh
dependencies: OM-001, OM-007 (nvim already in base packages; this adds LazyVim setup)
ordering: packaging phase; after base packages
translator-capability: run a post-install setup script that configures an editor's plugin ecosystem

---

### OM-020

category: arbitrary-script
intent: Deploy application icon assets to the user's local application icons directory so that custom launcher entries have the correct visual icons.
source: install/packaging/icons.sh
dependencies: OM-015 (desktop infrastructure must exist)
ordering: packaging phase; before mimetypes and webapp configs
translator-capability: copy icon files to a well-known application icon directory

---

### OM-021

category: arbitrary-script
intent: Install a curated set of web applications as Chromium Progressive Web Apps (PWAs) via a wrapper that creates .desktop launcher entries with custom icons and optional URL-scheme handler registrations. Applications: HEY (email, mailto handler), Basecamp, WhatsApp, Google Photos, Google Contacts, Google Messages, Google Maps, ChatGPT, YouTube, GitHub, X, Figma, Discord, Zoom (zoom scheme handler), Fizzy.
source: install/packaging/webapps.sh
dependencies: OM-008 (Chromium), OM-020 (icons)
ordering: packaging phase
translator-capability: create .desktop launcher entries for web apps; register URL scheme handlers

---

### OM-022

category: arbitrary-script
intent: Install terminal UI applications as windowed launcher entries with custom icons and layout mode (float or tile): Disk Usage (dua), Docker (lazydocker).
source: install/packaging/tuis.sh
dependencies: OM-020 (icons), OM-015 (launcher infrastructure)
ordering: packaging phase
translator-capability: create .desktop launcher entries for TUI apps with compositor layout hints

---

### OM-023

category: npm-global-install
intent: Install AI coding assistant and development tools via the Node.js package manager as global command-line tools: OpenAI Codex CLI, Google Gemini CLI, GitHub Copilot CLI, OpenCode AI, Playwright browser automation CLI, Pi coding agent, GitHub UI assistant (ghui), Hunk diff reviewer.
source: install/packaging/npm.sh
dependencies: OM-009 (mise + Node.js runtime must be available)
ordering: packaging phase; after mise-work (OM-054) which installs Node via mise
translator-capability: install npm global packages; resolve tool names to npm package identifiers; handle cross-package-manager installs separate from the OS package manager

---

### PHASE 2: Packaging — Hardware-Conditional

---

### OM-024

category: hardware-conditional
intent: Install ASUS ROG laptop control daemon (asusctl) for fan curve management, performance mode switching, and keyboard RGB control.
source: install/packaging/asus-rog.sh
dependencies: OM-001
ordering: packaging phase
condition: omarchy-hw-asus-rog helper returns true (DMI vendor matches ASUS ROG chassis)
translator-capability: conditionally install packages based on hardware detection; invoke DMI-based hardware detection helpers

---

### OM-025

category: hardware-conditional
intent: Install QMK HID tool for Framework 16 keyboard RGB and key configuration.
source: install/packaging/framework16.sh
dependencies: OM-001
ordering: packaging phase
condition: omarchy-hw-framework16 helper returns true (DMI matches Framework 16)
translator-capability: conditionally install packages based on hardware detection

---

### OM-026

category: hardware-conditional
intent: Install haptic touchpad driver for Dell XPS laptops with haptic touchpad hardware.
source: install/packaging/dell-xps-touchpad-haptics.sh
dependencies: OM-001
ordering: packaging phase
condition: omarchy-hw-dell-xps-haptic-touchpad helper returns true (DMI matches Dell XPS haptic touchpad variant)
translator-capability: conditionally install packages based on hardware detection

---

### OM-027

category: hardware-conditional
intent: Install Marvell WiFi firmware for Microsoft Surface devices.
source: install/packaging/surface.sh
dependencies: OM-001
ordering: packaging phase
condition: omarchy-hw-surface helper returns true (DMI matches Microsoft Surface)
translator-capability: conditionally install firmware packages based on hardware detection

---

### PHASE 3: Config — General

---

### OM-028

category: config-dotfile
intent: Deploy the complete set of application configuration files to the user's config directory, establishing the baseline Omarchy configuration for all desktop applications. Also deploy the Omarchy default shell startup file.
source: install/config/config.sh
dependencies: OM-006 through OM-017 (packages must be installed before config files are deployed)
ordering: config phase; first config step; before theme.sh which sets the active theme
translator-capability: recursively copy a config file tree to the user's config directory; deploy a shell startup file

---

### OM-029

category: theming
intent: Configure the desktop theme system: create the user theme directory, set up Chromium policy directory for theme integration, activate the default theme (Tokyo Night), create symbolic links for per-application theme assets (btop, mako), set Chromium to follow system color scheme.
source: install/config/theme.sh
dependencies: OM-028 (config files must be deployed), themes directory (OM-162 through OM-182)
ordering: config phase; after config.sh
translator-capability: manage a multi-application theme system; write browser policy files; create config symlinks

---

### OM-030

category: config-dotfile
intent: Allow users to customize the ASCII-art branding shown in the fastfetch system info display and the screensaver. Copies default branding assets to a user-owned branding directory.
source: install/config/branding.sh
dependencies: OM-028
ordering: config phase
translator-capability: copy branding asset files to a user config directory

---

### OM-031

category: config-dotfile
intent: Set the user's global git identity (name and email) from the values provided at install time, so git commits are correctly attributed from first use.
source: install/config/git.sh
dependencies: OM-009 (git is included via base packages)
ordering: config phase
translator-capability: write global git configuration settings

---

### OM-032

category: config-dotfile
intent: Configure the GPG daemon to use multiple keyservers for improved reliability. Deploy the configuration to the system-wide GnuPG config directory and restart the daemon.
source: install/config/gpg.sh
dependencies: none (gpg is a base system tool)
ordering: config phase
translator-capability: deploy a GPG daemon configuration file; restart the GPG daemon

---

### OM-033

category: arbitrary-script
intent: Grant passwordless sudo access to the timezone management commands (tzupdate and timedatectl) for all members of the wheel group, so users can update their timezone without a password prompt.
source: install/config/timezones.sh
dependencies: none
ordering: config phase
translator-capability: write a sudoers drop-in rule granting passwordless access to specific commands for a user group

---

### OM-034

category: config-dotfile
intent: Increase the maximum number of password entry attempts in the sudo and screen-lock tools from the default (3) to 10, reducing accidental lockouts from mistyped passwords.
source: install/config/increase-sudo-tries.sh
dependencies: none
ordering: config phase
translator-capability: write a sudoers drop-in to set passwd_tries; modify the PAM faillock configuration

---

### OM-035

category: config-dotfile
intent: Set the account lockout threshold to 10 failed attempts with a 2-minute unlock timeout (instead of the default shorter lockout), and ensure the auto-login session resets the fail counter on success.
source: install/config/increase-lockout-limit.sh
dependencies: none
ordering: config phase
translator-capability: modify PAM system-auth and sddm-autologin configuration to set faillock parameters

---

### OM-036

category: sysctl-param
intent: Enable TCP MTU probing to reduce SSH connection flakiness caused by path MTU issues.
source: install/config/ssh-flakiness.sh
dependencies: none
ordering: config phase
translator-capability: write a sysctl.d drop-in with a net.ipv4.tcp_mtu_probing=1 parameter

---

### OM-037

category: sysctl-param
intent: Increase the inotify file watcher limit from the default 8192 to 524288, sufficient for large development projects using tools like file watchers and bundlers.
source: install/config/increase-file-watchers.sh
dependencies: none
ordering: config phase
translator-capability: write a sysctl.d drop-in with fs.inotify.max_user_watches; apply sysctl settings immediately

---

### OM-038

category: sysctl-param
intent: Raise the system-wide and per-user file descriptor limit from the default 1024 to a soft limit of 65536 (hard 524288), so development tools, container daemons, and databases operate without running out of file handles.
source: install/config/increase-fd-limit.sh
dependencies: none
ordering: config phase
translator-capability: write systemd system.conf.d and user.conf.d drop-in files to set DefaultLimitNOFILE

---

### OM-039

category: config-dotfile
intent: Detect the keyboard layout and variant configured in the base system and propagate those settings into the Hyprland compositor input configuration, so keyboard layout carries over from the Arch install step without requiring manual re-entry.
source: install/config/detect-keyboard-layout.sh
dependencies: OM-028 (hyprland config files must be deployed)
ordering: config phase; after config.sh
translator-capability: read system keyboard configuration; write compositor-specific keyboard layout settings

---

### OM-040

category: config-dotfile
intent: Configure XCompose for fast emoji input and personal identity shortcuts (name and email triggered by key sequences). Uses CapsLock as the compose key.
source: install/config/xcompose.sh
dependencies: none
ordering: config phase
translator-capability: write an XCompose configuration file with include directives and user-identity shortcuts

---

### OM-041

category: arbitrary-script
intent: Set up a default work directory structure with a mise tool configuration that adds ./bin to the PATH for all projects under ~/Work. Install Node.js via mise (from a bundled tarball in chroot installs, or the latest version in online installs) and configure it globally.
source: install/config/mise-work.sh
dependencies: OM-009 (mise must be installed)
ordering: config phase; before npm global installs (OM-023)
translator-capability: configure a version manager; set global runtime versions; handle both chroot and online install modes

---

### OM-042

category: arbitrary-script
intent: Fix the powerprofilesctl script to use the system Python 3 interpreter instead of the mise-managed Python, preventing version conflicts when mise activates a different Python.
source: install/config/fix-powerprofilesctl-shebang.sh
dependencies: OM-041 (mise must be configured before this fix is needed)
ordering: config phase
translator-capability: rewrite a shebang line in a system script file

---

### OM-043

category: service-enable
intent: Configure the Docker daemon with log size limits, a fixed bridge IP, and DNS pointing to the host's systemd-resolved instance. Expose the systemd-resolved stub to the Docker bridge network. Enable Docker socket activation (on-demand startup). Add the current user to the docker group for privileged access. Prevent Docker from blocking system boot by removing its network-online dependency.
source: install/config/docker.sh
dependencies: OM-011 (Docker packages must be installed)
ordering: config phase
translator-capability: write a Docker daemon configuration JSON; configure systemd-resolved drop-in for DNS; enable a systemd socket unit; add user to a system group; write a systemd service drop-in

---

### OM-044

category: mime-type
intent: Configure the system default applications for all major MIME types: file manager for directories, imv for images (png/jpeg/gif/webp/bmp/tiff), Evince for PDF, Chromium for web and http/https URLs, mpv for all video formats, HEY for mailto links, nvim for text and source file types. Rebuild the desktop application database.
source: install/config/mimetypes.sh
dependencies: OM-008 (Chromium), OM-012 (mpv, imv), OM-013 (signal), OM-021 (HEY webapp), OM-015 (nautilus, evince)
ordering: config phase; after all application packages and webapp launchers are installed
translator-capability: invoke xdg-mime and xdg-settings to register MIME type associations; rebuild desktop application database

---

### OM-045

category: config-dotfile
intent: Configure XDG user directories to point standard directories (Downloads, Pictures, Videos) to $HOME subdirectories, and remap the unused Desktop, Templates, and Public directories to $HOME itself to prevent them from appearing in the file manager sidebar. Create a GTK bookmarks file for quick navigation.
source: install/config/user-dirs.sh
dependencies: OM-015 (nautilus)
ordering: config phase
translator-capability: invoke xdg-user-dirs-update to set directory paths; write a GTK bookmarks file

---

### OM-046

category: config-dotfile
intent: Initialize the Omarchy feature toggle state directory and create empty toggle configuration files for the three toggleable subsystems: screen lock (hyprlock), notification daemon (mako), and application launcher styling (walker). Toggles are user-visible on/off switches for optional features.
source: install/config/toggles.sh
dependencies: OM-028
ordering: config phase; before toggle-dependent configuration
translator-capability: create state files for a feature toggle system

---

### OM-047

category: arbitrary-script
intent: Install Nautilus file manager Python extensions for LocalSend (file sharing integration) and Transcode (video transcoding right-click action) by deploying the extension scripts to the Nautilus Python extensions directory.
source: install/config/nautilus-python.sh
dependencies: OM-015 (nautilus-python package must be installed)
ordering: config phase
translator-capability: deploy plugin files to an application extension directory

---

### OM-048

category: arbitrary-script
intent: Update the system file-locate database so that the locate command can find all files installed during the packaging phase.
source: install/config/localdb.sh
dependencies: all packaging opinions (OM-006 through OM-027)
ordering: config phase; after all packaging is complete
translator-capability: run the system file-locate database update command (updatedb)

---

### OM-049

category: arbitrary-script
intent: Configure the Walker application launcher as an autostart entry, set it up as a systemd user service with automatic restart on crash, create a package-manager post-upgrade hook to restart Walker after updates, and link the Omarchy visual theme menu configuration files.
source: install/config/walker-elephant.sh
dependencies: OM-015 (walker/omarchy-walker package), OM-028 (config deployed)
ordering: config phase
translator-capability: create XDG autostart entries; write systemd user service drop-ins; create package manager post-update hooks; create symlinks in config directories

---

### OM-050

category: config-dotfile
intent: Configure the systemd service manager for fast shutdown: reduce the default timeout for stopping services and user sessions to allow quicker power-off and reboot operations.
source: install/config/fast-shutdown.sh
dependencies: none
ordering: config phase
translator-capability: deploy systemd system.conf.d and user@.service.d drop-in files for shutdown timeout tuning; run daemon-reload

---

### OM-051

category: arbitrary-script
intent: Install a system-sleep hook that unmounts FUSE filesystems before suspend, preventing filesystem corruption on systems with active FUSE mounts (e.g. sshfs, rclone).
source: install/config/unmount-fuse.sh
dependencies: none
ordering: config phase
translator-capability: install an executable sleep-hook script into the systemd system-sleep hook directory

---

### OM-052

category: config-dotfile
intent: Grant passwordless sudo access specifically for the asdcontrol brightness-control tool, allowing the current user to control Apple Studio Display brightness without a password.
source: install/config/sudoless-asdcontrol.sh
dependencies: OM-014 (asdcontrol package)
ordering: config phase
translator-capability: write a sudoers drop-in granting passwordless NOPASSWD access to a specific binary

---

### OM-053

category: user-group
intent: Add the current user to the input group so that dictation tools and game controller applications can access /dev/input devices without requiring root.
source: install/config/input-group.sh
dependencies: none
ordering: config phase
translator-capability: add a user to a system group (input)

---

### OM-054

category: arbitrary-script
intent: Deploy the Omarchy AI skill definition as a symbolic link into each AI coding assistant's skills directory (~/.agents/skills, ~/.claude/skills, ~/.codex/skills, ~/.pi/agent/skills), making the Omarchy system context available to all installed AI agents immediately.
source: install/config/omarchy-ai-skill.sh
dependencies: OM-010 (claude-code), OM-023 (codex, pi via npm)
ordering: config phase
translator-capability: create symbolic links across multiple agent tool skill directories; handle multiple AI tool ecosystems in parallel

---

### OM-055

category: arbitrary-script
intent: Install a custom Pi coding agent extension that provides Omarchy system theme awareness to the Pi agent.
source: install/config/pi.sh
dependencies: OM-023 (pi npm tool must be installed)
ordering: config phase; after npm installs
translator-capability: deploy an agent extension file to the Pi agent extensions directory

---

### OM-056

category: config-dotfile
intent: Initialize the Omarchy Hyprland feature toggle flags file, which controls which optional compositor features are enabled or disabled at runtime.
source: install/config/omarchy-toggles.sh
dependencies: OM-028, OM-046
ordering: config phase; after toggles initialization
translator-capability: copy a Lua configuration file to a user state directory

---

### OM-057

category: service-enable
intent: Enable the kernel modules cleanup service, which removes orphaned kernel module files after kernel upgrades.
source: install/config/kernel-modules-hook.sh
dependencies: OM-001 (kernel-modules-hook package must be installed via base packages)
ordering: config phase
translator-capability: enable a systemd service unit (supports chroot context where systemctl enable cannot start the unit)

---

### OM-058

category: config-dotfile
intent: Configure automatic power profile switching via udev rules on AC adapter plug/unplug events (balanced on battery, performance on AC). Enable the power profiles daemon service. Only applied on systems with a battery.
source: install/config/powerprofilesctl-rules.sh
dependencies: OM-016 (power-profiles-daemon package)
ordering: config phase
condition: omarchy-battery-present helper returns true (detects presence of a battery/AC adapter)
translator-capability: write udev rules for power event handling; enable a systemd service; trigger udev rule reload

---

### OM-059

category: config-dotfile
intent: Configure automatic WiFi power-save mode switching via udev rules: enable power-save on battery, disable it on AC power. Only applied on systems with a battery.
source: install/config/wifi-powersave-rules.sh
dependencies: OM-014 (iwd for WiFi management)
ordering: config phase
condition: omarchy-battery-present helper returns true
translator-capability: write udev rules for WiFi power management; trigger udev rule reload

---

### OM-060

category: config-dotfile
intent: Configure the plocate file-locate database update service to only run when the system is on AC power, preventing battery drain from index rebuilds on portable hardware.
source: install/config/plocate-ac-only.sh
dependencies: OM-014 (plocate package)
ordering: config phase
translator-capability: write a systemd service drop-in with an AC power condition; run daemon-reload

---

### PHASE 3: Config — Hardware (General)

---

### OM-061

category: hardware-conditional
intent: Enable the wireless networking daemon and disable the systemd-networkd-wait-online service to prevent boot timeout on systems where network is managed by iwd rather than networkd.
source: install/config/hardware/network.sh
dependencies: OM-014 (iwd package)
ordering: config/hardware phase
condition: always applied (non-conditional; applies to all systems using iwd for WiFi)
translator-capability: enable a systemd service unit; disable and mask another systemd service unit

---

### OM-062

category: hardware-conditional
intent: Detect the system's wireless regulatory domain from the timezone configuration and set it in the wireless-regdom configuration file so WiFi operates at the correct power levels and channel set for the user's country.
source: install/config/hardware/set-wireless-regdom.sh
dependencies: OM-061 (iwd enabled), OM-033 (timezone configuration)
ordering: config/hardware phase
condition: applied when /etc/conf.d/wireless-regdom exists and the region is not already set; uses timezone to derive country code
translator-capability: read system timezone; derive wireless regulatory domain country code; write regulatory domain configuration; invoke wireless reg domain set command

---

### OM-063

category: config-dotfile
intent: Configure Apple-style keyboards (and compatible keyboards such as Lofree Flow84) to treat the top-row function keys as standard F-keys (fnmode=2) rather than media keys by default.
source: install/config/hardware/fix-fkeys.sh
dependencies: none
ordering: config/hardware phase
condition: always applied (harmless on non-Apple keyboards; activated only when hid_apple module loads)
translator-capability: write a kernel module option configuration file (/etc/modprobe.d/hid_apple.conf)

---

### OM-064

category: service-enable
intent: Enable the Bluetooth service so that Bluetooth devices can be paired and connected. Configure Bluetooth to remember the last power state across reboots (not forcing AutoEnable). Configure WirePlumber for automatic Bluetooth A2DP audio profile connection.
source: install/config/hardware/bluetooth.sh
dependencies: OM-014 (bluetui, bolt packages), OM-028 (wireplumber config deployed)
ordering: config/hardware phase
translator-capability: enable a systemd service unit (chroot-compatible); modify a configuration file entry; deploy a WirePlumber config drop-in

---

### OM-065

category: service-enable
intent: Enable CUPS printing service, Avahi mDNS daemon (for network printer discovery), and cups-browsed for automatic remote printer detection. Configure systemd-resolved to not handle multicast DNS (delegating it to Avahi). Configure nsswitch.conf to resolve .local domains via mDNS.
source: install/config/hardware/printer.sh
dependencies: OM-016 (cups packages, avahi package)
ordering: config/hardware phase
translator-capability: enable multiple systemd service units; write a resolved drop-in; modify nsswitch.conf; append to CUPS browsed configuration

---

### OM-066

category: config-dotfile
intent: Disable USB autosuspend globally to prevent peripherals (keyboards, mice, audio interfaces) from disconnecting when the system suspends their USB port.
source: install/config/hardware/usb-autosuspend.sh
dependencies: none
ordering: config/hardware phase
condition: always applied (prevents USB autosuspend system-wide via kernel module option)
translator-capability: write a kernel module option configuration file (/etc/modprobe.d/) to disable usbcore autosuspend

---

### OM-067

category: config-dotfile
intent: Configure the system to ignore the hardware power button press (instead of triggering shutdown), allowing the power button to be bound to a user-visible power menu in the desktop environment.
source: install/config/hardware/ignore-power-button.sh
dependencies: none
ordering: config/hardware phase
translator-capability: modify /etc/systemd/logind.conf to set HandlePowerKey=ignore

---

### OM-068

category: hardware-conditional
intent: Detect NVIDIA GPUs and install the appropriate driver stack: for Turing+ GPUs with GSP firmware support (RTX 20xx and newer) install open DKMS drivers with VAAPI support; for Maxwell/Pascal/Volta GPUs install the 580xx legacy DKMS drivers. Configure modprobe for early KMS loading, mkinitcpio for early module inclusion, and set Hyprland environment variables appropriate to the GPU generation.
source: install/config/hardware/nvidia.sh
dependencies: OM-001 (repo must be configured for NVIDIA packages), OM-028 (Hyprland envs.lua must be deployed)
ordering: config/hardware phase
condition: lspci | grep -qi 'nvidia' — at least one NVIDIA GPU present in PCI device list
translator-capability: detect GPU model; conditionally install DKMS driver packages; write modprobe kernel module options; write mkinitcpio config drop-in; append environment variable declarations to compositor config

---

### OM-069

category: hardware-conditional
intent: Install Vulkan graphics drivers matching detected GPU hardware: vulkan-intel for Intel GPUs, vulkan-radeon for AMD GPUs, vulkan-asahi for Apple Silicon GPUs. Multiple drivers may be installed if multiple GPUs are present.
source: install/config/hardware/vulkan.sh
dependencies: OM-001
ordering: config/hardware phase; after nvidia.sh (NVIDIA Vulkan comes from nvidia-utils, not this script)
condition: lspci detects a VGA/Display device matching Intel, AMD, or Apple vendor
translator-capability: detect GPU vendors from PCI device list; conditionally install vendor-specific Vulkan driver packages

---

### PHASE 3: Config — Hardware Intel

---

### OM-070

category: hardware-conditional
intent: Install Intel hardware video acceleration drivers: intel-media-driver + libvpl + vpl-gpu-rt for HD Graphics / Iris / Xe / Arc / Panther Lake GPUs; libva-intel-driver for older Intel GMA (pre-2014) GPUs.
source: install/config/hardware/intel/video-acceleration.sh
dependencies: OM-001
ordering: config/hardware/intel phase
condition: lspci detects an Intel GPU in the VGA/3D/Display PCI class
translator-capability: detect Intel GPU generation from lspci output; conditionally install generation-appropriate video acceleration packages

---

### OM-071

category: hardware-conditional
intent: Install and enable Intel Low Power Mode Daemon (intel-lpmd) for supported hybrid Intel CPUs (Alder Lake model 151/154, Raptor Lake 183/186/191, Meteor Lake 170/172, Lunar Lake 189, Panther Lake 204). Only applied on battery-equipped systems.
source: install/config/hardware/intel/lpmd.sh
dependencies: OM-001
ordering: config/hardware/intel phase
condition: omarchy-hw-intel AND omarchy-battery-present AND cpu model in [151,154,170,172,183,186,189,191,204]
translator-capability: read CPU model number from /proc/cpuinfo; conditionally install and enable a power management daemon

---

### OM-072

category: hardware-conditional
intent: Install and enable thermald for Intel laptops with Sandy Bridge (model >= 42) or newer CPUs. Only applied on battery-equipped systems. Thermald provides thermal management beyond the default CPU throttling.
source: install/config/hardware/intel/thermald.sh
dependencies: OM-001
ordering: config/hardware/intel phase
condition: omarchy-hw-intel AND omarchy-battery-present AND cpu model >= 42 (Sandy Bridge and newer)
translator-capability: read CPU model number from /proc/cpuinfo; conditionally install and enable a thermal management daemon

---

### OM-073

category: hardware-conditional
intent: Install MIPI camera support for Intel IPU7 hardware (identified by OVTI08F4 ACPI HID) found in recent Intel Panther Lake laptops.
source: install/config/hardware/intel/ipu7-camera.sh
dependencies: OM-001
ordering: config/hardware/intel phase
condition: /sys/bus/acpi/devices/*/hid contains "OVTI08F4"
translator-capability: probe ACPI device IDs; conditionally install camera firmware/driver packages

---

### OM-074

category: hardware-conditional
intent: Install the Panther Lake-specific kernel (linux-ptl) on Dell XPS Panther Lake systems, which includes audio driver patches not yet in mainline. Remove the standard linux kernel to avoid dual-kernel confusion. Configure the bootloader to show only the PTL kernel entry in the boot menu.
source: install/config/hardware/intel/ptl-kernel.sh
dependencies: OM-001
ordering: config/hardware/intel phase; before limine-snapper (which rebuilds the UKI)
condition: omarchy-hw-match "XPS" AND omarchy-hw-intel-ptl — DMI product name matches "XPS" and Intel PTL CPU detected
translator-capability: install an alternative kernel package; remove the default kernel package; write a bootloader entry-tool configuration file to control boot entry ordering

---

### OM-075

category: kernel-boot-param
intent: Enable the FRED (Flexible Return and Event Delivery) kernel feature on Intel Panther Lake systems by adding the fred=on kernel command-line parameter. FRED is a new interrupt-delivery mechanism that improves performance on Panther Lake.
source: install/config/hardware/intel/fred.sh
dependencies: none (limine-entry-tool.d drop-in consumed by limine-snapper)
ordering: config/hardware/intel phase; before limine-snapper
condition: omarchy-hw-intel-ptl — Intel Panther Lake CPU detected
translator-capability: write a bootloader entry-tool drop-in file that appends to the kernel command line

---

### OM-076

category: hardware-conditional
intent: Disable WiFi 7 (EHT / 802.11be) on Intel BE200 and BE211 cards [PCI IDs 8086:e440 and 8086:272b] via iwlwifi module option, working around a broken EHT RX data path that causes the AP to fall back to MCS 0 NSS 1 (effectively unusable). Falls back to WiFi 6 (802.11ax) at full speed. Intended as a temporary workaround until the firmware/driver is fixed.
source: install/config/hardware/intel/fix-wifi7-eht.sh
dependencies: none
ordering: config/hardware/intel phase
condition: lspci -nn | grep -qE '[8086:(e440|272b)]' — Intel BE200 or BE211 WiFi card present
translator-capability: write a kernel module option configuration file to disable a specific driver feature flag

---

### OM-077

category: hardware-conditional
intent: Install Sound Open Firmware (sof-firmware) for the audio DSP on Intel Panther Lake systems that are NOT Dell XPS (which uses linux-ptl that already hard-depends on sof-firmware). Without sof-firmware, the DSP fails to initialize and only a null audio device appears.
source: install/config/hardware/intel/sof-firmware.sh
dependencies: OM-001
ordering: config/hardware/intel phase
condition: omarchy-hw-intel-ptl AND NOT omarchy-hw-match "XPS"
translator-capability: conditionally install firmware packages; handle the logical NOT of another hardware condition

---

### PHASE 3: Config — Hardware ASUS

---

### OM-078

category: kernel-boot-param
intent: Fix the display backlight on ASUS ExpertBook B9406 and Zenbook UX5406AA (Panther Lake / Xe3 iGPU) by enabling DPCD AUX backlight mode (xe.enable_dpcd_backlight=1) in the kernel command line. Without this, brightness control changes the backlight register but produces no visible effect.
source: install/config/hardware/asus/fix-asus-ptl-display-backlight.sh
dependencies: none
ordering: config/hardware/asus phase; before limine-snapper
condition: omarchy-hw-asus-expertbook-b9406 OR omarchy-hw-asus-zenbook-ux5406aa — DMI product name matches ExpertBook B9406 or Zenbook UX5406AA
translator-capability: write a bootloader entry-tool drop-in file to append a kernel parameter

---

### OM-079

category: kernel-boot-param
intent: Disable Panel Replay (xe.enable_panel_replay=0) on ASUS ExpertBook B9406 to fix a display wake/update issue where the Xe3 GPU's new Panel Replay feature has a broken wake path on this specific eDP panel, causing the screen to freeze after display sleep.
source: install/config/hardware/asus/fix-asus-ptl-b9406-display.sh
dependencies: none
ordering: config/hardware/asus phase; before limine-snapper
condition: omarchy-hw-asus-expertbook-b9406 — DMI product name matches ExpertBook B9406
translator-capability: write a bootloader entry-tool drop-in file to append a kernel parameter

---

### OM-080

category: hardware-conditional
intent: Work around libinput's touch jump detection discarding all motion events from the ASUS ExpertBook B9406 touchpad (Pixart 093A:4F05) by masking the pressure axes via a libinput quirks file. Without this, the touchpad registers button clicks but not cursor movement.
source: install/config/hardware/asus/fix-asus-ptl-b9406-touchpad.sh
dependencies: none
ordering: config/hardware/asus phase
condition: omarchy-hw-asus-expertbook-b9406 — DMI product name matches ExpertBook B9406
translator-capability: write a libinput quirks override file to mask specific HID report axes

---

### OM-081

category: hardware-conditional
intent: Fix audio volume control on ASUS ROG laptops by enabling a PipeWire/WirePlumber soft mixer profile for the ALC285 codec. Also unmute the Master control which is often muted by default on these devices.
source: install/config/hardware/asus/fix-audio-mixer.sh
dependencies: OM-028 (wireplumber config directory deployed)
ordering: config/hardware/asus phase
condition: omarchy-hw-asus-rog — DMI chassis vendor matches ASUS ROG
translator-capability: deploy a WirePlumber configuration file; invoke alsa mixer controls to set volume levels

---

### OM-082

category: hardware-conditional
intent: Fix internal microphone gain on ASUS ROG laptops with Realtek ALC285 codec by setting the mic boost to 0 and capture level to 70%, then storing the ALSA state persistently. Default mic boost causes clipping during recording.
source: install/config/hardware/asus/fix-mic.sh
dependencies: OM-081 (audio card must be configured)
ordering: config/hardware/asus phase; after fix-audio-mixer
condition: omarchy-hw-asus-rog AND ALC285 codec present in /proc/asound
translator-capability: invoke alsa mixer controls; store ALSA state via alsactl

---

### OM-083

category: hardware-conditional
intent: Mark the ASUS ROG Flow Z13 detachable keyboard touchpad as an integrated (internal) input device via a udev rule, so libinput's disable-while-typing feature correctly pairs it with the keyboard and suppresses ghost touchpad taps from keyboard vibration.
source: install/config/hardware/asus/fix-z13-touchpad.sh
dependencies: none
ordering: config/hardware/asus phase
condition: omarchy-hw-asus-rog AND omarchy-hw-match "GZ302" — ASUS ROG AND DMI product name matches GZ302 (Flow Z13)
translator-capability: write a udev rules file to set device integration attributes; reload udev rules

---

### PHASE 3: Config — Hardware Framework

---

### OM-084

category: hardware-conditional
intent: Set the audio card profile on Framework 13 AMD laptops to a specific HiFi profile that enables both microphone inputs and the speaker, working around the default profile selection that often omits input channels.
source: install/config/hardware/framework/fix-f13-amd-audio-input.sh
dependencies: none
ordering: config/hardware/framework phase
condition: pactl detects an audio card with "Family 17h/19h" in its name (AMD Ryzen audio)
translator-capability: invoke PulseAudio/PipeWire card profile selection via pactl

---

### OM-085

category: hardware-conditional
intent: Install a udev rule that grants unprivileged access to the Framework 16 keyboard HID device for RGB control via qmk_hid. Without this rule, RGB control requires root.
source: install/config/hardware/framework/qmk-hid.sh
dependencies: OM-025 (qmk-hid package), OM-001
ordering: config/hardware/framework phase
condition: omarchy-hw-framework16 — DMI product name matches Framework 16
translator-capability: deploy a udev rules file; reload and trigger udev rules

---

### PHASE 3: Config — Hardware Apple

---

### OM-086

category: hardware-conditional
intent: Install the SPI keyboard driver (macbook12-spi-driver-dkms) for MacBook models that use SPI instead of USB for keyboard communication (MacBook 8,1 / 9,1 / 10,1 and MacBookPro 13,x / 14,x). Configure mkinitcpio to load the required modules early.
source: install/config/hardware/apple/fix-spi-keyboard.sh
dependencies: OM-001
ordering: config/hardware/apple phase
condition: /sys/class/dmi/id/product_name matches MacBook[89],1 | MacBook1[02],1 | MacBookPro13,[123] | MacBookPro14,[123]
translator-capability: read DMI product name; conditionally install DKMS kernel module packages; write mkinitcpio module configuration

---

### OM-087

category: hardware-conditional
intent: Fix NVMe drive wake-from-suspend failures on affected MacBook models (MacBook 8/9/10,1, MacBookPro 13/14,x) by installing a systemd service that disables D3cold power state for the NVMe controller at the known PCI address (0000:01:00.0).
source: install/config/hardware/apple/fix-suspend-nvme.sh
dependencies: none
ordering: config/hardware/apple phase
condition: /sys/class/dmi/id/product_name matches MacBook(8,1|9,1|10,1)|MacBookPro13,[123]|MacBookPro14,[123] AND /sys/bus/pci/devices/0000:01:00.0/d3cold_allowed exists
translator-capability: read DMI product name; write and enable a systemd service unit

---

### OM-088

category: hardware-conditional
intent: Install full Apple T2 chip support: T2-compatible kernel (linux-t2), audio firmware (apple-t2-audio-config), Broadcom WiFi/Bluetooth firmware (apple-bcm-firmware), fan control daemon (t2fanrd), Touch Bar driver (tiny-dfr). Add user to video group for Touch Bar access. Enable T2 services. Configure kernel module auto-loading for T2 hardware. Configure mkinitcpio for T2 boot modules. Set kernel parameters (intel_iommu=on, iommu=pt, pcie_ports=compat) for T2 compatibility. Configure fan speed curve.
source: install/config/hardware/apple/fix-t2.sh
dependencies: OM-001, OM-088a: also adds arch-mact2 repo (see post-install/pacman.sh OM-098)
ordering: config/hardware/apple phase; post-install/pacman.sh adds the arch-mact2 repo which provides linux-t2
condition: lspci -nn | grep -q "106b:180[12]" — Apple T2 security chip PCI ID detected
translator-capability: detect T2 chip by PCI ID; install T2-specific kernel and hardware support packages; enable multiple systemd services; write module-load configurations; write mkinitcpio module list; write bootloader kernel parameters; write hardware daemon configuration

---

### PHASE 3: Config — Hardware Lenovo

---

### OM-089

category: hardware-conditional
intent: Fix the bass speaker output on Lenovo Yoga Pro 7 14IAH10 by writing a kernel module option that applies a pin quirk to route audio to both AMP channels on the ALC287 codec.
source: install/config/hardware/lenovo/fix-yoga-pro7-bass-speakers.sh
dependencies: none
ordering: config/hardware/lenovo phase
condition: omarchy-hw-match "Yoga Pro 7 14IAH10" — DMI product name matches this exact Lenovo model
translator-capability: write a kernel module option configuration file for a sound driver pin quirk

---

### PHASE 3: Config — Hardware Misc (no sub-vendor dir)

---

### OM-090

category: hardware-conditional
intent: Install Broadcom WiFi drivers (broadcom-wl DKMS) for MacBook and other systems with BCM4360 (PCI 14e4:43a0) or BCM4331 (14e4:4331) chipsets.
source: install/config/hardware/fix-bcm43xx.sh
dependencies: OM-001
ordering: config/hardware phase
condition: lspci -nn detects PCI ID 14e4:43a0 (BCM4360) or 14e4:4331 (BCM4331)
translator-capability: read PCI device IDs; conditionally install DKMS WiFi driver packages

---

### OM-091

category: hardware-conditional
intent: Configure the mkinitcpio initramfs to include the Surface-specific kernel modules required for keyboard input on Microsoft Surface devices. Attempts to auto-detect the active pinctrl module and includes surface_aggregator, surface_hid, and related Surface bus modules.
source: install/config/hardware/fix-surface-keyboard.sh
dependencies: none
ordering: config/hardware phase
condition: omarchy-hw-surface — DMI product name matches Microsoft Surface family
translator-capability: detect loaded kernel modules (pinctrl); write mkinitcpio module list configuration

---

### OM-092

category: hardware-conditional
intent: Install DKMS driver for the Motorcomm YT6801 Ethernet adapter used in the Slimbook Executive laptop.
source: install/config/hardware/fix-yt6801-ethernet-adapter.sh
dependencies: OM-001
ordering: config/hardware phase
condition: lspci detects "YT6801" or "Motorcomm.*Ethernet" in PCI device list
translator-capability: detect PCI device by name pattern; conditionally install DKMS driver packages

---

### OM-093

category: hardware-conditional
intent: Enable Synaptics InterTouch (psmouse with synaptics_intertouch=1) for touchpad devices that present as Synaptics in the input device list but whose psmouse driver is not yet loaded, providing improved touchpad responsiveness.
source: install/config/hardware/fix-synaptic-touchpad.sh
dependencies: none
ordering: config/hardware phase
condition: /proc/bus/input/devices contains "synaptics" (case-insensitive) AND psmouse module is not loaded
translator-capability: detect touchpad type from kernel input devices; conditionally load a kernel module with specific options

---

### OM-094

category: hardware-conditional
intent: Install Tuxedo/Clevo keyboard backlight drivers (tuxedo-drivers-nocompatcheck-dkms) for Tuxedo laptops and Slimbook Executive (Clevo chassis). Blacklist the conflicting clevo_xsm_wmi legacy module. Remove any orphaned legacy module files.
source: install/config/hardware/fix-tuxedo-backlight.sh
dependencies: OM-001
ordering: config/hardware phase
condition: /sys/class/dmi/id/sys_vendor matches "TUXEDO" or "Slimbook" (case-insensitive)
translator-capability: detect system vendor from DMI; install DKMS driver packages; write kernel module blacklist configuration; remove specific kernel module files

---

### PHASE 4: Login

---

### OM-095

category: theming
intent: Install the Omarchy custom Plymouth boot splash theme and set it as the active default theme so the Omarchy branded animation appears during boot.
source: install/login/plymouth.sh
dependencies: OM-016 (plymouth package), default/plymouth/ theme assets
ordering: login phase; before limine-snapper (which triggers mkinitcpio rebuild that embeds the theme)
translator-capability: copy a Plymouth theme directory to the system themes location; set the default Plymouth theme

---

### OM-096

category: config-dotfile
intent: Create a pre-configured GNOME keyring file that is unlocked without a password (for use with auto-login setups). This prevents prompts for keyring unlock that would appear at desktop startup when no password-protected keyring exists.
source: install/login/default-keyring.sh
dependencies: OM-016 (gnome-keyring package)
ordering: login phase
translator-capability: write keyring descriptor files with correct permissions to the GNOME keyrings directory

---

### OM-097

category: display-manager
intent: Install the Omarchy SDDM display manager theme, configure SDDM for Wayland operation with a custom Hyprland compositor command, enable auto-login for the current user into the Omarchy session. Disable password-based SDDM logins from creating an encrypted login keyring (which conflicts with the passwordless auto-unlock keyring created in OM-096). Enable the SDDM service.
source: install/login/sddm.sh
dependencies: OM-015 (sddm package), OM-096 (default keyring configured)
ordering: login phase; after default-keyring.sh
translator-capability: deploy an SDDM theme; write SDDM configuration drop-ins for Wayland session and autologin; modify PAM SDDM configuration; enable a display manager systemd service

---

### OM-098

category: arbitrary-script
intent: Set up hibernation support: configure the swap device, set resume kernel parameters, and install the mkinitcpio resume hook. The --no-rebuild flag defers the initramfs rebuild to the limine-snapper step to avoid redundant rebuilds.
source: install/login/hibernation.sh
dependencies: none (uses omarchy-hibernation-setup helper from bin/)
ordering: login phase; before limine-snapper.sh which does the final mkinitcpio rebuild
translator-capability: configure a swap resume device; add kernel resume parameters; install mkinitcpio hooks; coordinate with the subsequent bootloader step

---

### OM-099

category: bootloader-config
intent: Configure the Limine bootloader with Omarchy-specific mkinitcpio hooks (including Plymouth, btrfs-overlayfs, encrypt), set up UKI generation, configure the btrfs snapshot boot menu via limine-snapper-sync, set up the Thunderbolt module, create a root snapper configuration and disable btrfs quotas for performance. Enable limine-snapper-sync service. Restore mkinitcpio hooks disabled in OM-002. Trigger a full UKI rebuild. Register the Limine EFI boot entry and remove the archinstall-created entry.
source: install/login/limine-snapper.sh
dependencies: OM-002 (mkinitcpio hooks were disabled here; restored here), OM-095 (Plymouth theme), OM-098 (hibernation), OM-074 (PTL kernel drop-ins), OM-075 (fred kernel param), OM-078/079 (ASUS display fixes)
ordering: login phase; last login step; must run after all kernel parameter drop-ins are written
translator-capability: configure mkinitcpio hooks and modules; manage a Limine bootloader installation; install limine-mkinitcpio-hook and limine-snapper-sync packages; create a snapper configuration; disable btrfs quota accounting; enable a systemd service; manage EFI boot entries

---

### PHASE 5: Post-Install

---

### OM-100

category: custom-repo
intent: Restore the final package manager configuration (from the default template) for the target mirror channel after the install is complete. For Apple T2 systems, also append the arch-mact2 repository (SigLevel = Never) to the package manager configuration, enabling linux-t2 kernel updates.
source: install/post-install/pacman.sh
dependencies: OM-001 (initial pacman config), OM-088 (T2 detection)
ordering: post-install phase; after all packaging and config; this is the final pacman.conf state
translator-capability: deploy a final package manager configuration file; conditionally append an additional unsigned repository (SigLevel = Never) for T2 hardware
schema-note: arch-mact2 with SigLevel = Never is a security-relevant opinion — translator must explicitly acknowledge unsigned repo trust; threat T-00-SIG

---

### PHASE 6: First-Run (execution-phase: first-run)

All opinions in this phase require a live Wayland display session and are deferred to after first login.

---

### OM-101

category: service-enable
intent: Check for network connectivity on first login. If offline, send desktop notifications prompting the user to set up WiFi (with an iwd-based WiFi UI). Whether online or offline, send a notification prompting the user to run a system update.
source: install/first-run/wifi.sh
dependencies: OM-061 (iwd), OM-015 (notification daemon)
ordering: first-run phase
execution-phase: first-run
translator-capability: send desktop notifications; detect network connectivity; integrate with a WiFi management UI

---

### OM-102

category: theming
intent: Apply GTK theme settings for dark mode (Adwaita-dark color scheme, Yaru-blue icon theme) via GNOME settings, and update the icon cache. This requires a running GNOME settings daemon / gsettings backend, hence deferred to first-run.
source: install/first-run/gnome-theme.sh
dependencies: OM-015 (Yaru icon theme installed), OM-029 (theme system configured)
ordering: first-run phase
execution-phase: first-run
translator-capability: invoke gsettings to set desktop theme and icon theme; invoke gtk-update-icon-cache

---

### OM-103

category: firewall-rule
intent: Configure the system firewall with a default deny-inbound policy, allow all outbound traffic, allow LocalSend peer-to-peer file sharing (UDP/TCP port 53317), allow Docker containers to use the host's DNS resolver (UDP port 53 from Docker bridge ranges), and enable the firewall service on boot. Install Docker firewall protection rules.
source: install/first-run/firewall.sh
dependencies: OM-016 (ufw, ufw-docker packages)
ordering: first-run phase; requires passwordless sudo configured in OM-004
execution-phase: first-run
translator-capability: invoke ufw to configure firewall rules and default policies; invoke ufw-docker to install Docker-specific rules; enable ufw as a systemd service

---

### OM-104

category: dns-config
intent: Configure the system DNS resolver to use systemd-resolved's stub listener by symlinking /etc/resolv.conf to /run/systemd/resolve/stub-resolv.conf, enabling DNSSEC validation and the resolved stub resolver for all DNS queries.
source: install/first-run/dns-resolver.sh
dependencies: none (systemd-resolved is part of base systemd)
ordering: first-run phase
execution-phase: first-run
translator-capability: create a symbolic link from /etc/resolv.conf to the systemd-resolved stub resolver path

---

### OM-105

category: config-dotfile
intent: Detect the current monitor's reported scale factor from the Hyprland compositor and write it as the default GDK_SCALE environment variable in the monitors configuration, ensuring GTK applications render at the correct scale on HiDPI displays without manual configuration.
source: install/first-run/gdk-scale.sh
dependencies: OM-028 (monitors.conf deployed), OM-006 (hyprland compositor running)
ordering: first-run phase
execution-phase: first-run
translator-capability: query the compositor for monitor scale; write a scale factor value into a compositor environment configuration file

---

### OM-106

category: service-enable
intent: Enable the battery monitoring timer service (omarchy-battery-monitor.timer) on systems with a battery to provide low-battery desktop notifications. On systems without a battery, set the power profile to "performance" mode as the permanent default.
source: install/first-run/battery-monitor.sh
dependencies: OM-016 (power-profiles-daemon), OM-058 (power profile switching)
ordering: first-run phase
execution-phase: first-run
condition: omarchy-battery-present — whether a battery is detected determines which branch executes
translator-capability: enable a systemd user timer unit; invoke powerprofilesctl to set the active power profile

---

### OM-107

category: arbitrary-script
intent: Remove the temporary first-run sudoers rule file that granted passwordless sudo access to the set of commands needed during first-run configuration, restoring normal sudo behavior after all first-run scripts have completed.
source: install/first-run/cleanup-reboot-sudoers.sh
dependencies: OM-004 (first-run sudoers rule was created here), all other first-run opinions (this should be last)
ordering: first-run phase; last first-run step
execution-phase: first-run
translator-capability: remove a specific sudoers drop-in file using sudo (the file itself grants this permission)

---

### OM-108

category: service-enable
intent: Enable and start the Elephant application launcher as a systemd user service so it is available immediately after first login and on all subsequent logins.
source: install/first-run/elephant.sh
dependencies: OM-049 (walker-elephant configured the autostart; this starts the service)
ordering: first-run phase
execution-phase: first-run
translator-capability: enable a user-scoped systemd service and start it immediately

---

### OM-109

category: config-dotfile
intent: Enable GTK primary paste (middle-click paste from selection clipboard) as a system-wide GTK preference.
source: install/first-run/gtk-primary-paste.sh
dependencies: none
ordering: first-run phase
execution-phase: first-run
translator-capability: invoke gsettings to set a GTK interface preference

---

### OM-110

category: service-enable
intent: Enable the omarchy-recover-internal-monitor systemd user service which automatically recovers the internal monitor if it becomes unavailable after an external display is disconnected.
source: install/first-run/recover-internal-monitor.sh
dependencies: OM-006 (Hyprland compositor)
ordering: first-run phase
execution-phase: first-run
translator-capability: enable a user-scoped systemd service unit

---

### OM-111

category: service-enable
intent: Start and enable the SwayOSD on-screen display server service, which provides visual feedback overlays for volume, brightness, and caps-lock changes.
source: install/first-run/swayosd.sh
dependencies: OM-015 (swayosd package)
ordering: first-run phase
execution-phase: first-run
translator-capability: reload the systemd user daemon; enable and start a user-scoped systemd service

---

### OM-112

category: arbitrary-script
intent: Send a desktop notification welcoming the user and explaining the key Omarchy keyboard shortcuts (Super+K for keybinding cheatsheet, Super+Space for application launcher, Super+Alt+Space for Omarchy Menu).
source: install/first-run/welcome.sh
dependencies: OM-015 (notification daemon)
ordering: first-run phase
execution-phase: first-run
translator-capability: send desktop notifications with structured text content

---

### OM-113

category: arbitrary-script
intent: Send a desktop notification offering to install Voxtype voice dictation. The hook script removes itself after running (one-shot notification). This is a deferred optional feature prompt, not an automatic install.
source: install/first-run/install-voxtype.hook
dependencies: OM-015 (notification daemon)
ordering: first-run phase
execution-phase: first-run
translator-capability: send a desktop notification; self-delete the hook script after execution

---

### PHASE 7: Themes

All 21 theme entries share the same structure. A theme is an opinionated set of visual assets (color palette, backgrounds, per-application color/style configs) that the user can switch between at runtime without reinstalling the system.

---

### OM-114

category: theming
intent: Provide the Catppuccin (dark Mocha variant) visual theme bundle including: desktop backgrounds, btop color theme, color palette definitions (colors.toml), icon theme link, Neovim color scheme, lock screen image, preview images for the theme selector, and VS Code color theme JSON.
source: themes/catppuccin/
dependencies: OM-029 (theme system must be initialized)
ordering: post-config (themes are selected at runtime; initial theme is set in OM-029)
translator-capability: deploy file asset payloads to theme directories; link per-application config files to the active theme

---

### OM-115

category: theming
intent: Provide the Catppuccin Latte (light variant) visual theme bundle with the same structure as OM-114 but with a light color palette.
source: themes/catppuccin-latte/
dependencies: OM-029
ordering: post-config
translator-capability: deploy file asset payloads to theme directories

---

### OM-116

category: theming
intent: Provide the Ethereal visual theme bundle.
source: themes/ethereal/
dependencies: OM-029
ordering: post-config
translator-capability: deploy file asset payloads to theme directories

---

### OM-117

category: theming
intent: Provide the Everforest visual theme bundle.
source: themes/everforest/
dependencies: OM-029
ordering: post-config
translator-capability: deploy file asset payloads to theme directories

---

### OM-118

category: theming
intent: Provide the Flexoki Light visual theme bundle.
source: themes/flexoki-light/
dependencies: OM-029
ordering: post-config
translator-capability: deploy file asset payloads to theme directories

---

### OM-119

category: theming
intent: Provide the Gruvbox visual theme bundle.
source: themes/gruvbox/
dependencies: OM-029
ordering: post-config
translator-capability: deploy file asset payloads to theme directories

---

### OM-120

category: theming
intent: Provide the Hackerman (dark green terminal-aesthetic) visual theme bundle.
source: themes/hackerman/
dependencies: OM-029
ordering: post-config
translator-capability: deploy file asset payloads to theme directories

---

### OM-121

category: theming
intent: Provide the Kanagawa visual theme bundle.
source: themes/kanagawa/
dependencies: OM-029
ordering: post-config
translator-capability: deploy file asset payloads to theme directories

---

### OM-122

category: theming
intent: Provide the Last Horizon visual theme bundle.
source: themes/last-horizon/
dependencies: OM-029
ordering: post-config
translator-capability: deploy file asset payloads to theme directories

---

### OM-123

category: theming
intent: Provide the Lumon (Severance-inspired) visual theme bundle.
source: themes/lumon/
dependencies: OM-029
ordering: post-config
translator-capability: deploy file asset payloads to theme directories

---

### OM-124

category: theming
intent: Provide the Matte Black visual theme bundle.
source: themes/matte-black/
dependencies: OM-029
ordering: post-config
translator-capability: deploy file asset payloads to theme directories

---

### OM-125

category: theming
intent: Provide the Miasma visual theme bundle.
source: themes/miasma/
dependencies: OM-029
ordering: post-config
translator-capability: deploy file asset payloads to theme directories

---

### OM-126

category: theming
intent: Provide the Nord visual theme bundle.
source: themes/nord/
dependencies: OM-029
ordering: post-config
translator-capability: deploy file asset payloads to theme directories

---

### OM-127

category: theming
intent: Provide the Osaka Jade visual theme bundle.
source: themes/osaka-jade/
dependencies: OM-029
ordering: post-config
translator-capability: deploy file asset payloads to theme directories

---

### OM-128

category: theming
intent: Provide the Retro 82 visual theme bundle.
source: themes/retro-82/
dependencies: OM-029
ordering: post-config
translator-capability: deploy file asset payloads to theme directories

---

### OM-129

category: theming
intent: Provide the Ristretto (warm coffee palette) visual theme bundle.
source: themes/ristretto/
dependencies: OM-029
ordering: post-config
translator-capability: deploy file asset payloads to theme directories

---

### OM-130

category: theming
intent: Provide the Rose Pine visual theme bundle.
source: themes/rose-pine/
dependencies: OM-029
ordering: post-config
translator-capability: deploy file asset payloads to theme directories

---

### OM-131

category: theming
intent: Provide the Solitude visual theme bundle.
source: themes/solitude/
dependencies: OM-029
ordering: post-config
translator-capability: deploy file asset payloads to theme directories

---

### OM-132

category: theming
intent: Provide the Tokyo Night visual theme bundle. This is the default theme activated during install (OM-029).
source: themes/tokyo-night/
dependencies: OM-029
ordering: post-config; this theme is activated by default in OM-029
translator-capability: deploy file asset payloads to theme directories

---

### OM-133

category: theming
intent: Provide the Vantablack (maximum contrast dark) visual theme bundle.
source: themes/vantablack/
dependencies: OM-029
ordering: post-config
translator-capability: deploy file asset payloads to theme directories

---

### OM-134

category: theming
intent: Provide the White (light) visual theme bundle.
source: themes/white/
dependencies: OM-029
ordering: post-config
translator-capability: deploy file asset payloads to theme directories

---

## Summary Statistics

| Phase | OM-NNN Range | Count |
|-------|-------------|-------|
| Preflight | OM-001 to OM-005 | 5 |
| Packaging/Base (logical groups) | OM-006 to OM-017 | 12 |
| Packaging/Specialized | OM-018 to OM-023 | 6 |
| Packaging/Hardware-Conditional | OM-024 to OM-027 | 4 |
| Config/General | OM-028 to OM-060 | 33 |
| Config/Hardware General | OM-061 to OM-069 | 9 |
| Config/Hardware Intel | OM-070 to OM-077 | 8 |
| Config/Hardware ASUS | OM-078 to OM-083 | 6 |
| Config/Hardware Framework | OM-084 to OM-085 | 2 |
| Config/Hardware Apple | OM-086 to OM-088 | 3 |
| Config/Hardware Lenovo | OM-089 | 1 |
| Config/Hardware Misc | OM-090 to OM-094 | 5 |
| Login | OM-095 to OM-099 | 5 |
| Post-Install | OM-100 | 1 |
| First-Run | OM-101 to OM-113 | 13 |
| Themes | OM-114 to OM-134 | 21 |
| **Total** | **OM-001 to OM-134** | **134** |

---

## Schema Surprises

The following categories were not in the original docs/08 taxonomy and represent schema extension requirements:

1. **File asset payloads** (theming bundles) — the schema must express file content references or inline payloads, not just package names
2. **Custom package repository registration** (`custom-repo` category) — adding a signed repo is itself an opinion with SigLevel, URL, mirror, keyring dependency
3. **npm global installs** (`npm-global-install` category) — entirely separate from the OS package manager; schema must express cross-tool installs
4. **Ordering is load-bearing at phase level** — preflight → packaging → config → login → post-install → first-run is mandatory; schema ordering constraints must be expressive enough to capture phase-level dependencies
5. **Multi-predicate hardware conditions** — conditions are compound (omarchy-hw-intel AND omarchy-battery-present AND cpu model in [list]); not simple boolean flags
6. **Arbitrary script payloads with declared capabilities** — omarchy-ai-skill.sh deploys symlinks to multiple agent directories; OM-054 is not expressible as package or config alone
7. **Deferred execution phase** — `execution-phase: first-run` is required to distinguish opinions that need a live Wayland session from install-time opinions

---

## Coverage Notes / Unmapped Scripts

Scripts that are NOT referenced by any `all.sh` and therefore not in the main install pipeline:

### Infrastructure helpers (not opinions — execution infrastructure)

- `install/helpers/` (all files) — logging, error handling, chroot wrappers; these are executor infrastructure, not opinions
- `install/preflight/guard.sh` — re-run guard; sourced directly by preflight/all.sh but is infrastructure, not an opinion
- `install/preflight/begin.sh` — welcome message display; infrastructure
- `install/post-install/allow-reboot.sh` — unlocks reboot after install; infrastructure
- `install/post-install/finished.sh` — displays completion message; infrastructure

### Scripts not called by any all.sh (enumerated, not inventoried)

The `bin/` directory contains 306 `omarchy-*` utility scripts. These are runtime helper commands (e.g. `omarchy-theme-set`, `omarchy-pkg-add`, `omarchy-hibernation-setup`), not install-time opinions. They are invoked BY opinion scripts but do not themselves represent post-install decisions. Noted for Plan 04 open-questions.

The `applications/` directory contains `.desktop` files. These are data files deployed by installer scripts; their contents are captured in the opinion that deploys them (e.g. OM-021 for webapps, OM-097 for SDDM session).

### Migrations (sampled, not inventoried per Pitfall 5)

**Count:** 313 timestamped migration scripts (UNIX timestamp named *.sh files)
**Pattern:** Each migration is a numbered idempotent catch-up script that updates an existing Omarchy installation. Examples sampled: adding packages (omarchy-pkg-add), removing packages (omarchy-pkg-drop), config file updates, service restarts, keybinding additions. Migrations represent the opinion evolution history of an installed system — they encode the delta between opinion states over time, not stable atomic opinions themselves.
**Open question:** Does the DebateOS schema need a migration concept, or is version pinning of opinions sufficient? Recorded in Plan 04 open-questions.

### Config and default directories

- `config/` (32 app subdirs) — dotfile templates deployed by OM-028 (config.sh); contents are captured as part of the config-dotfile opinions that deploy them
- `default/` (25+ subdirs) — system config templates deployed by various opinions; each is referenced in the opinion that uses it

---

## Spot-Check: No Arch Mechanics in intent: Fields

Verified: no `intent:` field in this inventory contains the strings `pacman`, `AUR`, or `mkarchiso`.
Arch-specific mechanics (pacman commands, AUR helper invocations, mkarchiso usage) are confined to `source:` citations and `translator-capability:` fields only.
