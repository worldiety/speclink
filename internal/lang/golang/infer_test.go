package golang

import (
	"path/filepath"
	"testing"

	"github.com/worldiety/speclink/internal/ir"
)

// TestInferKinds pins what the nago recognisers find in the conformant fixture.
//
// The summary line of the command only counts constructs, so a recogniser that
// returns the wrong kind would still add up. This asserts the kinds themselves,
// keyed by name, because that is what forward coverage and every diagnostic
// downstream depend on.
func TestInferKinds(t *testing.T) {
	root, err := filepath.Abs("../../../testdata/example")
	if err != nil {
		t.Fatal(err)
	}
	pkgs, err := Load(root, "./...")
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	got := map[string]ir.ConstructKind{}
	for _, p := range pkgs {
		for _, c := range p.Infer() {
			got[identOf(c.Name)] = c.Kind
		}
	}

	want := map[string]ir.ConstructKind{
		"SubmitQuote":       ir.ConstructUseCase,
		"ApproveQuote":      ir.ConstructUseCase,
		"FindQuoteOverview": ir.ConstructQuery,
		"SubmitQuoteCmd":    ir.ConstructCommand,
		"QuoteSubmitted":    ir.ConstructEvent,
		"QuoteAggregate":    ir.ConstructAggregate,
		// The read model is call driven: its type carries only Clone, so the
		// fact that it is a projection exists solely at construction.
		"QuoteOverview": ir.ConstructProjection,
		// The repository is a named type standing for the framework interface.
		"QuoteRepository": ir.ConstructRepository,
	}
	for name, kind := range want {
		if got[name] != kind {
			t.Errorf("%s: inferred %v, want %v", name, got[name], kind)
		}
	}
}

// TestCoverageObligations pins which construct kinds have to name a
// requirement. The set is a policy decision, and the one it replaces was never
// argued: a query fell out of it by omission, although every architecture rule
// already treats a query as a use case. Reading is a promise too.
func TestCoverageObligations(t *testing.T) {
	must := []ir.ConstructKind{
		ir.ConstructUseCase,
		ir.ConstructQuery,
		ir.ConstructCommand,
		ir.ConstructEvent,
		ir.ConstructProjection,
	}
	for _, k := range must {
		if !k.NeedsRequirement() {
			t.Errorf("%v carries business meaning and must name a requirement", k)
		}
	}

	// Structural kinds are covered through the use case that guards, writes or
	// holds them; demanding a binding for each would only produce noise.
	mustNot := []ir.ConstructKind{
		ir.ConstructAggregate,
		ir.ConstructPermission,
		ir.ConstructRepository,
	}
	for _, k := range mustNot {
		if k.NeedsRequirement() {
			t.Errorf("%v is structural and must not demand a requirement of its own", k)
		}
	}
}

// identOf reduces a fully qualified construct name to its identifier.
func identOf(qualified string) string {
	for i := len(qualified) - 1; i >= 0; i-- {
		if qualified[i] == '.' {
			return qualified[i+1:]
		}
	}
	return qualified
}
