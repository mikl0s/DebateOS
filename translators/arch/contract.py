"""
contract.py — Input contract loaders for the Arch translator.

Provides two loaders:
- ``load_resolved_speech(path)``  — reads the Phase 1 ResolvedSpeech JSON and
  returns a plain dict with keys: applied, skipped, dropped, install_order,
  explanations (all lists; explanations is a list of dicts).
- ``load_opinion_bodies(path)``   — accepts either:
    * a JSON file (array of Opinion dicts), or
    * a directory of *.yaml / *.json opinion files (single dict or list per file),
  and returns a ``dict[str, dict]`` mapping opinion ID → opinion dict.

Only stdlib json and PyYAML are used (no other dependencies per CONTEXT D).
"""

import glob
import json
import os

import yaml  # PyYAML


# ---------------------------------------------------------------------------
# load_resolved_speech
# ---------------------------------------------------------------------------

_RESOLVED_SPEECH_KEYS = ("applied", "skipped", "dropped", "install_order", "explanations")


def load_resolved_speech(path: str) -> dict:
    """Load and validate a ResolvedSpeech JSON file.

    Args:
        path: Path to the resolved speech JSON file emitted by the Phase 1
            resolver (``CanonicalJSON`` output).

    Returns:
        A dict with at least the keys ``applied``, ``skipped``, ``dropped``,
        ``install_order``, and ``explanations``.  All list fields default to
        empty lists if absent (robustness against older resolver output that
        omits empty slices via ``omitempty``).

    Raises:
        FileNotFoundError: if ``path`` does not exist.
        json.JSONDecodeError: if the file is not valid JSON.
    """
    with open(path) as fh:
        data = json.load(fh)

    # Ensure every expected key is present; default to empty list.
    # All keys (including explanations) default to [] — both branches of the
    # original ternary returned [], so the conditional was dead code (WR-04).
    for key in _RESOLVED_SPEECH_KEYS:
        if key not in data:
            data[key] = []

    return data


# ---------------------------------------------------------------------------
# load_opinion_bodies
# ---------------------------------------------------------------------------


def load_opinion_bodies(path: str) -> dict:
    """Load opinion bodies from a JSON file or a directory of YAML/JSON files.

    Accepted input formats:
    - **JSON file** (``*.json``): Must contain a JSON array of opinion dicts.
      Each dict must have an ``"id"`` key.
    - **Directory**: Globs for ``*.yaml``, ``*.yml``, and ``*.json`` files.
      Each file may contain a single opinion dict or a list of opinion dicts.
      All opinions from all files are merged into one index.

    Args:
        path: Path to a JSON file or a directory.

    Returns:
        ``dict[str, dict]`` mapping opinion ID string → opinion dict.

    Raises:
        FileNotFoundError: if ``path`` does not exist.
        ValueError: if a dict is missing the required ``"id"`` key.
    """
    if os.path.isdir(path):
        return _load_opinions_from_directory(path)
    else:
        return _load_opinions_from_json_file(path)


def _load_opinions_from_json_file(path: str) -> dict:
    """Load opinion bodies from a JSON array file."""
    with open(path) as fh:
        data = json.load(fh)

    if isinstance(data, dict):
        # Treat as a single opinion dict.
        data = [data]
    elif not isinstance(data, list):
        raise ValueError(f"Expected a JSON array or dict in {path!r}, got {type(data).__name__}")

    index = {}
    for item in data:
        if not isinstance(item, dict):
            raise ValueError(f"Expected opinion dicts in {path!r}, got {type(item).__name__}")
        if "id" not in item:
            raise ValueError(f"Opinion dict missing 'id' key in {path!r}: {item!r}")
        index[str(item["id"])] = item
    return index


def _load_opinions_from_directory(directory: str) -> dict:
    """Load opinion bodies from all *.yaml, *.yml, *.json files in a directory."""
    index = {}
    patterns = [
        os.path.join(directory, "*.yaml"),
        os.path.join(directory, "*.yml"),
        os.path.join(directory, "*.json"),
    ]
    files = []
    for pattern in patterns:
        files.extend(sorted(glob.glob(pattern)))

    for filepath in files:
        ext = os.path.splitext(filepath)[1].lower()
        with open(filepath) as fh:
            if ext == ".json":
                data = json.load(fh)
            else:
                data = yaml.safe_load(fh)

        if data is None:
            continue  # empty file

        if isinstance(data, dict):
            data = [data]

        for item in data:
            if not isinstance(item, dict):
                continue
            if "id" not in item:
                raise ValueError(
                    f"Opinion dict missing 'id' key in {filepath!r}: {item!r}"
                )
            index[str(item["id"])] = item

    return index
