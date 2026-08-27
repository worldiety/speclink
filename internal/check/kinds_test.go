package check

import (
	"bytes"
	"testing"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// The rules must work on a vocabulary they have never seen.
//
// Use case, command, event, aggregate and projection are one framework's words
// for a domain driven, event sourced design. A project on a different framework
// has other roles and the same three questions to answer about each of them, so
// what belongs in the rules is the question and not the answer.
//
// This uses roles from a framework that does not exist, precisely so that
// nothing here can be passing because it recognised a name.
func TestRulesWorkOnAForeignVocabulary(t *testing.T) {
	var (
		endpoint = ir.NewConstructKind("endpoint", "an endpoint", ir.NeedsRequirement())
		entity   = ir.NewConstructKind("entity", "an entity", ir.IsDomainModel(), ir.EmbodiesStorageDecision())
		mapper   = ir.NewConstructKind("mapper", "a mapper")
	)

	constructs := []ir.Construct{
		{Kind: endpoint, Name: "m/api.GetInvoice", Package: "m/api", Evidence: "is annotated as a route", Pos: ir.Position{File: "api/get.go", Line: 3}},
		{Kind: entity, Name: "m/api.Invoice", Package: "m/api", Evidence: "is annotated as an entity", Pos: ir.Position{File: "api/model.go", Line: 5}},
		{Kind: mapper, Name: "m/api.InvoiceMapper", Package: "m/api", Evidence: "converts between the two", Pos: ir.Position{File: "api/map.go", Line: 7}},
	}

	out := &diag.Set{}
	got := CoverConstructs(constructs, nil, plainDialect{}, out)

	// Only the role that says it carries business meaning is asked for one.
	if got.Required != 1 || len(got.Unbound) != 1 {
		t.Fatalf("got %d roles asked for a requirement and %d unbound, want 1 and 1 (the endpoint)", got.Required, len(got.Unbound))
	}
	if got.Unbound[0].Name != "m/api.GetInvoice" {
		t.Fatalf("the wrong role was asked: %s", got.Unbound[0].Name)
	}

	var buf bytes.Buffer
	if err := out.WriteText(&buf); err != nil {
		t.Fatal(err)
	}
	text := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("GetInvoice is an endpoint")) {
		t.Errorf("the finding does not speak the frontend's vocabulary:\n%s", text)
	}
	for _, unwanted := range []string{"Invoice is an entity", "InvoiceMapper"} {
		if bytes.Contains(buf.Bytes(), []byte(unwanted)) {
			t.Errorf("a structural role was asked for a requirement:\n%s", text)
		}
	}
}

// A kind nobody configured must not silently acquire an obligation, and must
// not render as an empty word either.
func TestZeroKindIsInertAndReadable(t *testing.T) {
	var k ir.ConstructKind

	if k.NeedsRequirement() || k.IsDomainModel() || k.EmbodiesStorageDecision() {
		t.Error("an undeclared role carries an obligation")
	}
	if k.String() != "unknown" || k.WithArticle() != "an unknown construct" {
		t.Errorf("an undeclared role renders as %q / %q", k.String(), k.WithArticle())
	}
}
