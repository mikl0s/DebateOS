# 04 — Conflict Resolution (Load-Bearing)

This is the core algorithmic model. It must stay **human-readable**: a user should always understand *why* a resolution happened. Resolution is **rule-based**, never a SAT solver.

## Opinion metadata (minimum set)

Every opinion declares:

- **Status within its point:** `required` | `nice-to-have`
- **Dependencies:** opinions/capabilities that must also be present
- **Conflicts:** opinions/capabilities that cannot coexist
- **Hardware conditions:** e.g. `requires: nvidia-gpu`, `requires: uefi`
- **Ordering constraints:** must-install-before / must-install-after relationships
- **Known patches:** references to patch opinions that resolve specific conflicts
- **Translator capability requirements:** what a translator must support to express this opinion

> The exact, complete schema is **derived from the Phase 0 Omarchy research** (`08`), then drafted in Phase 1. The set above is the validated floor; Phase 0 may expand it (and is expected to — e.g. arbitrary script payloads, theming assets).

## Resolution hierarchy (precise rules)

When two opinions conflict during a debate, apply in order:

1. **Required beats nice-to-have** → the required opinion wins automatically; the nice-to-have is dropped **visibly, with an explanation**.
2. **Required vs required** → **hard conflict**. The user must drop or replace one of the points — **unless** a patch opinion exists that makes both coexist, in which case it is offered as a resolution.
3. **Nice-to-have vs nice-to-have** → the system picks a sensible default or asks the user; either may be chosen.
4. **Patch opinions can override any of the above** → if a patch exists that makes both coexist, it is offered as the resolution.

## Patch opinions (first-class)

Community-contributed opinions whose purpose is to make otherwise-conflicting opinions coexist (e.g. a compatibility layer letting two stacks share a dependency at different versions).

- **Discoverable:** attached to the conflict pair in metadata; the resolver offers them automatically.
- **Not hacks:** versioned, maintained, attributable community solutions.
- **Cumulative:** over time the opinion graph becomes more connected and the resolver resolves more conflicts automatically.

## Hardware-aware resolution

The composer cross-checks opinions against the target machine's declared/scanned hardware.

Canonical example: user selects "NVIDIA local-AI point" + "AMD gaming point" but the machine has only an NVIDIA GPU → the system suggests *"You only have NVIDIA — you probably want the NVIDIA gaming point instead,"* with a one-click swap.

Version-level incompatibilities (point A needs version X, point B needs incompatible Y) use the same machinery: **declare, detect, suggest, patch.**

## Ordering

Ordering constraints (`must-install-before` / `must-install-after`) feed a **topological sort** that produces the concrete install order in the resolved speech. Cycles are a hard error surfaced at composition time with the offending opinions named.

## Community conflict-resolution workflow

When a new conflict is hit with no known resolution:

1. The system (or user) spins up a disposable VM/container reproducing the conflicting composition.
2. The community works the problem in that environment until a solution is found.
3. The solution is extracted into a **patch opinion** + metadata update via PR.
4. From then on the resolver knows: *"if people want both of these, here's how."*

This makes conflict resolution **collaborative and extractable** instead of forum-thread folklore. **The Forum** (`06`) hosts the conflict threads and links them to the resolving patch PRs — but the patch opinions themselves live in Git, so the knowledge is decentralized and survives The Forum.

## Readability requirement (invariant)

The conflict graph must remain human-readable. The visual layer (glass panes, red/green overlaps, suggested swaps) is the primary interface, with markdown/YAML underneath — but the system must be understandable from the YAML alone. **It must never decay into a byzantine dependency resolver.** When a choice is between "smarter automatic resolution" and "explainable resolution," choose explainable.
