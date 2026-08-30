package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// speclink writes drawing sources and runs nothing. PlantUML is a prerequisite
// of the environment rather than a dependency of this program, which is what
// keeps a checkout with no Java able to run every rule and every test — this
// one included.

func TestDiagramsWritesSourcesAndRunsNothing(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "puml")
	if got, code := runSpeclink(t, "diagrams", "../../testdata/bare", "-out", out, "./..."); code != 0 {
		t.Fatalf("diagrams failed with %d:\n%s", code, got)
	}

	for _, name := range []string{"context.puml", "blocks.puml"} {
		body, err := os.ReadFile(filepath.Join(out, name))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !strings.HasPrefix(string(body), "@startuml") {
			t.Errorf("%s is not a PlantUML source:\n%s", name, body)
		}
	}

	// No pictures. Rendering is the caller's line in a Makefile, and a tool
	// that rendered would have to be installed with a Java runtime beside it.
	entries, err := os.ReadDir(out)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".puml") {
			t.Errorf("something other than a source was written: %s", e.Name())
		}
	}
}

// One process, one diagram, named after it.
func TestDiagramsWritesOnePerProcess(t *testing.T) {
	t.Parallel()
	out := filepath.Join(t.TempDir(), "puml")
	if got, code := runSpeclink(t, "diagrams", "../../testdata/example", "-out", out, "./..."); code != 0 {
		t.Fatalf("diagrams failed with %d:\n%s", code, got)
	}

	body, err := os.ReadFile(filepath.Join(out, "process-P-QUOTE-DECISION.puml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"state aufteilen <<fork>>",
		"state zusammen <<join>>",
		"state pruefen <<choice>>",
		// The jump backwards, which is why the model is a graph.
		"pruefen --> abgeben : nachzubessern",
		// Two outcomes, two terminals.
		"state freigegeben <<end>>",
		"state verworfen <<end>>",
	} {
		if !strings.Contains(string(body), want) {
			t.Errorf("expected %q in the process diagram:\n%s", want, body)
		}
	}
}

// Two runs over one model produce byte identical sources. That is what makes a
// diagram in a review diff readable at all, and it has to hold whatever the
// renderer does with them afterwards.
func TestDiagramsAreReproducible(t *testing.T) {
	t.Parallel()
	first, second := filepath.Join(t.TempDir(), "a"), filepath.Join(t.TempDir(), "b")
	for _, dir := range []string{first, second} {
		if got, code := runSpeclink(t, "diagrams", "../../testdata/bare", "-out", dir, "./..."); code != 0 {
			t.Fatalf("diagrams failed with %d:\n%s", code, got)
		}
	}

	for _, name := range []string{"context.puml", "blocks.puml"} {
		a, err := os.ReadFile(filepath.Join(first, name))
		if err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(filepath.Join(second, name))
		if err != nil {
			t.Fatal(err)
		}
		if string(a) != string(b) {
			t.Errorf("%s differs between two runs over one model:\n%s\n---\n%s", name, a, b)
		}
	}
}

// A project that has declared neither is told so, rather than being handed an
// empty directory to wonder about.
func TestDiagramsRefusesWhenThereIsNothingToDraw(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")
	if err := os.RemoveAll(filepath.Join(dir, "topology")); err != nil {
		t.Fatal(err)
	}

	out, code := runSpeclink(t, "diagrams", dir, "-out", filepath.Join(t.TempDir(), "puml"), "./...")
	if code == 0 {
		t.Fatalf("expected a refusal:\n%s", out)
	}
	if !strings.Contains(out, "nothing to draw") {
		t.Errorf("the refusal does not say why:\n%s", out)
	}
}
