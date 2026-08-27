package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/baseline"
	"github.com/worldiety/speclink/internal/check"
	"github.com/worldiety/speclink/internal/config"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang/golang"
	"github.com/worldiety/speclink/internal/reqtree"
	"github.com/worldiety/speclink/spec"
)

// evidence records which tests actually demonstrated which requirements.
//
// It is the second half of spec.Verified, and the half that makes the first one
// worth anything. Reading the call out of the source proves that somebody wrote
// it down; only running it proves that anything happened. A call can sit behind
// a condition that never holds, or in a test that fails long before reaching
// it, and neither is distinguishable from a working test by any amount of
// static analysis.
//
// So the record here is a record of a moment. Every other entry in
// speclink.lock is a hash of text that can be read again at any time; this one
// cannot be reconstructed from the working tree at all, which is precisely why
// it has to be written down.
//
// It reads `go test -json` rather than running it. speclink does not run tests:
// the build order is Go compiler, then speclink, then tests, and a command that
// invoked the test suite would either violate that or duplicate it. It also
// makes the evidence something a CI system hands over rather than something
// speclink produces, which is the right way round for evidence.
//
// Only passing tests are recorded. A test that wrote the line and then failed
// claimed something it did not show.
func evidence(args []string) error {
	fs := flag.NewFlagSet("evidence", flag.ExitOnError)
	root := fs.String("root", ".", "repository root, holding "+baseline.FileName)
	cfgPath := fs.String("config", "", "layout configuration; defaults to "+config.FileName+" in the root")
	in := fs.String("in", "", "`file` holding the output of \"go test -json\"; standard input by default")
	dry := fs.Bool("n", false, "report what would be recorded, write nothing")
	if err := fs.Parse(args); err != nil {
		return err
	}

	absRoot, err := filepath.Abs(*root)
	if err != nil {
		return fmt.Errorf("resolve root: %w", err)
	}
	layout, err := loadLayout(absRoot, *cfgPath)
	if err != nil {
		return err
	}

	source := io.Reader(os.Stdin)
	if *in != "" {
		f, err := os.Open(*in)
		if err != nil {
			return err
		}
		defer f.Close()
		source = f
	}

	demonstrated, err := readTestOutput(source)
	if err != nil {
		return err
	}

	// The requirement tree is loaded to turn IDs back into requirements: the
	// record is bound to the wording a test ran against, and the wording lives
	// in the tree.
	discard := &diag.Set{}
	pkgs, err := golang.Load(absRoot, fs.Args()...)
	if err != nil {
		return err
	}
	if errs := golang.TypeErrors(pkgs); len(errs) > 0 {
		fmt.Fprintln(os.Stderr, "the Go build is broken; fix it before speclink can record anything:")
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, "  "+e.Error())
		}
		return errFindings
	}
	var reqs []*ir.Requirement
	for _, p := range pkgs {
		reqs = append(reqs, p.ReadRequirements(discard)...)
	}
	tree := reqtree.Build(absRoot, reqs, discard)
	_ = layout

	base, err := baseline.Load(absRoot)
	if err != nil {
		return err
	}

	changed, unknown := check.RecordVerifications(base, tree, demonstrated)
	for _, id := range unknown {
		// A record naming a requirement the tree does not have is a defect
		// worth saying out loud rather than skipping: it means a test is
		// verifying something that has been renamed or removed, and dropping it
		// silently would leave the test looking useful.
		fmt.Fprintf(os.Stderr, "ignored   %s: no such requirement\n", id)
	}
	if len(changed) == 0 {
		fmt.Fprintln(os.Stderr, "nothing to record; the record already matches this run.")
		return nil
	}
	for _, line := range changed {
		fmt.Fprintln(os.Stderr, line)
	}
	if *dry {
		fmt.Fprintln(os.Stderr, "\nnothing written (-n).")
		return nil
	}
	if err := base.Save(absRoot); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "\nrecorded in %s.\n", baseline.FileName)
	return nil
}

// testEvent is the part of the `go test -json` stream this needs.
type testEvent struct {
	Action string `json:"Action"`
	Test   string `json:"Test"`
	Output string `json:"Output"`
}

// readTestOutput collects the requirements each passing test demonstrated.
//
// Attribution comes from the stream rather than from the line, which is why
// spec.Verified writes through the test's own logger: go test tags every output
// event with the test that produced it, and without that a record could not be
// tied to a pass or a failure.
func readTestOutput(r io.Reader) (map[string][]string, error) {
	var (
		claimed = map[string][]string{}
		passed  = map[string]bool{}
		seen    bool
	)

	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for s.Scan() {
		var e testEvent
		if err := json.Unmarshal(s.Bytes(), &e); err != nil {
			// Not every line of the stream is ours to understand; a build
			// failure prints plain text into it.
			continue
		}
		seen = true

		switch e.Action {
		case "pass":
			if e.Test != "" {
				passed[e.Test] = true
			}
		case "output":
			if e.Test == "" {
				continue
			}
			ids, err := parseVerifiedLine(e.Output)
			if err != nil {
				return nil, err
			}
			claimed[e.Test] = append(claimed[e.Test], ids...)
		}
	}
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("read test output: %w", err)
	}
	if !seen {
		return nil, errors.New(`no "go test -json" events on the input; pipe "go test -json ./..." into this command, or point -in at its output`)
	}

	out := map[string][]string{}
	for test, ids := range claimed {
		// A test that claimed something and then failed showed nothing. The
		// claim is still in the source, so K14-VERIFICATION-STALE will report
		// it; recording it here would make the failure invisible instead.
		if !passed[test] {
			continue
		}
		for _, id := range ids {
			out[id] = appendUnique(out[id], test)
		}
	}
	return out, nil
}

// parseVerifiedLine extracts the requirement IDs from one output line.
//
// The marker is looked for anywhere in the line, not at its start, because
// testing prefixes its output with the file and line it came from.
func parseVerifiedLine(line string) ([]string, error) {
	i := strings.Index(line, spec.VerifiedMarker)
	if i < 0 {
		return nil, nil
	}
	payload := strings.TrimSpace(line[i+len(spec.VerifiedMarker):])

	var record struct {
		Version int      `json:"v"`
		Reqs    []string `json:"reqs"`
	}
	if err := json.Unmarshal([]byte(payload), &record); err != nil {
		return nil, fmt.Errorf("unreadable verification line %q: %w", payload, err)
	}
	// The project pins speclink/spec in its go.mod while the developer runs an
	// arbitrary speclink binary, so this is one of the few places where genuine
	// version skew is possible. Refusing is the only safe answer: recording
	// nothing looks exactly like a test that was never written.
	if record.Version != spec.VerifiedVersion {
		return nil, fmt.Errorf("the tests were built against spec.Verified version %d, this speclink reads version %d; align the speclink/spec requirement in go.mod with the binary",
			record.Version, spec.VerifiedVersion)
	}
	return record.Reqs, nil
}

func appendUnique(list []string, s string) []string {
	for _, have := range list {
		if have == s {
			return list
		}
	}
	return append(list, s)
}

func sortedStrings(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}
