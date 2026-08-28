package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The template is checked by running speclink over what it produces.
//
// Nothing else would do. A template is source that no compiler sees and no test
// covers until somebody uses it, which is the one kind of code that rots
// silently — and a starting point that speclink itself rejects would be the
// most embarrassing thing this repository could ship.
//
// So the test is the walkthrough from AGENTS.md, run in full: render, build,
// record what the tests demonstrated, record the shapes, then verify. The last
// step must come out clean in every measured direction.
func TestInitTemplateVerifiesClean(t *testing.T) {
	dir := renderTemplate(t, "go_bare_ddd1", "full", "example.com/erp", "sales")

	if out, err := exec.Command("gofmt", "-l", dir).CombinedOutput(); err != nil {
		t.Fatalf("gofmt: %v\n%s", err, out)
	} else if strings.TrimSpace(string(out)) != "" {
		t.Errorf("the rendered template is not gofmt clean:\n%s", out)
	}

	build := exec.Command("go", "build", "./...")
	build.Dir = dir
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("the rendered template does not build: %v\n%s", err, out)
	}

	vet := exec.Command("go", "vet", "./...")
	vet.Dir = dir
	if out, err := vet.CombinedOutput(); err != nil {
		t.Errorf("go vet on the rendered template: %v\n%s", err, out)
	}

	recordEvidence(t, dir)
	if out, code := runSpeclink(t, "freeze", dir); code != 0 {
		t.Fatalf("freeze failed with %d:\n%s", code, out)
	}

	out, code := runVerify(t, dir)
	if code != 0 {
		t.Fatalf("the rendered template does not verify:\n%s", out)
	}

	// Being clean is not enough: a direction that was never measured must not
	// read as one that came out clean, and a template whose sources or
	// bindings went unread would report exactly the same empty finding list.
	for _, want := range []string{
		"100% accounted", "100% bound", "100% covered",
		"100% verified", "100% demonstrated",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %s in the summary:\n%s", want, out)
		}
	}
	if strings.Contains(out, "not measured") {
		t.Errorf("the template leaves a direction unmeasured:\n%s", out)
	}
}

// The context name reaches directories, package names, import paths and a
// permission prefix. Rendering with a second one is what catches a placeholder
// that was hard coded in one of them.
func TestInitTemplateFollowsTheContextName(t *testing.T) {
	dir := renderTemplate(t, "go_bare_ddd1", "full", "example.com/other", "billing")

	build := exec.Command("go", "build", "./...")
	build.Dir = dir
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("the template does not build under a second context name: %v\n%s", err, out)
	}

	if _, err := os.Stat(filepath.Join(dir, "app", "billing", "model.go")); err != nil {
		t.Errorf("the context directory did not follow the name: %v", err)
	}
	// The binary is derived from the module path rather than asked for,
	// because a separate answer could contradict the import path.
	if _, err := os.Stat(filepath.Join(dir, "cmd", "other", "main.go")); err != nil {
		t.Errorf("the command directory was not derived from the module path: %v", err)
	}

	perm, err := os.ReadFile(filepath.Join(dir, "app", "billing", "perm.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(perm), `"billing.record.create"`) {
		t.Errorf("the permission id did not follow the context name:\n%s", perm)
	}
}

// Writing into a directory that already holds something is refused rather than
// merged. Hidden entries are tolerated so that git init may run first: refusing
// because of .git would impose an order that has no reason to be that way
// round.
func TestInitRefusesANonEmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, code := runInit(t, dir, "-profile", "go_bare_ddd1", "-template", "full",
		"-module", "example.com/erp", "-context", "sales")
	if code == 0 {
		t.Fatalf("init overwrote a directory that was not empty:\n%s", out)
	}
	if !strings.Contains(out, "notes.txt") {
		t.Errorf("the refusal does not name what was in the way:\n%s", out)
	}

	if _, err := os.Stat(filepath.Join(dir, "go.mod")); !os.IsNotExist(err) {
		t.Error("init wrote a file before refusing")
	}
}

func TestInitToleratesHiddenEntries(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	if out, code := runInit(t, dir, "-profile", "go_bare_ddd1", "-template", "full",
		"-module", "example.com/erp", "-context", "sales"); code != 0 {
		t.Fatalf("init refused a directory holding only .git:\n%s", out)
	}
}

// Every refusal names the alternatives, because the caller is usually an agent
// and a prompt is the one interface it cannot use.
func TestInitRefusalsAreMenus(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"no profile", nil, "available profiles:"},
		{"no template", []string{"-profile", "go_bare_ddd1"}, "templates for go_bare_ddd1:"},
		{
			"no module",
			[]string{"-profile", "go_bare_ddd1", "-template", "full"},
			`-module is required: the Go module path of the new project, for example "example.com/erp"`,
		},
		{
			"a context name that would collide",
			[]string{"-profile", "go_bare_ddd1", "-template", "full", "-module", "example.com/erp", "-context", "rest"},
			"already the name of a package in the generated project",
		},
		{
			"a profile that cannot start a project",
			[]string{"-profile", "java_springboot_ddd1"},
			"has no templates yet",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, code := runInit(t, t.TempDir(), tc.args...)
			if code == 0 {
				t.Fatalf("expected a refusal:\n%s", out)
			}
			if !strings.Contains(out, tc.want) {
				t.Errorf("expected %q:\n%s", tc.want, out)
			}
		})
	}
}

// -describe is the same catalogue for a caller that would rather plan than fail
// first. It is machine readable because the caller is a machine.
func TestInitDescribeIsMachineReadable(t *testing.T) {
	out, code := runInit(t, t.TempDir(), "-describe", "-format", "json")
	if code != 0 {
		t.Fatalf("describe failed with %d:\n%s", code, out)
	}

	var got struct {
		Profiles []struct {
			Name      string `json:"name"`
			Templates []struct {
				Name   string `json:"name"`
				Params []struct {
					Name     string `json:"name"`
					Example  string `json:"example"`
					Required bool   `json:"required"`
				} `json:"params"`
			} `json:"templates"`
		} `json:"profiles"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("describe is not valid JSON: %v\n%s", err, out)
	}

	var params []string
	for _, p := range got.Profiles {
		if p.Name != "go_bare_ddd1" {
			continue
		}
		for _, tpl := range p.Templates {
			if tpl.Name != "full" {
				continue
			}
			for _, param := range tpl.Params {
				if param.Example == "" {
					t.Errorf("parameter %s has no example, so the caller has to guess the shape", param.Name)
				}
				if !param.Required {
					t.Errorf("parameter %s is optional; an optional parameter is a default, and a default that shapes a project is a decision taken silently", param.Name)
				}
				params = append(params, param.Name)
			}
		}
	}
	if strings.Join(params, ",") != "module,context" {
		t.Errorf("describe lists %v, expected module and context", params)
	}
}

// A profile with no template says so. Being able to judge a project and being
// able to start one are separate capabilities, and a profile that quietly
// produced nothing would look like a broken template rather than an absent one.
func TestEveryProfileIsDescribed(t *testing.T) {
	out, code := runInit(t, t.TempDir(), "-describe")
	if code != 0 {
		t.Fatalf("describe failed with %d:\n%s", code, out)
	}
	for _, want := range []string{"go_bare_ddd1", "go_nago_ddd1", "java_springboot_ddd1"} {
		if !strings.Contains(out, want) {
			t.Errorf("describe does not mention %s:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "no templates") {
		t.Errorf("a profile without a template does not say so:\n%s", out)
	}
}

// renderTemplate runs init and points the result at this working copy.
//
// The replace is the test's own, not the template's. A generated project
// resolves speclink from the module proxy like any other dependency; pinning it
// here keeps the test hermetic and, more importantly, makes it check the
// template against the rules in this checkout rather than against whatever was
// last published.
func renderTemplate(t *testing.T, profileName, template, module, context string) string {
	t.Helper()

	dir := t.TempDir()
	if out, code := runInit(t, dir, "-profile", profileName, "-template", template,
		"-module", module, "-context", context); code != 0 {
		t.Fatalf("init failed with %d:\n%s", code, out)
	}

	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "go.mod")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	body = append(body, []byte("\nrequire github.com/worldiety/speclink v0.0.0\n\nreplace github.com/worldiety/speclink => "+root+"\n")...)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// runInit runs the init command. It cannot go through runSpeclink, which adds
// a -root that init does not have: init writes a directory rather than reading
// one, and calling the flag -root would suggest it reads what is already there.
func runInit(t *testing.T, dir string, extra ...string) (string, int) {
	t.Helper()

	cmd := exec.Command(speclinkBin, append([]string{"init", "-dir", dir}, extra...)...)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()

	code := 0
	var exit *exec.ExitError
	if err != nil {
		if ok := asExitError(err, &exit); ok {
			code = exit.ExitCode()
		} else {
			t.Fatalf("run speclink init: %v\n%s", err, out)
		}
	}
	return string(out), code
}
