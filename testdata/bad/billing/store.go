package billing

import (
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/json"
)

// LedgerID identifies a ledger entry.
type LedgerID string

// Ledger is stored through the sloppy repository, which serialises the domain
// model directly. That makes the domain model the wire format, so it is
// promised — and it was never recorded.
type Ledger struct {
	ID     LedgerID `json:"id"`
	Amount int      `json:"amount"`
}

// Identity makes the ledger entry addressable.
func (l Ledger) Identity() LedgerID { return l.ID }

// LedgerRepository stores ledger entries.
type LedgerRepository data.Repository[Ledger, LedgerID]

// NewLedgerRepository ties the domain model to stored data.
func NewLedgerRepository(store blob.Store) LedgerRepository {
	return json.NewSloppyJSONRepository[Ledger, LedgerID](store)
}
