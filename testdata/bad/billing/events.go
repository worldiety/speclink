// Package billing is frozen as a package, so the cascade starts at the type.
package billing

import (
	"context"

	"go.wdy.de/nago/application/evs"
)

// Aggregate is the fold target.
type Aggregate struct {
	Status string
}

// Invoiced is a draft type in a frozen package. Marking one of its fields
// again states nothing new, because the type already covers every field.
type Invoiced struct {
	ID     string
	Amount int
}

func (Invoiced) Discriminator() evs.Discriminator { return "bad.billing.invoiced.v1" }

func (e Invoiced) Evolve(_ context.Context, a *Aggregate) error {
	a.Status = "invoiced"
	return nil
}
