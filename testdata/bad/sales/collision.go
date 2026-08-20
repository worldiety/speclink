package sales

import (
	"context"

	"go.wdy.de/nago/application/evs"
)

// Invoiced carries the same tag as billing.Invoiced, in a different bounded
// context. Both write into one stream and read each other's messages.
//
// The framework cannot see this: its uniqueness check runs per aggregate
// handler, and these two belong to different ones.
type Invoiced struct {
	ID string
}

func (Invoiced) Discriminator() evs.Discriminator { return "bad.billing.invoiced.v1" }

func (e Invoiced) Evolve(_ context.Context, a *Aggregate) error {
	a.Status = "invoiced"
	return nil
}
