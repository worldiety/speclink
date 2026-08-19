package sales

import (
	"go.wdy.de/nago/pkg/blob"
	"go.wdy.de/nago/pkg/data"
	"go.wdy.de/nago/pkg/data/json"
)

// CustomerID identifies a customer.
type CustomerID string

// Customer is the domain model. It is deliberately not the stored one: the
// separation is what lets it be renamed, restructured or split without touching
// a single byte on disk.
type Customer struct {
	ID    CustomerID
	Name  string
	Notes []string
}

// Identity makes the customer an aggregate root.
func (c Customer) Identity() CustomerID { return c.ID }

// CustomerEntity is the stored form, and the only one that is promised.
type CustomerEntity struct {
	ID   CustomerID `json:"id"`
	Name string     `json:"name"`
}

// Identity makes the entity addressable in the store.
func (e CustomerEntity) Identity() CustomerID { return e.ID }

// CustomerRepository stores customers.
type CustomerRepository data.Repository[Customer, CustomerID]

// NewCustomerRepository maps between the two models, so the domain stays free
// to change while the stored shape holds still.
func NewCustomerRepository(store blob.Store) CustomerRepository {
	return json.NewJSONRepository[Customer, CustomerID, CustomerEntity, CustomerID](
		store,
		func(e CustomerEntity) (Customer, error) {
			return Customer{ID: e.ID, Name: e.Name}, nil
		},
		func(c Customer) (CustomerEntity, error) {
			return CustomerEntity{ID: c.ID, Name: c.Name}, nil
		},
	)
}
