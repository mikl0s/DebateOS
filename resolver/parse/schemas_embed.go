package parse

import (
	"github.com/mikkelraglan/debateos/schemas"
)

// schemaBytes returns the raw bytes of one embedded canonical schema file.
// The schemas package at the repo root is the single source of truth — this
// indirection exists so parse never reads the filesystem at runtime.
func schemaBytes(name string) ([]byte, error) {
	return schemas.FS.ReadFile(name)
}
