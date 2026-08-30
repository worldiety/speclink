package main

import (
	"strings"
	"testing"
)

// The world outside the code is the one thing speclink cannot read. No Go
// module states that a sales clerk exists or that a directory on disk is
// somebody else's responsibility, so it is declared — and then held against
// what the code does show, which is where the adapters are.

func TestTopologyIsChecked(t *testing.T) {
	t.Parallel()
	const topo = "topology/boundary.topology.go"

	for _, tc := range []struct {
		name   string
		break_ func(t *testing.T, dir string)
		want   string
	}{
		{
			name: "a channel end resolves to nothing",
			break_: func(t *testing.T, dir string) {
				rewrite(t, dir, topo, `To:       "dateiablage",`, `To:       "dateiablge",`)
			},
			want: `"dateiablge" is neither a declared participant nor a package of this module`,
		},
		{
			// The empty one is always the interesting one.
			name: "a channel does not say what protects it",
			break_: func(t *testing.T, dir string) {
				rewrite(t, dir, topo, `Crypto:   "entfällt, lokal",`, `Crypto:   "",`)
			},
			want: "leaves out Crypto",
		},
		{
			name: "a channel answers to no requirement",
			break_: func(t *testing.T, dir string) {
				rewrite(t, dir, topo, "\tSatisfies: []spec.Requirement{\n\t\tdec.RDecQuoteState,\n\t},\n", "")
				rewrite(t, dir, topo, "\t\"example.com/bare/requirements/dec\"\n", "")
			},
			want: "answers to no requirement",
		},
		{
			name: "nothing reaches a declared participant",
			break_: func(t *testing.T, dir string) {
				rewrite(t, dir, topo, `From:     "vertrieb",`, `From:     "app/sales/cli",`)
			},
			want: "Vertrieb is declared as an actor that nothing reaches",
		},
		{
			name: "two participants share an id",
			break_: func(t *testing.T, dir string) {
				rewrite(t, dir, topo, `ID:   "dateiablage",`, `ID:   "vertrieb",`)
			},
			want: `"vertrieb" is declared twice`,
		},
		{
			// The backward direction, and the one no declaration can produce
			// on its own: a way out of the system that never appeared in any
			// interface list.
			name: "the system reaches outside where no channel says so",
			break_: func(t *testing.T, dir string) {
				rewrite(t, dir, topo, `From:     "app/sales/adapter/fs",`, `From:     "app/sales",`)
			},
			want: "app/sales/adapter/fs reaches outside the system and no channel describes it",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := copyFixture(t, "../../testdata/bare")
			tc.break_(t, dir)

			out, code := runVerify(t, dir)
			if code == 0 {
				t.Fatalf("the topology is broken and nothing was reported:\n%s", out)
			}
			if !strings.Contains(out, tc.want) {
				t.Errorf("expected %q:\n%s", tc.want, out)
			}
		})
	}
}

// An in memory adapter is structurally an adapter and crosses nothing.
//
// The rule cannot see the difference, and weakening it to guess would be worse
// than the waiver: the waiver names the reason, counts toward the figure the
// way every other waiver does, and stops being true the moment somebody puts a
// database behind the port.
func TestInMemoryAdapterIsWaivedRatherThanExempted(t *testing.T) {
	t.Parallel()
	out, code := runVerify(t, "../../testdata/bare")
	if code != 0 {
		t.Fatalf("the bare fixture did not verify:\n%s", out)
	}
	if !strings.Contains(out, "2 channels (2 of 2 boundaries described)") {
		t.Errorf("the waived boundary is not counted like every other waiver:\n%s", summary(out))
	}

	dir := copyFixture(t, "../../testdata/bare")
	rewrite(t, dir, "app/billing/adapter/mem/invoices.annotation.go",
		`spec.Waive("K17-ADAPTER-NO-CHANNEL"`, `spec.Waive("K17-SOMETHING-ELSE"`)

	out, code = runVerify(t, dir)
	if code == 0 {
		t.Fatalf("the waiver was removed and nothing was reported:\n%s", out)
	}
	if !strings.Contains(out, "app/billing/adapter/mem reaches outside") {
		t.Errorf("the waiver was doing nothing:\n%s", out)
	}
}

// A project that has not described its boundary gets no figure about it. The
// same ramp processes are on, for the same reason: before the first
// declaration there is nothing to be incomplete against.
func TestTopologyFigureAppearsOnlyWhereDeclared(t *testing.T) {
	t.Parallel()
	out, code := runVerify(t, "../../testdata/example")
	if code != 0 {
		t.Fatalf("the nago fixture did not verify:\n%s", out)
	}
	if strings.Contains(out, "boundaries described") {
		t.Errorf("a project without a topology was given one:\n%s", summary(out))
	}
}

// A channel satisfies requirements through the same machinery a construct does.
func TestChannelCountsTowardsCoverage(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")
	// The persistence ruling loses its construct bindings and keeps only the
	// channel. It must still read as covered.
	rewrite(t, dir, "app/sales/model.annotation.go",
		"var _ = spec.For[Quote](\n\tspec.Satisfies(dec.RDecQuoteState),\n)",
		"var _ = spec.For[Quote](\n\tspec.Satisfies(quote.RQuoteLookup),\n)")

	out, _ := runVerify(t, dir)
	if strings.Contains(out, "R-DEC-QUOTE-STATE is satisfied by nothing") {
		t.Errorf("a requirement covered by a channel read as uncovered:\n%s", out)
	}
}
