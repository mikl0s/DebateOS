# Omarchy North-Star Composition

This directory contains the DebateOS north-star composition for
[Omarchy](https://github.com/basecamp/omarchy), pinned at commit
**9cf1852525a5f7de26d3162db9d61e2f5c1d5523** (version 4.0.0.alpha).

It is the canonical example proving that Omarchy's full opinionated Linux
desktop can be expressed as a schema-valid, conflict-free DebateOS speech.

## License

All content in `examples/omarchy/` is released under CC0 1.0 Universal.
See `LICENSE` for the full text.

## Structure

```
opinions/      134 opinion YAML files (OM-001.yaml .. OM-134.yaml)
points/        32 point YAML files (one per curated bundle)
speech.yaml    The north-star speech targeting vanilla Arch Linux
gen/           generate.py — idempotent authoring helper (CC0)
```

## Status Policy (RESOLVED OQ-1)

Status assignments follow the RESOLVED open question OQ-1:

- **required**: OM-001 (custom repo), OM-006 (compositor), OM-097
  (display manager), OM-099 (bootloader), and all hardware-conditional
  opinions (their `hardware_condition` gates them to `Skipped` on a
  generic machine — not a hard conflict).
- **nice-to-have**: visual themes OM-114..OM-134, and a small set of
  optional extras (npm tools, branding, first-run UX opinions).
- **all others**: required, reflecting their role as core Omarchy
  infrastructure.

## Resolution Behaviour

On a vanilla Arch Linux baseline (no special hardware declared):

- Hardware-conditional opinions (OM-024..027, OM-058..094, OM-106) land
  in **Skipped** — correct and expected; their `hardware_condition` simply
  evaluated false.
- All non-gated opinions land in **Applied**.
- No **Hard conflict** is present — the speech resolves cleanly.

## Capability Tokens

Every opinion carries `translator_capabilities` tokens that map it to what
the Arch translator (Phase 2 Plan 01, `translators/arch/`) must be able to
do. This is the ARCH-03 capability map: the union of tokens across all
opinions defines the required translator capability surface.

## Generator

`gen/generate.py` is the idempotent authoring helper. Re-running it from
the repo root produces identical output:

```bash
python examples/omarchy/gen/generate.py
```

The script reads `research/omarchy-opinion-inventory.md` and
`research/omarchy-points.md` and emits the same 134 + 32 YAML files.
It requires PyYAML (stdlib + PyYAML only; no other dependencies).

## Pinned Commit

All evidence traces to Omarchy commit `9cf1852525a5f7de26d3162db9d61e2f5c1d5523`.
To reproduce the source analysis:

```bash
git clone https://github.com/basecamp/omarchy /tmp/omarchy
git -C /tmp/omarchy checkout 9cf1852525a5f7de26d3162db9d61e2f5c1d5523
```
