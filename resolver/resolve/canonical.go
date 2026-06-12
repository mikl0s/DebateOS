package resolve

import (
	"encoding/json"
	"fmt"
)

// CanonicalJSON returns the deterministic canonical JSON representation of
// a ResolvedSpeech. It is the golden-file producer for the 01-05 WASM parity
// test: identical inputs on native and WASM must produce byte-identical output.
//
// Determinism guarantees (T-01-12):
//   - Uses encoding/json on well-typed structs — no map iteration.
//   - All slice fields in ResolvedSpeech are deterministically sorted by Resolve.
//   - No float64 fields anywhere in the type tree (no NaN/Inf risk).
//   - encoding/json sorts struct fields by field order (not map order), so
//     output is fully reproducible across runs.
//
// Returns an error only when json.Marshal itself fails (should never happen
// for this struct-based type — all fields are JSON-serialisable).
func CanonicalJSON(rs *ResolvedSpeech) ([]byte, error) {
	if rs == nil {
		return nil, fmt.Errorf("canonical: nil ResolvedSpeech")
	}
	b, err := json.Marshal(rs)
	if err != nil {
		return nil, fmt.Errorf("canonical: marshal failed: %w", err)
	}
	return b, nil
}
