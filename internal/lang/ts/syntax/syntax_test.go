package syntax

import (
	"testing"
)

func kinds(toks []Token) []Kind {
	out := make([]Kind, 0, len(toks))
	for _, t := range toks {
		out = append(out, t.Kind)
	}
	return out
}

func TestLexTemplatesKeepBracketsBalanced(t *testing.T) {
	toks, errs := Lex("const u = `${base}wep/${a({b: 1})}/x`; f(/[)}]/g, a / b)", Pos{})
	if len(errs) > 0 {
		t.Fatalf("errors: %v", errs)
	}
	depth := 0
	for _, tok := range toks {
		if isOpener(tok) {
			depth++
		}
		if isCloser(tok) {
			depth--
		}
	}
	if depth != 0 {
		t.Fatalf("brackets do not balance: %v", kinds(toks))
	}
	var regex, div int
	for _, tok := range toks {
		if tok.Kind == Regex {
			regex++
		}
		if tok.Is("/") {
			div++
		}
	}
	if regex != 1 || div != 1 {
		t.Errorf("want one regex and one division, got %d and %d", regex, div)
	}
}

func TestLexPositionsFollowTheRegion(t *testing.T) {
	src := "<template>\n  <p/>\n</template>\n<script setup lang=\"ts\">\nimport x from 'y';\n</script>\n"
	var script Block
	for _, b := range SFCBlocks(src) {
		if b.Tag == "script" {
			script = b
		}
	}
	if !script.Setup() {
		t.Fatalf("script setup not recognised: %+v", script)
	}
	f := ParseRegion("x.vue", src, script.From, script.To)
	if len(f.Items) != 1 || f.Items[0].Import == nil {
		t.Fatalf("want one import, got %+v", f.Items)
	}
	if got := f.Items[0].Pos.Line; got != 5 {
		t.Errorf("import on line %d, want 5", got)
	}
}

const useCase = `import type User from '@/domain/types/User.ts';
import AuthService from '@/domain/services/AuthService.ts';

/**
 * Determines the filter.
 */
export default class ResolveUseCase {
  private authService: AuthService;
  static count = 0

  constructor(authService: AuthService, private readonly other: Other) {
    this.authService = authService;
  }

  execute(a: string | undefined, b: User[]): { technician: User | undefined } {
    const info = this.authService.getUserInfo();
    if (info.permissions.canViewAllAppointments) {
      return { technician: b[0] };
    }
    return { technician: undefined };
  }

  private helper = () => {
    return 1;
  };

  get name(): string { return 'x' }
}
`

func TestParseClass(t *testing.T) {
	f := Parse("uc.ts", useCase)
	if len(f.Errors) > 0 {
		t.Fatalf("errors: %v", f.Errors)
	}
	if len(f.Items) != 3 {
		t.Fatalf("want 3 items, got %d", len(f.Items))
	}
	imp := f.Items[0].Import
	if imp == nil || !imp.TypeOnly || imp.Default != "User" || imp.Spec != "@/domain/types/User.ts" {
		t.Errorf("type import misread: %+v", imp)
	}
	c := f.Items[2]
	if c.Kind != ItemClass || c.Name != "ResolveUseCase" || !c.Exported || !c.Default {
		t.Fatalf("class misread: %+v", c)
	}
	var names []string
	for _, m := range c.Members {
		names = append(names, m.Name)
	}
	want := []string{"authService", "count", "constructor", "execute", "helper", "name"}
	if len(names) != len(want) {
		t.Fatalf("members %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("members %v, want %v", names, want)
		}
	}
	ctor := c.Members[2]
	if len(ctor.Params) != 2 || ctor.Params[1].Name != "other" || ctor.Params[1].Access != "private" || !ctor.Params[1].Readonly {
		t.Errorf("constructor parameters misread: %+v", ctor.Params)
	}
	exec := c.Members[3]
	if exec.Kind != MemberMethod || exec.Body.Empty() || exec.Private() {
		t.Errorf("execute misread: %+v", exec)
	}
	if !c.Members[1].Static || c.Members[1].Kind != MemberProperty {
		t.Errorf("static property misread: %+v", c.Members[1])
	}
	if !c.Members[4].Private() || c.Members[4].Init.Empty() {
		t.Errorf("arrow property misread: %+v", c.Members[4])
	}
	if c.Members[5].Kind != MemberGetter {
		t.Errorf("getter misread: %+v", c.Members[5])
	}
}

func TestParseStatements(t *testing.T) {
	src := `import { a, type B, c as d } from './x';
import * as spec from '@/speclink';
import './side.css';
export { e } from './e';
export interface Port extends Base<T> {
  get(): Promise<User[]>;
}
export enum Status { A, B }
export type Alias = { a: string } | null;
let counter = 0
const router = createRouter({
  routes: [{ path: '/', component: View }]
});
router.beforeEach((to) => {
  return true;
});
export default router;
export function f<T>(x: T): T { return x }
`
	f := Parse("m.ts", src)
	if len(f.Errors) > 0 {
		t.Fatalf("errors: %v", f.Errors)
	}
	var got []ItemKind
	for _, it := range f.Items {
		got = append(got, it.Kind)
	}
	want := []ItemKind{ItemImport, ItemImport, ItemImport, ItemImport, ItemInterface, ItemEnum, ItemType,
		ItemVar, ItemVar, ItemExpr, ItemExpr, ItemFunction}
	if len(got) != len(want) {
		t.Fatalf("items %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("items %v, want %v", got, want)
		}
	}
	named := f.Items[0].Import.Named
	if len(named) != 3 || !named[1].TypeOnly || named[2].Local != "d" {
		t.Errorf("named imports misread: %+v", named)
	}
	if f.Items[1].Import.Namespace != "spec" {
		t.Errorf("namespace import misread: %+v", f.Items[1].Import)
	}
	if !f.Items[3].Import.ReExport {
		t.Errorf("re-export misread: %+v", f.Items[3].Import)
	}
	if f.Items[7].VarKind != "let" || f.Items[7].Name != "counter" {
		t.Errorf("let misread: %+v", f.Items[7])
	}
	if !f.Items[10].Default {
		t.Errorf("export default expression misread: %+v", f.Items[10])
	}
	if f.Items[11].Name != "f" || !f.Items[11].Exported {
		t.Errorf("function misread: %+v", f.Items[11])
	}

	route := f.Value(f.Items[8].Vars[0].Init)
	if route.Kind != ValCall || route.Last() != "createRouter" {
		t.Fatalf("call misread: %+v", route)
	}
	routes, _ := route.Elems[0].Get("routes")
	if len(routes.Elems) != 1 {
		t.Fatalf("routes misread: %+v", routes)
	}
	path, _ := routes.Elems[0].Get("path")
	comp, _ := routes.Elems[0].Get("component")
	if path.Str != "/" || comp.Last() != "View" {
		t.Errorf("route misread: %+v %+v", path, comp)
	}
}

func TestValueLiterals(t *testing.T) {
	src := `export const RX = requirement({
  id: 'R-X-Y',
  kind: Kind.Functional,
  text: 'one ' +
    'two',
  detail: ` + "`multi\nline`" + `,
  derivedFrom: [RA, RB],
  sources: [{ doc: 'a.md', anchor: 'b' }],
  bad: compute(),
  worse: a ? b : c,
});`
	f := Parse("r.ts", src)
	v := f.Value(f.Items[0].Vars[0].Init)
	if v.Kind != ValCall || len(v.Elems) != 1 {
		t.Fatalf("call misread: %+v", v)
	}
	o := v.Elems[0]
	check := func(key string, kind ValueKind, str string) {
		t.Helper()
		got, ok := o.Get(key)
		if !ok || got.Kind != kind || str != "" && got.Str != str {
			t.Errorf("%s: got %+v", key, got)
		}
	}
	check("id", ValString, "R-X-Y")
	check("kind", ValMember, "")
	check("text", ValString, "one two")
	check("detail", ValString, "multi\nline")
	check("derivedFrom", ValArray, "")
	check("bad", ValCall, "")
	check("worse", ValOther, "")
	src2, _ := o.Get("sources")
	if doc, _ := src2.Elems[0].Get("doc"); doc.Str != "a.md" {
		t.Errorf("nested object misread: %+v", src2)
	}
}

func TestASIDoesNotSplitChains(t *testing.T) {
	src := "keycloak\n  .init({})\n  .then(() => {})\nconst a = 1\n"
	f := Parse("m.ts", src)
	if len(f.Items) != 2 || f.Items[0].Kind != ItemExpr || f.Items[1].Kind != ItemVar {
		t.Fatalf("items misread: %+v", f.Items)
	}
}
