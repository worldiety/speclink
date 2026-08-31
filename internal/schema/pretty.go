package schema

import "strings"

// Pretty sets a shape as a nested structure, one field to a line.
//
// # Why the one line form could not stay
//
// The shape grammar is written on one line because it is a fingerprint: it is
// compared, hashed and stored, and for that a single line is exactly right. It
// is not a thing to read. The desired configuration of a runner comes to six
// hundred characters on one line, which in a table cell runs off the page and
// in a paragraph is a wall nobody parses.
//
// It is also the wrong shape for a typesetter. A long unbroken run of text with
// no spaces cannot be broken at all, so it overflows the column rather than
// wrapping — the failure is not that it looks bad but that the text on the
// right is simply gone.
//
// # Why it looks like a struct
//
// Because the reader is going to write one. Whoever reads this is building the
// far end, in Go or Kotlin or TypeScript, and the shape they need is nested
// with one field to a line. Rendering it that way removes a translation step
// that every reader would otherwise do in their head, and each line is short
// enough that no typesetter has to break anything.
func Pretty(t Type) string {
	b := &strings.Builder{}
	write(b, t, 0)
	return b.String()
}

// PrettyShape parses and sets a shape, falling back to the original text.
//
// An unreadable shape is returned unchanged rather than reported. This is the
// document, not the verifier: a shape this package cannot parse is a defect in
// speclink, and printing it raw lets a reader see what it actually says instead
// of an apology.
func PrettyShape(shape string) string {
	t, err := Parse(shape)
	if err != nil {
		return shape
	}
	return Pretty(t)
}

// Composite reports whether a shape has an inside worth setting out.
//
// A string is not worth a block of its own, and neither is a list of strings.
// An object is, and so is a list of objects — those are the ones that run off
// the page.
func Composite(shape string) bool {
	t, err := Parse(shape)
	if err != nil {
		return false
	}
	return hasFields(t)
}

func hasFields(t Type) bool {
	switch t.Kind {
	case "object":
		return len(t.Fields) > 0
	case "array", "map":
		return t.Elem != nil && hasFields(*t.Elem)
	}
	return false
}

func write(b *strings.Builder, t Type, depth int) {
	switch t.Kind {
	case "object":
		if len(t.Fields) == 0 {
			b.WriteString("{}")
			return
		}
		b.WriteString("{\n")
		for _, f := range t.Fields {
			indent(b, depth+1)
			b.WriteString(f.Wire)
			b.WriteString(": ")
			write(b, f.Type, depth+1)
			b.WriteString("\n")
		}
		indent(b, depth)
		b.WriteString("}")

	case "array":
		b.WriteString("[]")
		if t.Elem != nil {
			write(b, *t.Elem, depth)
		}

	case "map":
		// The key of a map on the wire is always a string; what varies is the
		// value, and that is the half worth printing.
		b.WriteString("map[string]")
		if t.Elem != nil {
			write(b, *t.Elem, depth)
		}

	default:
		b.WriteString(t.Kind)
	}
}

func indent(b *strings.Builder, depth int) {
	for range depth {
		b.WriteString("    ")
	}
}
