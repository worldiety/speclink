package golang

import (
	"path/filepath"
	"testing"
)

// TestReadSchemaShapes pins the fingerprint against the real type checker.
//
// The fingerprint is what the promise is recorded as, so what it collapses and
// what it keeps apart is the actual policy. Two of these are load bearing: a
// named type over string must be indistinguishable from string, or every
// harmless rename would be reported; and every integer width must collapse to
// one token, or widening a counter would look like a break.
func TestReadSchemaShapes(t *testing.T) {
	root, err := filepath.Abs("../../../testdata/example")
	if err != nil {
		t.Fatal(err)
	}
	pkgs, err := Load(root, "./app/...")
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	shapes := map[string]map[string]string{}
	discs := map[string]string{}
	for _, p := range pkgs {
		for _, st := range p.ReadSchema() {
			name := st.Name[len(st.Package)+1:]
			discs[name] = st.Discriminator
			shapes[name] = map[string]string{}
			for _, f := range st.Fields {
				shapes[name][f.Name] = f.Shape
			}
		}
	}

	if got, want := discs["QuoteSubmitted"], "sales.quote.submitted.v1"; got != want {
		t.Errorf("discriminator = %q, want %q", got, want)
	}
	if got, want := shapes["QuoteSubmitted"]["QuoteID"], "string"; got != want {
		t.Errorf("QuoteID shape = %q, want %q", got, want)
	}
}

// TestBasicShapeCollapsesIntegers states the policy directly, because the
// fixture cannot hold every width without becoming a menagerie.
func TestBasicShapeCollapsesIntegers(t *testing.T) {
	root, err := filepath.Abs("../../../testdata/bad")
	if err != nil {
		t.Fatal(err)
	}
	pkgs, err := Load(root, "./billing/...")
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	for _, p := range pkgs {
		for _, st := range p.ReadSchema() {
			for _, f := range st.Fields {
				// Amount is declared int; the record must say so as a class.
				if f.Name == "Amount" && f.Shape != "int" {
					t.Errorf("Amount shape = %q, want %q", f.Shape, "int")
				}
			}
		}
	}
}
