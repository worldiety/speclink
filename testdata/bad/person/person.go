// Package person deliberately uses the generic CRUD factories, which are
// forbidden by K4-NO-GENERIC-CRUD.
//
// The call sites are shaped exactly as the framework requires them, generic
// parameters and configurator included. A simplified fixture would prove only
// that the rule matches a name, not that it matches real usage.
package person

import (
	"go.wdy.de/nago/application"
	"go.wdy.de/nago/application/ent"
	cfgent "go.wdy.de/nago/application/ent/cfg"
)

// ID identifies a person.
type ID string

// Person is a plain CRUD aggregate.
type Person struct {
	ID   ID
	Name string
}

// Identity makes Person an aggregate root.
func (p Person) Identity() ID { return p.ID }

// WithIdentity lets the generic create use case allocate an id.
func (p Person) WithIdentity(id ID) Person {
	p.ID = id
	return p
}

// Perms is derived from a prefix at run time instead of being declared.
var Perms = ent.DeclarePermissions[Person, ID]("my.person", "Person")

// UseCases are six use cases nobody wrote.
var UseCases = ent.NewUseCases[Person, ID](Perms, nil, ent.Options{})

// Setup wires a whole module in one call.
func Setup(cfg *application.Configurator) error {
	_, err := cfgent.Enable[Person, ID](cfg, "my.person", "Person", cfgent.Options[Person, ID]{})
	return err
}
