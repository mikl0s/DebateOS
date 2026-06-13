"""
generator.py — generate() entrypoint for the DebateOS Arch translator.

Composes the full translation pipeline:
  ResolvedSpeech JSON + opinion bodies
    → capability gate (ARCH-03)
    → BuildManifest (plan 01)
    → variant profile (plan 03)
    → archiso profile tree (plan 02)

Usage as a module (argv-stable for Phase 3 CLI subprocess invocation):
  python3 -m translators.arch.generator <resolved.json> <opinions-path> <profile> <out-dir>

The translate shell wrapper calls this module with the frozen argv:
  translate <resolved.json> --opinions <path> --profile <name> --out <dir>

Design decisions:
- generate() is the single public entrypoint; it returns out_dir on success.
- Capability gate fires before any file I/O (fast fail, ARCH-03 / SC-3).
- Source derivation is deterministic (BLD-03 groundwork via derive_source_date_epoch).
- No per-variant code branches (ARCH-04 invariant).
"""

import json
import os
import sys
from pathlib import Path

# Ensure translators/arch is on sys.path for bare-name imports regardless of
# how this module is invoked (pytest, python -m translators.arch.generator,
# or translate shell wrapper via PYTHONPATH).
_ARCH_DIR = os.path.dirname(os.path.abspath(__file__))
if _ARCH_DIR not in sys.path:
    sys.path.insert(0, _ARCH_DIR)

# Ensure translators/ (parent of arch/) is on sys.path so `import common`
# resolves when running from inside translators/arch/ context.
_TRANSLATORS_DIR = os.path.dirname(_ARCH_DIR)
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
    profile_name: str = "vanilla-arch",
    out_dir: str = "./arch-profile",
) -> str:
    """Generate a complete archiso profile tree from a resolved speech.

    Steps:
    1. Load resolved speech + opinion bodies (contract).
    2. Load translator capabilities.
    3. Build BuildManifest (includes capability gate — raises CapabilityError
       before any file I/O if a required opinion needs an unsupported capability).
    4. Load variant profile.
    5. Emit the full archiso profile tree.

    Args:
        resolved_path: Path to the ResolvedSpeech JSON file (Phase 1 output).
        opinions_path: Path to the opinion bodies source — a JSON array file
            or a directory of YAML/JSON opinion files.
        profile_name: Variant profile name (default "vanilla-arch").
            Must match a YAML file in translators/arch/profiles/.
        out_dir: Output directory for the generated archiso profile tree
            (default "./arch-profile"). Created if it does not exist.

    Returns:
        The out_dir path string on success.

    Raises:
        CapabilityError: If a required opinion declares an unsupported capability
            (ARCH-03 / SC-3 gate; fires before any file I/O).
        FileNotFoundError: If resolved_path, opinions_path, or the variant
            profile YAML does not exist.
        ValueError: If a file_asset dst path is absolute or traverses outside
            the target root (T-02-08 path sanitization).
    """
    # --- Step 1: Load resolved speech + opinion bodies ---
    # IN-02: Read the resolved speech file only once — derive both the parsed
    # dict and the raw bytes from a single open() call to avoid inconsistency
    # between the two reads if the file were modified between them.
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

    # --- Step 3: Build BuildManifest ---
    # SC-3 / ARCH-03: Run the capability gate BEFORE manifest assembly.
    # The shared common/manifest.py is foundation-neutral and does not call
    # check_capabilities internally — the translator (generator.py) owns the
    # gate call so each foundation's capability set applies correctly.
    check_capabilities(resolved, opinions_index, capabilities)
    manifest = BuildManifest.from_resolved(
        resolved=resolved,
        opinions_index=opinions_index,
        capabilities=capabilities,
        resolved_bytes=resolved_bytes,
    )

    # --- Step 4: Load variant profile ---
    variant = load_variant_profile(profile_name)

    # --- Step 5: Emit profile tree ---
    emit_profile_tree(
        out_dir=out_dir,
        manifest=manifest,
        variant=variant,
    )

    return out_dir


# ---------------------------------------------------------------------------
# __main__ — runnable as `python3 -m translators.arch.generator` (Phase 3 argv)
# ---------------------------------------------------------------------------


def _main(argv=None):
    """CLI entry for direct module invocation.

    Usage: python3 -m translators.arch.generator <resolved.json> <opinions-path> <profile> <out>

    Args are positional in this order to match the translate shell wrapper.
    """
    if argv is None:
        argv = sys.argv[1:]

    if len(argv) < 4 or argv[0] in ("-h", "--help"):
        print(
            "usage: python3 -m translators.arch.generator "
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
