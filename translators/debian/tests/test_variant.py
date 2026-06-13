"""
test_variant.py — RED tests for Debian variant.py (DEB-03 core: apt sig_level mapping).

TDD RED phase: These tests are written BEFORE variant.py exists.
They MUST fail now and pass after implementation (GREEN).

Coverage:
- load_variant_profile("debian") loads profiles/debian.yaml
- load_variant_profile(unknown) raises FileNotFoundError naming the profile
- apply_variant maps sig_level Required → signed-by apt option
- apply_variant maps sig_level RequiredDatabaseOptional → signed-by apt option
- apply_variant maps sig_level OptionalTrustAll → [trusted=yes] + trust_warning
- apply_variant maps sig_level Never → [trusted=yes] + LOUD trust_warning comment
- Result keys: apt_sources (list), keyring_install_before_repos (list), trust_warnings (list)
"""

import pytest
import os

from variant import load_variant_profile, apply_variant


PROFILES_DIR = os.path.join(os.path.dirname(__file__), "..", "profiles")


# ---------------------------------------------------------------------------
# Test: load_variant_profile
# ---------------------------------------------------------------------------

class TestLoadVariantProfile:

    def test_load_debian(self):
        """debian profile loads without error and has required fields."""
        profile = load_variant_profile("debian")
        assert profile["variant"] == "debian"
        assert "repos" in profile
        assert "kernel" in profile
        assert "defaults" in profile

    def test_debian_has_initramfs_tools(self):
        """debian profile defaults.initramfs must be initramfs-tools (not mkinitcpio)."""
        profile = load_variant_profile("debian")
        assert profile["defaults"]["initramfs"] == "initramfs-tools"

    def test_debian_has_grub2(self):
        """debian profile defaults.bootloader must be grub2."""
        profile = load_variant_profile("debian")
        assert profile["defaults"]["bootloader"] == "grub2"

    def test_debian_has_kernel_package(self):
        """debian profile must declare a kernel package."""
        profile = load_variant_profile("debian")
        assert "package" in profile["kernel"]
        assert "linux" in profile["kernel"]["package"]

    def test_unknown_profile_raises_file_not_found(self):
        """Loading a missing profile raises FileNotFoundError naming it."""
        with pytest.raises(FileNotFoundError) as exc_info:
            load_variant_profile("nonexistent-distro")
        assert "nonexistent-distro" in str(exc_info.value)


# ---------------------------------------------------------------------------
# Test: apply_variant — apt_sources mapping
# ---------------------------------------------------------------------------

class TestApplyVariantAptSources:

    def _make_repo(self, name, url, sig_level, keyring="https://example.com/key.asc"):
        return {"name": name, "url": url, "sig_level": sig_level, "keyring": keyring}

    def test_result_has_required_keys(self):
        """apply_variant result must have apt_sources, keyring_install_before_repos, trust_warnings."""
        variant = load_variant_profile("debian")
        result = apply_variant(variant)
        assert "apt_sources" in result
        assert "keyring_install_before_repos" in result
        assert "trust_warnings" in result

    def test_empty_repos_produces_empty_apt_sources(self):
        """A variant with no custom repos produces an empty apt_sources list."""
        variant = load_variant_profile("debian")
        # debian.yaml has repos: []
        result = apply_variant(variant)
        assert result["apt_sources"] == []
        assert result["trust_warnings"] == []

    def test_required_sig_level_produces_signed_by(self):
        """sig_level=Required must produce a sources line with signed-by=..."""
        variant = load_variant_profile("debian")
        variant = dict(variant)
        variant["repos"] = [
            self._make_repo(
                "myrepo",
                "https://myrepo.example.com/debian stable main",
                "Required",
                "https://myrepo.example.com/key.asc",
            )
        ]
        result = apply_variant(variant)
        assert len(result["apt_sources"]) == 1
        src = result["apt_sources"][0]
        assert "signed-by" in src["line"], f"Expected signed-by in: {src['line']}"
        assert "trusted=yes" not in src["line"], f"Must NOT use trusted=yes for Required"

    def test_required_database_optional_produces_signed_by(self):
        """sig_level=RequiredDatabaseOptional must map same as Required (signed-by)."""
        variant = load_variant_profile("debian")
        variant = dict(variant)
        variant["repos"] = [
            self._make_repo(
                "dbrepo",
                "https://dbrepo.example.com/debian stable main",
                "RequiredDatabaseOptional",
                "https://dbrepo.example.com/key.asc",
            )
        ]
        result = apply_variant(variant)
        src = result["apt_sources"][0]
        assert "signed-by" in src["line"], f"Expected signed-by in: {src['line']}"
        assert "trusted=yes" not in src["line"]

    def test_optional_trust_all_produces_trusted_yes(self):
        """sig_level=OptionalTrustAll must produce a sources line with [trusted=yes]."""
        variant = load_variant_profile("debian")
        variant = dict(variant)
        variant["repos"] = [
            self._make_repo(
                "untrustedrepo",
                "https://untrusted.example.com/debian stable main",
                "OptionalTrustAll",
            )
        ]
        result = apply_variant(variant)
        src = result["apt_sources"][0]
        assert "trusted=yes" in src["line"], f"Expected trusted=yes in: {src['line']}"
        assert len(result["trust_warnings"]) >= 1, "OptionalTrustAll must emit trust_warning"

    def test_never_produces_trusted_yes_with_loud_warning(self):
        """sig_level=Never must produce [trusted=yes] AND a LOUD warning comment."""
        variant = load_variant_profile("debian")
        variant = dict(variant)
        variant["repos"] = [
            self._make_repo(
                "unsignedrepo",
                "https://unsigned.example.com/debian stable main",
                "Never",
            )
        ]
        result = apply_variant(variant)
        src = result["apt_sources"][0]
        assert "trusted=yes" in src["line"], f"Expected trusted=yes in: {src['line']}"
        # LOUD warning: must emit at least one trust_warning
        assert len(result["trust_warnings"]) >= 1
        # The warning must mention sig_level=Never
        any_never_warning = any("Never" in w for w in result["trust_warnings"])
        assert any_never_warning, f"Expected Never warning in: {result['trust_warnings']}"
        # The archive content (comment field) must also contain a LOUD WARNING comment
        # The comment key must exist and contain "WARNING" for Never sig_level
        assert "comment" in src, "Never sig_level sources must include a comment field"
        assert "WARNING" in src.get("comment", ""), (
            f"LOUD WARNING comment required for sig_level=Never, got: {src.get('comment')}"
        )

    def test_required_signed_by_uses_keyring_path(self):
        """Required repo signed-by path should reference /etc/apt/trusted.gpg.d/."""
        variant = load_variant_profile("debian")
        variant = dict(variant)
        variant["repos"] = [
            self._make_repo(
                "myrepo",
                "https://myrepo.example.com/debian stable main",
                "Required",
                "https://myrepo.example.com/key.asc",
            )
        ]
        result = apply_variant(variant)
        src = result["apt_sources"][0]
        assert "/etc/apt/trusted.gpg.d/" in src["line"], (
            f"signed-by must reference /etc/apt/trusted.gpg.d/: {src['line']}"
        )

    def test_required_repo_produces_keyring_install_entry(self):
        """Required repo must record a keyring entry in keyring_install_before_repos."""
        variant = load_variant_profile("debian")
        variant = dict(variant)
        variant["repos"] = [
            self._make_repo(
                "myrepo",
                "https://myrepo.example.com/debian stable main",
                "Required",
                "https://myrepo.example.com/key.asc",
            )
        ]
        result = apply_variant(variant)
        assert len(result["keyring_install_before_repos"]) >= 1

    def test_no_trust_warnings_for_required(self):
        """sig_level=Required must produce zero trust_warnings."""
        variant = load_variant_profile("debian")
        variant = dict(variant)
        variant["repos"] = [
            self._make_repo(
                "myrepo",
                "https://myrepo.example.com/debian stable main",
                "Required",
                "https://myrepo.example.com/key.asc",
            )
        ]
        result = apply_variant(variant)
        assert result["trust_warnings"] == [], f"Required must have no trust_warnings: {result['trust_warnings']}"

    def test_apply_variant_accepts_only_variant_arg(self):
        """apply_variant(variant) with no base_repos arg must succeed (Debian has no base repos concept)."""
        variant = load_variant_profile("debian")
        result = apply_variant(variant)
        assert isinstance(result, dict)


# ---------------------------------------------------------------------------
# Test: manifest reuse from common
# ---------------------------------------------------------------------------

class TestManifestReuse:
    """Verify that BuildManifest from common/ aggregates the df fixture correctly."""

    def test_build_manifest_from_resolved(self, tmp_path):
        """BuildManifest.from_resolved aggregates all 5 DF opinions correctly."""
        import json
        import sys
        import os
        # Ensure common is importable
        repo_root = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", ".."))
        if repo_root not in sys.path:
            sys.path.insert(0, repo_root)
        from translators.common.manifest import BuildManifest

        fixtures = os.path.join(os.path.dirname(__file__), "fixtures")
        with open(os.path.join(fixtures, "df_resolved.json"), "rb") as fh:
            resolved_bytes = fh.read()
        resolved = json.loads(resolved_bytes.decode("utf-8"))
        with open(os.path.join(fixtures, "df_opinions.json")) as fh:
            opinions_list = json.load(fh)
        opinions_index = {op["id"]: op for op in opinions_list}
        caps = {
            "install-packages", "deploy-config-file-tree", "enable-systemd-service",
            "write-sysctl-drop-in", "add-user-to-group",
        }

        manifest = BuildManifest.from_resolved(resolved, opinions_index, caps, resolved_bytes)

        # install_order preserved
        assert manifest.install_order == ["DF-001", "DF-002", "DF-003", "DF-004", "DF-005"]
        # packages deduped
        assert set(manifest.target_packages) == {"git", "curl", "vim"}
        # file_assets present
        assert len(manifest.file_assets) == 1
        assert manifest.file_assets[0]["dst"] == "etc/motd"
        # system_services present
        assert len(manifest.system_services) == 1
        assert manifest.system_services[0]["name"] == "systemd-timesyncd.service"
        # sysctl_params present
        assert len(manifest.sysctl_params) == 1
        assert manifest.sysctl_params[0]["key"] == "net.ipv4.tcp_fastopen"
        # group_memberships present
        assert len(manifest.group_memberships) == 1
        assert manifest.group_memberships[0]["group"] == "video"
        # foundation
        assert manifest.foundation == "debian"
