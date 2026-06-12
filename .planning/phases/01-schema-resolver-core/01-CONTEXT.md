# Phase 1: Schema & Resolver Core - Context

**Gathered:** 2026-06-12
**Status:** Ready for planning
**Mode:** Smart discuss (autonomous — recommended answers auto-accepted per owner directive + ADR process notes)

<domain>
## Phase Boundary

Phase 1 delivers the Opinion/Point/Speech YAML schemas (in `schemas/`, CC0) derived from the Phase 0 evidence floor (SR-001..SR-022), and the rule-based Go resolver (`resolver/` per docs/11) compiled native + WASM with identical results — built test-first against the Phase 0 EC-NNN corpus (D19). Includes 3–4 example compositions in `examples/` (CC0) exercising the harness end-to-end.

Out of scope: translators (Phase 2/4), CLI (Phase 3), any UI (Phase 5), variant-profile *implementation* (Phase 2 — only schema hooks land here), SAT/constraint solving (forbidden by D6).

</domain>

<decisions>
## Implementation Decisions

### Schema Design
- YAML schema documentation + JSON Schema (draft 2020-12) validation files in `schemas/` with CC0 LICENSE; the resolver's parse layer enforces them.
- Explicit `schema: 1` version field on every Opinion/Point/Speech document from day one.
- Full SR traceability: every SR-001..SR-022 requirement maps to a schema field or a documented deferral; traceability table lives in `schemas/README.md`.
- Speech-level foundation target (SR-022) and foundation pre-seed semantic hooks are included now; the full variant-profile schema is Phase 2 (candidate sketch stays in research/).
- Schema surprises from Phase 0 must be expressible: file/asset payloads, custom repo + keyring registration (with per-repo trust level per SR-009), runtime-tool-install category (SR-010), execution-phase field install-time vs first-run (SR-011), compound hardware predicates, arbitrary script payloads with declared capabilities, phase-level ordering.

### Resolver Architecture
- Package layout locked by docs/11: `resolver/parse`, `resolver/graph`, `resolver/resolve`, `resolver/patch`, `resolver/hardware`, `resolver/wasm` — single Go module at repo root (`module github.com/<owner>/debateos`).
- First-class `Explanation` type attached to every resolution decision: human-readable text plus structured fields (rule applied, opinions involved, what was dropped/kept and why). Every EC scenario's "expected explanation" must be producible.
- Determinism discipline: stable sorts everywhere, no map-iteration-order leaks, canonical JSON output for resolved speeches.
- Dependencies minimal: `gopkg.in/yaml.v3` only; stdlib otherwise; NO SAT/constraint libraries (D6); rule-based resolution only: toposort + pairwise conflict detection + docs/04 hierarchy + patch lookup.
- Resolution hierarchy implemented exactly per docs/04: (1) required beats nice-to-have (visible drop + explanation); (2) required-vs-required = hard conflict UNLESS patch opinion exists (offered automatically); (3) nice-vs-nice = sensible default or ask; (4) patch opinions override; ordering cycles are hard errors naming the offending opinions; hardware-conditional resolution with swap suggestions.

### TDD Harness (D19 — locked)
- The EC-NNN corpus (research/resolver-edge-cases.md, 27 scenarios) is encoded as table-driven Go tests, one case per EC with the ID in the test name — written and committed RED before implementation (GREEN) per GSD TDD gates.
- Parity proof: canonical-JSON golden files; identical fixtures run native and `GOOS=js GOARCH=wasm` (Node-based wasm_exec runner); byte-identical outputs asserted by a repeatable script.
- Coverage: ≥90% on resolver packages overall, 100% of docs/04 rule branches; enforced by a coverage-check script that fails below threshold.
- Tests must include: required-vs-required, hardware mismatch, version clash, at least one patchable pair (per ROADMAP Phase 1 spec), ordering cycle error.

### Example Compositions
- `examples/` with CC0 LICENSE (D3). Evidence-derived, not invented: a mini-Omarchy subset using real OM-NNN opinions, a clean two-point speech, one deliberately conflicting speech, one hardware-conditional speech.
- Examples are harness fixtures loaded end-to-end (parse → resolve → explain), not just documentation.

### Claude's Discretion
- Exact Go type names, internal function decomposition, golden-file layout.
- How many JSON Schema files vs one combined file, as long as Opinion/Point/Speech are each fully specified.
- Node wasm runner mechanics (wasm_exec.js wiring) details.

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- research/resolver-edge-cases.md — 27 EC-NNN Given/When/Then scenarios with expected explanations (the test corpus).
- research/schema-requirements.md — SR-001..SR-022 schema floor with evidence citations.
- research/omarchy-opinion-inventory.md — 134 real opinions usable for examples.
- research/omarchy-points.md — 32 candidate points for example point definitions.
- No code exists yet — this phase creates the Go module.

### Established Patterns
- Monorepo layout per docs/11; Markdown docs only; AGPLv3 root LICENSE + CC0 for schemas/ and examples/ (D3) — license files must be added in this phase for the new dirs (root LICENSE if absent).
- Terminology contract per docs/02 — type and field names must use the locked terms (Opinion, Point, Speech, Debate, Translator, Foundation, Registry, Patch Opinion).

### Integration Points
- Phase 2 (Arch translator) consumes the resolved-speech canonical JSON/YAML output — the resolver's output format is the translator input contract (docs/11).
- Phase 3 CLI wraps the native resolver library; Phase 5 UI calls the WASM build — both depend on this module's public API.
- Go module path: use `module github.com/mikkelraglan/debateos` placeholder unless a repo remote exists (none configured); record as a decision in STATE.md.

</code_context>

<specifics>
## Specific Ideas

- North-star alignment: the schema must be able to express every OM-NNN opinion category found in Phase 0 (the floor is empirical, D17).
- Invariant 3 (human-readable): a person must be able to understand any example composition and its resolution from the YAML alone — schemas should prefer clarity over cleverness; resolution explanations must be plain language.
- Invariant 1 (OS-agnostic): no Arch/Debian mechanics may leak into schema fields or example opinions.
- WASM/native parity is an automated test, not an inspection (D19).

</specifics>

<deferred>
## Deferred Ideas

- Variant-profile schema implementation → Phase 2 (hooks only in Phase 1 per SR-022).
- OQ-001 migrations/update primitive → recorded in open-questions; decide only if schema work forces it; default defer post-v1.0.
- Resolver-as-a-service (Forum-side indexing use) → Phase 5.

</deferred>
