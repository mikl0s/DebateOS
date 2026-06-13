"""
manifest.py — Re-export shim for the Debian translator.

The canonical implementation lives in translators/common/manifest.py.
This shim re-exports all public names so that bare-name imports within the
Debian translator continue to work:

    from manifest import BuildManifest, derive_source_date_epoch  # works

Single source of truth: translators/common/manifest.py (DEB-03 / COMM-01)
"""
# noqa: F401,F403
from common.manifest import (  # noqa: F401
    BuildManifest,
    derive_source_date_epoch,
    _MIN_EPOCH,
    _MAX_EPOCH,
)

__all__ = [
    "BuildManifest",
    "derive_source_date_epoch",
]
