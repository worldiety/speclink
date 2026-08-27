package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A claim is not evidence, and these pin the difference.
//
// spec.Verified is read from the source, which is what makes a missing one
// reportable, and writes a line when it runs, which is what makes a present one
// believable. Three quite different mistakes reduce to the same finding, and
// none of them is visible in the source: the call can sit behind a condition
// that never holds, the test can fail before reaching it, or the requirement
// can be rewritten afterwards so that the evidence was for other words.

func TestClaimWithoutARunIsReported(t *testing.T) {
	for _, tc := range []struct {
		name   string
		break_ func(t *testing.T, dir string)
		rerun  bool
	}{
		{
			// Written, compiles, never executes. No amount of reading the
			// source distinguishes this from a working test.
			name: "the call is never reached",
			break_: func(t *testing.T, dir string) {
				rewrite(t, dir, "app/sales/sales_test.go",
					"\tspec.Verified(t, quote.RQuoteChannel)",
					"\tif false {\n\t\tspec.Verified(t, quote.RQuoteChannel)\n\t}")
			},
			rerun: true,
		},
		{
			// The line is written and then the test fails. Only passing tests
			// are recorded, so the claim stands and the evidence does not.
			name: "the test fails before the end",
			break_: func(t *testing.T, dir string) {
				rewrite(t, dir, "app/sales/sales_test.go",
					"\tspec.Verified(t, quote.RQuoteChannel)",
					"\tt.Fatal(\"deliberately failing\")\n\tspec.Verified(t, quote.RQuoteChannel)")
			},
			rerun: true,
		},
		{
			// Nothing about the test changed. The sentence it was written
			// against did, so what it demonstrated is no longer what is being
			// asked.
			name: "the requirement was rewritten after the run",
			break_: func(t *testing.T, dir string) {
				rewrite(t, dir, "requirements/fun/quote/R-QUOTE-CHANNEL.spec.go",
					"MUST be recorded with the submission",
					"MUST NOT be recorded with the submission")
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := copyFixture(t, "../../testdata/example")
			tc.break_(t, dir)
			if tc.rerun {
				recordEvidence(t, dir)
			}

			out, code := runVerify(t, dir)
			if code == 0 {
				t.Fatalf("expected a finding:\n%s", out)
			}
			if !strings.Contains(out, "claims R-QUOTE-CHANNEL, but no run has shown it") {
				t.Errorf("expected K14-VERIFICATION-STALE:\n%s", out)
			}
			// The claim is still in the source, so the first figure must not
			// move. The gap between the two is the whole point of having both.
			if !strings.Contains(out, "100% verified, 88% demonstrated") {
				t.Errorf("expected the two figures to disagree:\n%s", summary(out))
			}
		})
	}
}

// Withdrawing a demonstration is a change like any other. Reporting only
// additions was a real bug: with nothing new to record the command returned
// early and never wrote the file, so evidence for a test that had just stopped
// passing stayed in the baseline and the run went green.
func TestEvidenceWithdrawsWhatNoLongerPasses(t *testing.T) {
	dir := copyFixture(t, "../../testdata/example")
	rewrite(t, dir, "app/sales/sales_test.go",
		"\tspec.Verified(t, quote.RQuoteChannel)",
		"\tt.Fatal(\"deliberately failing\")\n\tspec.Verified(t, quote.RQuoteChannel)")

	out := recordEvidence(t, dir)
	if !strings.Contains(out, "withdrawn R-QUOTE-CHANNEL") {
		t.Fatalf("the stale evidence was not withdrawn:\n%s", out)
	}

	lock, err := os.ReadFile(filepath.Join(dir, "speclink.lock"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(lock), "TestSubmissionRecordsTheChannel") {
		t.Error("the withdrawal was reported but not written")
	}
}

// A test built against a spec version this binary does not read must stop the
// command. Recording nothing would look exactly like a test that was never
// written, which is the failure this whole mechanism exists to make impossible.
func TestEvidenceRefusesAnUnknownVersion(t *testing.T) {
	dir := copyFixture(t, "../../testdata/example")

	stream := filepath.Join(dir, "tests.json")
	line := `{"Action":"output","Test":"TestX","Output":"    x_test.go:1: speclink-verified:{\"v\":99,\"reqs\":[\"R-QUOTE-SUBMIT\"]}\n"}` + "\n"
	if err := os.WriteFile(stream, []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}

	out, code := runSpeclink(t, "evidence", dir, "-in", stream, "./...")
	if code == 0 {
		t.Fatalf("an unknown version was accepted:\n%s", out)
	}
	if !strings.Contains(out, "version 99") {
		t.Errorf("the refusal must name the version it found:\n%s", out)
	}
}

// Piping something that is not a test stream is a mistake worth naming, not an
// empty run that silently withdraws every demonstration in the project.
func TestEvidenceRefusesAnEmptyStream(t *testing.T) {
	dir := copyFixture(t, "../../testdata/example")

	stream := filepath.Join(dir, "tests.json")
	if err := os.WriteFile(stream, []byte("not a test stream\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, code := runSpeclink(t, "evidence", dir, "-in", stream, "./...")
	if code == 0 {
		t.Fatalf("an empty stream was accepted:\n%s", out)
	}
	if !strings.Contains(out, "go test -json") {
		t.Errorf("the refusal must say what to pipe in:\n%s", out)
	}
}

// recordEvidence runs the fixture's tests and hands the result to speclink,
// which is the build order this feature adds: compile, verify, test, record.
func recordEvidence(t *testing.T, dir string) string {
	t.Helper()

	run := exec.Command("go", "test", "-json", "./...")
	run.Dir = dir
	// A failing test is the normal case here, so its exit code is not an error.
	stream, _ := run.Output()

	path := filepath.Join(dir, "tests.json")
	if err := os.WriteFile(path, stream, 0o644); err != nil {
		t.Fatal(err)
	}

	out, code := runSpeclink(t, "evidence", dir, "-in", path, "./...")
	if code != 0 {
		t.Fatalf("evidence failed with %d:\n%s", code, out)
	}
	return out
}
