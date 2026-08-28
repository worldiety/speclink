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
	out, code := runSpeclink(t, "requirements", "../../testdata/java", "-profile", "java_springboot_ddd1")
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

// A direction a frontend cannot measure does not come out weak — it does not
// come out. Staying silent would let "no answer" read as "clean", which is the
// failure this tool spends most of its rules preventing.
func TestUnmeasuredDirectionsAreDisclosed(t *testing.T) {
	out, _ := runSpeclink(t, "requirements", "../../testdata/java", "-profile", "java_springboot_ddd1")

	// The JVM frontend reads no persisted shapes, so schema evolution is a
	// question it never puts.
	if !strings.Contains(out, "not measured: schema evolution") {
		t.Errorf("a direction this frontend cannot measure was not disclosed:\n%s", out)
	}
	// It does infer constructs, so it must not claim that gap.
	if strings.Contains(out, "not measured: forward coverage") {
		t.Errorf("a direction this frontend does measure was reported as a gap:\n%s", out)
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
	out, code := runSpeclink(t, "requirements", "../../testdata/java", "-profile", "java_springboot_ddd1", "-format", "json")
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

// A project must say which combination it is written in, and speclink must not
// guess. Language could be worked out from a go.mod and framework from an
// import, but style cannot be worked out from anything — and guessing it
// wrongly reports dozens of findings about a convention the project never meant
// to follow, which teaches the reader that the tool is wrong rather than that
// the project is.
func TestMissingProfileIsRefusedWithAMenu(t *testing.T) {
	dir := t.TempDir()

	out, code := runSpeclink(t, "requirements", dir)
	if code == 0 {
		t.Fatalf("a project without a profile was accepted:\n%s", out)
	}
	if !strings.Contains(out, "no profile set") {
		t.Errorf("the refusal does not say what is missing:\n%s", out)
	}
	// A rejection that does not say what the alternatives are makes the reader
	// go looking for documentation.
	for _, want := range []string{"go_nago_ddd1", "java_springboot_ddd1"} {
		if !strings.Contains(out, want) {
			t.Errorf("the refusal does not offer %s:\n%s", want, out)
		}
	}
}

// A key with no effect under the chosen profile is a mistaken expectation,
// almost always a profile that changed without the configuration following.
// Ignoring it is the ordinary thing to do and the wrong one: the symptom is a
// setting that quietly does nothing.
func TestForeignConfigurationIsRefused(t *testing.T) {
	dir := copyFixture(t, "../../testdata/example")
	writeConfig(t, dir, `{"profile":"go_nago_ddd1","classRoots":["build/classes"]}`)

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a foreign key was accepted:\n%s", out)
	}
	if !strings.Contains(out, "classRoots") || !strings.Contains(out, "does not use") {
		t.Errorf("the refusal does not name the key:\n%s", out)
	}
	// And it must say what the profile does understand, or the reader is left
	// guessing which of their keys was the wrong one.
	if !strings.Contains(out, "contextRoot") {
		t.Errorf("the refusal does not list what the profile understands:\n%s", out)
	}
}

// The profile carries the conventions and the project states its deviations.
// Before profiles those conventions lived in a Default() that read as what
// speclink believed about every project rather than as one style's layout.
func TestProfileCarriesTheConventions(t *testing.T) {
	dir := copyFixture(t, "../../testdata/example")

	// The fixture follows the style, so stating nothing must be clean.
	if out, code := runVerify(t, dir); code != 0 {
		t.Fatalf("a project stating no deviations did not verify:\n%s", out)
	}

	// And a stated deviation has to reach the rules.
	writeConfig(t, dir, `{"profile":"go_nago_ddd1","cmdRoot":"tools"}`)
	out, _ := runVerify(t, dir)
	if !strings.Contains(out, "does not live under tools/") {
		t.Errorf("the deviation did not reach the rules:\n%s", out)
	}
}

// A profile whose style prescribes nothing says so, on the same grounds as the
// capability lines: a rule family that never ran must not read as one that came
// out clean.
func TestEmptyStyleIsDisclosed(t *testing.T) {
	out, _ := runSpeclink(t, "verify", "../../testdata/java")
	if !strings.Contains(out, "prescribes no rules yet") {
		t.Errorf("an unwritten style passed quietly:\n%s", out)
	}

	out, _ = runVerify(t, "../../testdata/example")
	if strings.Contains(out, "prescribes no rules") {
		t.Errorf("a style with rules claimed to have none:\n%s", out)
	}
}

// Forward coverage on a project that is not Go, which is the question the
// vanilla Go exercise could not have answered.
//
// The doubt after the first frontend was that inference only worked because
// nago's shapes happened to be legible: method sets, type aliases, a named func
// type with a particular first parameter. Spring declares its architecture in
// annotations, and annotations are exactly what a declaration level reader
// sees. Inference is a property of a framework that says what it is, not of the
// framework it was first written for.
func TestVerifyMeasuresForwardCoverageOnSpring(t *testing.T) {
	out, code := runSpeclink(t, "verify", "../../testdata/java", "-profile", "java_springboot_ddd1")
	if code != 0 {
		t.Fatalf("the Java fixture did not verify:\n%s", out)
	}
	if !strings.Contains(out, "5 constructs (100% bound)") {
		t.Errorf("forward coverage was not measured:\n%s", summary(out))
	}
	if !strings.Contains(out, "3 normative requirements (100% covered") {
		t.Errorf("backward coverage was not measured:\n%s", summary(out))
	}
	// And the verification directions, which come from an annotation in the
	// bytecode joined with the report the build wrote — no runtime code at all.
	if !strings.Contains(out, "100% verified, 100% demonstrated") {
		t.Errorf("verification was not measured:\n%s", summary(out))
	}
}

// A question that was never put must not be answered.
//
// Running the verification rules over an empty set reported every requirement
// as unverified, which is a different claim from "nobody asked" and the more
// damaging of the two: it fails a project for not doing something the tool
// never looked for. The figures behave the same way — an unmeasured direction
// is absent from the summary rather than shown as zero.
func TestUnmeasuredDirectionsAreNotReportedAsFailing(t *testing.T) {
	out, _ := runSpeclink(t, "verify", "../../testdata/java", "-profile", "java_springboot_ddd1")

	// The JVM frontend reads no persisted shapes, so no promise can have been
	// broken and none is reported. The rules for it are skipped rather than
	// run over an empty set: running them would report every recorded type as
	// removed, which is a different claim from "nobody looked".
	if strings.Contains(out, "K9-") || strings.Contains(out, "was promised but no longer exists") {
		t.Errorf("a frontend that reads no shapes reported an evolution finding:\n%s", out)
	}
	if !strings.Contains(out, "not measured: schema evolution") {
		t.Errorf("the gap was not disclosed:\n%s", out)
	}

	// The Go frontend measures everything, so it claims no gap at all.
	out, _ = runVerify(t, "../../testdata/example")
	if strings.Contains(out, "not measured") {
		t.Errorf("a frontend that measures everything claimed a gap:\n%s", out)
	}
}
