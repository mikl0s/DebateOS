# 01 — Vision

## The idea in one paragraph

An "opinionated distro" like Omarchy is really just vanilla Arch plus a set of post-install decisions. **DebateOS makes those decisions first-class, composable objects.** Atomic choices are **Opinions**. Curated bundles of opinions are **Points**. A user's personal remix of points plus private customization is a **Speech**. The visual conflict-resolution process — where panes are layered and clashes resolved — is the **Debate**. OS-specific **Translators** turn the abstract speech into a fully unattended installer for Arch, Debian, Fedora, or whatever foundation is chosen. The base OS becomes irrelevant; the opinions are what matter.

## The problem

Linux personalization is trapped inside distributions. To get someone else's setup you fork their distro, copy dotfiles by hand, or follow a wiki page that rots. "Opinionated distros" multiply endlessly, each one a maintenance burden and a tribe to defend. The decisions that actually matter — packages, compositor, theming, keybindings, drivers — are buried inside install scripts and forks instead of being shareable, diffable, remixable artifacts.

This produces two harms:

1. **Maintenance waste.** Every opinionated distro re-solves the same base-OS problems and carries a fork forever.
2. **Cultural fragmentation ("distro wars").** Identity attaches to the *foundation* rather than to the *opinions*, so communities splinter along tribal lines and newcomers are told to "pick a side." This actively harms Linux adoption and cohesion.

## The mission

Decouple the act of running Linux from the choice of distribution. Shift maintainers from *software stewards* to *configuration gardeners*. Make the foundation interchangeable infrastructure so that gaming, AI, and development communities can collaborate on the **same opinions** regardless of which base OS anyone runs.

When the foundation is irrelevant, there is no tribal axis left to fight over. Newcomers enter through *"compose your own, learn why opinions matter, remix as you grow"* instead of *"pick a side and defend it."*

## Philosophy

- **Opinions over distributions.** People should collaborate on decisions, not maintain forks.
- **Intent over mechanics.** An opinion expresses *what* you want ("use Hyprland", "install NVIDIA proprietary drivers"); translators own *how* each distro achieves it.
- **Human-readable always.** The visual Debate is the eventual primary interface, but the YAML underneath must remain comprehensible on its own. The system must never decay into a byzantine dependency resolver.
- **Decentralized by default.** The core compose→build path runs on free public tooling and user-owned compute, with no central bottleneck. Social features are an optional enhancement, never a dependency.
- **Anti-dogmatic tone.** There are no conclusions, only debates. Tagline: *"That's just your opinion, man."* Playful, tongue-in-cheek rhetoric metaphor throughout — softened where it would otherwise obscure meaning, but it is core to the identity.

## What success looks like

- A user composes a speech from several curators' points, resolves conflicts visually, and produces a bootable, fully-unattended installer for the foundation of their choice — at zero hosting cost.
- **Omarchy is reproducible as a speech on vanilla Arch** (the concrete validation north star).
- Two foundations (Arch + Debian) prove the opinion/translator abstraction is real, not an Arch-shaped illusion.
- Curators accumulate subscribers; popular speeches become de-facto "distributions" maintained purely as configuration.

## Non-goals (v1.0)

- **Monetization.** The project is non-commercial and released into the wild. No paid tiers, no central SaaS dependency.
- **Post-install reconciliation.** Applying speech changes to an already-installed running system is out of scope for v1.0 (install-time only).
- **Hardware-scanning installer.** Deferred to Phase 6 (post-v1.0); v1.0 targets ISO/USB installer output.
