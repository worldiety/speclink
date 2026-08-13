// Package auth is a stub of the nago auth API.
package auth

import "go.wdy.de/nago/application/user"

// Subject is an authenticated identity.
//
// An alias, exactly as in the framework: the resolved type therefore reports
// the user package, which any recogniser matching on the package path has to
// account for.
type Subject = user.Subject
