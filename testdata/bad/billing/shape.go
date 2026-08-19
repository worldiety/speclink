package billing

import (
	"context"

	"go.wdy.de/nago/application/evs"
)

// Reconciled is recorded with Total stored as a string and Note declared
// optional. Both were changed here, and Extra was added without saying that
// older messages lack it.
type Reconciled struct {
	Total int
	Note  string
	Extra string
}

func (Reconciled) Discriminator() evs.Discriminator { return "bad.billing.reconciled.v1" }

func (e Reconciled) Evolve(_ context.Context, a *Aggregate) error {
	a.Status = "reconciled"
	return nil
}
