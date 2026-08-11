// Package ent is a stub of the nago generic CRUD module.
//
// Everything here is forbidden by K4-NO-GENERIC-CRUD: these constructors
// produce permissions, use cases and routes at run time from a prefix, which a
// static analysis can only see by reimplementing framework internals.
package ent

import "go.wdy.de/nago/application/permission"

// Permissions is the generated permission set of a CRUD module.
type Permissions struct {
	Create     permission.ID
	FindByID   permission.ID
	Update     permission.ID
	DeleteByID permission.ID
	EntityName string
}

// DeclarePermissions derives six permission IDs from a prefix.
func DeclarePermissions(prefix permission.ID, entityName string) Permissions {
	return Permissions{EntityName: entityName}
}

// UseCases bundles the generated CRUD use cases.
type UseCases struct{}

// NewUseCases derives six use cases from a repository.
func NewUseCases(perms Permissions) UseCases { return UseCases{} }
