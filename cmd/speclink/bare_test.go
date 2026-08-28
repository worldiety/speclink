package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The second Go style is what turned "a style is selectable" from a claim into
// two things that exist. It shares a language and a reader with the first and
// agrees with it about almost nothing else: no framework to import, no event
// sourcing, layers in their own directories, and one term the other forbids.

func TestBareProjectVerifies(t *testing.T) {
	out, code := runVerify(t, "../../testdata/bare")
	if code != 0 {
		t.Fatalf("the bare fixture did not verify:\n%s", out)
	}
	if !strings.Contains(out, "9 constructs (100% bound)") {
		t.Errorf("forward coverage was not measured:\n%s", summary(out))
	}
	if !strings.Contains(out, "100% covered, 100% verified, 100% demonstrated") {
		t.Errorf("a direction was lost:\n%s", summary(out))
	}
}

// Adding a required field to a stored type is what K9-FIELD-ADDED-REQUIRED
// exists to catch, because records written before it lack the field.
//
// This was silently unreachable under go_bare_ddd1. The recogniser matched
// nago's repository constructors and nothing else, so the persisted set was
// empty, every K9 rule iterated over nothing, and the change came out
// "0 findings" with 100% in every column — a clean bill of health for something
// nobody looked at. What this profile offers instead is the element type of a
// repository port, and an adapter that keeps a shape of its own says so with
// StoredAs, which moves the promise onto the shape that is actually written.
func TestBareGuardsStoredShapes(t *testing.T) {
	dir := copyFixture(t, "../../testdata/bare")
	rewrite(t, dir, "app/sales/adapter/fs/quotes.go",
		"\tStatus string `json:\"status\"`",
		"\tStatus string `json:\"status\"`\n\tNote   string `json:\"note\"`")
	rewrite(t, dir, "app/sales/adapter/fs/quotes.annotation.go",
		`var _ = spec.ForField[QuoteStore]("Status", spec.Satisfies(dec.RDecQuoteState))`,
		"var _ = spec.ForField[QuoteStore](\"Status\", spec.Satisfies(dec.RDecQuoteState))\nvar _ = spec.ForField[QuoteStore](\"Note\", spec.Satisfies(dec.RDecQuoteState))")

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a field was added to a promised shape without a word:\n%s", out)
	}
	if !strings.Contains(out, "field Note was added to QuoteStore without being declared optional") {
		t.Errorf("expected K9-FIELD-ADDED-REQUIRED:\n%s", out)
	}
}

// The promise sits on what is written, not on what happens to be in memory.
//
// The adapter maps the domain model onto a struct of its own, so freezing the
// domain model would guard a shape no file ever held, and every rename in it
// would read as a change to stored data. StoredAs says which of the two is on
// disk; without it the promise stays on the domain type, which is stricter and
// therefore the safe default.
func TestBareFreezesTheStoredShapeNotTheDomainModel(t *testing.T) {
	lock, err := os.ReadFile(filepath.Join("..", "..", "testdata", "bare", "speclink.lock"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(lock), "adapter/fs.QuoteStore") {
		t.Errorf("the written shape is not promised:\n%s", lock)
	}
	if strings.Contains(string(lock), `"example.com/bare/app/sales.Quote"`) {
		t.Errorf("the domain model was promised although the adapter insulates it:\n%s", lock)
	}
}

// The frontend level line and the profile level line say different things, and
// a profile must not be told off twice for one gap. The JVM frontend reads no
// schemas at all, so its own line is the accurate one.
func TestUnmeasuredSchemasAreReportedOnce(t *testing.T) {
	out, _ := runVerify(t, "../../testdata/java")

	count := strings.Count(out, "not measured: schema evolution")
	if count != 1 {
		t.Errorf("schema evolution is reported %d times, expected once:\n%s", count, out)
	}
	if !strings.Contains(out, "this frontend reads no persisted shapes") {
		t.Errorf("the reason given is the profile's rather than the frontend's:\n%s", out)
	}
}

// Four roles rather than nago's eight, and the four that are missing are
// missing because the architecture has nothing for those words to name.
func TestBareRecognisesFewerRoles(t *testing.T) {
	out, code := runSpeclink(t, "inventory", "../../testdata/bare", "./...")
	if code != 0 {
		t.Fatalf("inventory failed:\n%s", out)
	}
	for _, want := range []string{"use case", "aggregate", "repository", "permission"} {
		if !strings.Contains(out, want) {
			t.Errorf("the role %q was not recognised:\n%s", want, out)
		}
	}
	// Command, event and projection are the vocabulary of event sourcing, and
	// an architecture storing current state has nothing they could name.
	for _, unwanted := range []string{"event", "projection", "command"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("the role %q was invented where nothing declares it:\n%s", unwanted, out)
		}
	}
}

// A hand written interface states nothing, so it is marked — and the mark is
// the one fact this architecture annotates rather than infers.
func TestPersistenceMarkIsRecognised(t *testing.T) {
	out, _ := runSpeclink(t, "inventory", "../../testdata/bare", "./...")
	if !strings.Contains(out, "NumberRegistry") {
		t.Errorf("a marked port was not recognised as storage:\n%s", out)
	}
}

// The same term in the architecture that can infer it is a second source for a
// fact that already has one.
func TestPersistenceMarkIsRefusedWhereItIsInferable(t *testing.T) {
	dir := copyFixture(t, "../../testdata/example")
	appendTo(t, dir, "app/sales/customer.annotation.go",
		"\nvar _ = spec.For[CustomerEntity](spec.Persistence())\n")

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a redundant term was accepted:\n%s", out)
	}
	if !strings.Contains(out, "spec.Persistence is not available") {
		t.Errorf("the term was not refused:\n%s", out)
	}
	// The refusal has to say what already states the fact, or the reader is
	// told to remove something without being told why it is unnecessary.
	if !strings.Contains(out, "data.Repository") {
		t.Errorf("the refusal does not say what already states it:\n%s", out)
	}
}

// The layering rules are what the second style adds, and each has to fail on
// its own case rather than on a compile error.
func TestLayeringRules(t *testing.T) {
	for _, tc := range []struct {
		name, file, from, to, want string
	}{
		{
			// The domain importing its own presentation is an import cycle
			// that Go rejects first, so the rule exists for the direction
			// between contexts — which compiles, and is the one that is easy
			// to write and hard to see.
			name: "a context imports another context's presentation",
			file: "app/billing/usecases.go",
			from: "package billing",
			to:   "package billing\n\nimport _ \"example.com/bare/app/sales/rest\"",
			want: "is a domain package but imports",
		},
		{
			name: "presentation imports an adapter",
			file: "app/sales/rest/routes.go",
			from: "\t\"example.com/bare/foundation/rest\"",
			to:   "\t_ \"example.com/bare/app/sales/adapter/fs\"\n\t\"example.com/bare/foundation/rest\"",
			want: "but imports the adapter",
		},
		{
			// Added rather than swapped, so that the caller under cmd still
			// compiles: the point is what the rule says about the parameter,
			// not what the Go compiler says about the call.
			name: "a handler takes the whole bundle",
			file: "app/sales/rest/routes.go",
			from: "func Submit(who rest.Authenticator, submit sales.SubmitQuote) http.HandlerFunc {",
			to:   "func Bundled(who rest.Authenticator, uc sales.UseCases) http.HandlerFunc {\n\treturn Submit(who, uc.SubmitQuote)\n}\n\nfunc Submit(who rest.Authenticator, submit sales.SubmitQuote) http.HandlerFunc {",
			want: "takes the whole UseCases bundle",
		},
		{
			name: "a presentation package is named after neither its layer nor its context",
			file: "app/sales/rest/routes.go",
			from: "package restsales",
			to:   "package handlers",
			want: "expected restsales",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := copyFixture(t, "../../testdata/bare")
			rewrite(t, dir, tc.file, tc.from, tc.to)

			out, _ := runVerify(t, dir)
			if !strings.Contains(out, tc.want) {
				t.Errorf("expected %q:\n%s", tc.want, out)
			}
		})
	}
}

// Two styles over one language, and the seam between them is the profile.
func TestTwoGoStylesCoexist(t *testing.T) {
	for _, tc := range []struct{ dir, profile string }{
		{"../../testdata/example", "go_nago_ddd1"},
		{"../../testdata/bare", "go_bare_ddd1"},
	} {
		data, err := os.ReadFile(filepath.Join(tc.dir, "speclink.json"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), tc.profile) {
			t.Errorf("%s does not name %s", tc.dir, tc.profile)
		}
		if out, code := runVerify(t, tc.dir); code != 0 {
			t.Errorf("%s did not verify:\n%s", tc.dir, out)
		}
	}
}
