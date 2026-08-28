package jvm

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// ReadVerifications collects the claims tests make about requirements.
//
// The Go side writes a line when it runs, and the placement carries meaning:
// spec.Verified at the end of a test says the test got there. Here the claim is
// an annotation on the method, which cannot be put in a branch that never
// executes and therefore needs no placement discipline at all.
//
// It is still only a claim. The evidence is the test report a build writes
// anyway — Surefire, Failsafe and Gradle all record every test and its result —
// so there is nothing to run, no extension to install and no library to depend
// on. That is a better trade than the Go form got: there, proving a test
// actually ran meant putting code in it.
func (r *Reader) ReadVerifications(out *diag.Set) []ir.Binding {
	var bindings []ir.Binding

	for _, c := range r.classes {
		if c.IsSynthetic() {
			continue
		}
		for _, m := range c.Methods {
			if !declared(m) {
				continue
			}
			a, ok := find(m.Annotations, r.annotationType("Verifies"))
			if !ok {
				continue
			}
			file, line := r.pos.Of(c, m.Name)
			if m.Line > 0 {
				line = m.Line
			}
			pos := ir.Position{File: file, Line: line, Col: 1}

			bindings = append(bindings, ir.Binding{
				// The name is what a test report calls it, so the claim and the
				// result can be matched without either side guessing.
				Target: ir.Target{Kind: ir.TargetFunc, Package: c.Package(), Name: TestName(c.Name, m.Name)},
				Assertions: []ir.Assertion{{
					Kind:         ir.AssertVerified,
					Requirements: classRefs(a.Values["value"]),
					Pos:          pos,
				}},
				Pos: pos,
			})
		}
	}
	return bindings
}

// TestName renders a test the way a report names it: the class, then the
// method, separated by a hash.
//
// Both sides have to agree on this and neither can be asked. A claim read from
// bytecode knows the class and the method; a report knows the same two, spelled
// the same way. Anything else — a descriptor, a package alias — would be
// something one side had and the other did not.
func TestName(class, method string) string { return class + "#" + method }

// Verifications implements lang.VerificationReader.
func (m *Model) Verifications(out *diag.Set) []ir.Binding { return m.r.ReadVerifications(out) }

// ReportRoots are where the build tools leave their test reports.
var ReportRoots = []string{
	"target/surefire-reports",
	"target/failsafe-reports",
	"build/test-results/test",
	"build/test-results/testDebugUnitTest",
}

// surefireSuite is the part of a JUnit XML report this needs.
//
// The format is not specified anywhere normative — it grew out of Ant and every
// runner writes a dialect of it — so only the fields that all of them agree on
// are read: the case, its class, its name, and whether a failure or an error
// element sits inside it.
type surefireSuite struct {
	Cases []struct {
		Class   string    `xml:"classname,attr"`
		Name    string    `xml:"name,attr"`
		Failure *struct{} `xml:"failure"`
		Error   *struct{} `xml:"error"`
		Skipped *struct{} `xml:"skipped"`
	} `xml:"testcase"`
}

// ReadTestReports returns the tests that passed, named as [TestName] names them.
//
// Only passes. A test that claimed something and then failed showed nothing,
// and recording it would make the failure invisible — the claim is still in the
// bytecode, so the run reports it as claimed but never demonstrated, which is
// the accurate description.
//
// A skipped test is not a pass either. It is the case the Go side gets for free
// by writing its line at run time and the one this form has to be careful
// about: a report lists a skipped test exactly like a passing one except for a
// single element.
func ReadTestReports(root string, roots []string) (map[string]bool, []error) {
	if len(roots) == 0 {
		roots = ReportRoots
	}
	var (
		passed = map[string]bool{}
		errs   []error
		seen   bool
	)

	for _, rel := range roots {
		dir := filepath.Join(root, rel)
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			continue
		}
		seen = true

		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".xml") {
				return err
			}
			f, openErr := os.Open(path)
			if openErr != nil {
				errs = append(errs, openErr)
				return nil
			}
			defer f.Close()

			if parseErr := readSuite(f, passed); parseErr != nil {
				errs = append(errs, fmt.Errorf("%s: %w", rel, parseErr))
			}
			return nil
		})
		if err != nil {
			errs = append(errs, err)
		}
	}

	if !seen {
		errs = append(errs, fmt.Errorf("no test reports under %s; run the tests first, or set reportRoots", strings.Join(roots, ", ")))
	}
	return passed, errs
}

func readSuite(r io.Reader, passed map[string]bool) error {
	var suite surefireSuite
	if err := xml.NewDecoder(r).Decode(&suite); err != nil {
		return err
	}
	for _, c := range suite.Cases {
		if c.Failure != nil || c.Error != nil || c.Skipped != nil {
			continue
		}
		passed[TestName(c.Class, c.Name)] = true
	}
	return nil
}

// Demonstrations turns claims and results into what actually happened.
//
// It is the join the Go side gets from one line of test output: a claim without
// a matching pass is a test that exists and did not show anything, and the two
// have to be told apart by whoever reads the report.
func Demonstrations(claims []ir.Binding, passed map[string]bool, tree Requirements) map[string][]string {
	out := map[string][]string{}
	for _, b := range claims {
		if !passed[b.Target.Name] {
			continue
		}
		for _, a := range b.Assertions {
			if a.Kind != ir.AssertVerified {
				continue
			}
			for _, ref := range a.Requirements {
				if id, ok := tree.IDOf(ref); ok {
					out[id] = appendOnce(out[id], b.Target.Name)
				}
			}
		}
	}
	return out
}

// Requirements resolves a reference to a requirement ID.
//
// Narrow on purpose: this needs one lookup and taking the whole tree would make
// the frontend depend on a package it otherwise has no business knowing.
type Requirements interface {
	IDOf(ref string) (string, bool)
}

func appendOnce(list []string, s string) []string {
	for _, have := range list {
		if have == s {
			return list
		}
	}
	return append(list, s)
}
