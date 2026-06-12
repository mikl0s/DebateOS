"""
firstrun.py — First-run systemd user oneshot unit generator for the Arch translator.

Provides:
- render_firstrun_unit(opinion_id, description, exec_path) → str

Generated units follow 02-RESEARCH.md Pattern 2:
- User-scope systemd oneshot service (not system scope)
- ConditionPathExists=!/var/lib/debateos/.firstrun-<id>.done (flag-file guard, Pitfall 3)
- ExecStartPost=/bin/touch <flag-file> (idempotency)
- RemainAfterExit=yes
- WantedBy=graphical-session.target (requires live user session)
- After=graphical-session.target (ordering)

Security: T-02-11 — units are placed in etc/systemd/USER/ (not system/), ensuring
they run in the target user context, not root scope.
"""

# ---------------------------------------------------------------------------
# Unit file name helper
# ---------------------------------------------------------------------------

FLAG_FILE_DIR = "/var/lib/debateos"
FLAG_FILE_PREFIX = ".firstrun-"


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
# render_firstrun_unit
# ---------------------------------------------------------------------------

_UNIT_TEMPLATE = """\
[Unit]
Description=DebateOS first-run: {description}
ConditionPathExists=!{flag_file}
After=graphical-session.target

[Service]
Type=oneshot
ExecStart={exec_path}
ExecStartPost=/bin/touch {flag_file}
RemainAfterExit=yes

[Install]
WantedBy=graphical-session.target
"""


def render_firstrun_unit(
    opinion_id: str,
    description: str,
    exec_path: str,
) -> str:
    """Render a systemd user oneshot unit for a first-run opinion.

    The unit is flag-file guarded: it only runs if the flag file does not
    exist, and creates the flag file on success — ensuring idempotency across
    multiple user logins without relying on ConditionFirstBoot (Pitfall 3).

    Args:
        opinion_id: The opinion ID, e.g. "OM-102". Used to:
            - Name the flag file (/var/lib/debateos/.firstrun-OM-102.done)
            - Name the service unit (debateos-firstrun-OM-102.service)
        description: Human-readable description of the first-run action.
            Appears in the unit's Description= field.
        exec_path: Absolute path to the executable to run at first login.
            Appears in ExecStart=.

    Returns:
        A complete systemd unit file contents as a string, ready to be
        written to airootfs/etc/systemd/user/debateos-firstrun-<id>.service.
    """
    flag_file = _flag_file_path(opinion_id)
    return _UNIT_TEMPLATE.format(
        description=description,
        flag_file=flag_file,
        exec_path=exec_path,
    )
