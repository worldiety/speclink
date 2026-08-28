package baseline

import (
	"os"
	"path/filepath"
	"testing"
)

// The record is written by the tool and never edited by hand, so a version bump
// that stranded existing projects would be asking them to do the one thing the
// file forbids. Older known versions are upgraded on read; newer ones are
// refused, because guessing at a record that says more than this binary
// understands would report differences that are not there.
func TestOlderVersionsAreUpgraded(t *testing.T) {
	for _, version := range []string{"1", "2"} {
		dir := t.TempDir()
		write(t, dir, `{"version":`+version+`,"types":{}}`)

		f, err := Load(dir)
		if err != nil {
			t.Fatalf("version %s was refused: %v", version, err)
		}
		// The maps the older version did not know must be usable rather than
		// nil, or every caller has to check before writing.
		if f.Requirements == nil || f.Sources == nil || f.Verifications == nil {
			t.Errorf("version %s left a nil map: %+v", version, f)
		}
	}
}

func TestNewerVersionIsRefused(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, `{"version":99,"types":{}}`)

	if _, err := Load(dir); err == nil {
		t.Error("a record from a newer binary was accepted")
	}
}

// A save writes the current version. The bump for the verifications was missed
// once and the field shipped under the previous one — which is the failure the
// version exists to prevent, because an older binary would not have refused
// such a file, it would have parsed it, dropped the field it did not know, and
// written the result back.
func TestSaveWritesTheCurrentVersion(t *testing.T) {
	dir := t.TempDir()
	f := &File{}
	f.fill()
	f.Verifications["R-A"] = []Verification{{Test: "TestA", Text: "hash"}}

	if err := f.Save(dir); err != nil {
		t.Fatal(err)
	}
	reloaded, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Version != Version {
		t.Errorf("wrote version %d, want %d", reloaded.Version, Version)
	}
	if len(reloaded.Verifications["R-A"]) != 1 {
		t.Error("the verifications did not survive a round trip")
	}
}

// The hash covers the words and not how they were typed, so the same
// requirement expressed in two languages records identically. That is not a
// coincidence worth relying on by accident: it is what lets a record survive a
// project moving from one frontend to another.
func TestTextHashIsIndependentOfLanguage(t *testing.T) {
	const (
		text  = "On submitting an approved quote a sequential, duplicate free quote number MUST be drawn."
		title = "Quote number on submission"
	)
	// The same sentence as a Go source would hold it, wrapped, and as a Java
	// annotation would hold it, on one line.
	wrapped := "On submitting an approved quote a sequential,\n\tduplicate free quote number MUST be drawn."

	if HashText(text, title) != HashText(wrapped, title) {
		t.Error("the same requirement hashed differently depending on how it was wrapped")
	}
	if HashText(text, title) == HashText(text+" Always.", title) {
		t.Error("a changed sentence produced the same hash")
	}
}

func write(t *testing.T, dir, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
