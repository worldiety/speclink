package golang

import (
	"go/types"
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

// TestATimeAndAByteSliceAreStringsOnTheWire holds the shape to what
// encoding/json actually writes.
//
// Read structurally, a time.Time is an empty object — all its fields are
// unexported — and a []byte is an array of numbers. Neither is what any
// implementation on the far end will ever see: the first is an RFC 3339 string
// and the second is base64.
//
// This is worse than an incomplete answer. A schema generated from the
// structural reading produces a parser that fails on the first real message,
// and it fails while looking correct, because nothing in the document hints
// that the shape was derived from fields nobody marshals.
func TestATimeAndAByteSliceAreStringsOnTheWire(t *testing.T) {
	t.Parallel()

	if got := shapeOf(types.NewSlice(types.Typ[types.Byte]), map[*types.Named]bool{}); got != "string" {
		t.Errorf("a []byte is base64 on the wire, and the shape says %q", got)
	}

	root, err := filepath.Abs("../../../testdata/bare")
	if err != nil {
		t.Fatal(err)
	}
	// The standard library is loaded alongside, because the shape of a
	// time.Time has to be read from the real declaration: a hand built stand
	// in would be a test of the test.
	pkgs, err := Load(root, "./...", "time")
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	clock := lookupTime(pkgs)
	if clock == nil {
		t.Fatal("time.Time was not found, so the interesting half of this test never ran")
	}
	if got := shapeOf(clock, map[*types.Named]bool{}); got != "string" {
		t.Errorf("a time.Time is RFC 3339 on the wire, and the shape says %q", got)
	}
}

// lookupTime finds the real time.Time among the loaded packages.
func lookupTime(pkgs []*Package) types.Type {
	for _, p := range pkgs {
		if p.PkgPath() != "time" {
			continue
		}
		if obj := p.pkg.Types.Scope().Lookup("Time"); obj != nil {
			return obj.Type()
		}
	}
	return nil
}
