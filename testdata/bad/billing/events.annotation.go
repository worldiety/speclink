package billing

import "github.com/worldiety/speclink/spec"

// The package is frozen; only this one type is still open.
var _ = spec.For[Invoiced](
	spec.Proposal(),
)

// Redundant: the type above already covers every one of its fields.
var _ = spec.ForField[Invoiced]("Amount",
	spec.Proposal(),
)
