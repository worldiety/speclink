package source

import (
	"testing"

	"github.com/worldiety/speclink/internal/ir"
)

func TestSlug(t *testing.T) {
	tests := []struct {
		heading string
		want    string
	}{
		// The example from docs/annotations.md §5.2.
		{"## 8.1 Angebot (Kopf)", "81-angebot-kopf"},
		{"# 8. Abgabe", "8-abgabe"},
		{"### Angebot: Datenmodell", "angebot-datenmodell"},
		{"Vier-Augen-Prinzip", "vier-augen-prinzip"},
		{"  spaced  out  ", "spaced-out"},
		{"snake_case heading", "snake-case-heading"},
		{"Trailing punctuation!", "trailing-punctuation"},
		{"§ 139 BGB", "139-bgb"},
		// Non-ASCII letters are kept rather than dropped, otherwise German or
		// Greek headings would collapse to nothing.
		{"Übersicht", "übersicht"},
		{"---", ""},
	}
	for _, tt := range tests {
		if got := Slug(tt.heading); got != tt.want {
			t.Errorf("Slug(%q) = %q, want %q", tt.heading, got, tt.want)
		}
	}
}

func TestHeadings(t *testing.T) {
	const md = "# Title\n" +
		"text\n" +
		"## 8.1 Angebot (Kopf)\n" +
		"```go\n" +
		"// # not a heading\n" +
		"```\n" +
		"### Nested\n" +
		"#no-space-is-not-a-heading\n"

	got := Headings(md)
	want := []string{"title", "81-angebot-kopf", "nested"}

	if len(got) != len(want) {
		t.Fatalf("Headings() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Headings()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestKindDirsAreUsableAsGoPackages guards a trap the convention walked into
// once already.
//
// The requirement tree is Go, so every directory level is an import path
// element. "con" is a reserved device name on Windows and the Go toolchain
// refuses it outright — a constraint could be written but never compiled, and
// the failure appears as a malformed import path rather than as anything about
// requirements. A prescribed directory nobody can build is not a convention.
func TestKindDirsAreUsableAsGoPackages(t *testing.T) {
	// The names Windows reserves, lower cased as they would appear in a path.
	reserved := map[string]bool{
		"con": true, "prn": true, "aux": true, "nul": true,
		"com1": true, "com2": true, "com3": true, "com4": true,
		"lpt1": true, "lpt2": true, "lpt3": true, "lpt4": true,
	}

	for _, k := range []ir.Kind{ir.Functional, ir.NonFunctional, ir.Constraint, ir.Decision} {
		dir := k.Dir()
		if dir == "" {
			t.Errorf("%v has no directory", k)
			continue
		}
		if reserved[dir] {
			t.Errorf("%v maps to %q, which Go refuses as an import path element", k, dir)
		}
	}
}
