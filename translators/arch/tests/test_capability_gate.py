"""
Tests for the capability gate (ARCH-03 / SC-3).

RED phase: All tests in this file are written BEFORE any implementation.
They must fail when capabilities.py does not exist yet, and pass after implementation.
"""
import json
import os
import pytest
import sys

# Ensure the package root is on sys.path (pytest.ini sets pythonpath = .)
# but keep test self-contained.


# ---------------------------------------------------------------------------
# Fixtures helpers
# ---------------------------------------------------------------------------

FIXTURES_DIR = os.path.join(os.path.dirname(__file__), "fixtures")


def load_fixture(name):
    path = os.path.join(FIXTURES_DIR, name)
    with open(path) as f:
        return json.load(f)


# ---------------------------------------------------------------------------
# Tests: load_capabilities
# ---------------------------------------------------------------------------


def test_load_capabilities_returns_set():
    """load_capabilities() must return a set of strings."""
    from capabilities import load_capabilities  # noqa: E402

    caps = load_capabilities()
    assert isinstance(caps, set)
    assert len(caps) > 0


def test_load_capabilities_contains_install_named_packages():
    """capabilities.json must declare install-named-packages."""
    from capabilities import load_capabilities  # noqa: E402

    caps = load_capabilities()
    assert "install-named-packages" in caps


def test_load_capabilities_contains_add_signed_external_repo():
    """capabilities.json must declare add-signed-external-repo."""
    from capabilities import load_capabilities  # noqa: E402

    caps = load_capabilities()
    assert "add-signed-external-repo" in caps


# ---------------------------------------------------------------------------
# Tests: contract loaders
# ---------------------------------------------------------------------------


def test_load_resolved_speech_returns_dict_with_expected_keys():
    """load_resolved_speech must return a dict with applied/skipped/dropped/install_order/explanations."""
    from contract import load_resolved_speech  # noqa: E402

    rs = load_resolved_speech(os.path.join(FIXTURES_DIR, "minimal_resolved.json"))
    assert isinstance(rs, dict)
    for key in ("applied", "skipped", "dropped", "install_order", "explanations"):
        assert key in rs, f"missing key: {key}"


def test_load_resolved_speech_applied_is_list():
    """applied must be a list of opinion IDs."""
    from contract import load_resolved_speech  # noqa: E402

    rs = load_resolved_speech(os.path.join(FIXTURES_DIR, "minimal_resolved.json"))
    assert isinstance(rs["applied"], list)
    assert "OM-001" in rs["applied"]
    assert "OM-006" in rs["applied"]


def test_load_opinion_bodies_from_json_file():
    """load_opinion_bodies from a JSON file (array) returns a dict keyed by opinion id."""
    from contract import load_opinion_bodies  # noqa: E402

    idx = load_opinion_bodies(os.path.join(FIXTURES_DIR, "minimal_opinions.json"))
    assert isinstance(idx, dict)
    assert "OM-001" in idx
    assert "OM-006" in idx


def test_load_opinion_bodies_values_are_dicts():
    """Each value in the opinion index must be a dict."""
    from contract import load_opinion_bodies  # noqa: E402

    idx = load_opinion_bodies(os.path.join(FIXTURES_DIR, "minimal_opinions.json"))
    for v in idx.values():
        assert isinstance(v, dict)


# ---------------------------------------------------------------------------
# Tests: check_capabilities — happy path
# ---------------------------------------------------------------------------


def test_check_capabilities_passes_when_all_required_supported():
    """check_capabilities must not raise when all required capabilities are declared."""
    from capabilities import check_capabilities, CapabilityError  # noqa: E402

    rs = load_fixture("minimal_resolved.json")
    opinions = load_fixture("minimal_opinions.json")
    opinions_index = {op["id"]: op for op in opinions}
    caps = {"add-signed-external-repo", "import-gpg-key-by-fingerprint", "install-named-packages"}

    # Must not raise
    dropped = check_capabilities(rs, opinions_index, caps)
    assert isinstance(dropped, list)
    assert len(dropped) == 0


# ---------------------------------------------------------------------------
# Tests: check_capabilities — required opinion with unsupported capability
# ---------------------------------------------------------------------------


def test_unsupported_required_raises_capability_error():
    """A required opinion with an unsupported capability must raise CapabilityError."""
    from capabilities import check_capabilities, CapabilityError  # noqa: E402

    rs = load_fixture("unsupported_required_resolved.json")
    opinions = load_fixture("unsupported_required_opinions.json")
    opinions_index = {op["id"]: op for op in opinions}
    # Caps deliberately exclude "install-npm-global-packages"
    caps = {"install-named-packages", "add-signed-external-repo"}

    with pytest.raises(CapabilityError) as exc_info:
        check_capabilities(rs, opinions_index, caps)

    msg = str(exc_info.value)
    # Must name the opinion
    assert "OM-023" in msg
    # Must name the missing capability
    assert "install-npm-global-packages" in msg
    # Must say "composition time"
    assert "composition time" in msg


def test_unsupported_required_capability_error_message_format(
):
    """CapabilityError message must contain the opinion ID and capability name."""
    from capabilities import check_capabilities, CapabilityError  # noqa: E402

    rs = load_fixture("unsupported_required_resolved.json")
    opinions = load_fixture("unsupported_required_opinions.json")
    opinions_index = {op["id"]: op for op in opinions}
    caps = {"install-named-packages"}

    with pytest.raises(CapabilityError) as exc_info:
        check_capabilities(rs, opinions_index, caps)

    msg = str(exc_info.value)
    assert "OM-023" in msg
    assert "install-npm-global-packages" in msg
    assert "composition time" in msg


# ---------------------------------------------------------------------------
# Tests: check_capabilities — nice-to-have with unsupported capability
# ---------------------------------------------------------------------------


def test_nicetohave_drop_does_not_raise():
    """A nice-to-have opinion with an unsupported capability must NOT raise."""
    from capabilities import check_capabilities  # noqa: E402

    rs = load_fixture("minimal_resolved.json")
    # Inject a nice-to-have opinion that needs an unsupported capability
    opinions_index = {
        "OM-001": {
            "id": "OM-001",
            "status": "nice-to-have",
            "translator_capabilities": ["install-npm-global-packages"],
        },
        "OM-006": {
            "id": "OM-006",
            "status": "required",
            "translator_capabilities": ["install-named-packages"],
        },
    }
    caps = {"install-named-packages"}  # no npm support

    # Must not raise
    dropped = check_capabilities(rs, opinions_index, caps)
    assert isinstance(dropped, list)
    # The nice-to-have OM-001 must appear in dropped
    dropped_ids = [d[0] for d in dropped]
    assert "OM-001" in dropped_ids


def test_nicetohave_dropped_entry_has_reason():
    """Dropped (opinion_id, reason) tuples must have a non-empty reason."""
    from capabilities import check_capabilities  # noqa: E402

    rs = load_fixture("minimal_resolved.json")
    opinions_index = {
        "OM-001": {
            "id": "OM-001",
            "status": "nice-to-have",
            "translator_capabilities": ["install-npm-global-packages"],
        },
        "OM-006": {
            "id": "OM-006",
            "status": "required",
            "translator_capabilities": ["install-named-packages"],
        },
    }
    caps = {"install-named-packages"}

    dropped = check_capabilities(rs, opinions_index, caps)
    # Find OM-001 entry
    entry = next((d for d in dropped if d[0] == "OM-001"), None)
    assert entry is not None
    opinion_id, reason = entry
    assert reason  # non-empty
    assert "install-npm-global-packages" in reason


def test_nicetohave_with_supported_cap_not_in_dropped():
    """A nice-to-have opinion whose capabilities ARE declared must not appear in dropped."""
    from capabilities import check_capabilities  # noqa: E402

    rs = load_fixture("minimal_resolved.json")
    opinions_index = {
        "OM-001": {
            "id": "OM-001",
            "status": "nice-to-have",
            "translator_capabilities": ["add-signed-external-repo"],
        },
        "OM-006": {
            "id": "OM-006",
            "status": "required",
            "translator_capabilities": ["install-named-packages"],
        },
    }
    caps = {"install-named-packages", "add-signed-external-repo"}

    dropped = check_capabilities(rs, opinions_index, caps)
    dropped_ids = [d[0] for d in dropped]
    assert "OM-001" not in dropped_ids


# ---------------------------------------------------------------------------
# Tests: CapabilityError is importable
# ---------------------------------------------------------------------------


def test_capability_error_is_exception_subclass():
    """CapabilityError must be a subclass of Exception."""
    from capabilities import CapabilityError  # noqa: E402

    assert issubclass(CapabilityError, Exception)
    e = CapabilityError("test error")
    assert str(e) == "test error"
