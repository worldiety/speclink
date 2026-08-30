package doc

import (
	"fmt"
	"strings"
)

// Markdown renders the document as Markdown.
//
// It renders everywhere, it diffs, and a diff is the form in which this
// document is actually reviewed. That argument has not changed and this is
// still the default output; what changed is that it is now one renderer among
// two rather than the only way the document can exist.
type Markdown struct{}

func (Markdown) Ext() string { return "md" }

func (r Markdown) Render(d *Doc) string {
	b := &strings.Builder{}
	if d.Title != "" {
		fmt.Fprintf(b, "# %s\n\n", mdEscape(d.Title))
	}
	for _, blk := range d.Blocks {
		switch t := blk.(type) {
		case *Heading:
			fmt.Fprintf(b, "%s %s\n\n", strings.Repeat("#", t.Level), mdEscape(t.Text))
		case *Para:
			fmt.Fprintf(b, "%s\n\n", r.inlines(t.Text))
		case *Note:
			fmt.Fprintf(b, "> %s\n\n", r.inlines(t.Text))
		case *List:
			for _, it := range t.Items {
				fmt.Fprintf(b, "- %s\n", r.inlines(it.Text))
				for _, sub := range it.Sub {
					fmt.Fprintf(b, "  - %s\n", r.inlines(sub))
				}
			}
			b.WriteString("\n")
		case *Table:
			r.table(b, t)
		default:
			panic(fmt.Sprintf("doc: markdown has no case for %T", blk))
		}
	}
	return b.String()
}

func (r Markdown) table(b *strings.Builder, t *Table) {
	b.WriteString("|")
	for _, h := range t.Head {
		fmt.Fprintf(b, " %s |", h)
	}
	b.WriteString("\n|")
	for i := range t.Head {
		if t.AlignOf(i) == Right {
			b.WriteString("---:|")
		} else {
			b.WriteString("---|")
		}
	}
	b.WriteString("\n")
	for _, row := range t.Rows {
		b.WriteString("|")
		for _, c := range row {
			fmt.Fprintf(b, " %s |", r.inlines(c))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
}

func (r Markdown) inlines(in []Inline) string {
	b := &strings.Builder{}
	for _, i := range in {
		b.WriteString(r.inline(i))
	}
	return b.String()
}

func (r Markdown) inline(i Inline) string {
	switch t := i.(type) {
	case Text:
		return mdEscape(string(t))
	case Code:
		return "`" + string(t) + "`"
	case Emph:
		return "_" + mdEscape(string(t)) + "_"
	case Strong:
		return "**" + mdEscape(string(t)) + "**"
	case Link:
		return "[" + mdEscape(t.Text) + "](" + t.URL + ")"
	case Break:
		return "<br>"
	case Ref:
		// Markdown cannot check that this lands anywhere. It gets the anchor
		// GitHub and pandoc both derive from a heading, and the guarantee that
		// the target exists comes from the Typst rendering of the same tree.
		return "[" + mdEscape(t.Text) + "](#" + anchor(t.ID) + ")"
	default:
		panic(fmt.Sprintf("doc: markdown has no case for inline %T", i))
	}
}

// mdEscape protects the characters that would otherwise become markup.
//
// Requirement text is written by people in a document this tool does not own,
// so an underscore in an identifier or a pipe in a sentence has to survive
// into a table cell without splitting it.
func mdEscape(s string) string {
	return strings.NewReplacer(
		`\`, `\\`,
		"`", "\\`",
		"*", `\*`,
		"_", `\_`,
		"|", `\|`,
		"[", `\[`,
		"]", `\]`,
		"<", `\<`,
	).Replace(s)
}

// anchor derives the fragment a heading gets, the way GitHub does it.
func anchor(s string) string {
	b := &strings.Builder{}
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('-')
		}
	}
	return b.String()
}
