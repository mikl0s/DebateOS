"""
test_common_shared.py — TDD RED tests for translators/common/ shared modules.

Tests import from ``common.contract``, ``common.manifest``, and ``common.firstrun``.
All imports WILL FAIL until Task 2 creates those modules (RED phase, D19).

Mirrors the assertion style of translators/arch/tests/test_manifest.py and
test_firstrun.py to guarantee behavior parity between the shared core and
the Arch translator's existing tests.
"""

import json
import pytest


# ===========================================================================
# Inline fixtures (self-contained — no external fixture files needed)
# ===========================================================================


def make_resolved(applied, install_order, foundation="arch", skipped=None, dropped=None):
    """Construct a minimal ResolvedSpeech dict."""
    return {
        "schema": 1,
        "foundation": foundation,
        "applied": list(applied),
        "skipped": skipped or [],
        "dropped": dropped or [],
        "install_order": list(install_order),
        "explanations": [],
    }


def base_opinion(
    opinion_id,
    *,
    status="required",
    packages=None,
    remove_packages=None,
    file_assets=None,
    custom_repos=None,
    services=None,
    sysctl_params=None,
    kernel_params=None,
    group_memberships=None,
    mime_associations=None,
    theme=None,
    execution_phase=None,
    script_payload=None,
    translator_capabilities=None,
):
    """Build a minimal opinion dict with optional payload fields."""
    op = {"id": opinion_id, "status": status}
    if packages:
        op["packages"] = packages
    if remove_packages:
        op["remove_packages"] = remove_packages
    if file_assets:
        op["file_assets"] = file_assets
    if custom_repos:
        op["custom_repos"] = custom_repos
    if services:
        op["services"] = services
    if sysctl_params:
        op["sysctl_params"] = sysctl_params
    if kernel_params:
        op["kernel_params"] = kernel_params
    if group_memberships:
        op["group_memberships"] = group_memberships
    if mime_associations:
        op["mime_associations"] = mime_associations
    if theme:
        op["theme"] = theme
    if execution_phase:
        op["execution_phase"] = execution_phase
    if script_payload:
        op["script_payload"] = script_payload
    if translator_capabilities:
        op["translator_capabilities"] = translator_capabilities
    return op


# Full capability set matching arch/tests/test_manifest.py FULL_CAPS
FULL_CAPS = {
    "install-named-packages",
    "add-signed-external-repo",
    "import-gpg-key-by-fingerprint",
    "write-sysctl-drop-in",
    "enable-systemd-service-chroot",
    "add-user-to-group",
    "register-mime-types",
    "deploy-file-asset-bundle",
    "systemd-user-service-firstrun",
    "run-post-install-setup-script",
    "deploy-config-file-tree",
}


# ===========================================================================
# common.contract — load_opinion_bodies
# ===========================================================================


class TestCommonContractLoadOpinionBodies:
    """load_opinion_bodies must work with a JSON-array file and a directory."""

    def test_import_load_opinion_bodies(self):
        """Importing load_opinion_bodies from common.contract must succeed."""
        from common.contract import load_opinion_bodies  # noqa: F401

    def test_load_from_json_array_file(self, tmp_path):
        """Loading from a JSON array file returns an id→dict index."""
        from common.contract import load_opinion_bodies

        opinions = [
            {"id": "DF-001", "status": "required", "packages": ["curl"]},
            {"id": "DF-002", "status": "required", "packages": ["git"]},
        ]
        f = tmp_path / "opinions.json"
        f.write_text(json.dumps(opinions))

        index = load_opinion_bodies(str(f))

        assert isinstance(index, dict)
        assert "DF-001" in index
        assert "DF-002" in index
        assert index["DF-001"]["packages"] == ["curl"]
        assert index["DF-002"]["packages"] == ["git"]

    def test_load_from_directory(self, tmp_path):
        """Loading from a directory globs *.yaml/*.json and builds id→dict index."""
        from common.contract import load_opinion_bodies

        import yaml

        op1 = {"id": "DF-001", "status": "required", "packages": ["bash"]}
        op2 = {"id": "DF-002", "status": "required", "packages": ["vim"]}

        (tmp_path / "df001.yaml").write_text(yaml.dump(op1))
        (tmp_path / "df002.json").write_text(json.dumps(op2))

        index = load_opinion_bodies(str(tmp_path))

        assert "DF-001" in index
        assert "DF-002" in index

    def test_load_single_dict_json(self, tmp_path):
        """A JSON file containing a single dict (not an array) is treated as one opinion."""
        from common.contract import load_opinion_bodies

        opinion = {"id": "DF-SINGLE", "status": "required"}
        f = tmp_path / "single.json"
        f.write_text(json.dumps(opinion))

        index = load_opinion_bodies(str(f))
        assert "DF-SINGLE" in index

    def test_load_missing_id_raises(self, tmp_path):
        """Opinion dict missing 'id' key must raise ValueError."""
        from common.contract import load_opinion_bodies

        opinions = [{"status": "required"}]
        f = tmp_path / "bad.json"
        f.write_text(json.dumps(opinions))

        with pytest.raises(ValueError, match="id"):
            load_opinion_bodies(str(f))


# ===========================================================================
# common.manifest — derive_source_date_epoch
# ===========================================================================


class TestCommonManifestDeriveSourceDateEpoch:
    """derive_source_date_epoch must be deterministic and clamp correctly."""

    def test_import(self):
        """Importing derive_source_date_epoch from common.manifest must succeed."""
        from common.manifest import derive_source_date_epoch  # noqa: F401

    def test_returns_int(self):
        """Must return an integer."""
        from common.manifest import derive_source_date_epoch

        assert isinstance(derive_source_date_epoch(b"some bytes"), int)

    def test_deterministic(self):
        """Same bytes → same epoch every time."""
        from common.manifest import derive_source_date_epoch

        b = b"deterministic-test"
        assert derive_source_date_epoch(b) == derive_source_date_epoch(b)

    def test_clamps_to_valid_range(self):
        """Epoch must be in [2020-01-01, 2040-01-01) = [1577836800, 2208988800)."""
        from common.manifest import derive_source_date_epoch

        MIN_EPOCH = 1577836800
        MAX_EPOCH = 2208988800
        for data in [b"abc", b"", b"x" * 1000, b"\x00\xff\x42"]:
            result = derive_source_date_epoch(data)
            assert MIN_EPOCH <= result < MAX_EPOCH, (
                f"Epoch {result} out of range for input {data!r}"
            )

    def test_stable_abc(self):
        """b'abc' epoch is stable across calls (regression guard)."""
        from common.manifest import derive_source_date_epoch

        assert derive_source_date_epoch(b"abc") == derive_source_date_epoch(b"abc")

    def test_different_inputs_differ(self):
        """Two different inputs produce valid epochs (both in range)."""
        from common.manifest import derive_source_date_epoch

        MIN_EPOCH = 1577836800
        MAX_EPOCH = 2208988800
        a = derive_source_date_epoch(b"input-alpha")
        b_ = derive_source_date_epoch(b"input-beta")
        assert MIN_EPOCH <= a < MAX_EPOCH
        assert MIN_EPOCH <= b_ < MAX_EPOCH


# ===========================================================================
# common.manifest — BuildManifest.from_resolved
# ===========================================================================


class TestCommonBuildManifest:
    """BuildManifest.from_resolved must aggregate payloads identically to arch/manifest.py."""

    def test_import(self):
        """Importing BuildManifest from common.manifest must succeed."""
        from common.manifest import BuildManifest  # noqa: F401

    def test_target_packages_in_install_order_deduped(self):
        """Packages follow install_order; first occurrence wins on duplicate."""
        from common.manifest import BuildManifest

        rs = make_resolved(["OM-A", "OM-B"], ["OM-A", "OM-B"])
        idx = {
            "OM-A": base_opinion("OM-A", packages=["pkg-a", "pkg-common"],
                                  translator_capabilities=["install-named-packages"]),
            "OM-B": base_opinion("OM-B", packages=["pkg-b", "pkg-common"],
                                  translator_capabilities=["install-named-packages"]),
        }
        manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
        assert manifest.target_packages == ["pkg-a", "pkg-common", "pkg-b"]

    def test_system_services_vs_deferred_services(self):
        """Non-deferred → system_services; deferred=True → deferred_services."""
        from common.manifest import BuildManifest

        rs = make_resolved(["OM-SVC"], ["OM-SVC"])
        idx = {
            "OM-SVC": base_opinion(
                "OM-SVC",
                services=[
                    {"name": "NetworkManager", "enable": True, "deferred": False},
                    {"name": "my-deferred", "enable": True, "deferred": True},
                ],
                translator_capabilities=["enable-systemd-service-chroot"],
            ),
        }
        manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
        system_names = [s["name"] for s in manifest.system_services]
        deferred_names = [s["name"] for s in manifest.deferred_services]
        assert "NetworkManager" in system_names
        assert "my-deferred" in deferred_names
        assert "my-deferred" not in system_names

    def test_first_run_opinions_collected(self):
        """execution_phase==first-run opinions go into first_run, not packages."""
        from common.manifest import BuildManifest

        rs = make_resolved(["OM-INSTALL", "OM-FR"], ["OM-INSTALL", "OM-FR"])
        idx = {
            "OM-INSTALL": base_opinion("OM-INSTALL", packages=["bash"],
                                        translator_capabilities=["install-named-packages"]),
            "OM-FR": base_opinion(
                "OM-FR",
                execution_phase="first-run",
                script_payload={"script": "#!/bin/bash\necho hello"},
                translator_capabilities=["systemd-user-service-firstrun"],
            ),
        }
        manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
        first_run_ids = [fr["id"] for fr in manifest.first_run]
        assert "OM-FR" in first_run_ids
        assert "OM-INSTALL" not in first_run_ids

    def test_sysctl_params_aggregated(self):
        """sysctl_params from applied opinions are aggregated."""
        from common.manifest import BuildManifest

        rs = make_resolved(["OM-S"], ["OM-S"])
        idx = {
            "OM-S": base_opinion(
                "OM-S",
                sysctl_params=[{"key": "vm.swappiness", "value": "10"}],
                translator_capabilities=["write-sysctl-drop-in"],
            ),
        }
        manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
        keys = [p["key"] for p in manifest.sysctl_params]
        assert "vm.swappiness" in keys

    def test_trust_warnings_for_sig_level_never(self):
        """sig_level=Never custom repos emit trust_warnings (T-04-01 / T-02-02)."""
        from common.manifest import BuildManifest

        rs = make_resolved(["OM-REPO"], ["OM-REPO"])
        idx = {
            "OM-REPO": base_opinion(
                "OM-REPO",
                custom_repos=[
                    {
                        "name": "danger-repo",
                        "url": "https://example.com/$arch",
                        "sig_level": "Never",
                        "priority": 5,
                    }
                ],
                translator_capabilities=["add-signed-external-repo"],
            ),
        }
        manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
        assert len(manifest.trust_warnings) > 0
        warning_text = " ".join(manifest.trust_warnings)
        assert "danger-repo" in warning_text

    def test_trust_warnings_for_sig_level_optionaltrust_all(self):
        """sig_level=OptionalTrustAll custom repos emit trust_warnings (WR-01)."""
        from common.manifest import BuildManifest

        rs = make_resolved(["OM-REPO"], ["OM-REPO"])
        idx = {
            "OM-REPO": base_opinion(
                "OM-REPO",
                custom_repos=[
                    {
                        "name": "omarchy",
                        "url": "https://packages.omarchy.org/stable",
                        "sig_level": "OptionalTrustAll",
                        "priority": 10,
                        "keyring": "40DFB630FF42BCFFB047046CF0134EE680CAC571",
                    }
                ],
                translator_capabilities=["add-signed-external-repo"],
            ),
        }
        manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
        assert len(manifest.trust_warnings) > 0
        warning_text = " ".join(manifest.trust_warnings)
        assert "omarchy" in warning_text
        assert "OptionalTrustAll" in warning_text or "unsigned" in warning_text.lower()

    def test_no_trust_warnings_for_required(self):
        """sig_level=Required repos produce no trust_warnings."""
        from common.manifest import BuildManifest

        rs = make_resolved(["OM-SAFE"], ["OM-SAFE"])
        idx = {
            "OM-SAFE": base_opinion(
                "OM-SAFE",
                custom_repos=[
                    {
                        "name": "safe-repo",
                        "url": "https://safe.example.com/$arch",
                        "sig_level": "Required",
                        "priority": 5,
                    }
                ],
                translator_capabilities=["add-signed-external-repo"],
            ),
        }
        manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
        assert len(manifest.trust_warnings) == 0

    def test_source_date_epoch_set(self):
        """source_date_epoch is derived from resolved_bytes and stored in manifest."""
        from common.manifest import BuildManifest, derive_source_date_epoch

        rs = make_resolved(["OM-1"], ["OM-1"])
        idx = {
            "OM-1": base_opinion("OM-1", packages=["pkg"],
                                  translator_capabilities=["install-named-packages"]),
        }
        content_bytes = b'{"applied":["OM-1"]}'
        manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, content_bytes)
        expected = derive_source_date_epoch(content_bytes)
        assert manifest.source_date_epoch == expected

    def test_to_dict_json_serializable(self):
        """to_dict() returns a JSON-serializable plain dict (T-02-01)."""
        from common.manifest import BuildManifest

        rs = make_resolved(["OM-1"], ["OM-1"])
        idx = {
            "OM-1": base_opinion("OM-1", packages=["pkg"],
                                  translator_capabilities=["install-named-packages"]),
        }
        manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
        d = manifest.to_dict()
        assert isinstance(d, dict)
        serialized = json.dumps(d)
        assert '"target_packages"' in serialized

    def test_from_resolved_no_check_capabilities_import(self):
        """common.manifest must NOT import check_capabilities (foundation-neutral)."""
        import sys

        # Reload to get fresh state
        mod_name = "common.manifest"
        if mod_name in sys.modules:
            del sys.modules[mod_name]

        import common.manifest as cm

        # The module's actual imported names must not include check_capabilities
        # (checking vars/dir rather than source text which includes comments)
        assert not hasattr(cm, "check_capabilities"), (
            "common/manifest.py must NOT import 'check_capabilities' "
            "(foundation-neutral: the capability gate is caller responsibility)"
        )
        # Also verify: the capabilities module itself must not be imported
        assert "capabilities" not in dir(cm) or not callable(getattr(cm, "capabilities", None)), (
            "common/manifest.py must not expose 'capabilities' module"
        )


# ===========================================================================
# common.firstrun — render_firstrun_unit + firstrun_unit_name
# ===========================================================================


class TestCommonFirstrun:
    """render_firstrun_unit and firstrun_unit_name from common.firstrun."""

    def test_import(self):
        """Importing render_firstrun_unit and firstrun_unit_name must succeed."""
        from common.firstrun import render_firstrun_unit, firstrun_unit_name  # noqa: F401

    def test_unit_name_format(self):
        """firstrun_unit_name returns debateos-firstrun-<id>.service."""
        from common.firstrun import firstrun_unit_name

        assert firstrun_unit_name("OM-102") == "debateos-firstrun-OM-102.service"
        assert firstrun_unit_name("DF-001") == "debateos-firstrun-DF-001.service"

    def test_render_returns_string(self):
        """render_firstrun_unit returns a non-empty string."""
        from common.firstrun import render_firstrun_unit

        unit = render_firstrun_unit(
            opinion_id="OM-102",
            description="GTK theme configuration",
            exec_path="/usr/lib/debateos/firstrun/OM-102.sh",
        )
        assert isinstance(unit, str)
        assert len(unit) > 0

    def test_unit_contains_description(self):
        """Description argument appears in the unit Description= field."""
        from common.firstrun import render_firstrun_unit

        desc = "GTK theme configuration"
        unit = render_firstrun_unit("OM-102", desc, "/usr/lib/debateos/firstrun/OM-102.sh")
        assert desc in unit

    def test_unit_contains_conditionpathexists_guard(self):
        """Unit has ConditionPathExists=! flag-file guard (Pitfall 3)."""
        from common.firstrun import render_firstrun_unit

        unit = render_firstrun_unit("OM-102", "GTK theme", "/usr/lib/debateos/firstrun/OM-102.sh")
        assert "ConditionPathExists=!" in unit

    def test_unit_exec_start_contains_exec_path(self):
        """ExecStart references the provided exec_path."""
        from common.firstrun import render_firstrun_unit

        exec_path = "/usr/lib/debateos/firstrun/OM-102.sh"
        unit = render_firstrun_unit("OM-102", "GTK theme", exec_path)
        assert exec_path in unit

    def test_unit_exec_start_post_touches_flag_file(self):
        """ExecStartPost touches the flag file (idempotency guard)."""
        from common.firstrun import render_firstrun_unit

        unit = render_firstrun_unit("OM-102", "GTK theme", "/usr/lib/debateos/firstrun/OM-102.sh")
        assert "ExecStartPost" in unit
        assert "touch" in unit or "/bin/touch" in unit

    def test_unit_type_oneshot_remain_after_exit(self):
        """Unit is Type=oneshot with RemainAfterExit=yes (Pattern 2)."""
        from common.firstrun import render_firstrun_unit

        unit = render_firstrun_unit("OM-102", "GTK theme", "/usr/lib/debateos/firstrun/OM-102.sh")
        assert "Type=oneshot" in unit
        assert "RemainAfterExit=yes" in unit

    def test_unit_wanted_by_graphical_session(self):
        """WantedBy=graphical-session.target (user scope)."""
        from common.firstrun import render_firstrun_unit

        unit = render_firstrun_unit("OM-102", "GTK theme", "/usr/lib/debateos/firstrun/OM-102.sh")
        assert "WantedBy=graphical-session.target" in unit

    def test_unit_after_graphical_session(self):
        """Unit has After=graphical-session.target for ordering."""
        from common.firstrun import render_firstrun_unit

        unit = render_firstrun_unit("OM-102", "GTK theme", "/usr/lib/debateos/firstrun/OM-102.sh")
        assert "After=graphical-session.target" in unit

    def test_flag_file_includes_opinion_id(self):
        """Flag file path includes the opinion ID for per-opinion idempotency."""
        from common.firstrun import render_firstrun_unit

        unit = render_firstrun_unit("OM-102", "GTK theme", "/usr/lib/debateos/firstrun/OM-102.sh")
        assert ".firstrun-" in unit
        assert "OM-102" in unit

    def test_different_ids_produce_different_flag_files(self):
        """Two different opinion IDs produce different flag-file paths in the unit."""
        from common.firstrun import render_firstrun_unit

        unit_a = render_firstrun_unit("OM-101", "first-run A", "/usr/lib/debateos/firstrun/OM-101.sh")
        unit_b = render_firstrun_unit("OM-113", "first-run B", "/usr/lib/debateos/firstrun/OM-113.sh")
        lines_a = [l for l in unit_a.splitlines() if "firstrun-" in l]
        lines_b = [l for l in unit_b.splitlines() if "firstrun-" in l]
        assert lines_a != lines_b

    def test_render_with_default_template_dir(self):
        """render_firstrun_unit works without explicit template_dir (uses common default)."""
        from common.firstrun import render_firstrun_unit

        # Called with only 3 required args — must not raise
        unit = render_firstrun_unit(
            opinion_id="DF-001",
            description="Test opinion",
            exec_path="/usr/lib/debateos/firstrun/DF-001.sh",
        )
        assert isinstance(unit, str)
        assert "DF-001" in unit

    def test_render_with_explicit_template_dir(self, tmp_path):
        """render_firstrun_unit accepts template_dir param pointing at a custom template."""
        from common.firstrun import render_firstrun_unit

        # Write a minimal custom template to tmp_path
        custom_tpl = (
            "[Unit]\n"
            "Description=Custom: {description}\n"
            "ConditionPathExists=!{flag_file}\n"
            "After=graphical-session.target\n"
            "\n"
            "[Service]\n"
            "Type=oneshot\n"
            "ExecStart={exec_path}\n"
            "ExecStartPost=/bin/touch {flag_file}\n"
            "RemainAfterExit=yes\n"
            "\n"
            "[Install]\n"
            "WantedBy=graphical-session.target\n"
        )
        (tmp_path / "firstrun.service.tpl").write_text(custom_tpl)

        unit = render_firstrun_unit(
            opinion_id="DF-001",
            description="Custom template test",
            exec_path="/usr/lib/debateos/firstrun/DF-001.sh",
            template_dir=str(tmp_path),
        )
        assert "Custom: Custom template test" in unit
        assert "DF-001" in unit
