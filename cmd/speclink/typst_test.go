package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestTypstDocumentCompiles is the end of the chain this tool exists for.
//
// Everything upstream — the requirement tree, what implements it, what a run
// demonstrated — has been checkable for a while, and the document that carries
// it to a person has not been. A Typst file that looks plausible and does not
// compile delivers nothing, and no assertion about the Go strings inside it
// would notice. So the fixture is rendered and actually built.
//
// Skipped rather than failed when Typst is absent: speclink writes and never
// runs, so Typst is a prerequisite of the environment that renders the
// document, not of the one that builds this tool.
func TestTypstDocumentCompiles(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("typst"); err != nil {
		t.Skip("typst is not installed; the emitted document cannot be checked here")
	}

	for _, fixture := range []string{"../../testdata/example", "../../testdata/bare", "../../testdata/java"} {
		t.Run(filepath.Base(fixture), func(t *testing.T) {
			t.Parallel()
			extra := []string{"-format", "typst", "-date", "2026-01-01"}
			if strings.HasSuffix(fixture, "java") {
				extra = append(extra, "-profile", "java_springboot_ddd1")
			}
			out, _ := runSpeclink(t, "generate", fixture, extra...)

			dir := t.TempDir()
			src := filepath.Join(dir, "spec.typ")
			if err := os.WriteFile(src, []byte(out), 0o644); err != nil {
				t.Fatal(err)
			}
			if msg, err := exec.Command("typst", "compile", src, filepath.Join(dir, "spec.pdf")).CombinedOutput(); err != nil {
				t.Fatalf("the generated document does not compile: %v\n%s", err, msg)
			}
		})
	}
}

// TestGenerateIsReproducible protects the property that makes the document
// reviewable at all.
//
// It is meant to be committed and read as a diff, so generating it twice from
// the same tree has to produce the same bytes. The date on the title page is
// passed in for exactly this reason — reading it from the clock would put a
// spurious change in front of every reviewer every morning, and after a week
// of that nobody reads the diff.
func TestGenerateIsReproducible(t *testing.T) {
	t.Parallel()
	for _, format := range []string{"markdown", "typst"} {
		first, _ := runSpeclink(t, "generate", "../../testdata/example", "-format", format)
		second, _ := runSpeclink(t, "generate", "../../testdata/example", "-format", format)
		if first != second {
			t.Errorf("%s output changed between two runs over the same tree", format)
		}
	}
}

// TestBothFormatsCarryTheSameFacts is the guarantee the document model was
// built for.
//
// Two backends is a second thing to maintain only if each decides what the
// document says. They do not: both are handed the same tree and neither can
// see the model that produced it, so a fact can be spelled differently but
// cannot be present in one and missing from the other. This checks the claim
// on the real fixture rather than trusting the design.
func TestBothFormatsCarryTheSameFacts(t *testing.T) {
	t.Parallel()
	md, _ := runSpeclink(t, "generate", "../../testdata/example", "-format", "markdown")
	typ, _ := runSpeclink(t, "generate", "../../testdata/example", "-format", "typst")

	// Requirement identifiers are the load bearing content: every one that
	// reaches one document must reach the other.
	for _, id := range []string{
		"R-CUSTOMER-MASTERDATA", "R-QUOTE-SUBMIT", "R-QUOTE-APPROVE",
		"R-QUOTE-OVERVIEW", "R-NFR-TRACEABILITY", "R-DEC-CUSTOMER-STATE",
	} {
		if !strings.Contains(md, id) {
			t.Errorf("%s is missing from the markdown", id)
		}
		if !strings.Contains(typ, id) {
			t.Errorf("%s is missing from the typst", id)
		}
	}
	// So are the chapter titles, which is how a reader navigates either one.
	for _, chapter := range []string{
		"Where it stands", "Gaps", "The material", "The boundary",
		"What answers from outside", "Courses of business",
		"Requirements", "Source documents",
	} {
		if !strings.Contains(md, chapter) {
			t.Errorf("chapter %q is missing from the markdown", chapter)
		}
		if !strings.Contains(typ, chapter) {
			t.Errorf("chapter %q is missing from the typst", chapter)
		}
	}
}
