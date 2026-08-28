// Package profile names the combinations speclink actually supports.
//
// Language, framework and architectural style look like three axes and are not.
// They are a chain, and the code has always said so: the architecture rules ask
// the framework's recognisers what a use case is, and the recognisers ask the
// reader what a declaration is. A rule about bounded contexts cannot be written
// without knowing what a use case looks like in this framework, and that cannot
// be known without a reader for the language.
//
// So there is no go × springboot cell that nobody filled in — there is no such
// cell. What existed instead was a name that implied a product: -lang go said
// "the Go frontend" and silently meant "Go, on nago, in the one architectural
// style these rules were written for". Two of those three were never chosen and
// never stated.
//
// A profile states all three, in one name a project pins.
//
// # What the third part is
//
// A style, not an organisation. The rules under K5 to K8 are not nago's — nago
// says nothing about whether a use case lives in uc_<name>.go, or whether a
// context bundles its use cases in a UseCases struct. They are an architecture:
// a subset of DDD in three layers, extended, combined with a functional style
// adapted to what Go allows.
//
// That is why they have to be selectable. The reference ERP follows a different
// convention — 193 use case constructors named with a UC suffix rather than
// New<Type>, and 45 files laid out to match — and under one fixed style that
// reads as 238 defects. It is not 238 defects. It is a second style, and until
// there was a name for the first one there was no way to say so.
//
// # Numbers rather than versions
//
// A style is approved as a whole and pinned as a whole, so ddd1 and ddd2 are two
// styles rather than two versions of one. There is no compatibility relation
// between them to express, and semantic versioning would imply there is.
package profile

import (
	"fmt"
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/config"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/lang"
)

// Language is the reader a profile uses.
type Language string

const (
	// Go reads Go source through the type checker.
	Go Language = "go"
	// JVM reads compiled class files, which serves Java and Kotlin alike.
	JVM Language = "jvm"
)

// Profile is one supported combination.
type Profile struct {
	// Name is what a project pins, in the form <language>_<framework>_<style>.
	Name string
	// Language selects the reader.
	Language Language
	// Framework names the recognisers, for diagnostics and for the list.
	Framework string
	// Style names the architecture rules.
	Style string
	// Summary is one line saying what this combination is, so that the error
	// for an unknown profile is a menu rather than a rejection.
	Summary string

	// Layout are the conventions this profile assumes. A project's
	// speclink.json states deviations from them, not the conventions
	// themselves.
	Layout config.Config

	// Fields are the configuration keys this profile understands. Anything else
	// in a speclink.json is refused rather than ignored: classRoots under a Go
	// profile is not a harmless leftover, it is somebody expecting an effect
	// that will not happen.
	Fields []string

	// Architecture reports whether the style prescribes any rules yet. A
	// profile whose style is still empty says so rather than passing quietly,
	// on the same grounds as the capability lines: a rule family that was never
	// run must not read as one that came out clean.
	Architecture bool

	// Open builds the model. It takes the frontend's own arguments because
	// loading is where the readers differ most and neither signature
	// generalises without lying about the other.
	Open func(root string, layout config.Config, patterns []string, withTests bool, out *diag.Set) (lang.Model, error)

	// templates are the starting points init can write. A profile with none
	// says so rather than failing obscurely: being able to judge a project and
	// being able to start one are separate capabilities, and the second is the
	// one that has to be written by hand for every style.
	templates []Template
}

var registry = map[string]*Profile{}

// Register adds a profile. It panics on a duplicate, because two profiles of
// one name would make the thing a project pins ambiguous.
func Register(p *Profile) {
	if _, dup := registry[p.Name]; dup {
		panic("profile registered twice: " + p.Name)
	}
	registry[p.Name] = p
}

// Get returns a profile by name.
func Get(name string) (*Profile, error) {
	if p, ok := registry[name]; ok {
		return p, nil
	}
	if name == "" {
		return nil, fmt.Errorf("no profile set; put \"profile\" in %s or pass -profile.\n\n%s",
			config.FileName, List())
	}
	return nil, fmt.Errorf("unknown profile %q.\n\n%s", name, List())
}

// All returns every profile, ordered by name.
func All() []*Profile {
	out := make([]*Profile, 0, len(registry))
	for _, p := range registry {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// List renders the available profiles, for the message a project sees when it
// has not chosen one.
//
// Which style a project follows cannot be guessed, and guessing wrongly is
// expensive in a particular way: it reports dozens of findings about a
// convention the project never meant to follow, which teaches the reader that
// the tool is wrong rather than that the project is.
func List() string {
	var b strings.Builder
	b.WriteString("available profiles:\n")
	for _, p := range All() {
		fmt.Fprintf(&b, "  %-24s %s\n", p.Name, p.Summary)
	}
	return b.String()
}

// CheckConfig rejects configuration a profile does not understand.
//
// Silently ignoring it would be the ordinary thing to do and the wrong one.
// speclink.json is where a project writes down what it expects, and a key that
// has no effect under the chosen profile is a mistaken expectation — most often
// a profile that was changed without the configuration following.
func (p *Profile) CheckConfig(set []string) error {
	known := map[string]bool{}
	for _, f := range append(universalFields(), p.Fields...) {
		known[f] = true
	}

	var unknown []string
	for _, f := range set {
		if !known[f] {
			unknown = append(unknown, f)
		}
	}
	if len(unknown) == 0 {
		return nil
	}
	sort.Strings(unknown)
	return fmt.Errorf("%s sets %s, which profile %s does not use.\nIt understands: %s",
		config.FileName, strings.Join(unknown, ", "), p.Name,
		strings.Join(append(universalFields(), p.Fields...), ", "))
}

// universalFields are the keys every profile understands: where the raw
// requirement documents live, and which packages are measured at all.
func universalFields() []string {
	return []string{"profile", "sourceRoots", "scope", "exclude"}
}
