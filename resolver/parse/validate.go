package parse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

// compiledSchemas lazily compiles the three embedded JSON Schema 2020-12
// documents exactly once. Compilation failure is a programming error (the
// schemas are embedded constants) and is reported on first use.
var compiledSchemas = struct {
	once    sync.Once
	opinion *jsonschema.Schema
	point   *jsonschema.Schema
	speech  *jsonschema.Schema
	err     error
}{}

func compileAll() {
	compile := func(name string) (*jsonschema.Schema, error) {
		raw, err := schemaBytes(name)
		if err != nil {
			return nil, fmt.Errorf("read embedded schema %s: %w", name, err)
		}
		doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
		if err != nil {
			return nil, fmt.Errorf("unmarshal schema %s: %w", name, err)
		}
		c := jsonschema.NewCompiler()
		if err := c.AddResource(name, doc); err != nil {
			return nil, fmt.Errorf("add schema resource %s: %w", name, err)
		}
		s, err := c.Compile(name)
		if err != nil {
			return nil, fmt.Errorf("compile schema %s: %w", name, err)
		}
		return s, nil
	}
	cs := &compiledSchemas
	if cs.opinion, cs.err = compile("opinion.schema.json"); cs.err != nil {
		return
	}
	if cs.point, cs.err = compile("point.schema.json"); cs.err != nil {
		return
	}
	cs.speech, cs.err = compile("speech.schema.json")
}

func schemaFor(kind string) (*jsonschema.Schema, error) {
	compiledSchemas.once.Do(compileAll)
	if compiledSchemas.err != nil {
		return nil, compiledSchemas.err
	}
	switch kind {
	case "opinion":
		return compiledSchemas.opinion, nil
	case "point":
		return compiledSchemas.point, nil
	case "speech":
		return compiledSchemas.speech, nil
	}
	return nil, fmt.Errorf("unknown schema kind %q", kind)
}

// validateAgainst round-trips the typed value through JSON into the generic
// representation the validator expects, then validates it against the named
// embedded schema. This guarantees the validated shape is exactly what the
// canonical JSON encoder would emit.
func validateAgainst(kind string, v any) error {
	s, err := schemaFor(kind)
	if err != nil {
		return err
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal %s for validation: %w", kind, err)
	}
	generic, err := jsonschema.UnmarshalJSON(bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("reload %s for validation: %w", kind, err)
	}
	if err := s.Validate(generic); err != nil {
		return fmt.Errorf("%s schema validation: %w", kind, err)
	}
	return nil
}
