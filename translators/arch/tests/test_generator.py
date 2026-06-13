"""
test_generator.py — RED tests for generator.py + translate shell wrapper.

These tests MUST FAIL before implementation (TDD RED phase, D19).

Covers:
- generate() end-to-end on omarchy_subset fixtures
- Capability gate fires through generate() on unsupported-required fixture
- translate argv parsing (subprocess)
"""

import json
import os
import stat
import subprocess
import sys
import tempfile

import pytest

# ---------------------------------------------------------------------------
# Import generator (will fail if not implemented — RED)
# ---------------------------------------------------------------------------

from generator import generate


# ---------------------------------------------------------------------------
# Paths
# ---------------------------------------------------------------------------

_FIXTURES_DIR = os.path.join(os.path.dirname(__file__), "fixtures")
_ARCH_DIR = os.path.dirname(os.path.dirname(__file__))  # translators/arch/
_TRANSLATE = os.path.join(_ARCH_DIR, "translate")

_SUBSET_RESOLVED = os.path.join(_FIXTURES_DIR, "omarchy_subset_resolved.json")
_SUBSET_OPINIONS = os.path.join(_FIXTURES_DIR, "omarchy_subset_opinions.json")
_UNSUPPORTED_RESOLVED = os.path.join(_FIXTURES_DIR, "unsupported_required_resolved.json")
_UNSUPPORTED_OPINIONS = os.path.join(_FIXTURES_DIR, "unsupported_required_opinions.json")


# ---------------------------------------------------------------------------
# Test: generate() end-to-end on omarchy_subset
# ---------------------------------------------------------------------------

class TestGenerateEndToEnd:

    def test_generate_returns_out_dir(self):
        """generate() returns the out_dir path on success."""
        with tempfile.TemporaryDirectory() as out_dir:
            result = generate(
                resolved_path=_SUBSET_RESOLVED,
                opinions_path=_SUBSET_OPINIONS,
                profile_name="vanilla-arch",
                out_dir=out_dir,
            )
            assert result == out_dir

    def test_generate_creates_profiledef_sh(self):
        """generate() creates profiledef.sh in the output directory."""
        with tempfile.TemporaryDirectory() as out_dir:
            generate(
                resolved_path=_SUBSET_RESOLVED,
                opinions_path=_SUBSET_OPINIONS,
                profile_name="vanilla-arch",
                out_dir=out_dir,
            )
            assert os.path.isfile(os.path.join(out_dir, "profiledef.sh"))

    def test_generate_creates_installer_script(self):
        """generate() creates debateos-install.sh in airootfs/root/."""
        with tempfile.TemporaryDirectory() as out_dir:
            generate(
                resolved_path=_SUBSET_RESOLVED,
                opinions_path=_SUBSET_OPINIONS,
                profile_name="vanilla-arch",
                out_dir=out_dir,
            )
            installer = os.path.join(out_dir, "airootfs", "root", "debateos-install.sh")
            assert os.path.isfile(installer)

    def test_generate_creates_build_manifest_json(self):
        """generate() creates build-manifest.json in the output directory."""
        with tempfile.TemporaryDirectory() as out_dir:
            generate(
                resolved_path=_SUBSET_RESOLVED,
                opinions_path=_SUBSET_OPINIONS,
                profile_name="vanilla-arch",
                out_dir=out_dir,
            )
            bm = os.path.join(out_dir, "build-manifest.json")
            assert os.path.isfile(bm)
            # Must be valid JSON with expected keys (IN-04: use context manager)
            with open(bm) as fh:
                data = json.load(fh)
            assert "target_packages" in data
            assert "file_assets" in data

    def test_generate_creates_zlogin(self):
        """generate() creates .zlogin in airootfs/root/."""
        with tempfile.TemporaryDirectory() as out_dir:
            generate(
                resolved_path=_SUBSET_RESOLVED,
                opinions_path=_SUBSET_OPINIONS,
                profile_name="vanilla-arch",
                out_dir=out_dir,
            )
            zlogin = os.path.join(out_dir, "airootfs", "root", ".zlogin")
            assert os.path.isfile(zlogin)

    def test_generate_complete_profile_tree(self):
        """generate() produces all required profile tree components."""
        with tempfile.TemporaryDirectory() as out_dir:
            generate(
                resolved_path=_SUBSET_RESOLVED,
                opinions_path=_SUBSET_OPINIONS,
                profile_name="vanilla-arch",
                out_dir=out_dir,
            )
            # All required components
            required = [
                "profiledef.sh",
                "packages.x86_64",
                "pacman.conf",
                os.path.join("airootfs", "root", "debateos-install.sh"),
                os.path.join("airootfs", "root", ".zlogin"),
                "build-manifest.json",
            ]
            for rel_path in required:
                full_path = os.path.join(out_dir, rel_path)
                assert os.path.isfile(full_path), f"Missing: {rel_path}"

    def test_generate_with_cachyos_profile(self):
        """generate() works with cachyos profile (ARCH-04 no-fork verification)."""
        with tempfile.TemporaryDirectory() as out_dir:
            generate(
                resolved_path=_SUBSET_RESOLVED,
                opinions_path=_SUBSET_OPINIONS,
                profile_name="cachyos",
                out_dir=out_dir,
            )
            pacman_conf = open(os.path.join(out_dir, "pacman.conf")).read()
            assert "[cachyos]" in pacman_conf

    def test_generate_with_garuda_profile(self):
        """generate() works with garuda profile (ARCH-04 no-fork verification)."""
        with tempfile.TemporaryDirectory() as out_dir:
            generate(
                resolved_path=_SUBSET_RESOLVED,
                opinions_path=_SUBSET_OPINIONS,
                profile_name="garuda",
                out_dir=out_dir,
            )
            pacman_conf = open(os.path.join(out_dir, "pacman.conf")).read()
            assert "[chaotic-aur]" in pacman_conf


# ---------------------------------------------------------------------------
# Test: capability gate fires through generate()
# ---------------------------------------------------------------------------

class TestGenerateCapabilityGate:

    def test_unsupported_required_raises_capability_error(self):
        """generate() raises CapabilityError when a required opinion needs
        an unsupported capability (end-to-end ARCH-03 gate via generate())."""
        from capabilities import CapabilityError
        with tempfile.TemporaryDirectory() as out_dir:
            with pytest.raises(CapabilityError) as exc_info:
                generate(
                    resolved_path=_UNSUPPORTED_RESOLVED,
                    opinions_path=_UNSUPPORTED_OPINIONS,
                    profile_name="vanilla-arch",
                    out_dir=out_dir,
                )
            # Error must name the opinion + token + "composition time"
            err = str(exc_info.value)
            assert "composition time" in err

    def test_unsupported_required_no_partial_output(self):
        """When CapabilityError fires, generate() has not written files
        (fails fast before profile emission)."""
        from capabilities import CapabilityError
        with tempfile.TemporaryDirectory() as out_dir:
            try:
                generate(
                    resolved_path=_UNSUPPORTED_RESOLVED,
                    opinions_path=_UNSUPPORTED_OPINIONS,
                    profile_name="vanilla-arch",
                    out_dir=out_dir,
                )
            except CapabilityError:
                pass
            # profiledef.sh should NOT exist — gate fires before emit
            assert not os.path.isfile(os.path.join(out_dir, "profiledef.sh")), (
                "profiledef.sh must not be written when capability gate fires"
            )


# ---------------------------------------------------------------------------
# Test: generate() runnable as python -m translators.arch.generator
# ---------------------------------------------------------------------------

class TestGeneratorModuleMain:

    def test_generator_runnable_as_module(self):
        """generator.py can be invoked as python3 -m translators.arch.generator."""
        with tempfile.TemporaryDirectory() as out_dir:
            # translators/arch/ is two levels above tests/
            # repo root is two levels above translators/arch/
            repo_root = os.path.dirname(os.path.dirname(_ARCH_DIR))
            result = subprocess.run(
                [
                    sys.executable, "-m", "translators.arch.generator",
                    _SUBSET_RESOLVED, _SUBSET_OPINIONS, "vanilla-arch", out_dir,
                ],
                cwd=repo_root,
                capture_output=True,
                text=True,
            )
            assert result.returncode == 0, (
                f"Module run failed:\nstdout: {result.stdout}\nstderr: {result.stderr}"
            )
            assert os.path.isfile(os.path.join(out_dir, "profiledef.sh"))


# ---------------------------------------------------------------------------
# Test: translate shell wrapper argv parsing
# ---------------------------------------------------------------------------

class TestTranslateWrapper:

    def test_translate_is_executable(self):
        """translate wrapper file exists and is executable."""
        assert os.path.isfile(_TRANSLATE), f"translate not found at {_TRANSLATE}"
        mode = os.stat(_TRANSLATE).st_mode
        assert mode & stat.S_IXUSR, "translate must be executable"

    def test_translate_has_required_flags(self):
        """translate wrapper source contains --opinions, --profile, --out flags."""
        with open(_TRANSLATE) as fh:
            content = fh.read()
        assert "--opinions" in content, "translate must handle --opinions flag"
        assert "--profile" in content, "translate must handle --profile flag"
        assert "--out" in content, "translate must handle --out flag"

    def test_translate_unknown_flag_exits_nonzero(self):
        """translate exits non-zero for unknown flags."""
        with tempfile.TemporaryDirectory() as out_dir:
            result = subprocess.run(
                [_TRANSLATE, _SUBSET_RESOLVED, "--unknown-flag", "foo"],
                capture_output=True,
                text=True,
            )
            assert result.returncode != 0, (
                "translate must exit non-zero for unknown flags"
            )

    def test_translate_end_to_end(self):
        """translate runs generate() end-to-end via the wrapper."""
        with tempfile.TemporaryDirectory() as out_dir:
            result = subprocess.run(
                [
                    _TRANSLATE,
                    _SUBSET_RESOLVED,
                    "--opinions", _SUBSET_OPINIONS,
                    "--profile", "vanilla-arch",
                    "--out", out_dir,
                ],
                capture_output=True,
                text=True,
            )
            assert result.returncode == 0, (
                f"translate failed:\nstdout: {result.stdout}\nstderr: {result.stderr}"
            )
            assert os.path.isfile(os.path.join(out_dir, "profiledef.sh"))

    def test_translate_default_profile(self):
        """translate uses vanilla-arch as the default profile if --profile is omitted."""
        with tempfile.TemporaryDirectory() as out_dir:
            result = subprocess.run(
                [
                    _TRANSLATE,
                    _SUBSET_RESOLVED,
                    "--opinions", _SUBSET_OPINIONS,
                    "--out", out_dir,
                ],
                capture_output=True,
                text=True,
            )
            assert result.returncode == 0, (
                f"translate (default profile) failed:\n{result.stderr}"
            )
