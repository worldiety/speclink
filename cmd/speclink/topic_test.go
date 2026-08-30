package main

import (
	"strings"
	"testing"
)

// A theme orders from inside; a standard imposes from outside. Both head a
// chapter, and keeping them apart is what lets one be asked "which of these
// does nothing answer" while the other is simply how the people who built this
// think about it.

// A chapter heading with nothing under it reads as a part of the system that
// was left out, rather than as a heading somebody stopped using.
func TestEmptyThemeIsReported(t *testing.T) {
	dir := copyFixture(t, "../../testdata/bare")
	rewrite(t, dir, "requirements/dec/R-DEC-QUOTE-STATE.spec.go",
		"\tTopics:     []spec.Topic{requirements.Ablage},\n", "")
	rewrite(t, dir, "requirements/dec/R-DEC-QUOTE-STATE.spec.go",
		"\t\"example.com/bare/requirements\"\n", "")
	rewrite(t, dir, "requirements/dec/R-DEC-INVOICE-STATE.spec.go",
		"\tTopics:     []spec.Topic{requirements.Ablage},\n", "")
	rewrite(t, dir, "requirements/dec/R-DEC-INVOICE-STATE.spec.go",
		"\t\"example.com/bare/requirements\"\n", "")

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a theme has nothing under it and nothing was reported:\n%s", out)
	}
	if !strings.Contains(out, `no requirement is filed under "T-ABLAGE"`) {
		t.Errorf("expected K19-TOPIC-UNUSED:\n%s", out)
	}
}

// Two themes under one ID make both the chapter and every diagnostic about it
// ambiguous. Unlike a misspelled reference, the compiler has nothing to say
// about this one.
func TestDuplicateThemeIsReported(t *testing.T) {
	dir := copyFixture(t, "../../testdata/bare")
	rewrite(t, dir, "requirements/topics.spec.go", `ID:          "T-ABLAGE",`, `ID:          "T-ZUGRIFF",`)

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("two themes claim one id and nothing was reported:\n%s", out)
	}
	if !strings.Contains(out, `topic "T-ZUGRIFF" is declared twice`) {
		t.Errorf("expected K19-TOPIC-DUPLICATE:\n%s", out)
	}
}

// The chapters, and the one that matters: what carries no theme is counted
// rather than dropped, because a requirement left out of every chapter reads
// as one that does not exist.
func TestDocumentCarriesTheThemes(t *testing.T) {
	out, _ := runSpeclink(t, "generate", "../../testdata/bare", "./...")

	for _, want := range []string{
		"## Themes",
		"### Zugriff und Berechtigung",
		"| `R-QUOTE-SUBMIT` | Quote number on submission |",
		"### Under no theme",
		"3 requirements are filed under no theme",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in the document:\n%s", want, out)
		}
	}
}

func TestNoThemesNoChapter(t *testing.T) {
	out, _ := runSpeclink(t, "generate", "../../testdata/example", "./...")
	if strings.Contains(out, "## Themes") {
		t.Errorf("a project without themes was given a themes chapter:\n%s", out)
	}
}
