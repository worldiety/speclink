package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// renameStoredField changes the wire name of a field the file store writes.
//
// A rename rather than a retype, because a retype does not compile and speclink
// correctly refuses to read a broken build — which makes it useless as a way of
// asking whether this rule fires.
func renameStoredField(t *testing.T, dir, from, to string) {
	t.Helper()
	path := filepath.Join(dir, "app", "sales", "adapter", "fs", "quotes.go")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	out := strings.Replace(string(src), `json:"`+from+`"`, `json:"`+to+`"`, 1)
	if out == string(src) {
		t.Fatalf("the fixture no longer has a field tagged %q", from)
	}
	if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestAContractThisSystemReliesOnIsHeld is the other half of the surface rules.
//
// K20 holds this system to the shapes it offers: a field removed from a
// response breaks a promise to somebody who parsed it. Nothing held it to the
// shapes it relies on receiving, and that is the sharper of the two. A promise
// this system makes is broken by people who are in this repository and can be
// told. A promise it relies on is broken by somebody who has never heard of
// this repository, and the first sign is a field arriving empty in production.
func TestAContractThisSystemReliesOnIsHeld(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")

	out, code := runVerify(t, dir)
	if code != 0 {
		t.Fatalf("the fixture must start clean, got %d:\n%s", code, summary(out))
	}

	renameStoredField(t, dir, "number", "quoteNumber")

	out, code = runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a contract lost a field and the run passed:\n%s", summary(out))
	}
	// The text output prints the code; the rule name reaches a reader through
	// the JSON format and the README index.
	if !strings.Contains(out, "SPEC-V6-171") {
		t.Errorf("the loss of a relied-on field was not reported:\n%s", out)
	}
	// The message has to name the channel, because that is what a reader can
	// act on — the far end of it is the thing that has to be talked to.
	if !strings.Contains(out, "Angebotsablage") {
		t.Errorf("the finding does not say which way across is affected:\n%s", out)
	}
}

// TestDroppingTheContractIsItselfReported guards the easiest way to silence
// this rule.
//
// Deleting the Contract field makes the finding go away and changes nothing
// about what crosses the boundary. The shape is still arriving and this system
// still reads it; the only difference is that nothing compares it any more.
func TestDroppingTheContractIsItselfReported(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")

	path := filepath.Join(dir, "topology", "boundary.topology.go")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	out := strings.Replace(string(src), "Contract: fs.QuoteStore{},", "", 1)
	if out == string(src) {
		t.Fatal("the fixture no longer declares a contract")
	}
	// The import goes with it, or the build breaks for the wrong reason.
	out = strings.Replace(out, "\"example.com/bare/app/sales/adapter/fs\"\n", "", 1)
	if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
		t.Fatal(err)
	}

	msg, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a recorded contract was silently dropped:\n%s", summary(msg))
	}
	if !strings.Contains(msg, "SPEC-V6-170") {
		t.Errorf("dropping a contract was not reported:\n%s", msg)
	}
}

// TestAnUnrecordedContractIsNotAFinding keeps the rule from reporting the
// ordinary case.
//
// A contract nobody has recorded yet cannot have been broken. That is what
// freeze is for, and a rule that fired on a shape the first time it was
// declared would make declaring one a punishment.
func TestAnUnrecordedContractIsNotAFinding(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")

	// Remove the record and leave the declaration.
	lock := filepath.Join(dir, "speclink.lock")
	src, err := os.ReadFile(lock)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), "\"channels\"") {
		t.Fatal("the fixture lock records no channel contract")
	}
	if err := os.WriteFile(lock, []byte(stripChannels(string(src))), 0o644); err != nil {
		t.Fatal(err)
	}

	out, code := runVerify(t, dir)
	if code != 0 {
		t.Fatalf("an unrecorded contract was treated as a break, got %d:\n%s", code, summary(out))
	}
	if strings.Contains(out, "SPEC-V6-170") || strings.Contains(out, "SPEC-V6-171") ||
		strings.Contains(out, "SPEC-V6-172") {
		t.Errorf("a contract that was never promised was reported as broken:\n%s", out)
	}
}

// stripChannels removes the recorded channel contracts from a lock file.
func stripChannels(lock string) string {
	start := strings.Index(lock, "\t\"channels\": {")
	if start < 0 {
		return lock
	}
	end := strings.Index(lock[start:], "\n\t},\n")
	if end < 0 {
		return lock
	}
	return lock[:start] + lock[start+end+len("\n\t},\n"):]
}

// TestFreezeRecordsAContract is the step that makes the rest possible.
func TestFreezeRecordsAContract(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")

	lock := filepath.Join(dir, "speclink.lock")
	src, err := os.ReadFile(lock)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(lock, []byte(stripChannels(string(src))), 0o644); err != nil {
		t.Fatal(err)
	}

	out, code := runSpeclink(t, "freeze", dir, "./...")
	if code != 0 {
		t.Fatalf("freeze refused a clean fixture, got %d:\n%s", code, out)
	}
	after, err := os.ReadFile(lock)
	if err != nil {
		t.Fatal(err)
	}
	// The structure, not the type name: a rename is not a change to anything
	// the far end can see.
	if !strings.Contains(string(after), "id:string,number:string") {
		t.Errorf("the contract was not recorded:\n%s", after)
	}
}
