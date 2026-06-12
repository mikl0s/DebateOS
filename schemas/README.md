# DebateOS Schemas

JSON Schema 2020-12 definitions for the three DebateOS document types:

| File | Document type | What it is |
|------|---------------|------------|
| `opinion.schema.json` | Opinion | One atomic, OS-agnostic configuration decision |
| `point.schema.json` | Point | A curated, coherent bundle of opinions |
| `speech.schema.json` | Speech | A user's complete composition for a foundation |

All schema content is dedicated to the public domain under [CC0-1.0](LICENSE) (D3).
Documents are authored in YAML; the `.json` schema files are the canonical
validation artifacts loaded by the resolver's parse layer. Every document
carries a required top-level `schema: 1` version field.

The schema floor below is **empirical, not theoretical** (D17): every field
traces to evidence in `research/schema-requirements.md` (SR-001..SR-022),
which itself traces to real decisions found in the Omarchy source and the
CachyOS/Garuda variant study.

## SR-001..SR-022 Traceability

| SR | Requirement | Schema file | JSON Schema key | Status |
|----|-------------|-------------|-----------------|--------|
| SR-001 | Required vs nice-to-have status | opinion.schema.json | `status` enum | covered |
| SR-002 | Opinion dependencies | opinion.schema.json | `depends_on` → `$defs/opinionRef` | covered |
| SR-003 | Conflict declarations + known patches | opinion.schema.json | `conflicts`, `known_patches` → `$defs/patchRef` | covered |
| SR-004 | Single hardware predicate | opinion.schema.json | `hardware_condition` leaf node (`predicate`) | covered |
| SR-005 | Compound hardware predicates (AND/OR/NOT, set-membership, string-match) | opinion.schema.json | `hardware_condition` → recursive `$defs/hardwareExpr` (`operands`, `values`, `match`) | covered |
| SR-006 | Phase-level + relative ordering (incl. cross-phase exceptions) | opinion.schema.json | `install_phase` enum, `ordering.before/after` | covered |
| SR-007 | Translator capability declaration | opinion.schema.json | `translator_capabilities` | covered |
| SR-008 | File asset payloads | opinion.schema.json | `file_assets` → `$defs/fileAsset` | covered |
| SR-009 | Custom repo registration with trust levels | opinion.schema.json | `custom_repos` → `$defs/repoDecl` (`sig_level` enum: Required, RequiredDatabaseOptional, OptionalTrustAll, Never; `keyring`; `priority`) | covered |
| SR-010 | Runtime tool installs (language package managers) | opinion.schema.json | `runtime_tool_installs` → `$defs/runtimeToolInstall` | covered |
| SR-011 | Execution phase: install-time vs first-run | opinion.schema.json | `execution_phase` enum | covered |
| SR-012 | Arbitrary script payloads with declared capabilities | opinion.schema.json | `script_payload` → `$defs/scriptPayload` | covered |
| SR-013 | Display manager configuration | opinion.schema.json | `display_manager` → `$defs/displayManager` | covered |
| SR-014 | Bootloader configuration | opinion.schema.json | `bootloader` → `$defs/bootloader` | covered |
| SR-015 | Service enablement (incl. deferred/first-boot) | opinion.schema.json | `services` → `$defs/serviceDecl` (`deferred`) | covered |
| SR-016 | Sysctl parameters via drop-in files (per-key collision detection) | opinion.schema.json | `sysctl_params` → `$defs/sysctlParam` (`key`, `value`, `drop_in_file`) | covered |
| SR-017 | Kernel boot parameters | opinion.schema.json | `kernel_params` → `$defs/kernelParam` | covered |
| SR-018 | User/group memberships | opinion.schema.json | `group_memberships` → `$defs/groupMembership` | covered |
| SR-019 | MIME type associations | opinion.schema.json | `mime_associations` → `$defs/mimeAssoc` | covered |
| SR-020 | Theme bundles (assets + activation symlinks + default flag) | opinion.schema.json | `theme` → `$defs/themeDecl` | covered |
| SR-021 | Point: name, intent, members with per-point status | point.schema.json | top-level + `members[]` (`id`, `status`) | covered |
| SR-022 | Speech: foundation target + point list + declared hardware | speech.schema.json | `foundation`, `points[]`, `opinions[]`, `hardware` | covered |

### Documented deferrals

- **Migrations/update primitive (OQ-001):** intentionally absent from `schema: 1`;
  deferred post-v1.0. The `schema` version field exists so a future revision can
  add it without breaking existing documents.
- **Variant-profile document type (D20/OQ-007):** the candidate sketch lives in
  `research/arch-variants-delta.md`; the speech-level `foundation` field (SR-022)
  is the Phase 1 hook. The full variant-profile schema is Phase 2 scope.

## Design rules

- **OS-agnostic (invariant 1):** no distribution-specific vocabulary appears in
  any schema title, description, enum, or field name. Opinions express intent;
  translators own mechanics.
- **Human-readable (invariant 3):** documents must be comprehensible from the
  YAML alone; field names are plain words from the locked terminology contract
  (docs/02-concepts.md).
- **Strict by default:** `additionalProperties: false` at every level — typos
  are rejected, never silently ignored.
- **No floating-point fields:** numeric data is integers or strings, keeping
  canonical JSON byte-identical between native and WASM resolver builds.
