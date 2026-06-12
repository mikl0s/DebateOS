# DebateOS — Doc Ingest Conflicts

Mode: new | Docs: 12 | Generated: 2026-06-12

## Conflict Detection Report

### BLOCKERS (0)

None. Single locked ADR (docs/09-decisions.md); no LOCKED-vs-LOCKED contradictions, no UNKNOWN/low-confidence classifications, no existing .planning context to collide with (fresh bootstrap).

### WARNINGS (0)

None. The two PRDs (docs/01-vision.md, docs/06-social-layer.md) cover complementary scopes; no requirement appears in both with divergent acceptance criteria.

### INFO (5)

[INFO] Cross-reference cycle detected: docs/04-conflict-resolution.md <-> docs/06-social-layer.md
  Found: docs/04-conflict-resolution.md cross-refs 06 (Forum hosts conflict threads); docs/06-social-layer.md cross-refs 04 (collaborative conflict-resolution workflow)
  Note: Both edges are informational see-also references with no contradicting content. Synthesis order is fixed by manifest precedence integers (04=12, 06=21), not reference traversal, so no synthesis loop occurs. Recorded for transparency; no action required.

[INFO] Cross-reference cycle detected: docs/09-decisions.md <-> docs/10-prior-art-and-lessons.md
  Found: docs/09-decisions.md D18 cross-refs 10 (carry-forward scope); docs/10-prior-art-and-lessons.md cites decision IDs D6, D11, D13 in 09
  Note: Mutual citations, no contradiction — doc 10 consistently supports the locked decisions it cites. Precedence (09=0, 10=31) makes the ADR authoritative regardless of reference direction. Recorded for transparency; no action required.

[INFO] Auto-resolved: ADR > DOC on decision authority
  Note: docs/00-START-HERE.md (DOC, precedence 30) restates the locked tech stack and mandate ("Tech at a glance — all locked"); per precedence, docs/09-decisions.md (ADR, precedence 0) is the sole authoritative source for D1–D18 and the invariants. The restatements were verified consistent and treated as non-authoritative summary in synthesized intel (decisions.md sources only docs/09).

[INFO] Auto-resolved: ADR > PRD on v1.0 foundation scope
  Note: docs/01-vision.md mentions translators for "Arch, Debian, Fedora, or whatever foundation is chosen" as the general model; docs/09-decisions.md D9 locks v1.0 foundations to Arch (+ structure for 1–2 Arch variants) and Debian, and docs/07-roadmap.md places Fedora/other translators post-v1.0 (community-owned). Intel records Fedora-and-beyond as post-v1.0 vision, not a v1.0 requirement.

[INFO] Auto-resolved: ADR > SPEC scope clarification on install-time hardware detection
  Note: docs/03-architecture.md describes a small installer-side hardware scan, while docs/09-decisions.md D2 forbids building the Phase 6 hardware-scanning installer in v1.0. docs/03 itself disambiguates: v1.0 = declared hardware + basic install-time resolution; the full hardware-scanning installer is Phase 6 / post-v1.0. Intel (constraints.md "Hardware detection scope") records the v1.0-bounded reading consistent with D2.
