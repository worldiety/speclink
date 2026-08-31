package schema

import (
	"encoding/json"
	"sort"
	"strings"
)

// Dialect is the JSON Schema this package writes.
const Dialect = "https://json-schema.org/draft/2020-12/schema"

// Doc is one emitted schema file.
type Doc struct {
	// Name is the file stem, derived from the type.
	Name string
	// Body is the schema, ready to marshal.
	Body map[string]any
}

// Restriction is a rule about the values of a named type, with the cases that
// decide it.
type Restriction struct {
	Type    string
	Rule    string
	Valid   []string
	Invalid []string
}

// Shape is one structure to emit, with what is known about its top level.
type Shape struct {
	// Type is the qualified Go type that produced it.
	Type string
	// Shape is the grammar speclink writes.
	Shape string
	// Optional names the top level fields that may be absent, by wire name.
	Optional map[string]bool
	// FieldTypes maps a top level wire name onto the named type it was
	// declared with, so a restricted value can be pointed at its rule.
	FieldTypes map[string]string
	// Docs maps a top level wire name onto what the comment beside the field
	// says. JSON Schema calls it description, and it is the field a generator
	// turns back into a comment on the far end.
	Docs map[string]string
}

// Of renders one shape as a schema document.
//
// # Why the required list stops at the top level
//
// Because that is where the knowledge stops. speclink records whether a top
// level field may be omitted; below that it compares the structure as a whole
// and never asked the question. Emitting a required list for a nested object
// would be inventing one, and a schema that is confidently wrong about
// presence is worse than one that says nothing: a generator turns the first
// into a parser that dereferences an absent value.
func Of(s Shape, restricted map[string]Restriction) (Doc, error) {
	t, err := Parse(s.Shape)
	if err != nil {
		return Doc{}, err
	}

	body := jsonType(t, false)
	body["$schema"] = Dialect
	body["title"] = lastSegment(s.Type)
	body["$id"] = fileStem(s.Type) + ".schema.json"

	if t.Kind == "object" {
		applyTopLevel(body, t, s, restricted)
	}
	return Doc{Name: fileStem(s.Type), Body: body}, nil
}

// applyTopLevel adds what is known only about the outermost object: which
// fields may be absent, and which carry a rule about their values.
func applyTopLevel(body map[string]any, t Type, s Shape, restricted map[string]Restriction) {
	props, _ := body["properties"].(map[string]any)

	var required []string
	for _, f := range t.Fields {
		if !s.Optional[f.Wire] {
			required = append(required, f.Wire)
		}
		if text := s.Docs[f.Wire]; text != "" && props != nil {
			if p, ok := props[f.Wire].(map[string]any); ok {
				p["description"] = text
			}
		}
		named := s.FieldTypes[f.Wire]
		r, ok := restricted[named]
		if !ok || props == nil {
			continue
		}
		// The rule cannot be enforced by the schema — it is prose — so it is
		// carried where a reader will meet it, and the cases that decide it
		// stay in the vector file where a test can run them.
		if p, ok := props[f.Wire].(map[string]any); ok {
			p["$comment"] = r.Rule
			p["x-speclink-type"] = named
		}
	}
	sort.Strings(required)
	if len(required) > 0 {
		body["required"] = required
	}
}

// jsonType renders one node. nested marks everything below the outermost
// object, where presence was never recorded.
func jsonType(t Type, nested bool) map[string]any {
	switch t.Kind {
	case "object":
		props := map[string]any{}
		for _, f := range t.Fields {
			props[f.Wire] = jsonType(f.Type, true)
		}
		out := map[string]any{"type": "object", "properties": props}
		if nested && len(t.Fields) > 0 {
			out["$comment"] = "no required list: speclink compares nested structure as a whole and does not record which of these may be omitted"
		}
		return out

	case "array":
		return map[string]any{"type": "array", "items": jsonType(*t.Elem, true)}

	case "map":
		// A JSON object with unknown keys. additionalProperties carries the
		// value type, which is the whole of what a map promises.
		return map[string]any{"type": "object", "additionalProperties": jsonType(*t.Elem, true)}

	case "int":
		return map[string]any{"type": "integer"}
	case "number":
		return map[string]any{"type": "number"}
	case "bool":
		return map[string]any{"type": "boolean"}
	case "string":
		return map[string]any{"type": "string"}
	}
	// any: no constraint at all, which is what an interface promises.
	return map[string]any{}
}

// Vectors is the file of cases every conforming implementation is judged by.
//
// Written beside the schemas rather than inside them, and deliberately. The
// interesting half of a restriction is what must be refused, and a JSON Schema
// expresses that as a negated pattern which generators drop without a sound.
// A list of cases survives every generator, every language and every team,
// because "this must be rejected" means the same thing everywhere.
func Vectors(rs []Restriction) map[string]any {
	types := map[string]any{}
	for _, r := range rs {
		types[r.Type] = map[string]any{
			"rule":    r.Rule,
			"valid":   nonNil(r.Valid),
			"invalid": nonNil(r.Invalid),
		}
	}
	return map[string]any{
		"$comment": "cases a conforming implementation must accept and must reject; the rule beside them is prose and cannot be checked by a schema",
		"types":    types,
	}
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// Marshal renders a schema body with a stable, readable layout.
func Marshal(v any) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

// lastSegment is the type name a reader recognises.
func lastSegment(qualified string) string {
	if i := strings.LastIndexByte(qualified, '.'); i >= 0 {
		return qualified[i+1:]
	}
	return qualified
}

// fileStem turns a qualified type into a file name.
//
// The package is dropped and the name lowered with underscores, which is what
// the neighbouring generated files look like and what a shell completes
// without quoting.
func fileStem(qualified string) string {
	name := lastSegment(qualified)
	b := &strings.Builder{}
	for i, r := range name {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r - 'A' + 'a')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
