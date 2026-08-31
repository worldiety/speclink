package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/worldiety/speclink/internal/check"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang"
	"github.com/worldiety/speclink/internal/schema"
)

// schemas writes a JSON Schema for every shape that crosses a boundary.
//
// It writes and does not validate, for the reason diagrams writes and does not
// draw. Nothing here runs a validator, links one or pins a version of one, so a
// checkout with no JavaScript toolchain still runs every rule speclink has.
//
// # Why the schemas are derived rather than written
//
// Every structure emitted here was read from a Go type and is already what the
// rules compare a promise against. A schema written by hand beside those types
// would be the same fact in two places, and the two would disagree the first
// time somebody added a field — which is exactly the failure the schema existed
// to prevent. The types stay the one place; this is a projection of them into
// the format the far end reads.
//
// # Why the vectors are a separate file
//
// A restriction is prose and a schema cannot carry it. The half that matters is
// what must be refused, and JSON Schema expresses that as a negated pattern
// which generators drop without a sound. A list of cases survives every
// generator, every language and every team, because "this must be rejected"
// means the same thing everywhere.
func schemas(args []string) error {
	fs := flag.NewFlagSet("schema", flag.ExitOnError)
	var (
		root    = fs.String("root", ".", "repository root")
		cfgPath = fs.String("config", "", "layout configuration; defaults to speclink.json in the root")
		prof    = fs.String("profile", "", "overrides the profile from speclink.json")
		out     = fs.String("out", "build/schema", "directory to write the schemas into")
	)
	if err := fs.Parse(args); err != nil {
		return err
	}

	absRoot, err := absRootOf(*root)
	if err != nil {
		return err
	}
	model, _, _, err := open(absRoot, *cfgPath, *prof, fs.Args(), false)
	if err != nil {
		return err
	}

	// Findings are collected and dropped: this command writes what is there and
	// does not judge it. Whether the model is sound is what verify answers.
	quiet := &diag.Set{}

	var topo ir.Topology
	if tr, ok := model.(lang.TopologyReader); ok {
		topo = tr.Topology(quiet)
	}
	restricted := restrictionsByType(model, quiet)

	shapes := crossingShapes(topo)
	if len(shapes) == 0 && len(restricted) == 0 {
		return errors.New("nothing to write: no channel of this project states a contract or a\n" +
			"protocol, and no type states a rule about its values. See the README\n" +
			"sections on channels and on restricted values.")
	}

	if err := os.MkdirAll(*out, 0o755); err != nil {
		return err
	}

	written := 0
	for _, s := range shapes {
		doc, err := schema.Of(s, restricted)
		if err != nil {
			// A shape speclink itself wrote and cannot read back is a defect
			// in this program, not in the project. Say which shape, or the
			// report is unusable.
			return fmt.Errorf("cannot render %s: %w", s.Type, err)
		}
		body, err := schema.Marshal(doc.Body)
		if err != nil {
			return err
		}
		if err := writeFile(*out, doc.Name+".schema.json", body); err != nil {
			return err
		}
		written++
	}

	if len(restricted) > 0 {
		body, err := schema.Marshal(schema.Vectors(sortedRestrictions(restricted)))
		if err != nil {
			return err
		}
		if err := writeFile(*out, "vectors.json", body); err != nil {
			return err
		}
		written++
	}

	fmt.Fprintf(os.Stderr, "wrote %s into %s\n", plural(written, "file", "files"), *out)
	return nil
}

// crossingShapes is every structure that leaves this program, in a stable
// order.
//
// A contract and a message are the same thing to a reader on the far end: a
// shape they have to produce or parse. They are collected together and
// deduplicated by type, because one type may cross more than one boundary and
// two identical files with different names would leave a reader wondering
// which is authoritative.
func crossingShapes(t ir.Topology) []schema.Shape {
	seen := map[string]schema.Shape{}
	add := func(w *ir.WireShape) {
		if w == nil || w.Type == "" {
			return
		}
		if _, ok := seen[w.Type]; ok {
			return
		}
		s := schema.Shape{
			Type:       w.Type,
			Shape:      w.Shape,
			Optional:   map[string]bool{},
			FieldTypes: map[string]string{},
		}
		for _, f := range w.Fields {
			// Both spellings of absence. One is a decision somebody declared,
			// the other a fact about the encoding; either way the far end may
			// receive a payload without the field, and a required list that
			// ignored the difference would be wrong in the dangerous
			// direction.
			if f.Optional || f.OmitEmpty {
				s.Optional[f.Wire] = true
			}
			if f.Type != "" {
				s.FieldTypes[f.Wire] = f.Type
			}
		}
		seen[w.Type] = s
	}

	for _, c := range t.Channels {
		add(c.Contract)
		add(c.Envelope)
		for _, m := range c.Messages {
			add(m.Payload)
		}
	}

	out := make([]schema.Shape, 0, len(seen))
	for _, s := range seen {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Type < out[j].Type })
	return out
}

// restrictionsByType indexes the rules types state about their own values.
func restrictionsByType(model lang.Model, quiet *diag.Set) map[string]schema.Restriction {
	out := map[string]schema.Restriction{}
	inf, ok := model.(lang.ConstructInferrer)
	if !ok {
		return out
	}
	_ = inf
	for _, r := range check.Restrictions(model.Bindings(quiet), quiet) {
		out[r.Type] = schema.Restriction{
			Type: r.Type, Rule: r.Text, Valid: r.Valid, Invalid: r.Invalid,
		}
	}
	return out
}

func sortedRestrictions(m map[string]schema.Restriction) []schema.Restriction {
	out := make([]schema.Restriction, 0, len(m))
	for _, r := range m {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Type < out[j].Type })
	return out
}

func writeFile(dir, name string, body []byte) error {
	return os.WriteFile(filepath.Join(dir, name), body, 0o644)
}
