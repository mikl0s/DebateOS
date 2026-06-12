#!/usr/bin/env python3
# SPDX-License-Identifier: CC0-1.0
# generate.py — idempotent generator for examples/omarchy/
#
# Reads: research/omarchy-opinion-inventory.md + research/omarchy-points.md
# Emits: examples/omarchy/opinions/OM-NNN.yaml (134 files)
#        examples/omarchy/points/<slug>.yaml (32 files)
#
# Usage (from repo root):
#   python examples/omarchy/gen/generate.py
#
# Idempotent: re-running produces identical output (deterministic YAML emission).
# Requires: PyYAML (python-yaml on Arch; stdlib otherwise)

import os
import re
import sys
import yaml

# ── Paths ───────────────────────────────────────────────────────────────────
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
REPO_ROOT = os.path.abspath(os.path.join(SCRIPT_DIR, "..", "..", "..", ".."))
OPINIONS_DIR = os.path.join(SCRIPT_DIR, "..", "opinions")
POINTS_DIR = os.path.join(SCRIPT_DIR, "..", "points")
INVENTORY_MD = os.path.join(REPO_ROOT, "research", "omarchy-opinion-inventory.md")
POINTS_MD = os.path.join(REPO_ROOT, "research", "omarchy-points.md")

# ── Status policy (RESOLVED OQ-1) ───────────────────────────────────────────
# required: OM-001 (custom-repo), OM-006 (compositor), OM-097 (display-manager),
#           OM-099 (bootloader), all hardware-conditional opinions (gated by hw anyway)
# nice-to-have: themes OM-114..OM-134
# all others: required (most package/config opinions are core to Omarchy operation)
# A small set of truly optional extras are nice-to-have per inventory role.
REQUIRED_IDS = {
    "OM-001", "OM-006", "OM-097", "OM-099",
    # Core infrastructure
    "OM-002", "OM-003", "OM-004", "OM-005",
    "OM-007", "OM-008", "OM-009", "OM-010", "OM-011",
    "OM-012", "OM-013", "OM-014", "OM-015", "OM-016", "OM-017",
    "OM-018", "OM-019", "OM-028", "OM-029",
    "OM-031", "OM-032", "OM-033", "OM-034", "OM-035",
    "OM-036", "OM-037", "OM-038", "OM-039", "OM-040",
    "OM-041", "OM-042", "OM-043", "OM-044", "OM-045",
    "OM-046", "OM-047", "OM-048", "OM-049", "OM-050",
    "OM-051", "OM-052", "OM-053", "OM-057",
    "OM-061", "OM-062", "OM-063", "OM-064", "OM-065",
    "OM-066", "OM-067",
    "OM-095", "OM-096", "OM-098",
    "OM-100",
    "OM-101", "OM-103", "OM-104", "OM-107",
    "OM-108", "OM-110", "OM-111",
}
# All hardware-conditional opinions (OM-024..027, OM-058..060, OM-068..094, OM-106):
# status is required but they are gated by hardware_condition so they resolve
# to Skipped on a baseline vanilla-arch machine (correct and expected).
HARDWARE_CONDITIONAL_IDS = {
    "OM-024", "OM-025", "OM-026", "OM-027",
    "OM-058", "OM-059", "OM-060",
    "OM-068", "OM-069", "OM-070", "OM-071", "OM-072", "OM-073",
    "OM-074", "OM-075", "OM-076", "OM-077",
    "OM-078", "OM-079", "OM-080", "OM-081", "OM-082", "OM-083",
    "OM-084", "OM-085", "OM-086", "OM-087", "OM-088", "OM-089",
    "OM-090", "OM-091", "OM-092", "OM-093", "OM-094",
    "OM-106",
}
NICE_TO_HAVE_IDS = {
    # Themes OM-114..OM-134
    "OM-114", "OM-115", "OM-116", "OM-117", "OM-118", "OM-119",
    "OM-120", "OM-121", "OM-122", "OM-123", "OM-124", "OM-125",
    "OM-126", "OM-127", "OM-128", "OM-129", "OM-130", "OM-131",
    "OM-132", "OM-133", "OM-134",
    # Optional extras
    "OM-020", "OM-021", "OM-022", "OM-023",
    "OM-030", "OM-054", "OM-055", "OM-056",
    "OM-102", "OM-105", "OM-109", "OM-112", "OM-113",
}


def get_status(om_id: str) -> str:
    if om_id in NICE_TO_HAVE_IDS:
        return "nice-to-have"
    return "required"


# ── Canonical opinion data (hand-authored from inventory) ────────────────────
# Each entry: id → dict of YAML fields (omitting schema/id/status which are added)
# Only non-empty optional fields are included.
def opinion_data() -> dict:
    """Return the full set of 134 opinion records keyed by OM-NNN id."""
    data = {}

    # OM-001: Custom repo
    data["OM-001"] = {
        "name": "Omarchy Package Repository",
        "category": "custom-repo",
        "intent": "Register the Omarchy package repository with the system package manager, import its GPG signing key (ID 40DFB630FF42BCFFB047046CF0134EE680CAC571), and set the preferred mirror channel. This makes Omarchy-curated packages available before any other install step. Trust level: OptionalTrustAll.",
        "install_phase": "preflight",
        "translator_capabilities": ["add-custom-package-repo", "import-gpg-key"],
        "custom_repos": [{"name": "omarchy", "url": "https://packages.omarchy.org/stable", "sig_level": "OptionalTrustAll", "priority": 10, "keyring": "40DFB630FF42BCFFB047046CF0134EE680CAC571"}],
    }

    # OM-002
    data["OM-002"] = {
        "name": "Disable Initramfs Hooks During Install",
        "category": "service-enable",
        "intent": "Disable automatic initramfs regeneration hooks during the installation run to prevent multiple redundant rebuilds. Hooks are restored after the bootloader configuration phase completes.",
        "install_phase": "preflight",
        "ordering": {"before": [{"id": "OM-006"}]},
        "translator_capabilities": ["disable-package-manager-hooks", "restore-package-manager-hooks"],
        "services": [{"name": "mkinitcpio-trigger", "enable": False}],
    }

    # OM-003
    data["OM-003"] = {
        "name": "Apply Pending System Migrations",
        "category": "arbitrary-script",
        "intent": "Run all pending system-migration scripts so the existing installation is current before new opinions are applied. Mark each migration as applied so it will not re-run on future install invocations.",
        "install_phase": "preflight",
        "depends_on": [{"id": "OM-001"}],
        "ordering": {"after": [{"id": "OM-001"}], "before": [{"id": "OM-006"}]},
        "translator_capabilities": ["run-migration-scripts", "maintain-applied-migrations-state"],
        "script_payload": {"script": "run-pending-migrations", "capabilities": ["run-migration-scripts"]},
    }

    # OM-004
    data["OM-004"] = {
        "name": "First-Run Marker and Scoped Sudo",
        "category": "arbitrary-script",
        "intent": "Set a marker that tells deferred first-run scripts that they should execute after the user's first login session. Configure passwordless sudo access scoped to the exact set of commands those first-run scripts will need.",
        "install_phase": "preflight",
        "translator_capabilities": ["write-persistent-state-marker", "configure-scoped-sudoers"],
        "script_payload": {"script": "set-first-run-marker-and-sudoers", "capabilities": ["write-persistent-state-marker", "configure-scoped-sudoers"]},
    }

    # OM-005
    data["OM-005"] = {
        "name": "Log Installation Environment",
        "category": "arbitrary-script",
        "intent": "Log the installation environment variables to aid debugging and provide a reproducible record of the install-time inputs.",
        "install_phase": "preflight",
        "translator_capabilities": ["log-structured-key-value-data"],
        "script_payload": {"script": "log-install-env", "capabilities": ["log-structured-key-value-data"]},
    }

    # OM-006
    data["OM-006"] = {
        "name": "Wayland Compositor Stack",
        "category": "package-install",
        "intent": "Install the compositor and Wayland session management stack: tiling compositor (Hyprland), session launcher (UWSM), display portal, idle daemon, screen lock, and color temperature tool.",
        "install_phase": "packaging",
        "depends_on": [{"id": "OM-001"}],
        "translator_capabilities": ["install-packages"],
        "packages": ["hyprland", "uwsm", "hyprland-portal", "hypridle", "hyprlock", "hyprsunset"],
    }

    # OM-007
    data["OM-007"] = {
        "name": "Terminal and Shell Toolchain",
        "category": "package-install",
        "intent": "Install the primary terminal emulator (foot) and terminal-multiplexer (tmux) with supporting tools: shell completion, modern shell enhancements, navigation, file browsing, and search utilities.",
        "install_phase": "packaging",
        "depends_on": [{"id": "OM-001"}],
        "translator_capabilities": ["install-packages"],
        "packages": ["foot", "tmux", "starship", "zoxide", "bat", "eza", "fd", "ripgrep", "fzf", "tldr", "tree-sitter-cli", "fastfetch"],
    }

    # OM-008
    data["OM-008"] = {
        "name": "Web Browser",
        "category": "package-install",
        "intent": "Install the primary web browser (Chromium) and supporting libraries for web content rendering and network stack.",
        "install_phase": "packaging",
        "depends_on": [{"id": "OM-001"}],
        "translator_capabilities": ["install-packages"],
        "packages": ["chromium"],
    }

    # OM-009
    data["OM-009"] = {
        "name": "Developer Toolchain",
        "category": "package-install",
        "intent": "Install developer tooling: version manager, language runtimes (Ruby, Rust, Lua, Python), build tools, and database libraries.",
        "install_phase": "packaging",
        "depends_on": [{"id": "OM-001"}],
        "translator_capabilities": ["install-packages"],
        "packages": ["mise", "ruby", "rust", "lua", "python", "clang", "llvm", "luarocks", "python-poetry-core", "mariadb-libs", "postgresql-libs"],
    }

    # OM-010
    data["OM-010"] = {
        "name": "AI Assistant Native Packages",
        "category": "package-install",
        "intent": "Install AI assistant tools that run as native applications or integration layers.",
        "install_phase": "packaging",
        "depends_on": [{"id": "OM-001"}],
        "translator_capabilities": ["install-packages"],
        "packages": ["claude-code"],
    }

    # OM-011
    data["OM-011"] = {
        "name": "Containerization Tools",
        "category": "package-install",
        "intent": "Install containerization and virtualization tools: container engine, compose, build plugin, and cross-architecture emulation.",
        "install_phase": "packaging",
        "depends_on": [{"id": "OM-001"}],
        "translator_capabilities": ["install-packages"],
        "packages": ["docker", "docker-compose", "docker-buildx", "qemu-user-static-binfmt"],
    }

    # OM-012
    data["OM-012"] = {
        "name": "Media and Creative Tools",
        "category": "package-install",
        "intent": "Install media consumption and creation tools: video player, audio tools, screen recorder, image viewer, image editor, PDF viewer, screenshot tools, and screen annotation.",
        "install_phase": "packaging",
        "depends_on": [{"id": "OM-001"}],
        "translator_capabilities": ["install-packages"],
        "packages": ["mpv", "mpv-mpris", "alsa-utils", "pamixer", "playerctl", "wireplumber", "wiremix", "gpu-screen-recorder", "obs-studio", "imv", "pinta", "evince", "grim", "slurp", "satty", "xournalpp"],
    }

    # OM-013
    data["OM-013"] = {
        "name": "Productivity and Communication Applications",
        "category": "package-install",
        "intent": "Install personal productivity and communication applications: note-taking, secure messaging, office suite, markdown editor, audio streaming, and local file sharing.",
        "install_phase": "packaging",
        "depends_on": [{"id": "OM-001"}],
        "translator_capabilities": ["install-packages"],
        "packages": ["obsidian", "signal-desktop", "libreoffice-fresh", "spotify", "typora", "localsend"],
    }

    # OM-014
    data["OM-014"] = {
        "name": "System Utilities and Hardware Tools",
        "category": "package-install",
        "intent": "Install system utilities covering Bluetooth management, disk tools, file manager and extensions, display brightness control, system monitoring, file search, and network tools.",
        "install_phase": "packaging",
        "depends_on": [{"id": "OM-001"}],
        "translator_capabilities": ["install-packages"],
        "packages": ["bluetui", "bolt", "gnome-disk-utility", "dosfstools", "exfatprogs", "nautilus", "nautilus-python", "sushi", "brightnessctl", "asdcontrol", "btop", "inxi", "impala", "usage", "dua-cli", "plocate", "iwd", "inetutils", "wireless-regdb", "whois"],
    }

    # OM-015
    data["OM-015"] = {
        "name": "Desktop Shell Components",
        "category": "package-install",
        "intent": "Install graphical desktop shell components: status bar, notification daemon, application launcher, background setter, on-screen display, display manager, polkit agent, GTK themes, input method framework, screen tools, and desktop portals.",
        "install_phase": "packaging",
        "depends_on": [{"id": "OM-006"}],
        "translator_capabilities": ["install-packages"],
        "packages": ["waybar", "quickshell", "mako", "omarchy-walker", "swaybg", "swayosd", "sddm", "polkit-gnome", "gnome-themes-extra", "kvantum-qt5", "gvfs-mtp", "gvfs-nfs", "gvfs-smb", "fcitx5", "fcitx5-gtk", "fcitx5-qt", "hyprpicker", "hyprland-preview-share-picker", "xdg-desktop-portal-gtk", "xdg-desktop-portal-hyprland", "xdg-terminal-exec"],
    }

    # OM-016
    data["OM-016"] = {
        "name": "Security and Network Services",
        "category": "package-install",
        "intent": "Install system security and administration tools: firewall, printer support, network service discovery, cryptographic keyring, and power management daemon.",
        "install_phase": "packaging",
        "depends_on": [{"id": "OM-001"}],
        "translator_capabilities": ["install-packages"],
        "packages": ["ufw", "ufw-docker", "cups", "cups-browsed", "cups-filters", "cups-pdf", "avahi", "nss-mdns", "gnome-keyring", "libsecret", "gnome-calculator", "power-profiles-daemon", "plymouth"],
    }

    # OM-017
    data["OM-017"] = {
        "name": "Development Support Packages",
        "category": "package-install",
        "intent": "Install development support packages: version control tools, data processing utilities, session tools, Wayland clipboard, image processing, OCR, and system bindings.",
        "install_phase": "packaging",
        "depends_on": [{"id": "OM-001"}],
        "translator_capabilities": ["install-packages"],
        "packages": ["github-cli", "jq", "xmlstarlet", "wl-clipboard", "imagemagick", "ffmpegthumbnailer", "tesseract", "tesseract-data-eng", "python-gobject", "python-yaml"],
    }

    # OM-018
    data["OM-018"] = {
        "name": "Omarchy Brand Font",
        "category": "font-install",
        "intent": "Deploy the custom Omarchy brand font to the user's local font directory and rebuild the font cache so the font is immediately available to desktop applications.",
        "install_phase": "packaging",
        "depends_on": [{"id": "OM-001"}],
        "ordering": {"after": [{"id": "OM-006"}]},
        "translator_capabilities": ["deploy-font-file", "rebuild-font-cache"],
        "packages": ["omarchy-font"],
    }

    # OM-019
    data["OM-019"] = {
        "name": "Neovim with LazyVim Configuration",
        "category": "package-install",
        "intent": "Install the Neovim editor with the Omarchy-curated LazyVim configuration and color scheme themes via a post-install setup script.",
        "install_phase": "packaging",
        "depends_on": [{"id": "OM-001"}, {"id": "OM-007"}],
        "ordering": {"after": [{"id": "OM-007"}]},
        "translator_capabilities": ["install-packages", "run-post-install-setup-script"],
        "packages": ["neovim"],
        "script_payload": {"script": "setup-omarchy-nvim", "capabilities": ["run-post-install-setup-script"]},
    }

    # OM-020
    data["OM-020"] = {
        "name": "Application Icon Assets",
        "category": "arbitrary-script",
        "intent": "Deploy application icon assets to the user's local application icons directory so that custom launcher entries have the correct visual icons.",
        "install_phase": "packaging",
        "depends_on": [{"id": "OM-015"}],
        "ordering": {"after": [{"id": "OM-015"}]},
        "translator_capabilities": ["copy-icon-files-to-app-icon-directory"],
        "script_payload": {"script": "deploy-app-icons", "capabilities": ["copy-icon-files-to-app-icon-directory"]},
    }

    # OM-021
    data["OM-021"] = {
        "name": "Web Applications as Chromium PWAs",
        "category": "arbitrary-script",
        "intent": "Install a curated set of web applications as Chromium Progressive Web Apps with desktop launcher entries, custom icons, and URL-scheme handler registrations.",
        "install_phase": "packaging",
        "depends_on": [{"id": "OM-008"}, {"id": "OM-020"}],
        "ordering": {"after": [{"id": "OM-020"}]},
        "translator_capabilities": ["create-desktop-launcher-entries", "register-url-scheme-handlers"],
        "script_payload": {"script": "install-chromium-pwas", "capabilities": ["create-desktop-launcher-entries", "register-url-scheme-handlers"]},
    }

    # OM-022
    data["OM-022"] = {
        "name": "TUI Application Launcher Entries",
        "category": "arbitrary-script",
        "intent": "Install terminal UI applications as windowed launcher entries with custom icons and compositor layout mode hints.",
        "install_phase": "packaging",
        "depends_on": [{"id": "OM-020"}, {"id": "OM-015"}],
        "translator_capabilities": ["create-desktop-launcher-entries-with-compositor-hints"],
        "script_payload": {"script": "install-tui-launchers", "capabilities": ["create-desktop-launcher-entries-with-compositor-hints"]},
    }

    # OM-023
    data["OM-023"] = {
        "name": "AI Development Tools via Runtime Package Manager",
        "category": "npm-global-install",
        "intent": "Install AI coding assistant and development tools as global command-line tools via the Node.js package manager.",
        "install_phase": "packaging",
        "depends_on": [{"id": "OM-009"}, {"id": "OM-041"}],
        "ordering": {"after": [{"id": "OM-041"}]},
        "translator_capabilities": ["install-npm-global-packages"],
        "runtime_tool_installs": [{"manager": "npm", "packages": ["@openai/codex", "@google/gemini-cli", "@githubnext/github-copilot-cli", "opencode-ai", "playwright", "@pi-ai/pi", "ghui", "hunk"]}],
    }

    # OM-024
    data["OM-024"] = {
        "name": "ASUS ROG Control Daemon",
        "category": "hardware-conditional",
        "intent": "Install ASUS ROG laptop control daemon for fan curve management, performance mode switching, and keyboard RGB control.",
        "install_phase": "packaging",
        "depends_on": [{"id": "OM-001"}],
        "hardware_condition": {"type": "leaf", "predicate": "hw-dmi-chassis-vendor", "match": "ASUS"},
        "translator_capabilities": ["install-packages", "eval-dmi-hardware-condition"],
        "packages": ["asusctl", "supergfxctl"],
    }

    # OM-025
    data["OM-025"] = {
        "name": "Framework 16 Keyboard RGB Tool",
        "category": "hardware-conditional",
        "intent": "Install QMK HID tool for Framework 16 keyboard RGB and key configuration.",
        "install_phase": "packaging",
        "depends_on": [{"id": "OM-001"}],
        "hardware_condition": {"type": "leaf", "predicate": "hw-dmi-product-name", "match": "Framework 16"},
        "translator_capabilities": ["install-packages", "eval-dmi-hardware-condition"],
        "packages": ["qmk-hid"],
    }

    # OM-026
    data["OM-026"] = {
        "name": "Dell XPS Haptic Touchpad Driver",
        "category": "hardware-conditional",
        "intent": "Install haptic touchpad driver for Dell XPS laptops with haptic touchpad hardware.",
        "install_phase": "packaging",
        "depends_on": [{"id": "OM-001"}],
        "hardware_condition": {"type": "leaf", "predicate": "hw-dmi-product-name", "match": "XPS"},
        "translator_capabilities": ["install-packages", "eval-dmi-hardware-condition"],
        "packages": ["libpinch"],
    }

    # OM-027
    data["OM-027"] = {
        "name": "Microsoft Surface WiFi Firmware",
        "category": "hardware-conditional",
        "intent": "Install Marvell WiFi firmware for Microsoft Surface devices.",
        "install_phase": "packaging",
        "depends_on": [{"id": "OM-001"}],
        "hardware_condition": {"type": "leaf", "predicate": "hw-dmi-sys-vendor", "match": "Microsoft"},
        "translator_capabilities": ["install-packages", "eval-dmi-hardware-condition"],
        "packages": ["linux-firmware-marvell"],
    }

    # OM-028
    data["OM-028"] = {
        "name": "Application Configuration Tree",
        "category": "config-dotfile",
        "intent": "Deploy the complete set of application configuration files to the user's config directory, establishing the baseline Omarchy configuration for all desktop applications. Also deploy the Omarchy default shell startup file.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-006"}],
        "ordering": {"after": [{"id": "OM-017"}]},
        "translator_capabilities": ["deploy-config-file-tree", "deploy-shell-startup-file"],
        "file_assets": [{"src": "omarchy-config/", "dst": "~/.config/"}],
    }

    # OM-029
    data["OM-029"] = {
        "name": "Desktop Theme System Configuration",
        "category": "theming",
        "intent": "Configure the desktop theme system: create the user theme directory, set up browser policy for theme integration, activate the default theme (Tokyo Night), create per-application theme asset symlinks, and set the browser to follow system color scheme.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-028"}],
        "ordering": {"after": [{"id": "OM-028"}]},
        "translator_capabilities": ["manage-multi-app-theme-system", "write-browser-policy-files", "create-config-symlinks"],
        "theme": {"bundle_dir": "themes/tokyo-night", "is_default": True},
    }

    # OM-030
    data["OM-030"] = {
        "name": "ASCII Branding Assets",
        "category": "config-dotfile",
        "intent": "Allow users to customize the ASCII-art branding shown in the system info display and the screensaver. Copies default branding assets to a user-owned branding directory.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-028"}],
        "translator_capabilities": ["copy-branding-assets-to-user-config"],
        "file_assets": [{"src": "omarchy-branding/", "dst": "~/.config/omarchy/branding/"}],
    }

    # OM-031
    data["OM-031"] = {
        "name": "Global Git Identity",
        "category": "config-dotfile",
        "intent": "Set the user's global git identity (name and email) from the values provided at install time, so version control commits are correctly attributed from first use.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-009"}],
        "translator_capabilities": ["write-global-git-configuration"],
        "file_assets": [{"src": "git-config-template", "dst": "~/.gitconfig"}],
    }

    # OM-032
    data["OM-032"] = {
        "name": "GPG Daemon Keyserver Configuration",
        "category": "config-dotfile",
        "intent": "Configure the GPG daemon to use multiple keyservers for improved reliability. Deploy the configuration to the system-wide GnuPG config directory and restart the daemon.",
        "install_phase": "config",
        "translator_capabilities": ["deploy-gpg-daemon-config", "restart-gpg-daemon"],
        "file_assets": [{"src": "gnupg/gpg-agent.conf", "dst": "/etc/gnupg/gpg-agent.conf"}],
    }

    # OM-033
    data["OM-033"] = {
        "name": "Timezone Management Sudo Access",
        "category": "arbitrary-script",
        "intent": "Grant passwordless sudo access to the timezone management commands for all members of the wheel group, so users can update their timezone without a password prompt.",
        "install_phase": "config",
        "translator_capabilities": ["write-sudoers-drop-in"],
        "file_assets": [{"src": "sudoers-timezone", "dst": "/etc/sudoers.d/timezone"}],
    }

    # OM-034
    data["OM-034"] = {
        "name": "Increase Password Attempt Limit",
        "category": "config-dotfile",
        "intent": "Increase the maximum number of password entry attempts in the sudo and screen-lock tools from the default three to ten, reducing accidental lockouts from mistyped passwords.",
        "install_phase": "config",
        "translator_capabilities": ["write-sudoers-drop-in", "modify-pam-faillock-config"],
        "file_assets": [{"src": "sudoers-passwd-tries", "dst": "/etc/sudoers.d/passwd-tries"}],
    }

    # OM-035
    data["OM-035"] = {
        "name": "Account Lockout Threshold",
        "category": "config-dotfile",
        "intent": "Set the account lockout threshold to ten failed attempts with a two-minute unlock timeout, and ensure the auto-login session resets the fail counter on success.",
        "install_phase": "config",
        "translator_capabilities": ["modify-pam-system-auth", "modify-pam-sddm-autologin"],
        "file_assets": [{"src": "pam-faillock.conf", "dst": "/etc/security/faillock.conf"}],
    }

    # OM-036
    data["OM-036"] = {
        "name": "TCP MTU Probing",
        "category": "sysctl-param",
        "intent": "Enable TCP MTU probing to reduce connection flakiness caused by path MTU issues.",
        "install_phase": "config",
        "translator_capabilities": ["write-sysctl-drop-in"],
        "sysctl_params": [{"key": "net.ipv4.tcp_mtu_probing", "value": "1", "drop_in_file": "omarchy-tcp-mtu.conf"}],
    }

    # OM-037
    data["OM-037"] = {
        "name": "Inotify File Watcher Limit",
        "category": "sysctl-param",
        "intent": "Increase the inotify file watcher limit from the default 8192 to 524288, sufficient for large development projects using file watchers and bundlers.",
        "install_phase": "config",
        "translator_capabilities": ["write-sysctl-drop-in"],
        "sysctl_params": [{"key": "fs.inotify.max_user_watches", "value": "524288", "drop_in_file": "omarchy-inotify.conf"}],
    }

    # OM-038
    data["OM-038"] = {
        "name": "File Descriptor Limits",
        "category": "sysctl-param",
        "intent": "Raise the system-wide and per-user file descriptor limit so development tools, container daemons, and databases operate without running out of file handles.",
        "install_phase": "config",
        "translator_capabilities": ["write-systemd-system-conf-drop-in", "write-systemd-user-conf-drop-in"],
        "sysctl_params": [{"key": "fs.file-max", "value": "524288", "drop_in_file": "omarchy-fd-limit.conf"}],
    }

    # OM-039
    data["OM-039"] = {
        "name": "Keyboard Layout Detection",
        "category": "config-dotfile",
        "intent": "Detect the keyboard layout and variant configured in the base system and propagate those settings into the compositor input configuration, so keyboard layout carries over from the install step without requiring manual re-entry.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-028"}],
        "translator_capabilities": ["read-system-keyboard-config", "write-compositor-keyboard-layout"],
        "script_payload": {"script": "detect-and-set-keyboard-layout", "capabilities": ["read-system-keyboard-config", "write-compositor-keyboard-layout"]},
    }

    # OM-040
    data["OM-040"] = {
        "name": "XCompose Emoji and Identity Shortcuts",
        "category": "config-dotfile",
        "intent": "Configure XCompose for fast emoji input and personal identity shortcuts using CapsLock as the compose key.",
        "install_phase": "config",
        "translator_capabilities": ["write-xcompose-config"],
        "file_assets": [{"src": "xcompose-template", "dst": "~/.XCompose"}],
    }

    # OM-041
    data["OM-041"] = {
        "name": "Mise Work Directory and Node.js Setup",
        "category": "arbitrary-script",
        "intent": "Set up a default work directory structure with a version manager configuration that adds project bin/ to PATH. Install Node.js via the version manager and configure it globally.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-009"}],
        "ordering": {"after": [{"id": "OM-009"}], "before": [{"id": "OM-023"}]},
        "translator_capabilities": ["configure-version-manager", "set-global-runtime-versions"],
        "script_payload": {"script": "setup-mise-work-and-nodejs", "capabilities": ["configure-version-manager", "set-global-runtime-versions"]},
    }

    # OM-042
    data["OM-042"] = {
        "name": "Fix Powerprofilesctl Python Shebang",
        "category": "arbitrary-script",
        "intent": "Fix the powerprofilesctl script to use the system Python interpreter instead of the version-manager-managed Python, preventing version conflicts.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-041"}],
        "ordering": {"after": [{"id": "OM-041"}]},
        "translator_capabilities": ["rewrite-shebang-line"],
        "script_payload": {"script": "fix-powerprofilesctl-shebang", "capabilities": ["rewrite-shebang-line"]},
    }

    # OM-043
    data["OM-043"] = {
        "name": "Docker Daemon Configuration",
        "category": "service-enable",
        "intent": "Configure the container daemon with log size limits, a fixed bridge IP, and DNS pointing to the host resolver. Expose the stub resolver to the container bridge network. Enable socket activation. Add the current user to the container access group. Prevent the daemon from blocking system boot.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-011"}],
        "ordering": {"after": [{"id": "OM-011"}]},
        "translator_capabilities": ["write-docker-daemon-config", "configure-systemd-resolved-drop-in", "enable-systemd-socket", "add-user-to-group", "write-systemd-service-drop-in"],
        "services": [{"name": "docker.socket", "enable": True}],
        "group_memberships": [{"group": "docker"}],
    }

    # OM-044
    data["OM-044"] = {
        "name": "System Default MIME Associations",
        "category": "mime-type",
        "intent": "Configure the system default applications for all major MIME types: file manager for directories, image viewer for images, document viewer for PDF, browser for web URLs, video player for all video formats, email client for mailto links, and editor for text and source files.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-008"}, {"id": "OM-012"}, {"id": "OM-015"}],
        "ordering": {"after": [{"id": "OM-021"}]},
        "translator_capabilities": ["register-mime-type-associations", "rebuild-desktop-application-database"],
        "mime_associations": [
            {"mime_pattern": "inode/directory", "app_id": "org.gnome.Nautilus.desktop"},
            {"mime_pattern": "image/*", "app_id": "imv.desktop"},
            {"mime_pattern": "application/pdf", "app_id": "org.gnome.Evince.desktop"},
            {"mime_pattern": "text/html", "app_id": "chromium.desktop"},
            {"mime_pattern": "x-scheme-handler/http", "app_id": "chromium.desktop"},
            {"mime_pattern": "x-scheme-handler/https", "app_id": "chromium.desktop"},
            {"mime_pattern": "video/*", "app_id": "mpv.desktop"},
            {"mime_pattern": "x-scheme-handler/mailto", "app_id": "hey.desktop"},
            {"mime_pattern": "text/plain", "app_id": "nvim.desktop"},
        ],
    }

    # OM-045
    data["OM-045"] = {
        "name": "XDG User Directories",
        "category": "config-dotfile",
        "intent": "Configure XDG user directories to point standard directories to home subdirectories, and remap unused directories to home itself to prevent them from appearing in the file manager sidebar.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-015"}],
        "translator_capabilities": ["invoke-xdg-user-dirs-update", "write-gtk-bookmarks-file"],
        "file_assets": [{"src": "user-dirs.dirs", "dst": "~/.config/user-dirs.dirs"}],
    }

    # OM-046
    data["OM-046"] = {
        "name": "Feature Toggle State Directory",
        "category": "config-dotfile",
        "intent": "Initialize the feature toggle state directory and create empty toggle configuration files for the three toggleable subsystems: screen lock, notification daemon, and application launcher styling.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-028"}],
        "translator_capabilities": ["create-state-files-for-feature-toggle-system"],
        "file_assets": [{"src": "toggles/", "dst": "~/.config/omarchy/toggles/"}],
    }

    # OM-047
    data["OM-047"] = {
        "name": "Nautilus Python Extensions",
        "category": "arbitrary-script",
        "intent": "Install Nautilus file manager Python extensions for local file sharing integration and video transcoding right-click action.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-015"}],
        "translator_capabilities": ["deploy-plugin-files-to-application-extension-directory"],
        "file_assets": [{"src": "nautilus-extensions/", "dst": "~/.local/share/nautilus/extensions/"}],
    }

    # OM-048
    data["OM-048"] = {
        "name": "System File Locate Database Update",
        "category": "arbitrary-script",
        "intent": "Update the system file-locate database so that the locate command can find all files installed during the packaging phase.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-014"}],
        "ordering": {"after": [{"id": "OM-017"}]},
        "translator_capabilities": ["run-updatedb"],
        "script_payload": {"script": "updatedb", "capabilities": ["run-updatedb"]},
    }

    # OM-049
    data["OM-049"] = {
        "name": "Walker Application Launcher Setup",
        "category": "arbitrary-script",
        "intent": "Configure the Walker application launcher as an autostart entry, set it up as a systemd user service with automatic restart on crash, create a package manager post-upgrade hook to restart Walker after updates, and link the Omarchy visual theme menu configuration files.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-015"}, {"id": "OM-028"}],
        "translator_capabilities": ["create-xdg-autostart-entries", "write-systemd-user-service-drop-ins", "create-package-manager-post-update-hooks", "create-config-symlinks"],
        "script_payload": {"script": "setup-walker-launcher", "capabilities": ["create-xdg-autostart-entries", "write-systemd-user-service-drop-ins", "create-config-symlinks"]},
    }

    # OM-050
    data["OM-050"] = {
        "name": "Fast Shutdown Configuration",
        "category": "config-dotfile",
        "intent": "Configure the service manager for fast shutdown: reduce the default timeout for stopping services and user sessions to allow quicker power-off and reboot operations.",
        "install_phase": "config",
        "translator_capabilities": ["deploy-systemd-system-conf-drop-in", "deploy-systemd-user-service-drop-in"],
        "file_assets": [{"src": "systemd/fast-shutdown.conf", "dst": "/etc/systemd/system.conf.d/fast-shutdown.conf"}],
    }

    # OM-051
    data["OM-051"] = {
        "name": "FUSE Unmount on Suspend Hook",
        "category": "arbitrary-script",
        "intent": "Install a system-sleep hook that unmounts FUSE filesystems before suspend, preventing filesystem corruption on systems with active FUSE mounts.",
        "install_phase": "config",
        "translator_capabilities": ["install-executable-sleep-hook-script"],
        "file_assets": [{"src": "sleep-hooks/unmount-fuse.sh", "dst": "/usr/lib/systemd/system-sleep/unmount-fuse"}],
    }

    # OM-052
    data["OM-052"] = {
        "name": "Brightness Control Sudo Access",
        "category": "config-dotfile",
        "intent": "Grant passwordless sudo access specifically for the display brightness-control tool, allowing the current user to control brightness without a password.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-014"}],
        "translator_capabilities": ["write-sudoers-nopasswd-drop-in"],
        "file_assets": [{"src": "sudoers-asdcontrol", "dst": "/etc/sudoers.d/asdcontrol"}],
    }

    # OM-053
    data["OM-053"] = {
        "name": "Input Group Membership",
        "category": "user-group",
        "intent": "Add the current user to the input group so that dictation tools and game controller applications can access input devices without requiring root.",
        "install_phase": "config",
        "translator_capabilities": ["add-user-to-system-group"],
        "group_memberships": [{"group": "input"}],
    }

    # OM-054
    data["OM-054"] = {
        "name": "Omarchy AI Skill Symlinks",
        "category": "arbitrary-script",
        "intent": "Deploy the Omarchy AI skill definition as a symbolic link into each AI coding assistant's skills directory, making the Omarchy system context available to all installed AI agents immediately.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-010"}, {"id": "OM-023"}],
        "ordering": {"after": [{"id": "OM-023"}]},
        "translator_capabilities": ["create-symlinks-across-multiple-agent-skill-directories"],
        "script_payload": {"script": "deploy-omarchy-ai-skill-symlinks", "capabilities": ["create-symlinks-across-multiple-agent-skill-directories"]},
    }

    # OM-055
    data["OM-055"] = {
        "name": "Pi Agent Omarchy Extension",
        "category": "arbitrary-script",
        "intent": "Install a custom Pi coding agent extension that provides Omarchy system theme awareness to the Pi agent.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-023"}],
        "ordering": {"after": [{"id": "OM-023"}]},
        "translator_capabilities": ["deploy-agent-extension-file"],
        "script_payload": {"script": "deploy-pi-omarchy-extension", "capabilities": ["deploy-agent-extension-file"]},
    }

    # OM-056
    data["OM-056"] = {
        "name": "Hyprland Feature Toggle Flags",
        "category": "config-dotfile",
        "intent": "Initialize the Hyprland feature toggle flags file, which controls which optional compositor features are enabled or disabled at runtime.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-028"}, {"id": "OM-046"}],
        "ordering": {"after": [{"id": "OM-046"}]},
        "translator_capabilities": ["copy-lua-config-file-to-user-state-directory"],
        "file_assets": [{"src": "omarchy-toggles.lua", "dst": "~/.local/state/omarchy/toggles.lua"}],
    }

    # OM-057
    data["OM-057"] = {
        "name": "Kernel Modules Cleanup Service",
        "category": "service-enable",
        "intent": "Enable the kernel modules cleanup service, which removes orphaned kernel module files after kernel upgrades.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-001"}],
        "translator_capabilities": ["enable-systemd-service-unit"],
        "services": [{"name": "kernel-modules-hook.service", "enable": True}],
    }

    # OM-058
    data["OM-058"] = {
        "name": "Automatic Power Profile Switching",
        "category": "config-dotfile",
        "intent": "Configure automatic power profile switching via rules on AC adapter plug/unplug events. Enable the power profiles daemon service. Only applied on systems with a battery.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-016"}],
        "hardware_condition": {"type": "leaf", "predicate": "hw-battery-present"},
        "translator_capabilities": ["write-udev-rules", "enable-systemd-service"],
        "services": [{"name": "power-profiles-daemon.service", "enable": True}],
        "file_assets": [{"src": "udev/power-profiles.rules", "dst": "/etc/udev/rules.d/80-power-profiles.rules"}],
    }

    # OM-059
    data["OM-059"] = {
        "name": "WiFi Power Save Mode Switching",
        "category": "config-dotfile",
        "intent": "Configure automatic WiFi power-save mode switching via rules: enable power-save on battery, disable it on AC power. Only applied on systems with a battery.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-014"}],
        "hardware_condition": {"type": "leaf", "predicate": "hw-battery-present"},
        "translator_capabilities": ["write-udev-rules"],
        "file_assets": [{"src": "udev/wifi-powersave.rules", "dst": "/etc/udev/rules.d/80-wifi-powersave.rules"}],
    }

    # OM-060
    data["OM-060"] = {
        "name": "Plocate AC-Only Update",
        "category": "config-dotfile",
        "intent": "Configure the file-locate database update service to only run when the system is on AC power, preventing battery drain from index rebuilds on portable hardware.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-014"}],
        "hardware_condition": {"type": "leaf", "predicate": "hw-battery-present"},
        "translator_capabilities": ["write-systemd-service-drop-in-with-ac-power-condition"],
        "file_assets": [{"src": "systemd/plocate-ac-only.conf", "dst": "/etc/systemd/system/plocate-updatedb.service.d/ac-only.conf"}],
    }

    # OM-061
    data["OM-061"] = {
        "name": "Wireless Networking Daemon",
        "category": "hardware-conditional",
        "intent": "Enable the wireless networking daemon and disable the systemd-networkd-wait-online service to prevent boot timeout on systems where network is managed by the wireless daemon rather than networkd.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-014"}],
        "translator_capabilities": ["enable-systemd-service-unit", "disable-and-mask-systemd-service-unit"],
        "services": [
            {"name": "iwd.service", "enable": True},
            {"name": "systemd-networkd-wait-online.service", "enable": False},
        ],
    }

    # OM-062
    data["OM-062"] = {
        "name": "Wireless Regulatory Domain",
        "category": "hardware-conditional",
        "intent": "Detect the system's wireless regulatory domain from the timezone configuration and set it in the wireless-regdom configuration file so WiFi operates at the correct power levels and channel set for the user's country.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-061"}, {"id": "OM-033"}],
        "ordering": {"after": [{"id": "OM-061"}]},
        "translator_capabilities": ["read-system-timezone", "derive-wireless-regulatory-domain", "write-regulatory-domain-config"],
        "script_payload": {"script": "set-wireless-regdom", "capabilities": ["read-system-timezone", "derive-wireless-regulatory-domain", "write-regulatory-domain-config"]},
    }

    # OM-063
    data["OM-063"] = {
        "name": "Apple-Style Function Key Mode",
        "category": "config-dotfile",
        "intent": "Configure Apple-style keyboards to treat the top-row function keys as standard F-keys by default rather than media keys.",
        "install_phase": "config",
        "translator_capabilities": ["write-kernel-module-option-config"],
        "file_assets": [{"src": "modprobe/hid_apple.conf", "dst": "/etc/modprobe.d/hid_apple.conf"}],
    }

    # OM-064
    data["OM-064"] = {
        "name": "Bluetooth Service and Audio Profile",
        "category": "service-enable",
        "intent": "Enable the Bluetooth service so that Bluetooth devices can be paired and connected. Configure Bluetooth to remember the last power state across reboots. Configure audio for automatic Bluetooth audio profile connection.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-014"}, {"id": "OM-028"}],
        "translator_capabilities": ["enable-systemd-service-unit", "modify-config-file-entry", "deploy-wireplumber-config-drop-in"],
        "services": [{"name": "bluetooth.service", "enable": True}],
    }

    # OM-065
    data["OM-065"] = {
        "name": "Printing Services and mDNS",
        "category": "service-enable",
        "intent": "Enable printing service, mDNS daemon for network printer discovery, and automatic remote printer detection. Configure DNS to not handle multicast DNS. Configure name resolution for .local domains via mDNS.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-016"}],
        "translator_capabilities": ["enable-multiple-systemd-service-units", "write-resolved-drop-in", "modify-nsswitch-conf"],
        "services": [
            {"name": "cups.service", "enable": True},
            {"name": "avahi-daemon.service", "enable": True},
            {"name": "cups-browsed.service", "enable": True},
        ],
    }

    # OM-066
    data["OM-066"] = {
        "name": "Disable USB Autosuspend",
        "category": "config-dotfile",
        "intent": "Disable USB autosuspend globally to prevent peripherals from disconnecting when the system suspends their USB port.",
        "install_phase": "config",
        "translator_capabilities": ["write-kernel-module-option-config-to-disable-usb-autosuspend"],
        "file_assets": [{"src": "modprobe/usbcore-autosuspend.conf", "dst": "/etc/modprobe.d/usbcore-autosuspend.conf"}],
    }

    # OM-067
    data["OM-067"] = {
        "name": "Ignore Hardware Power Button",
        "category": "config-dotfile",
        "intent": "Configure the system to ignore the hardware power button press, allowing the power button to be bound to a user-visible power menu in the desktop environment.",
        "install_phase": "config",
        "translator_capabilities": ["modify-logind-conf-handle-power-key"],
        "file_assets": [{"src": "logind/ignore-power-button.conf", "dst": "/etc/systemd/logind.conf.d/ignore-power-button.conf"}],
    }

    # OM-068
    data["OM-068"] = {
        "name": "NVIDIA GPU Drivers",
        "category": "hardware-conditional",
        "intent": "Detect NVIDIA GPUs and install the appropriate driver stack: open DKMS drivers with video acceleration support for newer GPUs, or legacy DKMS drivers for older GPU generations. Configure early kernel module loading and set compositor environment variables appropriate to the GPU generation.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-001"}, {"id": "OM-028"}],
        "hardware_condition": {"type": "leaf", "predicate": "hw-pci-vendor", "values": ["10de"]},
        "translator_capabilities": ["detect-gpu-model", "install-dkms-driver-packages", "write-modprobe-kernel-module-options", "write-mkinitcpio-config-drop-in", "append-compositor-environment-variable-declarations"],
        "packages": ["nvidia-open-dkms", "nvidia-utils", "libva-nvidia-driver"],
    }

    # OM-069
    data["OM-069"] = {
        "name": "Vulkan Graphics Drivers",
        "category": "hardware-conditional",
        "intent": "Install Vulkan graphics drivers matching detected GPU hardware: Intel, AMD, or Apple Silicon. Multiple drivers may be installed if multiple GPUs are present.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-001"}],
        "ordering": {"after": [{"id": "OM-068"}]},
        "hardware_condition": {"type": "or", "operands": [
            {"type": "leaf", "predicate": "hw-pci-vendor", "values": ["8086"]},
            {"type": "leaf", "predicate": "hw-pci-vendor", "values": ["1002"]},
            {"type": "leaf", "predicate": "hw-pci-vendor", "values": ["106b"]},
        ]},
        "translator_capabilities": ["detect-gpu-vendors-from-pci-device-list", "install-vendor-specific-vulkan-driver-packages"],
        "packages": ["vulkan-intel", "vulkan-radeon"],
    }

    # OM-070
    data["OM-070"] = {
        "name": "Intel Video Acceleration Drivers",
        "category": "hardware-conditional",
        "intent": "Install Intel hardware video acceleration drivers based on GPU generation.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-001"}],
        "hardware_condition": {"type": "leaf", "predicate": "hw-pci-vendor", "values": ["8086"]},
        "translator_capabilities": ["detect-intel-gpu-generation", "install-generation-appropriate-video-acceleration-packages"],
        "packages": ["intel-media-driver", "libvpl", "vpl-gpu-rt"],
    }

    # OM-071
    data["OM-071"] = {
        "name": "Intel Low Power Mode Daemon",
        "category": "hardware-conditional",
        "intent": "Install and enable Intel Low Power Mode Daemon for supported hybrid Intel CPUs on battery-equipped systems.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-001"}],
        "hardware_condition": {"type": "and", "operands": [
            {"type": "leaf", "predicate": "hw-cpu-vendor", "values": ["GenuineIntel"]},
            {"type": "leaf", "predicate": "hw-battery-present"},
        ]},
        "translator_capabilities": ["read-cpu-model-from-proc-cpuinfo", "install-and-enable-power-management-daemon"],
        "packages": ["intel-lpmd"],
        "services": [{"name": "intel_lpmd.service", "enable": True}],
    }

    # OM-072
    data["OM-072"] = {
        "name": "Intel Thermald",
        "category": "hardware-conditional",
        "intent": "Install and enable thermald for Intel laptops with Sandy Bridge or newer CPUs. Only applied on battery-equipped systems.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-001"}],
        "hardware_condition": {"type": "and", "operands": [
            {"type": "leaf", "predicate": "hw-cpu-vendor", "values": ["GenuineIntel"]},
            {"type": "leaf", "predicate": "hw-battery-present"},
        ]},
        "translator_capabilities": ["read-cpu-model-from-proc-cpuinfo", "install-and-enable-thermal-management-daemon"],
        "packages": ["thermald"],
        "services": [{"name": "thermald.service", "enable": True}],
    }

    # OM-073
    data["OM-073"] = {
        "name": "Intel IPU7 MIPI Camera Support",
        "category": "hardware-conditional",
        "intent": "Install MIPI camera support for Intel IPU7 hardware found in recent Intel Panther Lake laptops.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-001"}],
        "hardware_condition": {"type": "leaf", "predicate": "hw-acpi-hid", "match": "OVTI08F4"},
        "translator_capabilities": ["probe-acpi-device-ids", "install-camera-firmware-driver-packages"],
        "packages": ["ipu7-camera-firmware"],
    }

    # OM-074
    data["OM-074"] = {
        "name": "Panther Lake Kernel Installation",
        "category": "hardware-conditional",
        "intent": "Install the Panther Lake-specific kernel on Dell XPS Panther Lake systems, which includes audio driver patches not yet in mainline. Remove the standard kernel to avoid dual-kernel confusion.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-001"}],
        "ordering": {"before": [{"id": "OM-099"}]},
        "hardware_condition": {"type": "and", "operands": [
            {"type": "leaf", "predicate": "hw-dmi-product-name", "match": "XPS"},
            {"type": "leaf", "predicate": "hw-cpu-vendor", "values": ["GenuineIntel"]},
        ]},
        "translator_capabilities": ["install-alternative-kernel-package", "remove-default-kernel-package"],
        "packages": ["linux-ptl"],
        "remove_packages": ["linux"],
    }

    # OM-075
    data["OM-075"] = {
        "name": "Intel PTL FRED Kernel Parameter",
        "category": "kernel-boot-param",
        "intent": "Enable the FRED kernel feature on Intel Panther Lake systems by adding a kernel command-line parameter. FRED is a new interrupt-delivery mechanism that improves performance on Panther Lake.",
        "install_phase": "config",
        "ordering": {"before": [{"id": "OM-099"}]},
        "hardware_condition": {"type": "leaf", "predicate": "hw-cpu-vendor", "values": ["GenuineIntel"]},
        "translator_capabilities": ["write-bootloader-entry-tool-drop-in"],
        "kernel_params": [{"key": "fred", "value": "on"}],
    }

    # OM-076
    data["OM-076"] = {
        "name": "Intel BE200/BE211 WiFi 7 EHT Workaround",
        "category": "hardware-conditional",
        "intent": "Disable WiFi 7 on Intel BE200 and BE211 cards via a driver module option, working around a broken receive data path that causes the access point to fall back to unusable rates.",
        "install_phase": "config",
        "hardware_condition": {"type": "or", "operands": [
            {"type": "leaf", "predicate": "hw-pci-id", "values": ["8086:e440"]},
            {"type": "leaf", "predicate": "hw-pci-id", "values": ["8086:272b"]},
        ]},
        "translator_capabilities": ["write-kernel-module-option-config"],
        "file_assets": [{"src": "modprobe/fix-wifi7-eht.conf", "dst": "/etc/modprobe.d/iwlwifi-no-eht.conf"}],
    }

    # OM-077
    data["OM-077"] = {
        "name": "Intel PTL Sound Open Firmware",
        "category": "hardware-conditional",
        "intent": "Install Sound Open Firmware for the audio DSP on Intel Panther Lake systems that are not Dell XPS. Without this firmware, the DSP fails to initialize and only a null audio device appears.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-001"}],
        "hardware_condition": {"type": "and", "operands": [
            {"type": "leaf", "predicate": "hw-cpu-vendor", "values": ["GenuineIntel"]},
            {"type": "not", "operands": [{"type": "leaf", "predicate": "hw-dmi-product-name", "match": "XPS"}]},
        ]},
        "translator_capabilities": ["install-firmware-packages"],
        "packages": ["sof-firmware"],
    }

    # OM-078
    data["OM-078"] = {
        "name": "ASUS PTL Display Backlight Fix",
        "category": "kernel-boot-param",
        "intent": "Fix the display backlight on specific ASUS Panther Lake laptops by enabling DPCD AUX backlight mode in the kernel command line.",
        "install_phase": "config",
        "ordering": {"before": [{"id": "OM-099"}]},
        "hardware_condition": {"type": "or", "operands": [
            {"type": "leaf", "predicate": "hw-dmi-product-name", "match": "ExpertBook B9406"},
            {"type": "leaf", "predicate": "hw-dmi-product-name", "match": "Zenbook UX5406"},
        ]},
        "translator_capabilities": ["write-bootloader-entry-tool-drop-in"],
        "kernel_params": [{"key": "xe.enable_dpcd_backlight", "value": "1"}],
    }

    # OM-079
    data["OM-079"] = {
        "name": "ASUS B9406 Panel Replay Disable",
        "category": "kernel-boot-param",
        "intent": "Disable Panel Replay on ASUS ExpertBook B9406 to fix a display wake/update issue where the GPU's new Panel Replay feature has a broken wake path on this specific panel.",
        "install_phase": "config",
        "ordering": {"before": [{"id": "OM-099"}]},
        "hardware_condition": {"type": "leaf", "predicate": "hw-dmi-product-name", "match": "ExpertBook B9406"},
        "translator_capabilities": ["write-bootloader-entry-tool-drop-in"],
        "kernel_params": [{"key": "xe.enable_panel_replay", "value": "0"}],
    }

    # OM-080
    data["OM-080"] = {
        "name": "ASUS B9406 Touchpad Fix",
        "category": "hardware-conditional",
        "intent": "Work around the touchpad's touch jump detection discarding all motion events on ASUS ExpertBook B9406 by masking pressure axes via a libinput quirks file.",
        "install_phase": "config",
        "hardware_condition": {"type": "leaf", "predicate": "hw-dmi-product-name", "match": "ExpertBook B9406"},
        "translator_capabilities": ["write-libinput-quirks-override-file"],
        "file_assets": [{"src": "libinput/asus-b9406-touchpad.quirks", "dst": "/etc/libinput/local-overrides.quirks"}],
    }

    # OM-081
    data["OM-081"] = {
        "name": "ASUS ROG Audio Volume Fix",
        "category": "hardware-conditional",
        "intent": "Fix audio volume control on ASUS ROG laptops by enabling a soft mixer profile for the ALC285 codec. Also unmute the Master control which is often muted by default.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-028"}],
        "hardware_condition": {"type": "leaf", "predicate": "hw-dmi-chassis-vendor", "match": "ASUS"},
        "translator_capabilities": ["deploy-wireplumber-config-file", "invoke-alsa-mixer-controls"],
        "file_assets": [{"src": "wireplumber/asus-rog-audio.conf", "dst": "~/.config/wireplumber/wireplumber.conf.d/asus-rog-audio.conf"}],
    }

    # OM-082
    data["OM-082"] = {
        "name": "ASUS ROG Microphone Gain Fix",
        "category": "hardware-conditional",
        "intent": "Fix internal microphone gain on ASUS ROG laptops with ALC285 codec by setting the mic boost to zero and capture level to 70%, then storing the ALSA state persistently.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-081"}],
        "ordering": {"after": [{"id": "OM-081"}]},
        "hardware_condition": {"type": "leaf", "predicate": "hw-dmi-chassis-vendor", "match": "ASUS"},
        "translator_capabilities": ["invoke-alsa-mixer-controls", "store-alsa-state"],
        "script_payload": {"script": "fix-asus-rog-mic-gain", "capabilities": ["invoke-alsa-mixer-controls", "store-alsa-state"]},
    }

    # OM-083
    data["OM-083"] = {
        "name": "ASUS ROG Flow Z13 Touchpad Classification",
        "category": "hardware-conditional",
        "intent": "Mark the ASUS ROG Flow Z13 detachable keyboard touchpad as an integrated input device via a rule, so disable-while-typing correctly pairs it with the keyboard.",
        "install_phase": "config",
        "hardware_condition": {"type": "and", "operands": [
            {"type": "leaf", "predicate": "hw-dmi-chassis-vendor", "match": "ASUS"},
            {"type": "leaf", "predicate": "hw-dmi-product-name", "match": "GZ302"},
        ]},
        "translator_capabilities": ["write-udev-rules-file-to-set-device-integration-attributes"],
        "file_assets": [{"src": "udev/asus-z13-touchpad.rules", "dst": "/etc/udev/rules.d/70-asus-z13-touchpad.rules"}],
    }

    # OM-084
    data["OM-084"] = {
        "name": "Framework 13 AMD Audio Profile",
        "category": "hardware-conditional",
        "intent": "Set the audio card profile on Framework 13 AMD laptops to a specific profile that enables both microphone inputs and the speaker.",
        "install_phase": "config",
        "hardware_condition": {"type": "leaf", "predicate": "hw-audio-card-name", "match": "Family 17h/19h"},
        "translator_capabilities": ["invoke-pipewire-card-profile-selection"],
        "script_payload": {"script": "set-framework-13-amd-audio-profile", "capabilities": ["invoke-pipewire-card-profile-selection"]},
    }

    # OM-085
    data["OM-085"] = {
        "name": "Framework 16 Keyboard HID udev Rule",
        "category": "hardware-conditional",
        "intent": "Install a rule that grants unprivileged access to the Framework 16 keyboard HID device for RGB control. Without this rule, RGB control requires root.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-025"}, {"id": "OM-001"}],
        "hardware_condition": {"type": "leaf", "predicate": "hw-dmi-product-name", "match": "Framework 16"},
        "translator_capabilities": ["deploy-udev-rules-file"],
        "file_assets": [{"src": "udev/framework16-qmk-hid.rules", "dst": "/etc/udev/rules.d/70-framework16-qmk-hid.rules"}],
    }

    # OM-086
    data["OM-086"] = {
        "name": "MacBook SPI Keyboard Driver",
        "category": "hardware-conditional",
        "intent": "Install the SPI keyboard driver for MacBook models that use SPI instead of USB for keyboard communication. Configure the initramfs to load the required modules early.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-001"}],
        "hardware_condition": {"type": "leaf", "predicate": "hw-dmi-product-name", "match": "MacBook"},
        "translator_capabilities": ["read-dmi-product-name", "install-dkms-kernel-module-packages", "write-mkinitcpio-module-configuration"],
        "packages": ["macbook12-spi-driver-dkms"],
    }

    # OM-087
    data["OM-087"] = {
        "name": "MacBook NVMe Suspend Fix",
        "category": "hardware-conditional",
        "intent": "Fix NVMe drive wake-from-suspend failures on affected MacBook models by installing a service that disables D3cold power state for the NVMe controller.",
        "install_phase": "config",
        "hardware_condition": {"type": "leaf", "predicate": "hw-dmi-product-name", "match": "MacBook"},
        "translator_capabilities": ["read-dmi-product-name", "write-and-enable-systemd-service-unit"],
        "services": [{"name": "fix-macbook-nvme-d3cold.service", "enable": True}],
    }

    # OM-088
    data["OM-088"] = {
        "name": "Apple T2 Full Support Stack",
        "category": "hardware-conditional",
        "intent": "Install full Apple T2 chip support including T2-compatible kernel, audio firmware, WiFi/Bluetooth firmware, fan control daemon, and Touch Bar driver. Add user to video group for Touch Bar access. Enable T2 services. Configure kernel modules and parameters for T2 compatibility.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-001"}],
        "hardware_condition": {"type": "leaf", "predicate": "hw-pci-id", "values": ["106b:1801", "106b:1802"]},
        "translator_capabilities": ["detect-t2-chip-by-pci-id", "install-t2-specific-kernel-and-hardware-packages", "enable-multiple-systemd-services", "write-module-load-configurations", "write-mkinitcpio-module-list", "write-bootloader-kernel-parameters"],
        "packages": ["linux-t2", "apple-t2-audio-config", "apple-bcm-firmware", "t2fanrd", "tiny-dfr"],
        "services": [{"name": "t2fanrd.service", "enable": True}, {"name": "tiny-dfr.service", "enable": True}],
        "group_memberships": [{"group": "video"}],
        "kernel_params": [{"key": "intel_iommu", "value": "on"}, {"key": "iommu", "value": "pt"}, {"key": "pcie_ports", "value": "compat"}],
    }

    # OM-089
    data["OM-089"] = {
        "name": "Lenovo Yoga Pro 7 Bass Speaker Fix",
        "category": "hardware-conditional",
        "intent": "Fix the bass speaker output on Lenovo Yoga Pro 7 14IAH10 by writing a kernel module option that applies a pin quirk to route audio to both amplifier channels on the ALC287 codec.",
        "install_phase": "config",
        "hardware_condition": {"type": "leaf", "predicate": "hw-dmi-product-name", "match": "Yoga Pro 7 14IAH10"},
        "translator_capabilities": ["write-kernel-module-option-config-for-sound-driver-pin-quirk"],
        "file_assets": [{"src": "modprobe/yoga-pro7-bass.conf", "dst": "/etc/modprobe.d/yoga-pro7-bass.conf"}],
    }

    # OM-090
    data["OM-090"] = {
        "name": "Broadcom WiFi DKMS Drivers",
        "category": "hardware-conditional",
        "intent": "Install Broadcom WiFi DKMS drivers for MacBook and other systems with BCM4360 or BCM4331 chipsets.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-001"}],
        "hardware_condition": {"type": "or", "operands": [
            {"type": "leaf", "predicate": "hw-pci-id", "values": ["14e4:43a0"]},
            {"type": "leaf", "predicate": "hw-pci-id", "values": ["14e4:4331"]},
        ]},
        "translator_capabilities": ["read-pci-device-ids", "install-dkms-wifi-driver-packages"],
        "packages": ["broadcom-wl-dkms"],
    }

    # OM-091
    data["OM-091"] = {
        "name": "Microsoft Surface Keyboard Modules",
        "category": "hardware-conditional",
        "intent": "Configure the initramfs to include the Surface-specific kernel modules required for keyboard input on Microsoft Surface devices.",
        "install_phase": "config",
        "hardware_condition": {"type": "leaf", "predicate": "hw-dmi-sys-vendor", "match": "Microsoft"},
        "translator_capabilities": ["detect-loaded-kernel-modules", "write-mkinitcpio-module-list-configuration"],
        "file_assets": [{"src": "mkinitcpio/surface-keyboard.conf", "dst": "/etc/mkinitcpio.conf.d/surface-keyboard.conf"}],
    }

    # OM-092
    data["OM-092"] = {
        "name": "Motorcomm YT6801 Ethernet Driver",
        "category": "hardware-conditional",
        "intent": "Install DKMS driver for the Motorcomm YT6801 Ethernet adapter used in the Slimbook Executive laptop.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-001"}],
        "hardware_condition": {"type": "leaf", "predicate": "hw-pci-vendor", "values": ["0x1d17"]},
        "translator_capabilities": ["detect-pci-device-by-name-pattern", "install-dkms-driver-packages"],
        "packages": ["yt6801-dkms"],
    }

    # OM-093
    data["OM-093"] = {
        "name": "Synaptics InterTouch Touchpad",
        "category": "hardware-conditional",
        "intent": "Enable Synaptics InterTouch for touchpad devices that present as Synaptics in the input device list but whose driver is not yet loaded, providing improved touchpad responsiveness.",
        "install_phase": "config",
        "hardware_condition": {"type": "leaf", "predicate": "hw-input-device-name", "match": "synaptics"},
        "translator_capabilities": ["detect-touchpad-type-from-kernel-input-devices", "conditionally-load-kernel-module-with-options"],
        "script_payload": {"script": "enable-synaptics-intertouch", "capabilities": ["conditionally-load-kernel-module-with-options"]},
    }

    # OM-094
    data["OM-094"] = {
        "name": "Tuxedo/Clevo Keyboard Backlight Drivers",
        "category": "hardware-conditional",
        "intent": "Install Tuxedo/Clevo keyboard backlight DKMS drivers for Tuxedo laptops and Slimbook Executive. Blacklist the conflicting legacy module. Remove any orphaned legacy module files.",
        "install_phase": "config",
        "depends_on": [{"id": "OM-001"}],
        "hardware_condition": {"type": "or", "operands": [
            {"type": "leaf", "predicate": "hw-dmi-sys-vendor", "match": "TUXEDO"},
            {"type": "leaf", "predicate": "hw-dmi-sys-vendor", "match": "Slimbook"},
        ]},
        "translator_capabilities": ["detect-system-vendor-from-dmi", "install-dkms-driver-packages", "write-kernel-module-blacklist-configuration"],
        "packages": ["tuxedo-drivers-nocompatcheck-dkms"],
        "file_assets": [{"src": "modprobe/clevo-blacklist.conf", "dst": "/etc/modprobe.d/clevo-blacklist.conf"}],
    }

    # OM-095
    data["OM-095"] = {
        "name": "Plymouth Boot Splash Theme",
        "category": "theming",
        "intent": "Install the Omarchy custom Plymouth boot splash theme and set it as the active default theme so the Omarchy branded animation appears during boot.",
        "install_phase": "login",
        "depends_on": [{"id": "OM-016"}],
        "ordering": {"before": [{"id": "OM-099"}]},
        "translator_capabilities": ["copy-plymouth-theme-directory", "set-default-plymouth-theme"],
        "file_assets": [{"src": "plymouth-themes/omarchy/", "dst": "/usr/share/plymouth/themes/omarchy/"}],
    }

    # OM-096
    data["OM-096"] = {
        "name": "Passwordless GNOME Keyring",
        "category": "config-dotfile",
        "intent": "Create a pre-configured GNOME keyring file that is unlocked without a password for use with auto-login setups. This prevents prompts for keyring unlock that would appear at desktop startup.",
        "install_phase": "login",
        "depends_on": [{"id": "OM-016"}],
        "translator_capabilities": ["write-keyring-descriptor-files-with-correct-permissions"],
        "file_assets": [{"src": "gnome-keyrings/default.keyring", "dst": "~/.local/share/keyrings/default.keyring"}],
    }

    # OM-097
    data["OM-097"] = {
        "name": "SDDM Display Manager",
        "category": "display-manager",
        "intent": "Install the Omarchy SDDM display manager theme, configure SDDM for Wayland operation with a custom compositor command, enable auto-login for the current user into the Omarchy session. Disable password-based SDDM logins from creating an encrypted login keyring. Enable the SDDM service.",
        "install_phase": "login",
        "depends_on": [{"id": "OM-015"}, {"id": "OM-096"}],
        "ordering": {"after": [{"id": "OM-096"}]},
        "translator_capabilities": ["deploy-sddm-theme", "write-sddm-config-drop-ins", "modify-pam-sddm-config", "enable-display-manager-systemd-service"],
        "display_manager": {"name": "sddm", "greeter": "omarchy", "auto_login": True},
        "services": [{"name": "sddm.service", "enable": True}],
    }

    # OM-098
    data["OM-098"] = {
        "name": "Hibernation Support",
        "category": "arbitrary-script",
        "intent": "Set up hibernation support: configure the swap device, set resume kernel parameters, and install the initramfs resume hook. The initramfs rebuild is deferred to the bootloader configuration step.",
        "install_phase": "login",
        "ordering": {"before": [{"id": "OM-099"}]},
        "translator_capabilities": ["configure-swap-resume-device", "add-kernel-resume-parameters", "install-initramfs-hooks"],
        "script_payload": {"script": "setup-hibernation", "capabilities": ["configure-swap-resume-device", "add-kernel-resume-parameters", "install-initramfs-hooks"]},
    }

    # OM-099
    data["OM-099"] = {
        "name": "Limine Bootloader Configuration",
        "category": "bootloader-config",
        "intent": "Configure the Limine bootloader with mkinitcpio hooks including Plymouth and btrfs-overlayfs. Set up unified kernel image generation. Configure the btrfs snapshot boot menu. Set up the Thunderbolt module. Create a root snapper configuration and disable btrfs quotas for performance. Enable the snapshot sync service. Restore initramfs hooks disabled earlier. Trigger a full unified kernel image rebuild. Register the Limine EFI boot entry.",
        "install_phase": "login",
        "depends_on": [{"id": "OM-002"}, {"id": "OM-095"}, {"id": "OM-098"}],
        "ordering": {"after": [{"id": "OM-095"}, {"id": "OM-098"}]},
        "translator_capabilities": ["configure-mkinitcpio-hooks-and-modules", "manage-limine-bootloader-installation", "create-snapper-configuration", "disable-btrfs-quota-accounting", "enable-systemd-service", "manage-efi-boot-entries"],
        "bootloader": {"name": "limine", "snapshot": True},
        "packages": ["limine", "limine-mkinitcpio-hook", "limine-snapper-sync", "snapper"],
        "services": [{"name": "limine-snapper-sync.service", "enable": True}],
    }

    # OM-100
    data["OM-100"] = {
        "name": "Final Package Manager Configuration",
        "category": "custom-repo",
        "intent": "Restore the final package manager configuration for the target mirror channel after the install is complete. For Apple T2 systems, also append the arch-mact2 repository to the package manager configuration, enabling kernel updates.",
        "install_phase": "post-install",
        "depends_on": [{"id": "OM-001"}, {"id": "OM-088"}],
        "translator_capabilities": ["deploy-final-package-manager-config", "conditionally-append-unsigned-repository-for-t2-hardware"],
        "custom_repos": [{"name": "arch-mact2", "url": "https://pkgs.t2linux.org/archlinux/x86_64/", "sig_level": "Never"}],
    }

    # OM-101: first-run
    data["OM-101"] = {
        "name": "First-Run Network Check and Update Prompt",
        "category": "service-enable",
        "intent": "Check for network connectivity on first login. If offline, send a desktop notification prompting the user to set up wireless networking. Whether online or offline, send a notification prompting the user to run a system update.",
        "install_phase": "first-run",
        "execution_phase": "first-run",
        "depends_on": [{"id": "OM-061"}, {"id": "OM-015"}],
        "translator_capabilities": ["send-desktop-notifications", "detect-network-connectivity"],
        "script_payload": {"script": "first-run-network-check", "capabilities": ["send-desktop-notifications", "detect-network-connectivity"]},
    }

    # OM-102: first-run
    data["OM-102"] = {
        "name": "GTK Dark Mode and Icon Theme",
        "category": "theming",
        "intent": "Apply GTK theme settings for dark mode and an appropriate icon theme via desktop settings, and update the icon cache. Requires a running display session and settings backend.",
        "install_phase": "first-run",
        "execution_phase": "first-run",
        "depends_on": [{"id": "OM-015"}, {"id": "OM-029"}],
        "translator_capabilities": ["invoke-gsettings-to-set-desktop-theme-and-icon-theme", "invoke-gtk-update-icon-cache"],
        "script_payload": {"script": "apply-gtk-dark-mode", "capabilities": ["invoke-gsettings-to-set-desktop-theme-and-icon-theme", "invoke-gtk-update-icon-cache"]},
    }

    # OM-103: first-run
    data["OM-103"] = {
        "name": "Firewall Configuration",
        "category": "firewall-rule",
        "intent": "Configure the system firewall with a default deny-inbound policy, allow all outbound traffic, allow local file sharing peer-to-peer port, allow container DNS, and enable the firewall service on boot.",
        "install_phase": "first-run",
        "execution_phase": "first-run",
        "depends_on": [{"id": "OM-016"}],
        "translator_capabilities": ["invoke-ufw-to-configure-firewall-rules", "invoke-ufw-docker-for-container-rules", "enable-ufw-systemd-service"],
        "script_payload": {"script": "configure-firewall", "capabilities": ["invoke-ufw-to-configure-firewall-rules", "invoke-ufw-docker-for-container-rules", "enable-ufw-systemd-service"]},
    }

    # OM-104: first-run
    data["OM-104"] = {
        "name": "Systemd-Resolved DNS Stub",
        "category": "dns-config",
        "intent": "Configure the system DNS resolver to use systemd-resolved's stub listener, enabling DNSSEC validation and the resolved stub resolver for all DNS queries.",
        "install_phase": "first-run",
        "execution_phase": "first-run",
        "translator_capabilities": ["create-symlink-from-resolv-conf-to-systemd-resolved-stub"],
        "script_payload": {"script": "setup-dns-stub-resolver", "capabilities": ["create-symlink-from-resolv-conf-to-systemd-resolved-stub"]},
    }

    # OM-105: first-run
    data["OM-105"] = {
        "name": "HiDPI GDK Scale Detection",
        "category": "config-dotfile",
        "intent": "Detect the current monitor's reported scale factor from the compositor and write it as the default display scale environment variable in the monitors configuration, ensuring applications render at the correct scale on HiDPI displays.",
        "install_phase": "first-run",
        "execution_phase": "first-run",
        "depends_on": [{"id": "OM-028"}, {"id": "OM-006"}],
        "translator_capabilities": ["query-compositor-for-monitor-scale", "write-scale-factor-to-compositor-environment-config"],
        "script_payload": {"script": "detect-and-set-gdk-scale", "capabilities": ["query-compositor-for-monitor-scale", "write-scale-factor-to-compositor-environment-config"]},
    }

    # OM-106: first-run
    data["OM-106"] = {
        "name": "Battery Monitor or Performance Mode",
        "category": "service-enable",
        "intent": "Enable the battery monitoring timer service on systems with a battery to provide low-battery desktop notifications. On systems without a battery, set the power profile to performance mode as the permanent default.",
        "install_phase": "first-run",
        "execution_phase": "first-run",
        "depends_on": [{"id": "OM-016"}, {"id": "OM-058"}],
        "hardware_condition": {"type": "leaf", "predicate": "hw-battery-present"},
        "translator_capabilities": ["enable-systemd-user-timer-unit", "invoke-powerprofilesctl"],
        "services": [{"name": "omarchy-battery-monitor.timer", "enable": True, "deferred": True}],
    }

    # OM-107: first-run
    data["OM-107"] = {
        "name": "Cleanup First-Run Sudoers Rule",
        "category": "arbitrary-script",
        "intent": "Remove the temporary first-run sudoers rule file that granted passwordless sudo access to the set of commands needed during first-run configuration, restoring normal sudo behavior after all first-run scripts have completed.",
        "install_phase": "first-run",
        "execution_phase": "first-run",
        "depends_on": [{"id": "OM-004"}],
        "ordering": {"after": [{"id": "OM-103"}, {"id": "OM-104"}]},
        "translator_capabilities": ["remove-specific-sudoers-drop-in-file"],
        "script_payload": {"script": "cleanup-first-run-sudoers", "capabilities": ["remove-specific-sudoers-drop-in-file"]},
    }

    # OM-108: first-run
    data["OM-108"] = {
        "name": "Elephant Application Launcher Service",
        "category": "service-enable",
        "intent": "Enable and start the Elephant application launcher as a user service so it is available immediately after first login and on all subsequent logins.",
        "install_phase": "first-run",
        "execution_phase": "first-run",
        "depends_on": [{"id": "OM-049"}],
        "translator_capabilities": ["enable-user-scoped-systemd-service-and-start-immediately"],
        "services": [{"name": "walker.service", "enable": True, "deferred": True}],
    }

    # OM-109: first-run
    data["OM-109"] = {
        "name": "GTK Primary Paste",
        "category": "config-dotfile",
        "intent": "Enable GTK primary paste (middle-click paste from selection clipboard) as a system-wide GTK preference.",
        "install_phase": "first-run",
        "execution_phase": "first-run",
        "translator_capabilities": ["invoke-gsettings-to-set-gtk-interface-preference"],
        "script_payload": {"script": "enable-gtk-primary-paste", "capabilities": ["invoke-gsettings-to-set-gtk-interface-preference"]},
    }

    # OM-110: first-run
    data["OM-110"] = {
        "name": "Internal Monitor Recovery Service",
        "category": "service-enable",
        "intent": "Enable the internal monitor recovery user service which automatically recovers the internal monitor if it becomes unavailable after an external display is disconnected.",
        "install_phase": "first-run",
        "execution_phase": "first-run",
        "depends_on": [{"id": "OM-006"}],
        "translator_capabilities": ["enable-user-scoped-systemd-service-unit"],
        "services": [{"name": "omarchy-recover-internal-monitor.service", "enable": True, "deferred": True}],
    }

    # OM-111: first-run
    data["OM-111"] = {
        "name": "SwayOSD On-Screen Display Service",
        "category": "service-enable",
        "intent": "Start and enable the SwayOSD on-screen display server service, which provides visual feedback overlays for volume, brightness, and caps-lock changes.",
        "install_phase": "first-run",
        "execution_phase": "first-run",
        "depends_on": [{"id": "OM-015"}],
        "translator_capabilities": ["reload-systemd-user-daemon", "enable-and-start-user-scoped-systemd-service"],
        "services": [{"name": "swayosd-server.service", "enable": True, "deferred": True}],
    }

    # OM-112: first-run
    data["OM-112"] = {
        "name": "Welcome Notification",
        "category": "arbitrary-script",
        "intent": "Send a desktop notification welcoming the user and explaining the key keyboard shortcuts for the keybinding cheatsheet, application launcher, and system menu.",
        "install_phase": "first-run",
        "execution_phase": "first-run",
        "depends_on": [{"id": "OM-015"}],
        "translator_capabilities": ["send-desktop-notifications-with-structured-text"],
        "script_payload": {"script": "send-welcome-notification", "capabilities": ["send-desktop-notifications-with-structured-text"]},
    }

    # OM-113: first-run
    data["OM-113"] = {
        "name": "Voxtype Install Offer Notification",
        "category": "arbitrary-script",
        "intent": "Send a desktop notification offering to install Voxtype voice dictation. The hook script removes itself after running. This is a deferred optional feature prompt, not an automatic install.",
        "install_phase": "first-run",
        "execution_phase": "first-run",
        "depends_on": [{"id": "OM-015"}],
        "translator_capabilities": ["send-desktop-notification", "self-delete-hook-script-after-execution"],
        "script_payload": {"script": "offer-voxtype-install", "capabilities": ["send-desktop-notification", "self-delete-hook-script-after-execution"]},
    }

    # Themes OM-114..OM-134
    themes = [
        ("OM-114", "Catppuccin Dark Theme", "catppuccin", "Provide the Catppuccin dark (Mocha variant) visual theme bundle including desktop backgrounds, color palette definitions, icon theme link, editor color scheme, lock screen image, and preview images."),
        ("OM-115", "Catppuccin Latte Theme", "catppuccin-latte", "Provide the Catppuccin Latte (light variant) visual theme bundle with a light color palette and the same structure as the Catppuccin dark bundle."),
        ("OM-116", "Ethereal Theme", "ethereal", "Provide the Ethereal visual theme bundle."),
        ("OM-117", "Everforest Theme", "everforest", "Provide the Everforest visual theme bundle."),
        ("OM-118", "Flexoki Light Theme", "flexoki-light", "Provide the Flexoki Light visual theme bundle."),
        ("OM-119", "Gruvbox Theme", "gruvbox", "Provide the Gruvbox visual theme bundle."),
        ("OM-120", "Hackerman Theme", "hackerman", "Provide the Hackerman dark green terminal-aesthetic visual theme bundle."),
        ("OM-121", "Kanagawa Theme", "kanagawa", "Provide the Kanagawa visual theme bundle."),
        ("OM-122", "Last Horizon Theme", "last-horizon", "Provide the Last Horizon visual theme bundle."),
        ("OM-123", "Lumon Theme", "lumon", "Provide the Lumon Severance-inspired visual theme bundle."),
        ("OM-124", "Matte Black Theme", "matte-black", "Provide the Matte Black visual theme bundle."),
        ("OM-125", "Miasma Theme", "miasma", "Provide the Miasma visual theme bundle."),
        ("OM-126", "Nord Theme", "nord", "Provide the Nord visual theme bundle."),
        ("OM-127", "Osaka Jade Theme", "osaka-jade", "Provide the Osaka Jade visual theme bundle."),
        ("OM-128", "Retro 82 Theme", "retro-82", "Provide the Retro 82 visual theme bundle."),
        ("OM-129", "Ristretto Theme", "ristretto", "Provide the Ristretto warm coffee palette visual theme bundle."),
        ("OM-130", "Rose Pine Theme", "rose-pine", "Provide the Rose Pine visual theme bundle."),
        ("OM-131", "Solitude Theme", "solitude", "Provide the Solitude visual theme bundle."),
        ("OM-132", "Tokyo Night Theme", "tokyo-night", "Provide the Tokyo Night visual theme bundle. This is the default theme activated during install."),
        ("OM-133", "Vantablack Theme", "vantablack", "Provide the Vantablack maximum contrast dark visual theme bundle."),
        ("OM-134", "White Theme", "white", "Provide the White light visual theme bundle."),
    ]
    for om_id, name, bundle, intent in themes:
        entry = {
            "name": name,
            "category": "theming",
            "intent": intent,
            "install_phase": "first-run",
            "depends_on": [{"id": "OM-029"}],
            "translator_capabilities": ["deploy-file-asset-payloads-to-theme-directories"],
            "theme": {"bundle_dir": f"themes/{bundle}"},
        }
        if om_id == "OM-132":
            entry["theme"]["is_default"] = True
        data[om_id] = entry

    return data


# ── Point data (from research/omarchy-points.md) ────────────────────────────
def point_data() -> list:
    """Return the 32 points as a list of dicts."""
    return [
        {
            "id": "omarchy/repository-bootstrap-and-preflight",
            "name": "Repository Bootstrap and Preflight",
            "intent": "Establish the package manager configuration, apply any pending migrations, and set the install-time execution context before any packages are installed.",
            "members": [
                {"id": "OM-001", "status": "required"},
                {"id": "OM-002", "status": "required"},
                {"id": "OM-003", "status": "required"},
                {"id": "OM-004", "status": "required"},
                {"id": "OM-005", "status": "required"},
            ],
        },
        {
            "id": "omarchy/hyprland-desktop-stack",
            "name": "Hyprland Desktop Stack",
            "intent": "Install and configure the complete Wayland compositor session stack including the compositor, session launcher, display portal, idle daemon, screen lock, color temperature control, status bar, notification daemon, application launcher, background setter, on-screen display, polkit agent, and the display manager with auto-login.",
            "members": [
                {"id": "OM-006", "status": "required"},
                {"id": "OM-015", "status": "required"},
                {"id": "OM-039", "status": "required"},
                {"id": "OM-040", "status": "required"},
                {"id": "OM-046", "status": "required"},
                {"id": "OM-049", "status": "required"},
                {"id": "OM-056", "status": "nice-to-have"},
            ],
        },
        {
            "id": "omarchy/terminal-and-shell-toolchain",
            "name": "Terminal and Shell Toolchain",
            "intent": "Install and configure the primary terminal emulator, multiplexer, and shell enhancement tools that provide the daily command-line working environment.",
            "members": [
                {"id": "OM-007", "status": "required"},
            ],
        },
        {
            "id": "omarchy/browser-and-web-access",
            "name": "Browser and Web Access",
            "intent": "Install the primary web browser and its supporting libraries.",
            "members": [
                {"id": "OM-008", "status": "required"},
            ],
        },
        {
            "id": "omarchy/developer-toolchain",
            "name": "Developer Toolchain",
            "intent": "Install the core development tooling: language runtimes and version manager, build tools, database libraries, the editor with curated plugin configuration, version control tooling, and the work-directory setup.",
            "members": [
                {"id": "OM-009", "status": "required"},
                {"id": "OM-017", "status": "required"},
                {"id": "OM-019", "status": "required"},
                {"id": "OM-041", "status": "required"},
                {"id": "OM-042", "status": "required"},
            ],
        },
        {
            "id": "omarchy/ai-tooling",
            "name": "AI Tooling",
            "intent": "Install and configure all AI coding assistant and agent tooling, including native packages, runtime global installs, agent skill links, and the Pi agent extension.",
            "members": [
                {"id": "OM-010", "status": "required"},
                {"id": "OM-023", "status": "nice-to-have"},
                {"id": "OM-054", "status": "nice-to-have"},
                {"id": "OM-055", "status": "nice-to-have"},
            ],
        },
        {
            "id": "omarchy/containerization-and-virtualization",
            "name": "Containerization and Virtualization",
            "intent": "Install container and virtualization tooling and configure the container daemon with production-ready settings including log limits, DNS, network configuration, and service activation.",
            "members": [
                {"id": "OM-011", "status": "required"},
                {"id": "OM-043", "status": "required"},
            ],
        },
        {
            "id": "omarchy/media-and-creative-applications",
            "name": "Media and Creative Applications",
            "intent": "Install the full suite of media consumption and creation tools: video player, audio tools, screen recorder, image viewer, image editor, PDF viewer, screenshot tools, and screen annotation.",
            "members": [
                {"id": "OM-012", "status": "required"},
            ],
        },
        {
            "id": "omarchy/productivity-and-communication",
            "name": "Productivity and Communication Applications",
            "intent": "Install personal productivity and communication applications: note-taking, secure messaging, office suite, markdown editor, audio streaming, and local file sharing.",
            "members": [
                {"id": "OM-013", "status": "required"},
            ],
        },
        {
            "id": "omarchy/system-utilities-and-hardware-tools",
            "name": "System Utilities and Hardware Tools",
            "intent": "Install system utilities covering Bluetooth management, disk tools, file manager and extensions, display brightness control, system monitoring, file search, and network tools.",
            "members": [
                {"id": "OM-014", "status": "required"},
                {"id": "OM-047", "status": "required"},
                {"id": "OM-053", "status": "required"},
            ],
        },
        {
            "id": "omarchy/security-and-network-services",
            "name": "Security and Network Services",
            "intent": "Install and configure system security tools including firewall, printing support, network service discovery, cryptographic keyring, and power management daemon.",
            "members": [
                {"id": "OM-016", "status": "required"},
            ],
        },
        {
            "id": "omarchy/fonts-and-visual-assets",
            "name": "Fonts and Visual Assets",
            "intent": "Install and register fonts used by the Omarchy desktop environment.",
            "members": [
                {"id": "OM-018", "status": "required"},
            ],
        },
        {
            "id": "omarchy/web-applications-and-launcher-entries",
            "name": "Web Applications and Launcher Entries",
            "intent": "Install web applications as Chromium PWAs and configure TUI applications as windowed launcher entries with custom icons and compositor layout hints.",
            "members": [
                {"id": "OM-020", "status": "nice-to-have"},
                {"id": "OM-021", "status": "nice-to-have"},
                {"id": "OM-022", "status": "nice-to-have"},
            ],
        },
        {
            "id": "omarchy/application-configuration-and-defaults",
            "name": "Application Configuration and Defaults",
            "intent": "Deploy the baseline configuration file tree, configure MIME type associations, set XDG user directories, configure GPG, write sudoers rules, initialize locate database, and establish default user preferences.",
            "members": [
                {"id": "OM-028", "status": "required"},
                {"id": "OM-030", "status": "nice-to-have"},
                {"id": "OM-031", "status": "required"},
                {"id": "OM-032", "status": "required"},
                {"id": "OM-033", "status": "required"},
                {"id": "OM-034", "status": "required"},
                {"id": "OM-035", "status": "required"},
                {"id": "OM-044", "status": "required"},
                {"id": "OM-045", "status": "required"},
                {"id": "OM-048", "status": "required"},
            ],
        },
        {
            "id": "omarchy/system-performance-tuning",
            "name": "System Performance Tuning",
            "intent": "Tune system kernel and resource parameters for a developer-oriented desktop: inotify watchers, file descriptor limits, TCP MTU probing, and fast shutdown timeouts.",
            "members": [
                {"id": "OM-036", "status": "required"},
                {"id": "OM-037", "status": "required"},
                {"id": "OM-038", "status": "required"},
                {"id": "OM-050", "status": "required"},
            ],
        },
        {
            "id": "omarchy/system-infrastructure-and-hooks",
            "name": "System Infrastructure and Hooks",
            "intent": "Enable and configure supporting system services and infrastructure hooks that underpin the desktop: kernel modules cleanup, FUSE unmount hook, and brightness control sudo access.",
            "members": [
                {"id": "OM-051", "status": "required"},
                {"id": "OM-052", "status": "required"},
                {"id": "OM-057", "status": "required"},
            ],
        },
        {
            "id": "omarchy/networking-and-connectivity-configuration",
            "name": "Networking and Connectivity Configuration",
            "intent": "Enable and configure network management, wireless regulatory domain, Bluetooth, printing services, and DNS resolution.",
            "members": [
                {"id": "OM-061", "status": "required"},
                {"id": "OM-062", "status": "required"},
                {"id": "OM-064", "status": "required"},
                {"id": "OM-065", "status": "required"},
            ],
        },
        {
            "id": "omarchy/hardware-peripheral-configuration",
            "name": "Hardware Peripheral Configuration",
            "intent": "Configure input and peripheral hardware behavior applicable to many systems: Apple-style function key mode, USB autosuspend policy, power button behavior, and ASUS ROG laptop audio fixes.",
            "members": [
                {"id": "OM-063", "status": "required"},
                {"id": "OM-066", "status": "required"},
                {"id": "OM-067", "status": "required"},
                {"id": "OM-081", "status": "required"},
                {"id": "OM-082", "status": "required"},
            ],
        },
        {
            "id": "omarchy/gpu-and-display-drivers",
            "name": "GPU and Display Drivers",
            "intent": "Detect and install GPU drivers and video acceleration packages for NVIDIA, AMD, Intel, and Apple Silicon GPUs, including Vulkan drivers.",
            "members": [
                {"id": "OM-068", "status": "required"},
                {"id": "OM-069", "status": "required"},
                {"id": "OM-070", "status": "required"},
            ],
        },
        {
            "id": "omarchy/intel-hardware-optimizations",
            "name": "Intel Hardware Optimizations",
            "intent": "Configure Intel-specific power management, thermal control, camera support, kernel selection, and hardware workarounds for supported Intel CPU and chipset generations.",
            "members": [
                {"id": "OM-071", "status": "required"},
                {"id": "OM-072", "status": "required"},
                {"id": "OM-073", "status": "required"},
                {"id": "OM-074", "status": "required"},
                {"id": "OM-075", "status": "required"},
                {"id": "OM-076", "status": "required"},
                {"id": "OM-077", "status": "required"},
            ],
        },
        {
            "id": "omarchy/asus-hardware-support",
            "name": "ASUS Hardware Support",
            "intent": "Install ASUS-specific hardware control daemons and fix device-specific hardware issues for ASUS ROG and ExpertBook/Zenbook variants.",
            "members": [
                {"id": "OM-024", "status": "required"},
                {"id": "OM-078", "status": "required"},
                {"id": "OM-079", "status": "required"},
                {"id": "OM-080", "status": "required"},
                {"id": "OM-083", "status": "required"},
            ],
        },
        {
            "id": "omarchy/framework-hardware-support",
            "name": "Framework Hardware Support",
            "intent": "Install and configure Framework-specific hardware: QMK HID tool for Framework 16 RGB control, audio profile correction for Framework 13 AMD, and udev access rules.",
            "members": [
                {"id": "OM-025", "status": "required"},
                {"id": "OM-084", "status": "required"},
                {"id": "OM-085", "status": "required"},
            ],
        },
        {
            "id": "omarchy/apple-hardware-support",
            "name": "Apple Hardware Support",
            "intent": "Install and configure Apple-specific hardware support: SPI keyboard driver, NVMe suspend fix, and full T2 chip support stack.",
            "members": [
                {"id": "OM-086", "status": "required"},
                {"id": "OM-087", "status": "required"},
                {"id": "OM-088", "status": "required"},
            ],
        },
        {
            "id": "omarchy/other-vendor-hardware-support",
            "name": "Other Vendor Hardware Support",
            "intent": "Install hardware support for Dell, Microsoft Surface, Lenovo, Broadcom WiFi, Motorcomm Ethernet, Synaptics touchpad, and Tuxedo/Clevo keyboard backlight hardware.",
            "members": [
                {"id": "OM-026", "status": "required"},
                {"id": "OM-027", "status": "required"},
                {"id": "OM-089", "status": "required"},
                {"id": "OM-090", "status": "required"},
                {"id": "OM-091", "status": "required"},
                {"id": "OM-092", "status": "required"},
                {"id": "OM-093", "status": "required"},
                {"id": "OM-094", "status": "required"},
            ],
        },
        {
            "id": "omarchy/battery-and-power-management",
            "name": "Battery and Power Management",
            "intent": "Configure power profile switching and WiFi power-save mode based on AC adapter state, restrict the locate database update to AC power, and manage battery monitoring.",
            "members": [
                {"id": "OM-058", "status": "required"},
                {"id": "OM-059", "status": "required"},
                {"id": "OM-060", "status": "required"},
            ],
        },
        {
            "id": "omarchy/boot-login-and-snapshot-configuration",
            "name": "Boot, Login, and Snapshot Configuration",
            "intent": "Configure the complete boot and login stack: Plymouth boot splash, GNOME keyring for auto-login, SDDM display manager with auto-login, hibernation support, and the Limine bootloader with initramfs hooks, unified kernel image generation, btrfs snapshot menu, and EFI entry registration.",
            "members": [
                {"id": "OM-095", "status": "required"},
                {"id": "OM-096", "status": "required"},
                {"id": "OM-097", "status": "required"},
                {"id": "OM-098", "status": "required"},
                {"id": "OM-099", "status": "required"},
            ],
        },
        {
            "id": "omarchy/package-manager-finalization",
            "name": "Package Manager Finalization and Post-Install Repositories",
            "intent": "Write the final package manager configuration for the installed system, including the conditional addition of the arch-mact2 unsigned repository for Apple T2 hardware post-install updates.",
            "members": [
                {"id": "OM-100", "status": "required"},
            ],
        },
        {
            "id": "omarchy/first-run-system-setup",
            "name": "First-Run System Setup",
            "intent": "Perform one-time setup steps that require a live display session after the user's first login: network connectivity check, firewall configuration, DNS resolver symlink, battery monitor setup, and post-first-run sudoers cleanup.",
            "members": [
                {"id": "OM-101", "status": "required"},
                {"id": "OM-103", "status": "required"},
                {"id": "OM-104", "status": "required"},
                {"id": "OM-105", "status": "nice-to-have"},
                {"id": "OM-106", "status": "required"},
                {"id": "OM-107", "status": "required"},
            ],
        },
        {
            "id": "omarchy/first-run-desktop-services",
            "name": "First-Run Desktop Services",
            "intent": "Start and enable desktop services that require an active user session on first login: the Elephant application launcher, SwayOSD on-screen display server, and internal monitor recovery service.",
            "members": [
                {"id": "OM-108", "status": "required"},
                {"id": "OM-110", "status": "required"},
                {"id": "OM-111", "status": "required"},
            ],
        },
        {
            "id": "omarchy/first-run-gtk-and-user-preferences",
            "name": "First-Run GTK and User Preferences",
            "intent": "Apply GTK/GNOME theme settings, enable primary paste, and deliver onboarding notifications — all of which require a running display session and settings backend.",
            "members": [
                {"id": "OM-102", "status": "nice-to-have"},
                {"id": "OM-109", "status": "nice-to-have"},
                {"id": "OM-112", "status": "nice-to-have"},
                {"id": "OM-113", "status": "nice-to-have"},
            ],
        },
        {
            "id": "omarchy/visual-themes",
            "name": "Visual Themes",
            "intent": "Provide the complete set of 21 selectable visual theme bundles that a user can switch between at runtime without reinstalling the system.",
            "members": [
                {"id": f"OM-{i}", "status": "nice-to-have"}
                for i in range(114, 135)
            ],
        },
        {
            "id": "omarchy/desktop-theming-configuration",
            "name": "Desktop Theming Configuration",
            "intent": "Configure the active theme system including the browser policy integration, application-specific symlinks, and Plymouth boot splash branding.",
            "members": [
                {"id": "OM-029", "status": "required"},
            ],
        },
    ]


# ── YAML emitter ─────────────────────────────────────────────────────────────
def dump_yaml(obj: dict) -> str:
    """Emit YAML with consistent style (no aliases, block sequences)."""
    return yaml.dump(
        obj,
        default_flow_style=False,
        allow_unicode=True,
        sort_keys=False,
        width=120,
    )


# ── Main ─────────────────────────────────────────────────────────────────────
def main():
    os.makedirs(OPINIONS_DIR, exist_ok=True)
    os.makedirs(POINTS_DIR, exist_ok=True)

    opinions = opinion_data()
    points = point_data()

    # Verify coverage: every OM-001..OM-134 appears exactly once across all points
    covered = []
    for pt in points:
        for m in pt["members"]:
            covered.append(m["id"])
    assert len(covered) == 134, f"Expected 134 member references, got {len(covered)}"
    assert len(set(covered)) == 134, f"Duplicate member IDs: {[x for x in covered if covered.count(x) > 1]}"

    # Emit opinions
    for om_id, fields in opinions.items():
        status = get_status(om_id)
        doc = {"schema": 1, "id": om_id, "status": status}
        doc.update(fields)
        path = os.path.join(OPINIONS_DIR, f"{om_id}.yaml")
        with open(path, "w") as f:
            f.write(dump_yaml(doc))

    print(f"Wrote {len(opinions)} opinion files to {OPINIONS_DIR}/")

    # Emit points
    for pt in points:
        slug = pt["id"].replace("/", "_").replace("-", "_")
        doc = {"schema": 1, "id": pt["id"], "name": pt["name"],
               "intent": pt["intent"], "curator": "omarchy@basecamp.com",
               "members": pt["members"]}
        # Use slug from name for filename
        name_slug = pt["name"].lower().replace(",", "").replace("/", "-").replace(" ", "-")
        path = os.path.join(POINTS_DIR, f"{name_slug}.yaml")
        with open(path, "w") as f:
            f.write(dump_yaml(doc))

    print(f"Wrote {len(points)} point files to {POINTS_DIR}/")
    print("Done.")


if __name__ == "__main__":
    main()
