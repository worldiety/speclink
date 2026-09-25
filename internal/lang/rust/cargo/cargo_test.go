package cargo

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseTOML(t *testing.T) {
	doc, err := parseTOML(`
# comment
[package]
name = "my-service" # trailing
version = "0.1.0"
"quoted key" = 'literal \n'
description = """
two
lines"""

[dependencies]
serde = { version = "1", features = ["derive",
  "std"] }
spec.path = "../spec"

[[bin]]
name = "a"
[[bin]]
name = "b"
path = "src/b.rs"

[target.'cfg(unix)'.dependencies]
libc = "0.2"
`)
	if err != nil {
		t.Fatal(err)
	}
	pkg := doc["package"].(map[string]any)
	if pkg["name"] != "my-service" || pkg["version"] != "0.1.0" || pkg["quoted key"] != `literal \n` || pkg["description"] != "two\nlines" {
		t.Errorf("package: %#v", pkg)
	}
	deps := doc["dependencies"].(map[string]any)
	serde := deps["serde"].(map[string]any)
	if !reflect.DeepEqual(serde["features"], []any{"derive", "std"}) {
		t.Errorf("serde: %#v", serde)
	}
	if deps["spec"].(map[string]any)["path"] != "../spec" {
		t.Errorf("dotted key: %#v", deps["spec"])
	}
	bins := doc["bin"].([]map[string]any)
	if len(bins) != 2 || bins[1]["path"] != "src/b.rs" {
		t.Errorf("bins: %#v", bins)
	}
	if doc["target"].(map[string]any)["cfg(unix)"] == nil {
		t.Error("quoted table key")
	}
}

func TestParseTOMLErrors(t *testing.T) {
	for _, src := range []string{"a = ", "a = \"x", "[a\nb=1", "a = 1\na = 2", "a = [1 2]"} {
		if _, err := parseTOML(src); err == nil {
			t.Errorf("%q: no error", src)
		}
	}
}

func write(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		p := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestTargets(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"Cargo.toml": `
[package]
name = "my-service"
edition = "2021"
[[bin]]
name = "tool"
path = "tools/tool.rs"
`,
		"src/lib.rs":            "",
		"src/main.rs":           "",
		"src/bin/extra.rs":      "",
		"src/bin/multi/main.rs": "",
		"tools/tool.rs":         "",
	})
	m, err := ReadManifest(filepath.Join(root, "Cargo.toml"))
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, tg := range m.Targets() {
		rel, _ := filepath.Rel(root, tg.Path)
		got = append(got, tg.Kind.String()+" "+tg.Name+" "+filepath.ToSlash(rel))
	}
	want := []string{
		"lib my_service src/lib.rs",
		"bin tool tools/tool.rs",
		"bin my_service src/main.rs",
		"bin extra src/bin/extra.rs",
		"bin multi src/bin/multi/main.rs",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("targets:\n got  %q\n want %q", got, want)
	}
}

func TestLibOverride(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"Cargo.toml":  "[package]\nname = \"x\"\nautobins = false\n[lib]\nname = \"core-x\"\npath = \"lib/core.rs\"\n",
		"src/main.rs": "",
	})
	m, err := ReadManifest(filepath.Join(root, "Cargo.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if m.Lib == nil || m.Lib.Name != "core_x" || m.Lib.Path != filepath.Join(root, "lib", "core.rs") {
		t.Errorf("lib: %+v", m.Lib)
	}
	if len(m.Bins) != 0 {
		t.Errorf("autobins = false still found %+v", m.Bins)
	}
}

func TestWorkspace(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"Cargo.toml": `
[workspace]
members = ["crates/*", "app"]
exclude = ["crates/old"]
[workspace.dependencies]
spec = { path = "crates/spec" }
`,
		"crates/spec/Cargo.toml": "[package]\nname = \"spec\"\n",
		"crates/spec/src/lib.rs": "",
		"crates/old/Cargo.toml":  "[package]\nname = \"old\"\n",
		"crates/notes/README.md": "",
		"app/Cargo.toml":         "[package]\nname = \"app\"\n[dependencies]\nspec.workspace = true\nserde = \"1\"\n[dev-dependencies]\nhelper = { path = \"../helper\", package = \"test-helper\" }\n",
		"app/src/main.rs":        "",
	})
	w, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, p := range w.Packages {
		names = append(names, p.Package.Name)
	}
	if !reflect.DeepEqual(names, []string{"spec", "app"}) {
		t.Errorf("packages: %q", names)
	}
	app := w.PackageAt(filepath.Join(root, "app"))
	if d := app.Deps["spec"]; d.Path != filepath.Join(root, "crates", "spec") || d.Inherited {
		t.Errorf("inherited dependency: %+v", d)
	}
	if d := app.Deps["serde"]; d.Path != "" {
		t.Errorf("registry dependency has a path: %+v", d)
	}
	if d := app.Deps["helper"]; !d.Dev || d.Package != "test-helper" || d.Path != filepath.Join(root, "helper") {
		t.Errorf("renamed dev dependency: %+v", d)
	}
}

func TestWorkspaceMemberMissing(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{"Cargo.toml": "[workspace]\nmembers = [\"nope\"]\n"})
	if _, err := Load(root); err == nil {
		t.Error("a member matching nothing was accepted")
	}
}
