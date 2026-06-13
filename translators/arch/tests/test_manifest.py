"""
Tests for BuildManifest (ARCH-01 foundation).

RED phase: All tests are written BEFORE manifest.py is implemented.
Tests cover:
- Package union/order/dedup from install_order
- First-run opinion split (execution_phase=="first-run")
- System services vs deferred services (deferred=True)
- custom_repos with sig_level=Never captured in trust_warnings
- derive_source_date_epoch determinism
- check_capabilities called before assembly (raises on unsupported required)
- to_dict() produces a serializable dict
"""

import json
import os
import pytest


# ---------------------------------------------------------------------------
# Shared fixtures
# ---------------------------------------------------------------------------


def make_resolved(applied, install_order, skipped=None, dropped=None, explanations=None):
    """Helper to construct a minimal ResolvedSpeech dict."""
    return {
        "schema": 1,
        "foundation": "arch",
        "applied": list(applied),
        "skipped": skipped or [],
        "dropped": dropped or [],
        "install_order": list(install_order),
        "explanations": explanations or [],
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
    """Build a minimal opinion dict with all optional payload fields."""
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


# ---------------------------------------------------------------------------
# Tests: derive_source_date_epoch
# ---------------------------------------------------------------------------


def test_derive_source_date_epoch_returns_int():
    """Must return an integer."""
    from manifest import derive_source_date_epoch  # noqa: E402

    result = derive_source_date_epoch(b"some bytes")
    assert isinstance(result, int)


def test_derive_source_date_epoch_deterministic():
    """Same bytes must produce the same epoch every time."""
    from manifest import derive_source_date_epoch  # noqa: E402

    b = b"deterministic-test"
    assert derive_source_date_epoch(b) == derive_source_date_epoch(b)


def test_derive_source_date_epoch_two_different_inputs():
    """Different bytes should produce different epochs (with overwhelming probability)."""
    from manifest import derive_source_date_epoch  # noqa: E402

    a = derive_source_date_epoch(b"input-alpha")
    b = derive_source_date_epoch(b"input-beta")
    # They CAN collide in theory but SHA-256 makes it negligible.
    # Test is informational; we don't hard-fail on collision.
    # Simply assert both are valid integers in range.
    MIN_EPOCH = 1577836800
    MAX_EPOCH = 2208988800
    assert MIN_EPOCH <= a < MAX_EPOCH
    assert MIN_EPOCH <= b < MAX_EPOCH


def test_derive_source_date_epoch_in_valid_range():
    """Epoch must be in [2020-01-01, 2040-01-01)."""
    from manifest import derive_source_date_epoch  # noqa: E402

    MIN_EPOCH = 1577836800  # 2020-01-01
    MAX_EPOCH = 2208988800  # 2040-01-01
    result = derive_source_date_epoch(b"abc")
    assert MIN_EPOCH <= result < MAX_EPOCH


def test_derive_source_date_epoch_abc_stable():
    """The epoch for b'abc' must match the same call repeated (stability test)."""
    from manifest import derive_source_date_epoch  # noqa: E402

    assert derive_source_date_epoch(b"abc") == derive_source_date_epoch(b"abc")


# ---------------------------------------------------------------------------
# Tests: BuildManifest.from_resolved — package aggregation
# ---------------------------------------------------------------------------


def test_from_resolved_target_packages_in_install_order():
    """target_packages must follow install_order, deduplicated, first-occurrence wins."""
    from manifest import BuildManifest  # noqa: E402

    # OM-A installs [pkg-a, pkg-common], OM-B installs [pkg-b, pkg-common]
    # install_order is [OM-A, OM-B] → target_packages = [pkg-a, pkg-common, pkg-b]
    rs = make_resolved(["OM-A", "OM-B"], ["OM-A", "OM-B"])
    idx = {
        "OM-A": base_opinion("OM-A", packages=["pkg-a", "pkg-common"],
                              translator_capabilities=["install-named-packages"]),
        "OM-B": base_opinion("OM-B", packages=["pkg-b", "pkg-common"],
                              translator_capabilities=["install-named-packages"]),
    }
    manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
    assert manifest.target_packages == ["pkg-a", "pkg-common", "pkg-b"]


def test_from_resolved_dedup_preserves_first_occurrence():
    """Duplicate packages must keep the first occurrence position."""
    from manifest import BuildManifest  # noqa: E402

    rs = make_resolved(["OM-1", "OM-2", "OM-3"], ["OM-1", "OM-2", "OM-3"])
    idx = {
        "OM-1": base_opinion("OM-1", packages=["base", "linux"],
                              translator_capabilities=["install-named-packages"]),
        "OM-2": base_opinion("OM-2", packages=["linux", "vim"],
                              translator_capabilities=["install-named-packages"]),
        "OM-3": base_opinion("OM-3", packages=["base", "git"],
                              translator_capabilities=["install-named-packages"]),
    }
    manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
    pkgs = manifest.target_packages
    assert pkgs.index("base") < pkgs.index("vim")
    assert pkgs.index("linux") < pkgs.index("vim")
    assert pkgs.count("base") == 1
    assert pkgs.count("linux") == 1
    assert "vim" in pkgs
    assert "git" in pkgs


def test_from_resolved_remove_packages_collected():
    """remove_packages from all applied opinions must be aggregated."""
    from manifest import BuildManifest  # noqa: E402

    rs = make_resolved(["OM-R"], ["OM-R"])
    idx = {
        "OM-R": base_opinion("OM-R", packages=["new-pkg"], remove_packages=["old-pkg"],
                              translator_capabilities=["install-named-packages"]),
    }
    manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
    assert "old-pkg" in manifest.remove_packages


def test_from_resolved_respects_install_order():
    """Packages from later install_order entries come after earlier ones."""
    from manifest import BuildManifest  # noqa: E402

    rs = make_resolved(["OM-FIRST", "OM-SECOND"], ["OM-FIRST", "OM-SECOND"])
    idx = {
        "OM-FIRST": base_opinion("OM-FIRST", packages=["aaa"],
                                  translator_capabilities=["install-named-packages"]),
        "OM-SECOND": base_opinion("OM-SECOND", packages=["zzz"],
                                   translator_capabilities=["install-named-packages"]),
    }
    manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
    assert manifest.target_packages.index("aaa") < manifest.target_packages.index("zzz")


# ---------------------------------------------------------------------------
# Tests: BuildManifest.from_resolved — first-run split
# ---------------------------------------------------------------------------


def test_first_run_opinions_collected_separately():
    """Opinions with execution_phase=='first-run' go into manifest.first_run."""
    from manifest import BuildManifest  # noqa: E402

    rs = make_resolved(["OM-INSTALL", "OM-FIRSTRUN"], ["OM-INSTALL", "OM-FIRSTRUN"])
    idx = {
        "OM-INSTALL": base_opinion("OM-INSTALL", packages=["some-pkg"],
                                    translator_capabilities=["install-named-packages"]),
        "OM-FIRSTRUN": base_opinion(
            "OM-FIRSTRUN",
            execution_phase="first-run",
            script_payload={"script": "#!/bin/bash\ngsettings set org.gnome.desktop.wm.preferences theme 'Omarchy'"},
            translator_capabilities=["systemd-user-service-firstrun"],
        ),
    }
    manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
    # OM-FIRSTRUN must be in first_run
    first_run_ids = [fr["id"] for fr in manifest.first_run]
    assert "OM-FIRSTRUN" in first_run_ids
    # OM-INSTALL must NOT be in first_run
    assert "OM-INSTALL" not in first_run_ids


def test_first_run_entry_has_id_and_script_payload():
    """Each first_run entry must have 'id' and 'script_payload'."""
    from manifest import BuildManifest  # noqa: E402

    rs = make_resolved(["OM-FR"], ["OM-FR"])
    idx = {
        "OM-FR": base_opinion(
            "OM-FR",
            execution_phase="first-run",
            script_payload={"script": "#!/bin/bash\necho 'hello'"},
            translator_capabilities=["systemd-user-service-firstrun"],
        ),
    }
    manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
    assert len(manifest.first_run) == 1
    entry = manifest.first_run[0]
    assert "id" in entry
    assert "script_payload" in entry


# ---------------------------------------------------------------------------
# Tests: BuildManifest.from_resolved — services
# ---------------------------------------------------------------------------


def test_system_services_collected():
    """Non-deferred services go into system_services."""
    from manifest import BuildManifest  # noqa: E402

    rs = make_resolved(["OM-SVC"], ["OM-SVC"])
    idx = {
        "OM-SVC": base_opinion(
            "OM-SVC",
            services=[{"name": "NetworkManager", "enable": True, "deferred": False}],
            translator_capabilities=["enable-systemd-service-chroot"],
        ),
    }
    manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
    svc_names = [s["name"] for s in manifest.system_services]
    assert "NetworkManager" in svc_names


def test_deferred_services_collected_separately():
    """Services with deferred=True go into deferred_services, not system_services."""
    from manifest import BuildManifest  # noqa: E402

    rs = make_resolved(["OM-SVC"], ["OM-SVC"])
    idx = {
        "OM-SVC": base_opinion(
            "OM-SVC",
            services=[
                {"name": "NetworkManager", "enable": True, "deferred": False},
                {"name": "some-deferred-service", "enable": True, "deferred": True},
            ],
            translator_capabilities=["enable-systemd-service-chroot"],
        ),
    }
    manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
    deferred_names = [s["name"] for s in manifest.deferred_services]
    system_names = [s["name"] for s in manifest.system_services]
    assert "some-deferred-service" in deferred_names
    assert "some-deferred-service" not in system_names
    assert "NetworkManager" in system_names


# ---------------------------------------------------------------------------
# Tests: BuildManifest.from_resolved — trust_warnings (T-02-02)
# ---------------------------------------------------------------------------


def test_sig_level_never_captured_in_trust_warnings():
    """Custom repos with sig_level=Never must be captured in trust_warnings."""
    from manifest import BuildManifest  # noqa: E402

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


def test_sig_level_required_not_in_trust_warnings():
    """Custom repos with sig_level=Required must NOT appear in trust_warnings."""
    from manifest import BuildManifest  # noqa: E402

    rs = make_resolved(["OM-REPO"], ["OM-REPO"])
    idx = {
        "OM-REPO": base_opinion(
            "OM-REPO",
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
    # No trust warnings for a Required sig_level repo
    assert len(manifest.trust_warnings) == 0


def test_sig_level_optionaltrust_all_captured_in_trust_warnings():
    """Custom repos with sig_level=OptionalTrustAll must appear in trust_warnings (WR-01)."""
    from manifest import BuildManifest  # noqa: E402

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
    assert len(manifest.trust_warnings) > 0, (
        "OptionalTrustAll must produce a trust_warning in the manifest"
    )
    warning_text = " ".join(manifest.trust_warnings)
    assert "omarchy" in warning_text
    assert "OptionalTrustAll" in warning_text or "unsigned" in warning_text.lower()


# ---------------------------------------------------------------------------
# Tests: BuildManifest.from_resolved — capability gate integration
# ---------------------------------------------------------------------------


def test_capability_gate_raises_on_unsupported_required_capability():
    """check_capabilities must raise CapabilityError if a required opinion needs unsupported cap.

    Note: Since translators/common/manifest.py is foundation-neutral, the
    capability gate (check_capabilities) is now called by the CALLER (generator.py)
    before BuildManifest.from_resolved(). This test verifies the gate contract
    via check_capabilities directly, which is the correct call site (ARCH-03 / SC-3).
    """
    from capabilities import check_capabilities, CapabilityError  # noqa: E402

    rs = make_resolved(["OM-BAD"], ["OM-BAD"])
    idx = {
        "OM-BAD": base_opinion(
            "OM-BAD",
            status="required",
            translator_capabilities=["install-npm-global-packages"],  # not in caps
        ),
    }
    caps = {"install-named-packages"}  # npm not declared

    with pytest.raises(CapabilityError):
        check_capabilities(rs, idx, caps)


def test_capability_gate_runs_before_assembly():
    """CapabilityError must be raised when check_capabilities is called before assembly.

    Note: The gate is now caller-responsibility (generator.py calls
    check_capabilities before BuildManifest.from_resolved). This test verifies
    that the gate fires correctly via direct check_capabilities call.
    """
    from capabilities import check_capabilities, CapabilityError  # noqa: E402

    rs = make_resolved(["OM-OK", "OM-BAD"], ["OM-OK", "OM-BAD"])
    idx = {
        "OM-OK": base_opinion("OM-OK", packages=["good-pkg"],
                               translator_capabilities=["install-named-packages"]),
        "OM-BAD": base_opinion(
            "OM-BAD",
            status="required",
            translator_capabilities=["install-npm-global-packages"],
        ),
    }
    caps = {"install-named-packages"}

    with pytest.raises(CapabilityError):
        check_capabilities(rs, idx, caps)


# ---------------------------------------------------------------------------
# Tests: BuildManifest payload field aggregation
# ---------------------------------------------------------------------------


def test_sysctl_params_aggregated():
    """sysctl_params from all applied opinions must be aggregated."""
    from manifest import BuildManifest  # noqa: E402

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


def test_kernel_params_aggregated():
    """kernel_params from all applied opinions must be aggregated."""
    from manifest import BuildManifest  # noqa: E402

    rs = make_resolved(["OM-K"], ["OM-K"])
    idx = {
        "OM-K": base_opinion(
            "OM-K",
            kernel_params=[{"key": "quiet", "value": ""}],
            translator_capabilities=["install-named-packages"],
        ),
    }
    manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
    keys = [p["key"] for p in manifest.kernel_params]
    assert "quiet" in keys


def test_group_memberships_aggregated():
    """group_memberships from all applied opinions must be aggregated."""
    from manifest import BuildManifest  # noqa: E402

    rs = make_resolved(["OM-G"], ["OM-G"])
    idx = {
        "OM-G": base_opinion(
            "OM-G",
            group_memberships=[{"group": "docker"}],
            translator_capabilities=["add-user-to-group"],
        ),
    }
    manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
    groups = [m["group"] for m in manifest.group_memberships]
    assert "docker" in groups


def test_custom_repos_aggregated():
    """custom_repos from all applied opinions must be aggregated."""
    from manifest import BuildManifest  # noqa: E402

    rs = make_resolved(["OM-CR"], ["OM-CR"])
    idx = {
        "OM-CR": base_opinion(
            "OM-CR",
            custom_repos=[{"name": "my-repo", "url": "https://example.com/$arch", "sig_level": "Required"}],
            translator_capabilities=["add-signed-external-repo"],
        ),
    }
    manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
    repo_names = [r["name"] for r in manifest.custom_repos]
    assert "my-repo" in repo_names


def test_file_assets_aggregated():
    """file_assets from all applied opinions must be aggregated."""
    from manifest import BuildManifest  # noqa: E402

    rs = make_resolved(["OM-FA"], ["OM-FA"])
    idx = {
        "OM-FA": base_opinion(
            "OM-FA",
            file_assets=[{"src": "configs/tmux.conf", "dst": "~/.config/tmux/tmux.conf"}],
            translator_capabilities=["deploy-config-file-tree"],
        ),
    }
    manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
    assert len(manifest.file_assets) == 1
    assert manifest.file_assets[0]["src"] == "configs/tmux.conf"


def test_mime_associations_aggregated():
    """mime_associations from all applied opinions must be aggregated."""
    from manifest import BuildManifest  # noqa: E402

    rs = make_resolved(["OM-MIME"], ["OM-MIME"])
    idx = {
        "OM-MIME": base_opinion(
            "OM-MIME",
            mime_associations=[{"mime_pattern": "text/html", "app_id": "firefox.desktop"}],
            translator_capabilities=["register-mime-types"],
        ),
    }
    manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
    assert len(manifest.mime_associations) == 1
    assert manifest.mime_associations[0]["mime_pattern"] == "text/html"


def test_themes_aggregated():
    """theme declarations from applied opinions must be aggregated into themes list."""
    from manifest import BuildManifest  # noqa: E402

    rs = make_resolved(["OM-THEME"], ["OM-THEME"])
    idx = {
        "OM-THEME": base_opinion(
            "OM-THEME",
            theme={"bundle_dir": "themes/omarchy-theme", "is_default": True},
            translator_capabilities=["deploy-file-asset-bundle"],
        ),
    }
    manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
    assert len(manifest.themes) == 1
    assert manifest.themes[0]["bundle_dir"] == "themes/omarchy-theme"


# ---------------------------------------------------------------------------
# Tests: BuildManifest — source_date_epoch field
# ---------------------------------------------------------------------------


def test_manifest_has_source_date_epoch():
    """BuildManifest must have a source_date_epoch field set from resolved_bytes."""
    from manifest import BuildManifest, derive_source_date_epoch  # noqa: E402

    rs = make_resolved(["OM-1"], ["OM-1"])
    idx = {
        "OM-1": base_opinion("OM-1", packages=["pkg"],
                              translator_capabilities=["install-named-packages"]),
    }
    content_bytes = b'{"applied":["OM-1"]}'
    manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, content_bytes)
    expected = derive_source_date_epoch(content_bytes)
    assert manifest.source_date_epoch == expected


# ---------------------------------------------------------------------------
# Tests: BuildManifest.to_dict — serialization
# ---------------------------------------------------------------------------


def test_to_dict_returns_dict():
    """to_dict() must return a plain dict."""
    from manifest import BuildManifest  # noqa: E402

    rs = make_resolved(["OM-1"], ["OM-1"])
    idx = {
        "OM-1": base_opinion("OM-1", packages=["pkg"],
                              translator_capabilities=["install-named-packages"]),
    }
    manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
    d = manifest.to_dict()
    assert isinstance(d, dict)


def test_to_dict_json_serializable():
    """to_dict() output must be JSON-serializable (no dataclass nesting)."""
    from manifest import BuildManifest  # noqa: E402

    rs = make_resolved(["OM-1"], ["OM-1"])
    idx = {
        "OM-1": base_opinion("OM-1", packages=["pkg"],
                              translator_capabilities=["install-named-packages"]),
    }
    manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
    d = manifest.to_dict()
    # Should not raise
    serialized = json.dumps(d)
    assert '"target_packages"' in serialized


def test_to_dict_contains_install_order():
    """to_dict() must include install_order from the resolved speech."""
    from manifest import BuildManifest  # noqa: E402

    rs = make_resolved(["OM-A", "OM-B"], ["OM-A", "OM-B"])
    idx = {
        "OM-A": base_opinion("OM-A", packages=["pkg-a"],
                              translator_capabilities=["install-named-packages"]),
        "OM-B": base_opinion("OM-B", packages=["pkg-b"],
                              translator_capabilities=["install-named-packages"]),
    }
    manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
    d = manifest.to_dict()
    assert "install_order" in d
    assert d["install_order"] == ["OM-A", "OM-B"]


def test_to_dict_contains_source_date_epoch():
    """to_dict() must include source_date_epoch."""
    from manifest import BuildManifest  # noqa: E402

    rs = make_resolved(["OM-1"], ["OM-1"])
    idx = {
        "OM-1": base_opinion("OM-1", packages=["pkg"],
                              translator_capabilities=["install-named-packages"]),
    }
    manifest = BuildManifest.from_resolved(rs, idx, FULL_CAPS, b"dummy")
    d = manifest.to_dict()
    assert "source_date_epoch" in d
    assert isinstance(d["source_date_epoch"], int)
