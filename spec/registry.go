package spec

import (
	"encoding/json"
	"io"
	"sort"
	"sync"
)

// DumpVersion is the schema version of the JSON produced by [DumpJSON].
//
// This is the one place in speclink where genuine version skew is possible: the
// target project pins speclink/spec in its go.mod, while the developer runs an
// arbitrary speclink binary. The reader must refuse a mismatching version with
// a clear message rather than report an incomplete comparison as a finding.
//
// The intermediate model inside speclink needs no version: one binary, one
// process, no boundary.
const DumpVersion = 1

var (
	registryMu sync.Mutex
	registry   []Entry
)

// Entry is one recorded binding term with its source position.
type Entry struct {
	File       string           `json:"file"`
	Line       int              `json:"line"`
	TargetKind string           `json:"targetKind"`
	Target     string           `json:"target,omitempty"`
	Assertions []EntryAssertion `json:"assertions,omitempty"`
}

// EntryAssertion is one recorded assertion. Only the fields relevant for its
// kind are populated.
type EntryAssertion struct {
	Kind         string   `json:"kind"`
	Requirements []string `json:"requirements,omitempty"`
	EventType    string   `json:"eventType,omitempty"`
	State        string   `json:"state,omitempty"`
	Text         string   `json:"text,omitempty"`
	Term         string   `json:"term,omitempty"`
	Rule         string   `json:"rule,omitempty"`
}

// Dump is the envelope written by [DumpJSON].
type Dump struct {
	Version int     `json:"version"`
	Entries []Entry `json:"entries"`
}

func entryOf(t target, as []Assertion, file string, line int) Entry {
	e := Entry{File: file, Line: line, TargetKind: t.kind.String(), Target: t.name()}
	for _, a := range as {
		e.Assertions = append(e.Assertions, a.entry())
	}
	return e
}

func (a Assertion) entry() EntryAssertion {
	ea := EntryAssertion{Kind: a.kind.String()}
	switch a.kind {
	case kindSatisfies:
		for _, r := range a.reqs {
			ea.Requirements = append(ea.Requirements, string(r.ID))
		}
	case kindTransition:
		ea.EventType = typeName(a.eventType)
		ea.State = string(a.state)
	case kindHelp, kindRationale:
		ea.Text = a.text
	case kindTerm:
		ea.Term = string(a.term.ID)
	case kindWaive:
		ea.Rule = string(a.rule)
		ea.Text = a.text
	}
	return ea
}

func (k assertionKind) String() string {
	switch k {
	case kindSatisfies:
		return "satisfies"
	case kindTransition:
		return "transition"
	case kindExternal:
		return "external"
	case kindHelp:
		return "help"
	case kindTerm:
		return "term"
	case kindRationale:
		return "rationale"
	case kindWaive:
		return "waive"
	case kindDraft:
		return "draft"
	case kindOptional:
		return "optional"
	case kindPersistence:
		return "persistence"
	}
	return "unknown"
}

func appendEntry(e Entry) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry = append(registry, e)
}

// Entries returns a copy of the registry, sorted by position.
//
// The sort makes the output stable for diffing. It does not make the model
// order significant: Go initialises package level variables in dependency
// order, so the recording order is deterministic but unrelated to source order.
func Entries() []Entry {
	registryMu.Lock()
	out := make([]Entry, len(registry))
	copy(out, registry)
	registryMu.Unlock()

	sort.Slice(out, func(i, j int) bool {
		if out[i].File != out[j].File {
			return out[i].File < out[j].File
		}
		return out[i].Line < out[j].Line
	})
	return out
}

// DumpJSON writes the recorded model as JSON. It is the entry point used by
// `speclink selfreport`, which links the target packages, runs their package
// initialisation and compares the result against the statically read model.
func DumpJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(Dump{Version: DumpVersion, Entries: Entries()})
}
