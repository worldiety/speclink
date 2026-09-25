package main

import (
	"os"
	"strings"
	"testing"
)

// TestCommandsDoNotKnowAFrontend guards the seam between the commands and the
// readers.
//
// A command asks a model for a capability; the profile decides which model it
// gets. The moment a command imports a frontend it starts answering the
// question for that one language — a type assertion to report skipped packages,
// a switch to find the test results — and the next frontend has to be added in
// the command as well as beside the others. Both of those happened, and both
// were taken out again.
func TestCommandsDoNotKnowAFrontend(t *testing.T) {
	banned := []string{
		`"github.com/worldiety/speclink/internal/lang/golang"`,
		`"github.com/worldiety/speclink/internal/lang/jvm"`,
		`"github.com/worldiety/speclink/internal/lang/rust`,
		`"github.com/worldiety/speclink/spec"`,
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		for _, b := range banned {
			if strings.Contains(string(data), b) {
				t.Errorf("%s imports %s; ask the model for a capability in internal/lang instead", name, b)
			}
		}
	}
}
