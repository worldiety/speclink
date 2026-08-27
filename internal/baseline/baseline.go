// Package baseline records what has already been promised.
//
// speclink otherwise decides everything from the current source, and for the
// evolution rules that is not enough: no snapshot of a working tree can say
// whether a field used to be an int, or whether a discriminator was ever
// spelled differently. That knowledge has to come from somewhere outside the
// snapshot.
//
// It is deliberately not a second source of intent. Intent stays in the code —
// the field type states the shape, spec.Draft states the status. This file
// records facts: what was actually committed to. It has the same relationship
// to the source that go.sum has to go.mod, and the same rule follows from it:
// it is written by the tool and never edited by hand.
//
// The diff of this file is the moment a shape stops being an experiment.
//
// # One mechanism, three edges
//
// The file started out recording persisted shapes and now records everything
// the Go compiler cannot check. That is not scope creep but the same principle
// reaching its natural extent: a link speclink cannot re-derive from the
// current source needs a record of what it used to mean, or a change to the far
// end goes unreported while the link still resolves.
//
// There are exactly three such edges, and they fail identically.
//
//   - A persisted shape outlives the code declaring it. Recorded since the
//     beginning.
//   - A requirement's text can be rewritten while its identifier and every
//     reference to it stay valid. With a generated requirement tree this is
//     not an edge case but the likeliest defect there is: nothing about a
//     regenerated tree is checked by the compiler, and every satisfier stays
//     green through an arbitrary change of meaning.
//   - A source segment can be rewritten while the anchor citing it still
//     resolves. The literature treats this one as unsolved and answers it by
//     hand — a revision number somebody remembers to increment, a suspect flag
//     somebody remembers to review. Here it is computed.
//
// In all three cases the answer is the same and so is the workflow: the run
// reports what moved, `speclink freeze` records the new state, and the diff of
// this file is what a human actually reviews. That is the point of doing it
// this way rather than with a status field somebody maintains — the moment of
// review is a diff in a pull request, not a discipline.
package baseline

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/ir"
)

// FileName is the baseline in the project root.
const FileName = "speclink.lock"

// Version is the schema version of the file.
//
// A reader that does not know the version must refuse rather than guess:
// comparing against a half understood record would report differences that are
// not there. An *older* known version is a different matter and is upgraded on
// read. This file is written by the tool and never edited by hand, so a version
// bump that stranded existing projects would be asking them to do the one thing
// the file forbids.
//
// Version 2 added the requirement and source records. Upgrading from 1 leaves
// both empty, which is exactly right: nothing has been recorded about them yet,
// so nothing can have drifted, and the next freeze reads them for the first
// time.
const Version = 2

// File is the on disk form. Maps are used so the encoding is stable under
// Go's sorted map marshalling, and so a diff stays readable per type.
type File struct {
	Version int `json:"version"`
	// Types are the promised persisted shapes.
	Types map[string]Entry `json:"types"`
	// Requirements are the recorded requirement texts, keyed by requirement ID.
	Requirements map[string]Requirement `json:"requirements,omitempty"`
	// Sources are the recorded document segments, keyed by segment reference
	// in the form "path/to/doc.md#anchor".
	Sources map[string]Segment `json:"sources,omitempty"`
}

// Requirement is one recorded requirement text.
//
// Text is a hash rather than the words themselves. The words are in the
// .spec.go file, and repeating them here would make this a second source of
// intent, which it must never be. A hash records only the fact that a review
// happened against a particular wording.
//
// From is the segment the requirement was written against. It is what keeps an
// identifier stable across a regeneration of the tree: a generator that reuses
// the ID already recorded for a segment does not rename RQuoteSubmit on every
// run, and every spec.Satisfies in the code keeps compiling. Without it the
// property that a rename is a refactoring rather than a broken link only holds
// for trees written by hand.
type Requirement struct {
	Text  string `json:"text"`
	Title string `json:"title,omitempty"`
	From  string `json:"from,omitempty"`
}

// Segment is one recorded source document segment.
type Segment struct {
	Fingerprint string `json:"fingerprint"`
	// Kind is markdown or image, carried so a diff says what moved without
	// the reader having to look the path up.
	Kind string `json:"kind,omitempty"`
}

// Entry is one promised type.
type Entry struct {
	Discriminator string  `json:"discriminator"`
	Fields        []Field `json:"fields"`
}

// Field is one promised field.
type Field struct {
	Name     string `json:"name"`
	Wire     string `json:"wire"`
	Shape    string `json:"shape"`
	Optional bool   `json:"optional,omitempty"`
}

// Field returns the promised field with the given Go name.
func (e Entry) Field(name string) (Field, bool) {
	for _, f := range e.Fields {
		if f.Name == name {
			return f, true
		}
	}
	return Field{}, false
}

// ByWire returns the promised field carrying the given wire name.
func (e Entry) ByWire(wire string) (Field, bool) {
	for _, f := range e.Fields {
		if f.Wire == wire {
			return f, true
		}
	}
	return Field{}, false
}

// Load reads the baseline from the project root. A missing file is not an
// error: it means nothing has been promised yet.
func Load(root string) (*File, error) {
	f := &File{Version: Version}
	f.fill()

	data, err := os.ReadFile(filepath.Join(root, FileName))
	if errors.Is(err, fs.ErrNotExist) {
		return f, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", FileName, err)
	}
	if err := json.Unmarshal(data, f); err != nil {
		return nil, fmt.Errorf("parse %s: %w", FileName, err)
	}
	if f.Version > Version || f.Version < 1 {
		return nil, fmt.Errorf("%s has version %d, this speclink reads versions 1 to %d", FileName, f.Version, Version)
	}
	f.fill()
	return f, nil
}

// fill replaces the nil maps of a fresh or partially populated record, so
// every caller can write into them without checking first.
func (f *File) fill() {
	if f.Types == nil {
		f.Types = map[string]Entry{}
	}
	if f.Requirements == nil {
		f.Requirements = map[string]Requirement{}
	}
	if f.Sources == nil {
		f.Sources = map[string]Segment{}
	}
}

// Save writes the baseline to the project root.
func (f *File) Save(root string) error {
	f.Version = Version
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, FileName), append(data, '\n'), 0o644)
}

// EntryOf converts a read schema type into the recorded form.
func EntryOf(t ir.SchemaType) Entry {
	e := Entry{Discriminator: t.Discriminator}
	for _, f := range t.Fields {
		e.Fields = append(e.Fields, Field{
			Name:     f.Name,
			Wire:     f.Wire,
			Shape:    f.Shape,
			Optional: f.Optional,
		})
	}
	return e
}

// Names returns the recorded type names in a stable order.
func (f *File) Names() []string { return sortedKeys(f.Types) }

// RequirementIDs returns the recorded requirement IDs in a stable order.
func (f *File) RequirementIDs() []string { return sortedKeys(f.Requirements) }

// SegmentRefs returns the recorded segment references in a stable order.
func (f *File) SegmentRefs() []string { return sortedKeys(f.Sources) }

// RequirementOf returns the ID already recorded for a source segment.
//
// This is the lookup a generator uses to keep identifiers stable: a regenerated
// tree that reuses the recorded ID for a segment leaves every reference in the
// code compiling, where a freshly invented one would break all of them at once.
func (f *File) RequirementOf(segment string) (string, bool) {
	if segment == "" {
		return "", false
	}
	for _, id := range f.RequirementIDs() {
		if f.Requirements[id].From == segment {
			return id, true
		}
	}
	return "", false
}

// HashText reduces a requirement text to what a change to it would have to
// mean.
//
// Leading and trailing whitespace and the line breaks of a wrapped sentence are
// artefacts of how the text was typed, not of what it says. Everything else is
// kept: wording, capitalisation and punctuation all carry meaning in a
// normative sentence, and a hash that forgave them would be quietly deciding
// that some rewrites do not count.
func HashText(parts ...string) string {
	var b strings.Builder
	for i, p := range parts {
		if i > 0 {
			b.WriteByte(0)
		}
		b.WriteString(strings.Join(strings.Fields(p), " "))
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
