package reqtree

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// TestOrderIndependence pins the central design guarantee: no annotation
// changes the meaning of another, so permuting the input must not change the
// result by a single byte.
//
// This is what the two pass design buys. Declarations are collected first and
// references resolved afterwards, which makes forward references legal and the
// order irrelevant. It also restores, at the model level, the determinism that
// LLM generated code gives up.
func TestOrderIndependence(t *testing.T) {
	reqs := func() []*ir.Requirement {
		return []*ir.Requirement{
			req("R-DEC-BASE", "m/dec.RDecBase", ir.Decision, ir.Abstract),
			// Forward reference: declared before the requirement it derives
			// from appears in the slice.
			derive(req("R-QUOTE-SUBMIT", "m/quote.RQuoteSubmit", ir.Functional, ir.Normative), "m/dec.RDecBase"),
			derive(req("R-QUOTE-APPROVE", "m/quote.RQuoteApprove", ir.Functional, ir.Normative), "m/dec.RDecBase"),
			req("R-NFR-AUDIT", "m/nfr.RNfrAudit", ir.NonFunctional, ir.Normative),
		}
	}

	want := build(t, reqs())

	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 20; i++ {
		in := reqs()
		rng.Shuffle(len(in), func(a, b int) { in[a], in[b] = in[b], in[a] })

		if got := build(t, in); !bytes.Equal(got, want) {
			t.Fatalf("permutation %d changed the result:\nwant:\n%s\ngot:\n%s", i, want, got)
		}
	}
}

// TestCycleDetected guards the acyclicity of the derivation graph. A cycle
// means none of the requirements in it can be justified.
func TestCycleDetected(t *testing.T) {
	a := derive(req("R-QUOTE-A", "m/quote.A", ir.Functional, ir.Normative), "m/quote.B")
	b := derive(req("R-QUOTE-B", "m/quote.B", ir.Functional, ir.Normative), "m/quote.A")

	out := &diag.Set{}
	Build(t.TempDir(), []*ir.Requirement{a, b}, out)

	if out.Len() == 0 {
		t.Fatal("expected a cycle to be reported")
	}
	if got := render(t, out); !bytes.Contains(got, []byte("cycle in DerivedFrom")) {
		t.Errorf("expected a cycle finding, got:\n%s", got)
	}
}

// TestDuplicateID guards requirement identity.
func TestDuplicateID(t *testing.T) {
	out := &diag.Set{}
	Build(t.TempDir(), []*ir.Requirement{
		req("R-QUOTE-X", "m/quote.X1", ir.Functional, ir.Normative),
		req("R-QUOTE-X", "m/quote.X2", ir.Functional, ir.Normative),
	}, out)

	if got := render(t, out); !bytes.Contains(got, []byte("declared twice")) {
		t.Errorf("expected a duplicate ID finding, got:\n%s", got)
	}
}

// build resolves the tree and renders both the graph and the diagnostics, so a
// permutation cannot hide behind identical findings but a different graph.
func build(t *testing.T, reqs []*ir.Requirement) []byte {
	t.Helper()

	out := &diag.Set{}
	tree := Build(t.TempDir(), reqs, out)

	var buf bytes.Buffer
	for _, r := range tree.All() {
		buf.WriteString(r.ID + " " + r.Kind.String() + " " + r.Status.String())
		for _, d := range r.DerivedFrom {
			buf.WriteString(" <- " + d)
		}
		buf.WriteByte('\n')
	}
	buf.Write(render(t, out))
	return buf.Bytes()
}

func render(t *testing.T, out *diag.Set) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := out.WriteText(&buf); err != nil {
		t.Fatalf("render findings: %v", err)
	}
	return buf.Bytes()
}

func req(id, goIdent string, k ir.Kind, s ir.Status) *ir.Requirement {
	return &ir.Requirement{
		ID:      id,
		GoIdent: goIdent,
		Kind:    k,
		Status:  s,
		Text:    "text of " + id,
		Pos:     ir.Position{File: "mem://" + id, Line: 1, Col: 1},
	}
}

func derive(r *ir.Requirement, from ...string) *ir.Requirement {
	r.DerivedFrom = from
	return r
}
