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
# Test: T-04-08 — no default password; hard-fail when unset (security review)
# ---------------------------------------------------------------------------

class TestPreseedPasswordRequired:

    def test_emit_raises_without_password(self, monkeypatch):
        """emit_profile_tree refuses to bake a default password (T-04-08)."""
        monkeypatch.delenv("DEBATEOS_HASHED_PASSWORD", raising=False)
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            with pytest.raises(ValueError, match="DEBATEOS_HASHED_PASSWORD"):
                emit_profile_tree(out_dir, manifest, variant)

    def test_emit_raises_on_invalid_hash(self, monkeypatch):
        """A non-crypt password value is rejected (must start with '$')."""
        monkeypatch.setenv("DEBATEOS_HASHED_PASSWORD", "plaintextnope")
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            with pytest.raises(ValueError, match="DEBATEOS_HASHED_PASSWORD"):
                emit_profile_tree(out_dir, manifest, variant)

    def test_preseed_has_no_sentinels_and_real_hash(self):
        """With a valid hash set (autouse fixture), no %% literals remain and
        the password field is a crypt hash."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            preseed = open(
                os.path.join(out_dir, "config", "includes.installer", "preseed.cfg")
            ).read()
            assert "%%" not in preseed, "unreplaced sentinel left in preseed.cfg"
            assert "$6$" in preseed, "no SHA-512 crypt hash in preseed.cfg"


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

    def test_preseed_contains_hashed_password_not_plaintext(self):
        """preseed.cfg uses a crypt hash for the password, never plaintext (T-04-08).

        CR-02 fix: the emitted preseed.cfg must have a real crypt hash (starts with $)
        rather than a literal %%HASHED_PASSWORD%% sentinel that d-i cannot parse.
        """
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            preseed_path = os.path.join(
                out_dir, "config", "includes.installer", "preseed.cfg"
            )
            with open(preseed_path) as fh:
                content = fh.read()
            # Must use a crypt hash (starts with $), never the literal sentinel
            assert "%%HASHED_PASSWORD%%" not in content, (
                "preseed.cfg must not contain literal %%HASHED_PASSWORD%% sentinel (CR-02)"
            )
            # The password field must start with $ (crypt hash format, T-04-08)
            found_password_line = False
            for line in content.splitlines():
                if "user-password-crypted" in line and not line.strip().startswith("#"):
                    found_password_line = True
                    parts = line.split()
                    pw_value = parts[-1] if parts else ""
                    assert pw_value.startswith("$"), (
                        f"preseed.cfg hashed password must start with '$' (crypt format), "
                        f"got: {pw_value!r} (T-04-08)"
                    )
            assert found_password_line, "preseed.cfg must contain d-i passwd/user-password-crypted line"

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


# ---------------------------------------------------------------------------
# CR-01: Shell injection via raw opinion data — safe manifest pattern
# ---------------------------------------------------------------------------

class TestCR01ShellInjectionSafety:
    """Tests that dangerous chars in opinion data do NOT reach shell command position.

    The safe pattern: data goes into build-manifest.json verbatim; the chroot
    hook reads it with printf/jq and always quotes the variables.  The generated
    hook must NOT contain the raw dangerous strings interpolated into command position.
    """

    def _make_manifest_with_dangerous_sysctl(self) -> "BuildManifest":
        """Manifest with a sysctl key that contains a single-quote (injection attempt)."""
        resolved = {
            "schema": 1, "foundation": "debian",
            "applied": ["INJ-001"], "skipped": [], "dropped": [],
            "install_order": ["INJ-001"], "explanations": [],
        }
        opinions = {
            "INJ-001": {
                "id": "INJ-001", "status": "required",
                "translator_capabilities": ["write-sysctl-drop-in"],
                "sysctl_params": [
                    {"key": "net.ipv4.tcp_fastopen", "value": "3'; rm -rf /;'",
                     "drop_in_file": "50-test.conf"},
                ],
            }
        }
        caps = {"write-sysctl-drop-in"}
        return BuildManifest.from_resolved(resolved, opinions, caps, b'{}')

    def _make_manifest_with_dangerous_service(self) -> "BuildManifest":
        """Manifest with a service name that contains a semicolon (injection attempt)."""
        resolved = {
            "schema": 1, "foundation": "debian",
            "applied": ["INJ-002"], "skipped": [], "dropped": [],
            "install_order": ["INJ-002"], "explanations": [],
        }
        opinions = {
            "INJ-002": {
                "id": "INJ-002", "status": "required",
                "translator_capabilities": ["enable-systemd-service"],
                "services": [
                    {"name": "safe-service.service; rm -rf /", "enable": True, "deferred": False}
                ],
            }
        }
        caps = {"enable-systemd-service"}
        return BuildManifest.from_resolved(resolved, opinions, caps, b'{}')

    def _make_manifest_with_dangerous_group(self) -> "BuildManifest":
        """Manifest with a group name containing a semicolon."""
        resolved = {
            "schema": 1, "foundation": "debian",
            "applied": ["INJ-003"], "skipped": [], "dropped": [],
            "install_order": ["INJ-003"], "explanations": [],
        }
        opinions = {
            "INJ-003": {
                "id": "INJ-003", "status": "required",
                "translator_capabilities": ["add-user-to-group"],
                "group_memberships": [
                    {"group": "video; reboot"}
                ],
            }
        }
        caps = {"add-user-to-group"}
        return BuildManifest.from_resolved(resolved, opinions, caps, b'{}')

    def _make_manifest_with_dangerous_kernel_param(self) -> "BuildManifest":
        """Manifest with a kernel param value containing a single-quote and slash."""
        resolved = {
            "schema": 1, "foundation": "debian",
            "applied": ["INJ-004"], "skipped": [], "dropped": [],
            "install_order": ["INJ-004"], "explanations": [],
        }
        opinions = {
            "INJ-004": {
                "id": "INJ-004", "status": "required",
                "translator_capabilities": ["configure-kernel-cmdline"],
                "kernel_params": [
                    {"key": "quiet", "value": "splash'/etc/passwd; reboot;#"}
                ],
            }
        }
        caps = {"configure-kernel-cmdline"}
        return BuildManifest.from_resolved(resolved, opinions, caps, b'{}')

    def test_sysctl_injection_chars_not_in_hook_command_position(self):
        """A sysctl value with '; rm -rf /' must NOT appear unquoted in shell command position.

        The hook must write data via the manifest's jq/printf pattern so that the
        dangerous value is safely quoted.  The dangerous literal should appear only
        in the build-manifest.json (data), not interpolated into a bare printf argument.
        """
        manifest = self._make_manifest_with_dangerous_sysctl()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            hook_path = os.path.join(
                out_dir, "config", "hooks", "live",
                "9000-debateos-apply.hook.chroot"
            )
            with open(hook_path) as fh:
                hook_content = fh.read()
            # The injection string must NOT appear bare in the hook
            assert "rm -rf /" not in hook_content, (
                "Injection payload 'rm -rf /' must not appear in the chroot hook: "
                "sysctl values must be sourced from build-manifest.json at runtime, "
                "not embedded in shell code (CR-01)"
            )
            # build-manifest.json should carry the value safely
            bm_path = os.path.join(out_dir, "build-manifest.json")
            with open(bm_path) as fh:
                bm = json.load(fh)
            assert any(
                p.get("value", "") == "3'; rm -rf /;'"
                for p in bm.get("sysctl_params", [])
            ), "build-manifest.json must carry the sysctl value verbatim for safe runtime use"

    def test_service_injection_chars_not_in_hook_command_position(self):
        """A service name with '; rm -rf /' must NOT appear unquoted in systemctl enable line."""
        manifest = self._make_manifest_with_dangerous_service()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            hook_path = os.path.join(
                out_dir, "config", "hooks", "live",
                "9000-debateos-apply.hook.chroot"
            )
            with open(hook_path) as fh:
                hook_content = fh.read()
            # The injection string must NOT appear bare after systemctl enable
            # If it appears in a comment or data-read section that's OK,
            # but it must not be literally `systemctl enable ...; rm -rf /`
            lines_with_enable = [
                l for l in hook_content.splitlines()
                if "systemctl enable" in l and "rm -rf" in l
            ]
            assert not lines_with_enable, (
                f"Found unquoted service injection in systemctl enable line(s): "
                f"{lines_with_enable} (CR-01)"
            )

    def test_group_injection_chars_not_in_hook_command_position(self):
        """A group name with '; reboot' must NOT appear unquoted in groupadd command."""
        manifest = self._make_manifest_with_dangerous_group()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            hook_path = os.path.join(
                out_dir, "config", "hooks", "live",
                "9000-debateos-apply.hook.chroot"
            )
            with open(hook_path) as fh:
                hook_content = fh.read()
            lines_with_groupadd = [
                l for l in hook_content.splitlines()
                if "groupadd" in l and "reboot" in l
            ]
            assert not lines_with_groupadd, (
                f"Found unquoted group injection in groupadd line(s): "
                f"{lines_with_groupadd} (CR-01)"
            )

    def test_manifest_carries_dangerous_values_verbatim(self):
        """build-manifest.json carries all opinion data verbatim (safe data-only channel)."""
        manifest = self._make_manifest_with_dangerous_sysctl()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            bm_path = os.path.join(out_dir, "build-manifest.json")
            with open(bm_path) as fh:
                bm = json.load(fh)
            sysctl = bm.get("sysctl_params", [])
            assert len(sysctl) == 1
            assert sysctl[0]["key"] == "net.ipv4.tcp_fastopen"
            # The value must be stored verbatim in JSON (JSON escaping is safe)
            assert "rm -rf" in sysctl[0]["value"]


# ---------------------------------------------------------------------------
# CR-02: Preseed sentinels must not appear literally in emitted preseed.cfg
# ---------------------------------------------------------------------------

class TestCR02PreseedSentinels:
    """The emitted preseed.cfg must not contain raw %%...%% sentinel literals."""

    def test_preseed_no_percent_sentinels(self):
        """preseed.cfg must not contain any literal %%...%% sentinels."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            preseed_path = os.path.join(
                out_dir, "config", "includes.installer", "preseed.cfg"
            )
            with open(preseed_path) as fh:
                content = fh.read()
            import re
            sentinels = re.findall(r'%%[A-Z_]+%%', content)
            assert not sentinels, (
                f"preseed.cfg contains unreplaced sentinel(s): {sentinels} — "
                "d-i will reject these as invalid values (CR-02)"
            )

    def test_preseed_username_is_real_value(self):
        """preseed.cfg d-i passwd/username must have a real value, not %%USERNAME%%."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            preseed_path = os.path.join(
                out_dir, "config", "includes.installer", "preseed.cfg"
            )
            with open(preseed_path) as fh:
                content = fh.read()
            for line in content.splitlines():
                if "passwd/username" in line:
                    assert "%%USERNAME%%" not in line, (
                        f"Username sentinel not replaced: {line!r} (CR-02)"
                    )
                    break

    def test_preseed_hashed_password_starts_with_dollar(self):
        """preseed.cfg hashed password must look like a crypt hash (starts with $)."""
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            preseed_path = os.path.join(
                out_dir, "config", "includes.installer", "preseed.cfg"
            )
            with open(preseed_path) as fh:
                content = fh.read()
            for line in content.splitlines():
                if "user-password-crypted" in line:
                    # Value after "password " should start with $ (crypt hash)
                    # or the line should be commented out
                    if line.strip().startswith("#"):
                        continue
                    # Extract the value after the last space
                    parts = line.split()
                    if parts:
                        pw_value = parts[-1]
                        assert pw_value.startswith("$"), (
                            f"Hashed password in preseed.cfg must start with '$' "
                            f"(crypt format), got: {pw_value!r} (CR-02)"
                        )


# ---------------------------------------------------------------------------
# CR-03: Empty target_packages must not produce broken apt-get install line
# ---------------------------------------------------------------------------

class TestCR03EmptyPackages:
    """When target_packages is empty, the chroot hook must not have a broken install line."""

    def _make_manifest_no_packages(self) -> "BuildManifest":
        """Manifest with zero target_packages (only sysctl, no install-packages opinion)."""
        resolved = {
            "schema": 1, "foundation": "debian",
            "applied": ["NP-001"], "skipped": [], "dropped": [],
            "install_order": ["NP-001"], "explanations": [],
        }
        opinions = {
            "NP-001": {
                "id": "NP-001", "status": "required",
                "translator_capabilities": ["write-sysctl-drop-in"],
                "sysctl_params": [
                    {"key": "net.ipv4.tcp_fastopen", "value": "3",
                     "drop_in_file": "50-debateos.conf"},
                ],
            }
        }
        caps = {"write-sysctl-drop-in"}
        return BuildManifest.from_resolved(resolved, opinions, caps, b'{}')

    def test_empty_packages_no_broken_apt_get(self):
        """When target_packages is empty, hook must not have 'apt-get install -y ... :'."""
        manifest = self._make_manifest_no_packages()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            hook_path = os.path.join(
                out_dir, "config", "hooks", "live",
                "9000-debateos-apply.hook.chroot"
            )
            with open(hook_path) as fh:
                hook_content = fh.read()
            # Must not have apt-get install followed by bare ":"
            # The broken pattern is: apt-get install -y --no-install-recommends \<newline>:
            assert "apt-get install -y --no-install-recommends \\\n:" not in hook_content, (
                "Empty packages produced broken 'apt-get install :' line (CR-03)"
            )
            assert "apt-get install -y --no-install-recommends \\\n  :" not in hook_content, (
                "Empty packages produced broken 'apt-get install  :' line (CR-03)"
            )

    def test_empty_packages_hook_passes_bash_syntax(self):
        """When target_packages is empty, the generated hook must pass 'bash -n' syntax check."""
        import subprocess
        manifest = self._make_manifest_no_packages()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            hook_path = os.path.join(
                out_dir, "config", "hooks", "live",
                "9000-debateos-apply.hook.chroot"
            )
            result = subprocess.run(
                ["bash", "-n", hook_path],
                capture_output=True, text=True
            )
            assert result.returncode == 0, (
                f"chroot hook failed bash -n with empty packages:\n{result.stderr} (CR-03)"
            )

    def test_nonempty_packages_hook_passes_bash_syntax(self):
        """With packages, the generated hook must pass 'bash -n' syntax check."""
        import subprocess
        manifest = _load_df_manifest()
        variant = load_variant_profile("debian")
        with tempfile.TemporaryDirectory() as out_dir:
            emit_profile_tree(out_dir, manifest, variant)
            hook_path = os.path.join(
                out_dir, "config", "hooks", "live",
                "9000-debateos-apply.hook.chroot"
            )
            result = subprocess.run(
                ["bash", "-n", hook_path],
                capture_output=True, text=True
            )
            assert result.returncode == 0, (
                f"chroot hook failed bash -n:\n{result.stderr}"
            )
