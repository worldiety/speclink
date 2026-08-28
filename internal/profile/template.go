package profile

import (
	"embed"
	"errors"
	"fmt"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"golang.org/x/mod/module"
)

// templates holds the starting points, compiled into the binary.
//
// The all: prefix is required because the context directory is named __ctx__
// and go:embed skips names beginning with an underscore. Compiling them in
// rather than fetching them means the templates cannot drift from the rules
// that judge them: one binary carries both, and the walkthrough test renders
// the template and runs the whole pipeline over the result.
//
//go:embed all:templates
var templates embed.FS

// Param is one thing a template cannot derive and will not guess.
//
// Anything that follows from another answer is computed instead of asked. The
// binary name is the last segment of the module path, so it is not a parameter;
// asking for it would invite an answer that contradicts the import path.
type Param struct {
	// Name is the flag, without the dash.
	Name string
	// Description says what the value means, in one line.
	Description string
	// Example is a value that would be accepted, so that the error for a
	// missing parameter shows the shape rather than only the name.
	Example string
	// Validate rejects a value the template cannot use. It runs before
	// anything is written, because a project that is half created is worse
	// than one that was refused.
	Validate func(string) error
}

// Template is one starting point within a profile.
//
// A profile can have several because transports differ: a command line tool
// that is handed foundation/rest and app/<context>/rest begins by deleting
// them, and a starting point whose first instruction is to remove code teaches
// the wrong thing about what the layout means.
//
// They are a named, approved set rather than a composition mechanism. cli, rest
// and full are three templates somebody reviewed; a flag-driven combination
// would be 2^n arrangements nobody has ever seen compile.
type Template struct {
	// Name is what -template pins.
	Name string
	// Purpose is the one line shown in the menu.
	Purpose string
	// Scope says what the rendered project contains, for the reader deciding
	// between two templates.
	Scope string
	// Params are the values it needs.
	Params []Param
	// Dir is the path within the embedded file system.
	Dir string
}

// Templates returns the profile's starting points, ordered by name.
func (p *Profile) Templates() []Template {
	out := append([]Template(nil), p.templates...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Template returns one starting point by name.
func (p *Profile) Template(name string) (Template, error) {
	if len(p.templates) == 0 {
		return Template{}, fmt.Errorf("profile %s has no templates yet; it can check a project but cannot start one", p.Name)
	}
	for _, t := range p.templates {
		if t.Name == name {
			return t, nil
		}
	}
	if name == "" {
		return Template{}, fmt.Errorf("no template chosen.\n\n%s", p.TemplateList())
	}
	return Template{}, fmt.Errorf("profile %s has no template %q.\n\n%s", p.Name, name, p.TemplateList())
}

// TemplateList renders the menu, in the same shape as the profile menu.
//
// There is deliberately no default even when a profile has exactly one
// template. A default would make adding a second one a breaking change for
// every caller that relied on it, and naming the template costs one flag.
func (p *Profile) TemplateList() string {
	var b strings.Builder
	fmt.Fprintf(&b, "templates for %s:\n", p.Name)
	for _, t := range p.Templates() {
		fmt.Fprintf(&b, "  %-10s %s\n", t.Name, t.Purpose)
	}
	return b.String()
}

// Values are the answers, plus what follows from them.
type Values struct {
	// Module is the Go module path of the generated project.
	Module string
	// Context is the name of its first bounded context, which is also the
	// package name and the directory under app/.
	Context string
	// Binary is the command name, derived from the module path.
	Binary string
}

// Bind checks the answers and computes the derived ones.
func (t Template) Bind(answers map[string]string) (Values, error) {
	for _, p := range t.Params {
		v, ok := answers[p.Name]
		if !ok || v == "" {
			return Values{}, fmt.Errorf("-%s is required: %s, for example %q", p.Name, p.Description, p.Example)
		}
		if p.Validate != nil {
			if err := p.Validate(v); err != nil {
				return Values{}, fmt.Errorf("-%s: %w", p.Name, err)
			}
		}
	}

	v := Values{Module: answers["module"], Context: answers["context"]}
	v.Binary = path.Base(v.Module)
	return v, nil
}

// contextPlaceholder is the directory standing for the bounded context. A
// literal {{.Context}} on disk would work but reads badly in a file listing and
// invites a tool to try to parse it.
const contextPlaceholder = "__ctx__"

// validModule rejects a module path the go tool would not accept.
//
// It is checked here rather than left to go mod tidy because the path is
// written into every import of the generated project: a bad one fails as
// hundreds of unresolved imports, which names the symptom and not the cause.
func validModule(s string) error {
	if err := module.CheckPath(s); err != nil {
		return fmt.Errorf("%q is not a usable module path: %w", s, err)
	}
	return nil
}

// reserved are package names the generated project already uses.
//
// The entry point imports the context alongside the foundation packages and the
// adapter, so a context called rest or fs would collide there. The collision is
// real but its error message points at main.go, which is the one file the
// author of the name never looked at.
var reserved = map[string]bool{
	"main": true, "auth": true, "permission": true, "data": true,
	"rest": true, "flag": true, "fs": true, "cli": true,
	"record": true, "dec": true, "spec": true,
}

// validPackage rejects a context name that is not a usable Go package name.
func validPackage(s string) error {
	for i, r := range s {
		switch {
		case unicode.IsLower(r) && unicode.IsLetter(r):
		case unicode.IsDigit(r) && i > 0:
		default:
			return fmt.Errorf("%q is not a usable package name: use lower case letters and digits, starting with a letter", s)
		}
	}
	if s == "" {
		return errors.New("a context name is required")
	}
	if token.Lookup(s).IsKeyword() {
		return fmt.Errorf("%q is a Go keyword", s)
	}
	if reserved[s] {
		return fmt.Errorf("%q is already the name of a package in the generated project; choose a name for what this context is about, such as sales or billing", s)
	}
	return nil
}

// Render writes the template into dir, which must be empty.
//
// Hidden entries are tolerated so that git init may run first: refusing a
// directory because it contains .git would force the caller into an order that
// has no reason to be that way round.
func (t Template) Render(dir string, v Values) error {
	if err := requireEmpty(dir); err != nil {
		return err
	}

	root := t.Dir
	return fs.WalkDir(templates, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		// Every file is a template, including its path: the context name is a
		// directory, a package clause and an import path all at once.
		rel = strings.ReplaceAll(rel, contextPlaceholder, v.Context)
		rel = strings.ReplaceAll(rel, "__bin__", v.Binary)
		rel = strings.TrimSuffix(rel, ".tmpl")

		body, err := templates.ReadFile(p)
		if err != nil {
			return err
		}
		out, err := expand(rel, string(body), v)
		if err != nil {
			return err
		}

		target := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, out, 0o644)
	})
}

// expand runs one file through the template engine.
func expand(name, body string, v Values) ([]byte, error) {
	// Missing keys are an error rather than an empty string. A typo in a
	// placeholder would otherwise render as a blank import path, and the
	// resulting failure would name a compiler error rather than the typo.
	tpl, err := template.New(name).Option("missingkey=error").Parse(body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	var b strings.Builder
	if err := tpl.Execute(&b, v); err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return []byte(b.String()), nil
}

// requireEmpty refuses to write into a directory that already holds something.
//
// Overwriting is not offered. The one case where it would help is re-running
// init after a mistake, and the cost of getting it wrong is somebody's work.
func requireEmpty(dir string) error {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return os.MkdirAll(dir, 0o755)
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		return fmt.Errorf("%s is not empty: it contains %s.\nRun init in an empty directory; hidden entries such as .git are allowed", dir, e.Name())
	}
	return nil
}
