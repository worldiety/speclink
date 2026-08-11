// Package evs is a stub of the nago event sourcing API.
package evs

import (
	"context"

	"go.wdy.de/nago/auth"
)

// Discriminator is the stable serialisation tag of an event.
type Discriminator string

// SeqID is the commit sequence of an append.
type SeqID int64

// Cmd is a command which also provides the Decide implementation.
type Cmd[Aggregate, SuperEvt any] interface {
	Decide(auth.Subject, Aggregate) ([]SuperEvt, error)
}

// Evt is an event which also provides the Evolve implementation and a stable
// discriminator.
type Evt[Aggregate any] interface {
	Evolve(context.Context, Aggregate) error
	Discriminator() Discriminator
}
