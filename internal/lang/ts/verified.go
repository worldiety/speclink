package ts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang"
	"github.com/worldiety/speclink/internal/lang/ts/syntax"
)

// testFunctions open a test or a group of tests, in Vitest and Jest alike.
var testFunctions = map[string]bool{"describe": true, "suite": true, "it": true, "test": true}

// testModifiers may stand between a test function and its call.
var testModifiers = map[string]bool{
	"only": true, "skip": true, "concurrent": true, "sequential": true, "fails": true, "todo": true,
}

// TestName renders a test the way the claim and the report both name it: the
// file relative to the root, a dot, and the titles of the enclosing groups and
// the test joined by " > ".
//
// A report shortens a qualified name to what follows its last dot, so a dot in
// a title is spelled as a one dot leader. Otherwise "handles 1.5 weeks" would
// be reported as "5 weeks", and the reader would look for a test that does not
// exist.
func TestName(file string, titles []string) string {
	return file + "." + strings.ReplaceAll(strings.Join(titles, " > "), ".", "\u2024")
}

// Verifications reads the verified(…) calls of the tests.
//
// The claim is a call, and where it sits says which test makes it: the titles
// of the describe and test calls around it. That is the one spelling the
// report shares, so the join needs nothing else.
func (p *Project) Verifications(out *diag.Set) []ir.Binding {
	var bindings []ir.Binding
	claims, _ := p.allClaims(out)
	for _, c := range claims {
		bindings = append(bindings, c.Binding)
	}
	return bindings
}

// claim is one verified(…) call, with what is needed to find its results.
type claim struct {
	ir.Binding
	// Each records that a group or test around the claim is parametrised with
	// .each, so its title is a pattern the report fills in once per case.
	Each bool
}

// matches returns the passed tests a claim stands for: the one of its name,
// or for a parametrised test every case of it.
//
// literal holds the tests whose names are written out in full. A case of a
// parametrised group can look exactly like one of them — "Test for %s" fills
// in to "Test for merchant", which may also be a group of its own — and such a
// test made no claim, so it is never counted for one.
func (c claim) matches(passed, literal map[string]bool) []string {
	if !c.Each {
		if passed[c.Target.Name] {
			return []string{c.Target.Name}
		}
		return nil
	}
	re := eachPattern(c.Target.Name)
	var out []string
	for name := range passed {
		if !literal[name] && re.MatchString(name) {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

// placeholder is what Vitest substitutes in the title of a parametrised test:
// printf verbs and $-references to a field of the case.
var placeholder = regexp.MustCompile(`%[sdifjo#%]|\$[A-Za-z_][A-Za-z0-9_․]*`)

func eachPattern(name string) *regexp.Regexp {
	var b strings.Builder
	b.WriteString("^")
	last := 0
	for _, loc := range placeholder.FindAllStringIndex(name, -1) {
		b.WriteString(regexp.QuoteMeta(name[last:loc[0]]))
		b.WriteString(".*")
		last = loc[1]
	}
	b.WriteString(regexp.QuoteMeta(name[last:]))
	b.WriteString("$")
	return regexp.MustCompile(b.String())
}

// allClaims returns every claim of the tests, and the names of every test
// written out without parameters.
func (p *Project) allClaims(out *diag.Set) ([]claim, map[string]bool) {
	var claims []claim
	literal := map[string]bool{}
	for _, m := range p.Modules {
		if p.Layer(m) != LayerTest {
			continue
		}
		for _, f := range m.Scripts {
			claims = append(claims, p.claims(m, f, literal, out)...)
		}
	}
	return claims, literal
}

func (p *Project) claims(m *Module, f *syntax.File, literal map[string]bool, out *diag.Set) []claim {
	type frame struct {
		title string
		end   int
		each  bool
	}
	var (
		stack    []frame
		bindings []claim
	)
	toks := f.Toks
	for i := 0; i < len(toks); i++ {
		for len(stack) > 0 && i >= stack[len(stack)-1].end {
			stack = stack[:len(stack)-1]
		}
		t := toks[i]
		if t.Kind != syntax.Ident || i > 0 && toks[i-1].Is(".") {
			continue
		}
		if testFunctions[t.Text] {
			j := i + 1
			each := false
			for j+1 < len(toks) && toks[j].Is(".") {
				if toks[j+1].Is("each") && j+2 < len(toks) && toks[j+2].Is("(") {
					// describe.each(cases)(title, fn): the cases first, then
					// the call that carries the title.
					each = true
					j = f.GroupEnd(j + 2)
					continue
				}
				if !testModifiers[toks[j+1].Text] {
					break
				}
				j += 2
			}
			if j+1 < len(toks) && toks[j].Is("(") && (toks[j+1].Kind == syntax.String || toks[j+1].Kind == syntax.Template) {
				stack = append(stack, frame{title: toks[j+1].Text, end: f.GroupEnd(j), each: each})
				if t.Text == "test" || t.Text == "it" {
					var titles []string
					parametrised := false
					for _, fr := range stack {
						titles = append(titles, fr.title)
						parametrised = parametrised || fr.each
					}
					if !parametrised {
						literal[TestName(m.Path, titles)] = true
					}
				}
			}
			continue
		}
		if i+1 >= len(toks) || !toks[i+1].Is("(") || m.vocabularyCall([]string{t.Text}) != "verified" {
			continue
		}
		end := f.GroupEnd(i + 1)
		call := f.Value(syntax.Range{From: i, To: end})
		pos := m.Pos(t.Pos)
		if len(stack) == 0 {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseBinding, 23),
				Pos:  pos,
				What: "verified is called outside a test.",
				Why:  "A verification is a claim a test makes, and the test is what a result is recorded against. Outside one there is nothing a report could confirm.",
				How:  "Call verified(…) inside the test(…) or it(…) that demonstrates the requirement.",
			})
			continue
		}
		var titles []string
		each := false
		for _, fr := range stack {
			titles = append(titles, fr.title)
			each = each || fr.each
		}
		a := ir.Assertion{Kind: ir.AssertVerified, Pos: pos}
		for _, e := range call.Elems {
			if s, ok := p.reference(m, e); ok {
				a.Requirements = append(a.Requirements, s)
				continue
			}
			unresolved(out, m.Pos(e.Pos), e)
		}
		bindings = append(bindings, claim{Each: each, Binding: ir.Binding{
			Target:     ir.Target{Kind: ir.TargetFunc, Package: m.Dir, Name: TestName(m.Path, titles)},
			Assertions: []ir.Assertion{a},
			Pos:        pos,
		}})
		i = end - 1
	}
	return bindings
}

// vitestReport is the part of Vitest's JSON reporter output this needs. Jest
// writes the same shape, which is where Vitest took it from.
type vitestReport struct {
	TestResults []struct {
		Name             string `json:"name"`
		AssertionResults []struct {
			AncestorTitles []string `json:"ancestorTitles"`
			Title          string   `json:"title"`
			Status         string   `json:"status"`
		} `json:"assertionResults"`
	} `json:"testResults"`
}

// Demonstrations joins the claims in the tests with the results of a run.
//
// The results are what vitest run --reporter=json writes, handed over on
// standard input. Only passes count: a test that claimed something and then
// failed showed nothing, and a skipped one did not run.
func (p *Project) Demonstrations(in io.Reader, reqs lang.RequirementIndex) (map[string][]string, error) {
	data, err := io.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("read the test results: %w", err)
	}
	data, err = p.reportBody(data)
	if err != nil {
		return nil, err
	}
	var report vitestReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("read the test results: %w; pipe the output of vitest run --reporter=json", err)
	}
	passed := map[string]bool{}
	for _, file := range report.TestResults {
		rel := p.testFile(file.Name)
		for _, a := range file.AssertionResults {
			if a.Status != "passed" {
				continue
			}
			passed[TestName(rel, append(append([]string(nil), a.AncestorTitles...), a.Title))] = true
		}
	}

	out := map[string][]string{}
	claims, literal := p.allClaims(&diag.Set{})
	for _, c := range claims {
		tests := c.matches(passed, literal)
		if len(tests) == 0 {
			continue
		}
		for _, a := range c.Assertions {
			for _, ref := range a.Requirements {
				id, ok := reqs.IDOf(ref)
				if !ok {
					continue
				}
				for _, test := range tests {
					seen := false
					for _, have := range out[id] {
						seen = seen || have == test
					}
					if !seen {
						out[id] = append(out[id], test)
					}
				}
			}
		}
	}
	return out, nil
}

// testFile turns the file name a report gives into the root relative path the
// claims use.
//
// A report names the file absolutely, as it was on the machine that ran it,
// and that is rarely this one: the tests run in CI and the record is written
// wherever the report is handed over. So the name is matched by its ending
// against the tests this project has, and only taken relative to the root when
// none of them matches.
func (p *Project) testFile(name string) string {
	slashed := filepath.ToSlash(name)
	best := ""
	for _, m := range p.Modules {
		if p.Layer(m) != LayerTest {
			continue
		}
		if (slashed == m.Path || strings.HasSuffix(slashed, "/"+m.Path)) && len(m.Path) > len(best) {
			best = m.Path
		}
	}
	if best != "" {
		return best
	}
	if filepath.IsAbs(name) {
		if r, err := filepath.Rel(p.Root, name); err == nil {
			return filepath.ToSlash(r)
		}
	}
	return slashed
}

// writtenTo is the line Vitest prints instead of the report when it writes the
// report to a file, which recent versions do by default.
var writtenTo = regexp.MustCompile(`JSON report written to (\S+)`)

// reportBody returns the report itself: the input when it is one, or the file
// the input names.
//
// Whether the JSON reporter prints to standard output or to a file depends on
// the Vitest version and its configuration, and the pipe the documentation
// shows has to work either way. Following the one line Vitest prints is what
// keeps it working, rather than a flag that is right for one version.
func (p *Project) reportBody(data []byte) ([]byte, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && trimmed[0] == '{' {
		return trimmed, nil
	}
	m := writtenTo.FindSubmatch(data)
	if m == nil {
		return trimmed, nil
	}
	file := string(m[1])
	if !filepath.IsAbs(file) {
		file = filepath.Join(p.Root, file)
	}
	body, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read the test results Vitest wrote to %s: %w", file, err)
	}
	return body, nil
}
