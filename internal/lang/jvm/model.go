package jvm

import (
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang"
)

// Model is the JVM frontend's view of a project.
//
// It implements only lang.Model and none of the capability interfaces, which is
// the accurate description of what this frontend can do today: it reads a
// requirement tree and the bindings into it, and infers nothing. A run over a
// JVM project therefore measures the backward direction and says out loud that
// it did not measure the forward one.
//
// Growing into ConstructInferrer is what Spring's annotations are for —
// @RestController, @Entity, @Repository are architectural roles declared in
// exactly the place a class file reader can see them — but claiming the
// capability before it exists would make an unmeasured direction read as a
// clean one.
type Model struct{ r *Reader }

var _ lang.Model = (*Model)(nil)

// NewModel wraps a reader as a frontend model.
func NewModel(r *Reader) *Model { return &Model{r: r} }

func (m *Model) Name() string { return "jvm" }

func (m *Model) Requirements(out *diag.Set) []*ir.Requirement { return m.r.ReadRequirements(out) }
func (m *Model) Bindings(out *diag.Set) []ir.Binding          { return m.r.ReadBindings(out) }
func (m *Model) Dialect() ir.Dialect                          { return Dialect{} }
