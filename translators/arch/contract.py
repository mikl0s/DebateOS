"""
contract.py — Re-export shim for the Arch translator.

The canonical implementation lives in translators/common/contract.py.
This shim re-exports all public names so that existing bare-name imports
within the Arch translator continue to work unchanged:

    from contract import load_resolved_speech, load_opinion_bodies  # still works

Single source of truth: translators/common/contract.py
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
