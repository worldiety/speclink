package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/worldiety/speclink/internal/profile"
)

// init writes a starting point for a new project.
//
// It asks nothing interactively. The caller is usually an agent, for which a
// prompt is the worst possible interface: it has to notice that one appeared,
// parse prose to learn what is wanted, and answer in the right order, and any
// mismatch hangs the process. So init behaves like the rest of speclink and
// refuses with a message that names the alternatives — the same shape as the
// profile menu, which a caller can read, choose from, and retry.
//
// -describe is the same catalogue for a caller that would rather plan than
// fail first.
func initialise(args []string) error {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	var (
		profileName = fs.String("profile", "", "the profile to start from")
		template    = fs.String("template", "", "the starting point within the profile")
		module      = fs.String("module", "", "the Go module path of the new project")
		context     = fs.String("context", "", "the name of the first bounded context")
		dir         = fs.String("dir", ".", "the directory to write into, which must be empty")
		describe    = fs.Bool("describe", false, "print the available profiles, templates and parameters")
		format      = fs.String("format", "text", "output format for -describe: text or json")
	)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *describe {
		return describeTemplates(*profileName, *format)
	}

	p, err := profile.Get(*profileName)
	if err != nil {
		if *profileName == "" {
			return fmt.Errorf("no profile chosen; pass -profile.\n\n%s", profile.List())
		}
		return err
	}

	t, err := p.Template(*template)
	if err != nil {
		return err
	}

	values, err := t.Bind(map[string]string{"module": *module, "context": *context})
	if err != nil {
		return err
	}
	if err := t.Render(*dir, values); err != nil {
		return err
	}

	// The next steps are printed because the first verify of a new project
	// legitimately fails: nothing has been tested and no shape has been
	// recorded. A caller that did not expect that would read a correct report
	// as a broken template.
	fmt.Fprintf(os.Stderr, "wrote %s into %s\n\n", t.Name, *dir)
	fmt.Fprint(os.Stderr, ""+
		"next:\n"+
		"  go mod tidy\n"+
		"  go build ./...\n"+
		"  go test -json ./... | speclink evidence\n"+
		"  speclink freeze\n"+
		"  speclink verify\n\n"+
		"verify reports findings until evidence and freeze have run once. That is\n"+
		"the truth about a project where nothing has been tested or approved yet.\n"+
		"AGENTS.md in the new project says the rest.\n")
	return nil
}

// describeTemplates prints the catalogue.
func describeTemplates(name, format string) error {
	profiles := profile.All()
	if name != "" {
		p, err := profile.Get(name)
		if err != nil {
			return err
		}
		profiles = []*profile.Profile{p}
	}

	switch format {
	case "json":
		return json.NewEncoder(os.Stdout).Encode(describeJSON(profiles))
	case "text":
		for _, p := range profiles {
			fmt.Printf("%s\n  %s\n", p.Name, p.Summary)
			if len(p.Templates()) == 0 {
				fmt.Println("  no templates; this profile can check a project but not start one")
			}
			for _, t := range p.Templates() {
				fmt.Printf("  template %s\n    %s\n    %s\n", t.Name, t.Purpose, t.Scope)
				for _, param := range t.Params {
					fmt.Printf("    -%-8s %s, for example %q\n", param.Name, param.Description, param.Example)
				}
			}
			fmt.Println()
		}
		return nil
	default:
		return errors.New("unknown format " + format + ", expected text or json")
	}
}

// describeJSON is the machine readable catalogue.
//
// It is a type of its own rather than the profile structs so that the wire
// shape is decided here: Profile carries an Open function and a layout, neither
// of which means anything to a caller planning an init.
type describedProfile struct {
	Name      string              `json:"name"`
	Summary   string              `json:"summary"`
	Templates []describedTemplate `json:"templates"`
}

type describedTemplate struct {
	Name    string           `json:"name"`
	Purpose string           `json:"purpose"`
	Scope   string           `json:"scope"`
	Params  []describedParam `json:"params"`
}

type describedParam struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Example     string `json:"example"`
	Required    bool   `json:"required"`
}

func describeJSON(profiles []*profile.Profile) map[string]any {
	out := make([]describedProfile, 0, len(profiles))
	for _, p := range profiles {
		d := describedProfile{Name: p.Name, Summary: p.Summary, Templates: []describedTemplate{}}
		for _, t := range p.Templates() {
			dt := describedTemplate{Name: t.Name, Purpose: t.Purpose, Scope: t.Scope, Params: []describedParam{}}
			for _, param := range t.Params {
				dt.Params = append(dt.Params, describedParam{
					Name:        param.Name,
					Description: param.Description,
					Example:     param.Example,
					// Every parameter is required. Optional ones would be
					// defaults, and a default that shapes a whole project is a
					// decision taken on the caller's behalf without saying so.
					Required: true,
				})
			}
			dt.Params = append([]describedParam{}, dt.Params...)
			d.Templates = append(d.Templates, dt)
		}
		out = append(out, d)
	}
	return map[string]any{"profiles": out}
}
