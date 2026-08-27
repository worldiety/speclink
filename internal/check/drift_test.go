package check

import (
	"bytes"
	"testing"

	"github.com/worldiety/speclink/internal/baseline"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/reqtree"
	"github.com/worldiety/speclink/internal/source"
)

// The defect these rules exist for: everything still resolves, every figure is
// still at a hundred percent, and the meaning has moved. Nothing else in
// speclink can see it, because there is nothing dangling to see.

func TestRewrittenRequirementIsReported(t *testing.T) {
	root := t.TempDir()
	r := cite("R-A", "src/flow.md", "abgabe")
	r.Text = "Eine fortlaufende Nummer MUSS gezogen werden."

	base := recorded(t, root, r)
	r.Text = "Eine zufällige Nummer DARF vergeben werden."

	got := drift(t, root, base, r)
	if !bytes.Contains(got, []byte("the text of R-A changed")) {
		t.Fatalf("a rewritten requirement was not reported:\n%s", got)
	}
}

// Wrapping a sentence differently is not a rewrite. A rule that fires on
// reflowing is one that gets waived by habit.
func TestReflowingIsNotAChange(t *testing.T) {
	root := t.TempDir()
	r := cite("R-A", "src/flow.md", "abgabe")
	r.Text = "Eine fortlaufende Nummer MUSS gezogen werden."

	base := recorded(t, root, r)
	r.Text = "Eine fortlaufende\n  Nummer MUSS   gezogen werden.  "

	if got := drift(t, root, base, r); len(got) != 0 {
		t.Fatalf("reflowing reported as a change:\n%s", got)
	}
}

func TestRewrittenSourceSegmentIsReported(t *testing.T) {
	root := t.TempDir()
	write(t, root, "src/flow.md", "# Abgabe\n\nEine fortlaufende Nummer wird gezogen.\n")
	r := cite("R-A", "src/flow.md", "abgabe")

	base := recorded(t, root, r)
	write(t, root, "src/flow.md", "# Abgabe\n\nEine zufällige Nummer wird vergeben.\n")

	got := drift(t, root, base, r)
	if !bytes.Contains(got, []byte("changed since it was last recorded")) {
		t.Fatalf("a rewritten section was not reported:\n%s", got)
	}
	// The finding has to name what to re-read, or acting on it means searching
	// the tree for whatever cited this section.
	if !bytes.Contains(got, []byte("R-A")) {
		t.Errorf("the finding does not name the requirement to re-read:\n%s", got)
	}
}

// Nothing was promised about it yet, so nothing can have drifted. A newly
// written requirement is read for the first time by the freeze that records it.
func TestUnrecordedRequirementDoesNotDrift(t *testing.T) {
	root := t.TempDir()
	write(t, root, "src/flow.md", "# Abgabe\n\nText.\n")

	base := &baseline.File{Version: baseline.Version}
	got := drift(t, root, base, cite("R-A", "src/flow.md", "abgabe"))

	if len(got) != 0 {
		t.Fatalf("an unrecorded requirement was reported:\n%s", got)
	}
}

// Freezing is what turns a change into a read. It is the moment of review, and
// the diff of the file is what the review actually looks at.
func TestFreezingClearsTheDrift(t *testing.T) {
	root := t.TempDir()
	write(t, root, "src/flow.md", "# Abgabe\n\nErste Fassung.\n")
	r := cite("R-A", "src/flow.md", "abgabe")

	base := recorded(t, root, r)
	write(t, root, "src/flow.md", "# Abgabe\n\nZweite Fassung.\n")

	if got := drift(t, root, base, r); len(got) == 0 {
		t.Fatal("expected drift before freezing")
	}
	record(t, root, base, r)
	if got := drift(t, root, base, r); len(got) != 0 {
		t.Fatalf("drift survived the freeze:\n%s", got)
	}
}

// The record is what lets a generator keep an identifier stable across a
// regeneration. Without it a regenerated tree renames the requirement and
// breaks every spec.Satisfies in the code at once.
func TestRecordedIDIsFoundByItsSegment(t *testing.T) {
	root := t.TempDir()
	write(t, root, "src/flow.md", "# Abgabe\n\nText.\n")
	base := recorded(t, root, cite("R-QUOTE-SUBMIT", "src/flow.md", "abgabe"))

	got, ok := base.RequirementOf("src/flow.md#abgabe")
	if !ok || got != "R-QUOTE-SUBMIT" {
		t.Fatalf("got (%q, %v), want R-QUOTE-SUBMIT", got, ok)
	}
}

func TestWaiverSuppressesRequirementDrift(t *testing.T) {
	root := t.TempDir()
	r := cite("R-A", "src/flow.md", "abgabe")
	r.Text = "Erste Fassung."

	base := recorded(t, root, r)
	r.Text = "Zweite Fassung."

	waivers := ir.Waivers{}
	waivers.Waive("R-A", RuleReqChanged, "the change was editorial and was reviewed out of band")

	if got := driftWith(t, root, base, waivers, r); len(got) != 0 {
		t.Fatalf("waived drift reported:\n%s", got)
	}
}

// recorded builds a baseline holding the current state, as a freeze would.
func recorded(t *testing.T, root string, reqs ...*ir.Requirement) *baseline.File {
	t.Helper()
	base := &baseline.File{Version: baseline.Version}
	record(t, root, base, reqs...)
	return base
}

func record(t *testing.T, root string, base *baseline.File, reqs ...*ir.Requirement) {
	t.Helper()
	fill(base)

	out := &diag.Set{}
	tree := reqtree.Build(root, clone(reqs), out)
	set := source.NewSet(root)
	docs, _ := source.Walk(root, []string{"src"})
	Record(base, tree, set, docs)
}

func drift(t *testing.T, root string, base *baseline.File, reqs ...*ir.Requirement) []byte {
	t.Helper()
	return driftWith(t, root, base, nil, reqs...)
}

func driftWith(t *testing.T, root string, base *baseline.File, waivers ir.Waivers, reqs ...*ir.Requirement) []byte {
	t.Helper()

	out := &diag.Set{}
	tree := reqtree.Build(root, clone(reqs), out)
	set := source.NewSet(root)
	docs, _ := source.Walk(root, []string{"src"})

	cov := Coverage{BySatisfier: map[string][]ir.Target{}}
	for _, r := range reqs {
		cov.BySatisfier[r.ID] = []ir.Target{{Kind: ir.TargetType, Name: "m/q.Impl" + r.ID}}
	}
	srcCov := CoverSources(tree, set, docs, ir.Waivers{}, &diag.Set{})
	Drift(tree, set, docs, cov, srcCov, base, waivers, out)

	var buf bytes.Buffer
	if err := out.WriteText(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// clone keeps the tree from mutating the fixtures a test reuses across calls.
func clone(reqs []*ir.Requirement) []*ir.Requirement {
	out := make([]*ir.Requirement, 0, len(reqs))
	for _, r := range reqs {
		c := *r
		c.DerivedFrom = append([]string(nil), r.DerivedFrom...)
		c.Sources = append([]ir.Source(nil), r.Sources...)
		out = append(out, &c)
	}
	return out
}

func fill(f *baseline.File) {
	if f.Requirements == nil {
		f.Requirements = map[string]baseline.Requirement{}
	}
	if f.Sources == nil {
		f.Sources = map[string]baseline.Segment{}
	}
	if f.Types == nil {
		f.Types = map[string]baseline.Entry{}
	}
}
