package spec

// TopicID identifies a theme.
type TopicID string

// Topic is a theme a requirement belongs to, and a chapter of the document.
//
// # Why this is not the same as a standard
//
// A standard imposes from outside: its clauses are obligations somebody
// accepted, they can be enumerated, and the interesting question is which of
// them nothing answers. A topic orders from inside: it is how the people who
// built the thing think about it, it cannot be enumerated in advance, and there
// is no such thing as a topic nobody has covered.
//
// The two are kept apart because folding them together would make one of the
// questions unanswerable. A theme in a list of obligations reads as an
// obligation nobody imposed; an obligation in a list of themes loses the
// coverage that is the whole reason it is written down.
//
// # Why it is not required
//
// A requirement without a topic lands in a chapter of its own, counted. Forcing
// one on every requirement buys a complete table of contents at the price of a
// decision at every declaration, most of which would be made carelessly — and a
// carelessly assigned theme is worse than none, because it looks like somebody
// thought about it.
type Topic struct {
	// ID is stable and appears in diagnostics.
	ID TopicID
	// Title heads the chapter.
	Title string
	// Description says what belongs here and what does not, which is the half
	// that keeps a theme from swallowing everything next to it.
	Description string
}
