// Package schemas embeds the canonical JSON Schema 2020-12 validation
// artifacts so the resolver (native and WASM) validates documents without
// filesystem access. The .schema.json files are the single source of truth.
//
// Schema content in this directory is CC0-1.0 (see LICENSE); this embed shim
// is part of the AGPL-3.0 codebase like all other Go source.
package schemas

import "embed"

// FS holds the three canonical schema files.
//
//go:embed opinion.schema.json point.schema.json speech.schema.json
var FS embed.FS
