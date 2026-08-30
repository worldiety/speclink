package spec

// Chapter places written prose in the generated document.
//
// # Why the document needs any
//
// Every other chapter is derived: it states what the model states, and each
// sentence is written in exactly one place. That property is the point of the
// whole tool, and it has a limit. Why the system is cut this way, what was
// tried before, what a reader has to understand before the diagrams mean
// anything — none of it follows from any model, and a document without it
// describes a module to somebody who already knows what the module is for.
//
// # Why it is a declaration and not a list in a configuration file
//
// The prose lives in a Markdown file, and the one thing that can go wrong is
// that the file moves or is deleted while the outline still names it. As a
// declaration this is caught when the specification is checked, with the
// position of the offending line. In a configuration file it would be caught
// by whoever next reads the document and wonders why a chapter is missing —
// and a missing chapter looks exactly like a chapter nobody wrote.
//
// The prose keeps its own headings. The title of the chapter is the first
// heading of the file, because a title in both places is one fact written
// twice, and the copy is the one that goes stale.
type Chapter struct {
	// ID is stable, appears in diagnostics and anchors the chapter so other
	// parts of the document can point at it.
	ID string
	// Doc is the Markdown file, relative to the repository root.
	Doc string
	// At says where the chapter goes.
	At Place
}

// Place names a point in the generated document.
//
// Each value names the generated chapter the prose goes in front of, rather
// than an ordinal. An ordinal would have to be renumbered whenever a chapter
// is added, and every outline in the project would silently mean something
// else afterwards.
//
// Several chapters may share a place. They are then ordered by the file they
// are declared in and their position within it, which is stable across runs
// and is the order somebody reading the requirement tree would expect.
type Place int

const (
	// Beginning is after the note on how to read the document and before the
	// figures on where the work stands.
	Beginning Place = iota + 1
	// BeforeArchitecture precedes the layers and the rules enforced on them.
	BeforeArchitecture
	// BeforeComposition precedes the package graph.
	BeforeComposition
	// BeforeBoundary precedes the actors, the foreign systems and the
	// channels between them.
	BeforeBoundary
	// BeforeSurface precedes the addresses the system answers on.
	BeforeSurface
	// BeforeProcesses precedes the process drawings.
	BeforeProcesses
	// BeforeRequirements precedes the register of requirements.
	BeforeRequirements
	// Appendix is after everything, including the list of source documents.
	Appendix
)
