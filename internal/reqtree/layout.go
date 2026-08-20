package reqtree

import (
	"path/filepath"
	"strings"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// Layout checks tie together three representations of the same fact: the
// directory a requirement lives in, the prefix of its ID and its Kind field.
//
// Redundancy is harmless once it is checked; unchecked redundancy is the main
// error pattern this whole design attacks (M1). The ID stays authoritative and
// lives in the term, so moving a file never breaks a reference — but a file in
// the wrong place is a mistake worth reporting.
//
// Layout:
//
//	requirements/dec/R-DEC-EVENTSOURCING.spec.go        Kind = Decision
//	requirements/nfr/R-NFR-AUDIT.spec.go                Kind = NonFunctional
//	requirements/cst/R-CST-TENANT.spec.go               Kind = Constraint
//	requirements/fun/quote/R-QUOTE-SUBMIT.spec.go       Kind = Functional
//	requirements/fun/quote/R-QUOTE-SUBMIT/              attachments, no .go
//
// Cross cutting kinds are grouped by kind because they have no domain home by
// definition. Functional requirements are the majority and are grouped by
// domain, since a package named "functional" would carry no information at the
// call site.

// CheckLayout verifies the directory, file name and ID prefix of every
// requirement against its Kind.
func (t *Tree) CheckLayout(out *diag.Set) {
	for _, id := range t.sortedIDs() {
		r := t.ByID[id]
		if r.Kind == 0 {
			continue // already reported
		}
		t.checkFileName(r, out)
		t.checkPrefix(r, out)
		t.checkDir(r, out)
	}
}

// checkFileName requires <ID>.spec.go.
func (t *Tree) checkFileName(r *ir.Requirement, out *diag.Set) {
	base := filepath.Base(r.Pos.File)
	want := r.ID + ".spec.go"
	if base == want {
		return
	}
	out.Add(diag.Finding{
		Code: diag.Code(diag.PhaseResolve, 30),
		Pos:  r.Pos,
		What: "requirement " + r.ID + " is declared in " + base + ", expected " + want + ".",
		Why:  "One requirement per file, named after its ID. This is what lets an attachment folder of the same name sit next to it, and what makes the tree navigable without an index.",
		How:  "Rename the file to " + want + ", or move this declaration into its own file.",
	})
}

// checkPrefix requires the ID prefix to match the Kind for cross cutting kinds.
//
// The correlation was not invented here: in the reference project all 125
// decisions carry R-DEC-, all 10 quality goals R-NFR- and all 7 constraints
// R-CST-, without a single exception. The convention already existed; it was
// only never written down and never checked.
func (t *Tree) checkPrefix(r *ir.Requirement, out *diag.Set) {
	prefix, ok := idPrefix(r.ID)
	if !ok {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseResolve, 31),
			Pos:  r.Pos,
			What: "requirement ID " + r.ID + " does not follow the pattern R-<PREFIX>-<NAME>.",
			Why:  "The prefix groups the requirement and must agree with its Kind and its directory.",
			How:  "Rename it, e.g. R-QUOTE-SUBMIT for a functional requirement of the quote domain.",
		})
		return
	}

	want, fixed := kindPrefix(r.Kind)
	if !fixed {
		return // functional requirements carry a domain prefix, checked in checkDir
	}
	if prefix == want {
		return
	}
	out.Add(diag.Finding{
		Code: diag.Code(diag.PhaseResolve, 32),
		Pos:  r.Pos,
		What: r.ID + " has Kind " + r.Kind.String() + " but prefix " + prefix + ", expected " + want + ".",
		Why:  "Cross cutting requirements are grouped by kind, and directory, prefix and Kind must state the same thing.",
		How:  "Rename it to R-" + want + "-… , or correct the Kind field.",
	})
}

// checkDir requires the directory to match the Kind, and for functional
// requirements the domain directory to match the ID prefix.
func (t *Tree) checkDir(r *ir.Requirement, out *diag.Set) {
	dir := filepath.Dir(r.Pos.File)
	rel, err := filepath.Rel(t.root, dir)
	if err != nil {
		return
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")

	level, ok := firstOf(parts, "dec", "nfr", "cst", "fun")
	if !ok {
		return // outside the conventional tree; layout is not enforced there
	}

	if want := r.Kind.Dir(); parts[level] != want {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseResolve, 33),
			Pos:  r.Pos,
			What: r.ID + " has Kind " + r.Kind.String() + " but lives in " + parts[level] + "/, expected " + want + "/.",
			Why:  "The first directory level is the Kind. Moving a requirement between them changes its identity, which is why it also changes the ID.",
			How:  "Move the file to " + want + "/, or correct the Kind field. If the kind genuinely changed, give it a new ID and list the predecessor in Supersedes.",
		})
		return
	}

	if r.Kind != ir.Functional {
		return
	}
	if level+1 >= len(parts) {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseResolve, 34),
			Pos:  r.Pos,
			What: "functional requirement " + r.ID + " lies directly in fun/.",
			Why:  "Functional requirements are grouped by domain; fun/ itself is only a structuring level and holds no package.",
			How:  "Move it into a domain directory, e.g. fun/" + strings.ToLower(prefixOf(r.ID)) + "/.",
		})
		return
	}

	domain := parts[level+1]
	prefix, ok := idPrefix(r.ID)
	if !ok || strings.EqualFold(domain, prefix) {
		return
	}
	out.Add(diag.Finding{
		Code: diag.Code(diag.PhaseResolve, 35),
		Pos:  r.Pos,
		What: r.ID + " lies in fun/" + domain + "/ but carries the prefix " + prefix + ".",
		Why:  "Directory, package name and ID prefix are one fact in three representations and must agree, otherwise the tree misleads its reader.",
		How:  "Move the file to fun/" + strings.ToLower(prefix) + "/, or rename it to R-" + strings.ToUpper(domain) + "-… .",
	})
}

// kindPrefix returns the mandatory ID prefix for a kind. The second result is
// false for Functional, which carries a domain prefix instead.
func kindPrefix(k ir.Kind) (string, bool) {
	switch k {
	case ir.Decision:
		return "DEC", true
	case ir.NonFunctional:
		return "NFR", true
	case ir.Constraint:
		return "CST", true
	}
	return "", false
}

// idPrefix extracts the middle segment of R-<PREFIX>-<NAME>.
func idPrefix(id string) (string, bool) {
	parts := strings.Split(id, "-")
	if len(parts) < 3 || parts[0] != "R" || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func prefixOf(id string) string {
	p, _ := idPrefix(id)
	return p
}

// firstOf returns the index of the first element of parts equal to one of the
// candidates.
func firstOf(parts []string, candidates ...string) (int, bool) {
	for i, p := range parts {
		for _, c := range candidates {
			if p == c {
				return i, true
			}
		}
	}
	return 0, false
}
