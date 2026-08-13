// Package user is a stub of the nago user API.
package user

import "go.wdy.de/nago/application/permission"

// PermissionDeniedErr is returned by use cases that refuse an operation
// without going through Audit.
var PermissionDeniedErr = errString("permission denied")

// Subject is an authenticated identity, actor or subject.
type Subject interface {
	permission.Auditable
	ID() string
	Valid() bool
	HasRole(id string) bool
	HasGroup(id string) bool
}

type errString string

func (e errString) Error() string { return string(e) }
