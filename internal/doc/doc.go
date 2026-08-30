// Package doc is the shape of the specification, with no opinion on how it is
// spelled.
//
// The document generator used to write Markdown directly, and the comment above
// it argued — correctly — that a second backend would be a second thing to
// maintain, which is the exact failure this tool exists to end. A parallel
// writeTypst walking the same graph would have been that: two places deciding
// what the document says, drifting apart on the first change.
//
// This is the other way out. There is still exactly one place that decides what
// the specification contains; what moved is the spelling. A renderer here
// cannot invent a fact, because it is handed a tree of headings, tables and
// sentences and gets no access to the model that produced them. Two outputs
// that disagree is not a bug that can be introduced by editing a renderer, and
// that property is worth more than either output.
//
// The set of node types is deliberately small — it is what the specification
// actually uses, not what a document format can express. Every addition here is
// a new obligation for every backend, so the bar is a chapter that cannot be
// written without it.
package doc

import "fmt"

// Doc is a whole document: a title, and the blocks under it in order.
type Doc struct {
	Title  string
	Blocks []Block
}

// New starts a document.
func New(title string) *Doc { return &Doc{Title: title} }

// Block is one thing at the top level of a document.
//
// The interface is closed by an unexported method: a backend has to handle
// every case, and adding a case is a compile error in every backend rather
// than silently rendering as nothing.
type Block interface{ block() }

// Heading opens a chapter or a section.
//
// Level is 2 or deeper. Level 1 is the document title, which is a field on Doc
// rather than a block, because a document has exactly one and a backend needs
// it before it writes anything else.
type Heading struct {
	Level int
	Text  string

	// ID, when set, makes this heading referable by Ref.
	//
	// Markdown gets an anchor it cannot check; Typst gets a label it can. That
	// difference is the whole reason the document is a tree and not a string:
	// a cross reference to a requirement that no longer exists should be
	// caught by the renderer that is able to catch it.
	ID string
}

// Para is a paragraph.
type Para struct{ Text []Inline }

// Note is an aside, set apart from the argument around it.
//
// Used for what a run could not measure and for scope that was excluded. It is
// a distinct type rather than an italic paragraph because those two facts are
// the ones a reader must not skim past, and a backend that can make them
// unskippable should be allowed to.
type Note struct{ Text []Inline }

// List is a sequence of items, each of which may carry a flat list under it.
//
// One level of nesting, and no more. It is here because a requirement lists
// what implements it and each of those carries a source range, which is a
// hierarchy and not a sequence — the bar set for adding a node was a chapter
// that cannot be written without it, and that chapter is one. Faking the
// indentation with spaces inside the item text was tried and produced a list
// that looked nested in Markdown and was flat in every other sense, which is
// the outcome this whole package exists to make impossible.
//
// A second level would be a design failure in the document, so a backend
// never has to decide what one looks like.
type List struct{ Items []Bullet }

// Bullet is one item, and whatever hangs under it.
type Bullet struct {
	Text []Inline
	Sub  [][]Inline
}

// Table is a header row and the rows under it.
type Table struct {
	Head  []string
	Align []Align
	Rows  [][][]Inline
}

// Align is how a column is set.
type Align int

const (
	Left Align = iota
	Right
)

func (*Heading) block() {}
func (*Para) block()    {}
func (*Note) block()    {}
func (*List) block()    {}
func (*Table) block()   {}

// Inline is a run of text inside a block.
type Inline interface{ inline() }

// Text is literal words. A backend must escape it; it is never markup.
type Text string

// Code is an identifier, a path or a symbol: something spelled exactly.
type Code string

// Emph is stress.
type Emph string

// Strong is a label the eye should land on, such as the lead-in of a list item.
type Strong string

// Link points out of the document.
type Link struct{ Text, URL string }

// Ref points inside it, at a Heading carrying the same ID.
type Ref struct{ ID, Text string }

// Break ends a line without ending the block.
//
// Used in table cells that carry a label over its endpoints. It is a node
// rather than a literal <br> in the text because that string is HTML: correct
// in one of the two backends, escaped into visible noise by the other, and
// wrong in any third.
type Break struct{}

func (Text) inline()   {}
func (Code) inline()   {}
func (Emph) inline()   {}
func (Strong) inline() {}
func (Link) inline()   {}
func (Ref) inline()    {}
func (Break) inline()  {}

// T is literal text.
//
// It deliberately takes no arguments. A single T(format, args...) would make
// every call site a printf call, and the text flowing through here is written
// by people in requirement documents this tool does not own — a percent sign
// in "covers 100% of the ledger" would become a verb and eat the sentence.
// vet catches that in this package and nowhere in a caller, so the two jobs
// are two functions.
func T(s string) Text { return Text(s) }

// Tf is text assembled from values speclink computed itself.
func Tf(format string, args ...any) Text { return Text(fmt.Sprintf(format, args...)) }

// H adds a heading.
func (d *Doc) H(level int, text string) *Doc {
	d.Blocks = append(d.Blocks, &Heading{Level: level, Text: text})
	return d
}

// HID adds a heading that can be referred to.
func (d *Doc) HID(level int, id, text string) *Doc {
	d.Blocks = append(d.Blocks, &Heading{Level: level, Text: text, ID: id})
	return d
}

// P adds a paragraph.
func (d *Doc) P(text ...Inline) *Doc {
	d.Blocks = append(d.Blocks, &Para{Text: text})
	return d
}

// Pf adds a paragraph of formatted text, the common case.
func (d *Doc) Pf(format string, args ...any) *Doc {
	return d.P(Tf(format, args...))
}

// Note adds an aside.
func (d *Doc) Note(text ...Inline) *Doc {
	d.Blocks = append(d.Blocks, &Note{Text: text})
	return d
}

// Notef adds an aside of formatted text.
func (d *Doc) Notef(format string, args ...any) *Doc {
	return d.Note(Tf(format, args...))
}

// Bullets adds a list. An empty list adds nothing, so callers do not each have
// to guard a heading against a chapter that turned out to have no entries.
func (d *Doc) Bullets(items ...Bullet) *Doc {
	if len(items) == 0 {
		return d
	}
	d.Blocks = append(d.Blocks, &List{Items: items})
	return d
}

// Item is one list item, spelled to keep call sites readable.
func Item(text ...Inline) Bullet { return Bullet{Text: text} }

// Under hangs a sub item off a bullet.
func (b Bullet) Under(text ...Inline) Bullet {
	b.Sub = append(b.Sub, text)
	return b
}

// Row is one table row.
func Row(cells ...[]Inline) [][]Inline { return cells }

// Cell is one table cell.
func Cell(text ...Inline) []Inline { return text }

// Grid adds a table. Align may be shorter than Head, and missing columns are
// left aligned.
func (d *Doc) Grid(head []string, align []Align, rows ...[][]Inline) *Doc {
	t := &Table{Head: head, Align: align, Rows: rows}
	d.Blocks = append(d.Blocks, t)
	return d
}

// Table adds an empty table and returns it, for callers that build rows in a
// loop.
func (d *Doc) Table(head ...string) *Table {
	t := &Table{Head: head}
	d.Blocks = append(d.Blocks, t)
	return t
}

// Aligned sets the column alignment and returns the table.
func (t *Table) Aligned(a ...Align) *Table {
	t.Align = a
	return t
}

// Add appends a row.
func (t *Table) Add(cells ...[]Inline) *Table {
	t.Rows = append(t.Rows, cells)
	return t
}

// AlignOf reports how column i is set, defaulting to Left.
func (t *Table) AlignOf(i int) Align {
	if i < len(t.Align) {
		return t.Align[i]
	}
	return Left
}

// Renderer turns a document into bytes.
type Renderer interface {
	Render(*Doc) string
	// Ext is the file extension, without the dot, that this output belongs in.
	Ext() string
}
