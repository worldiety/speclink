// Package schema turns the shapes speclink already reads into JSON Schema.
//
// # Why this is a translation and not a description
//
// Nothing here is declared. Every structure it emits was read from a Go type
// and is already used to decide whether a promise was broken. Writing a schema
// by hand beside those types would be the same fact in two places, and the two
// would disagree the first time somebody added a field — which is precisely
// the failure the schema was supposed to prevent.
//
// # What it deliberately does not claim
//
// A nested object carries no required list. speclink compares nested structure
// as a whole and does not record which of its fields may be omitted, so a
// required list there would be invented. An absent list means unconstrained,
// which is the honest reading: a consumer that generates a parser from it will
// handle absence, and handling an absence that never happens costs nothing.
// Claiming a field is required and being wrong costs a crash.
package schema

import (
	"sort"
	"strings"
)

// Type is one node of the shape grammar speclink writes.
type Type struct {
	// Kind is "string", "int", "bool", "number", "any", "object", "array" or
	// "map".
	Kind string
	// Elem is the element of an array or the value of a map.
	Elem *Type
	// Fields are the members of an object, in the order they were written.
	Fields []Field
}

// Field is one member of an object in the shape grammar.
type Field struct {
	Wire string
	Type Type
}

// Parse reads one shape as speclink writes it.
//
// The grammar is closed and this package's own output produces it, so anything
// unrecognised is reported rather than guessed at: a schema derived from a
// shape nobody could read would be a confident description of nothing.
func Parse(shape string) (Type, error) {
	p := &parser{in: shape}
	t, err := p.parse()
	if err != nil {
		return Type{}, err
	}
	if p.pos != len(p.in) {
		return Type{}, errAt(p.in, p.pos, "trailing text")
	}
	return t, nil
}

type parser struct {
	in  string
	pos int
}

func (p *parser) parse() (Type, error) {
	switch {
	case p.has("[]"):
		p.pos += 2
		elem, err := p.parse()
		if err != nil {
			return Type{}, err
		}
		return Type{Kind: "array", Elem: &elem}, nil

	case p.has("map["):
		return p.parseMap()

	case p.has("{"):
		return p.parseObject()
	}
	return p.parseBasic()
}

// parseMap reads map[K]V.
//
// The key is read and dropped. JSON object keys are strings whatever the Go
// type says, so recording the key type here would describe a distinction the
// encoding does not have.
func (p *parser) parseMap() (Type, error) {
	p.pos += len("map[")
	depth := 1
	start := p.pos
	for p.pos < len(p.in) && depth > 0 {
		switch p.in[p.pos] {
		case '[':
			depth++
		case ']':
			depth--
		}
		p.pos++
	}
	if depth != 0 {
		return Type{}, errAt(p.in, start, "unterminated map key")
	}
	elem, err := p.parse()
	if err != nil {
		return Type{}, err
	}
	return Type{Kind: "map", Elem: &elem}, nil
}

func (p *parser) parseObject() (Type, error) {
	p.pos++ // {
	out := Type{Kind: "object"}
	if p.has("}") {
		p.pos++
		return out, nil
	}
	for {
		name := p.until(':')
		if name == "" {
			return Type{}, errAt(p.in, p.pos, "field without a name")
		}
		p.pos++ // :
		ft, err := p.parse()
		if err != nil {
			return Type{}, err
		}
		out.Fields = append(out.Fields, Field{Wire: name, Type: ft})

		switch {
		case p.has(","):
			p.pos++
		case p.has("}"):
			p.pos++
			return out, nil
		default:
			return Type{}, errAt(p.in, p.pos, "expected , or }")
		}
	}
}

func (p *parser) parseBasic() (Type, error) {
	start := p.pos
	for p.pos < len(p.in) && !strings.ContainsRune(",}", rune(p.in[p.pos])) {
		p.pos++
	}
	word := p.in[start:p.pos]
	switch word {
	case "string":
		return Type{Kind: "string"}, nil
	case "int":
		return Type{Kind: "int"}, nil
	case "bool":
		return Type{Kind: "bool"}, nil
	case "float32", "float64":
		return Type{Kind: "number"}, nil
	case "any", "unknown", "...":
		// Three different reasons for the same silence, and none of them is a
		// structure: an interface decides its content elsewhere, an
		// unrecognised type was never read, and a recursive one was cut off.
		return Type{Kind: "any"}, nil
	}
	return Type{}, errAt(p.in, start, "unknown shape "+quote(word))
}

// until reads to the next occurrence of c at the top nesting level.
func (p *parser) until(c byte) string {
	start, depth := p.pos, 0
	for p.pos < len(p.in) {
		switch p.in[p.pos] {
		case '{', '[':
			depth++
		case '}', ']':
			depth--
		case c:
			if depth == 0 {
				return p.in[start:p.pos]
			}
		}
		p.pos++
	}
	return ""
}

func (p *parser) has(s string) bool { return strings.HasPrefix(p.in[p.pos:], s) }

// sortedWires returns the field names of an object, for a stable required list.
func sortedWires(fields []Field) []string {
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		out = append(out, f.Wire)
	}
	sort.Strings(out)
	return out
}

func quote(s string) string { return `"` + s + `"` }
