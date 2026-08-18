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

// TestProjectionNeedsRequirement guards the one coverage decision that is not
// obvious: a projection is aggregate crossing, so nothing covers it
// transitively and it must name the requirement it answers. A repository is
// structural and must not, or every context would carry noise.
func TestProjectionNeedsRequirement(t *testing.T) {
	if !ir.ConstructProjection.NeedsRequirement() {
		t.Error("a projection must be bound to a requirement")
	}
	if ir.ConstructRepository.NeedsRequirement() {
		t.Error("a repository must not demand a requirement of its own")
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
