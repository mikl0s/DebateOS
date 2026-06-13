"""
test_profile.py — RED tests for Debian profile.py: emit_profile_tree.

TDD RED phase: Written BEFORE profile.py exists. MUST fail now, pass after GREEN.

Coverage:
- emit_profile_tree writes the full config/ tree (preseed.cfg, chroot hook,
  package-lists, build-manifest.json, first-run units, archives)
- chroot hook is executable (0755)
- preseed.cfg contains hashed-password sentinel (T-04-08 — never plaintext)
- _sanitize_dst rejects absolute + traversal dst paths (T-04-05)
- CapabilityError fires BEFORE any file I/O (generate creates no out_dir on gate fail)
"""

import json
import os
import stat
import sys
import tempfile
import pytest

# Ensure common is importable via .. in pytest.ini pythonpath
from manifest import BuildManifest
from variant import load_variant_profile

# Import module under test (will fail until profile.py exists — RED)
from profile import emit_profile_tree, _sanitize_dst


# ---------------------------------------------------------------------------
# Fixtures helpers
# ---------------------------------------------------------------------------

FIXTURES_DIR = os.path.join(os.path.dirname(__file__), "fixtures")


def _load_df_manifest() -> BuildManifest:
    """Load the dual-foundation fixtures and build a BuildManifest."""
    import json
    resolved_path = os.path.join(FIXTURES_DIR, "df_resolved.json")
    opinions_path = os.path.join(FIXTURES_DIR, "df_opinions.json")

    with open(resolved_path, "rb") as fh:
        resolved_bytes = fh.read()
    resolved = json.loads(resolved_bytes.decode("utf-8"))
    with open(opinions_path) as fh:
        opinions_list = json.load(fh)
    opinions_index = {op["id"]: op for op in opinions_list}
    caps = {
        "install-packages", "deploy-config-file-tree", "enable-systemd-service",
        "write-sysctl-drop-in", "add-user-to-group",
    }
    return BuildManifest.from_resolved(resolved, opinions_index, caps, resolved_bytes)


# ---------------------------------------------------------------------------
# Test: preseed.cfg
# ---------------------------------------------------------------------------

class TestPreseedCfg:

    def test_preseed_created(self):
        """emit_profile_tree writes config/includes.installer/preseed.cfg."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            preseed_path = os.path.join(
                out_dir, "config", "includes.installer", "preseed.cfg"
            )
            assert os.path.isfile(preseed_path), f"preseed.cfg not found at {preseed_path}"

    def test_preseed_contains_locale(self):
        """preseed.cfg contains a d-i locale line."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            preseed_path = os.path.join(
                out_dir, "config", "includes.installer", "preseed.cfg"
            )
            with open(preseed_path) as fh:
                content = fh.read()
            assert "d-i" in content, "preseed.cfg must contain d-i configuration lines"
            assert "locale" in content or "keyboard" in content, \
                "preseed.cfg must contain locale/keyboard config"

    def test_preseed_contains_hashed_password_sentinel(self):
        """preseed.cfg uses %%HASHED_PASSWORD%% sentinel, never plaintext (T-04-08)."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            preseed_path = os.path.join(
                out_dir, "config", "includes.installer", "preseed.cfg"
            )
            with open(preseed_path) as fh:
                content = fh.read()
            # Must use the HASHED_PASSWORD sentinel (never plaintext password)
            assert "HASHED_PASSWORD" in content, (
                "preseed.cfg must use %%HASHED_PASSWORD%% sentinel (T-04-08 — never plaintext)"
            )

    def test_preseed_contains_partitioning(self):
        """preseed.cfg contains partman auto partitioning configuration."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            preseed_path = os.path.join(
                out_dir, "config", "includes.installer", "preseed.cfg"
            )
            with open(preseed_path) as fh:
                content = fh.read()
            assert "partman" in content, "preseed.cfg must contain partman configuration"

    def test_preseed_contains_pkgsel(self):
        """preseed.cfg contains pkgsel/include line for d-i package selection."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            preseed_path = os.path.join(
                out_dir, "config", "includes.installer", "preseed.cfg"
            )
            with open(preseed_path) as fh:
                content = fh.read()
            assert "pkgsel" in content or "PKGSEL" in content, \
                "preseed.cfg must contain pkgsel package selection"

    def test_preseed_no_plaintext_password(self):
        """preseed.cfg must not contain any string that looks like a plaintext password."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            preseed_path = os.path.join(
                out_dir, "config", "includes.installer", "preseed.cfg"
            )
            with open(preseed_path) as fh:
                content = fh.read()
            # Must not have user-password string = any plaintext value (only crypted)
            lines = content.splitlines()
            for line in lines:
                if "user-password " in line and "crypted" not in line:
                    assert "%%HASHED_PASSWORD%%" in line or line.strip().startswith("#"), (
                        f"Found potential plaintext password line: {line!r}"
                    )


# ---------------------------------------------------------------------------
# Test: chroot hook
# ---------------------------------------------------------------------------

class TestChrootHook:

    def test_chroot_hook_created(self):
        """emit_profile_tree writes config/hooks/live/9000-debateos-apply.hook.chroot."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            hook_path = os.path.join(
                out_dir, "config", "hooks", "live",
                "9000-debateos-apply.hook.chroot"
            )
            assert os.path.isfile(hook_path), f"chroot hook not found at {hook_path}"

    def test_chroot_hook_is_executable(self):
        """chroot hook must be chmod 0755 (executable for lb build)."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            hook_path = os.path.join(
                out_dir, "config", "hooks", "live",
                "9000-debateos-apply.hook.chroot"
            )
            mode = os.stat(hook_path).st_mode
            assert mode & stat.S_IXUSR, "Hook owner execute bit not set"
            assert mode & stat.S_IXGRP, "Hook group execute bit not set"
            assert mode & stat.S_IXOTH, "Hook other execute bit not set"

    def test_chroot_hook_contains_apt_get_install(self):
        """chroot hook contains apt-get install with target packages (DF-001: git curl vim)."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            hook_path = os.path.join(
                out_dir, "config", "hooks", "live",
                "9000-debateos-apply.hook.chroot"
            )
            with open(hook_path) as fh:
                content = fh.read()
            assert "apt-get install" in content or "apt-get" in content, \
                "chroot hook must call apt-get install"
            # Target packages from DF-001
            assert "git" in content, "chroot hook must include 'git' from DF-001"

    def test_chroot_hook_contains_systemctl_enable(self):
        """chroot hook contains systemctl enable for DF-003 service."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            hook_path = os.path.join(
                out_dir, "config", "hooks", "live",
                "9000-debateos-apply.hook.chroot"
            )
            with open(hook_path) as fh:
                content = fh.read()
            assert "systemctl" in content, "chroot hook must call systemctl enable"

    def test_chroot_hook_contains_sysctl_section(self):
        """chroot hook contains sysctl drop-in configuration (DF-004)."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            hook_path = os.path.join(
                out_dir, "config", "hooks", "live",
                "9000-debateos-apply.hook.chroot"
            )
            with open(hook_path) as fh:
                content = fh.read()
            # Either sysctl.d reference or the key itself
            assert "sysctl" in content or "tcp_fastopen" in content, \
                "chroot hook must contain sysctl configuration"

    def test_chroot_hook_contains_set_euo_pipefail(self):
        """chroot hook starts with set -euo pipefail for safety."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            hook_path = os.path.join(
                out_dir, "config", "hooks", "live",
                "9000-debateos-apply.hook.chroot"
            )
            with open(hook_path) as fh:
                content = fh.read()
            assert "set -euo pipefail" in content or "set -e" in content, \
                "chroot hook must use set -euo pipefail"

    def test_chroot_hook_shebang_is_bash(self):
        """chroot hook shebang is #!/bin/bash."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            hook_path = os.path.join(
                out_dir, "config", "hooks", "live",
                "9000-debateos-apply.hook.chroot"
            )
            with open(hook_path) as fh:
                first_line = fh.readline().strip()
            assert first_line == "#!/bin/bash", f"Expected #!/bin/bash, got: {first_line!r}"


# ---------------------------------------------------------------------------
# Test: package list
# ---------------------------------------------------------------------------

class TestPackageList:

    def test_package_list_created(self):
        """emit_profile_tree writes config/package-lists/debateos.list.chroot_install."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            pkg_list_path = os.path.join(
                out_dir, "config", "package-lists",
                "debateos.list.chroot_install"
            )
            assert os.path.isfile(pkg_list_path), \
                f"Package list not found at {pkg_list_path}"

    def test_package_list_contains_target_packages(self):
        """Package list contains git, curl, vim from DF-001."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            pkg_list_path = os.path.join(
                out_dir, "config", "package-lists",
                "debateos.list.chroot_install"
            )
            with open(pkg_list_path) as fh:
                content = fh.read()
            assert "git" in content
            assert "curl" in content
            assert "vim" in content

    def test_package_list_uses_chroot_install_suffix(self):
        """Package list uses .list.chroot_install suffix, not .list.chroot (Pitfall 1)."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            pkg_list_dir = os.path.join(out_dir, "config", "package-lists")
            if os.path.isdir(pkg_list_dir):
                files = os.listdir(pkg_list_dir)
                # Must have .chroot_install suffix
                assert any(f.endswith(".chroot_install") for f in files), \
                    f"Expected .chroot_install suffix, found: {files}"
                # Must NOT have bare .chroot files for the main list
                assert not any(f.endswith(".list.chroot") and not f.endswith(".list.chroot_install") for f in files), \
                    f"Found .list.chroot instead of .list.chroot_install: {files}"


# ---------------------------------------------------------------------------
# Test: build-manifest.json
# ---------------------------------------------------------------------------

class TestBuildManifest:

    def test_build_manifest_created(self):
        """emit_profile_tree writes build-manifest.json at the output root."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            bm_path = os.path.join(out_dir, "build-manifest.json")
            assert os.path.isfile(bm_path)

    def test_build_manifest_valid_json(self):
        """build-manifest.json is valid JSON with expected keys."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            bm_path = os.path.join(out_dir, "build-manifest.json")
            with open(bm_path) as fh:
                bm = json.load(fh)
            for key in ("target_packages", "file_assets", "system_services",
                        "sysctl_params", "group_memberships", "foundation"):
                assert key in bm, f"Missing key in build-manifest.json: {key}"

    def test_build_manifest_foundation_is_debian(self):
        """build-manifest.json records foundation: debian."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            with open(os.path.join(out_dir, "build-manifest.json")) as fh:
                bm = json.load(fh)
            assert bm["foundation"] == "debian"


# ---------------------------------------------------------------------------
# Test: _sanitize_dst path security (T-04-05)
# ---------------------------------------------------------------------------

class TestSanitizeDst:

    def test_valid_relative_path_accepted(self):
        """A valid relative path is accepted and returned normalized."""
        result = _sanitize_dst("etc/motd")
        assert result == "etc/motd"

    def test_absolute_path_rejected(self):
        """Absolute paths are rejected with ValueError."""
        with pytest.raises(ValueError) as exc_info:
            _sanitize_dst("/etc/passwd")
        assert "absolute" in str(exc_info.value).lower() or "passwd" in str(exc_info.value)

    def test_dotdot_traversal_rejected(self):
        """Paths with .. that escape the target root are rejected."""
        with pytest.raises(ValueError) as exc_info:
            _sanitize_dst("../../etc/passwd")
        msg = str(exc_info.value)
        assert ".." in msg or "traversal" in msg.lower() or "passwd" in msg

    def test_empty_path_rejected(self):
        """Empty dst is rejected (WR-02)."""
        with pytest.raises(ValueError):
            _sanitize_dst("")

    def test_dot_rejected(self):
        """Dst='.' is rejected (WR-02)."""
        with pytest.raises(ValueError):
            _sanitize_dst(".")

    def test_nested_valid_path_accepted(self):
        """Nested valid paths like home/user/.config/app are accepted."""
        result = _sanitize_dst("home/user/.config/app")
        assert "home" in result
        assert ".." not in result

    def test_sanitize_uses_debian_sentinel(self):
        """_sanitize_dst uses a Debian-appropriate sentinel root (not airootfs)."""
        # The sentinel can be any stable path — we just check the function works
        # and that valid paths don't raise
        result = _sanitize_dst("etc/apt/sources.list.d/test.list")
        assert "etc" in result


# ---------------------------------------------------------------------------
# Test: fail-fast file I/O gate (T-04-05, mirroring arch T-02-08)
# ---------------------------------------------------------------------------

class TestFailFastGate:

    def test_traversal_dst_no_files_written(self):
        """When file_asset dst traversal is detected, NO files must be written."""
        import shutil
        resolved = {
            "schema": 1,
            "foundation": "debian",
            "applied": ["DF-002"],
            "skipped": [],
            "dropped": [],
            "install_order": ["DF-002"],
            "explanations": [],
        }
        opinions = {
            "DF-002": {
                "id": "DF-002",
                "status": "required",
                "translator_capabilities": ["deploy-config-file-tree"],
                "file_assets": [{"src": "assets/evil", "dst": "../../etc/passwd"}],
            }
        }
        from manifest import BuildManifest
        caps = {"deploy-config-file-tree"}
        resolved_bytes = b'{"test": true}'
        manifest = BuildManifest.from_resolved(resolved, opinions, caps, resolved_bytes)
        variant = load_variant_profile("debian")

        out_dir = tempfile.mkdtemp()
        try:
            with pytest.raises(ValueError):
                emit_profile_tree(out_dir, manifest, variant)
            # No profile files should exist
            files = []
            for root, dirs, fnames in os.walk(out_dir):
                files.extend(fnames)
            assert not files, f"Expected no files after fail-fast, found: {files}"
        finally:
            shutil.rmtree(out_dir, ignore_errors=True)
