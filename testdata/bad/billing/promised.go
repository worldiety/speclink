package billing

import (
	"context"

	"go.wdy.de/nago/application/evs"
)

// Settled is recorded in speclink.lock with the discriminator
// bad.billing.settled.v1, a field stored as "amount" and a field stored as
// "ID". The tag was changed, the discriminator was changed and the field was
// dropped, so all three must be reported.
type Settled struct {
	Amount int `json:"betrag"`
}

func (Settled) Discriminator() evs.Discriminator { return "bad.billing.settled.v2" }

func (e Settled) Evolve(_ context.Context, a *Aggregate) error {
	a.Status = "settled"
	return nil
}

// Unrecorded is frozen by default and has never been promised, so it must ask
// for the decision rather than pass silently.
type Unrecorded struct {
	ID string
}

func (Unrecorded) Discriminator() evs.Discriminator { return "bad.billing.unrecorded.v1" }

func (e Unrecorded) Evolve(_ context.Context, a *Aggregate) error {
	a.Status = "unrecorded"
	return nil
}

// Demoted was promised and is now marked as a proposal again, which claims that
// nothing was committed to. Stored messages say otherwise.
type Demoted struct {
	ID string
}

func (Demoted) Discriminator() evs.Discriminator { return "bad.billing.demoted.v1" }

func (e Demoted) Evolve(_ context.Context, a *Aggregate) error {
	a.Status = "demoted"
	return nil
}
