// Package cfgent is a stub of the nago CRUD module wiring.
package cfgent

import (
	"go.wdy.de/nago/application/ent"
	"go.wdy.de/nago/application/permission"
)

// Module is a wired CRUD module.
type Module struct {
	UseCases    ent.UseCases
	Permissions ent.Permissions
}

// Enable configures a whole CRUD module instance: six permissions, a
// repository and three routes, all created at run time from the prefix.
func Enable(prefix permission.ID, entityName string) (Module, error) {
	return Module{}, nil
}
