package diag

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
)

// Set collects findings. The zero value is ready to use.
type Set struct {
	mu       sync.Mutex
	findings []Finding
}

// Add records a finding.
func (s *Set) Add(f Finding) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.findings = append(s.findings, f)
}

// Len returns the number of findings.
func (s *Set) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.findings)
}

// Empty reports whether the run is clean.
func (s *Set) Empty() bool { return s.Len() == 0 }

// Findings returns a copy sorted by position and then by code, so that output
// is stable across runs regardless of traversal order.
func (s *Set) Findings() []Finding {
	s.mu.Lock()
	out := make([]Finding, len(s.findings))
	copy(out, s.findings)
	s.mu.Unlock()

	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Pos.File != b.Pos.File {
			return a.Pos.File < b.Pos.File
		}
		if a.Pos.Line != b.Pos.Line {
			return a.Pos.Line < b.Pos.Line
		}
		if a.Pos.Col != b.Pos.Col {
			return a.Pos.Col < b.Pos.Col
		}
		return a.Code < b.Code
	})
	return out
}

// WriteText renders findings for humans and editors:
//
//	file:line:col: [CODE] what
//	    why
//	    how
//
// The first line is copy-paste friendly and matches the format editors parse.
func (s *Set) WriteText(w io.Writer) error {
	for _, f := range s.Findings() {
		if _, err := fmt.Fprintf(w, "%s: [%s] %s\n", f.Pos, f.Code, f.What); err != nil {
			return err
		}
		for _, line := range indent(f.Why) {
			if _, err := fmt.Fprintln(w, line); err != nil {
				return err
			}
		}
		for _, line := range indent(f.How) {
			if _, err := fmt.Fprintln(w, line); err != nil {
				return err
			}
		}
	}
	return nil
}

func indent(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	raw := strings.Split(strings.TrimRight(s, "\n"), "\n")
	out := make([]string, 0, len(raw))
	for _, l := range raw {
		out = append(out, "    "+l)
	}
	return out
}

// jsonFinding is the wire shape of a finding. Kept separate from Finding so the
// public JSON contract does not drift with internal refactoring.
type jsonFinding struct {
	Code   string `json:"code"`
	File   string `json:"file"`
	Line   int    `json:"line"`
	Col    int    `json:"col"`
	What   string `json:"what"`
	Why    string `json:"why,omitempty"`
	How    string `json:"how,omitempty"`
	Rule   string `json:"rule,omitempty"`
	Phase  string `json:"phase"`
	Waived bool   `json:"waived,omitempty"`
}

type jsonReport struct {
	Version  int           `json:"version"`
	Findings []jsonFinding `json:"findings"`
}

// ReportVersion is the schema version of the JSON diagnostics.
//
// Unlike the intermediate model this is a genuine external contract: the LLM
// loop consumes it. It is versioned deliberately.
const ReportVersion = 1

// WriteJSON renders findings for the LLM loop. What, Why and How stay separate
// fields so a consumer can act on them individually.
func (s *Set) WriteJSON(w io.Writer) error {
	fs := s.Findings()
	out := jsonReport{Version: ReportVersion, Findings: make([]jsonFinding, 0, len(fs))}
	for _, f := range fs {
		out.Findings = append(out.Findings, jsonFinding{
			Code:  f.Code,
			File:  f.Pos.File,
			Line:  f.Pos.Line,
			Col:   f.Pos.Col,
			What:  f.What,
			Why:   f.Why,
			How:   f.How,
			Rule:  f.Rule,
			Phase: phaseOf(f.Code),
		})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

// phaseOf extracts the phase from a code of the form SPEC-<phase>-<n>.
func phaseOf(code string) string {
	parts := strings.Split(code, "-")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}
