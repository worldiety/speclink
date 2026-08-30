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
	)
	if tr, ok := model.(lang.TopologyReader); ok {
		topo = tr.Topology(quiet)
	}
	if pr, ok := model.(lang.ProcessReader); ok {
		processes = pr.Processes(quiet)
	}
	if !topo.Declared() && len(processes) == 0 {
		return errors.New("nothing to draw: this project declares no topology and no process.\n" +
			"Add a *.topology.go or a *.process.go; see the README sections on both.")
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
	for _, p := range processes {
		if err := writeDiagram(*out, "process-"+p.ID+".puml", render.Process(p)); err != nil {
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
