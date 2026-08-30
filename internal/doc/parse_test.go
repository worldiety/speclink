package doc

import "testing"

// TestProseKeepsWhatItClaims covers the constructs the prose of a
// specification actually uses.
//
// The parser is not a Markdown implementation and does not try to be. What it
// must do is never quietly change what a passage says: a numbered list that
// comes out bulleted has lost the claim that the order matters, and a listing
// that loses its newlines has lost its entire content.
func TestProseKeepsWhatItClaims(t *testing.T) {
	t.Parallel()

	t.Run("an ordered list stays ordered", func(t *testing.T) {
		blocks := ParseMarkdown("1. erstens\n2. zweitens\n", 2)
		l, ok := single[*List](t, blocks)
		if !ok {
			return
		}
		if !l.Ordered {
			t.Error("a numbered list came out unnumbered, which drops the claim that the order carries meaning")
		}
		if len(l.Items) != 2 {
			t.Fatalf("got %d items, want 2", len(l.Items))
		}
	})

	t.Run("a wrapped item stays one item", func(t *testing.T) {
		// Prose written for review is hard wrapped. Without lazy
		// continuation the tail of every wrapped item breaks out of the list,
		// lands in the text as a paragraph and restarts the numbering at one.
		blocks := ParseMarkdown("1. erstens und\n   noch mehr\n2. zweitens\n", 2)
		l, ok := single[*List](t, blocks)
		if !ok {
			return
		}
		if len(l.Items) != 2 {
			t.Fatalf("got %d items, want 2 — a continuation line broke the list", len(l.Items))
		}
		if got := plain(l.Items[0].Text); got != "erstens und noch mehr" {
			t.Errorf("got %q, want the wrapped item joined", got)
		}
	})

	t.Run("a listing keeps its lines and its indentation", func(t *testing.T) {
		blocks := ParseMarkdown("```go\nif a {\n    b()\n}\n```\n", 2)
		l, ok := single[*Listing](t, blocks)
		if !ok {
			return
		}
		if l.Lang != "go" {
			t.Errorf("got language %q, want go", l.Lang)
		}
		if l.Text != "if a {\n    b()\n}" {
			t.Errorf("got %q, the newlines or the indentation were lost", l.Text)
		}
	})

	t.Run("a code span is left alone inside", func(t *testing.T) {
		// The one place an asterisk is an asterisk. Handling emphasis first
		// would turn a documented glob into italics.
		got := ParseInline("siehe `*.spec.go` dort")
		want := []Inline{Text("siehe "), Code("*.spec.go"), Text(" dort")}
		if len(got) != len(want) {
			t.Fatalf("got %#v, want %#v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("part %d: got %#v, want %#v", i, got[i], want[i])
			}
		}
	})

	t.Run("frontmatter is not content", func(t *testing.T) {
		blocks := ParseMarkdown("---\nid: \"0004\"\n---\n\n# Titel\n", 2)
		if len(blocks) != 1 {
			t.Fatalf("got %d blocks, want only the heading: %#v", len(blocks), blocks)
		}
		if FirstHeading(blocks) != "Titel" {
			t.Errorf("got %q, want Titel", FirstHeading(blocks))
		}
	})

	t.Run("a heading is placed under the chapters around it", func(t *testing.T) {
		// The prose does not know where it was slotted in, so the level of
		// its own top heading is decided here. Passing it through unchanged
		// left every narrative chapter one rank too shallow and put it beside
		// the title of the document.
		blocks := ParseMarkdown("# oben\n\n## darunter\n", 2)
		if len(blocks) != 2 {
			t.Fatalf("got %d blocks, want 2", len(blocks))
		}
		if h := blocks[0].(*Heading); h.Level != 2 {
			t.Errorf("top heading is level %d, want 2", h.Level)
		}
		if h := blocks[1].(*Heading); h.Level != 3 {
			t.Errorf("nested heading is level %d, want 3", h.Level)
		}
	})

	t.Run("a table keeps its alignment", func(t *testing.T) {
		blocks := ParseMarkdown("| a | b |\n|---|---:|\n| 1 | 2 |\n", 2)
		tb, ok := single[*Table](t, blocks)
		if !ok {
			return
		}
		if len(tb.Rows) != 1 {
			t.Fatalf("got %d rows, want 1", len(tb.Rows))
		}
		if tb.AlignOf(1) != Right {
			t.Error("the right aligned column came out left aligned")
		}
	})
}

// single asserts the blocks are exactly one of the wanted type.
func single[T Block](t *testing.T, blocks []Block) (T, bool) {
	t.Helper()
	var zero T
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1: %#v", len(blocks), blocks)
		return zero, false
	}
	got, ok := blocks[0].(T)
	if !ok {
		t.Fatalf("got %T, want %T", blocks[0], zero)
		return zero, false
	}
	return got, true
}

// plain flattens inline parts to their text, for comparing what a passage says.
func plain(parts []Inline) string {
	out := ""
	for _, p := range parts {
		switch v := p.(type) {
		case Text:
			out += string(v)
		case Code:
			out += string(v)
		case Emph:
			out += string(v)
		case Strong:
			out += string(v)
		}
	}
	return out
}
