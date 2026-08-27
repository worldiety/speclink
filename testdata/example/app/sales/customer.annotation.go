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
	spec.Waive("K14-REQ-UNVERIFIED", "The ruling is that a customer is stored as state rather than as facts, and the type having no event and a repository behind it is what discharges it. A test could only assert that the code compiles as written."),
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

// Added after the shape had been promised, so the messages stored until now do
// not carry it.
var _ = spec.ForField[CustomerEntity]("Notes",
	spec.Satisfies(customer.RCustomerMasterdata),
	spec.Optional(),
)

// The domain model is asked separately from the stored form, and the two are
// not a formality apart. Notes existed here and nowhere in CustomerEntity, so
// every note was dropped on save and read back empty — the mapping functions
// simply did not mention the field. Nothing else in the project could see that:
// both types compiled, both round tripped, and the loss only showed as a domain
// field that traced to nothing.
var _ = spec.ForField[Customer]("ID",
	spec.Satisfies(nfr.RNfrTraceability),
)

var _ = spec.ForField[Customer]("Name",
	spec.Satisfies(customer.RCustomerMasterdata),
)

var _ = spec.ForField[Customer]("Notes",
	spec.Satisfies(customer.RCustomerMasterdata),
)
