# DebateOS Example Compositions

Evidence-derived example compositions demonstrating the DebateOS schema and
resolver end-to-end (parse → resolve → explain). All content is CC0-1.0
(see `examples/LICENSE`).

Each example consists of two files:
- **`speech.yaml`** — the composition: which opinions the user selects.
- **`opinions.yaml`** — the opinion definitions included in this composition.

These examples are harness fixtures exercised by `examples/examples_test.go`.
They also generate the committed golden files in
`resolver/resolve/testdata/golden/` used by the WASM parity test.

---

## omarchy-mini

**What it shows:** A clean, non-conflicting subset of real Omarchy opinions
from the Hyprland desktop stack.

**Opinions included:**
- **OM-001** (custom-repo / required): Register the Omarchy package repository.
- **OM-006** (package-install / required): Wayland compositor stack (Hyprland,
  UWSM, hypridle, hyprlock). Depends on OM-001.
- **OM-007** (package-install / required): Terminal and shell tools (foot, tmux,
  starship, zoxide, bat, eza, fd, ripgrep, fzf). Depends on OM-001.
- **OM-015** (package-install / required): Desktop shell components (waybar,
  mako, swaybg, sddm, polkit-gnome). Depends on OM-006.

**Resolution outcome:** All four opinions applied; topological install order
OM-001 → OM-006 → OM-007 → OM-015 (OM-007 is a peer of OM-006, both after
OM-001, before OM-015).

---

## two-point-clean

**What it shows:** Two non-conflicting opinions that resolve cleanly with no
dropped or skipped opinions.

**Opinions included:**
- **OM-007** (package-install / required): Terminal and shell tools.
- **OM-064** (service-enable / nice-to-have): Bluetooth service and configuration.

**Resolution outcome:** Both opinions applied; no rule1/2/3/4 conflict firings.

---

## conflicting

**What it shows:** Two required opinions occupying the same exclusive slot
(display manager) produce a **hard conflict** (Rule 2). No patch opinion is
available.

**Opinions included:**
- **OM-015** (package-install / required): Desktop shell stack with **SDDM**
  as the display manager.
- **OM-015-greetd** (package-install / required): Alternative desktop shell
  stack using **greetd + tuigreet** instead of SDDM. Both opinions declare
  mutual `conflicts:` entries.

**Resolution outcome:** `Resolve` returns a non-nil error with
`"Hard conflict: OM-015 (required) and OM-015-greetd (required) cannot coexist."`.
The partial `ResolvedSpeech` carries the conflict `Explanation` so the caller
can display it (Invariant 3).

---

## hardware-conditional

**What it shows:** A hardware-gated opinion (OM-068 NVIDIA Driver Stack) that
is Applied when the declared hardware profile includes a matching NVIDIA PCI ID
and Skipped otherwise.

**Opinions included:**
- **OM-006** (package-install / required): Wayland compositor stack. Always applied.
- **OM-068** (hardware-conditional / nice-to-have): NVIDIA open DKMS driver
  stack, gated on PCI-ID set membership (10de:2204 / 10de:2206 / 10de:2208 /
  10de:1f04 / 10de:1f08 — Turing+ and Ampere GPUs).

**Resolution outcome (matching):** OM-068 Applied with `Rule="hardware-apply"`;
explanation confirms the hardware condition matched.

**Resolution outcome (non-matching):** OM-068 Skipped with `Rule="hardware-skip"`;
explanation confirms `"hardware condition not met for declared hardware profile"`.

---

## Running the Tests

```bash
go test ./examples/ -v
```

## Regenerating Golden Files

The committed golden files in `resolver/resolve/testdata/golden/` are the
byte-identical parity baseline for `scripts/wasm-parity-test.sh`. To regenerate:

```bash
GOLDEN_UPDATE=1 go test ./resolver/resolve/ -run TestCanonicalGolden -count=1
```

Then commit the updated `*.json` files.
