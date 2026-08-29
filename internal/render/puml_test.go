package render

import (
	"strings"
	"testing"

	"github.com/worldiety/speclink/internal/ir"
)

// A process is a graph and PlantUML's activity syntax is block structured, so
// the renderer emits a state diagram instead. These pin the parts of that
// choice that would otherwise be easy to lose.

func sample() *ir.Process {
	return &ir.Process{
		ID:    "P-QUOTE",
		Title: "Angebot",
		Nodes: []ir.ProcessNode{
			{Kind: ir.NodeStart, ID: "entwurf", Label: "Angebot ist entworfen", Pos: ir.Position{Line: 1}},
			{Kind: ir.NodeActivity, ID: "abgeben", Ref: "example.com/m/sales.SubmitQuote", Pos: ir.Position{Line: 2}},
			{Kind: ir.NodeEmit, ID: "abgegeben", Ref: "example.com/m/sales.QuoteSubmitted", Pos: ir.Position{Line: 3}},
			{Kind: ir.NodeFork, ID: "auf", Pos: ir.Position{Line: 4}},
			{Kind: ir.NodeJoin, ID: "zu", Pos: ir.Position{Line: 5}},
			{Kind: ir.NodeChoice, ID: "pruefen", Pos: ir.Position{Line: 6}},
			{Kind: ir.NodeEnd, ID: "ja", Label: "freigegeben", Pos: ir.Position{Line: 7}},
			{Kind: ir.NodeEnd, ID: "nein", Label: "verworfen", Pos: ir.Position{Line: 8}},
		},
		Edges: []ir.ProcessEdge{
			{From: "entwurf", To: "abgeben"},
			{From: "pruefen", To: "ja", When: "angenommen"},
			{From: "pruefen", To: "nein", When: "abgelehnt"},
			// A jump backwards, which is the whole reason the model is a graph.
			{From: "pruefen", To: "abgeben", When: "nachzubessern"},
		},
	}
}

// Several outcomes are the reason a process records more than one end.
// Spelling every one of them [*] collapses them into a single circle, and
// approved and withdrawn are not one ending with two labels.
func TestEveryOutcomeGetsItsOwnTerminal(t *testing.T) {
	got := Process(sample())

	for _, want := range []string{"state ja <<end>>", "state nein <<end>>"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q:\n%s", want, got)
		}
	}
	// The start is the one endpoint PlantUML spells [*], and it appears only
	// on the left of an arrow.
	if !strings.Contains(got, "[*] --> abgeben") {
		t.Errorf("the start is not drawn as one:\n%s", got)
	}
	if strings.Contains(got, "--> [*]") {
		t.Errorf("an outcome was collapsed into the shared terminal:\n%s", got)
	}
}

// PlantUML draws a terminal circle without a name, so the outcome has to be
// written where a reader will see it.
func TestOutcomeNameRidesOnTheArrow(t *testing.T) {
	got := Process(sample())
	if !strings.Contains(got, "pruefen --> ja : angenommen — freigegeben") {
		t.Errorf("the outcome is not named on the arrow:\n%s", got)
	}
}

func TestGatewaysCarryTheirStereotype(t *testing.T) {
	got := Process(sample())
	for _, want := range []string{"state auf <<fork>>", "state zu <<join>>", "state pruefen <<choice>>"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q:\n%s", want, got)
		}
	}
}

// A cycle survives the rendering. It is the one thing an activity diagram could
// not have carried without recovering block structure from an arbitrary graph.
func TestABackwardsJumpIsDrawn(t *testing.T) {
	got := Process(sample())
	if !strings.Contains(got, "pruefen --> abgeben : nachzubessern") {
		t.Errorf("the loop was lost:\n%s", got)
	}
}

// Node names come from the model and go into a syntax that has its own idea of
// an identifier.
func TestIdentifiersAreMadeSafe(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"abgeben", "abgeben"},
		{"P-QUOTE-DECISION", "P_QUOTE_DECISION"},
		{"1st", "n1st"},
		{"", "n"},
	} {
		if got := ident(tc.in); got != tc.want {
			t.Errorf("ident(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// A quotation mark in a label would end the name it sits in.
func TestLabelsCannotBreakOutOfTheirQuotes(t *testing.T) {
	p := sample()
	p.Nodes[0].Label = `sagt "ja"` + "\nund mehr"
	got := Process(p)

	if strings.Contains(got, `"ja"`) {
		t.Errorf("a quotation mark survived into the output:\n%s", got)
	}
	for _, line := range strings.Split(got, "\n") {
		if strings.Contains(line, "und mehr") && !strings.Contains(line, "sagt") {
			t.Errorf("a newline split a label across two lines:\n%s", got)
		}
	}
}

// The context view answers where the system ends. A channel between two
// packages does not cross that boundary and belongs in the other view.
func TestContextLeavesOutWhatDoesNotCrossTheBoundary(t *testing.T) {
	topo := ir.Topology{
		Participants: []ir.Participant{
			{Kind: ir.ParticipantActor, ID: "vertrieb", Name: "Vertrieb"},
		},
		Packages: map[string]bool{"app/sales": true, "app/sales/rest": true},
		Channels: []ir.Channel{
			{From: "vertrieb", To: "app/sales/rest", Label: "Selbstbedienung", Protocol: "HTTP"},
			{From: "app/sales/rest", To: "app/sales", Label: "Aufruf", Protocol: "in-process"},
		},
	}

	got := Context(topo, "erp")
	if !strings.Contains(got, "vertrieb --> system : Selbstbedienung") {
		t.Errorf("the crossing channel is missing:\n%s", got)
	}
	if strings.Contains(got, "Aufruf") {
		t.Errorf("an internal channel appeared in the context view:\n%s", got)
	}
	// Everything inside is one box, because a reader asking where the system
	// ends does not want to know that it has a sales context.
	if strings.Count(got, "rectangle") != 1 {
		t.Errorf("the system was opened up in the context view:\n%s", got)
	}
}
