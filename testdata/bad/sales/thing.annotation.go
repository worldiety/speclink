package sales

import (
	"example.com/bad/anforderungen/fun/quote"
	"github.com/worldiety/speclink/spec"
)

// A helper function has no business in an annotation file.
func helper() string { return "nope" }

// External may only be attached to a type, Help not to a field.
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
