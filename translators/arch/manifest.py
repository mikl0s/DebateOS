"""
manifest.py — Re-export shim for the Arch translator.

The canonical implementation lives in translators/common/manifest.py.
This shim re-exports the public API so that existing bare-name imports
within the Arch translator continue to work unchanged:

    from manifest import BuildManifest, derive_source_date_epoch  # still works

WR-04 fix: private names (_MIN_EPOCH, _MAX_EPOCH) are NOT re-exported here.
Any code that needs them should import directly from translators.common.manifest
to make the dependency explicit and avoid silent breakage if internals are renamed.

Single source of truth: translators/common/manifest.py

Note on capability gate: the common/manifest.py BuildManifest.from_resolved
does NOT call check_capabilities — this is intentional (foundation-neutral
design). The Arch generator.py now calls check_capabilities() explicitly
before BuildManifest.from_resolved() to preserve the ARCH-03 / SC-3 gate.
"""
# noqa: F401
from common.manifest import (  # noqa: F401
    BuildManifest,
    derive_source_date_epoch,
)

__all__ = [
    "BuildManifest",
    "derive_source_date_epoch",
]
