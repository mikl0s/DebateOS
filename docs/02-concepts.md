# 02 — Core Concepts & Terminology

These terms are load-bearing. Use them consistently across code, schemas, docs, and UI. The rhetoric metaphor is intentional.

## Opinion

The **atomic unit**: a single discrete, OS-agnostic configuration decision applied on top of a base OS install.

Examples: "Install NVIDIA proprietary drivers", "Use Hyprland as the compositor", "Remove the default office suite", "Set kernel parameter X", "Deploy this dotfile to `~/.config/...`".

Properties:
- Expresses **intent**, never distro mechanics. Mechanics belong to translators.
- Carries **metadata**: dependencies, conflicts, hardware conditions, ordering constraints, known patches, required-vs-nice-to-have status, translator capability requirements (see `04-conflict-resolution.md`).
- May encompass: package install/removal, script payloads, config/dotfile deployment, service enablement, kernel/boot parameters, hardware-conditional logic, theming, keybindings.

## Point

A **curated, coherent bundle of opinions**, maintained by a person or organization.

Examples: "Linus Tech Tips gaming setup", "NVIDIA official local-AI stack", "Sarah's Hyprland rice", "Headless server console".

Properties:
- Each opinion in a point is marked **required** or **nice-to-have**.
- May declare ordering constraints relative to other points.
- Versioned, forkable, subscribable.
- **Foundation-agnostic** — a point does not know or care which base OS it targets.

## Speech

A user's **complete composition**: the full manifest of selected points, individually selected opinions, and private customization that generates the finished installed system.

Properties:
- Expressed as **data** (YAML) — versionable, diffable, shareable.
- Composed of **public panes** (community points the user subscribed to) and **one private pane** (personal layer: SSH keys, dotfiles, `.bashrc`, home-folder structure, environment variables) that **never leaves the user's machine**.
- Compiles into a fully-unattended installer for the chosen foundation.

## Debate

The **composition and conflict-resolution process**, primarily experienced through a visual interface.

Mental model: a **foundation** (base OS + installer + bootloader) with transparent glass **panes** (points) stacked on top. The system highlights conflicts visually (red overlaps = incompatible, green = compatible), shows dependency ordering, and offers resolutions (auto-apply required-wins rules, suggest patch opinions, or let the user choose). When everything aligns, the debate **compresses into the final speech**.

Rhetorical principle: *there are no conclusions, only debates.*

## Translator

The **OS-specific layer** that effectuates abstract opinions into concrete, foundation-specific instructions. One per supported foundation (Arch translator, Debian translator, …).

Properties:
- Knows the mechanics: package manager, config file locations, service enablement, how to build an unattended installer image.
- **Declares which opinions/capabilities it supports.** Unsupported opinions break **visibly at composition time**, never silently at install time.
- Ownership model: the project bootstraps Arch (+ 1–2 Arch variants) and Debian; distributions are invited to maintain their own translators (Ubuntu owns Ubuntu's), and community PRs are welcome.

## Foundation

The **base OS + bootloader + installer** choice. Deliberately boring, interchangeable infrastructure. In the final form even the installer is itself an expression of opinions: a custom unattended installer generated per-speech.

## Registry

The public, **Git-backed index** of points and public speeches. Discovery, versioning, forking, and reputation. Built on GitHub repositories + a static index on GitHub Pages (GitLab parity desired but not required for v1.0). The optional **Forum** discovery service (see `05`/`06`) provides richer search, reputation, and collaborative conflict resolution on top of this Git-backed data.

## Patch Opinion

A first-class, community-contributed opinion whose purpose is to make two otherwise-conflicting opinions coexist (e.g. a compatibility shim letting two stacks share a dependency at different versions). Discoverable (attached to the conflict pair in metadata), versioned, maintained, attributable — not a hack. Detailed in `04-conflict-resolution.md`.
