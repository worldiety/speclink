package render

import (
	"strings"
	"testing"

	"github.com/worldiety/speclink/internal/ir"
)

// TestLanesAreDerivedNotDeclared is the argument for one model and two
// drawings.
//
// Nothing states who performs a step. Each activity names a use case, the use
// case lives in a package, and the package is the part of the system doing the
// work. Declaring the lane as well would be that fact written twice, and the
// copy is the one that goes stale when a use case moves.
func TestLanesAreDerivedNotDeclared(t *testing.T) {
	t.Parallel()

	p := &ir.Process{
		ID:    "P-X",
		Title: "Angebot bis Rechnung",
		Nodes: []ir.ProcessNode{
			{Kind: ir.NodeStart, ID: "los", Label: "Vertrieb beginnt", Actor: "example.com/x/topology.Vertrieb"},
			{Kind: ir.NodeActivity, ID: "abgeben", Ref: "example.com/x/app/sales.SubmitQuote", RefPackage: "example.com/x/app/sales"},
			{Kind: ir.NodeActivity, ID: "rechnen", Ref: "example.com/x/app/billing.DraftInvoice", RefPackage: "example.com/x/app/billing"},
			{Kind: ir.NodeEnd, ID: "fertig", Label: "fertig"},
		},
		Edges: []ir.ProcessEdge{
			{From: "los", To: "abgeben"},
			{From: "abgeben", To: "rechnen"},
			{From: "rechnen", To: "fertig"},
		},
	}
	parts := map[string]ir.Participant{
		"example.com/x/topology.Vertrieb": {Kind: ir.ParticipantActor, ID: "vertrieb", Name: "Vertrieb"},
	}

	got := Sequence(p, parts)

	// The bounded context and not the whole import path: app/sales/adapter/fs
	// and app/sales are one participant in a picture of who talks to whom.
	for _, want := range []string{
		`actor "Vertrieb" as vertrieb`,
		`participant "sales" as sales`,
		`participant "billing" as billing`,
		"sales -> billing: DraftInvoice",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the drawing is missing %q:\n%s", want, got)
		}
	}
}

// TestConcurrencyIsNotDrawnAsAnOrder keeps the two views honest about where
// they differ.
//
// Time runs downwards in a sequence, so every arrow is ordered — and two
// branches of a fork have no order between them. Stacking them would invent a
// sequence the model does not have.
func TestConcurrencyIsNotDrawnAsAnOrder(t *testing.T) {
	t.Parallel()

	p := &ir.Process{
		ID: "P-F",
		Nodes: []ir.ProcessNode{
			{Kind: ir.NodeStart, ID: "los"},
			{Kind: ir.NodeFork, ID: "teilen"},
			{Kind: ir.NodeActivity, ID: "a", Ref: "x/app/sales.A", RefPackage: "x/app/sales"},
			{Kind: ir.NodeActivity, ID: "b", Ref: "x/app/billing.B", RefPackage: "x/app/billing"},
		},
		Edges: []ir.ProcessEdge{
			{From: "los", To: "teilen"},
			{From: "teilen", To: "a"},
			{From: "teilen", To: "b"},
		},
	}

	got := Sequence(p, nil)
	if !strings.Contains(got, "par") {
		t.Errorf("a fork was not drawn as concurrent:\n%s", got)
	}

	// A choice is one branch or the other, which is a different claim.
	p.Nodes[1].Kind = ir.NodeChoice
	p.Edges[1].When = "angenommen"
	got = Sequence(p, nil)
	if !strings.Contains(got, "alt angenommen") {
		t.Errorf("a choice was not drawn as an alternative:\n%s", got)
	}
}

// TestANoteSurvivesTheDrawing covers what a reader must know at a step and
// cannot read off the graph.
func TestANoteSurvivesTheDrawing(t *testing.T) {
	t.Parallel()

	p := &ir.Process{
		ID: "P-N",
		Nodes: []ir.ProcessNode{
			{Kind: ir.NodeStart, ID: "los"},
			{Kind: ir.NodeActivity, ID: "a", Ref: "x/app/sales.A", RefPackage: "x/app/sales",
				Note: "Die Sperre liegt in der Control Plane, nicht im Objektspeicher."},
			{Kind: ir.NodeActivity, ID: "b", Ref: "x/app/billing.B", RefPackage: "x/app/billing"},
		},
		Edges: []ir.ProcessEdge{{From: "los", To: "a"}, {From: "a", To: "b"}},
	}

	got := Sequence(p, nil)
	if !strings.Contains(got, "note right: Die Sperre liegt in der Control Plane") {
		t.Errorf("the note did not reach the drawing:\n%s", got)
	}
}
