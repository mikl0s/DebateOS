"""
capabilities.py — Arch translator capability gate (ARCH-03 / SC-3).

Declares the supported translator capability tokens and enforces the gate:
- Required opinions with unsupported capabilities fail loudly at composition time
  (CapabilityError naming opinion ID + missing token + "composition time").
- Nice-to-have opinions with unsupported capabilities are dropped with a recorded
  reason (never silently).

Usage::

    from capabilities import load_capabilities, check_capabilities, CapabilityError

    caps = load_capabilities()
    dropped = check_capabilities(resolved_speech, opinions_index, caps)
    # dropped: list of (opinion_id, reason_str) for unsupported nice-to-haves.
"""

import json
import os


# ---------------------------------------------------------------------------
# CapabilityError
# ---------------------------------------------------------------------------


class CapabilityError(Exception):
    """Raised at composition time when a REQUIRED opinion declares a capability
    that the Arch translator has not implemented.

    The message always names the opinion ID, the missing capability token, and
    the phrase "composition time" so callers can surface it clearly (SC-3).
    """


# ---------------------------------------------------------------------------
# load_capabilities
# ---------------------------------------------------------------------------

# Path to capabilities.json relative to this module file.
_CAPABILITIES_JSON = os.path.join(os.path.dirname(__file__), "capabilities.json")


def load_capabilities(path: str = _CAPABILITIES_JSON) -> set:
    """Load the declared capability tokens from ``capabilities.json``.

    Returns:
        A ``set[str]`` of all declared capability tokens.

    Raises:
        FileNotFoundError: if the capabilities.json file is missing.
        KeyError: if the JSON does not have a "capabilities" key.
    """
    with open(path) as fh:
        data = json.load(fh)
    return set(data["capabilities"])


# ---------------------------------------------------------------------------
# check_capabilities  (SC-3 gate)
# ---------------------------------------------------------------------------


def check_capabilities(resolved: dict, opinions_index: dict, capabilities: set) -> list:
    """Enforce the capability gate (ARCH-03 / SC-3).

    Iterates over every opinion in ``resolved["applied"]`` and compares its
    ``translator_capabilities`` list against the declared ``capabilities`` set.

    - REQUIRED opinion with an unsupported capability → raises :class:`CapabilityError`
      immediately.  The error message names the opinion ID, the missing capability
      token, and contains the phrase ``"composition time"``.
    - NICE-TO-HAVE opinion with an unsupported capability → appended to the
      ``dropped`` result list as ``(opinion_id, reason_str)`` where ``reason_str``
      names every unsupported capability for that opinion.

    Args:
        resolved: The loaded ResolvedSpeech dict (keys: applied, skipped, dropped,
            install_order, explanations).
        opinions_index: Dict mapping opinion ID → opinion dict.
        capabilities: Set of declared capability token strings.

    Returns:
        A list of ``(opinion_id: str, reason: str)`` tuples for nice-to-have opinions
        that were silently dropped because they require unsupported capabilities.
        Empty list when all applied opinions' capabilities are satisfied.

    Raises:
        CapabilityError: When a required opinion declares a capability not present
            in ``capabilities``.
    """
    dropped = []

    for opinion_id in resolved.get("applied", []):
        opinion = opinions_index.get(opinion_id)
        if opinion is None:
            # Opinion missing from index — treat as empty capability set, pass through.
            continue

        status = opinion.get("status", "nice-to-have")
        caps_needed = opinion.get("translator_capabilities", [])

        unsupported = [c for c in caps_needed if c not in capabilities]
        if not unsupported:
            continue

        if status == "required":
            # Fail loudly: name the opinion, the first unsupported capability, and
            # use the phrase "composition time" so the message is recognizable.
            raise CapabilityError(
                f"Opinion {opinion_id} requires capability '{unsupported[0]}' "
                f"which is not declared by the Arch translator "
                f"at composition time. "
                f"Add it to translators/arch/capabilities.json when implemented. "
                f"(All unsupported: {unsupported})"
            )
        else:
            # Nice-to-have: drop with explanation, never raise.
            reason = f"unsupported capabilities: {unsupported}"
            dropped.append((opinion_id, reason))

    return dropped
