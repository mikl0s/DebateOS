"""
contract.py — Re-export shim for the Arch translator.

The canonical implementation lives in translators/common/contract.py.
This shim re-exports the public API so that existing bare-name imports
within the Arch translator continue to work unchanged:

    from contract import load_resolved_speech, load_opinion_bodies  # still works

WR-04 fix: private names (_load_opinions_from_json_file, _load_opinions_from_directory,
_RESOLVED_SPEECH_KEYS) are NOT re-exported here. Any code that needs them should
import directly from translators.common.contract to make the dependency explicit
and avoid silent breakage when common/ internals are renamed.

Single source of truth: translators/common/contract.py
"""
# noqa: F401
from common.contract import (  # noqa: F401
    load_resolved_speech,
    load_opinion_bodies,
)

__all__ = [
    "load_resolved_speech",
    "load_opinion_bodies",
]
