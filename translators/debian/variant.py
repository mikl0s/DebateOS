"""
variant.py — Variant profile loader and apt sources applicator for the Debian translator.

Provides:
- load_variant_profile(name) → dict
- apply_variant(variant) → dict

DEB-03 core: sig_level → apt sources mapping (Pattern 3 from 04-RESEARCH.md):
  Required / RequiredDatabaseOptional → [signed-by=/etc/apt/trusted.gpg.d/NAME.asc]
  OptionalTrustAll / Never            → [trusted=yes]  + LOUD trust_warning

Security (T-04-07):
  sig_level Never/OptionalTrustAll → [trusted=yes] + LOUD warning comment in archive file.
  sig_level Required / RequiredDatabaseOptional → signed-by keyring (never silently trust).

ARCH-04 invariant applies to Debian: one code path, multiple profiles, zero per-variant
branches. All differences are in the YAML data; the code never branches on variant name.
"""

import os
import yaml  # PyYAML

# ---------------------------------------------------------------------------
# Profile directory (relative to this module file)
# ---------------------------------------------------------------------------

_PROFILES_DIR = os.path.join(os.path.dirname(__file__), "profiles")

# Keyring install path inside the target system (used in signed-by option)
_APT_TRUSTED_DIR = "/etc/apt/trusted.gpg.d"


# ---------------------------------------------------------------------------
# load_variant_profile
# ---------------------------------------------------------------------------

def load_variant_profile(name: str) -> dict:
    """Load a variant profile YAML by name.

    Args:
        name: Profile name, e.g. "debian", "ubuntu".
              Corresponds to profiles/<name>.yaml.

    Returns:
        The parsed profile dict.

    Raises:
        FileNotFoundError: if no profiles/<name>.yaml exists, with a clear
            message naming the missing profile.
        yaml.YAMLError: if the file is not valid YAML.
    """
    profile_path = os.path.join(_PROFILES_DIR, f"{name}.yaml")
    if not os.path.isfile(profile_path):
        available = _list_available_profiles()
        raise FileNotFoundError(
            f"Variant profile '{name}' not found. "
            f"Expected file: {profile_path}. "
            f"Available profiles: {available}"
        )
    with open(profile_path) as fh:
        profile = yaml.safe_load(fh)
    return profile


def _list_available_profiles() -> list:
    """List available profile names (without .yaml extension)."""
    if not os.path.isdir(_PROFILES_DIR):
        return []
    return [
        os.path.splitext(f)[0]
        for f in sorted(os.listdir(_PROFILES_DIR))
        if f.endswith(".yaml") and not f.startswith(".")
    ]


# ---------------------------------------------------------------------------
# apply_variant — apt sources mapping (DEB-03 core)
# ---------------------------------------------------------------------------

# sig_level values that require signed-by (verified signature required)
_SIGNED_BY_LEVELS = {"Required", "RequiredDatabaseOptional"}
# sig_level values that use trusted=yes (unsigned — emit LOUD warning)
_TRUSTED_YES_LEVELS = {"OptionalTrustAll", "Never"}


def apply_variant(variant: dict) -> dict:
    """Map a variant profile's custom repos to apt sources.list.d stanzas.

    DEB-03 sig_level → apt option mapping (Pattern 3 / 04-RESEARCH.md):
      Required                → [signed-by=/etc/apt/trusted.gpg.d/debateos-NAME.asc]
      RequiredDatabaseOptional → [signed-by=/etc/apt/trusted.gpg.d/debateos-NAME.asc]
      OptionalTrustAll         → [trusted=yes]  + trust_warning (T-04-07)
      Never                    → [trusted=yes]  + LOUD trust_warning + WARNING comment (T-04-07)

    Args:
        variant: Loaded variant profile dict (from load_variant_profile).

    Returns:
        A dict with:
          "apt_sources": list of dicts, one per custom repo. Each dict has:
            - "filename": str  — target file under config/archives/
              (e.g. "debateos-REPONAME.list.chroot_install")
            - "line": str      — the deb [...] <url> line
            - "comment": str   — inline comment to prepend (empty unless Never)
            - "keyring": str | None — keyring URL/path (for signed-by repos)
          "keyring_install_before_repos": list of dicts with {name, keyring} for
              repos that require a keyring to be installed before apt fetch (Pitfall 4).
          "trust_warnings": list of human-readable warning strings for repos with
              sig_level=OptionalTrustAll or Never (T-04-07, mirrors T-02-10 pattern).
    """
    variant_repos = variant.get("repos", []) or []
    keyring_pkgs = list(variant.get("keyring_install_before_repos", []) or [])

    apt_sources = []
    keyring_install_before_repos = list(keyring_pkgs)
    trust_warnings = []

    for repo in variant_repos:
        name = repo.get("name", "unknown")
        url = repo.get("url", "")
        sig = repo.get("sig_level", "Required")
        keyring = repo.get("keyring")  # URL or path — translator-interpreted (DEB-03 Finding 6)

        filename = f"debateos-{name}.list.chroot_install"
        comment = ""

        if sig in _SIGNED_BY_LEVELS:
            # Require signed keyring; keyring file will be placed in config/archives/NAME.key.chroot
            keyring_path = f"{_APT_TRUSTED_DIR}/debateos-{name}.asc"
            line = f"deb [signed-by={keyring_path}] {url}"
            # Record for keyring install ordering (Pitfall 4)
            if keyring:
                keyring_install_before_repos.append({
                    "name": name,
                    "keyring": keyring,
                    "keyring_file": f"debateos-{name}.key.chroot",
                })

        elif sig in _TRUSTED_YES_LEVELS:
            line = f"deb [trusted=yes] {url}"
            if sig == "Never":
                # LOUD WARNING comment for sig_level=Never (T-04-07)
                comment = (
                    f"# WARNING: sig_level=Never for repo '{name}' — "
                    f"signatures bypassed entirely; verify repo source before use (T-04-07)"
                )
                trust_warnings.append(
                    f"WARNING: repo '{name}' has sig_level=Never "
                    f"(variant {variant.get('variant', '?')}) — all package signatures bypassed; "
                    f"verify repo source before use (T-04-07)."
                )
            else:
                # OptionalTrustAll
                trust_warnings.append(
                    f"WARNING: repo '{name}' has sig_level=OptionalTrustAll "
                    f"(variant {variant.get('variant', '?')}) — unsigned packages accepted; "
                    f"verify repo source before use (T-04-07, WR-01)."
                )
        else:
            # Unknown sig_level — default to signed-by and emit a warning
            keyring_path = f"{_APT_TRUSTED_DIR}/debateos-{name}.asc"
            line = f"deb [signed-by={keyring_path}] {url}"
            trust_warnings.append(
                f"WARNING: repo '{name}' has unknown sig_level '{sig}'; defaulted to signed-by."
            )

        apt_sources.append({
            "filename": filename,
            "line": line,
            "comment": comment,
            "keyring": keyring if sig in _SIGNED_BY_LEVELS else None,
        })

    return {
        "apt_sources": apt_sources,
        "keyring_install_before_repos": keyring_install_before_repos,
        "trust_warnings": trust_warnings,
    }
