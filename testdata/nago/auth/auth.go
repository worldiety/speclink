// Package auth is a stub of the nago auth API.
package auth

import "go.wdy.de/nago/application/permission"

// Subject is an authenticated identity, actor or subject.
type Subject interface {
	permission.Auditable
	ID() string
	Valid() bool
}
