package sales

import "github.com/worldiety/speclink/spec"

// The package is a draft, so nothing in it is promised yet.
var _ = spec.ForPackage(
	spec.Draft(),
)

// Redundant: the package above already says this.
var _ = spec.For[Opened](
	spec.Draft(),
)

// Redundant for the same reason, one level further down. A field term only
// means something once the type itself is frozen.
var _ = spec.ForField[Closed]("Reason",
	spec.Draft(),
)
