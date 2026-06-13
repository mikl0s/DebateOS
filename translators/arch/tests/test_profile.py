"""
test_profile.py — RED tests for profile.py: emit_profile_tree.

These tests MUST FAIL before implementation (TDD RED phase, D19).

Threat model: T-02-08 — file_asset dst path traversal test (ARCH-01 security gate).
"""

import json
import os
import stat
import tempfile
import pytest

from manifest import BuildManifest, derive_source_date_epoch
from contract import load_resolved_speech, load_opinion_bodies
from capabilities import load_capabilities
from variant import load_variant_profile, apply_variant
from profile import emit_profile_tree


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

_FIXTURES_DIR = os.path.join(os.path.dirname(__file__), "fixtures")


def _load_subset_manifest(
    resolved_name="omarchy_subset_resolved.json",
    opinions_name="omarchy_subset_opinions.json",
) -> BuildManifest:
    """Load the omarchy subset fixtures and build a BuildManifest."""
    resolved_path = os.path.join(_FIXTURES_DIR, resolved_name)
    opinions_path = os.path.join(_FIXTURES_DIR, opinions_name)
    resolved = load_resolved_speech(resolved_path)
    opinions = load_opinion_bodies(opinions_path)
    capabilities = load_capabilities()
    with open(resolved_path, "rb") as fh:
        resolved_bytes = fh.read()
    return BuildManifest.from_resolved(resolved, opinions, capabilities, resolved_bytes)


# ---------------------------------------------------------------------------
# Test: profile tree structure
# ---------------------------------------------------------------------------

class TestProfileTreeStructure:

    def test_profiledef_sh_created(self):
        """emit_profile_tree writes profiledef.sh at the profile root."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            assert os.path.isfile(os.path.join(out_dir, "profiledef.sh"))

    def test_packages_x86_64_created(self):
        """emit_profile_tree writes packages.x86_64 at the profile root."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            assert os.path.isfile(os.path.join(out_dir, "packages.x86_64"))

    def test_pacman_conf_created(self):
        """emit_profile_tree writes pacman.conf at the profile root."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            assert os.path.isfile(os.path.join(out_dir, "pacman.conf"))

    def test_installer_path(self):
        """emit_profile_tree writes debateos-install.sh at airootfs/root/."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            installer_path = os.path.join(out_dir, "airootfs", "root", "debateos-install.sh")
            assert os.path.isfile(installer_path), f"Installer not found at {installer_path}"

    def test_installer_is_0755(self):
        """debateos-install.sh is written with 0755 permissions."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            installer_path = os.path.join(out_dir, "airootfs", "root", "debateos-install.sh")
            mode = os.stat(installer_path).st_mode
            # Check execute bits for owner, group, others
            assert mode & stat.S_IXUSR, "Owner execute bit not set"
            assert mode & stat.S_IXGRP, "Group execute bit not set"
            assert mode & stat.S_IXOTH, "Other execute bit not set"

    def test_profiledef_contains_installer_permission_entry(self):
        """profiledef.sh file_permissions block includes the installer at 0:0:755."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            with open(os.path.join(out_dir, "profiledef.sh")) as fh:
                profiledef = fh.read()
            # Must contain the installer path with 0:0:755 permission entry
            assert "debateos-install.sh" in profiledef
            assert "0:0:755" in profiledef

    def test_zlogin_created(self):
        """emit_profile_tree writes .zlogin at airootfs/root/."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            zlogin_path = os.path.join(out_dir, "airootfs", "root", ".zlogin")
            assert os.path.isfile(zlogin_path), f".zlogin not found at {zlogin_path}"

    def test_zlogin_references_installer_and_tty1(self):
        """.zlogin references the installer script and /dev/tty1 (Pattern 1)."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            with open(os.path.join(out_dir, "airootfs", "root", ".zlogin")) as fh:
                zlogin = fh.read()
            assert "debateos-install.sh" in zlogin
            assert "tty1" in zlogin

    def test_build_manifest_json_created(self):
        """emit_profile_tree writes build-manifest.json at the profile root."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            bm_path = os.path.join(out_dir, "build-manifest.json")
            assert os.path.isfile(bm_path)

    def test_build_manifest_json_valid(self):
        """build-manifest.json is valid JSON with expected top-level keys."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            bm_path = os.path.join(out_dir, "build-manifest.json")
            with open(bm_path) as fh:
                bm = json.load(fh)
            # Must have the canonical BuildManifest keys
            assert "target_packages" in bm
            assert "file_assets" in bm
            assert "system_services" in bm


# ---------------------------------------------------------------------------
# Test: packages.x86_64 stays minimal (Pitfall 2 — target set in build-manifest)
# ---------------------------------------------------------------------------

class TestPackagesX86_64Minimal:

    def test_packages_x86_64_contains_live_env_packages(self):
        """packages.x86_64 contains live-env installer deps (arch-install-scripts etc.)."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            with open(os.path.join(out_dir, "packages.x86_64")) as fh:
                pkgs = fh.read()
            assert "arch-install-scripts" in pkgs

    def test_packages_x86_64_does_not_contain_target_packages(self):
        """packages.x86_64 does NOT contain opinion target packages (Pitfall 2).

        The target package set lives in build-manifest.json, not packages.x86_64.
        This test ensures the live ISO stays minimal.
        """
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            with open(os.path.join(out_dir, "packages.x86_64")) as fh:
                pkgs = fh.read()
            # hyprland is in the subset opinions target_packages — must NOT be in live env
            assert "hyprland" not in pkgs, (
                "hyprland is a target package that must be in build-manifest.json, "
                "not packages.x86_64 (Pitfall 2)"
            )

    def test_build_manifest_contains_target_packages(self):
        """build-manifest.json contains the opinion target packages (e.g. hyprland)."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            bm_path = os.path.join(out_dir, "build-manifest.json")
            with open(bm_path) as fh:
                bm = json.load(fh)
            assert "hyprland" in bm["target_packages"], (
                "hyprland must be in build-manifest.json target_packages"
            )


# ---------------------------------------------------------------------------
# Test: first-run units emitted (T-02-11)
# ---------------------------------------------------------------------------

class TestFirstRunUnits:

    def test_firstrun_unit_files_emitted(self):
        """emit_profile_tree creates first-run unit files for first-run opinions."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            user_unit_dir = os.path.join(out_dir, "airootfs", "etc", "systemd", "user")
            assert os.path.isdir(user_unit_dir), f"systemd/user dir not found at {user_unit_dir}"
            unit_files = os.listdir(user_unit_dir)
            assert len(unit_files) >= 1, "Expected at least one first-run unit file"

    def test_firstrun_unit_named_correctly(self):
        """First-run unit filename follows debateos-firstrun-<id>.service pattern."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            user_unit_dir = os.path.join(out_dir, "airootfs", "etc", "systemd", "user")
            unit_files = os.listdir(user_unit_dir)
            # OM-102 is a first-run opinion in the fixture
            assert any("OM-102" in f for f in unit_files), (
                f"Expected OM-102 unit in {unit_files}"
            )

    def test_firstrun_unit_has_flag_file_condition(self):
        """First-run unit contains ConditionPathExists=! guard."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            user_unit_dir = os.path.join(out_dir, "airootfs", "etc", "systemd", "user")
            # Find the OM-102 unit file
            unit_files = [
                f for f in os.listdir(user_unit_dir) if "OM-102" in f
            ]
            assert unit_files, "OM-102 first-run unit not found"
            with open(os.path.join(user_unit_dir, unit_files[0])) as fh:
                unit_content = fh.read()
            assert "ConditionPathExists=!" in unit_content
            assert ".firstrun-" in unit_content


# ---------------------------------------------------------------------------
# Test: profiledef.sh correctness
# ---------------------------------------------------------------------------

class TestProfiledefSh:

    def test_profiledef_iso_name(self):
        """profiledef.sh sets iso_name=debateos."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            with open(os.path.join(out_dir, "profiledef.sh")) as fh:
                profiledef = fh.read()
            assert 'iso_name="debateos"' in profiledef

    def test_profiledef_bootmodes(self):
        """profiledef.sh declares both bios.syslinux and uefi.systemd-boot bootmodes."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            with open(os.path.join(out_dir, "profiledef.sh")) as fh:
                profiledef = fh.read()
            assert "bios.syslinux" in profiledef
            assert "uefi.systemd-boot" in profiledef

    def test_profiledef_deterministic(self):
        """Two calls with the same manifest produce identical profiledef.sh content."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out1, tempfile.TemporaryDirectory() as out2:
            emit_profile_tree(out1, manifest, variant)
            emit_profile_tree(out2, manifest, variant)
            with open(os.path.join(out1, "profiledef.sh")) as fh:
                content1 = fh.read()
            with open(os.path.join(out2, "profiledef.sh")) as fh:
                content2 = fh.read()
            assert content1 == content2, "profiledef.sh must be deterministic"


# ---------------------------------------------------------------------------
# Test: pacman.conf variant repo injection
# ---------------------------------------------------------------------------

class TestPacmanConf:

    def test_pacman_conf_contains_core_repo(self):
        """pacman.conf always contains [core] repo section."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            with open(os.path.join(out_dir, "pacman.conf")) as fh:
                pacman_conf = fh.read()
            assert "[core]" in pacman_conf

    def test_pacman_conf_cachyos_repos_injected(self):
        """CachyOS repos appear in pacman.conf when cachyos variant is used."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("cachyos")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            with open(os.path.join(out_dir, "pacman.conf")) as fh:
                pacman_conf = fh.read()
            assert "[cachyos]" in pacman_conf

    def test_pacman_conf_cachyos_before_core(self):
        """CachyOS [cachyos] repo appears before [core] in pacman.conf."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("cachyos")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            with open(os.path.join(out_dir, "pacman.conf")) as fh:
                pacman_conf = fh.read()
            cachyos_pos = pacman_conf.find("[cachyos]")
            core_pos = pacman_conf.find("[core]")
            assert cachyos_pos != -1
            assert core_pos != -1
            assert cachyos_pos < core_pos, "cachyos repo must appear before core in pacman.conf"


# ---------------------------------------------------------------------------
# Test: installer script — jq-driven, no shell injection
# ---------------------------------------------------------------------------

class TestInstallerScript:

    def test_installer_references_build_manifest(self):
        """Installer script references build-manifest.json (jq-driven, Pitfall 6)."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            with open(os.path.join(out_dir, "airootfs", "root", "debateos-install.sh")) as fh:
                installer = fh.read()
            assert "build-manifest" in installer or "build_manifest" in installer

    def test_installer_uses_noconfirm(self):
        """Installer script uses --noconfirm in pacman/pacstrap calls."""
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            with open(os.path.join(out_dir, "airootfs", "root", "debateos-install.sh")) as fh:
                installer = fh.read()
            assert "--noconfirm" in installer

    def test_installer_nvme_partition_naming(self):
        """Installer uses 'p' suffix for NVMe/loop devices (CR-03).

        For disk names ending in a digit (nvme0n1, loop0), partitions must be
        named <disk>p1 / <disk>p2.  For SATA disks (sda, vda) partitions are
        <disk>1 / <disk>2.  The template must contain the branching logic.
        """
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            with open(os.path.join(out_dir, "airootfs", "root", "debateos-install.sh")) as fh:
                installer = fh.read()
            # Template must contain the nvme/loop regex branch
            assert "nvme" in installer, "installer must detect nvme disk type"
            # And produce the p-suffix partition naming
            assert "p1" in installer, "installer must have p-suffix partition naming for nvme"

    def test_installer_custom_repos_section(self):
        """Installer configures custom repos from build-manifest.json before pacstrap (CR-01).

        The installer must read custom_repos from the manifest and append them to
        /etc/pacman.conf so pacstrap can fetch packages from those repos.
        """
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            with open(os.path.join(out_dir, "airootfs", "root", "debateos-install.sh")) as fh:
                installer = fh.read()
            # Must reference custom_repos from manifest
            assert "custom_repos" in installer, (
                "installer must read custom_repos from build-manifest.json"
            )
            # Must append to pacman.conf
            assert "pacman.conf" in installer, (
                "installer must append custom repos to /etc/pacman.conf"
            )

    def test_installer_bootloader_section(self):
        """Installer contains a bootloader installation block (CR-02).

        The installer must install a bootloader after mkinitcpio.  For a limine
        manifest the limine block must appear; for a default manifest a systemd-boot
        fallback must be present.
        """
        manifest = _load_subset_manifest()
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            with open(os.path.join(out_dir, "airootfs", "root", "debateos-install.sh")) as fh:
                installer = fh.read()
            # bootctl install is the default fallback
            assert "bootctl" in installer, "installer must have systemd-boot fallback bootloader install"


# ---------------------------------------------------------------------------
# Test: file_asset dst path sanitization (T-02-08 security gate, ARCH-01)
# ---------------------------------------------------------------------------

class TestFileAssetDstSanitization:

    def _manifest_with_traversal_dst(self, dst_value: str) -> BuildManifest:
        """Create a minimal manifest with a file_asset that has a traversal dst."""
        resolved = {
            "schema": 1,
            "foundation": "arch",
            "applied": ["OM-006"],
            "skipped": [],
            "dropped": [],
            "install_order": ["OM-006"],
            "explanations": [],
        }
        opinions = {
            "OM-006": {
                "id": "OM-006",
                "name": "Traversal test",
                "category": "config-dotfile",
                "status": "required",
                "translator_capabilities": ["install-packages"],
                "packages": [],
                "file_assets": [
                    {"src": "config/evil", "dst": dst_value, "mode": "0644"}
                ],
            }
        }
        capabilities = load_capabilities()
        resolved_bytes = b'{"traversal-test": true}'
        return BuildManifest.from_resolved(resolved, opinions, capabilities, resolved_bytes)

    def test_absolute_dst_raises(self):
        """A file_asset with absolute dst path (e.g. /etc/passwd) raises ValueError."""
        manifest = self._manifest_with_traversal_dst("/etc/passwd")
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            with pytest.raises(ValueError) as exc_info:
                emit_profile_tree(out_dir, manifest, variant)
            assert "passwd" in str(exc_info.value) or "/etc/passwd" in str(exc_info.value)

    def test_dotdot_traversal_dst_raises(self):
        """A file_asset with .. traversal dst raises ValueError."""
        manifest = self._manifest_with_traversal_dst("../../etc/passwd")
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            with pytest.raises(ValueError) as exc_info:
                emit_profile_tree(out_dir, manifest, variant)
            # Error message should identify the offending dst
            assert ".." in str(exc_info.value) or "traversal" in str(exc_info.value).lower() or "passwd" in str(exc_info.value)

    def test_valid_relative_dst_accepted(self):
        """A valid relative dst path (e.g. home/user/.config) is accepted without error."""
        resolved = {
            "schema": 1,
            "foundation": "arch",
            "applied": ["OM-006"],
            "skipped": [],
            "dropped": [],
            "install_order": ["OM-006"],
            "explanations": [],
        }
        opinions = {
            "OM-006": {
                "id": "OM-006",
                "name": "Valid dst test",
                "category": "config-dotfile",
                "status": "required",
                "translator_capabilities": ["install-packages"],
                "packages": [],
                "file_assets": [
                    {"src": "config/good", "dst": "home/user/.config/good", "mode": "0644"}
                ],
            }
        }
        capabilities = load_capabilities()
        resolved_bytes = b'{"valid-test": true}'
        manifest = BuildManifest.from_resolved(resolved, opinions, capabilities, resolved_bytes)
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            # Should not raise
            emit_profile_tree(out_dir, manifest, variant)

    def test_traversal_with_leading_slash(self):
        """A file_asset with leading slash in dst raises ValueError."""
        manifest = self._manifest_with_traversal_dst("/home/user/.bashrc")
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            with pytest.raises(ValueError):
                emit_profile_tree(out_dir, manifest, variant)

    def test_empty_dst_raises(self):
        """A file_asset with empty dst raises ValueError (WR-02)."""
        manifest = self._manifest_with_traversal_dst("")
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            with pytest.raises(ValueError, match=r"empty|concrete"):
                emit_profile_tree(out_dir, manifest, variant)

    def test_dot_dst_raises(self):
        """A file_asset with dst='.' raises ValueError (WR-02)."""
        manifest = self._manifest_with_traversal_dst(".")
        variant = load_variant_profile("vanilla-arch")
        with tempfile.TemporaryDirectory() as out_dir:
            with pytest.raises(ValueError, match=r"empty|concrete|\\."):
                emit_profile_tree(out_dir, manifest, variant)

    def test_traversal_dst_no_files_written(self):
        """When file_asset dst traversal is detected, NO files must be written (CR-04 fail-fast).

        Validation runs BEFORE any file I/O so the output directory remains
        empty/absent on a path-traversal error.
        """
        import shutil
        manifest = self._manifest_with_traversal_dst("../../etc/passwd")
        variant = load_variant_profile("vanilla-arch")
        out_dir = tempfile.mkdtemp()
        try:
            with pytest.raises(ValueError):
                emit_profile_tree(out_dir, manifest, variant)
            # The out_dir was created by mkdtemp; after the failed call it must
            # contain NO profile files (profiledef.sh is the first file written).
            files = []
            for root, dirs, fnames in os.walk(out_dir):
                files.extend(fnames)
            assert not files, (
                f"Expected no files after fail-fast validation, found: {files}"
            )
        finally:
            shutil.rmtree(out_dir, ignore_errors=True)
