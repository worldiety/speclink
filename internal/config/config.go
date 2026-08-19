// Package config carries the few things speclink cannot know: where a project
// keeps its bounded contexts, its commands and its infrastructure packages.
//
// Everything else is hardcoded. Rules and recognisers know the framework, never
// the project (docs/annotations.md §1.7) — but a directory layout is genuinely
// project knowledge, and no amount of framework insight reveals whether
// contexts live under app/, application/ or at the module root.
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

	// Exclude are path patterns the architecture rules ignore, matched with
	// path.Match against the project relative directory. Examples and
	// generated documentation copies are the usual candidates.
	Exclude []string `json:"exclude,omitempty"`
}

// Default returns the conventional layout.
func Default() Config {
	return Config{
		ContextRoot: "app",
		CmdRoot:     "cmd",
		InfraRoots:  []string{"pkg", "foundation"},
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
}

// Excluded reports whether a project relative directory is out of scope for the
// architecture rules.
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
