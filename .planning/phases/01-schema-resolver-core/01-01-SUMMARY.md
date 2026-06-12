---
phase: 01-schema-resolver-core
plan: 01
status: complete
completed: 2026-06-12
requirements: [SCHM-01, SCHM-02]
key_files:
  created:
    - go.mod
    - go.sum
    - LICENSE
    - schemas/LICENSE
    - schemas/opinion.schema.json
    - schemas/point.schema.json
    - schemas/speech.schema.json
    - schemas/README.md
    - schemas/embed.go
    - examples/LICENSE
    - resolver/types.go
    - resolver/parse/parse.go
    - resolver/parse/validate.go
    - resolver/parse/schemas_embed.go
    - resolver/parse/parse_test.go
    - resolver/parse/testdata/opinion-sr-full.yaml
    - resolver/parse/testdata/point-valid.yaml
    - resolver/parse/testdata/speech-valid.yaml
    - resolver/parse/testdata/opinion-unknown-field.yaml
commits:
  - a57eacb feat(01-01) module bootstrap, licenses, shared types
  - 3a03fc5 test(01-01) RED parse tests + fixtures
  - 65f6241 feat(01-01) JSON Schemas + SR traceability README
  - 087da33 feat(01-01) GREEN parse layer
---

# Plan 01-01 Summary

Executed INLINE by the orchestrator (three executor subagent attempts were
terminated by an API content-filtering error at the same point; per the
execute-phase runtime fallback, the orchestrator ran the plan directly).

## What was built

- **Module:** `github.com/mikkelraglan/debateos` (placeholder path — no git
  remote configured; revisit when a remote exists), `go 1.24`. Exactly two
  external deps: `go.yaml.in/yaml/v3 v3.0.4` (maintained fork implementing the
  locked "yaml.v3" decision; archived `gopkg.in/yaml.v3` forbidden) and
  `github.com/santhosh-tekuri/jsonschema/v6 v6.0.2`.
- **Licenses (D3):** root `LICENSE` AGPL-3.0; `schemas/LICENSE` and
  `examples/LICENSE` CC0-1.0.
- **Shared types (`resolver/types.go`):** Opinion, Point (+PointMember),
  Speech (+PointRef, HardwareProfile), OpinionID, OpinionRef, PatchRef,
  Ordering, HardwareExpr (discriminated union leaf/and/or/not + values +
  match), FileAsset, RepoDecl (SigLevel enum: Required,
  RequiredDatabaseOptional, OptionalTrustAll, Never), RuntimeToolInstall,
  ScriptPayload, DisplayManagerConfig, BootloaderConfig, ServiceDecl,
  SysctlParam, KernelParam, GroupMembership, MimeAssoc, ThemeDecl. No
  floating-point fields anywhere (canonical-JSON parity safety).
- **Schemas:** three JSON Schema 2020-12 files covering SR-001..SR-022 with
  `additionalProperties: false` everywhere; `schemas/README.md` carries the
  full 22-row SR traceability table + documented deferrals (OQ-001 migrations,
  variant-profile → Phase 2).
- **Parse layer:** `ParseOpinion/ParsePoint/ParseSpeech(io.Reader)` — strict
  YAML decode (`KnownFields(true)`), then 2020-12 validation against embedded
  schemas, wrapped errors, no panics. TDD: RED commit 3a03fc5 precedes GREEN
  087da33 (D19).

## Deviations from plan

1. **Embed location:** `go:embed` cannot reference `../../schemas`, so the
   embed shim lives in a new root `schemas` Go package (`schemas/embed.go`,
   exporting `FS`); `resolver/parse/schemas_embed.go` reads through it. Single
   source of truth preserved (no copied schema files).
2. **Inline execution** (see header) — same tasks, same TDD sequence, same
   acceptance criteria; all verified mechanically.

## Public API for downstream plans (01-02/01-03/01-04)

```go
// package resolver
type OpinionID string
type Opinion struct{ Schema int; ID OpinionID; Name, Category, Intent, Status string; DependsOn, Conflicts []OpinionRef; KnownPatches []PatchRef; HardwareCondition *HardwareExpr; InstallPhase string; Ordering *Ordering; TranslatorCapabilities []string; Packages, RemovePackages []string; FileAssets []FileAsset; CustomRepos []RepoDecl; RuntimeToolInstalls []RuntimeToolInstall; ExecutionPhase string; ScriptPayload *ScriptPayload; DisplayManager *DisplayManagerConfig; Bootloader *BootloaderConfig; Services []ServiceDecl; SysctlParams []SysctlParam; KernelParams []KernelParam; GroupMemberships []GroupMembership; MimeAssociations []MimeAssoc; Theme *ThemeDecl }
type Point struct{ Schema int; ID, Name, Intent, Curator string; Members []PointMember }
type Speech struct{ Schema int; ID, Name, Foundation string; Points []PointRef; Opinions []OpinionRef; Hardware *HardwareProfile }
type HardwareExpr struct{ Type HardwareExprType; Predicate string; Values []string; Match string; Operands []HardwareExpr }

// package parse
func ParseOpinion(io.Reader) (*resolver.Opinion, error)
func ParsePoint(io.Reader) (*resolver.Point, error)
func ParseSpeech(io.Reader) (*resolver.Speech, error)
```

## Self-Check: PASSED

- go build ./... && go vet ./... clean
- go test ./resolver/parse/ green (6 tests)
- SR traceability: 25 SR-0 refs in README (≥22)
- OS-agnostic grep on schemas: clean
- RED commit precedes GREEN commit in history
