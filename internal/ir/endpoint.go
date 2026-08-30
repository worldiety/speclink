package ir

// Endpoint is one address the system answers on.
//
// It is recognised rather than declared, for the same reason a persisted type
// is: the code that serves a route already says everything about it that can be
// said, and a second declaration would be a second source of one fact. What
// makes it a promise rather than an implementation detail is not being written
// twice — it is being frozen, exactly as a stored shape is.
type Endpoint struct {
	// Method and Path are the address. Path is the pattern as the router
	// understands it, with its parameter placeholders intact.
	Method, Path string

	// Handler is the expression that serves it, for a diagnostic to point at
	// something a reader recognises.
	Handler string

	// Package is the import path of the package that mounts it. Recorded so
	// that a later run can tell "this address is gone" from "the package that
	// mounted it was not loaded this time", which look identical from the
	// baseline alone and mean opposite things.
	Package string

	// UseCases are the recognised use cases the handler reaches. Empty when
	// the trace found none, which is a finding rather than a silence: an
	// address the system answers on and nothing accountable behind it is
	// exactly the sort of surface nobody remembers exists.
	UseCases []string

	// Request and Response are the qualified type names crossing the wire,
	// where the framework says what they are. Empty where it does not.
	Request, Response string

	// ShapesStated records that the dialect which mounted this route reports
	// its wire types at all.
	//
	// Without it an empty Response cannot be read. On a builder that states
	// them, empty means the route promises no shape — it writes bytes. On the
	// standard library's router, empty means nobody asked the question. A
	// catalogue that printed both as a blank would be answering a question it
	// had not put, which is the one thing a catalogue must not do.
	ShapesStated bool

	// Request and Response are shown and not frozen, which is deliberate.
	//
	// What a caller depends on is the shape of the body, and the name of the
	// type is not the shape. Freezing the name would report a rename that
	// changes nothing as a break, and pass a field removed from the same type
	// as unchanged — wrong in both directions at once, which is worse than
	// unmeasured. What is frozen instead is below.

	// RequestShape and ResponseShape are what actually crosses: the structure
	// of each body, expanded through named types, and its fields where it has
	// any. This is the same reading a persisted type gets, for the same
	// reason — a caller and a stored message both depend on the shape and
	// neither can see the identifier that produced it.
	//
	// Nil where the dialect states no shape at all, which is not the same as a
	// body with no fields. See ShapesStated.
	RequestShape, ResponseShape *WireShape

	// Truncated records that the trace hit its depth limit before it ran out
	// of places to look. The distinction matters: no use case found because
	// there is none is a defect in the code, and no use case found because
	// this stopped looking is a defect in the tool, and they must not report
	// as the same thing.
	Truncated bool

	// LeftScope records that the trace reached a package of this module that
	// this run did not load. It is kept apart from Truncated for the same
	// reason Truncated is kept apart from an empty UseCases: truncation is a
	// defect in speclink, and leaving the scope is a choice the operator made,
	// and neither is a defect in the code.
	//
	// It exists because the alternative is a lie. Without it a run narrowed to
	// one package reports every use case it was never asked to look at as
	// missing — the mirror of the rule this tool is built on. A direction that
	// was not measured may not be called clean, and it may not be called
	// broken either.
	LeftScope bool

	Pos Position
}

// Ref renders the endpoint as a reader addresses it.
func (e Endpoint) Ref() string { return e.Method + " " + e.Path }

// WireShape is one body crossing the boundary.
//
// It carries both readings on purpose. Fields is what a rule works with,
// because the two directions break asymmetrically and only a field by field
// view can tell a removal from an addition. Shape is the whole structure as one
// string, which is what answers for a body that is not a struct at all — a
// list, a string, a map — where there are no fields to compare and something
// still has to be promised.
type WireShape struct {
	// Type is the qualified name of the type that produced it, for a reader
	// who wants to find it. Nothing is decided on this.
	Type string
	// Shape is the structure expanded through named types, so that a change
	// inside a nested type reaches this string.
	Shape string
	// Fields are the serialised fields, empty where the body is not a struct.
	Fields []SchemaField
}
