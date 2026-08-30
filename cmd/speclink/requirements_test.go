package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRequirementsCommand checks the tree of the conformant fixture on its own.
//
// The summary counts normative requirements separately because that is the
// number a migration is steered by: everything else is an explicit, justified
// exemption from coverage.
func TestRequirementsCommand(t *testing.T) {
	t.Parallel()
	out, code := runSpeclink(t, "requirements", "../../testdata/example", "./requirements/...")
	if code != 0 {
		t.Fatalf("expected a clean tree, got exit %d:\n%s", code, out)
	}
	if !strings.Contains(out, "9 requirements (8 normative, 0 reviewed)") {
		t.Errorf("unexpected summary:\n%s", out)
	}
}

// TestReviewIsBoundToTheWording pins what a recorded review means.
//
// The people this is for never see the source. They read a requirement that
// was extracted for them and say whether it is right, and the record of that
// is the only thing standing behind the claim that anybody read it. So it is
// recorded rather than declared — a field in the .spec.go file would be
// written by the same model that wrote the requirement — and it is bound to
// the wording it was given for. Rewrite the text and the review is gone,
// because what was read is not what is there.
func TestReviewIsBoundToTheWording(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/example")

	out, code := runSpeclink(t, "freeze", dir, "-reviewer", "Frau Meier", "./...")
	if code != 0 {
		t.Fatalf("freeze failed with %d:\n%s", code, out)
	}
	if out, _ = runSpeclink(t, "requirements", dir, "./requirements/..."); !strings.Contains(out, "9 reviewed") {
		t.Fatalf("the review was not recorded:\n%s", out)
	}

	rewrite(t, dir, "requirements/fun/quote/R-QUOTE-SUBMIT.spec.go",
		"MUST be drawn", "MUST NOT be drawn")

	out, _ = runSpeclink(t, "requirements", dir, "./requirements/...")
	if !strings.Contains(out, "8 reviewed") {
		t.Errorf("a rewritten requirement kept its review:\n%s", out)
	}
}

// TestExportCarriesWhatAReviewerNeeds pins the read surface.
//
// The audience is a person working through extracted requirements in a browser.
// They need the requirement as data and, above all, the segment it came from,
// because judging whether an extraction is faithful means reading it next to
// its origin.
func TestExportCarriesWhatAReviewerNeeds(t *testing.T) {
	t.Parallel()
	out, code := runSpeclink(t, "requirements", "../../testdata/example", "-format", "json", "./requirements/...")
	if code != 0 {
		t.Fatalf("export failed with %d:\n%s", code, out)
	}

	var report struct {
		Version      int `json:"version"`
		Requirements []struct {
			ID         string   `json:"id"`
			Text       string   `json:"text"`
			Sources    []string `json:"sources"`
			Satisfiers []string `json:"satisfiers"`
			Reviewed   bool     `json:"reviewed"`
			File       string   `json:"file"`
		} `json:"requirements"`
		Sources []struct {
			Ref          string   `json:"ref"`
			Kind         string   `json:"kind"`
			Requirements []string `json:"requirements"`
			Informative  bool     `json:"informative"`
		} `json:"sources"`
	}
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatalf("export is not valid JSON: %v\n%s", err, out)
	}
	if report.Version != TreeVersion {
		t.Errorf("got version %d, want %d", report.Version, TreeVersion)
	}

	var submit int = -1
	for i, r := range report.Requirements {
		if r.ID == "R-QUOTE-SUBMIT" {
			submit = i
		}
		// A surface has no notion of the machine the run happened on.
		if filepath.IsAbs(r.File) {
			t.Errorf("%s carries an absolute path %q", r.ID, r.File)
		}
	}
	if submit < 0 {
		t.Fatalf("R-QUOTE-SUBMIT missing from the export:\n%s", out)
	}
	r := report.Requirements[submit]
	if r.Text == "" || len(r.Sources) == 0 {
		t.Errorf("R-QUOTE-SUBMIT lacks text or origin: %+v", r)
	}
	// This command reads no annotation on purpose, so it knows of no
	// implementation. That is the case the reviewer works in, and the name of
	// the Go type would tell them nothing anyway.
	if len(r.Satisfiers) != 0 {
		t.Errorf("requirements must not claim to know the implementation: %v", r.Satisfiers)
	}

	// Both source kinds have to be there, or a mockup is invisible to the very
	// surface that is supposed to let people edit it.
	kinds := map[string]bool{}
	for _, s := range report.Sources {
		kinds[s.Kind] = true
	}
	if !kinds["markdown"] || !kinds["image"] {
		t.Errorf("the export does not carry both source kinds: %v", kinds)
	}
}

// TestRequirementsReportsTreeDefects pins that the tree checks still run, and
// that nothing else does. The bad fixture violates whitelist and binding rules
// too; in this mode they must stay silent, because they are questions about
// code and this command asks about the tree.
func TestRequirementsReportsTreeDefects(t *testing.T) {
	t.Parallel()
	out, code := runSpeclink(t, "requirements", "../../testdata/bad", "./requirements/...")
	if code == 0 {
		t.Fatalf("expected findings, got a clean run:\n%s", out)
	}

	for _, want := range []string{
		"SPEC-V5-004", // decision without a rationale
		"SPEC-V5-020", // normative requirement without a source
		"SPEC-V5-023", // source document does not exist
		"SPEC-V5-030", // file name does not match the requirement ID
		"SPEC-V5-032", // ID prefix contradicts Kind
		"SPEC-V5-033", // directory contradicts Kind
		"SPEC-V5-035", // domain directory contradicts the ID prefix
	} {
		if n := strings.Count(out, "["+want+"]"); n != 1 {
			t.Errorf("expected %s exactly once, found %d times", want, n)
		}
	}

	// Coverage is not measured here. Asking whether code satisfies a tree that
	// is still being built produces one finding per construct and buries the
	// defects of the tree itself, which are what has to be fixed first.
	for _, unwanted := range []string{"SPEC-V6-", "SPEC-V1-", "SPEC-V3-"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("%s findings must not appear in the requirements mode:\n%s", unwanted, out)
		}
	}
}

// TestRequirementsSurvivesBrokenImplementation is the reason the command
// exists. A tree is built over weeks; during that time the implementation
// around it is regularly in pieces, and a mode that needs the whole module to
// compile could not be used when it is needed most.
func TestRequirementsSurvivesBrokenImplementation(t *testing.T) {
	t.Parallel()
	// On a copy, like every other test that breaks something. Writing the
	// broken file into the checked in fixture worked for as long as the suite
	// ran serially and became a race the moment it did not: two other tests
	// read testdata/bad, and for a few hundred milliseconds it does not
	// compile. It never failed, which is the worst way for a race to behave.
	dir := copyFixture(t, "../../testdata/bad")
	broken := filepath.Join(dir, "sales", "zz_broken_test_fixture.go")
	if err := os.WriteFile(broken, []byte("package sales\n\nfunc broken() int { return \"\" }\n"), 0o644); err != nil {
		t.Fatalf("write broken file: %v", err)
	}

	// verify must refuse: without a compilable build there is nothing
	// meaningful to say about annotations.
	out, code := runSpeclink(t, "verify", dir, "./...")
	if code == 0 || !strings.Contains(out, "the build is broken") {
		t.Errorf("verify should refuse on a broken build, got exit %d:\n%s", code, out)
	}

	// requirements must not, as long as the tree itself compiles.
	out, code = runSpeclink(t, "requirements", dir, "./requirements/...")
	if strings.Contains(out, "the build is broken") {
		t.Errorf("requirements must not depend on the implementation compiling:\n%s", out)
	}
	if code != 1 || !strings.Contains(out, "requirements (") {
		t.Errorf("expected the tree summary and its findings, got exit %d:\n%s", code, out)
	}
}

// TestRequirementsJSON guards the machine readable contract of the mode, which
// is what a migration script consumes.
func TestRequirementsJSON(t *testing.T) {
	t.Parallel()
	out, _ := runSpeclink(t, "requirements", "../../testdata/bad", "-format", "json", "./requirements/...")
	for _, want := range []string{`"version": 1`, `"code":`, `"what":`, `"why":`, `"how":`} {
		if !strings.Contains(out, want) {
			t.Errorf("JSON report is missing %s:\n%s", want, out)
		}
	}
}
