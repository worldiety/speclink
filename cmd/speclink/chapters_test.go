package main

import (
	"strings"
	"testing"
)

// TestTheDocumentSaysWhatGetsBuilt is the chapter a reader needs before any
// other one makes sense.
//
// A module builds programs, and every other chapter speaks as though there
// were one system. Which binary a statement is about changes the answer to
// almost every question somebody arrives with.
func TestTheDocumentSaysWhatGetsBuilt(t *testing.T) {
	t.Parallel()
	out, _ := runSpeclink(t, "generate", "../../testdata/bare")

	if !strings.Contains(out, "## What gets built") {
		t.Fatalf("the document does not say what this module produces:\n%s", chaptersOf(out))
	}
	// Read from the language: the name and the directory.
	for _, want := range []string{"### erp", "`cmd/erp`"} {
		if !strings.Contains(out, want) {
			t.Errorf("the program entry is missing %q", want)
		}
	}
	// Read from the import graph: what it assembles and what it chooses. The
	// second is the one that matters for a deployment — it is the only place
	// the choice of technology is made.
	for _, want := range []string{"`sales`", "`billing`", "app/sales/adapter/fs"} {
		if !strings.Contains(out, want) {
			t.Errorf("the wiring of the program is missing %q", want)
		}
	}
}

// TestInferredInvocationIsMarkedAsInferred is the honesty this chapter turns on.
//
// Nothing in Go declares how a program is called, so the subcommands are read
// off comparisons against the argument vector. That covers the common shape
// and misses any program dispatching through a table. Printing the result
// beside facts read from the type system, with nothing to separate them, would
// let a guess be quoted as a contract.
func TestInferredInvocationIsMarkedAsInferred(t *testing.T) {
	t.Parallel()
	out, _ := runSpeclink(t, "generate", "../../testdata/bare")

	if !strings.Contains(out, "Subcommands: `serve`") {
		t.Errorf("the recognisable subcommand was not found:\n%s", out)
	}
	if !strings.Contains(out, "inferred from the code rather than declared") {
		t.Error("a guess was presented without saying it was one")
	}
}

// TestTheArchitectureIsStatedWhenNothingIsBroken is the gap this closes.
//
// Every sentence explaining an architectural rule used to be written at the
// moment the rule was violated, and named the file that violated it. So a
// project in good order — the normal case, and the one handed to a reviewer —
// described its own architecture nowhere. A developer joining it could learn
// the rules only by breaking one.
func TestTheArchitectureIsStatedWhenNothingIsBroken(t *testing.T) {
	t.Parallel()
	out, code := runVerify(t, "../../testdata/bare")
	if !strings.Contains(out, "0 findings") {
		t.Fatalf("this fixture is supposed to be clean, got %d:\n%s", code, summary(out))
	}

	out, _ = runSpeclink(t, "generate", "../../testdata/bare")
	if !strings.Contains(out, "## How it is put together") {
		t.Fatalf("a project with no findings described its architecture nowhere:\n%s", chaptersOf(out))
	}
	// The layers come from the project's own configuration, not the profile
	// defaults, so a project that moved its contexts is described where it
	// moved them to.
	for _, want := range []string{"`app/<context>`", "`cmd/<binary>`", "pkg/, foundation/"} {
		if !strings.Contains(out, want) {
			t.Errorf("the layout is missing %q", want)
		}
	}
	// Each rule carries the identifier a finding would carry, so a reader who
	// has seen the failure can find the sentence and the other way round.
	for _, want := range []string{"K8-MAIN-LOCATION", "K7-INFRA-DOMAIN-FREE", "K6-ADAPTER-WIRED-IN-CMD"} {
		if !strings.Contains(out, want) {
			t.Errorf("rule %s is enforced but not described", want)
		}
	}
}

// TestOnlyEnforcedRulesAreDescribed is the discipline that keeps the chapter
// worth reading.
//
// A description that outruns its checks is worse than none, because a clean
// run will never contradict it. The layering rules run only where the profile
// separates presentation from adapters, so the nago profile — which has
// neither directory — must not claim them.
func TestOnlyEnforcedRulesAreDescribed(t *testing.T) {
	t.Parallel()
	out, _ := runSpeclink(t, "generate", "../../testdata/example")

	if strings.Contains(out, "K6-ADAPTER-WIRED-IN-CMD") {
		t.Error("a rule this profile does not enforce is described as if it were")
	}
	// It still describes the rules it does enforce.
	if !strings.Contains(out, "K8-MAIN-EXISTS") {
		t.Errorf("a rule this profile does enforce is missing:\n%s", chaptersOf(out))
	}
}

// TestDevelopersGetTheFieldsNotJustTheTypeName is what makes the surface
// chapter usable rather than merely correct.
//
// The catalogue names the type that crosses an address, which is enough to
// audit the surface and useless for calling it: the name of a Go struct is not
// a wire format. The fields were always read — they are how a breaking change
// is detected — and were thrown away before the document.
func TestDevelopersGetTheFieldsNotJustTheTypeName(t *testing.T) {
	t.Parallel()
	out, _ := runSpeclink(t, "generate", "../../testdata/example")

	if !strings.Contains(out, "### What crosses each address") {
		t.Fatalf("the surface is catalogued but not described:\n%s", chaptersOf(out))
	}
	// The wire name, which is what actually appears in the payload and is not
	// the field name.
	for _, want := range []string{"`quoteId`", "`title`", "`sequence`"} {
		if !strings.Contains(out, want) {
			t.Errorf("the payload is missing the wire name %q", want)
		}
	}
}
