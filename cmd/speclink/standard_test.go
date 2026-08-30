package main

import (
	"strings"
	"testing"
)

// An external standard is a source document whose segments are its clauses,
// which is what makes the audit chapter fall out of machinery that already
// existed rather than out of a second implementation written for auditors.

// A clause nothing answers is not covered and not excluded either. It is
// unaccounted for, which is the state an audit exists to find.
func TestUnansweredClauseIsReported(t *testing.T) {
	dir := copyFixture(t, "../../testdata/bare")
	rewrite(t, dir, "requirements/fun/quote/R-QUOTE-SUBMIT.spec.go",
		"\t\t{Doc: \"requirements/_sources/vorgaben.standard.json\", Anchor: \"IAM-01\"},\n", "")

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a clause is answered by nothing and nothing was reported:\n%s", out)
	}
	if !strings.Contains(out, "clause IAM-01 of requirements/_sources/vorgaben.standard.json became no requirement") {
		t.Errorf("expected K12-SOURCE-UNCOVERED in the words of a standard:\n%s", out)
	}
	// The way out is the catalogue's, not Markdown's.
	if !strings.Contains(out, `exclude the clause with "applicable": false`) {
		t.Errorf("the fix offered is not the one this document kind has:\n%s", out)
	}
}

// A standard is reissued and the wording of a clause moves. Everything derived
// from it has to be read again, and this is the only mechanism that says which.
func TestRewordedClauseNamesWhatDependsOnIt(t *testing.T) {
	dir := copyFixture(t, "../../testdata/bare")
	rewrite(t, dir, "requirements/_sources/vorgaben.standard.json",
		"Die Prüfung erfolgt vor jeder Wirkung.", "Die Prüfung erfolgt nach der Wirkung.")

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a clause was reworded and nothing was reported:\n%s", out)
	}
	if !strings.Contains(out, "clause IAM-01") || !strings.Contains(out, "changed since it was last recorded") {
		t.Errorf("expected K13-SOURCE-DRIFT on the clause:\n%s", out)
	}
	if !strings.Contains(out, "R-QUOTE-SUBMIT") {
		t.Errorf("the finding does not name what was derived from it:\n%s", out)
	}
}

// The two figures answer different questions and a project is routinely in a
// different place on each. A mean of them tells nobody either.
func TestStandardsAreCountedApartFromOwnMaterial(t *testing.T) {
	out, code := runVerify(t, "../../testdata/bare")
	if code != 0 {
		t.Fatalf("the bare fixture did not verify:\n%s", out)
	}

	// Four Markdown sections, not six: the two applicable clauses are counted
	// in their standard and nowhere else.
	if !strings.Contains(out, "4 source segments (100% accounted)") {
		t.Errorf("clauses leaked into the figure for the project's own material:\n%s", summary(out))
	}
	if !strings.Contains(out, "2 of 2 applicable clauses answered, 2 not applicable") {
		t.Errorf("the standard has no figure of its own:\n%s", out)
	}
}

// The audit chapter and the statement of applicability, which ISO 27001
// requires as a document in its own right and nobody maintains by hand for long.
func TestDocumentCarriesTheStandardAndItsApplicability(t *testing.T) {
	out, _ := runSpeclink(t, "generate", "../../testdata/bare", "./...")

	for _, want := range []string{
		"## Standards",
		"### Hausstandard für Fachanwendungen, Fassung 1",
		"| `IAM-01` | Jede fachliche Handlung wird gegen eine Berechtigung geprüft | R-QUOTE-SUBMIT |",
		"#### Statement of applicability",
		"| `PS-01` | Zutrittskontrolle zum Rechenzentrum | Die Anwendung stellt keinen Betrieb bereit.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in the document:\n%s", want, out)
		}
	}
}

// A project with no standard gets no chapter, rather than an empty heading.
func TestNoStandardNoChapter(t *testing.T) {
	out, _ := runSpeclink(t, "generate", "../../testdata/example", "./...")
	if strings.Contains(out, "## Standards") {
		t.Errorf("a project without a standard was given a standards chapter:\n%s", out)
	}
}
