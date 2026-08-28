package golang

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/worldiety/speclink/internal/config"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

func TestSnakeCase(t *testing.T) {
	tests := []struct{ in, want string }{
		// The framework's own file names, which the rule has to reproduce.
		{"Create", "create"},
		{"FindByID", "find_by_id"},
		{"ListPermissions", "list_permissions"},
		{"UpdatePermissions", "update_permissions"},
		{"FindMyRoles", "find_my_roles"},
		{"SubmitQuoteUC", "submit_quote_uc"},
		// Acronyms must not fall apart into single letters.
		{"IDToken", "id_token"},
		{"HTTPServer", "http_server"},
		{"ParseURL", "parse_url"},
		{"A", "a"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := snakeCase(tt.in); got != tt.want {
			t.Errorf("snakeCase(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestUseCaseFileName(t *testing.T) {
	if got, want := DDD1.UseCaseFile("FindByID"), "uc_find_by_id.go"; got != want {
		t.Errorf("DDD1.UseCaseFile(FindByID) = %q, want %q", got, want)
	}
}

// A style says how a thing is spelled, and that two projects spell it
// differently while following the same architecture is the observation the
// seam exists for.
//
// The reference ERP names its use case constructors with a UC suffix and lays
// out 45 files to match. Under a fixed convention that reads as 238 defects. It
// is not 238 defects — it is a second convention, and until there was somewhere
// to write one down there was no way to say so.
//
// No second style is registered here: a style prescribes libraries and a layout
// and is approved as a whole, so inventing one to prove a mechanism would be
// putting something into the tool that nobody reviewed. What is proved is that
// the rules ask rather than assume.
func TestRulesAskTheStyleRatherThanAssuming(t *testing.T) {
	suffix := Style{
		Name:          "test-only",
		UseCaseFile:   func(name string) string { return snakeCase(name) + ".go" },
		Constructor:   func(name string) string { return name + "UC" },
		PermissionVar: func(name string) string { return name + "Perm" },
	}

	for _, tc := range []struct {
		style            Style
		file, ctor, perm string
	}{
		{DDD1, "uc_submit_quote.go", "NewSubmitQuote", "PermSubmitQuote"},
		{suffix, "submit_quote.go", "SubmitQuoteUC", "SubmitQuotePerm"},
	} {
		t.Run(tc.style.Name, func(t *testing.T) {
			if got := tc.style.UseCaseFile("SubmitQuote"); got != tc.file {
				t.Errorf("file is %q, want %q", got, tc.file)
			}
			if got := tc.style.Constructor("SubmitQuote"); got != tc.ctor {
				t.Errorf("constructor is %q, want %q", got, tc.ctor)
			}
			if got := tc.style.PermissionVar("SubmitQuote"); got != tc.perm {
				t.Errorf("permission is %q, want %q", got, tc.perm)
			}
		})
	}
}

// And the seam has to reach the rules, not merely the functions.
//
// The reference project is clean under its own style. Under a different
// convention the same unchanged code reports the same use cases against
// different expectations — which is exactly the reference ERP's position, and
// exactly the thing that used to be indistinguishable from being wrong.
func TestTheSameCodeUnderTwoConventions(t *testing.T) {
	dir, err := filepath.Abs(filepath.Join("..", "..", "..", "testdata", "example"))
	if err != nil {
		t.Fatal(err)
	}
	pkgs, err := Load(dir, "./...")
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{ContextRoot: "app", CmdRoot: "cmd", InfraRoots: []string{"pkg", "foundation"}}

	clean := &diag.Set{}
	CheckUseCases(pkgs, cfg, dir, DDD1, ir.Waivers{}, clean)
	if clean.Len() != 0 {
		var buf bytes.Buffer
		_ = clean.WriteText(&buf)
		t.Fatalf("the reference project is not clean under its own style:\n%s", buf.String())
	}

	other := &diag.Set{}
	CheckUseCases(pkgs, cfg, dir, Style{
		Name:          "test-only",
		UseCaseFile:   func(n string) string { return snakeCase(n) + ".go" },
		Constructor:   func(n string) string { return n + "UC" },
		PermissionVar: func(n string) string { return n + "Perm" },
	}, ir.Waivers{}, other)

	if other.Len() == 0 {
		t.Fatal("a different convention changed nothing, so the rules are not asking")
	}

	var buf bytes.Buffer
	if err := other.WriteText(&buf); err != nil {
		t.Fatal(err)
	}
	// The finding has to name what the style expects, not what the previous one
	// did, or the reader is told to rename a file to the name it already has.
	if !bytes.Contains(buf.Bytes(), []byte("expected approve_quote.go")) {
		t.Errorf("the finding does not speak the style's convention:\n%s", buf.String())
	}
	if bytes.Contains(buf.Bytes(), []byte("expected uc_approve_quote.go")) {
		t.Errorf("the finding still speaks the other style's convention:\n%s", buf.String())
	}
}

// A style reaching into the language rather than into the rules is new, and it
// earns that because the alternative is worse.
//
// A framework that hands out a Repository type has already said which types are
// storage; marking them again would be stating a fact twice, which is the one
// thing this language forbids. A framework of hand written interfaces has said
// nothing, and without a mark there is no set of stored types at all. The same
// term is therefore redundant in one architecture and indispensable in the next.
func TestStyleDecidesWhichTermsExist(t *testing.T) {
	bare := Style{Name: "bare", Terms: map[ir.AssertionKind]bool{ir.AssertPersistence: true}}

	if DDD1.Admits(ir.AssertPersistence) {
		t.Error("a style whose framework states persistence still admits the term")
	}
	if !bare.Admits(ir.AssertPersistence) {
		t.Error("a style that cannot infer persistence does not admit the term")
	}
	// Everything a style does not decide about stays universal, so adding a
	// term does not become a change to every style.
	for _, k := range []ir.AssertionKind{ir.AssertSatisfies, ir.AssertWaive, ir.AssertDraft, ir.AssertVerified} {
		if !DDD1.Admits(k) || !bare.Admits(k) {
			t.Errorf("%v was refused by a style that never decided about it", k)
		}
	}
}
