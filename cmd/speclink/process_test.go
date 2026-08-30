package main

import (
	"strings"
	"testing"
)

// A process is a graph, so the compiler stops helping at the edges: nodes are
// joined by name and a name is a string. Everything a type checker would have
// caught has to be caught here, and these pin that it is.
//
// The fixture's process is deliberately awkward — a fork that waits, a choice
// with three branches and one of them going backwards, two different endings —
// because a rule set proved only against a straight line is a rule set nobody
// has tested.

func TestProcessGraphIsChecked(t *testing.T) {
	t.Parallel()
	const proc = "processes/quote.process.go"

	for _, tc := range []struct {
		name   string
		break_ func(t *testing.T, dir string)
		want   string
	}{
		{
			// The commonest mistake there is, and the one the compiler used to
			// catch before the graph went to strings.
			name: "an edge names a node that does not exist",
			break_: func(t *testing.T, dir string) {
				rewrite(t, dir, proc, `{From: "liste", To: "zusammen"},`, `{From: "liste", To: "zusamen"},`)
			},
			want: `has an edge to "zusamen", which is not a node of it`,
		},
		{
			name: "two nodes share a name",
			break_: func(t *testing.T, dir string) {
				rewrite(t, dir, proc, `spec.Do[sales.ListQuotes]("liste")`, `spec.Do[sales.ListQuotes]("uebersicht")`)
			},
			want: `has two nodes called "uebersicht"`,
		},
		{
			// One corner loops on itself while the rest of the graph still
			// finishes, so the report is per node rather than per process.
			name: "one step can never finish although the process can",
			break_: func(t *testing.T, dir string) {
				rewrite(t, dir, proc, `{From: "zurueckgezogen", To: "verworfen"},`, `{From: "zurueckgezogen", To: "zurueckgezogen"},`)
			},
			want: `from "zurueckgezogen" the process can never finish`,
		},
		{
			// Every branch loops back, so nothing reaches an end. Reported once
			// against the process rather than once per node: a graph that
			// cannot finish is one mistake, and listing every step of it would
			// bury the fact in its own consequences.
			name: "no end is reachable from anywhere",
			break_: func(t *testing.T, dir string) {
				rewrite(t, dir, proc, `{From: "freigeben", To: "freigegeben"},`, `{From: "freigeben", To: "abgeben"},`)
				rewrite(t, dir, proc, `{From: "zurueckgezogen", To: "verworfen"},`, `{From: "zurueckgezogen", To: "abgeben"},`)
			},
			want: "can never finish: no end is reachable from anywhere in it",
		},
		{
			name: "an alternative states no condition",
			break_: func(t *testing.T, dir string) {
				rewrite(t, dir, proc, `{From: "pruefen", To: "abgeben", When: "nachzubessern"},`, `{From: "pruefen", To: "abgeben"},`)
			},
			want: `the branch from "pruefen" to "abgeben" states no condition`,
		},
		{
			// A fan out nothing named: a reader cannot tell whether both follow
			// or one of them does, which is exactly what a gateway is for.
			name: "a step has two outgoing edges",
			break_: func(t *testing.T, dir string) {
				rewrite(t, dir, proc, `{From: "abgeben", To: "abgegeben"},`,
					"{From: \"abgeben\", To: \"abgegeben\"},\n\t\t{From: \"abgeben\", To: \"freigegeben\"},")
			},
			want: `"abgeben" has 2 outgoing edges`,
		},
		{
			name: "a gateway divides nothing",
			break_: func(t *testing.T, dir string) {
				rewrite(t, dir, proc, `{From: "aufteilen", To: "liste"},`, `{From: "zusammen", To: "liste"},`)
			},
			want: `"aufteilen" is a fork with 1 outgoing edge`,
		},
		{
			// The graph is held against the code it claims to describe. An
			// event is a fact that was recorded, not work that is done.
			name: "an activity names an event",
			break_: func(t *testing.T, dir string) {
				rewrite(t, dir, proc, `spec.Do[sales.ListQuotes]("liste")`, `spec.Do[sales.QuoteSubmitted]("liste")`)
			},
			want: "QuoteSubmitted is an event and cannot be an activity",
		},
		{
			name: "a raised event names a use case",
			break_: func(t *testing.T, dir string) {
				rewrite(t, dir, proc, `spec.Emit[sales.QuoteWithdrawn]("zurueckgezogen")`, `spec.Emit[sales.WithdrawQuote]("zurueckgezogen")`)
			},
			want: "WithdrawQuote is a use case and is not an event",
		},
		{
			// A process is the promise that the separate actions add up to
			// something somebody asked for. Which promise cannot be read off
			// the graph.
			name: "the process names no requirement",
			break_: func(t *testing.T, dir string) {
				rewrite(t, dir, proc, "\t\tquote.RQuoteSubmit,\n\t\tquote.RQuoteApprove,\n", "")
				rewrite(t, dir, proc, "\t\"example.com/erp/requirements/fun/quote\"\n", "")
			},
			want: "process P-QUOTE-DECISION names no requirement",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := copyFixture(t, "../../testdata/example")
			tc.break_(t, dir)

			out, code := runVerify(t, dir)
			if code == 0 {
				t.Fatalf("the graph is broken and nothing was reported:\n%s", out)
			}
			if !strings.Contains(out, tc.want) {
				t.Errorf("expected %q:\n%s", tc.want, out)
			}
		})
	}
}

// A start with something leading into it is not a start, and an end with
// something leading out of it is not an end. Both are removals rather than
// rewrites, so they get their own cases.
func TestProcessNeedsABeginningAndAnEnd(t *testing.T) {
	t.Parallel()
	const proc = "processes/quote.process.go"

	t.Run("no start", func(t *testing.T) {
		t.Parallel()
		dir := copyFixture(t, "../../testdata/example")
		rewrite(t, dir, proc, `spec.Start("entwurf", "Angebot ist entworfen"),`, "")
		rewrite(t, dir, proc, `{From: "entwurf", To: "abgeben"},`, "")

		out, code := runVerify(t, dir)
		if code == 0 {
			t.Fatalf("a process without a start passed:\n%s", out)
		}
		if !strings.Contains(out, "process P-QUOTE-DECISION has no start") {
			t.Errorf("expected K16-PROCESS-NO-START:\n%s", out)
		}
	})

	t.Run("a start is led into", func(t *testing.T) {
		t.Parallel()
		dir := copyFixture(t, "../../testdata/example")
		rewrite(t, dir, proc, `{From: "freigeben", To: "freigegeben"},`, `{From: "freigeben", To: "entwurf"},`)

		out, code := runVerify(t, dir)
		if code == 0 {
			t.Fatalf("a start with an incoming edge passed:\n%s", out)
		}
		if !strings.Contains(out, `the start "entwurf" has 1 incoming edge`) {
			t.Errorf("expected the degree rule to name the start:\n%s", out)
		}
	})
}

// The graph is not judged where the recognisers never looked.
//
// This is the K9 mistake one step removed: a scope that leaves out the contexts
// would make every activity read as naming something that does not exist, and
// the run would fail a project for a setting it chose on purpose.
func TestProcessRefsAreNotJudgedOutOfScope(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/example")
	writeConfig(t, dir, `{"profile":"go_nago_ddd1","scope":["processes","requirements/fun/quote","requirements/dec"]}`)

	out, _ := runVerify(t, dir)
	if strings.Contains(out, "is not a recognised construct") {
		t.Errorf("a reference into an unmeasured package was reported as a mistake:\n%s", out)
	}
}

// A project that has not adopted processes gets no figure about them. A share
// of nothing is not a hundred percent, it is no claim at all.
func TestProcessFigureAppearsOnlyWhereThereAreProcesses(t *testing.T) {
	t.Parallel()
	out, code := runVerify(t, "../../testdata/bare")
	if code != 0 {
		t.Fatalf("the bare fixture did not verify:\n%s", out)
	}
	if strings.Contains(out, "process") {
		t.Errorf("a project without processes was given a process figure:\n%s", summary(out))
	}

	out, code = runVerify(t, "../../testdata/example")
	if code != 0 {
		t.Fatalf("the nago fixture did not verify:\n%s", out)
	}
	if !strings.Contains(out, "1 process (1 sound, 5 of 5 steps placed)") {
		t.Errorf("the declared process is not counted:\n%s", summary(out))
	}
}

// The backward direction, and the reason the model is worth more than a
// drawing: a use case says what one action promises and never where that action
// sits in the business.
func TestWorkOutsideEveryProcessIsReported(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/example")
	rewrite(t, dir, "processes/quote.process.go", "\t\tspec.Do[sales.ListQuotes](\"liste\"),\n", "")
	rewrite(t, dir, "processes/quote.process.go", "\t\t{From: \"aufteilen\", To: \"liste\"},\n", "")
	rewrite(t, dir, "processes/quote.process.go", "\t\t{From: \"liste\", To: \"zusammen\"},\n", "")

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a use case belongs to no process and nothing was reported:\n%s", out)
	}
	if !strings.Contains(out, "ListQuotes belongs to no process") {
		t.Errorf("expected K16-WORK-OUTSIDE-PROCESS:\n%s", out)
	}
	// The figure is the honest half of the same statement: four of five, not a
	// bare percentage with the denominator left off.
	if !strings.Contains(out, "4 of 5 steps placed") {
		t.Errorf("the share of placed work did not move:\n%s", summary(out))
	}
}

// A fact that outlives the code is recorded at some moment in the business.
// An event no course mentions is a moment nobody wrote down.
func TestEventNoProcessMentionsIsReported(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/example")
	rewrite(t, dir, "processes/quote.process.go", "\t\tspec.Emit[sales.QuoteWithdrawn](\"zurueckgezogen\"),\n", "")
	rewrite(t, dir, "processes/quote.process.go",
		"\t\t{From: \"zurueckziehen\", To: \"zurueckgezogen\"},\n\t\t{From: \"zurueckgezogen\", To: \"verworfen\"},\n",
		"\t\t{From: \"zurueckziehen\", To: \"verworfen\"},\n")

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("an event is in no process and nothing was reported:\n%s", out)
	}
	if !strings.Contains(out, "QuoteWithdrawn is raised or awaited by no process") {
		t.Errorf("expected K16-EVENT-UNPLACED:\n%s", out)
	}
}

// spec.External finally does something.
//
// It was recorded and read by nothing for as long as it existed, and its own
// documentation named a check that never was. This is the rule it was written
// for: nothing here produces the fact, so no process can be expected to.
func TestExternalExemptsAnEventFromEveryProcess(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/example")
	rewrite(t, dir, "processes/quote.process.go", "\t\tspec.Emit[sales.QuoteWithdrawn](\"zurueckgezogen\"),\n", "")
	rewrite(t, dir, "processes/quote.process.go",
		"\t\t{From: \"zurueckziehen\", To: \"zurueckgezogen\"},\n\t\t{From: \"zurueckgezogen\", To: \"verworfen\"},\n",
		"\t\t{From: \"zurueckziehen\", To: \"verworfen\"},\n")
	rewrite(t, dir, "app/sales/model.annotation.go",
		"\tspec.Transition[QuoteWithdrawn](\"withdrawn\"),\n",
		"\tspec.Transition[QuoteWithdrawn](\"withdrawn\"),\n\tspec.External(),\n")

	out, _ := runVerify(t, dir)
	if strings.Contains(out, "QuoteWithdrawn is raised or awaited by no process") {
		t.Errorf("an event declared as arriving from outside was still demanded:\n%s", out)
	}
}

// A process satisfies requirements the way a construct does, through the same
// machinery rather than a second path — because a second path is where two
// answers to one question come from.
func TestProcessCountsTowardsCoverage(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/example")
	// The approval requirement loses its construct binding and keeps only the
	// process. It must still read as covered.
	rewrite(t, dir, "app/sales/uc_approve_quote.annotation.go",
		"spec.Satisfies(quote.RQuoteApprove)", "spec.Satisfies(quote.RQuoteSubmit)")

	out, _ := runVerify(t, dir)
	if strings.Contains(out, "R-QUOTE-APPROVE") && strings.Contains(out, "is satisfied by nothing") {
		t.Errorf("a requirement covered only by a process read as uncovered:\n%s", out)
	}
}
