"""
test_variant.py — RED tests for variant.py: load_variant_profile, apply_variant,
surface_conflicts, ARCH-04 no-fork invariant.

These tests MUST FAIL before implementation (TDD RED phase, D19).
"""

import pytest
import sys
import os

# Ensure translators/arch is on the path (pytest.ini pythonpath = .)
# from the translators/arch directory.

# ---------------------------------------------------------------------------
# Import the modules under test (will fail if not implemented yet — RED)
# ---------------------------------------------------------------------------

from variant import (
    load_variant_profile,
    apply_variant,
    surface_conflicts,
)


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------

PROFILE_NAMES = ["vanilla-arch", "cachyos", "garuda"]

# A minimal set of Omarchy opinion IDs that trigger Garuda conflicts.
# From garuda.yaml conflicts_with_omarchy, the affected opinions are:
# OM-002, OM-086, OM-099, OM-091, OM-097 — at least 4 unique opinions.
OMARCHY_SUBSET_IDS = ["OM-002", "OM-086", "OM-091", "OM-097", "OM-099"]

BASE_PACMAN_REPOS = [
    {"name": "core", "url": "https://geo.mirror.pkgbuild.com/$repo/os/$arch", "sig_level": "Required"},
    {"name": "extra", "url": "https://geo.mirror.pkgbuild.com/$repo/os/$arch", "sig_level": "Required"},
]


# ---------------------------------------------------------------------------
# Test: load_variant_profile
# ---------------------------------------------------------------------------

class TestLoadVariantProfile:

    def test_load_vanilla_arch(self):
        """vanilla-arch profile loads without error and has required fields."""
        profile = load_variant_profile("vanilla-arch")
        assert profile["variant"] == "vanilla-arch"
        assert "repos" in profile
        assert "keyring_install_before_repos" in profile
        assert "kernel" in profile

    def test_load_cachyos(self):
        """cachyos profile loads and has at least one repo."""
        profile = load_variant_profile("cachyos")
        assert profile["variant"] == "cachyos"
        assert len(profile["repos"]) >= 1

    def test_load_garuda(self):
        """garuda profile loads and has keyring packages."""
        profile = load_variant_profile("garuda")
        assert profile["variant"] == "garuda"
        assert len(profile["keyring_install_before_repos"]) >= 1

    def test_unknown_profile_raises(self):
        """Loading an unknown profile raises a clear error."""
        with pytest.raises(Exception) as exc_info:
            load_variant_profile("nonexistent-distro")
        # Error message should name the profile
        assert "nonexistent-distro" in str(exc_info.value)


# ---------------------------------------------------------------------------
# Test: apply_variant — CachyOS repo ordering + keyring-first
# ---------------------------------------------------------------------------

class TestApplyVariantCachyos:

    def test_cachyos_repos_injected(self):
        """CachyOS repos appear in the output repo section."""
        variant = load_variant_profile("cachyos")
        result = apply_variant(variant, BASE_PACMAN_REPOS)
        repo_names = [r["name"] for r in result["repos"]]
        assert "cachyos" in repo_names

    def test_cachyos_keyring_before_repos(self):
        """cachyos-keyring is in the keyring_install_before_repos list."""
        variant = load_variant_profile("cachyos")
        result = apply_variant(variant, BASE_PACMAN_REPOS)
        assert "cachyos-keyring" in result["keyring_install_before_repos"]

    def test_cachyos_above_core_ordering(self):
        """CachyOS repos with above_core=true appear BEFORE core in the repo list."""
        variant = load_variant_profile("cachyos")
        result = apply_variant(variant, BASE_PACMAN_REPOS)
        repo_names = [r["name"] for r in result["repos"]]
        assert "cachyos" in repo_names
        assert "core" in repo_names
        cachyos_idx = repo_names.index("cachyos")
        core_idx = repo_names.index("core")
        assert cachyos_idx < core_idx, (
            f"cachyos (idx={cachyos_idx}) must appear before core (idx={core_idx})"
        )

    def test_vanilla_arch_no_extra_repos(self):
        """vanilla-arch adds no repos beyond the base set."""
        variant = load_variant_profile("vanilla-arch")
        result = apply_variant(variant, BASE_PACMAN_REPOS)
        repo_names = [r["name"] for r in result["repos"]]
        assert "cachyos" not in repo_names
        assert "chaotic-aur" not in repo_names

    def test_vanilla_arch_no_keyring_packages(self):
        """vanilla-arch has empty keyring_install_before_repos."""
        variant = load_variant_profile("vanilla-arch")
        result = apply_variant(variant, BASE_PACMAN_REPOS)
        assert result["keyring_install_before_repos"] == []


# ---------------------------------------------------------------------------
# Test: apply_variant — Garuda ordering (above_core=false → AFTER core/extra)
# ---------------------------------------------------------------------------

class TestApplyVariantGaruda:

    def test_garuda_repos_injected(self):
        """Garuda repos (chaotic-aur, garuda) appear in the output repo section."""
        variant = load_variant_profile("garuda")
        result = apply_variant(variant, BASE_PACMAN_REPOS)
        repo_names = [r["name"] for r in result["repos"]]
        assert "chaotic-aur" in repo_names

    def test_garuda_above_core_false_ordering(self):
        """Garuda repos (above_core=false) appear AFTER core/extra in the repo list."""
        variant = load_variant_profile("garuda")
        result = apply_variant(variant, BASE_PACMAN_REPOS)
        repo_names = [r["name"] for r in result["repos"]]
        core_idx = repo_names.index("core")
        chaotic_idx = repo_names.index("chaotic-aur")
        assert chaotic_idx > core_idx, (
            f"chaotic-aur (idx={chaotic_idx}) must appear after core (idx={core_idx})"
        )

    def test_garuda_keyring_present(self):
        """Garuda keyring (chaotic-keyring) is in keyring_install_before_repos."""
        variant = load_variant_profile("garuda")
        result = apply_variant(variant, BASE_PACMAN_REPOS)
        assert "chaotic-keyring" in result["keyring_install_before_repos"]


# ---------------------------------------------------------------------------
# Test: ARCH-04 no-fork invariant — same apply_variant for all three profiles
# ---------------------------------------------------------------------------

class TestAllVariantsNoFork:
    """
    All three variant profiles must pass through apply_variant without
    raising — proving ARCH-04: one code path, three profiles, no name-branching.
    """

    @pytest.mark.parametrize("profile_name", PROFILE_NAMES)
    def test_all_variants_apply_without_error(self, profile_name):
        """apply_variant succeeds for every declared variant profile."""
        variant = load_variant_profile(profile_name)
        result = apply_variant(variant, BASE_PACMAN_REPOS)
        # Every result must have a 'repos' key with at least the base repos
        assert "repos" in result
        assert "keyring_install_before_repos" in result
        # core must always be present (from BASE_PACMAN_REPOS)
        repo_names = [r["name"] for r in result["repos"]]
        assert "core" in repo_names

    @pytest.mark.parametrize("profile_name", PROFILE_NAMES)
    def test_all_variants_produce_repo_section(self, profile_name):
        """apply_variant returns a non-empty repos list for every profile."""
        variant = load_variant_profile(profile_name)
        result = apply_variant(variant, BASE_PACMAN_REPOS)
        assert len(result["repos"]) >= len(BASE_PACMAN_REPOS)


# ---------------------------------------------------------------------------
# Test: surface_conflicts — Garuda returns >=4 when Omarchy opinions applied
# ---------------------------------------------------------------------------

class TestSurfaceConflicts:

    def test_garuda_conflicts_returned_for_omarchy_opinions(self):
        """Garuda + Omarchy subset returns at least 4 conflict entries."""
        variant = load_variant_profile("garuda")
        conflicts = surface_conflicts(variant, OMARCHY_SUBSET_IDS)
        assert len(conflicts) >= 4, (
            f"Expected >= 4 Garuda/Omarchy conflicts, got {len(conflicts)}: {conflicts}"
        )

    def test_vanilla_arch_no_conflicts(self):
        """vanilla-arch returns empty conflicts regardless of applied opinions."""
        variant = load_variant_profile("vanilla-arch")
        conflicts = surface_conflicts(variant, OMARCHY_SUBSET_IDS)
        assert conflicts == []

    def test_cachyos_no_conflicts_with_omarchy(self):
        """cachyos profile returns no hard conflicts with Omarchy opinions."""
        variant = load_variant_profile("cachyos")
        conflicts = surface_conflicts(variant, OMARCHY_SUBSET_IDS)
        assert conflicts == []

    def test_garuda_empty_applied_returns_no_conflicts(self):
        """surface_conflicts returns [] when no opinions are applied."""
        variant = load_variant_profile("garuda")
        conflicts = surface_conflicts(variant, [])
        assert conflicts == []

    def test_garuda_conflict_entries_have_mechanism_and_opinions(self):
        """Each returned conflict entry has 'mechanism' and 'omarchy_opinions' keys."""
        variant = load_variant_profile("garuda")
        conflicts = surface_conflicts(variant, OMARCHY_SUBSET_IDS)
        for c in conflicts:
            assert "mechanism" in c, f"Missing 'mechanism' in conflict: {c}"
            assert "omarchy_opinions" in c, f"Missing 'omarchy_opinions' in conflict: {c}"

    def test_garuda_conflict_includes_dracut_mechanism(self):
        """Garuda conflicts include the dracut/mkinitcpio hard conflict for Omarchy."""
        variant = load_variant_profile("garuda")
        conflicts = surface_conflicts(variant, OMARCHY_SUBSET_IDS)
        mechanisms = [c.get("mechanism", "") for c in conflicts]
        assert any("dracut" in m for m in mechanisms), (
            f"Expected dracut conflict in mechanisms: {mechanisms}"
        )


# ---------------------------------------------------------------------------
# Test: trust_warning surfacing for non-Required sig_level repos
# ---------------------------------------------------------------------------

class TestTrustWarnings:

    def test_apply_variant_returns_trust_warnings_key(self):
        """apply_variant result includes a 'trust_warnings' key."""
        variant = load_variant_profile("cachyos")
        result = apply_variant(variant, BASE_PACMAN_REPOS)
        assert "trust_warnings" in result

    def test_no_trust_warnings_for_required_sig_level(self):
        """CachyOS repos with sig_level='Required DatabaseOptional' produce no trust warnings."""
        variant = load_variant_profile("cachyos")
        result = apply_variant(variant, BASE_PACMAN_REPOS)
        # CachyOS uses 'Required DatabaseOptional', not 'Never' — no trust warnings expected.
        # (Trust warnings are for sig_level=Never only, per T-02-10)
        never_warnings = [
            w for w in result["trust_warnings"]
            if "Never" in str(w)
        ]
        # sig_level=Never repos produce a warning; Required DatabaseOptional does NOT.
        # CachyOS has Required DatabaseOptional — should produce zero Never warnings.
        assert len(never_warnings) == 0
