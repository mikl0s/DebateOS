// Package parse strictly decodes Opinion/Point/Speech YAML documents into
// the typed structs from the resolver package and validates them against the
// embedded JSON Schema 2020-12 definitions.
//
// Decoding is strict (unknown fields are errors — typos never pass silently),
// validation is structural (required fields, enums, recursive hardware
// expressions), and all failures return wrapped errors; nothing panics on
// malformed input. Alias expansion is bounded by the YAML library's built-in
// protection, so hostile documents cannot blow up memory.
package parse

import (
	"fmt"
	"io"

	yaml "go.yaml.in/yaml/v3"

	resolver "github.com/mikkelraglan/debateos/resolver"
)

// decodeStrict decodes one YAML document into out, rejecting unknown fields.
func decodeStrict(r io.Reader, out any) error {
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	if err := dec.Decode(out); err != nil {
		return err
	}
	return nil
}

// ParseOpinion reads one Opinion YAML document, strictly decodes it, and
// validates it against the embedded opinion schema.
func ParseOpinion(r io.Reader) (*resolver.Opinion, error) {
	var op resolver.Opinion
	if err := decodeStrict(r, &op); err != nil {
		return nil, fmt.Errorf("parse opinion: %w", err)
	}
	if err := validateAgainst("opinion", &op); err != nil {
		return nil, fmt.Errorf("parse opinion: %w", err)
	}
	return &op, nil
}

// ParsePoint reads one Point YAML document, strictly decodes it, and
// validates it against the embedded point schema.
func ParsePoint(r io.Reader) (*resolver.Point, error) {
	var p resolver.Point
	if err := decodeStrict(r, &p); err != nil {
		return nil, fmt.Errorf("parse point: %w", err)
	}
	if err := validateAgainst("point", &p); err != nil {
		return nil, fmt.Errorf("parse point: %w", err)
	}
	return &p, nil
}

// ParseSpeech reads one Speech YAML document, strictly decodes it, and
// validates it against the embedded speech schema.
func ParseSpeech(r io.Reader) (*resolver.Speech, error) {
	var s resolver.Speech
	if err := decodeStrict(r, &s); err != nil {
		return nil, fmt.Errorf("parse speech: %w", err)
	}
	if err := validateAgainst("speech", &s); err != nil {
		return nil, fmt.Errorf("parse speech: %w", err)
	}
	return &s, nil
}
