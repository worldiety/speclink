// Package baseline records what has already been promised.
//
// speclink otherwise decides everything from the current source, and for the
// evolution rules that is not enough: no snapshot of a working tree can say
// whether a field used to be an int, or whether a discriminator was ever
// spelled differently. That knowledge has to come from somewhere outside the
// snapshot.
//
// It is deliberately not a second source of intent. Intent stays in the code —
// the field type states the shape, spec.Proposal states the status. This file
// records facts: what was actually committed to. It has the same relationship
// to the source that go.sum has to go.mod, and the same rule follows from it:
// it is written by the tool and never edited by hand.
//
// The diff of this file is the moment a shape stops being an experiment.
package baseline

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/worldiety/speclink/internal/ir"
)

// FileName is the baseline in the project root.
const FileName = "speclink.lock"

// Version is the schema version of the file. A reader that does not know the
// version must refuse rather than guess: comparing against a half understood
// record would report differences that are not there.
const Version = 1

// File is the on disk form. Maps are used so the encoding is stable under
// Go's sorted map marshalling, and so a diff stays readable per type.
type File struct {
	Version int              `json:"version"`
	Types   map[string]Entry `json:"types"`
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
	f := &File{Version: Version, Types: map[string]Entry{}}

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
	if f.Version != Version {
		return nil, fmt.Errorf("%s has version %d, this speclink writes version %d", FileName, f.Version, Version)
	}
	if f.Types == nil {
		f.Types = map[string]Entry{}
	}
	return f, nil
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
func (f *File) Names() []string {
	out := make([]string, 0, len(f.Types))
	for name := range f.Types {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
