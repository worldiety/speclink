// Package person deliberately uses the generic CRUD factories, which are
// forbidden by K4-NO-GENERIC-CRUD.
package person

import (
	"go.wdy.de/nago/application/ent"
	cfgent "go.wdy.de/nago/application/ent/cfg"
)

// Perms is derived from a prefix at run time instead of being declared.
var Perms = ent.DeclarePermissions("my.person", "Person")

// UseCases are six use cases nobody wrote.
var UseCases = ent.NewUseCases(Perms)

// Setup wires a whole module in one call.
func Setup() error {
	_, err := cfgent.Enable("my.person", "Person")
	return err
}
