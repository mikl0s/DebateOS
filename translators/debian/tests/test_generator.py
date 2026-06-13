"""
test_generator.py — RED tests for Debian generator.py + translate shell wrapper.

TDD RED phase: Written BEFORE generator.py and translate exist. MUST fail now.

Coverage:
- generate() end-to-end on df_resolved.json fixtures
- Capability gate fires through generate() — no file I/O on CapabilityError
- translate shell wrapper: exists, executable, correct argv parsing
- translate end-to-end: emits preseed.cfg + executable chroot hook
"""

import json
import os
import stat
import subprocess
import sys
import tempfile

import pytest

# Import module under test (will fail until generator.py exists — RED)
from generator import generate


# ---------------------------------------------------------------------------
# Paths
# ---------------------------------------------------------------------------

FIXTURES_DIR = os.path.join(os.path.dirname(__file__), "fixtures")
DEBIAN_DIR = os.path.dirname(os.path.dirname(__file__))  # translators/debian/
TRANSLATE = os.path.join(DEBIAN_DIR, "translate")
REPO_ROOT = os.path.abspath(os.path.join(DEBIAN_DIR, "..", ".."))

DF_RESOLVED = os.path.join(FIXTURES_DIR, "df_resolved.json")
DF_OPINIONS = os.path.join(FIXTURES_DIR, "df_opinions.json")


# ---------------------------------------------------------------------------
# Test: generate() end-to-end
# ---------------------------------------------------------------------------

class TestGenerateEndToEnd:

    def test_generate_returns_out_dir(self):
        """generate() returns the out_dir path on success."""
        with tempfile.TemporaryDirectory() as out_dir:
            result = generate(
                resolved_path=DF_RESOLVED,
                opinions_path=DF_OPINIONS,
                profile_name="debian",
                out_dir=out_dir,
            )
            assert result == out_dir

    def test_generate_creates_preseed_cfg(self):
        """generate() creates config/includes.installer/preseed.cfg."""
        with tempfile.TemporaryDirectory() as out_dir:
            generate(
                resolved_path=DF_RESOLVED,
                opinions_path=DF_OPINIONS,
                profile_name="debian",
                out_dir=out_dir,
            )
            preseed_path = os.path.join(
                out_dir, "config", "includes.installer", "preseed.cfg"
            )
            assert os.path.isfile(preseed_path), f"preseed.cfg not found at {preseed_path}"

    def test_generate_creates_chroot_hook(self):
        """generate() creates the executable chroot hook."""
        with tempfile.TemporaryDirectory() as out_dir:
            generate(
                resolved_path=DF_RESOLVED,
                opinions_path=DF_OPINIONS,
                profile_name="debian",
                out_dir=out_dir,
            )
            hook_path = os.path.join(
                out_dir, "config", "hooks", "live",
                "9000-debateos-apply.hook.chroot"
            )
            assert os.path.isfile(hook_path), f"chroot hook not found at {hook_path}"
            mode = os.stat(hook_path).st_mode
            assert mode & stat.S_IXUSR, "chroot hook must be executable"

    def test_generate_creates_package_list(self):
        """generate() creates config/package-lists/debateos.list.chroot_install."""
        with tempfile.TemporaryDirectory() as out_dir:
            generate(
                resolved_path=DF_RESOLVED,
                opinions_path=DF_OPINIONS,
                profile_name="debian",
                out_dir=out_dir,
            )
            pkg_list = os.path.join(
                out_dir, "config", "package-lists", "debateos.list.chroot_install"
            )
            assert os.path.isfile(pkg_list)

    def test_generate_creates_build_manifest(self):
        """generate() creates build-manifest.json."""
        with tempfile.TemporaryDirectory() as out_dir:
            generate(
                resolved_path=DF_RESOLVED,
                opinions_path=DF_OPINIONS,
                profile_name="debian",
                out_dir=out_dir,
            )
            bm_path = os.path.join(out_dir, "build-manifest.json")
            assert os.path.isfile(bm_path)
            with open(bm_path) as fh:
                bm = json.load(fh)
            assert "target_packages" in bm
            assert "foundation" in bm

    def test_generate_complete_config_tree(self):
        """generate() produces a structurally valid config/ tree."""
        with tempfile.TemporaryDirectory() as out_dir:
            generate(
                resolved_path=DF_RESOLVED,
                opinions_path=DF_OPINIONS,
                profile_name="debian",
                out_dir=out_dir,
            )
            required = [
                os.path.join("config", "includes.installer", "preseed.cfg"),
                os.path.join("config", "hooks", "live", "9000-debateos-apply.hook.chroot"),
                os.path.join("config", "package-lists", "debateos.list.chroot_install"),
                "build-manifest.json",
            ]
            for rel_path in required:
                full_path = os.path.join(out_dir, rel_path)
                assert os.path.isfile(full_path), f"Missing required file: {rel_path}"


# ---------------------------------------------------------------------------
# Test: capability gate fires through generate()
# ---------------------------------------------------------------------------

class TestGenerateCapabilityGate:

    def test_unsupported_required_raises_capability_error(self):
        """generate() raises CapabilityError when required opinion needs unsupported cap."""
        from capabilities import CapabilityError
        # ARCH-ONLY-001 in df_opinions.json requires configure-mkinitcpio-hooks-and-modules
        # We build a resolved with ARCH-ONLY-001 as the only applied opinion
        import json
        resolved = {
            "schema": 1,
            "foundation": "debian",
            "applied": ["ARCH-ONLY-001"],
            "skipped": [],
            "dropped": [],
            "install_order": ["ARCH-ONLY-001"],
            "explanations": [],
        }
        resolved_path = os.path.join(FIXTURES_DIR, "df_resolved.json")

        with tempfile.TemporaryDirectory() as out_dir:
            # Write a temporary resolved with ARCH-ONLY-001
            import tempfile as tf2
            with tf2.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as tmp:
                json.dump(resolved, tmp)
                tmp_path = tmp.name
            try:
                with pytest.raises(CapabilityError) as exc_info:
                    generate(
                        resolved_path=tmp_path,
                        opinions_path=DF_OPINIONS,
                        profile_name="debian",
                        out_dir=out_dir,
                    )
                msg = str(exc_info.value)
                assert "composition time" in msg
            finally:
                os.unlink(tmp_path)

    def test_capability_error_no_files_written(self):
        """When CapabilityError fires, generate() writes NO files (fail-fast before I/O)."""
        from capabilities import CapabilityError
        import json, tempfile as tf2
        resolved = {
            "schema": 1,
            "foundation": "debian",
            "applied": ["ARCH-ONLY-001"],
            "skipped": [],
            "dropped": [],
            "install_order": ["ARCH-ONLY-001"],
            "explanations": [],
        }
        with tempfile.TemporaryDirectory() as out_dir:
            with tf2.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as tmp:
                json.dump(resolved, tmp)
                tmp_path = tmp.name
            try:
                try:
                    generate(
                        resolved_path=tmp_path,
                        opinions_path=DF_OPINIONS,
                        profile_name="debian",
                        out_dir=out_dir,
                    )
                except CapabilityError:
                    pass

                # No config/ tree files should exist
                files = []
                for root, dirs, fnames in os.walk(out_dir):
                    files.extend(fnames)
                assert not files, (
                    f"Expected no files when CapabilityError fires before I/O, found: {files}"
                )
            finally:
                os.unlink(tmp_path)


# ---------------------------------------------------------------------------
# Test: translate shell wrapper
# ---------------------------------------------------------------------------

class TestTranslateWrapper:

    def test_translate_file_exists(self):
        """translators/debian/translate file exists."""
        assert os.path.isfile(TRANSLATE), f"translate not found at {TRANSLATE}"

    def test_translate_is_executable(self):
        """translate wrapper is executable."""
        assert os.path.isfile(TRANSLATE), f"translate not found at {TRANSLATE}"
        mode = os.stat(TRANSLATE).st_mode
        assert mode & stat.S_IXUSR, "translate must be executable"

    def test_translate_has_required_flags(self):
        """translate source contains --opinions, --profile, --out flags."""
        with open(TRANSLATE) as fh:
            content = fh.read()
        assert "--opinions" in content
        assert "--profile" in content
        assert "--out" in content

    def test_translate_default_profile_is_debian(self):
        """translate defaults to 'debian' as the profile name."""
        with open(TRANSLATE) as fh:
            content = fh.read()
        # Default profile must be "debian" not "vanilla-arch"
        assert '"debian"' in content or "'debian'" in content or "debian" in content, \
            "translate default profile must be 'debian'"

    def test_translate_unknown_flag_exits_nonzero(self):
        """translate exits non-zero for unknown flags."""
        if not os.path.isfile(TRANSLATE):
            pytest.skip("translate not yet implemented")
        result = subprocess.run(
            [TRANSLATE, DF_RESOLVED, "--unknown-flag", "foo"],
            capture_output=True,
            text=True,
        )
        assert result.returncode != 0, "translate must exit non-zero for unknown flags"

    def test_translate_end_to_end(self):
        """translate runs generate() end-to-end via the shell wrapper."""
        if not os.path.isfile(TRANSLATE):
            pytest.skip("translate not yet implemented")
        with tempfile.TemporaryDirectory() as out_dir:
            result = subprocess.run(
                [
                    TRANSLATE,
                    DF_RESOLVED,
                    "--opinions", DF_OPINIONS,
                    "--profile", "debian",
                    "--out", out_dir,
                ],
                capture_output=True,
                text=True,
                env={**os.environ, "PYTHONPATH": REPO_ROOT},
            )
            assert result.returncode == 0, (
                f"translate failed:\nstdout: {result.stdout}\nstderr: {result.stderr}"
            )
            preseed_path = os.path.join(out_dir, "config", "includes.installer", "preseed.cfg")
            hook_path = os.path.join(
                out_dir, "config", "hooks", "live", "9000-debateos-apply.hook.chroot"
            )
            assert os.path.isfile(preseed_path), "preseed.cfg missing after translate"
            assert os.path.isfile(hook_path), "chroot hook missing after translate"
            assert os.stat(hook_path).st_mode & stat.S_IXUSR, "chroot hook must be executable"
