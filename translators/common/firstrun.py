"""
firstrun.py — First-run systemd user oneshot unit generator (foundation-neutral).

Provides:
- render_firstrun_unit(opinion_id, description, exec_path, template_dir=None) → str
- firstrun_unit_name(opinion_id) → str

Generated units follow the flag-file guard pattern (02-RESEARCH.md Pattern 2):
- User-scope systemd oneshot service (not system scope)
- ConditionPathExists=!/var/lib/debateos/.firstrun-<id>.done (flag-file guard)
- ExecStartPost=/bin/touch <flag-file> (idempotency)
- RemainAfterExit=yes
- WantedBy=graphical-session.target (requires live user session)
- After=graphical-session.target (ordering)

Foundation-neutral parameterization:
- ``template_dir`` defaults to ``translators/common/templates/`` when not
  provided. Any translator can pass its own template directory to use a
  customized firstrun.service.tpl (e.g. if the template requires translation-
  specific adjustments for a future foundation).

Source: Adapted from translators/arch/firstrun.py — the only change is
parameterizing ``_TEMPLATES_DIR`` via the ``template_dir`` parameter.
"""

import os

# ---------------------------------------------------------------------------
# Constants
# ---------------------------------------------------------------------------

FLAG_FILE_DIR = "/var/lib/debateos"
FLAG_FILE_PREFIX = ".firstrun-"

# Default template directory: translators/common/templates/
_DEFAULT_TEMPLATES_DIR = os.path.join(os.path.dirname(__file__), "templates")


# ---------------------------------------------------------------------------
# Unit file name helper
# ---------------------------------------------------------------------------


def firstrun_unit_name(opinion_id: str) -> str:
    """Return the canonical service unit filename for a first-run opinion.

    Format: debateos-firstrun-<opinion_id>.service

    Args:
        opinion_id: The opinion ID string, e.g. "OM-102".

    Returns:
        The unit filename string, e.g. "debateos-firstrun-OM-102.service".
    """
    return f"debateos-firstrun-{opinion_id}.service"


def _flag_file_path(opinion_id: str) -> str:
    """Return the absolute flag file path for a first-run opinion.

    Format: /var/lib/debateos/.firstrun-<opinion_id>.done

    Args:
        opinion_id: The opinion ID string.

    Returns:
        Absolute path string for the flag file.
    """
    return f"{FLAG_FILE_DIR}/{FLAG_FILE_PREFIX}{opinion_id}.done"


# ---------------------------------------------------------------------------
# Template loading
# ---------------------------------------------------------------------------


def _load_unit_template(template_dir: str) -> str:
    """Load the firstrun.service.tpl template from the given directory.

    Args:
        template_dir: Path to the directory containing firstrun.service.tpl.

    Returns:
        The raw template string.

    Raises:
        FileNotFoundError: if firstrun.service.tpl is missing.
    """
    path = os.path.join(template_dir, "firstrun.service.tpl")
    with open(path) as fh:
        return fh.read()


# ---------------------------------------------------------------------------
# render_firstrun_unit
# ---------------------------------------------------------------------------


def render_firstrun_unit(
    opinion_id: str,
    description: str,
    exec_path: str,
    template_dir: str = None,
) -> str:
    """Render a systemd user oneshot unit for a first-run opinion.

    The unit is flag-file guarded: it only runs if the flag file does not
    exist, and creates the flag file on success — ensuring idempotency across
    multiple user logins without relying on ConditionFirstBoot (Pitfall 3).

    Foundation-neutral: the template directory is parameterized so any
    translator can provide a custom ``firstrun.service.tpl``. When
    ``template_dir`` is None, defaults to ``translators/common/templates/``.

    Args:
        opinion_id: The opinion ID, e.g. "OM-102" or "DF-001". Used to:
            - Name the flag file (/var/lib/debateos/.firstrun-OM-102.done)
            - Appear in the unit content
        description: Human-readable description of the first-run action.
            Appears in the unit's Description= field.
        exec_path: Absolute path to the executable to run at first login.
            Appears in ExecStart=.
        template_dir: Optional path to a directory containing
            ``firstrun.service.tpl``. Defaults to the common templates/
            directory when None.

    Returns:
        A complete systemd unit file contents as a string.
    """
    if template_dir is None:
        template_dir = _DEFAULT_TEMPLATES_DIR

    flag_file = _flag_file_path(opinion_id)
    template = _load_unit_template(template_dir)
    return template.format(
        description=description,
        flag_file=flag_file,
        exec_path=exec_path,
    )
