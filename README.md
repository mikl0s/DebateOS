# DebateOS

> *"That's just your opinion, man."*

A declarative configuration abstraction layer that decouples the base Linux distribution from personalization and stack choices — letting people compose, share, and remix their entire OS setup as layered **opinions**, without anyone maintaining a distro fork.

Atomic choices are **Opinions**. Curated bundles are **Points**. Your personal remix is a **Speech**. The visual conflict-resolution process is the **Debate**. OS-specific **Translators** turn an abstract speech into a fully unattended installer for Arch, Debian, or any foundation. The base OS becomes irrelevant; the opinions are what matter.

## Status

Pre-implementation. The complete founding context lives in [`docs/`](docs/) and is the source of truth for building **v1.0 (Phases 0–5)**. Start with [`docs/00-START-HERE.md`](docs/00-START-HERE.md).

## Documentation map

| File | Contents |
|---|---|
| [docs/00-START-HERE.md](docs/00-START-HERE.md) | Kickoff context, build order, operating mandate |
| [docs/01-vision.md](docs/01-vision.md) | Mission, philosophy, the problem |
| [docs/02-concepts.md](docs/02-concepts.md) | Opinion · Point · Speech · Debate · Translator · Foundation · Registry |
| [docs/03-architecture.md](docs/03-architecture.md) | Components, data flow, technical decisions |
| [docs/04-conflict-resolution.md](docs/04-conflict-resolution.md) | Metadata model, resolution rules, patches |
| [docs/05-distribution-and-infra.md](docs/05-distribution-and-infra.md) | Zero-cost build/registry, The Forum, secrets model |
| [docs/06-social-layer.md](docs/06-social-layer.md) | Curators, discovery, reputation |
| [docs/07-roadmap.md](docs/07-roadmap.md) | Phases 0–5 with goals and success criteria |
| [docs/08-omarchy-research.md](docs/08-omarchy-research.md) | Phase 0 research brief (gates the schema) |
| [docs/09-decisions.md](docs/09-decisions.md) | **Locked decisions + invariants** |
| [docs/10-prior-art-and-lessons.md](docs/10-prior-art-and-lessons.md) | Reuse / drop / avoid from the prior attempt |
| [docs/11-repo-layout.md](docs/11-repo-layout.md) | Monorepo directory tree |

## Tech at a glance

Go resolver (native + WASM) · Go CLI · shell/Python translators (Arch, Debian) · SvelteKit static Debate UI running the WASM resolver · Docker + GitHub Actions builds · GitHub-Pages registry · optional lean Go discovery service with embedded SQLite ("The Forum"). License: AGPLv3 (code) + CC0 (schemas/content). Zero-cost, non-commercial, released into the wild.
