package ts

import (
	"io"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang"
)

// Model is the TypeScript frontend's view of a project.
//
// It reads the requirement tree, the bindings into it, the use cases, the
// routes and the claims the tests make, and it enforces the hex1 style. It
// reads no persisted shapes and says so rather than reporting zero of them.
type Model struct {
	P *Project
}

var (
	_ lang.Model                 = (*Model)(nil)
	_ lang.ConstructInferrer     = (*Model)(nil)
	_ lang.VerificationReader    = (*Model)(nil)
	_ lang.EvidenceReader        = (*Model)(nil)
	_ lang.SyntaxChecker         = (*Model)(nil)
	_ lang.ArchitectureChecker   = (*Model)(nil)
	_ lang.ArchitectureDescriber = (*Model)(nil)
	_ lang.EndpointReader        = (*Model)(nil)
	_ lang.ScopeReporter         = (*Model)(nil)
	_ lang.PackageGrapher        = (*Model)(nil)
)

func (m *Model) Name() string        { return "ts" }
func (m *Model) Dialect() ir.Dialect { return Dialect{} }

func (m *Model) Requirements(out *diag.Set) []*ir.Requirement { return m.P.Requirements(out) }
func (m *Model) Bindings(out *diag.Set) []ir.Binding          { return m.P.Bindings(out) }
func (m *Model) Constructs(out *diag.Set) []ir.Construct      { return m.P.Constructs() }
func (m *Model) Verifications(out *diag.Set) []ir.Binding     { return m.P.Verifications(out) }
func (m *Model) CheckSyntax(out *diag.Set)                    { m.P.CheckSyntax(out) }
func (m *Model) CheckArchitecture(out *diag.Set)              { m.P.CheckArchitecture(out) }
func (m *Model) DescribeArchitecture() ir.Architecture        { return m.P.DescribeArchitecture() }
func (m *Model) Endpoints() []ir.Endpoint                     { return m.P.Endpoints() }
func (m *Model) PackageGraph() ir.PackageGraph                { return m.P.PackageGraph() }

func (m *Model) Demonstrations(in io.Reader, reqs lang.RequirementIndex) (map[string][]string, error) {
	return m.P.Demonstrations(in, reqs)
}

// SkippedUnits is the number of directories with modules the scope left out.
func (m *Model) SkippedUnits() int {
	dirs := map[string]bool{}
	for _, mod := range m.P.Modules {
		if !m.P.Layout.InScope(mod.Dir) {
			dirs[mod.Dir] = true
		}
	}
	return len(dirs)
}
