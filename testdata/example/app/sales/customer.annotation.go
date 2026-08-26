package sales

import (
	"example.com/erp/requirements/dec"
	"example.com/erp/requirements/fun/customer"
	"example.com/erp/requirements/nfr"
	"github.com/worldiety/speclink/spec"
)

// The customer is kept as current state while the quote next to it is event
// sourced. Recording that per aggregate is the point: the choice belongs to
// the data, not to the directory it happens to sit in.
var _ = spec.For[Customer](
	spec.Satisfies(dec.RDecCustomerState),
)

// The stored form carries the decision as well. It is a separate type on
// purpose — the domain model may be restructured without touching a byte on
// disk — and the promise that its shape holds still is the other half of the
// same ruling.
var _ = spec.For[CustomerEntity](
	spec.Satisfies(dec.RDecCustomerState),
)

var _ = spec.ForDecl(NewCustomerRepository,
	spec.Satisfies(dec.RDecCustomerState),
	spec.Rationale("The mapping between the two models lives here, which is where the choice to store state rather than facts actually takes effect."),
)

var _ = spec.ForField[CustomerEntity]("ID",
	spec.Satisfies(nfr.RNfrTraceability),
)

var _ = spec.ForField[CustomerEntity]("Name",
	spec.Satisfies(customer.RCustomerMasterdata),
)
