// Package config carries the few things speclink cannot know: where a project
// keeps its bounded contexts, its commands and its infrastructure packages.
//
// Everything else is hardcoded. Rules and recognisers know the framework, never
// the project: a framework is shared by many projects and that knowledge
// amortises, while project knowledge never does. A directory layout is the
// exception, because no amount of framework insight reveals whether contexts
// live under app/, application/ or at the module root.
//
// The defaults are the convention. A speclink.json is the exception, not the
// normal case, and a project that follows the convention needs none.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FileName is the optional configuration file in the project root.
const FileName = "speclink.json"

// Config describes the directory layout of a project.
type Config struct {
	// ContextRoot is the directory holding the bounded contexts, relative to
	// the project root. "." means the contexts sit at the root itself.
	ContextRoot string `json:"contextRoot,omitempty"`

	// CmdRoot is the directory every main package must live under.
	CmdRoot string `json:"cmdRoot,omitempty"`

	// InfraRoots are the directories reserved for infrastructure helpers.
	// Nothing domain specific may live there.
	InfraRoots []string `json:"infraRoots,omitempty"`

	// SourceRoots are the directories holding the raw requirement documents.
	//
	// They have to be named rather than discovered, because forward coverage
	// of the sources only means something if the set of sources is known. A
	// tool that checked only the documents somebody happened to cite could
	// never report the document nobody read: the one place where a whole
	// feature goes missing without a single finding.
	SourceRoots []string `json:"sourceRoots,omitempty"`

	// Scope restricts speclink to the packages matching one of these patterns.
	// Empty means the whole module, which is the normal case.
	//
	// This is the only dial speclink has, and it is deliberately the only one.
	// There are no warnings and no severities: a finding is an error and the
	// run fails. The reasons compound. The Go compiler behaves the same way,
	// and softening here would be incoherent anyway, because the annotations
	// sit in the ordinary build and a compile error cannot be downgraded to a
	// warning. Warnings meant for a migration become a permanent excuse — that
	// is not a guess but the standing behaviour of every codebase with a
	// warning backlog. And the reader is a model, which iterates until green:
	// commented-out code is visibly unfinished, a suppressed warning is
	// invisibly unfinished.
	//
	// So a codebase is brought in package by package rather than rule by rule.
	// The difference matters: "this package is not under speclink yet" is a
	// true statement, "this rule half applies here" is not one.
	Scope []string `json:"scope,omitempty"`

	// Exclude are path patterns speclink ignores, matched with path.Match
	// against the project relative directory. Examples and generated
	// documentation copies are the usual candidates.
	Exclude []string `json:"exclude,omitempty"`
}

// Default returns the conventional layout.
func Default() Config {
	return Config{
		ContextRoot: "app",
		CmdRoot:     "cmd",
		InfraRoots:  []string{"pkg", "foundation"},
		SourceRoots: []string{"requirements/_sources"},
	}
}

// Load reads speclink.json from root. A missing file is not an error: it means
// the project follows the convention.
func Load(root string) (Config, error) {
	return LoadFile(filepath.Join(root, FileName), true)
}

// LoadFile reads a configuration from an explicit path.
//
// optional says whether a missing file is acceptable. It is for the
// conventional location, where absence means "the convention applies", and it
// is not for a path the caller named: asking for a file that is not there is a
// mistake worth reporting rather than silently falling back to defaults that
// were deliberately being overridden.
func LoadFile(path string, optional bool) (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) && optional {
		return cfg, nil
	}
	if err != nil {
		return cfg, fmt.Errorf("read %s: %w", path, err)
	}

	// Decode into the defaults so an omitted field keeps its conventional
	// value and a project only states what it deviates in.
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse %s: %w", path, err)
	}
	cfg.normalise()
	return cfg, nil
}

func (c *Config) normalise() {
	c.ContextRoot = filepath.ToSlash(strings.Trim(c.ContextRoot, "/"))
	c.CmdRoot = filepath.ToSlash(strings.Trim(c.CmdRoot, "/"))
	for i, r := range c.InfraRoots {
		c.InfraRoots[i] = filepath.ToSlash(strings.Trim(r, "/"))
	}
	for i, r := range c.SourceRoots {
		c.SourceRoots[i] = filepath.ToSlash(strings.Trim(r, "/"))
	}
	for i, r := range c.Scope {
		c.Scope[i] = filepath.ToSlash(strings.Trim(r, "/"))
	}
}

// InScope reports whether a project relative directory is checked at all.
//
// The decision is binary and it is the whole of it. Inside the scope
// completeness is claimed; outside, nothing is. A package is either measured or
// it is not, and which one is stated in the configuration rather than implied
// by a tolerance somebody set once.
func (c Config) InScope(rel string) bool {
	rel = filepath.ToSlash(rel)
	if c.Excluded(rel) {
		return false
	}
	if len(c.Scope) == 0 {
		return true
	}
	for _, pattern := range c.Scope {
		if matchPath(pattern, rel) {
			return true
		}
	}
	return false
}

// Restricted reports whether the scope covers less than the whole module.
//
// A run that measures part of a project must say so. Without it a summary
// reading a hundred percent would be true of what was looked at and silent
// about what was not, which is the one way this tool could mislead by telling
// the truth.
func (c Config) Restricted() bool { return len(c.Scope) > 0 || len(c.Exclude) > 0 }

// Excluded reports whether a project relative directory is explicitly out of
// scope.
func (c Config) Excluded(rel string) bool {
	rel = filepath.ToSlash(rel)
	for _, pattern := range c.Exclude {
		if matchPath(pattern, rel) {
			return true
		}
	}
	return false
}

// matchPath supports a trailing /** to mean "this directory and everything
// below it", which path.Match alone cannot express.
func matchPath(pattern, rel string) bool {
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")
		return rel == prefix || strings.HasPrefix(rel, prefix+"/")
	}
	ok, err := filepath.Match(pattern, rel)
	return err == nil && ok
}

// InContextRoot reports whether a project relative directory lies inside the
// context root, and returns the context name.
//
// With ContextRoot "." the first path segment is the context, which is how a
// project that keeps its contexts at the module root is described. Command and
// infrastructure roots are excluded, otherwise pkg/ and cmd/ would be mistaken
// for bounded contexts and every rule would fire on them.
func (c Config) InContextRoot(rel string) (context string, ok bool) {
	rel = filepath.ToSlash(rel)
	if rel == "" || rel == "." {
		return "", false
	}
	if c.InInfraRoot(rel) || c.UnderCmdRoot(rel) {
		return "", false
	}
	parts := strings.Split(rel, "/")

	if c.ContextRoot == "" || c.ContextRoot == "." {
		return parts[0], true
	}
	prefix := strings.Split(c.ContextRoot, "/")
	if len(parts) <= len(prefix) {
		return "", false
	}
	for i, p := range prefix {
		if parts[i] != p {
			return "", false
		}
	}
	return parts[len(prefix)], true
}

// InInfraRoot reports whether a project relative directory lies inside one of
// the infrastructure roots.
func (c Config) InInfraRoot(rel string) bool {
	rel = filepath.ToSlash(rel)
	for _, r := range c.InfraRoots {
		if rel == r || strings.HasPrefix(rel, r+"/") {
			return true
		}
	}
	return false
}

// UnderCmdRoot reports whether a project relative directory lies under the
// command root, at any depth: both cmd/erp and example/cmd/tutorial qualify.
func (c Config) UnderCmdRoot(rel string) bool {
	rel = filepath.ToSlash(rel)
	if c.CmdRoot == "" {
		return true
	}
	for _, part := range strings.Split(rel, "/") {
		if part == c.CmdRoot {
			return true
		}
	}
	return false
}
