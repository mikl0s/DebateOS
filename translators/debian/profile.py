"""
profile.py — live-build config/ tree emitter for the DebateOS Debian translator.

Provides:
- emit_profile_tree(out_dir, manifest, variant) — writes a complete live-build
  config/ directory from a BuildManifest + variant profile dict.
- _sanitize_dst(dst) — path security gate (T-04-05, mirrors arch T-02-08).

Config tree structure (live-build lb config compatible):
  out_dir/
    config/
      includes.installer/
        preseed.cfg              — d-i automation (locale/partman/user/pkgsel)
      hooks/live/
        9000-debateos-apply.hook.chroot  — chroot-time opinion application (0755)
      package-lists/
        debateos.list.chroot_install     — target packages (.chroot_install = both envs)
      archives/
        debateos-<name>.list.chroot_install  — custom apt repos (if any)
        debateos-<name>.key.chroot           — keyrings for signed-by repos
      includes.chroot/
        etc/systemd/user/
          debateos-firstrun-<id>.service   — first-run systemd units
    build-manifest.json          — full manifest (read by installer)

Security:
- T-04-05: _sanitize_dst rejects absolute paths and .. traversal escapes.
- T-04-06: %%SENTINEL%% replacement only (never str.format with raw opinion data).
- T-04-07: sig_level Never/OptionalTrustAll → LOUD warning comment in archive file.
- T-04-08: preseed.cfg uses %%HASHED_PASSWORD%% sentinel (never plaintext).

ARCH-04 invariant: no per-variant code branches; differences are YAML data only.
"""

import json
import os
import stat
from pathlib import Path
from typing import Union

# Shim imports (see manifest.py, contract.py in this package)
import sys
_DEBIAN_DIR = os.path.dirname(os.path.abspath(__file__))
if _DEBIAN_DIR not in sys.path:
    sys.path.insert(0, _DEBIAN_DIR)
_TRANSLATORS_DIR = os.path.dirname(_DEBIAN_DIR)
if _TRANSLATORS_DIR not in sys.path:
    sys.path.insert(0, _TRANSLATORS_DIR)

from common.firstrun import render_firstrun_unit, firstrun_unit_name
from variant import apply_variant

# ---------------------------------------------------------------------------
# Template loading
# ---------------------------------------------------------------------------

_TEMPLATES_DIR = os.path.join(os.path.dirname(__file__), "templates")


def _load_template(name: str) -> str:
    """Load a template string from the templates/ directory.

    Templates use %%SENTINEL%% replacement (not str.format) to avoid conflicts
    with shell ${VAR} syntax (T-04-06, mirrors T-02-01/arch pattern).

    Args:
        name: Template filename (e.g. "preseed.cfg.tpl").

    Returns:
        The raw template string.

    Raises:
        FileNotFoundError: if the template file is missing.
    """
    path = os.path.join(_TEMPLATES_DIR, name)
    with open(path) as fh:
        return fh.read()


# ---------------------------------------------------------------------------
# Path sanitization (T-04-05 security gate, mirrors arch T-02-08)
# ---------------------------------------------------------------------------


def _sanitize_dst(dst: str) -> str:
    """Validate and return a file_asset dst path.

    Accepts only relative paths that do not escape the target root via
    ``..`` components or absolute paths (T-04-05, mirrors arch T-02-08).

    Args:
        dst: The destination path from a file_asset record.

    Returns:
        The normalized relative path string.

    Raises:
        ValueError: If ``dst`` is empty, ``'.'``, absolute, or contains ``..``
            that would escape the target root (T-04-05).
    """
    # Reject empty or sentinel-root dst values (WR-02)
    if not dst or dst.strip() in (".", ""):
        raise ValueError(
            f"file_asset dst is empty or '.' — a concrete relative path is required "
            f"(T-04-05 path traversal guard)."
        )

    # Reject absolute paths
    if os.path.isabs(dst):
        raise ValueError(
            f"file_asset dst '{dst}' is an absolute path. "
            f"All dst paths must be relative to the target root "
            f"(no absolute paths — T-04-05 path traversal guard)."
        )

    # Normalize to detect .. traversal; anchor to a sentinel root
    sentinel = "/debateos-chroot-root"
    joined = os.path.normpath(os.path.join(sentinel, dst))
    if not joined.startswith(sentinel + "/") and joined != sentinel:
        raise ValueError(
            f"file_asset dst '{dst}' traverses outside the target root "
            f"(resolved to '{joined}'). "
            f"Path components '..' that escape the chroot root are "
            f"rejected (T-04-05 path traversal guard)."
        )

    # Return the normalized relative form
    relative = os.path.relpath(joined, sentinel)
    return relative


# ---------------------------------------------------------------------------
# Main emitter
# ---------------------------------------------------------------------------


def emit_profile_tree(
    out_dir: Union[str, Path],
    manifest,           # BuildManifest instance
    variant: dict,      # loaded variant profile dict
) -> None:
    """Emit a complete live-build config/ tree from a BuildManifest + variant.

    Args:
        out_dir: Output directory path. Created if it does not exist.
        manifest: A BuildManifest instance (from common/manifest.py).
        variant: A loaded variant profile dict (from variant.py).

    Side effects:
        Writes all config/ tree files.
        9000-debateos-apply.hook.chroot is chmod 0755.

    Raises:
        ValueError: If any file_asset dst path is absolute or traverses
            outside the target root (T-04-05). Raised BEFORE any file I/O
            (fail-fast gate matching arch T-02-08 / CR-04).
    """
    out_dir = str(out_dir)

    # --- T-04-05 / CR-04: Validate ALL file_asset dst paths BEFORE any file I/O ---
    # Fail-fast: if any dst is invalid, raise immediately and write nothing.
    for fa in manifest.file_assets:
        _sanitize_dst(fa.get("dst", ""))

    os.makedirs(out_dir, exist_ok=True)

    # Apply variant to get apt sources + keyring + trust warnings
    applied = apply_variant(variant)
    apt_sources = applied["apt_sources"]
    keyring_install_before_repos = applied["keyring_install_before_repos"]
    trust_warnings = manifest.trust_warnings + applied["trust_warnings"]

    # --- config/includes.installer/preseed.cfg ---
    _write_preseed(out_dir)

    # --- config/hooks/live/9000-debateos-apply.hook.chroot (0755) ---
    _write_chroot_hook(out_dir, manifest)

    # --- config/package-lists/debateos.list.chroot_install ---
    _write_package_list(out_dir, manifest.target_packages)

    # --- config/archives/<repo>.list.chroot_install + .key.chroot ---
    _write_archives(out_dir, apt_sources)

    # --- config/includes.chroot/etc/systemd/user/ first-run units ---
    _write_firstrun_units(out_dir, manifest.first_run)

    # --- build-manifest.json ---
    manifest_dict = manifest.to_dict()
    manifest_dict["keyring_install_before_repos"] = keyring_install_before_repos
    manifest_dict["trust_warnings"] = trust_warnings
    _write_build_manifest(out_dir, manifest_dict)


# ---------------------------------------------------------------------------
# Private writers
# ---------------------------------------------------------------------------


def _write_preseed(out_dir: str) -> None:
    """Write config/includes.installer/preseed.cfg from the template."""
    tpl = _load_template("preseed.cfg.tpl")
    # %%SENTINEL%% replacement for the preseed template (T-04-06, T-04-08).
    # Default sentinels for build-time replacement:
    content = tpl
    content = content.replace("%%USERNAME%%", "%%USERNAME%%")
    content = content.replace("%%USER_FULLNAME%%", "%%USER_FULLNAME%%")
    content = content.replace("%%HASHED_PASSWORD%%", "%%HASHED_PASSWORD%%")
    content = content.replace("%%PKGSEL_PACKAGES%%", "ssh openssh-server")

    preseed_dir = os.path.join(out_dir, "config", "includes.installer")
    os.makedirs(preseed_dir, exist_ok=True)
    _write_file(os.path.join(preseed_dir, "preseed.cfg"), content)


def _write_chroot_hook(out_dir: str, manifest) -> None:
    """Write the chroot hook (0755) from the template using %%SENTINEL%% replacement."""
    tpl = _load_template("chroot-install.hook.tpl")

    # Build package install stanza
    packages = manifest.target_packages
    if packages:
        # Each package on its own line with continuation
        pkg_lines = " \\\n  ".join(packages)
    else:
        # No packages — use apt-get install with : (no-op)
        pkg_lines = ":"

    # Build remove-packages stanza
    remove_stanza = ""
    if manifest.remove_packages:
        remove_stanza = (
            "DEBIAN_FRONTEND=noninteractive apt-get remove -y "
            + " ".join(manifest.remove_packages)
        )
    else:
        remove_stanza = "# No packages to remove"

    # Build file-asset deployment stanza (T-04-06: paths already sanitized)
    file_asset_lines = []
    for fa in manifest.file_assets:
        dst = _sanitize_dst(fa.get("dst", ""))
        src = fa.get("src", "")
        mode = fa.get("mode", "0644")
        # Files are placed in includes.chroot overlay; in the hook we write them directly
        # using install -Dm<mode>. The src is relative to the assets dir.
        file_asset_lines.append(
            f"# Deploy file asset: {src} → /{dst}\n"
            f"if [ -f /debateos-assets/{dst} ]; then\n"
            f"  install -Dm{mode} /debateos-assets/{dst} /{dst}\n"
            f"fi"
        )
    file_asset_stanza = "\n\n".join(file_asset_lines) if file_asset_lines else "# No file assets"

    # Build service enable stanza
    service_lines = []
    for svc in manifest.system_services:
        if svc.get("enable", False):
            service_lines.append(f"systemctl enable {svc['name']}")
    service_enable_stanza = "\n".join(service_lines) if service_lines else "# No services to enable"

    # Build sysctl stanza
    sysctl_lines = []
    for param in manifest.sysctl_params:
        drop_in = param.get("drop_in_file", "50-debateos.conf")
        key = param["key"]
        value = param["value"]
        sysctl_lines.append(
            f"# sysctl: {key}\n"
            f"mkdir -p /etc/sysctl.d\n"
            f"printf '{key} = {value}\\n' >> /etc/sysctl.d/{drop_in}"
        )
    sysctl_stanza = "\n".join(sysctl_lines) if sysctl_lines else "# No sysctl params"

    # Build kernel param stanza (GRUB_CMDLINE_LINUX)
    kernel_param_lines = []
    for param in manifest.kernel_params:
        key = param.get("key", "")
        value = param.get("value", "")
        kernel_param_lines.append(f"  {key}={value}" if value else f"  {key}")
    if kernel_param_lines:
        params_str = " ".join(p.strip() for p in kernel_param_lines)
        kernel_param_stanza = (
            f"# Kernel parameters (GRUB_CMDLINE_LINUX)\n"
            f"if [ -f /etc/default/grub ]; then\n"
            f"  sed -i 's/^GRUB_CMDLINE_LINUX=.*/GRUB_CMDLINE_LINUX=\"{params_str}\"/' /etc/default/grub\n"
            f"  update-grub 2>/dev/null || true\n"
            f"fi"
        )
    else:
        kernel_param_stanza = "# No kernel params"

    # Build group membership stanza
    group_lines = []
    for gm in manifest.group_memberships:
        group = gm["group"]
        group_lines.append(
            f"# Add primary user to group: {group}\n"
            f"getent group {group} >/dev/null 2>&1 || groupadd {group}\n"
            f"# Note: user addition done at install time via preseed passwd/user-default-groups"
        )
    group_membership_stanza = "\n\n".join(group_lines) if group_lines else "# No group memberships"

    # %%SENTINEL%% replacement (T-04-06 — never str.format with raw opinion data)
    content = tpl
    content = content.replace("%%PACKAGES%%", pkg_lines)
    content = content.replace("%%REMOVE_PACKAGES_STANZA%%", remove_stanza)
    content = content.replace("%%FILE_ASSET_STANZA%%", file_asset_stanza)
    content = content.replace("%%SERVICE_ENABLE_STANZA%%", service_enable_stanza)
    content = content.replace("%%SYSCTL_STANZA%%", sysctl_stanza)
    content = content.replace("%%KERNEL_PARAM_STANZA%%", kernel_param_stanza)
    content = content.replace("%%GROUP_MEMBERSHIP_STANZA%%", group_membership_stanza)

    hook_dir = os.path.join(out_dir, "config", "hooks", "live")
    os.makedirs(hook_dir, exist_ok=True)
    hook_path = os.path.join(hook_dir, "9000-debateos-apply.hook.chroot")
    _write_file(hook_path, content)

    # chmod 0755 — required for lb build hook execution
    os.chmod(
        hook_path,
        stat.S_IRWXU | stat.S_IRGRP | stat.S_IXGRP | stat.S_IROTH | stat.S_IXOTH,
    )


def _write_package_list(out_dir: str, target_packages: list) -> None:
    """Write config/package-lists/debateos.list.chroot_install.

    Uses .chroot_install suffix so packages land in BOTH live env and installed
    system (Pitfall 1 — not .list.chroot which is live-only).
    """
    pkg_dir = os.path.join(out_dir, "config", "package-lists")
    os.makedirs(pkg_dir, exist_ok=True)
    content = "\n".join(target_packages) + "\n" if target_packages else ""
    _write_file(os.path.join(pkg_dir, "debateos.list.chroot_install"), content)


def _write_archives(out_dir: str, apt_sources: list) -> None:
    """Write config/archives/<repo>.list.chroot_install + <repo>.key.chroot per apt source.

    Security (T-04-07): Never/OptionalTrustAll sources get LOUD WARNING comments.
    """
    if not apt_sources:
        return

    archives_dir = os.path.join(out_dir, "config", "archives")
    os.makedirs(archives_dir, exist_ok=True)

    for src in apt_sources:
        filename = src["filename"]
        line = src["line"]
        comment = src.get("comment", "")

        # Build the .list.chroot_install content
        lines = []
        if comment:
            lines.append(comment)
        lines.append(line)
        list_content = "\n".join(lines) + "\n"
        _write_file(os.path.join(archives_dir, filename), list_content)

        # Write a stub .key.chroot if a keyring is specified
        keyring = src.get("keyring")
        if keyring:
            repo_name = filename.split("-", 1)[1].split(".")[0] if "-" in filename else "unknown"
            key_filename = f"debateos-{repo_name}.key.chroot"
            key_content = (
                f"# Keyring stub for {repo_name}\n"
                f"# Source URL: {keyring}\n"
                f"# live-build installs this keyring before apt fetch (Pitfall 4)\n"
            )
            _write_file(os.path.join(archives_dir, key_filename), key_content)


def _write_firstrun_units(out_dir: str, first_run_opinions: list) -> None:
    """Write systemd user oneshot units for first-run opinions.

    Units land in config/includes.chroot/etc/systemd/user/ — live-build
    overlays includes.chroot/ directly into the chroot filesystem (Pattern 4).
    """
    user_unit_dir = os.path.join(
        out_dir, "config", "includes.chroot", "etc", "systemd", "user"
    )
    os.makedirs(user_unit_dir, exist_ok=True)

    for entry in first_run_opinions:
        opinion_id = entry.get("id", "unknown")
        description = entry.get("description", f"first-run {opinion_id}")
        exec_path = f"/usr/lib/debateos/firstrun/{opinion_id}.sh"

        # Use common firstrun template (parameterized template_dir)
        unit_content = render_firstrun_unit(
            opinion_id=opinion_id,
            description=description,
            exec_path=exec_path,
        )
        unit_name = firstrun_unit_name(opinion_id)
        _write_file(os.path.join(user_unit_dir, unit_name), unit_content)


def _write_build_manifest(out_dir: str, manifest_dict: dict) -> None:
    """Write build-manifest.json — the runtime data for post-install operations."""
    content = json.dumps(manifest_dict, indent=2)
    _write_file(os.path.join(out_dir, "build-manifest.json"), content)


def _write_file(path: str, content: str) -> None:
    """Write text content to path, creating parent directories as needed."""
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as fh:
        fh.write(content)
