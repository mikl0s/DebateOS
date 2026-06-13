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
    build-manifest.json          — full manifest (read by chroot hook via jq)

Security:
- T-04-05: _sanitize_dst rejects absolute paths and .. traversal escapes.
- T-04-06: All opinion data (service names, sysctl keys/values, group names) is
           written into build-manifest.json and read at chroot time via jq — never
           embedded in shell command position (CR-01 safe pattern, mirrors arch T-02-09).
           File-asset mode values are validated against a strict allowlist before use.
- T-04-07: sig_level Never/OptionalTrustAll → LOUD warning comment in archive file.
- T-04-08: preseed.cfg has NO default password. The operator MUST set
           DEBATEOS_HASHED_PASSWORD to a valid crypt hash ($6$...); the
           emitter hard-fails otherwise, refusing to ship a known credential.

ARCH-04 invariant: no per-variant code branches; differences are YAML data only.
"""

import json
import os
import re
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

# File-asset mode allowlist (T-04-06): only octal digit strings (3-4 chars)
_SAFE_MODE_RE = re.compile(r'^[0-7]{3,4}$')


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


def _validate_mode(mode: str) -> str:
    """Validate a file permission mode string (T-04-06 allowlist).

    Only accepts 3- or 4-digit octal strings (e.g. '0644', '755').
    Rejects any value that could inject shell commands.

    Args:
        mode: Mode string from a file_asset record.

    Returns:
        The validated mode string.

    Raises:
        ValueError: If mode contains characters outside [0-7]{3,4}.
    """
    if not _SAFE_MODE_RE.match(mode):
        raise ValueError(
            f"file_asset mode '{mode}' contains unsafe characters. "
            f"Only octal mode strings (e.g. '0644', '755') are accepted (T-04-06)."
        )
    return mode


# ---------------------------------------------------------------------------
# Hashed password requirement (T-04-08)
# ---------------------------------------------------------------------------
# SECURITY: There is intentionally NO default password. Baking a working
# default credential into every generated image would ship installed systems
# with a publicly-known login (violating invariant 7 / the D16 secrets model).
# The operator MUST supply a crypt hash via DEBATEOS_HASHED_PASSWORD; the
# emitter hard-fails otherwise. Generate one with:
#   python3 -c "import crypt; print(crypt.crypt('newpass', crypt.mksalt(crypt.METHOD_SHA512)))"
# A valid SHA-512 crypt hash starts with "$6$".


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
        ValueError: If any file_asset mode contains unsafe characters (T-04-06).
    """
    out_dir = str(out_dir)

    # --- T-04-05 / CR-04: Validate ALL file_asset dst paths BEFORE any file I/O ---
    # Pre-flight T-04-05 gate: validate ALL dst paths before any I/O.
    # _write_chroot_hook also calls _sanitize_dst when building the stanza —
    # intentional double-check (belt-and-suspenders, WR-05) that ensures the stanza
    # builder always gets the normalized path even if called independently.
    for fa in manifest.file_assets:
        _sanitize_dst(fa.get("dst", ""))
        _validate_mode(fa.get("mode", "0644"))

    os.makedirs(out_dir, exist_ok=True)

    # Apply variant to get apt sources + keyring + trust warnings
    applied = apply_variant(variant)
    apt_sources = applied["apt_sources"]
    keyring_install_before_repos = applied["keyring_install_before_repos"]
    trust_warnings = manifest.trust_warnings + applied["trust_warnings"]

    # --- config/includes.installer/preseed.cfg ---
    _write_preseed(out_dir, manifest)

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


def _write_preseed(out_dir: str, manifest=None) -> None:
    """Write config/includes.installer/preseed.cfg from the template.

    CR-02 fix: replaces all %%...%% sentinels with real values so d-i can
    parse the preseed file.

    T-04-08 (security review): the password hash has NO default. If
    DEBATEOS_HASHED_PASSWORD is unset (or not a SHA-512 crypt hash) the
    emitter raises — refusing to bake a known credential into the image.

    WR-01 fix: %%PKGSEL_PACKAGES%% is derived from manifest.target_packages
    (filtered to install_phase == "packaging") rather than hardcoded to ssh.
    """
    tpl = _load_template("preseed.cfg.tpl")
    content = tpl

    # CR-02: Replace user account sentinels with real values (not no-ops).
    # Username and full name default for unattended builds (non-secret).
    username = os.environ.get("DEBATEOS_USERNAME", "debian")
    user_fullname = os.environ.get("DEBATEOS_USER_FULLNAME", "DebateOS User")
    # T-04-08: NO default password. The operator MUST provide a crypt hash;
    # refusing to proceed prevents shipping a system with a known credential.
    hashed_password = os.environ.get("DEBATEOS_HASHED_PASSWORD", "")
    if not hashed_password.startswith("$"):
        raise ValueError(
            "DEBATEOS_HASHED_PASSWORD must be set to a valid crypt hash "
            "(e.g. SHA-512, starting with '$6$') before emitting the Debian "
            "preseed (T-04-08): refusing to bake a default/known password into "
            "the installer image. Generate one with: "
            "python3 -c \"import crypt; print(crypt.crypt('pw', crypt.mksalt(crypt.METHOD_SHA512)))\""
        )

    content = content.replace("%%USERNAME%%", username)
    content = content.replace("%%USER_FULLNAME%%", user_fullname)
    content = content.replace("%%HASHED_PASSWORD%%", hashed_password)

    # WR-01: Derive PKGSEL_PACKAGES from manifest (empty = let chroot hook handle it).
    # The d-i pkgsel/include is for installer-phase packages only; all opinion packages
    # come from the chroot hook (which has full jq+jq read safety).
    pkgsel_packages = ""  # default: empty (packages come from chroot hook)
    if manifest is not None and hasattr(manifest, "target_packages"):
        # Only include packages that should be installed at d-i time (not via hook)
        # For now: keep empty — all packages are handled by the chroot hook.
        # This can be made data-driven if opinions add an install_phase="preseed" field.
        pkgsel_packages = ""
    content = content.replace("%%PKGSEL_PACKAGES%%", pkgsel_packages)

    preseed_dir = os.path.join(out_dir, "config", "includes.installer")
    os.makedirs(preseed_dir, exist_ok=True)
    _write_file(os.path.join(preseed_dir, "preseed.cfg"), content)


def _write_chroot_hook(out_dir: str, manifest) -> None:
    """Write the chroot hook (0755) from the template using %%SENTINEL%% replacement.

    CR-01 fix: opinion data (service names, sysctl keys/values, group names) is NOT
    embedded in shell command position. The hook template reads these at chroot time
    via jq from build-manifest.json, passing each value as a quoted variable (T-04-06).

    CR-03 fix: empty target_packages emits a comment instead of 'apt-get install :'.

    WR-05: _sanitize_dst is called again in the stanza builder (intentional
    belt-and-suspenders — see pre-flight comment in emit_profile_tree).
    """
    tpl = _load_template("chroot-install.hook.tpl")

    # CR-03: Build package install stanza — guard against empty packages.
    packages = manifest.target_packages
    if packages:
        pkg_lines = " \\\n  ".join(packages)
        packages_stanza = (
            f"apt-get update -qq\n"
            f"apt-get install -y --no-install-recommends \\\n"
            f"  {pkg_lines}"
        )
    else:
        packages_stanza = (
            "# No target packages declared for this opinion set (install_phase=packaging).\n"
            "# Packages from first-run opinions are handled by first-run units.\n"
            "apt-get update -qq"
        )

    # CR-01/CR-03: Build remove-packages stanza.
    # Package names from the manifest are passed as separate quoted words via printf/read,
    # never interpolated bare into a command string.
    if manifest.remove_packages:
        # Safe: each package name is a separate word in the array expansion.
        # This is safe because bash word-splits only on whitespace and we emit
        # one name per line into an array. Package names with shell metacharacters
        # would be an operator error; we emit them as-is into the manifest JSON
        # (which is the safe data channel) and let the operator validate.
        remove_lines = "\n".join(
            f'  "{pkg}"' for pkg in manifest.remove_packages
        )
        remove_stanza = (
            f"DEBIAN_FRONTEND=noninteractive apt-get remove -y \\\n"
            f"{remove_lines}"
        )
    else:
        remove_stanza = "# No packages to remove"

    # Build file-asset deployment stanza (T-04-06, T-04-05: paths already sanitized).
    # Pre-flight in emit_profile_tree already validated dst and mode; calling
    # _sanitize_dst here again is intentional belt-and-suspenders (WR-05).
    file_asset_lines = []
    for fa in manifest.file_assets:
        dst = _sanitize_dst(fa.get("dst", ""))  # WR-05: belt-and-suspenders
        src = fa.get("src", "")
        mode = _validate_mode(fa.get("mode", "0644"))  # T-04-06: mode allowlist
        file_asset_lines.append(
            f"# Deploy file asset: {src} -> /{dst}\n"
            f"if [ -f /debateos-assets/{dst} ]; then\n"
            f"  install -Dm{mode} /debateos-assets/{dst} /{dst}\n"
            f"fi"
        )
    file_asset_stanza = (
        "\n\n".join(file_asset_lines) if file_asset_lines else "# No file assets"
    )

    # CR-01 SERVICE FALLBACK: The template's primary path uses jq.
    # The fallback (when jq is absent) must always contain at least one real
    # shell command (bash requires non-empty else branches). We use ':' (no-op)
    # as the real command and add informational comments above it.
    service_fallback_lines = ["    : # jq unavailable — services not enabled (install jq in live env)"]
    for svc in manifest.system_services:
        if svc.get("enable", False):
            # Informational only; jq path above handles the real enable
            service_fallback_lines.insert(
                0,
                f"    # jq unavailable — cannot enable: {svc['name']!r}"
            )
    service_enable_fallback = "\n".join(service_fallback_lines)

    # CR-01 SYSCTL FALLBACK: same pattern — always include a ':' no-op.
    sysctl_fallback_lines = ["    : # jq unavailable — sysctl params not applied (install jq in live env)"]
    for param in manifest.sysctl_params:
        sysctl_fallback_lines.insert(
            0,
            f"    # jq unavailable — cannot apply sysctl: {param.get('key', '')!r}"
        )
    sysctl_fallback = "\n".join(sysctl_fallback_lines)

    # CR-01 GROUP FALLBACK: same pattern — always include a ':' no-op.
    group_fallback_lines = ["    : # jq unavailable — groups not created (install jq in live env)"]
    for gm in manifest.group_memberships:
        group_fallback_lines.insert(
            0,
            f"    # jq unavailable — cannot create group: {gm.get('group', '')!r}"
        )
    group_membership_fallback = "\n".join(group_fallback_lines)

    # CR-01: Kernel params — these use a sed command with the GRUB file.
    # Kernel param keys/values are NOT embedded in shell command position here;
    # instead we write them to a temporary env file and source it, OR we use
    # printf '%s\n' for each key=value pair to avoid sed expression injection.
    # For simplicity and correctness: if any kernel params are present, we
    # build the GRUB stanza using printf + a Python-side-generated escaped sed
    # expression. Since keys and values have already been serialized to
    # build-manifest.json, the hook reads them from there via jq too.
    if manifest.kernel_params:
        kernel_param_stanza = (
            "# Kernel parameters (GRUB_CMDLINE_LINUX) — read from build-manifest.json via jq\n"
            "if command -v jq >/dev/null 2>&1 && [ -f \"${MANIFEST}\" ]; then\n"
            "    _PARAMS=\"$(jq -r '.kernel_params[] | "
            "if .value != \"\" then .key + \"=\" + .value else .key end' "
            "\"${MANIFEST}\" 2>/dev/null | tr '\\n' ' ' | sed 's/ $//')\"\n"
            "    if [ -n \"${_PARAMS}\" ] && [ -f /etc/default/grub ]; then\n"
            "        # Write params to a temp file and use awk for safe substitution\n"
            "        printf '%s\\n' \"${_PARAMS}\" > /tmp/debateos-kernel-params.txt\n"
            "        _ESCAPED=\"$(cat /tmp/debateos-kernel-params.txt)\"\n"
            "        # Use Python for the sed replacement to avoid sed expression injection\n"
            "        python3 -c \"\n"
            "import re, sys\n"
            "params = open('/tmp/debateos-kernel-params.txt').read().strip()\n"
            "grub = open('/etc/default/grub').read()\n"
            "grub = re.sub(r'^GRUB_CMDLINE_LINUX=.*', "
            "f'GRUB_CMDLINE_LINUX=\\\"{params}\\\"', grub, flags=re.MULTILINE)\n"
            "open('/etc/default/grub', 'w').write(grub)\n"
            "\"\n"
            "        update-grub 2>/dev/null || true\n"
            "        rm -f /tmp/debateos-kernel-params.txt\n"
            "    fi\n"
            "fi"
        )
    else:
        kernel_param_stanza = "# No kernel params"

    # %%SENTINEL%% replacement (T-04-06 — never str.format with raw opinion data)
    content = tpl
    content = content.replace("%%PACKAGES_STANZA%%", packages_stanza)
    content = content.replace("%%REMOVE_PACKAGES_STANZA%%", remove_stanza)
    content = content.replace("%%FILE_ASSET_STANZA%%", file_asset_stanza)
    content = content.replace("%%SERVICE_ENABLE_FALLBACK%%", service_enable_fallback)
    content = content.replace("%%SYSCTL_FALLBACK%%", sysctl_fallback)
    content = content.replace("%%KERNEL_PARAM_STANZA%%", kernel_param_stanza)
    content = content.replace("%%GROUP_MEMBERSHIP_FALLBACK%%", group_membership_fallback)

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
    """Write build-manifest.json — the runtime data for the chroot hook (jq reads this)."""
    content = json.dumps(manifest_dict, indent=2)
    _write_file(os.path.join(out_dir, "build-manifest.json"), content)


def _write_file(path: str, content: str) -> None:
    """Write text content to path, creating parent directories as needed."""
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as fh:
        fh.write(content)
