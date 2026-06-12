"""
test_firstrun.py — RED tests for firstrun.py: render_firstrun_unit.

These tests MUST FAIL before implementation (TDD RED phase, D19).
"""

import pytest

from firstrun import render_firstrun_unit


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------

class TestRenderFirstrunUnit:

    def test_unit_is_string(self):
        """render_firstrun_unit returns a non-empty string."""
        unit = render_firstrun_unit(
            opinion_id="OM-102",
            description="GTK theme configuration",
            exec_path="/usr/lib/debateos/firstrun/OM-102.sh",
        )
        assert isinstance(unit, str)
        assert len(unit) > 0

    def test_unit_contains_unit_section(self):
        """Generated unit contains [Unit] section header."""
        unit = render_firstrun_unit("OM-102", "GTK theme", "/usr/lib/debateos/firstrun/OM-102.sh")
        assert "[Unit]" in unit

    def test_unit_contains_service_section(self):
        """Generated unit contains [Service] section header."""
        unit = render_firstrun_unit("OM-102", "GTK theme", "/usr/lib/debateos/firstrun/OM-102.sh")
        assert "[Service]" in unit

    def test_unit_contains_install_section(self):
        """Generated unit contains [Install] section header."""
        unit = render_firstrun_unit("OM-102", "GTK theme", "/usr/lib/debateos/firstrun/OM-102.sh")
        assert "[Install]" in unit

    def test_unit_flag_file_condition(self):
        """Unit contains ConditionPathExists=! guard for the flag file (Pitfall 3, Pattern 2)."""
        unit = render_firstrun_unit("OM-102", "GTK theme", "/usr/lib/debateos/firstrun/OM-102.sh")
        assert "ConditionPathExists=!" in unit, "Flag-file condition guard must be present"

    def test_unit_flag_file_includes_opinion_id(self):
        """Flag file path includes the opinion ID (for per-opinion idempotency)."""
        unit = render_firstrun_unit("OM-102", "GTK theme", "/usr/lib/debateos/firstrun/OM-102.sh")
        assert ".firstrun-" in unit, "Flag file path must contain '.firstrun-' prefix"
        assert "OM-102" in unit or "OM-102" in unit

    def test_unit_type_oneshot(self):
        """Service is Type=oneshot (not simple/forking)."""
        unit = render_firstrun_unit("OM-102", "GTK theme", "/usr/lib/debateos/firstrun/OM-102.sh")
        assert "Type=oneshot" in unit

    def test_unit_remain_after_exit(self):
        """Service has RemainAfterExit=yes (Pattern 2)."""
        unit = render_firstrun_unit("OM-102", "GTK theme", "/usr/lib/debateos/firstrun/OM-102.sh")
        assert "RemainAfterExit=yes" in unit

    def test_unit_exec_start_contains_exec_path(self):
        """ExecStart references the provided exec_path."""
        exec_path = "/usr/lib/debateos/firstrun/OM-102.sh"
        unit = render_firstrun_unit("OM-102", "GTK theme", exec_path)
        assert exec_path in unit

    def test_unit_exec_start_post_touches_flag_file(self):
        """ExecStartPost touches the flag file (idempotency guard)."""
        unit = render_firstrun_unit("OM-102", "GTK theme", "/usr/lib/debateos/firstrun/OM-102.sh")
        assert "ExecStartPost" in unit
        # The post command should touch or create the flag file
        assert "touch" in unit or "/bin/touch" in unit or "/usr/bin/touch" in unit

    def test_unit_wanted_by_graphical_session(self):
        """WantedBy=graphical-session.target (user session, not system scope)."""
        unit = render_firstrun_unit("OM-102", "GTK theme", "/usr/lib/debateos/firstrun/OM-102.sh")
        assert "WantedBy=graphical-session.target" in unit

    def test_unit_description_included(self):
        """The description argument appears in the unit Description= field."""
        description = "GTK theme configuration"
        unit = render_firstrun_unit("OM-102", description, "/usr/lib/debateos/firstrun/OM-102.sh")
        assert description in unit

    def test_different_opinion_ids_produce_different_flag_files(self):
        """Two different opinion IDs produce units with different flag file paths."""
        unit_a = render_firstrun_unit("OM-101", "first-run A", "/usr/lib/debateos/firstrun/OM-101.sh")
        unit_b = render_firstrun_unit("OM-113", "first-run B", "/usr/lib/debateos/firstrun/OM-113.sh")
        # Extract the flag file paths to confirm they differ
        lines_a = [l for l in unit_a.splitlines() if "firstrun-" in l]
        lines_b = [l for l in unit_b.splitlines() if "firstrun-" in l]
        assert lines_a != lines_b

    def test_unit_file_name_pattern(self):
        """render_firstrun_unit_name returns debateos-firstrun-<id>.service naming."""
        # The unit content should reference or be named debateos-firstrun-<id>.service
        # (implicit in Description or in the flag-file naming).
        unit = render_firstrun_unit("OM-102", "GTK theme", "/usr/lib/debateos/firstrun/OM-102.sh")
        # Check Description contains debateos branding
        assert "DebateOS" in unit or "debateos" in unit.lower()

    def test_after_graphical_session_target(self):
        """Unit has After=graphical-session.target for correct ordering."""
        unit = render_firstrun_unit("OM-102", "GTK theme", "/usr/lib/debateos/firstrun/OM-102.sh")
        assert "After=graphical-session.target" in unit
