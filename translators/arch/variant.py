"""
variant.py — Variant profile loader and applicator for the Arch translator.

Provides:
- load_variant_profile(name) → dict
- apply_variant(variant, base_pacman_repos) → dict
- surface_conflicts(variant, applied_opinion_ids) → list[dict]

ARCH-04 invariant: one code path, three profiles, zero per-variant branches.
All differences are in the YAML data; the code never branches on variant name.

Security: T-02-10 — sig_level=Never repos emit trust_warnings in apply_variant output.
"""

import os
import yaml  # PyYAML

# ---------------------------------------------------------------------------
# Profile directory (relative to this module file)
# ---------------------------------------------------------------------------

_PROFILES_DIR = os.path.join(os.path.dirname(__file__), "profiles")


# ---------------------------------------------------------------------------
# load_variant_profile
# ---------------------------------------------------------------------------

def load_variant_profile(name: str) -> dict:
    """Load a variant profile YAML by name.

    Args:
        name: Profile name, e.g. "vanilla-arch", "cachyos", "garuda".
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
# apply_variant
# ---------------------------------------------------------------------------

def apply_variant(variant: dict, base_pacman_repos: list) -> dict:
    """Apply a variant profile onto the base pacman repo set.

    Produces an ordered repo list suitable for generating pacman.conf:
    - Repos with above_core=True are prepended BEFORE core/extra.
    - Repos with above_core=False (or absent) are appended AFTER core/extra.

    Keyring packages (keyring_install_before_repos) are returned unchanged
    to ensure the caller installs them first (Pitfall 4).

    Trust warnings are emitted for any repo with sig_level="Never" (T-02-10).

    Args:
        variant: Loaded variant profile dict (from load_variant_profile).
        base_pacman_repos: List of base repo dicts (e.g. core + extra).
            Each dict has at least "name" and "url" keys.

    Returns:
        A dict with:
          "repos": ordered list of repo dicts (above_core variant repos
              first, then base_pacman_repos, then below_core variant repos)
          "keyring_install_before_repos": list of keyring package names to
              install before enabling custom repos (Pitfall 4)
          "trust_warnings": list of human-readable warning strings for
              repos with sig_level="Never" (T-02-10)
    """
    variant_repos = variant.get("repos", []) or []
    keyring_pkgs = list(variant.get("keyring_install_before_repos", []) or [])

    above_core = []   # variant repos with above_core=True
    below_core = []   # variant repos with above_core=False or unset

    trust_warnings = []

    for repo in variant_repos:
        repo_copy = dict(repo)
        # Remove non-pacman-conf keys (YAML-specific metadata)
        repo_copy.pop("note", None)

        sig = repo.get("sig_level", "")
        if sig == "Never":
            trust_warnings.append(
                f"WARNING: repo '{repo['name']}' has sig_level=Never "
                f"(variant {variant.get('variant', '?')}) — unsigned packages accepted; "
                f"verify repo source before use."
            )

        if repo.get("above_core", False):
            above_core.append(repo_copy)
        else:
            below_core.append(repo_copy)

    ordered_repos = above_core + list(base_pacman_repos) + below_core

    return {
        "repos": ordered_repos,
        "keyring_install_before_repos": keyring_pkgs,
        "trust_warnings": trust_warnings,
    }


# ---------------------------------------------------------------------------
# surface_conflicts
# ---------------------------------------------------------------------------

def surface_conflicts(variant: dict, applied_opinion_ids: list) -> list:
    """Return the subset of conflicts_with_omarchy whose opinions are applied.

    Reads the conflicts_with_omarchy list from the variant profile YAML.
    Each conflict entry has an omarchy_opinions list; if any of those opinions
    is in applied_opinion_ids, the conflict is included in the returned list.

    Args:
        variant: Loaded variant profile dict (from load_variant_profile).
        applied_opinion_ids: List of opinion IDs that are applied in the
            current resolution (from ResolvedSpeech.applied).

    Returns:
        A list of conflict dicts, each with at least:
          "mechanism": str — the Garuda mechanism description
          "omarchy_opinions": list[str] — the Omarchy opinion IDs involved
        Only conflicts where at least one omarchy_opinion is in
        applied_opinion_ids are returned.
    """
    conflicts_list = variant.get("conflicts_with_omarchy", []) or []
    applied_set = set(applied_opinion_ids)

    surfaced = []
    for conflict in conflicts_list:
        affected_opinions = conflict.get("omarchy_opinions", [])
        # Include this conflict if any of its opinions are in the applied set.
        if any(op_id in applied_set for op_id in affected_opinions):
            surfaced.append(dict(conflict))

    return surfaced
