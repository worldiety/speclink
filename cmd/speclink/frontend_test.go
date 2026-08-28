package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// The second frontend is what turned the language boundary from a claim into a
// measurement. These pin what it bought.

// The tree command reads a requirement tree out of compiled Java bytecode, and
// every rule that then runs over it — identity, the derivation graph, the
// layout, the source layer — is the same code that runs over Go.
func TestRequirementsRunsOnAJavaProject(t *testing.T) {
	out, code := runSpeclink(t, "requirements", "../../testdata/java", "-lang", "jvm")
	if code != 0 {
		t.Fatalf("a clean Java fixture did not verify:\n%s", out)
	}
	if !strings.Contains(out, "3 requirements (3 normative") {
		t.Errorf("the tree was not read:\n%s", summary(out))
	}
	// The source layer is language neutral and this is where that stops being
	// an assertion: the same segmentation, anchors and forward coverage over a
	// Markdown document cited from a Java annotation.
	if !strings.Contains(out, "2 source segments (100% accounted)") {
		t.Errorf("the source layer did not run on a Java project:\n%s", summary(out))
	}
}

// A frontend that infers no constructs does not measure forward coverage
// weakly — it does not measure it. Staying silent would let "no answer" read as
// "clean", which is the failure this tool spends most of its rules preventing.
func TestUnmeasuredDirectionsAreDisclosed(t *testing.T) {
	out, _ := runSpeclink(t, "requirements", "../../testdata/java", "-lang", "jvm")

	if !strings.Contains(out, "not measured: forward coverage") {
		t.Errorf("a direction this frontend cannot measure was not disclosed:\n%s", out)
	}

	// The Go frontend measures all of them, so it must say nothing.
	out, _ = runSpeclink(t, "requirements", "../../testdata/example", "./requirements/...")
	if strings.Contains(out, "not measured") {
		t.Errorf("a frontend that measures everything claimed a gap:\n%s", out)
	}
}

// The export is the read surface, and it must describe a Java project as
// faithfully as a Go one — a reviewer working in a browser cannot tell which
// language produced the tree and should not have to.
func TestExportWorksAcrossFrontends(t *testing.T) {
	out, code := runSpeclink(t, "requirements", "../../testdata/java", "-lang", "jvm", "-format", "json")
	if code != 0 {
		t.Fatalf("export failed:\n%s", out)
	}

	var report struct {
		Requirements []struct {
			ID          string   `json:"id"`
			Text        string   `json:"text"`
			Sources     []string `json:"sources"`
			DerivedFrom []string `json:"derivedFrom"`
			File        string   `json:"file"`
		} `json:"requirements"`
	}
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, out)
	}

	for _, r := range report.Requirements {
		if r.ID != "R-QUOTE-SUBMIT" {
			continue
		}
		if r.Text == "" || len(r.Sources) == 0 {
			t.Errorf("the requirement lost its text or origin: %+v", r)
		}
		// The derivation edge came from a Java class literal, resolved by the
		// Java compiler, and arrives here as a requirement ID like any other.
		if len(r.DerivedFrom) != 1 || r.DerivedFrom[0] != "R-DEC-NUMBERING" {
			t.Errorf("derivedFrom is %v", r.DerivedFrom)
		}
		// The position points at the .java file, not at the .class it was
		// actually read from.
		if !strings.HasSuffix(r.File, ".java") {
			t.Errorf("the position points at %q rather than the source", r.File)
		}
		return
	}
	t.Fatalf("R-QUOTE-SUBMIT is missing from the export:\n%s", out)
}

// A repository holding both must not have to be told which it is, except when
// it genuinely is both — which is the case here, and the reason the flag exists.
func TestFrontendIsDetected(t *testing.T) {
	if got := detectFrontend("../../testdata/java"); got != "jvm" {
		t.Errorf("a project with compiled classes and no go.mod was read as %q", got)
	}
	if got := detectFrontend("../../testdata/example"); got != "go" {
		t.Errorf("a Go module was read as %q", got)
	}
}
