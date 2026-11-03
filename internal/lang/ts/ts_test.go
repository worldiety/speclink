package ts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/worldiety/speclink/internal/config"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// writeProject lays out a project in a temporary directory.
func writeProject(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func load(t *testing.T, root string) *Project {
	t.Helper()
	p, errs := Load(root, config.Config{SrcRoot: "src", TestRoots: []string{"tests"}})
	if len(errs) > 0 {
		t.Fatalf("load: %v", errs)
	}
	return p
}

func rules(out *diag.Set) map[string]int {
	got := map[string]int{}
	for _, f := range out.Findings() {
		got[f.Rule]++
	}
	return got
}

const vocabulary = `export function bind(..._a: unknown[]) {}
export function bindModule(..._a: unknown[]) {}
export function waive(_r: string, _w: string) { return {}; }
export function satisfies(..._r: unknown[]) { return {}; }
export function requirement(r: unknown) { return r; }
export function verified(..._r: unknown[]) {}
export enum Kind { Functional = 'Functional' }
export enum Status { Normative = 'Normative' }
`

const tsconfig = `{ "compilerOptions": { "paths": { "@/*": ["./src/*"] } } }`

// The vocabulary is recognised by the file it was imported from, never by the
// spelling of the call. A project function called bind is not the annotation
// language, and an aliased import of the real one still is.
func TestVocabularyIsRecognisedByOrigin(t *testing.T) {
	t.Parallel()
	root := writeProject(t, map[string]string{
		"tsconfig.json":   tsconfig,
		"src/speclink.ts": vocabulary,
		"src/domain/usecases/AUseCase.ts": `export default class AUseCase {
  execute(s: { audit(p: string): void }) { s.audit('x'); }
}`,
		"src/domain/usecases/AUseCase.annotation.ts": `import { bind as tie, satisfies } from '@/speclink.ts';
import AUseCase from '@/domain/usecases/AUseCase.ts';
import { RX } from '../../../requirements/fun/x/R-X-Y.req.ts';

tie(AUseCase, satisfies(RX));`,
		"src/domain/usecases/BUseCase.ts": `function bind(..._a: unknown[]) {}
export default class BUseCase {
  execute(s: { audit(p: string): void }) { s.audit('x'); }
}
bind(BUseCase);`,
		"requirements/fun/x/R-X-Y.req.ts": `import * as spec from '../../../src/speclink.ts';

export const RX = spec.requirement({
  id: 'R-X-Y',
  kind: spec.Kind.Functional,
  status: spec.Status.Normative,
  text: 'It MUST.'
});`,
	})
	p := load(t, root)
	out := &diag.Set{}

	reqs := p.Requirements(out)
	if len(reqs) != 1 || reqs[0].ID != "R-X-Y" || reqs[0].Kind != ir.Functional || reqs[0].Status != ir.Normative {
		t.Fatalf("requirement misread: %+v", reqs)
	}
	if reqs[0].Symbol != "requirements/fun/x/R-X-Y.req.RX" {
		t.Errorf("symbol %q", reqs[0].Symbol)
	}

	bindings := p.Bindings(out)
	if len(bindings) != 1 {
		t.Fatalf("want the aliased bind and not the local one, got %+v", bindings)
	}
	b := bindings[0]
	if b.Target.Name != "src/domain/usecases/AUseCase.AUseCase" {
		t.Errorf("target %q", b.Target.Name)
	}
	if len(b.Assertions) != 1 || len(b.Assertions[0].Requirements) != 1 || b.Assertions[0].Requirements[0] != reqs[0].Symbol {
		t.Errorf("satisfies misread: %+v", b.Assertions)
	}
	if !out.Empty() {
		t.Errorf("unexpected findings: %+v", out.Findings())
	}

	cs := p.Constructs()
	if len(cs) != 2 || cs[0].Name != b.Target.Name {
		t.Errorf("constructs %+v", cs)
	}
}

// A requirement is read and never run, so a value that would have to be
// evaluated is refused rather than guessed at.
func TestRequirementValuesMustBeLiterals(t *testing.T) {
	t.Parallel()
	root := writeProject(t, map[string]string{
		"src/speclink.ts": vocabulary,
		"requirements/fun/x/R-X-Y.req.ts": `import { requirement, Kind, Status } from '../../../src/speclink.ts';

const prefix = 'It';
export const RX = requirement({
  id: 'R-X-Y',
  kind: Kind.Functional,
  status: Status.Normative,
  text: prefix + ' MUST.',
  colour: 'red'
});`,
	})
	p := load(t, root)
	out := &diag.Set{}
	p.Requirements(out)
	p.CheckSyntax(out)

	codes := map[string]int{}
	for _, f := range out.Findings() {
		codes[f.Code]++
	}
	for _, want := range []string{"SPEC-V1-020", "SPEC-V1-021", "SPEC-V1-022"} {
		if codes[want] != 1 {
			t.Errorf("want one %s, got %v", want, codes)
		}
	}
}

// A waiver narrows to the construct or the module it names and to nothing
// else, and the entry point may only be waived by saying so.
func TestWaiversAndTheMissingEntryPoint(t *testing.T) {
	t.Parallel()
	root := writeProject(t, map[string]string{
		"tsconfig.json":   tsconfig,
		"src/speclink.ts": vocabulary,
		"src/domain/usecases/OpenUseCase.ts": `export default class OpenUseCase {
  execute() { return 1; }
}`,
		"src/domain/usecases/OpenUseCase.annotation.ts": `import { bind, waive } from '@/speclink.ts';
import OpenUseCase from '@/domain/usecases/OpenUseCase.ts';

bind(OpenUseCase, waive('K5-UC-AUTHZ', 'open to every signed in user'));`,
		"src/domain/usecases/ClosedUseCase.ts": `export default class ClosedUseCase {
  execute() { return 1; }
}`,
		"src/domain/services/Browser.ts": `export default class Browser {
  origin() { return window.location.origin; }
}`,
		"src/domain/services/Browser.annotation.ts": `import { bindModule, waive } from '@/speclink.ts';

bindModule(waive('K6-DOMAIN-NO-FRAMEWORK', 'kept for the migration'));`,
	})
	p := load(t, root)
	out := &diag.Set{}
	p.CheckArchitecture(out)

	got := rules(out)
	if got[RuleUCAuthz] != 1 {
		t.Errorf("want K5-UC-AUTHZ once, for the unwaived use case: %v", got)
	}
	if got[RuleDomainNoFramework] != 0 {
		t.Errorf("a module waiver did not apply: %v", got)
	}
	if got[RuleMainExists] != 1 {
		t.Errorf("a project without an entry point was not reported: %v", got)
	}
}

// A route is an address the application answers on, and what stands behind it
// is the use cases its view constructs, directly or through its components.
func TestRoutesAreTracedThroughComponents(t *testing.T) {
	t.Parallel()
	root := writeProject(t, map[string]string{
		"tsconfig.json": tsconfig,
		"src/domain/usecases/ShowUseCase.ts": `export default class ShowUseCase {
  execute() {}
}`,
		"src/domain/services/ServiceLocator.ts": `export default class ServiceLocator {}`,
		"src/adapters/ui/router/index.ts": `import { createRouter } from 'vue-router';
import HomeView from '@/adapters/ui/views/HomeView.vue';

export default createRouter({
  routes: [
    { path: '/', component: HomeView, children: [{ path: 'about', component: () => import('../views/AboutView.vue') }] },
    { path: '/old', redirect: '/' }
  ]
});`,
		"src/adapters/ui/views/HomeView.vue": `<template><Card /></template>
<script setup lang="ts">
import Card from '@/adapters/ui/components/Card.vue';
</script>`,
		"src/adapters/ui/views/AboutView.vue": `<template><p>about</p></template>`,
		"src/adapters/ui/components/Card.vue": `<template><p/></template>
<script setup lang="ts">
import ShowUseCase from '@/domain/usecases/ShowUseCase.ts';
import ServiceLocator from '@/domain/services/ServiceLocator.ts';
new ShowUseCase(ServiceLocator).execute();
</script>`,
	})
	p := load(t, root)
	eps := p.Endpoints()
	if len(eps) != 2 {
		t.Fatalf("want 2 routes, got %+v", eps)
	}
	if eps[0].Path != "/" || len(eps[0].UseCases) != 1 || eps[0].UseCases[0] != "src/domain/usecases/ShowUseCase.ShowUseCase" {
		t.Errorf("home misread: %+v", eps[0])
	}
	if eps[1].Path != "/about" || !strings.HasSuffix(eps[1].Handler, "AboutView.vue") || len(eps[1].UseCases) != 0 {
		t.Errorf("lazy child misread: %+v", eps[1])
	}
}

// The claim and the report meet on one name, and the report's file name is
// matched by its ending, because it was written on whatever machine ran the
// tests.
func TestDemonstrationsJoinClaimsAndResults(t *testing.T) {
	t.Parallel()
	root := writeProject(t, map[string]string{
		"src/speclink.ts": vocabulary,
		"requirements/fun/x/R-X-Y.req.ts": `import { requirement, Kind, Status } from '../../../src/speclink.ts';
export const RX = requirement({ id: 'R-X-Y', kind: Kind.Functional, status: Status.Normative, text: 'It MUST.' });`,
		"tests/x.test.ts": `import { verified } from '../src/speclink.ts';
import { RX } from '../requirements/fun/x/R-X-Y.req.ts';

describe('group', () => {
  test.only('handles 1.5 weeks', () => {
    verified(RX);
  });
  it('fails', () => {
    verified(RX);
  });
});`,
	})
	p := load(t, root)
	claims := p.Verifications(&diag.Set{})
	if len(claims) != 2 || claims[0].Target.Name != "tests/x.test.ts.group > handles 1\u20245 weeks" {
		t.Fatalf("claims misread: %+v", claims)
	}

	report := `{"testResults":[{"name":"/ci/build/tests/x.test.ts","assertionResults":[
  {"ancestorTitles":["group"],"title":"handles 1.5 weeks","status":"passed"},
  {"ancestorTitles":["group"],"title":"fails","status":"failed"}]}]}`
	shown, err := p.Demonstrations(strings.NewReader(report), index{"requirements/fun/x/R-X-Y.req.RX": "R-X-Y"})
	if err != nil {
		t.Fatal(err)
	}
	if len(shown["R-X-Y"]) != 1 || shown["R-X-Y"][0] != claims[0].Target.Name {
		t.Errorf("demonstrations %+v", shown)
	}
}

type index map[string]string

func (i index) IDOf(ref string) (string, bool) {
	id, ok := i[ref]
	return id, ok
}

// A parametrised test claims for every case, and the report names each case
// with its parameter filled in.
func TestParametrisedClaimsMatchEveryCase(t *testing.T) {
	t.Parallel()
	root := writeProject(t, map[string]string{
		"src/speclink.ts": vocabulary,
		"requirements/fun/x/R-X-Y.req.ts": `import { requirement, Kind, Status } from '../../../src/speclink.ts';
export const RX = requirement({ id: 'R-X-Y', kind: Kind.Functional, status: Status.Normative, text: 'It MUST.' });`,
		"tests/x.test.js": `import { verified } from '../src/speclink.ts';
import { RX } from '../requirements/fun/x/R-X-Y.req.ts';

describe.each(['admin', 'advisor'])('Test for %s', (role) => {
  test('sees everything', () => {
    verified(RX);
  });
});

describe('Test for merchant', () => {
  test('sees everything', () => {});
});`,
	})
	p := load(t, root)
	report := `{"testResults":[{"name":"tests/x.test.js","assertionResults":[
  {"ancestorTitles":["Test for admin"],"title":"sees everything","status":"passed"},
  {"ancestorTitles":["Test for advisor"],"title":"sees everything","status":"passed"},
  {"ancestorTitles":["Test for merchant"],"title":"sees everything","status":"passed"},
  {"ancestorTitles":["Other"],"title":"sees everything","status":"passed"}]}]}`
	shown, err := p.Demonstrations(strings.NewReader(report), index{"requirements/fun/x/R-X-Y.req.RX": "R-X-Y"})
	if err != nil {
		t.Fatal(err)
	}
	if len(shown["R-X-Y"]) != 2 {
		t.Errorf("want both cases and nothing else, got %v", shown["R-X-Y"])
	}
}

// Recent Vitest versions write the JSON report to a file and print only where.
// The documented pipe has to keep working, so the line is followed.
func TestReportWrittenToAFileIsFollowed(t *testing.T) {
	t.Parallel()
	root := writeProject(t, map[string]string{
		"src/speclink.ts": vocabulary,
		"requirements/fun/x/R-X-Y.req.ts": `import { requirement, Kind, Status } from '../../../src/speclink.ts';
export const RX = requirement({ id: 'R-X-Y', kind: Kind.Functional, status: Status.Normative, text: 'It MUST.' });`,
		"tests/x.test.ts": `import { verified } from '../src/speclink.ts';
import { RX } from '../requirements/fun/x/R-X-Y.req.ts';
test('works', () => { verified(RX); });`,
		".vitest/json/output.json": `{"testResults":[{"name":"tests/x.test.ts","assertionResults":[{"ancestorTitles":[],"title":"works","status":"passed"}]}]}`,
	})
	p := load(t, root)
	shown, err := p.Demonstrations(strings.NewReader("JSON report written to "+filepath.Join(root, ".vitest/json/output.json")+"\n"),
		index{"requirements/fun/x/R-X-Y.req.RX": "R-X-Y"})
	if err != nil {
		t.Fatal(err)
	}
	if len(shown["R-X-Y"]) != 1 {
		t.Errorf("the report the line names was not read: %v", shown)
	}
}
