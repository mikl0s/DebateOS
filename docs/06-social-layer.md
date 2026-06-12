# 06 — Social Layer

The social layer turns DebateOS from a build tool into an ecosystem. In v1.0 it is delivered by two cooperating pieces: the **Git-backed registry** (decentralized source of truth) and **The Forum** (an optional, owner-hosted discovery service that makes the registry searchable and social). The registry is authoritative; The Forum is an index and a meeting place.

## Curators as distribution channels

DebateOS flips the distro power dynamic. Instead of *"pick a distro and hope its maintainers care about what you care about,"* it becomes **"follow the creators and tinkerers you actually trust."**

Illustrative scenario:
- A hardware reviewer publishes/maintains a gaming point set (~10 points).
- A trusted AI researcher maintains a local-inference point set.
- A rice enthusiast maintains a desktop-aesthetics point set.
- A user subscribes to all three; the system merges them into one coherent speech, resolves overlaps/conflicts, and generates a single installer.

Users don't take everything from one curator: *"I love his Claude Code setup point — I'll subscribe to just that one."*

## Reputation & discovery

- Speeches and points are shareable, forkable artifacts that build curator reputation.
- Good ones accumulate subscribers; weak ones are ignored or improved via forks.
- Popular speeches become de-facto "distributions" — maintained purely as configuration, with zero software-maintenance burden.
- The Forum surfaces popularity, freshness, and compatibility information; the underlying data lives in Git.

## What The Forum provides (v1.0)

- **Search & discovery** across indexed points and speeches (by curator, tag, popularity, freshness, foundation compatibility).
- **Subscriptions** — follow curators; see their updates.
- **Ratings / reputation** — lightweight, GitHub-identity-backed.
- **Collaborative conflict resolution** — hosts the conflict threads from the `04` workflow: a registry of known conflicts, the disposable-environment repros, discussion, and links to the **patch opinions** (GitHub PRs) that resolve them. This is the mechanism that turns conflict resolution from forum-thread folklore into extractable, reusable patches.

## What The Forum is NOT

- Not the system of record for content (that's Git/GitHub).
- Not a build service (no code execution).
- Not an account system (identity is GitHub OAuth; there are no DebateOS-native passwords).
- Not required to compose or build a speech.

Because of these boundaries, The Forum is **secure by design and disposable**: lose it and you re-index; it never holds anything you can't already get from GitHub.

## Maintainers of configuration, not software

Community energy redirects from defending distro camps to curating opinions. Distributions are invited to own their **translators** (Ubuntu controls and optimizes Ubuntu's); curators own **points and speeches**. Long-term thesis: as AI increasingly maintains the software layer, humans curate the opinions.

## Cultural goal

By making the foundation irrelevant, there is no tribal axis left. Gaming, AI, and development communities collaborate on the same points regardless of foundation. Newcomers — including the next generation — enter Linux through *"compose your own, learn why opinions matter, remix as you grow"* rather than *"pick a side and defend it."*

## Brand voice

Tongue-in-cheek rhetoric metaphor throughout: opinions, points, speeches, debates. Tagline: *"That's just your opinion, man."* Possibly softened at launch, but the playful, anti-dogmatic tone is core identity. **There are no conclusions — only debates.**
