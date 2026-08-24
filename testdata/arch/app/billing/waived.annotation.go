package billing

import "github.com/worldiety/speclink/spec"

// K5-UC-AUTHZ is violated by both use cases of this fixture and waived for one
// of them, so the rule still fires once and exactly one finding is suppressed.
//
// The fixture exists to violate things, so this is not there to tidy it up. It
// pins that spec.Waive reaches the architecture rules at all — for a long time
// it did not, while their own guidance was telling people to reach for it.
var _ = spec.For[IssueInvoice](
	spec.Waive("K5-UC-AUTHZ", "Fixture: pins that a waiver reaches the architecture rules."),
)
