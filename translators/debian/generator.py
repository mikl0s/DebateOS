"""
generator.py — generate() entrypoint for the DebateOS Debian translator.

Composes the full translation pipeline:
  ResolvedSpeech JSON + opinion bodies
    → capability gate (DEB-01 / SC-3)
    → BuildManifest (common)
    → variant profile (debian.yaml)
    → live-build config/ tree

Usage as a module (argv-stable for Phase 3 CLI subprocess invocation):
  python3 -m translators.debian.generator <resolved.json> <opinions-path> <profile> <out-dir>

The translate shell wrapper calls this module with the frozen argv:
  translate <resolved.json> --opinions <path> --profile <name> --out <dir>

Design decisions:
- generate() is the single public entrypoint; it returns out_dir on success.
- Capability gate fires before any file I/O (fail fast, DEB-01 / SC-3).
- Source derivation is deterministic (BLD-03 groundwork via derive_source_date_epoch).
- No per-variant code branches (ARCH-04 invariant applies to Debian).
- Imports from translators/common/ (not arch copy-paste) — DEB-03 / COMM-01.
"""

import json
import os
import sys
from pathlib import Path

# Ensure translators/debian is on sys.path for bare-name imports regardless of
# how this module is invoked (pytest, python -m translators.debian.generator,
# or translate shell wrapper via PYTHONPATH).
_DEBIAN_DIR = os.path.dirname(os.path.abspath(__file__))
if _DEBIAN_DIR not in sys.path:
    sys.path.insert(0, _DEBIAN_DIR)

# Ensure translators/ (parent of debian/) is on sys.path so `import common`
# resolves when running from inside translators/debian/ context.
_TRANSLATORS_DIR = os.path.dirname(_DEBIAN_DIR)
if _TRANSLATORS_DIR not in sys.path:
    sys.path.insert(0, _TRANSLATORS_DIR)

from capabilities import load_capabilities, check_capabilities
from contract import load_opinion_bodies
from manifest import BuildManifest
from profile import emit_profile_tree
from variant import load_variant_profile


# ---------------------------------------------------------------------------
# Public entrypoint
# ---------------------------------------------------------------------------


def generate(
    resolved_path: str,
    opinions_path: str,
    profile_name: str = "debian",
    out_dir: str = "./debian-profile",
) -> str:
    """Generate a complete live-build config/ tree from a resolved speech.

    Steps:
    1. Load resolved speech + opinion bodies (contract).
    2. Load translator capabilities (Debian set — no mkinitcpio/limine).
    3. Run capability gate (check_capabilities) — raises CapabilityError BEFORE
       any file I/O if a required opinion needs an unsupported capability (SC-3).
    4. Build BuildManifest (foundation-neutral, from common).
    5. Load variant profile (profiles/debian.yaml by default).
    6. Emit the full live-build config/ tree.

    Args:
        resolved_path: Path to the ResolvedSpeech JSON file (Phase 1 output).
        opinions_path: Path to the opinion bodies source — a JSON array file
            or a directory of YAML/JSON opinion files.
        profile_name: Variant profile name (default "debian").
            Must match a YAML file in translators/debian/profiles/.
        out_dir: Output directory for the generated config/ tree
            (default "./debian-profile"). Created if it does not exist.

    Returns:
        The out_dir path string on success.

    Raises:
        CapabilityError: If a required opinion declares an unsupported capability
            (DEB-01 / SC-3 gate; fires before any file I/O).
        FileNotFoundError: If resolved_path, opinions_path, or the variant
            profile YAML does not exist.
        ValueError: If a file_asset dst path is absolute or traverses outside
            the target root (T-04-05 path sanitization).
    """
    # --- Step 1: Load resolved speech + opinion bodies ---
    # Read the resolved speech file only once — derive both the parsed dict and
    # the raw bytes from a single open() call.
    with open(resolved_path, "rb") as fh:
        resolved_bytes = fh.read()
    resolved = json.loads(resolved_bytes.decode("utf-8"))
    # Apply the same key-defaulting as contract.load_resolved_speech
    for _key in ("applied", "skipped", "dropped", "install_order", "explanations"):
        if _key not in resolved:
            resolved[_key] = []
    opinions_index = load_opinion_bodies(opinions_path)

    # --- Step 2: Load capabilities ---
    capabilities = load_capabilities()

    # --- Step 3: Run capability gate BEFORE any file I/O ---
    # SC-3 / DEB-01: check_capabilities raises CapabilityError immediately
    # if a required opinion needs an unsupported Debian capability.
    # The gate fires here (not inside BuildManifest.from_resolved) because
    # foundation-neutral manifest.py does not know which capabilities apply.
    check_capabilities(resolved, opinions_index, capabilities)

    # --- Step 4: Build BuildManifest ---
    manifest = BuildManifest.from_resolved(
        resolved=resolved,
        opinions_index=opinions_index,
        capabilities=capabilities,
        resolved_bytes=resolved_bytes,
    )

    # --- Step 5: Load variant profile ---
    variant = load_variant_profile(profile_name)

    # --- Step 6: Emit live-build config/ tree ---
    emit_profile_tree(
        out_dir=out_dir,
        manifest=manifest,
        variant=variant,
    )

    return out_dir


# ---------------------------------------------------------------------------
# __main__ — runnable as `python3 -m translators.debian.generator` (Phase 3 argv)
# ---------------------------------------------------------------------------


def _main(argv=None):
    """CLI entry for direct module invocation.

    Usage: python3 -m translators.debian.generator <resolved.json> <opinions-path> <profile> <out>

    Args are positional in this order to match the translate shell wrapper.
    """
    if argv is None:
        argv = sys.argv[1:]

    if len(argv) < 4 or argv[0] in ("-h", "--help"):
        print(
            "usage: python3 -m translators.debian.generator "
            "<resolved.json> <opinions-path> <profile> <out-dir>",
            file=sys.stderr,
        )
        sys.exit(1)

    resolved_path, opinions_path, profile_name, out_dir = argv[:4]

    try:
        result = generate(
            resolved_path=resolved_path,
            opinions_path=opinions_path,
            profile_name=profile_name,
            out_dir=out_dir,
        )
        print(f"Profile tree generated at: {result}")
    except Exception as exc:  # noqa: BLE001
        print(f"ERROR: {exc}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    _main()
