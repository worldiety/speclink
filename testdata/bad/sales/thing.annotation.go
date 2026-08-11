package sales

import (
	"example.com/bad/anforderungen/fun/quote"
	"github.com/worldiety/speclink/spec"
)

// A helper function has no business in an annotation file.
func helper() string { return "nope" }

// External may only be attached to a type, and the field does not exist.
var _ = spec.ForField[Thing]("Missing",
	spec.Satisfies(quote.RWrongPrefix),
	spec.External(),
)

// A named binding, and Help given twice.
var named = spec.For[Thing](
	spec.Help("first"),
	spec.Help("second"),
	spec.Waive("K3-REQ-UNCOVERED", ""),
)

// The address-of operator is no longer part of the language.
var _ = spec.ForDecl(&Counter,
	spec.Rationale("taking the address adds nothing"),
)

// A literal names no declaration.
var _ = spec.ForDecl(42,
	spec.Rationale("this names nothing at all"),
)
