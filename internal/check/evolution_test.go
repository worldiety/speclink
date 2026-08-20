package check

import (
	"strings"
	"testing"

	"github.com/worldiety/speclink/internal/baseline"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

const evPkg = "example.com/m/sales"
const evName = evPkg + ".QuoteSubmitted"

// promised is the recorded shape every case below starts from.
func promised() *baseline.File {
	return &baseline.File{
		Version: baseline.Version,
		Types: map[string]baseline.Entry{
			evName: {
				Discriminator: "sales.quote.submitted.v1",
				Fields: []baseline.Field{
					{Name: "QuoteID", Wire: "quoteID", Shape: "string"},
					{Name: "Number", Wire: "number", Shape: "string"},
				},
			},
		},
	}
}

// current builds the shape as the source currently declares it.
func current(fields ...ir.SchemaField) []ir.SchemaType {
	return []ir.SchemaType{{
		Name:          evName,
		Package:       evPkg,
		Discriminator: "sales.quote.submitted.v1",
		Fields:        fields,
	}}
}

func field(name, wire, shape string) ir.SchemaField {
	return ir.SchemaField{Name: name, Wire: wire, Shape: shape}
}

var scope = map[string]bool{evPkg: true}

// TestEvolutionHolds pins what may change without breaking a promise. These are
// the cases a developer meets every day, and reporting any of them would make
// the guard something people switch off.
func TestEvolutionHolds(t *testing.T) {
	tests := []struct {
		name     string
		schema   []ir.SchemaType
		optional map[string]bool
	}{
		{
			name:   "unchanged",
			schema: current(field("QuoteID", "quoteID", "string"), field("Number", "number", "string")),
		},
		{
			// The Go name is not stored anywhere, so renaming it is free as
			// long as the json tag keeps the wire name.
			name:   "go field renamed, wire name kept",
			schema: current(field("QuoteID", "quoteID", "string"), field("QuoteNumber", "number", "string")),
		},
		{
			// Growing a persisted model is normal; saying that old messages
			// lack the new field is the whole obligation.
			name: "field added and declared optional",
			schema: current(
				field("QuoteID", "quoteID", "string"),
				field("Number", "number", "string"),
				field("Reason", "reason", "string"),
			),
			optional: map[string]bool{"Reason": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &diag.Set{}
			Evolution(tt.schema, freezeWith(tt.optional), promised(), scope, nil, out)
			if !out.Empty() {
				t.Errorf("expected no finding, got:\n%s", render(out))
			}
		})
	}
}

// freezeWith builds the resolved status for the event under test.
func freezeWith(optional map[string]bool) map[string]Freeze {
	if optional == nil {
		optional = map[string]bool{}
	}
	return map[string]Freeze{evName: {Type: evName, OptionalFields: optional}}
}

// TestEvolutionBreaks pins what must never pass. Each of these silently
// destroys the readability of data that was valid when it was written.
func TestEvolutionBreaks(t *testing.T) {
	tests := []struct {
		name   string
		schema []ir.SchemaType
		want   string
	}{
		{
			name: "discriminator changed",
			schema: []ir.SchemaType{{
				Name: evName, Package: evPkg, Discriminator: "sales.quote.submitted.v2",
				Fields: []ir.SchemaField{field("QuoteID", "quoteID", "string"), field("Number", "number", "string")},
			}},
			want: RuleDiscriminatorFrozen,
		},
		{
			name:   "field removed",
			schema: current(field("QuoteID", "quoteID", "string")),
			want:   RuleFieldRemoved,
		},
		{
			name:   "wire name changed",
			schema: current(field("QuoteID", "quoteID", "string"), field("Number", "nr", "string")),
			want:   RuleFieldRenamed,
		},
		{
			name:   "type gone",
			schema: nil,
			want:   RuleTypeRemoved,
		},
		{
			// The value written yesterday is a string; a reader expecting a
			// number either fails on it or coerces it into something wrong.
			name:   "shape changed",
			schema: current(field("QuoteID", "quoteID", "string"), field("Number", "number", "int")),
			want:   RuleFieldShape,
		},
		{
			name: "field added without saying it may be absent",
			schema: current(
				field("QuoteID", "quoteID", "string"),
				field("Number", "number", "string"),
				field("Reason", "reason", "string"),
			),
			want: RuleFieldAddedRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &diag.Set{}
			Evolution(tt.schema, nil, promised(), scope, nil, out)
			if !strings.Contains(render(out), tt.want) {
				t.Errorf("expected %s, got:\n%s", tt.want, render(out))
			}
		})
	}
}

// TestEvolutionIgnoresDrafts guards the purpose of the marker. A draft
// that was never recorded has promised nothing, so nothing about it can break.
func TestEvolutionIgnoresDrafts(t *testing.T) {
	status := map[string]Freeze{evName: {Type: evName, Draft: true}}
	empty := &baseline.File{Version: baseline.Version, Types: map[string]baseline.Entry{}}

	out := &diag.Set{}
	Evolution(current(field("Totally", "different", "int")), status, empty, scope, nil, out)
	if !out.Empty() {
		t.Errorf("a draft must not be held to a promise:\n%s", render(out))
	}
}

// TestEvolutionCannotUnpromise is the other half. Marking something a draft
// after it was recorded claims that nothing was committed to, which is untrue
// the moment the first message is stored. Allowing it would turn the baseline
// from a record into a suggestion.
func TestEvolutionCannotUnpromise(t *testing.T) {
	status := map[string]Freeze{evName: {Type: evName, Draft: true}}
	out := &diag.Set{}
	Evolution(current(field("QuoteID", "quoteID", "string")), status, promised(), scope, nil, out)

	if !strings.Contains(render(out), RuleDraftFrozen) {
		t.Errorf("expected %s, got:\n%s", RuleDraftFrozen, render(out))
	}
	// And it must not be mistaken for a deletion; the type is right there.
	if strings.Contains(render(out), RuleTypeRemoved) {
		t.Errorf("a demoted type is not a removed one:\n%s", render(out))
	}
}

// TestEvolutionMissingBaseline is the rule that drives adoption. Without it an
// empty baseline would stay empty and the guard would never engage.
func TestEvolutionMissingBaseline(t *testing.T) {
	empty := &baseline.File{Version: baseline.Version, Types: map[string]baseline.Entry{}}
	out := &diag.Set{}
	Evolution(current(field("QuoteID", "quoteID", "string")), nil, empty, scope, nil, out)

	if !strings.Contains(render(out), RuleBaselineMissing) {
		t.Errorf("an unrecorded frozen shape must ask for a decision:\n%s", render(out))
	}
}

// TestEvolutionScope guards against the false alarm that would discredit the
// whole check: a run over one directory must not call every type elsewhere
// removed.
func TestEvolutionScope(t *testing.T) {
	out := &diag.Set{}
	Evolution(nil, nil, promised(), map[string]bool{"example.com/m/billing": true}, nil, out)
	if !out.Empty() {
		t.Errorf("a package that was not loaded says nothing about its types:\n%s", render(out))
	}
}

func render(s *diag.Set) string {
	var b strings.Builder
	for _, f := range s.Findings() {
		b.WriteString(f.Rule + ": " + f.What + "\n")
	}
	return b.String()
}

// TestOptionalCannotBeRevoked pins the one property of optionality that is not
// symmetric. A field may become optional at any time, because that only widens
// what a reader must cope with. It can never stop being optional, because the
// messages that lack it are already written and no release can reach them.
func TestOptionalCannotBeRevoked(t *testing.T) {
	base := &baseline.File{
		Version: baseline.Version,
		Types: map[string]baseline.Entry{
			evName: {
				Discriminator: "sales.quote.submitted.v1",
				Fields: []baseline.Field{
					{Name: "QuoteID", Wire: "quoteID", Shape: "string"},
					{Name: "Reason", Wire: "reason", Shape: "string", Optional: true},
				},
			},
		},
	}
	schema := current(field("QuoteID", "quoteID", "string"), field("Reason", "reason", "string"))

	out := &diag.Set{}
	Evolution(schema, freezeWith(nil), base, scope, nil, out)
	if !strings.Contains(render(out), RuleOptionalRevoked) {
		t.Errorf("expected %s, got:\n%s", RuleOptionalRevoked, render(out))
	}

	// Still declared: nothing to report.
	out = &diag.Set{}
	Evolution(schema, freezeWith(map[string]bool{"Reason": true}), base, scope, nil, out)
	if !out.Empty() {
		t.Errorf("a field that keeps its optionality is fine:\n%s", render(out))
	}
}

// TestIntegerWidthIsNotAShapeChange guards a decision that is easy to get wrong
// in the other direction. On the wire an integer is a JSON number whatever its
// width, so the fingerprint records the class and a widening never reaches this
// comparison at all.
func TestIntegerWidthIsNotAShapeChange(t *testing.T) {
	base := &baseline.File{
		Version: baseline.Version,
		Types: map[string]baseline.Entry{
			evName: {
				Discriminator: "sales.quote.submitted.v1",
				Fields:        []baseline.Field{{Name: "Count", Wire: "count", Shape: "int"}},
			},
		},
	}
	out := &diag.Set{}
	Evolution(current(field("Count", "count", "int")), freezeWith(nil), base, scope, nil, out)
	if !out.Empty() {
		t.Errorf("an integer stays an integer:\n%s", render(out))
	}
}
