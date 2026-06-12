# 08 — Omarchy Research Brief (Phase 0)

**This research gates the schema and therefore everything else. Do it first; do not invent the schema before it is done.**

## Why Omarchy

Omarchy (DHH / 37signals) is an opinionated, pre-configured Arch + Hyprland setup aimed at developers — exactly the "opinions on top of a base OS" bundle that DebateOS generalizes. It is the ideal reverse-engineering target because it is:

- **Openly developed** — install scripts, package manifests, and configs are inspectable.
- **Aggressively opinionated** — deliberate package choices and removals, theming, keybindings, and hardware-specific touches (e.g. Apple-monitor bindings).
- **Self-installing on top of Arch** — it ships a de-facto single-OS "translator" that is itself worth studying.

**Source:** `https://github.com/basecamp/omarchy` (clone and analyze the actual source; do not rely on blog summaries).

## Research goal

Derive the **baseline opinion schema from real-world data, not theory.** The output must answer: *what must an opinion, at minimum, be able to express?*

## Method

1. Obtain the Omarchy source: install scripts, package lists, config files, dotfiles, theming, post-install hooks.
2. Walk **every decision made *after* a base Arch install** and record it as a candidate atomic opinion.
3. For each candidate opinion, record:
   - **Category** — package install/removal, config/dotfile deployment, service enablement, kernel/boot parameter, theming, keybinding, hardware-conditional, arbitrary script, …
   - **What's needed to express it OS-agnostically** (the intent, stripped of Arch mechanics).
   - **Dependencies and ordering** relative to other opinions.
   - **Anything that *cannot* be made OS-agnostic** → record as a **translator capability requirement**.
4. Group opinions into natural candidate **points** (e.g. "Hyprland desktop", "developer toolchain", "Omarchy theming", "Apple display support").
5. Note **surprises that expand the schema** (e.g. if boot steps are reordered → ordering metadata is mandatory; if custom scripts ship → opinions must support arbitrary script payloads; if theming ships asset files → opinions must carry file payloads).

## Deliverables (write under `research/`)

1. `research/omarchy-opinion-inventory.md` — full structured list of extracted opinions with categories and metadata observations.
2. `research/omarchy-points.md` — proposed point groupings.
3. `research/schema-requirements.md` — the minimum expressive surface an opinion/point/speech schema must support, each requirement backed by evidence from the inventory.
4. `research/open-questions.md` — surprises, ambiguities, and anything that can't be made OS-agnostic.

## Constraint & validation target

This research **gates** schema design (Phase 1). The schema is then extended through design work, but its **floor comes from this analysis.** The end-state validation (Phase 2 milestone) is the inverse operation:

> **North star: Omarchy must be reproducible as a speech on vanilla Arch.**

If DebateOS can express Omarchy as points + opinions on vanilla Arch and produce an equivalent installed system, the model works. Keep that target in mind while doing the research — every Omarchy decision you catalog is something the schema must eventually be able to express and the Arch translator must eventually be able to build.
