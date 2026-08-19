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
		for _, st := range p.ReadSchema(nil) {
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
		for _, st := range p.ReadSchema(nil) {
			for _, f := range st.Fields {
				// Amount is declared int; the record must say so as a class.
				if f.Name == "Amount" && f.Shape != "int" {
					t.Errorf("Amount shape = %q, want %q", f.Shape, "int")
				}
			}
		}
	}
}

// TestPersistedModelsSeparatesDomainFromStorage pins the distinction the
// framework offers and most projects get wrong.
//
// NewJSONRepository maps between two models, so only the persistence model is
// promised and the domain model stays free to be renamed or restructured.
// NewSloppyJSONRepository serialises the domain model directly; the framework's
// own documentation calls it a shorthand for throw-away prototypes. Choosing it
// makes every rename in the domain a change to stored data, and the check has
// to say so rather than quietly treat the two as equivalent.
func TestPersistedModelsSeparatesDomainFromStorage(t *testing.T) {
	tests := []struct {
		fixture  string
		pattern  string
		promised string
		free     string
	}{
		// Two models, mapped: the entity is stored, the domain type is not.
		{"../../../testdata/example", "./app/...", "CustomerEntity", "Customer"},
		// One model, serialised as it stands: the domain type is stored.
		{"../../../testdata/bad", "./billing/...", "Ledger", ""},
	}

	for _, tt := range tests {
		t.Run(tt.promised, func(t *testing.T) {
			root, err := filepath.Abs(tt.fixture)
			if err != nil {
				t.Fatal(err)
			}
			pkgs, err := Load(root, tt.pattern)
			if err != nil {
				t.Fatalf("load fixture: %v", err)
			}

			models := map[string]bool{}
			for _, p := range pkgs {
				for name := range p.PersistedModels() {
					models[name] = true
				}
			}

			if !hasSuffixIn(models, "."+tt.promised) {
				t.Errorf("%s is stored and must be part of the promised set, got %v", tt.promised, keysOf(models))
			}
			if tt.free != "" && hasSuffixIn(models, "."+tt.free) {
				t.Errorf("%s is only a domain model and must stay free of the promise", tt.free)
			}
		})
	}
}

func hasSuffixIn(set map[string]bool, suffix string) bool {
	for name := range set {
		if len(name) >= len(suffix) && name[len(name)-len(suffix):] == suffix {
			return true
		}
	}
	return false
}

func keysOf(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	return out
}
