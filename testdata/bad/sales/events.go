package sales

import (
	"context"

	"go.wdy.de/nago/application/evs"
)

// Aggregate is the fold target of the events below.
type Aggregate struct {
	Status string
}

// Opened is an event whose package is already a draft, so marking the type
// again states nothing new.
type Opened struct {
	ID   string
	Note string
}

func (Opened) Discriminator() evs.Discriminator { return "bad.sales.opened.v1" }

func (e Opened) Evolve(_ context.Context, a *Aggregate) error {
	a.Status = "open"
	return nil
}

// Closed carries a field marked as a draft although the whole package is
// one, which is the case the cascade exists to catch.
type Closed struct {
	ID     string
	Reason string
}

func (Closed) Discriminator() evs.Discriminator { return "bad.sales.closed.v1" }

func (e Closed) Evolve(_ context.Context, a *Aggregate) error {
	a.Status = "closed"
	return nil
}
