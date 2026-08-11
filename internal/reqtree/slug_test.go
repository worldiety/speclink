package reqtree

import "testing"

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
