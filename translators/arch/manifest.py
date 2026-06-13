"""
manifest.py — Re-export shim for the Arch translator.

The canonical implementation lives in translators/common/manifest.py.
This shim re-exports all public names so that existing bare-name imports
within the Arch translator continue to work unchanged:

    from manifest import BuildManifest, derive_source_date_epoch  # still works

Single source of truth: translators/common/manifest.py

Note on capability gate: the common/manifest.py BuildManifest.from_resolved
does NOT call check_capabilities — this is intentional (foundation-neutral
design). The Arch generator.py now calls check_capabilities() explicitly
before BuildManifest.from_resolved() to preserve the ARCH-03 / SC-3 gate.
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
