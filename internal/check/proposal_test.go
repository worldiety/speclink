package check

import (
	"testing"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// TestProposalCascade pins the resolution itself. The diagnostics are covered
// end to end by the fixtures; what is checked here is the answer the rest of
// the tool will ask for: given these terms, is this field promised?
func TestProposalCascade(t *testing.T) {
	const pkg = "example.com/m/sales"

	event := func(name string) ir.Construct {
		return ir.Construct{Kind: ir.ConstructEvent, Name: pkg + "." + name, Package: pkg}
	}
	proposal := func(target ir.Target) ir.Binding {
		return ir.Binding{Target: target, Assertions: []ir.Assertion{{Kind: ir.AssertProposal}}}
	}

	tests := []struct {
		name       string
		bindings   []ir.Binding
		wantFrozen map[string]bool // field -> promised?
		wantType   bool            // whole type still a proposal?
	}{
		{
			// The default is the important one: saying nothing promises
			// everything. Forgetting the term must fail safe.
			name:       "unmarked is frozen",
			wantFrozen: map[string]bool{"ID": true, "Reason": true},
		},
		{
			name:       "package cascades to every field",
			bindings:   []ir.Binding{proposal(ir.Target{Kind: ir.TargetPackage, Package: pkg})},
			wantType:   true,
			wantFrozen: map[string]bool{"ID": false, "Reason": false},
		},
		{
			name:       "type cascades to every field",
			bindings:   []ir.Binding{proposal(ir.Target{Kind: ir.TargetType, Name: pkg + ".Opened"})},
			wantType:   true,
			wantFrozen: map[string]bool{"ID": false, "Reason": false},
		},
		{
			// The only level at which a field term means anything.
			name:       "a single field inside a frozen type",
			bindings:   []ir.Binding{proposal(ir.Target{Kind: ir.TargetField, Name: pkg + ".Opened", Field: "Reason"})},
			wantFrozen: map[string]bool{"ID": true, "Reason": false},
		},
		{
			// A term on a neighbour must not leak sideways.
			name:       "another type stays frozen",
			bindings:   []ir.Binding{proposal(ir.Target{Kind: ir.TargetType, Name: pkg + ".Closed"})},
			wantFrozen: map[string]bool{"ID": true, "Reason": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Proposals([]ir.Construct{event("Opened")}, tt.bindings, &diag.Set{})
			f, ok := got[pkg+".Opened"]
			if !ok {
				t.Fatal("the event carries no freeze status at all")
			}
			if f.Proposal != tt.wantType {
				t.Errorf("type proposal = %v, want %v", f.Proposal, tt.wantType)
			}
			for field, want := range tt.wantFrozen {
				if f.Frozen(field) != want {
					t.Errorf("Frozen(%q) = %v, want %v", field, f.Frozen(field), want)
				}
			}
		})
	}
}

// TestProposalOnlyForPersisted guards the scope. A use case has no shape on the
// wire, so nothing about it can be promised and it must not appear here.
func TestProposalOnlyForPersisted(t *testing.T) {
	constructs := []ir.Construct{
		{Kind: ir.ConstructEvent, Name: "m.Opened", Package: "m"},
		{Kind: ir.ConstructUseCase, Name: "m.SubmitQuote", Package: "m"},
		{Kind: ir.ConstructAggregate, Name: "m.Quote", Package: "m"},
		{Kind: ir.ConstructProjection, Name: "m.Overview", Package: "m"},
	}
	got := Proposals(constructs, nil, &diag.Set{})

	if _, ok := got["m.Opened"]; !ok {
		t.Error("an event is persisted and must carry a freeze status")
	}
	for _, name := range []string{"m.SubmitQuote", "m.Quote", "m.Overview"} {
		if _, ok := got[name]; ok {
			t.Errorf("%s has no shape on the wire and must not be frozen", name)
		}
	}
}
