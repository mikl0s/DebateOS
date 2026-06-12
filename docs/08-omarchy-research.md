# 08 — Omarchy Research Brief (Phase 0)

**This research gates the schema and therefore everything else. Do it first; do not invent the schema before it is done.**

> **Scope note (2026-06-12):** Phase 0 covers Omarchy (primary, deep) **plus** an Arch-variant substitution study of **CachyOS** and **Garuda Linux** (secondary, targeted — see "Arch-variant substitution study" below). The variants validate that a foundation can be swapped for another of the same base, feed the translator's package-repo handling, and seed the resolver's edge-case corpus.

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

## Arch-variant substitution study (CachyOS, Garuda Linux)

Omarchy answers *"what must an opinion express?"* The variant study answers the complementary question: *"what must a **foundation swap within the same base** preserve, and where does it leak?"*

**Targets:** CachyOS and Garuda Linux — both Arch-based, both opinionated in ways vanilla Arch is not, and opinionated *differently from each other*:

- **CachyOS** — performance-focused: custom repos with `x86-64-v3`/`v4` optimized packages, custom kernels (cachyos-bore etc.), its own mirror/repo priority scheme layered above the standard Arch repos.
- **Garuda** — gaming/aesthetics-focused: Chaotic-AUR as a first-class repo, btrfs + snapper snapshots by default, heavy theming, its own tooling (garuda-update wrapping pacman).

**Validation question:** *can a speech targeting "Arch" be retargeted to an Arch variant unchanged?* Concretely: would Omarchy-as-a-speech build and install correctly on CachyOS or Garuda, and what breaks if not?

**Method (targeted, not a second deep inventory):**

1. For each variant, catalog the **delta from vanilla Arch**: extra/replacement package repos, repo priority and mirror handling, keyrings and signing, kernel variants, default filesystem/bootloader choices, pre-seeded configs and tooling that overlaps with what opinions would manage.
2. Determine what the **translator capability surface** needs to absorb these deltas *without bloating*: ideally the Arch translator gains a small, declarative "variant profile" (repo list + keyring + kernel + defaults), not a fork per variant. Record what fits that shape and what does not.
3. Record every case where the **same opinion would effectuate differently** across vanilla Arch / CachyOS / Garuda (e.g. a package exists in the variant repo with different optimization flags; a kernel opinion collides with the variant's default kernel; theming opinions collide with Garuda's pre-seeded theming). Each such case is BOTH a system-support requirement AND a candidate resolver test scenario.
4. Note which variant behaviors are really **pre-installed opinions in disguise** (Garuda's theming, CachyOS's kernel choice) — these inform how the schema models "the foundation already has an opinion about this".

**Deliverable framing:** the variant study output feeds three consumers — (a) the schema floor (foundation/variant metadata an opinion may need), (b) the Phase 2 Arch translator design (variant profiles, multi-repo handling), and (c) the Phase 1 resolver conflict-test harness (real-world edge cases, e.g. foundation-default vs user-opinion collisions).

## Deliverables (write under `research/`)

1. `research/omarchy-opinion-inventory.md` — full structured list of extracted opinions with categories and metadata observations.
2. `research/omarchy-points.md` — proposed point groupings.
3. `research/schema-requirements.md` — the minimum expressive surface an opinion/point/speech schema must support, each requirement backed by evidence from the inventory **and the variant study**.
4. `research/open-questions.md` — surprises, ambiguities, and anything that can't be made OS-agnostic.
5. `research/arch-variants-delta.md` — CachyOS + Garuda deltas from vanilla Arch: repos, keyrings, kernels, defaults, pre-seeded opinions; with a proposed variant-profile shape for the Arch translator.
6. `research/resolver-edge-cases.md` — concrete conflict/edge-case scenarios harvested from Omarchy and the variant study (foundation-default vs opinion, repo-priority collisions, kernel-variant conflicts, theming collisions), each written as a candidate Phase 1 resolver test scenario.

## Constraint & validation target

This research **gates** schema design (Phase 1). The schema is then extended through design work, but its **floor comes from this analysis.** The end-state validation (Phase 2 milestone) is the inverse operation:

> **North star: Omarchy must be reproducible as a speech on vanilla Arch.**

If DebateOS can express Omarchy as points + opinions on vanilla Arch and produce an equivalent installed system, the model works. Keep that target in mind while doing the research — every Omarchy decision you catalog is something the schema must eventually be able to express and the Arch translator must eventually be able to build.

**Stretch validation (same-base substitution):** the same Omarchy speech, retargeted to an Arch variant (CachyOS or Garuda) via a variant profile, should build with only declared, explainable differences — proving foundations of the same base are interchangeable. This is a Phase 2 stretch target, not a gate; the Phase 0 deliverable is knowing *exactly what it would take*.
