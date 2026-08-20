package check

import (
	"testing"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// TestDraftCascade pins the resolution itself. The diagnostics are covered
// end to end by the fixtures; what is checked here is the answer the rest of
// the tool will ask for: given these terms, is this field promised?
func TestDraftCascade(t *testing.T) {
	const pkg = "example.com/m/sales"

	event := func(name string) ir.SchemaType {
		return ir.SchemaType{Name: pkg + "." + name, Package: pkg}
	}
	draft := func(target ir.Target) ir.Binding {
		return ir.Binding{Target: target, Assertions: []ir.Assertion{{Kind: ir.AssertDraft}}}
	}

	tests := []struct {
		name       string
		bindings   []ir.Binding
		wantFrozen map[string]bool // field -> promised?
		wantType   bool            // whole type still a draft?
	}{
		{
			// The default is the important one: saying nothing promises
			// everything. Forgetting the term must fail safe.
			name:       "unmarked is frozen",
			wantFrozen: map[string]bool{"ID": true, "Reason": true},
		},
		{
			name:       "package cascades to every field",
			bindings:   []ir.Binding{draft(ir.Target{Kind: ir.TargetPackage, Package: pkg})},
			wantType:   true,
			wantFrozen: map[string]bool{"ID": false, "Reason": false},
		},
		{
			name:       "type cascades to every field",
			bindings:   []ir.Binding{draft(ir.Target{Kind: ir.TargetType, Name: pkg + ".Opened"})},
			wantType:   true,
			wantFrozen: map[string]bool{"ID": false, "Reason": false},
		},
		{
			// The only level at which a field term means anything.
			name:       "a single field inside a frozen type",
			bindings:   []ir.Binding{draft(ir.Target{Kind: ir.TargetField, Name: pkg + ".Opened", Field: "Reason"})},
			wantFrozen: map[string]bool{"ID": true, "Reason": false},
		},
		{
			// A term on a neighbour must not leak sideways.
			name:       "another type stays frozen",
			bindings:   []ir.Binding{draft(ir.Target{Kind: ir.TargetType, Name: pkg + ".Closed"})},
			wantFrozen: map[string]bool{"ID": true, "Reason": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Drafts([]ir.SchemaType{event("Opened")}, tt.bindings, &diag.Set{})
			f, ok := got[pkg+".Opened"]
			if !ok {
				t.Fatal("the event carries no freeze status at all")
			}
			if f.Draft != tt.wantType {
				t.Errorf("type draft = %v, want %v", f.Draft, tt.wantType)
			}
			for field, want := range tt.wantFrozen {
				if f.Frozen(field) != want {
					t.Errorf("Frozen(%q) = %v, want %v", field, f.Frozen(field), want)
				}
			}
		})
	}
}

// TestDraftScopeIsTheSchema guards where the persisted set comes from.
//
// It is the schema, not the construct list. An event is persisted because of
// what it is, a persistence model because somewhere a repository was built over
// it; a use case has no shape on the wire at all. Deriving the set from the
// schema keeps that one decision in one place.
func TestDraftScopeIsTheSchema(t *testing.T) {
	got := Drafts([]ir.SchemaType{
		{Name: "m.Opened", Package: "m"},
		{Name: "m.PersonEntity", Package: "m"},
	}, nil, &diag.Set{})

	for _, name := range []string{"m.Opened", "m.PersonEntity"} {
		if _, ok := got[name]; !ok {
			t.Errorf("%s is in the schema and must carry a freeze status", name)
		}
	}
	if _, ok := got["m.SubmitQuote"]; ok {
		t.Error("a type outside the schema has no shape to promise")
	}
}
