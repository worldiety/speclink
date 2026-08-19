package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDefaultsWithoutFile pins that a project following the convention needs no
// configuration at all. The file is the exception, not the normal case.
func TestDefaultsWithoutFile(t *testing.T) {
	cfg, err := Load(t.TempDir())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ContextRoot != "app" || cfg.CmdRoot != "cmd" {
		t.Errorf("unexpected defaults: %+v", cfg)
	}
}

// TestPartialOverride guards that stating one deviation does not silently drop
// the remaining conventions.
func TestPartialOverride(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, `{"contextRoot": "."}`)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ContextRoot != "" && cfg.ContextRoot != "." {
		t.Errorf("contextRoot = %q", cfg.ContextRoot)
	}
	if cfg.CmdRoot != "cmd" {
		t.Errorf("cmdRoot should keep its default, got %q", cfg.CmdRoot)
	}
	if len(cfg.InfraRoots) != 2 {
		t.Errorf("infraRoots should keep their defaults, got %v", cfg.InfraRoots)
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
	cfg := Default()
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
	if cfg.CmdRoot != "cmd" {
		t.Errorf("an omitted field must keep its conventional value, got %q", cfg.CmdRoot)
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
