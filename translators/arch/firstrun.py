"""
firstrun.py — Re-export shim for the Arch translator.

The canonical implementation lives in translators/common/firstrun.py.
This shim re-exports the public API so that existing bare-name imports
within the Arch translator continue to work unchanged:

    from firstrun import render_firstrun_unit, firstrun_unit_name  # still works

WR-04 fix: private names (_DEFAULT_TEMPLATES_DIR, _flag_file_path, _load_unit_template)
are NOT re-exported here. Any code that needs them should import directly from
translators.common.firstrun to make the dependency explicit and avoid silent
breakage if common/ internals are renamed.

Single source of truth: translators/common/firstrun.py

The common implementation defaults to translators/common/templates/ for the
firstrun.service.tpl template. The Arch translator calls render_firstrun_unit
without a template_dir override (the common template is identical to the
former arch template — byte-for-byte identical).
"""
# noqa: F401
from common.firstrun import (  # noqa: F401
    render_firstrun_unit,
    firstrun_unit_name,
    FLAG_FILE_DIR,
    FLAG_FILE_PREFIX,
)

__all__ = [
    "render_firstrun_unit",
    "firstrun_unit_name",
    "FLAG_FILE_DIR",
    "FLAG_FILE_PREFIX",
]
