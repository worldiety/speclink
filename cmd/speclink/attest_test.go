package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Who wrote a declaration and who has read it is a record of what happened, not
// a claim the code makes about itself. The code could say it and it would be
// worthless: the same machine that writes the code writes the annotation.

// A specialised reviewer reads a few declarations at a time. Recording a whole
// run as read because somebody looked at one use case is the fastest way to
// make the figure meaningless.
func TestAttestWorksOnASingleDeclaration(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")

	if out, code := runSpeclink(t, "attest", dir, "-origin", "llm", "./..."); code != 0 {
		t.Fatalf("attest failed with %d:\n%s", code, out)
	}
	// By short name, because a reviewer working through one package should not
	// have to type an import path.
	if out, code := runSpeclink(t, "attest", dir, "-reviewer", "TS", "SubmitQuote"); code != 0 {
		t.Fatalf("attest failed with %d:\n%s", code, out)
	}

	lock, err := os.ReadFile(filepath.Join(dir, "speclink.lock"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(lock), `"reviewedBy": "TS"`) {
		t.Errorf("the review was not recorded:\n%s", lock)
	}

	out, _ := runVerify(t, dir)
	if !strings.Contains(out, "1 read by a person") {
		t.Errorf("exactly one declaration was read and the figure says otherwise:\n%s", summary(out))
	}
}

// The whole point of the record, and the one rule here.
//
// A review that outlived its subject is worse than no review: it is somebody's
// name attached to text they never saw.
func TestReviewDoesNotSurviveAChange(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")
	if out, code := runSpeclink(t, "attest", dir, "-reviewer", "TS", "SubmitQuote"); code != 0 {
		t.Fatalf("attest failed with %d:\n%s", code, out)
	}

	// The logic of a use case lives in its constructor, not in the named func
	// type. A fingerprint over the signature alone would survive this.
	rewrite(t, dir, "app/sales/uc_submit_quote.go",
		"\t\treturn number, quotes.Save(", "\t\t_ = number\n\t\treturn \"\", quotes.Save(")

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a reviewed declaration was rewritten and nothing was reported:\n%s", out)
	}
	if !strings.Contains(out, "SubmitQuote was read by TS and has changed since") {
		t.Errorf("expected K18-REVIEW-STALE:\n%s", out)
	}
}

// Silence must not be able to pass for handwork. A declaration nothing has said
// anything about is neither machine written nor read.
func TestUnattestedIsNotHumanWritten(t *testing.T) {
	t.Parallel()
	out, code := runVerify(t, "../../testdata/bare")
	if code != 0 {
		t.Fatalf("the bare fixture did not verify:\n%s", out)
	}
	if !strings.Contains(out, "(0 machine written, 0 read by a person)") {
		t.Errorf("an unattested project was credited with something:\n%s", summary(out))
	}
}

// There is deliberately no finding for unread code. In a project whose code a
// machine writes and a person samples, a build that stays red until everything
// has been read is a build that is never green.
func TestUnreadCodeIsAFigureAndNotAFinding(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")
	if out, code := runSpeclink(t, "attest", dir, "-origin", "llm", "./..."); code != 0 {
		t.Fatalf("attest failed with %d:\n%s", code, out)
	}

	out, code := runVerify(t, dir)
	if code != 0 {
		t.Fatalf("machine written code with nobody having read it must still verify:\n%s", out)
	}
	if !strings.Contains(out, "machine written, 0 read by a person") {
		t.Errorf("the gap is not reported as a figure:\n%s", summary(out))
	}
}

// Recording an origin for text that has since changed would create exactly the
// state K18 exists to catch, so it must not be possible to create it here.
func TestAttestRecordsTheTextAsItStands(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")
	if out, code := runSpeclink(t, "attest", dir, "-reviewer", "TS", "SubmitQuote"); code != 0 {
		t.Fatalf("attest failed with %d:\n%s", code, out)
	}
	rewrite(t, dir, "app/sales/uc_submit_quote.go",
		"\t\treturn number, quotes.Save(", "\t\t_ = number\n\t\treturn \"\", quotes.Save(")

	// Re-attesting takes the new text with it, and the finding goes away.
	if out, code := runSpeclink(t, "attest", dir, "-reviewer", "TS", "SubmitQuote"); code != 0 {
		t.Fatalf("attest failed with %d:\n%s", code, out)
	}
	out, code := runVerify(t, dir)
	if code != 0 {
		t.Fatalf("a re-read declaration still reported stale:\n%s", out)
	}
}

func TestAttestRefusesWhatItCannotRecord(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")

	if out, code := runSpeclink(t, "attest", dir); code == 0 {
		t.Errorf("attest with nothing to record was accepted:\n%s", out)
	}
	if out, code := runSpeclink(t, "attest", dir, "-origin", "intern"); code == 0 {
		t.Errorf("an unknown origin was accepted:\n%s", out)
	} else if !strings.Contains(out, "expected llm or human") {
		t.Errorf("the refusal does not name what is allowed:\n%s", out)
	}
	if out, code := runSpeclink(t, "attest", dir, "-reviewer", "TS", "NoSuchThing"); code == 0 {
		t.Errorf("a construct that does not exist was accepted:\n%s", out)
	}
}

// The chain an audit asks for, end to end: the obligation, the requirement, the
// lines that implement it, how much of them a run reached, and the test that
// demonstrated it. Every link was knowable and the last two were never shown.
func TestDocumentCarriesTheLineLevelChain(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")

	profile := filepath.Join(t.TempDir(), "cover.out")
	stream := filepath.Join(t.TempDir(), "tests.json")
	runGoTest(t, dir, profile, stream)

	if out, code := runSpeclink(t, "evidence", dir, "-in", stream, "-coverprofile", profile, "./..."); code != 0 {
		t.Fatalf("evidence failed with %d:\n%s", code, out)
	}

	out, _ := runSpeclink(t, "generate", dir, "./...")
	if !strings.Contains(out, "app/sales/uc_submit_quote.go:") {
		t.Errorf("the document does not say where the requirement is implemented:\n%s", out)
	}
	if !strings.Contains(out, "statements exercised") {
		t.Errorf("the document does not say how much of it a run reached:\n%s", out)
	}

	verified, _ := runVerify(t, dir)
	if !strings.Contains(verified, "% of statements exercised") {
		t.Errorf("the summary carries no coverage figure:\n%s", summary(verified))
	}
}

// A figure measured before a rewrite is not a figure about what is there now.
func TestCoverageDoesNotSurviveARewrite(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")

	profile := filepath.Join(t.TempDir(), "cover.out")
	stream := filepath.Join(t.TempDir(), "tests.json")
	runGoTest(t, dir, profile, stream)
	if out, code := runSpeclink(t, "evidence", dir, "-in", stream, "-coverprofile", profile, "./..."); code != 0 {
		t.Fatalf("evidence failed with %d:\n%s", code, out)
	}

	rewrite(t, dir, "app/sales/uc_submit_quote.go",
		"\t\treturn number, quotes.Save(", "\t\t_ = number\n\t\treturn \"\", quotes.Save(")

	out, _ := runSpeclink(t, "generate", dir, "./...")
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "sales.SubmitQuote`") && strings.Contains(line, "statements exercised") {
			t.Errorf("a coverage figure outlived the text it was measured on:\n%s", line)
		}
	}
}

// runGoTest runs the fixture's tests with a coverage profile.
func runGoTest(t *testing.T, dir, profile, stream string) {
	t.Helper()

	cmd := exec.Command("go", "test", "-json", "-coverprofile="+profile, "./...")
	cmd.Dir = dir
	// A failing test is not this test's business; the streams are what matter.
	body, _ := cmd.Output()
	if err := os.WriteFile(stream, body, 0o644); err != nil {
		t.Fatal(err)
	}
}
