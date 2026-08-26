package billing

// Invoice carries an identity and is therefore an aggregate, and nothing
// anywhere says why an invoice is stored the way it is.
//
// The fixture needs it: without an aggregate in a bounded context the
// persistence rule has nothing to look at, and a rule that is never exercised
// by the negative fixture can rot without anyone noticing.
type Invoice struct {
	ID    string
	Total int
}

// Identity makes the invoice an aggregate root.
func (i Invoice) Identity() string { return i.ID }
