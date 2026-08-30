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
			// An explicit anchor, not the one a renderer derives from the
			// heading text. Deriving it means the link silently breaks the
			// day somebody adds a word to a title, and a broken anchor in
			// Markdown reads exactly like a working one.
			if t.ID != "" {
				fmt.Fprintf(b, "<a id=%q></a>\n", anchor(t.ID))
			}
			fmt.Fprintf(b, "%s %s\n\n", strings.Repeat("#", t.Level), mdEscape(t.Text))
		case *Para:
			fmt.Fprintf(b, "%s\n\n", r.inlines(t.Text))
		case *Note:
			fmt.Fprintf(b, "> %s\n\n", r.inlines(t.Text))
		case *List:
			for i, it := range t.Items {
				fmt.Fprintf(b, "%s %s\n", mdMarker(t.Ordered, i), r.inlines(it.Text))
				for j, sub := range it.Sub {
					fmt.Fprintf(b, "  %s %s\n", mdMarker(t.Ordered, j), r.inlines(sub))
				}
			}
			b.WriteString("\n")
		case *Listing:
			// The fence is grown past the longest run of backticks inside, so
			// a listing that itself shows Markdown does not end the block
			// early and spill its remainder into the prose.
			f := strings.Repeat("`", maxRun(t.Text, '`')+1)
			if len(f) < 3 {
				f = "```"
			}
			fmt.Fprintf(b, "%s%s\n%s\n%s\n\n", f, t.Lang, t.Text, f)
		case *Figure:
			// Markdown has no figure numbering, so the caption carries the
			// weight and the anchor is a heading-shaped comment nobody sees.
			fmt.Fprintf(b, "![%s](%s)\n\n", mdEscape(t.Caption), t.Path)
			if t.Caption != "" {
				fmt.Fprintf(b, "*%s*\n\n", mdEscape(t.Caption))
			}
		case *Table:
			r.table(b, t)
		default:
			panic(fmt.Sprintf("doc: markdown has no case for %T", blk))
		}
	}
	return b.String()
}

// mdMarker is the bullet or the number of one item.
func mdMarker(ordered bool, i int) string {
	if ordered {
		return fmt.Sprintf("%d.", i+1)
	}
	return "-"
}

// maxRun is the longest run of c in s.
func maxRun(s string, c byte) int {
	best, run := 0, 0
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			run++
			best = max(best, run)
			continue
		}
		run = 0
	}
	return best
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
	case Mark:
		return mdMark(t)
	case Ref:
		// Markdown cannot check that this lands anywhere. It gets the anchor
		// GitHub and pandoc both derive from a heading, and the guarantee that
		// the target exists comes from the Typst rendering of the same tree.
		return "[" + mdEscape(t.Text) + "](#" + anchor(t.ID) + ")"
	default:
		panic(fmt.Sprintf("doc: markdown has no case for inline %T", i))
	}
}

// mdMark spells a verdict.
//
// Plain characters rather than emoji: this file is read in a terminal, in a
// diff and in a browser, and an emoji is a different width in each of them,
// which turns every table into a ragged mess exactly where it is meant to be
// scanned down a column.
func mdMark(m Mark) string {
	switch m {
	case Yes:
		return "yes"
	case Partly:
		return "part"
	case No:
		return "no"
	case Unknown:
		return "?"
	case NotRequired:
		return "n/a"
	default:
		panic(fmt.Sprintf("doc: markdown has no case for mark %d", int(m)))
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

// anchor is the fragment a heading is given and a reference points at.
//
// It is derived from the identifier, never from the title, so the two sides
// cannot drift. Same input as the Typst label, so a document that resolves in
// one format resolves in the other.
func anchor(id string) string { return label(id) }
