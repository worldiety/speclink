package jvm

import (
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang"
)

// Model is the JVM frontend's view of a project.
//
// It reads a requirement tree, the bindings into it, the architectural roles a
// Spring project declares, and the claims its tests make. It reads no persisted
// shapes, and says so rather than reporting zero of them: an unmeasured
// direction must not read as a clean one.
type Model struct{ r *Reader }

var (
	_ lang.Model              = (*Model)(nil)
	_ lang.ConstructInferrer  = (*Model)(nil)
	_ lang.VerificationReader = (*Model)(nil)
)

// NewModel wraps a reader as a frontend model.
func NewModel(r *Reader) *Model { return &Model{r: r} }

func (m *Model) Name() string { return "jvm" }

func (m *Model) Requirements(out *diag.Set) []*ir.Requirement { return m.r.ReadRequirements(out) }
func (m *Model) Bindings(out *diag.Set) []ir.Binding          { return m.r.ReadBindings(out) }
func (m *Model) Dialect() ir.Dialect                          { return Dialect{} }
