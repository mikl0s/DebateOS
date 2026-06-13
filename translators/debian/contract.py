"""
contract.py — Re-export shim for the Debian translator.

The canonical implementation lives in translators/common/contract.py.
This shim re-exports all public names so that bare-name imports within the
Debian translator continue to work:

    from contract import load_resolved_speech, load_opinion_bodies  # works

Single source of truth: translators/common/contract.py (DEB-03 / COMM-01)
"""
# noqa: F401,F403
from common.contract import (  # noqa: F401
    load_resolved_speech,
    load_opinion_bodies,
    _load_opinions_from_json_file,
    _load_opinions_from_directory,
    _RESOLVED_SPEECH_KEYS,
)

__all__ = [
    "load_resolved_speech",
    "load_opinion_bodies",
]
