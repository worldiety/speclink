package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang"
	"github.com/worldiety/speclink/internal/render"
)

// diagrams writes the drawing sources of the model.
//
// It writes and does not run. PlantUML is a prerequisite of the environment
// rather than a dependency of this program: nothing here invokes it, links it
// or pins a version of it, so a checkout with no Java can still run every rule.
// Turning the .puml files into pictures is one line of a Makefile, and it is
// the caller's line.
//
// The separation earns something beyond tidiness. Two runs of speclink over one
// model produce byte identical sources, which is what makes a diagram in a
// review diff readable at all — and it stays true whatever the renderer does.
func diagrams(args []string) error {
	fs := flag.NewFlagSet("diagrams", flag.ExitOnError)
	var (
		root    = fs.String("root", ".", "repository root")
		cfgPath = fs.String("config", "", "layout configuration; defaults to speclink.json in the root")
		prof    = fs.String("profile", "", "overrides the profile from speclink.json")
		out     = fs.String("out", "build/puml", "directory to write the diagram sources into")
		title   = fs.String("title", "", "name of the system in the drawings; defaults to the directory name")
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

	// Findings are collected and dropped: this command draws what is there and
	// does not judge it. Whether the model is sound is what verify answers, and
	// answering it twice in two voices is how the two come to disagree.
	quiet := &diag.Set{}

	var (
		topo      ir.Topology
		processes []*ir.Process
		pkgs      ir.PackageGraph
	)
	if pg, ok := model.(lang.PackageGrapher); ok {
		pkgs = pg.PackageGraph()
	}
	if tr, ok := model.(lang.TopologyReader); ok {
		topo = tr.Topology(quiet)
	}
	if pr, ok := model.(lang.ProcessReader); ok {
		processes = pr.Processes(quiet)
	}
	if !topo.Declared() && len(processes) == 0 && !pkgs.Declared() {
		return errors.New("nothing to draw: this frontend reports no package graph, and this project\n" +
			"declares no topology and no process. Add a *.topology.go or a *.process.go;\n" +
			"see the README sections on both.")
	}

	if err := os.MkdirAll(*out, 0o755); err != nil {
		return err
	}

	// The same name the document uses, or the drawings label the system with a
	// directory name while the page around them calls it something else.
	system := *title
	if system == "" {
		system = filepath.Base(absRoot)
	}
	written := 0
	if topo.Declared() {
		for name, body := range map[string]string{
			"context.puml": render.Context(topo, system),
			"blocks.puml":  render.Blocks(topo, system),
		} {
			if err := writeDiagram(*out, name, body); err != nil {
				return err
			}
			written++
		}
	}
	if pkgs.Declared() {
		// One picture while it is still readable, and a map plus one picture
		// per context once it is not. The split is decided here rather than by
		// the caller because the caller has no way to judge it: what makes a
		// drawing unreadable is how many nodes end up in it, which only this
		// side knows.
		if !render.Crowded(pkgs) {
			if err := writeDiagram(*out, "packages.puml", render.Packages(pkgs, system)); err != nil {
				return err
			}
			written++
		} else {
			if err := writeDiagram(*out, "context-map.puml", render.ContextMap(pkgs, system)); err != nil {
				return err
			}
			written++
			for _, ctx := range pkgs.Contexts() {
				name := "packages-" + slug(ctx) + ".puml"
				if err := writeDiagram(*out, name, render.PackagesOf(pkgs, ctx, system)); err != nil {
					return err
				}
				written++
			}
		}
	}
	for _, p := range processes {
		// The same graph, pictured the way the declaration asks for. A course
		// that crosses no boundary has nothing to gain from a sequence, which
		// is why the flow drawing stays the default.
		body := render.Process(p)
		if p.Drawn == ir.AsSequence {
			body = render.Sequence(p, participantsByIdent(topo))
		}
		if err := writeDiagram(*out, "process-"+p.ID+".puml", body); err != nil {
			return err
		}
		written++
	}

	fmt.Fprintf(os.Stderr, "wrote %s into %s\n\n", plural(written, "diagram", "diagrams"), *out)
	fmt.Fprint(os.Stderr, "render them with a PlantUML of your choosing, for example:\n"+
		"  plantuml -tsvg "+filepath.Join(*out, "*.puml")+"\n")
	return nil
}

func writeDiagram(dir, name, body string) error {
	return os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644)
}

// participantsByIdent indexes the declared participants by the identifier a
// process names them with.
//
// The identifier and not the ID, because a process names the declaration and
// the compiler checks that: a drawing tied to a string would go on naming a
// participant that had been renamed out from under it.
func participantsByIdent(t ir.Topology) map[string]ir.Participant {
	out := make(map[string]ir.Participant, len(t.Participants))
	for _, p := range t.Participants {
		out[p.GoIdent] = p
	}
	return out
}
