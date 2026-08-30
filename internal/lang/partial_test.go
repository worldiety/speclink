package lang

import (
	"testing"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// everything implements every capability, standing in for a frontend that can
// answer any question when it is given the whole module.
type everything struct{}

func (everything) Requirements(*diag.Set) []*ir.Requirement { return nil }
func (everything) Bindings(*diag.Set) []ir.Binding          { return nil }
func (everything) Dialect() ir.Dialect                      { return nil }
func (everything) Name() string                             { return "everything" }
func (everything) Constructs(*diag.Set) []ir.Construct      { return nil }
func (everything) Processes(*diag.Set) []*ir.Process        { return nil }
func (everything) Topics() []ir.Topic                       { return nil }
func (everything) Topology(*diag.Set) ir.Topology           { return ir.Topology{} }
func (everything) Endpoints() []ir.Endpoint                 { return nil }
func (everything) CheckSyntax(*diag.Set)                    {}

// TestOnlyTreeCannotAnswerForAModule is the guard that three defects were
// missing.
//
// Each of them was a module-wide question put to a partial view: an entry point
// reported as missing because cmd/ had not been read, a use case reported as
// absent because its package had not been, and a requirement tree reported as
// fully covered having never been loaded. In every case the model was willing
// to answer, because willingness was a property of the concrete type and the
// restriction lived in a comment.
//
// A partial model must not implement the capabilities that speak for a module.
// Not "must decline to answer" — must not have the method, so the call cannot
// be written.
func TestOnlyTreeCannotAnswerForAModule(t *testing.T) {
	t.Parallel()
	whole := Model(everything{})
	part := OnlyTree(whole)

	if c := Of(whole); !c.Constructs || !c.Processes || !c.Topology || !c.Endpoints {
		t.Fatalf("the stand-in does not implement everything: %+v", c)
	}

	for name, implemented := range map[string]bool{
		"constructs": Of(part).Constructs,
		"processes":  Of(part).Processes,
		"topics":     Of(part).Topics,
		"topology":   Of(part).Topology,
		"endpoints":  Of(part).Endpoints,
		"schemas":    Of(part).Schemas,
	} {
		if implemented {
			t.Errorf("a partial model still answers for %s of the whole module", name)
		}
	}
}

// TestOnlyTreeKeepsWhatIsLocal is the other half.
//
// Whether a requirement file obeys the closed subset it is written in is
// answered by that file. Nothing about the rest of the module changes the
// verdict, which is what separates it from the rules that were also once
// reachable from a partial load.
func TestOnlyTreeKeepsWhatIsLocal(t *testing.T) {
	t.Parallel()
	part := OnlyTree(everything{})

	if _, ok := part.(SyntaxChecker); !ok {
		t.Error("a partial model lost the one check that is genuinely local")
	}
	if part.Name() != "everything" {
		t.Errorf("the frontend must still identify itself, got %q", part.Name())
	}
}

// TestOnlyTreeDoesNotPromoteByEmbedding pins why the wrapper is written out.
//
// Embedding the model would promote whatever the underlying type happens to
// implement, which is exactly the leak this exists to close. The test is here
// because the shorter spelling is the one somebody will reach for later.
func TestOnlyTreeDoesNotPromoteByEmbedding(t *testing.T) {
	t.Parallel()
	if _, ok := OnlyTree(everything{}).(ArchitectureChecker); ok {
		t.Error("a partial model can be asked to enforce the architecture of a module")
	}
}
