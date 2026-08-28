// Package auth carries who is acting.
//
// The subject is a parameter of every use case rather than something fished out
// of a context, and that is deliberate. A use case that takes a subject cannot
// be called without deciding who is calling; one that takes only a context can,
// and the check then depends on a value somebody remembered to put there.
//
// The context travels with the subject instead, because a use case that reaches
// for a database or a clock needs one and the subject is already there.
package auth

import (
	"context"
	"errors"
	"slices"

	"example.com/bare/foundation/permission"
)

// ErrPermissionDenied is returned when the subject lacks the permission.
var ErrPermissionDenied = errors.New("permission denied")

// Subject is who is acting.
type Subject interface {
	permission.Auditable

	// ID identifies the actor for the audit trail.
	ID() string
	// Context is bound to whatever outlives the subject: a request, a command
	// invocation, or nothing at all for System.
	Context() context.Context
}

// WithContext returns the subject bound to a context, which is what a
// presentation does once per request.
func WithContext(s Subject, ctx context.Context) Subject { return bound{Subject: s, ctx: ctx} }

type bound struct {
	Subject
	ctx context.Context
}

func (b bound) Context() context.Context { return b.ctx }

// System passes every check.
//
// For wiring, and for a command line tool that runs as the machine rather than
// as a person. It is deliberately not the zero value: a subject that grants
// everything has to be asked for.
func System() Subject { return system{} }

type system struct{}

func (system) ID() string                       { return "system" }
func (system) Audit(permission.ID) error        { return nil }
func (system) HasPermission(permission.ID) bool { return true }
func (system) Context() context.Context         { return context.Background() }

// Anonymous passes nothing.
//
// It is what an unauthenticated request gets, so that a missing subject fails
// closed rather than panicking somewhere further in.
func Anonymous() Subject { return Static("anonymous") }

// Static returns a subject with a fixed set of permissions.
func Static(id string, granted ...permission.ID) Subject {
	return static{id: id, granted: granted}
}

type static struct {
	id      string
	granted []permission.ID
}

func (s static) ID() string               { return s.id }
func (s static) Context() context.Context { return context.Background() }

func (s static) HasPermission(p permission.ID) bool { return slices.Contains(s.granted, p) }

func (s static) Audit(p permission.ID) error {
	if s.HasPermission(p) {
		return nil
	}
	return ErrPermissionDenied
}
