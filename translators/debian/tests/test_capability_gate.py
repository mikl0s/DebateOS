"""
test_capability_gate.py — RED tests for Debian capabilities.py (DEB-01 / DEB-03).

TDD RED phase: These tests are written BEFORE capabilities.py exists.
They MUST fail now and pass after implementation (GREEN).

Coverage:
- load_capabilities() returns a set containing the 5 dual-foundation tokens
- load_capabilities() EXCLUDES Arch-specific tokens (mkinitcpio/limine)
- check_capabilities() raises CapabilityError for required+unsupported (names opinion+token+"composition time")
- check_capabilities() drops nice-to-have+unsupported (not raised)
- CapabilityError is importable as an Exception subclass
"""

import json
import os
import pytest

FIXTURES_DIR = os.path.join(os.path.dirname(__file__), "fixtures")


def load_fixture(name):
    path = os.path.join(FIXTURES_DIR, name)
    with open(path) as f:
        return json.load(f)


# ---------------------------------------------------------------------------
# Tests: load_capabilities
# ---------------------------------------------------------------------------

class TestLoadCapabilities:

    def test_returns_set(self):
        """load_capabilities() must return a set of strings."""
        from capabilities import load_capabilities
        caps = load_capabilities()
        assert isinstance(caps, set)
        assert len(caps) > 0

    def test_contains_install_packages(self):
        """capabilities.json must declare install-packages (DF-001)."""
        from capabilities import load_capabilities
        caps = load_capabilities()
        assert "install-packages" in caps

    def test_contains_deploy_config_file_tree(self):
        """capabilities.json must declare deploy-config-file-tree (DF-002)."""
        from capabilities import load_capabilities
        caps = load_capabilities()
        assert "deploy-config-file-tree" in caps

    def test_contains_enable_systemd_service(self):
        """capabilities.json must declare enable-systemd-service (DF-003)."""
        from capabilities import load_capabilities
        caps = load_capabilities()
        assert "enable-systemd-service" in caps

    def test_contains_write_sysctl_drop_in(self):
        """capabilities.json must declare write-sysctl-drop-in (DF-004)."""
        from capabilities import load_capabilities
        caps = load_capabilities()
        assert "write-sysctl-drop-in" in caps

    def test_contains_add_user_to_group(self):
        """capabilities.json must declare add-user-to-group (DF-005)."""
        from capabilities import load_capabilities
        caps = load_capabilities()
        assert "add-user-to-group" in caps

    def test_excludes_mkinitcpio_token(self):
        """Debian caps must NOT declare configure-mkinitcpio-hooks-and-modules (Arch-only)."""
        from capabilities import load_capabilities
        caps = load_capabilities()
        assert "configure-mkinitcpio-hooks-and-modules" not in caps

    def test_excludes_limine_token(self):
        """Debian caps must NOT declare manage-limine-bootloader-installation (Arch-only)."""
        from capabilities import load_capabilities
        caps = load_capabilities()
        assert "manage-limine-bootloader-installation" not in caps

    def test_excludes_write_mkinitcpio_config(self):
        """Debian caps must NOT declare write-mkinitcpio-config-drop-in (Arch-only)."""
        from capabilities import load_capabilities
        caps = load_capabilities()
        assert "write-mkinitcpio-config-drop-in" not in caps


# ---------------------------------------------------------------------------
# Tests: check_capabilities — happy path (all 5 DF tokens supported)
# ---------------------------------------------------------------------------

class TestCheckCapabilitiesHappyPath:

    def test_passes_when_all_required_supported(self):
        """check_capabilities must not raise when all required caps are declared."""
        from capabilities import check_capabilities
        rs = load_fixture("df_resolved.json")
        opinions = load_fixture("df_opinions.json")
        # Only the 5 DF opinions are applied; ARCH-ONLY opinions are not in resolved
        opinions_index = {op["id"]: op for op in opinions}
        caps = {
            "install-packages",
            "deploy-config-file-tree",
            "enable-systemd-service",
            "write-sysctl-drop-in",
            "add-user-to-group",
        }
        dropped = check_capabilities(rs, opinions_index, caps)
        assert isinstance(dropped, list)
        assert len(dropped) == 0

    def test_returns_empty_dropped_on_clean_pass(self):
        """Empty dropped list when all capabilities are satisfied."""
        from capabilities import check_capabilities
        rs = {"applied": ["DF-001"], "skipped": [], "dropped": [], "install_order": ["DF-001"]}
        opinions_index = {
            "DF-001": {"id": "DF-001", "status": "required", "translator_capabilities": ["install-packages"]}
        }
        caps = {"install-packages"}
        dropped = check_capabilities(rs, opinions_index, caps)
        assert dropped == []


# ---------------------------------------------------------------------------
# Tests: check_capabilities — required + unsupported → CapabilityError
# ---------------------------------------------------------------------------

class TestCheckCapabilitiesRequired:

    def test_required_unsupported_raises_capability_error(self):
        """Required opinion with unsupported capability must raise CapabilityError."""
        from capabilities import check_capabilities, CapabilityError
        # ARCH-ONLY-001 is required and needs configure-mkinitcpio-hooks-and-modules
        rs = {
            "applied": ["ARCH-ONLY-001"],
            "skipped": [],
            "dropped": [],
            "install_order": ["ARCH-ONLY-001"],
        }
        opinions = load_fixture("df_opinions.json")
        opinions_index = {op["id"]: op for op in opinions}
        caps = {"install-packages", "deploy-config-file-tree"}  # no mkinitcpio

        with pytest.raises(CapabilityError) as exc_info:
            check_capabilities(rs, opinions_index, caps)

        msg = str(exc_info.value)
        assert "ARCH-ONLY-001" in msg, f"Expected opinion ID in: {msg}"
        assert "configure-mkinitcpio-hooks-and-modules" in msg, f"Expected token in: {msg}"
        assert "composition time" in msg, f"Expected 'composition time' in: {msg}"

    def test_capability_error_names_opinion_and_token(self):
        """CapabilityError message must name both opinion ID and token."""
        from capabilities import check_capabilities, CapabilityError
        rs = {
            "applied": ["DF-001"],
            "skipped": [],
            "dropped": [],
            "install_order": ["DF-001"],
        }
        opinions_index = {
            "DF-001": {
                "id": "DF-001",
                "status": "required",
                "translator_capabilities": ["some-unsupported-token"],
            }
        }
        caps = {"install-packages"}

        with pytest.raises(CapabilityError) as exc_info:
            check_capabilities(rs, opinions_index, caps)

        msg = str(exc_info.value)
        assert "DF-001" in msg
        assert "some-unsupported-token" in msg
        assert "composition time" in msg

    def test_capability_error_says_debian_translator(self):
        """CapabilityError message must mention 'Debian translator' (not 'Arch translator')."""
        from capabilities import check_capabilities, CapabilityError
        rs = {
            "applied": ["DF-001"],
            "skipped": [],
            "dropped": [],
            "install_order": ["DF-001"],
        }
        opinions_index = {
            "DF-001": {
                "id": "DF-001",
                "status": "required",
                "translator_capabilities": ["unsupported-token"],
            }
        }
        caps = set()

        with pytest.raises(CapabilityError) as exc_info:
            check_capabilities(rs, opinions_index, caps)

        msg = str(exc_info.value)
        # Must say "Debian translator", not "Arch translator"
        assert "Debian translator" in msg, f"Expected 'Debian translator' in: {msg}"


# ---------------------------------------------------------------------------
# Tests: check_capabilities — nice-to-have + unsupported → dropped (not raised)
# ---------------------------------------------------------------------------

class TestCheckCapabilitiesNiceToHave:

    def test_nicetohave_unsupported_does_not_raise(self):
        """Nice-to-have opinion with unsupported capability must NOT raise."""
        from capabilities import check_capabilities
        rs = {
            "applied": ["ARCH-ONLY-NICE"],
            "skipped": [],
            "dropped": [],
            "install_order": ["ARCH-ONLY-NICE"],
        }
        opinions = load_fixture("df_opinions.json")
        opinions_index = {op["id"]: op for op in opinions}
        caps = {"install-packages"}  # no mkinitcpio

        # Must NOT raise
        dropped = check_capabilities(rs, opinions_index, caps)
        assert isinstance(dropped, list)
        dropped_ids = [d[0] for d in dropped]
        assert "ARCH-ONLY-NICE" in dropped_ids

    def test_nicetohave_dropped_entry_has_reason(self):
        """Dropped entries must have a non-empty reason naming the missing capability."""
        from capabilities import check_capabilities
        rs = {
            "applied": ["ARCH-ONLY-NICE"],
            "skipped": [],
            "dropped": [],
            "install_order": ["ARCH-ONLY-NICE"],
        }
        opinions = load_fixture("df_opinions.json")
        opinions_index = {op["id"]: op for op in opinions}
        caps = {"install-packages"}

        dropped = check_capabilities(rs, opinions_index, caps)
        entry = next((d for d in dropped if d[0] == "ARCH-ONLY-NICE"), None)
        assert entry is not None
        opinion_id, reason = entry
        assert reason
        assert "configure-mkinitcpio-hooks-and-modules" in reason

    def test_nicetohave_with_supported_cap_not_in_dropped(self):
        """Nice-to-have whose capabilities ARE declared must not appear in dropped."""
        from capabilities import check_capabilities
        rs = {"applied": ["DF-001"], "skipped": [], "dropped": [], "install_order": ["DF-001"]}
        opinions_index = {
            "DF-001": {"id": "DF-001", "status": "nice-to-have", "translator_capabilities": ["install-packages"]}
        }
        caps = {"install-packages"}
        dropped = check_capabilities(rs, opinions_index, caps)
        dropped_ids = [d[0] for d in dropped]
        assert "DF-001" not in dropped_ids


# ---------------------------------------------------------------------------
# Tests: CapabilityError is importable
# ---------------------------------------------------------------------------

class TestCapabilityError:

    def test_is_exception_subclass(self):
        """CapabilityError must be a subclass of Exception."""
        from capabilities import CapabilityError
        assert issubclass(CapabilityError, Exception)
        e = CapabilityError("test error")
        assert str(e) == "test error"
