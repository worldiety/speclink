package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMissingFileIsEmpty pins that a project deviating in nothing needs no
// configuration at all. The conventions are the profile's, so an absent file
// means "no deviations" rather than "no conventions".
func TestMissingFileIsEmpty(t *testing.T) {
	cfg, err := Load(t.TempDir())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ContextRoot != "" || cfg.CmdRoot != "" || len(cfg.Keys()) != 0 {
		t.Errorf("an absent file produced values: %+v", cfg)
	}
}

// Stating one deviation must not silently drop the remaining conventions. That
// is what Over is for, and it is the reason the load step stops short of
// filling anything in.
func TestPartialOverride(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, `{"contextRoot": "."}`)

	declared, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	cfg := declared.Over(Config{
		ContextRoot: "app",
		CmdRoot:     "cmd",
		InfraRoots:  []string{"pkg", "foundation"},
	})

	if cfg.ContextRoot != "." {
		t.Errorf("contextRoot = %q", cfg.ContextRoot)
	}
	if cfg.CmdRoot != "cmd" {
		t.Errorf("cmdRoot should keep the profile's value, got %q", cfg.CmdRoot)
	}
	if len(cfg.InfraRoots) != 2 {
		t.Errorf("infraRoots should keep the profile's values, got %v", cfg.InfraRoots)
	}
}

func TestInContextRoot(t *testing.T) {
	tests := []struct {
		root, dir, want string
		ok              bool
	}{
		{"app", "app/sales", "sales", true},
		{"app", "app/sales/ui", "sales", true},
		{"app", "pkg/data", "", false},
		{"app", "app", "", false},
		// A project keeping its contexts at the module root.
		{".", "sales/ui", "sales", true},
		{".", "", "", false},
		{"application", "application/role/ui", "role", true},
		// With the contexts at the module root, infrastructure and commands
		// must not be mistaken for contexts.
		{".", "pkg/uix", "", false},
		{".", "cmd/erp", "", false},
		{".", "foundation/log", "", false},
	}
	for _, tt := range tests {
		cfg := Config{ContextRoot: tt.root, CmdRoot: "cmd", InfraRoots: []string{"pkg", "foundation"}}
		got, ok := cfg.InContextRoot(tt.dir)
		if got != tt.want || ok != tt.ok {
			t.Errorf("InContextRoot(%q, %q) = %q,%v; want %q,%v", tt.root, tt.dir, got, ok, tt.want, tt.ok)
		}
	}
}

func TestUnderCmdRoot(t *testing.T) {
	cfg := Config{CmdRoot: "cmd"}
	// Both a top level command and a nested one qualify: nago keeps its
	// tutorials under example/cmd/.
	for _, dir := range []string{"cmd/erp", "example/cmd/tutorial-01"} {
		if !cfg.UnderCmdRoot(dir) {
			t.Errorf("%q should be under the command root", dir)
		}
	}
	if cfg.UnderCmdRoot("tools") {
		t.Error("tools should not be under the command root")
	}
}

func TestExcluded(t *testing.T) {
	cfg := Config{Exclude: []string{"docs/**", "example"}}
	for _, dir := range []string{"docs", "docs/a/b", "example"} {
		if !cfg.Excluded(dir) {
			t.Errorf("%q should be excluded", dir)
		}
	}
	if cfg.Excluded("app/sales") {
		t.Error("app/sales must not be excluded")
	}
}

func write(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

// TestLoadFileExplicit guards the path that lets a project be measured without
// being modified. Trying speclink on an unfamiliar codebase should not require
// a commit to it, and the first run is exactly where the layout is least likely
// to match the convention.
func TestLoadFileExplicit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "elsewhere.json")
	if err := os.WriteFile(path, []byte(`{"contextRoot": ".", "infraRoots": ["pkg", "kernel"]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFile(path, false)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if cfg.ContextRoot != "." {
		t.Errorf("contextRoot = %q, want %q", cfg.ContextRoot, ".")
	}
	// An omitted field stays empty here and is filled by the profile, which is
	// what carries the conventions now. Filling them in at load time would make
	// a deviation indistinguishable from a convention agreed with, and the
	// difference is what the profile refuses foreign keys on.
	if cfg.CmdRoot != "" {
		t.Errorf("an omitted field was given a value before the profile applied, got %q", cfg.CmdRoot)
	}
}

// TestLoadFileMissingIsAnError separates the two cases. A missing file at the
// conventional location means the convention applies; a missing file the caller
// named is a mistake, and falling back to the very defaults that were being
// overridden would produce a measurement nobody asked for.
func TestLoadFileMissingIsAnError(t *testing.T) {
	if _, err := LoadFile(filepath.Join(t.TempDir(), "absent.json"), false); err == nil {
		t.Error("naming a configuration that does not exist must fail")
	}
	if _, err := LoadFile(filepath.Join(t.TempDir(), "absent.json"), true); err != nil {
		t.Errorf("the conventional location may be absent: %v", err)
	}
}

// A project states what it deviates in, never what it agrees with.
//
// The conventions used to live in a Default() that read as what speclink
// believed about every project rather than as one style's layout, so a project
// on a different style had no way to say so except by restating all of it.
func TestOverAppliesOnlyWhatWasWritten(t *testing.T) {
	base := Config{
		ContextRoot: "app",
		CmdRoot:     "cmd",
		InfraRoots:  []string{"pkg", "foundation"},
		SourceRoots: []string{"requirements/_sources"},
	}
	declared := Config{Profile: "go_nago_ddd1", CmdRoot: "tools"}

	got := declared.Over(base)

	if got.CmdRoot != "tools" {
		t.Errorf("the deviation did not apply: %q", got.CmdRoot)
	}
	if got.ContextRoot != "app" || len(got.InfraRoots) != 2 {
		t.Errorf("a convention was lost by stating something else: %+v", got)
	}
	if got.Profile != "go_nago_ddd1" {
		t.Errorf("the profile was lost: %q", got.Profile)
	}
}

// Decoding alone cannot tell an omitted field from one set to its zero value,
// and the difference is exactly what makes a mistaken expectation visible: a
// key that has no effect under the chosen profile is almost always a profile
// that changed without the configuration following.
func TestKeysReportsWhatWasWritten(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)
	if err := os.WriteFile(path, []byte(`{"profile":"go_nago_ddd1","cmdRoot":"","scope":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFile(path, false)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"profile": true, "cmdRoot": true, "scope": true}
	if len(cfg.Keys()) != len(want) {
		t.Fatalf("got %v, want %v", cfg.Keys(), want)
	}
	for _, k := range cfg.Keys() {
		if !want[k] {
			t.Errorf("unexpected key %q", k)
		}
	}
}
