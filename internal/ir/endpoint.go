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

	// UseCases are the recognised use cases the handler reaches. Empty when
	// the trace found none, which is a finding rather than a silence: an
	// address the system answers on and nothing accountable behind it is
	// exactly the sort of surface nobody remembers exists.
	UseCases []string

	// Request and Response are the qualified type names crossing the wire,
	// where the framework says what they are. Empty where it does not.
	Request, Response string

	// Truncated records that the trace hit its depth limit before it ran out
	// of places to look. The distinction matters: no use case found because
	// there is none is a defect in the code, and no use case found because
	// this stopped looking is a defect in the tool, and they must not report
	// as the same thing.
	Truncated bool

	Pos Position
}

// Ref renders the endpoint as a reader addresses it.
func (e Endpoint) Ref() string { return e.Method + " " + e.Path }
