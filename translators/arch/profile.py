"""
profile.py — archiso profile tree emitter for the DebateOS Arch translator.

Provides:
- emit_profile_tree(out_dir, manifest, variant) — writes a complete archiso
  profile directory from a BuildManifest + variant profile dict.

Profile tree structure (archiso 88-1 releng-compatible):
  out_dir/
    profiledef.sh              — ISO metadata + build config
    packages.x86_64            — minimal live-env packages (NOT target install set)
    pacman.conf                — pacman config with variant repos injected
    airootfs/
      root/
        debateos-install.sh    — generated installer script (0755)
        .zlogin                — hooks installer on tty1 (Pattern 1)
      etc/
        systemd/
          user/
            debateos-firstrun-<id>.service  — per first-run opinion
    build-manifest.json        — full manifest (read by installer via jq, Pitfall 6)

Security:
- T-02-08: _sanitize_dst rejects absolute paths and .. traversal escapes.
- T-02-09: all opinion data serialized as JSON; installer reads via jq.
- T-02-10: sig_level=Never repos get a warning comment in pacman.conf.
- T-02-11: first-run units are systemd USER services (etc/systemd/user/).

ARCH-04 invariant: no per-variant code branches; differences are YAML data only.
"""

import json
import os
import stat
from pathlib import Path
from typing import Union

from firstrun import render_firstrun_unit, firstrun_unit_name
from variant import apply_variant

# ---------------------------------------------------------------------------
# Minimal live-env package set (Pitfall 2)
# ---------------------------------------------------------------------------
# These go into packages.x86_64 — the live ISO's squashfs.
# The target package set (Omarchy ~285 packages) goes into build-manifest.json.
# Source: releng baseline + installer deps required for debateos-install.sh.

_LIVE_ENV_PACKAGES = [
    # Releng baseline essentials (installer needs these to be on the live env)
    "base",
    "linux",
    "linux-firmware",
    # Boot support packages required by profiledef.sh bootmodes (02-RESEARCH.md Minimal Deviation)
    # The generator overlays packages.x86_64 on top of the releng baseline, so these
    # must be included here to satisfy mkarchiso's bootmode validation.
    "syslinux",               # required for bios.syslinux bootmode (BIOS legacy)
    # Installer deps (used by debateos-install.sh)
    "arch-install-scripts",   # pacstrap + arch-chroot + genfstab
    "btrfs-progs",            # mkfs.btrfs + btrfs subvol operations
    "dosfstools",             # mkfs.fat (EFI partition)
    "e2fsprogs",              # mkfs.ext4 (fallback / general)
    "gdisk",                  # sgdisk for GPT partitioning
    "jq",                     # JSON manifest parsing at install time (Pitfall 6)
    # Network
    "iwd",
    "dhcpcd",
    # Live env convenience
    "terminus-font",
    "zsh",
]


# ---------------------------------------------------------------------------
# Template loading
# ---------------------------------------------------------------------------

_TEMPLATES_DIR = os.path.join(os.path.dirname(__file__), "templates")


def _load_template(name: str) -> str:
    """Load a template string from the templates/ directory.

    Templates use Python str.format() style placeholders: {variable_name}.
    Literal braces in shell content must be escaped as {{ and }}.

    Args:
        name: Template filename (e.g. "profiledef.sh.tpl").

    Returns:
        The raw template string.

    Raises:
        FileNotFoundError: if the template file is missing.
    """
    path = os.path.join(_TEMPLATES_DIR, name)
    with open(path) as fh:
        return fh.read()


# ---------------------------------------------------------------------------
# Path sanitization (T-02-08 security gate)
# ---------------------------------------------------------------------------


def _sanitize_dst(dst: str) -> str:
    """Validate and return a file_asset dst path.

    Accepts only relative paths that do not escape the target root via
    ``..`` components or absolute paths (T-02-08).

    Args:
        dst: The destination path from a file_asset record.

    Returns:
        The normalized relative path string.

    Raises:
        ValueError: If ``dst`` is empty, ``'.'``, absolute, or contains ``..``
            that would escape the target root. The error message names the
            offending dst (WR-02, T-02-08).
    """
    # WR-02: Reject empty or sentinel-root dst values.
    if not dst or dst.strip() in (".", ""):
        raise ValueError(
            f"file_asset dst is empty or '.' — a concrete relative path is required "
            f"(T-02-08 path traversal guard)."
        )

    # Reject absolute paths
    if os.path.isabs(dst):
        raise ValueError(
            f"file_asset dst '{dst}' is an absolute path. "
            f"All dst paths must be relative to the target root "
            f"(no absolute paths — T-02-08 path traversal guard)."
        )

    # Normalize to detect .. traversal
    # We anchor to a sentinel root and check containment.
    sentinel = "/debateos-airootfs-root"
    joined = os.path.normpath(os.path.join(sentinel, dst))
    if not joined.startswith(sentinel + "/") and joined != sentinel:
        raise ValueError(
            f"file_asset dst '{dst}' traverses outside the target root "
            f"(resolved to '{joined}'). "
            f"Path components '..' that escape the airootfs/target root are "
            f"rejected (T-02-08 path traversal guard)."
        )

    # Return the normalized relative form (strip any leading slash artifact)
    relative = os.path.relpath(joined, sentinel)
    return relative


# ---------------------------------------------------------------------------
# pacman.conf repo section builder
# ---------------------------------------------------------------------------

def _build_repo_section(repo: dict) -> str:
    """Build a single pacman.conf repo section string from a repo dict.

    Args:
        repo: Repo dict with at least "name" and "url" keys; optional "sig_level".

    Returns:
        A formatted repo section string including trust warning comment for
        sig_level=Never repos (T-02-10).
    """
    lines = []
    sig = repo.get("sig_level", "")
    if sig == "Never":
        lines.append(
            f"# WARNING: sig_level=Never for [{repo['name']}] — "
            f"all package signatures bypassed; verify source (T-02-10)"
        )
    elif sig == "OptionalTrustAll":
        lines.append(
            f"# WARNING: sig_level=OptionalTrustAll for [{repo['name']}] — "
            f"unsigned packages accepted; verify source (T-02-10, WR-01)"
        )
    lines.append(f"[{repo['name']}]")
    lines.append(f"Server = {repo['url']}")
    if sig and sig != "Required DatabaseOptional":
        lines.append(f"SigLevel = {sig}")
    return "\n".join(lines)


# ---------------------------------------------------------------------------
# Main emitter
# ---------------------------------------------------------------------------


def emit_profile_tree(
    out_dir: Union[str, Path],
    manifest,          # BuildManifest instance
    variant: dict,     # loaded variant profile dict
) -> None:
    """Emit a complete archiso profile tree from a BuildManifest + variant.

    Args:
        out_dir: Output directory path. Created if it does not exist.
        manifest: A BuildManifest instance (from manifest.py).
        variant: A loaded variant profile dict (from variant.py).

    Side effects:
        Writes all profile files. debateos-install.sh is chmod 0755.

    Raises:
        ValueError: If any file_asset dst path is absolute or traverses
            outside the target root (T-02-08).
    """
    out_dir = str(out_dir)

    # --- T-02-08 / CR-04: Validate ALL file_asset dst paths BEFORE any file I/O ---
    # Fail-fast: if any dst is invalid, raise immediately and write nothing.
    for fa in manifest.file_assets:
        _sanitize_dst(fa.get("dst", ""))

    os.makedirs(out_dir, exist_ok=True)

    # Apply variant to get ordered repos + keyring packages
    applied = apply_variant(variant, _default_base_repos())
    ordered_repos = applied["repos"]
    keyring_pkgs = applied["keyring_install_before_repos"]
    trust_warnings = applied["trust_warnings"]

    source_date_epoch = manifest.source_date_epoch

    # --- profiledef.sh ---
    _write_profiledef(out_dir, source_date_epoch)

    # --- packages.x86_64 ---
    _write_packages_x86_64(out_dir)

    # --- pacman.conf ---
    _write_pacman_conf(out_dir, ordered_repos)

    # --- airootfs/root/debateos-install.sh (0755) ---
    _write_installer(out_dir, source_date_epoch, keyring_pkgs)

    # --- airootfs/root/.zlogin ---
    _write_zlogin(out_dir)

    # --- airootfs/etc/systemd/user/ first-run units ---
    _write_firstrun_units(out_dir, manifest.first_run)

    # --- build-manifest.json ---
    # Inject keyring_install_before_repos into manifest dict for installer
    manifest_dict = manifest.to_dict()
    manifest_dict["keyring_install_before_repos"] = keyring_pkgs
    manifest_dict["trust_warnings"] = manifest.trust_warnings + trust_warnings
    _write_build_manifest(out_dir, manifest_dict)


def _default_base_repos() -> list:
    """Return the standard base Arch repo list (core + extra)."""
    return [
        {
            "name": "core",
            "url": "https://geo.mirror.pkgbuild.com/$repo/os/$arch",
            "sig_level": "Required",
        },
        {
            "name": "extra",
            "url": "https://geo.mirror.pkgbuild.com/$repo/os/$arch",
            "sig_level": "Required",
        },
    ]


def _write_profiledef(out_dir: str, source_date_epoch: int) -> None:
    """Write profiledef.sh from the template."""
    tpl = _load_template("profiledef.sh.tpl")
    content = tpl.format(
        source_date_epoch=source_date_epoch,
    )
    _write_file(os.path.join(out_dir, "profiledef.sh"), content)


def _write_packages_x86_64(out_dir: str) -> None:
    """Write packages.x86_64 with the minimal live-env set (Pitfall 2)."""
    content = "\n".join(_LIVE_ENV_PACKAGES) + "\n"
    _write_file(os.path.join(out_dir, "packages.x86_64"), content)


def _write_pacman_conf(out_dir: str, ordered_repos: list) -> None:
    """Write pacman.conf with variant repos injected (T-02-10 warnings inline)."""
    # Split repos into above_core and below_core sets
    # The ordered_repos list from apply_variant is already sorted correctly:
    # above_core repos first, then base repos (core, extra), then below_core repos.
    # We find the split points by looking for "core" and "extra" in the list.
    base_names = {"core", "extra"}

    above = []
    base = []
    below = []

    encountered_base = False
    finished_base = False

    for repo in ordered_repos:
        name = repo.get("name", "")
        if name in base_names:
            encountered_base = True
            base.append(repo)
        elif not encountered_base:
            # Variant repos before base repos
            above.append(repo)
        else:
            # finished base section
            finished_base = True
            below.append(repo)

    # Build sections
    above_section = ""
    for repo in above:
        above_section += _build_repo_section(repo) + "\n\n"

    below_section = ""
    for repo in below:
        below_section += _build_repo_section(repo) + "\n\n"

    tpl = _load_template("pacman.conf.tpl")
    content = tpl.format(
        variant_repos_above=above_section.rstrip("\n"),
        variant_repos_below=below_section.rstrip("\n"),
    )
    _write_file(os.path.join(out_dir, "pacman.conf"), content)


def _write_installer(
    out_dir: str,
    source_date_epoch: int,
    keyring_pkgs: list,
) -> None:
    """Write debateos-install.sh at 0755 (T-02-09 jq-driven, no shell injection)."""
    airootfs_root = os.path.join(out_dir, "airootfs", "root")
    os.makedirs(airootfs_root, exist_ok=True)

    # Compute the ISO label for the installer header comment
    import datetime
    try:
        dt = datetime.datetime.fromtimestamp(source_date_epoch, tz=datetime.timezone.utc)
        iso_label = f"DEBATEOS_{dt.strftime('%Y%m')}"
    except Exception:
        iso_label = "DEBATEOS"

    # The installer template contains many shell ${VAR} expansions.
    # We perform a targeted header substitution (not str.format on the whole file)
    # to avoid KeyError on shell variable names used as format placeholders.
    tpl = _load_template("installer.sh.tpl")
    # Replace only our sentinel markers (not shell ${...} syntax)
    content = tpl.replace("%%SOURCE_DATE_EPOCH%%", str(source_date_epoch))
    content = content.replace("%%ISO_LABEL%%", iso_label)

    installer_path = os.path.join(airootfs_root, "debateos-install.sh")
    _write_file(installer_path, content)
    # chmod 0755 — required for profiledef.sh file_permissions + ISO execution
    os.chmod(
        installer_path,
        stat.S_IRWXU | stat.S_IRGRP | stat.S_IXGRP | stat.S_IROTH | stat.S_IXOTH,
    )


def _write_zlogin(out_dir: str) -> None:
    """Write airootfs/root/.zlogin — calls installer on /dev/tty1 (Pattern 1)."""
    airootfs_root = os.path.join(out_dir, "airootfs", "root")
    os.makedirs(airootfs_root, exist_ok=True)
    content = (
        "# airootfs/root/.zlogin — generated by DebateOS Arch translator\n"
        "# Source: releng profile (archiso 88-1) Pattern 1\n"
        "# Call the unattended installer when running on tty1 (autologin hook).\n"
        "if [[ \"$(tty)\" == \"/dev/tty1\" ]]; then\n"
        "    /root/debateos-install.sh 2>&1 | tee /root/install.log\n"
        "fi\n"
    )
    _write_file(os.path.join(airootfs_root, ".zlogin"), content)


def _write_firstrun_units(out_dir: str, first_run_opinions: list) -> None:
    """Write systemd user oneshot units for first-run opinions (T-02-11)."""
    user_unit_dir = os.path.join(
        out_dir, "airootfs", "etc", "systemd", "user"
    )
    os.makedirs(user_unit_dir, exist_ok=True)

    for entry in first_run_opinions:
        opinion_id = entry.get("id", "unknown")
        description = entry.get("description", f"first-run {opinion_id}")
        exec_path = f"/usr/lib/debateos/firstrun/{opinion_id}.sh"

        unit_content = render_firstrun_unit(
            opinion_id=opinion_id,
            description=description,
            exec_path=exec_path,
        )
        unit_name = firstrun_unit_name(opinion_id)
        _write_file(os.path.join(user_unit_dir, unit_name), unit_content)


def _write_build_manifest(out_dir: str, manifest_dict: dict) -> None:
    """Write build-manifest.json — the runtime data the installer reads (Pitfall 6)."""
    content = json.dumps(manifest_dict, indent=2)
    _write_file(os.path.join(out_dir, "build-manifest.json"), content)


def _write_file(path: str, content: str) -> None:
    """Write text content to path, creating parent directories as needed."""
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as fh:
        fh.write(content)
