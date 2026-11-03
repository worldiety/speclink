package ts

import (
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang/ts/syntax"
)

// The rules of the hex1 style.
//
// Ports and adapters over a domain of use cases, services, ports and types,
// with the user interface as one adapter among several. Every rule here is a
// direction of dependency or a shape a reader relies on, and every one of them
// was a choice in the project the style was derived from rather than an
// accident of it.
const (
	RuleUCFile             = "K5-UC-FILE"
	RuleUCShape            = "K5-UC-SHAPE"
	RuleUCDeps             = "K5-UC-DEPS"
	RuleUCStateless        = "K5-UC-STATELESS"
	RuleUCAuthz            = "K5-UC-AUTHZ"
	RuleDomainNoAdapter    = "K6-DOMAIN-NO-ADAPTER-IMPORT"
	RuleDomainNoFramework  = "K6-DOMAIN-NO-FRAMEWORK"
	RuleTypesPure          = "K6-TYPES-PURE"
	RulePortsPure          = "K6-PORTS-PURE"
	RuleServiceNoUseCase   = "K6-SERVICE-NO-USECASE"
	RuleAdapterImplements  = "K6-ADAPTER-IMPLEMENTS-PORT"
	RuleAdapterWiredInMain = "K6-ADAPTER-WIRED-IN-MAIN"
	RuleAdapterIsolated    = "K6-ADAPTER-ISOLATED"
	RuleUINoPort           = "K6-UI-NO-PORT"
	RuleUIViaUseCase       = "K6-UI-VIA-USECASE"
	RuleLocatorInitInMain  = "K6-LOCATOR-INIT-IN-MAIN"
	RuleLayerUnknown       = "K6-LAYER-UNKNOWN"
	RuleRouteTargetsView   = "K7-ROUTE-TARGETS-VIEW"
	RuleI18nKeyParity      = "K7-I18N-KEY-PARITY"
	RuleI18nKeyExists      = "K7-I18N-KEY-EXISTS"
	RuleMainExists         = "K8-MAIN-EXISTS"
	RuleMainComposition    = "K8-MAIN-COMPOSITION"
)

// browserGlobals are the names through which code reaches the platform it runs
// on. The domain names none of them: what it needs from the platform it asks
// for through a port.
var browserGlobals = map[string]bool{
	"window": true, "document": true, "localStorage": true, "sessionStorage": true,
	"fetch": true, "navigator": true, "location": true, "XMLHttpRequest": true,
	"indexedDB": true, "WebSocket": true,
}

type checker struct {
	p      *Project
	out    *diag.Set
	waived ir.Waivers
}

func (c *checker) add(subjects []string, f diag.Finding) {
	for _, s := range subjects {
		if c.waived.Has(s, f.Rule) {
			return
		}
	}
	c.out.Add(f)
}

func finding(code int, rule string, pos ir.Position, what, why, how string) diag.Finding {
	return diag.Finding{Code: diag.Code(diag.PhaseSemantic, code), Rule: rule, Pos: pos, What: what, Why: why, How: how}
}

// CheckArchitecture enforces the hex1 style over the measured modules.
func (p *Project) CheckArchitecture(out *diag.Set) {
	c := &checker{p: p, out: out, waived: ir.CollectWaivers(p.Bindings(&diag.Set{}))}

	for _, m := range p.Modules {
		if !p.Layout.InScope(m.Dir) {
			continue
		}
		layer := p.Layer(m)
		switch layer {
		case LayerOutside, LayerTest, LayerSpec, LayerVocabulary:
			continue
		case LayerNone:
			c.layerUnknown(m)
			continue
		}
		subjects := []string{m.ID}

		if layer.domain() {
			c.domainImports(m, subjects)
			c.domainGlobals(m, subjects)
		}
		switch layer {
		case LayerUseCase:
			c.useCase(m)
		case LayerService:
			c.service(m, subjects)
		case LayerType:
			c.types(m, subjects)
		case LayerPort:
			c.port(m, subjects)
		case LayerAdapter:
			c.adapter(m, subjects)
		case LayerUI:
			c.ui(m, subjects)
		}
		if layer != LayerUseCase {
			c.useCaseElsewhere(m)
		}
		if layer != LayerEntry && layer != LayerAdapter && !layer.domain() {
			// The domain reports the same import under its own rule.
			c.adapterImports(m, subjects)
		}
		if layer != LayerEntry {
			c.locatorInit(m, subjects)
		}
	}
	c.routes()
	c.i18n()
	c.entry()
}

func (c *checker) layerUnknown(m *Module) {
	c.add([]string{m.ID}, finding(256, RuleLayerUnknown, ir.Position{File: m.Abs, Line: 1},
		m.Path+" belongs to no layer of the style.",
		"Every rule of the style is a statement about a layer. A module in none of them is held to nothing, and code that is held to nothing is where an architecture quietly ends.",
		"Move it to "+c.p.srcPath(dirUseCases)+", "+c.p.srcPath(dirServices)+", "+c.p.srcPath(dirPorts)+", "+
			c.p.srcPath(dirTypes)+", "+c.p.srcPath(dirUI)+" or "+c.p.srcPath(dirAdapters)+"/<technology>."))
}

// --- K5: use cases ---

func baseName(file string) string {
	b := path.Base(file)
	if i := strings.IndexByte(b, '.'); i >= 0 {
		return b[:i]
	}
	return b
}

func (c *checker) useCase(m *Module) {
	classes := m.classes()
	subjects := []string{m.ID}
	for _, cl := range classes {
		subjects = append(subjects, cl.Name())
	}
	pos := ir.Position{File: m.Abs, Line: 1}
	want := baseName(m.Path)

	switch {
	case m.Vue:
		c.add(subjects, finding(230, RuleUCFile, pos,
			m.Path+" is a component in the use case directory.",
			"A use case is what the system does, and a component is one way of showing it. Kept apart, the same use case can be reached from a second screen or a test without a renderer.",
			"Move the component to "+c.p.srcPath(dirUI)+" and the work it does into a use case class."))
		return
	case len(classes) == 0:
		c.add(subjects, finding(230, RuleUCFile, pos,
			m.Path+" declares no use case.",
			"The use case directory holds use cases and nothing else, one per file. A file without one is either misplaced or a use case written as something the rules cannot see.",
			"Declare export default class "+want+" { execute(…) { … } }, or move the file to the layer it belongs to."))
		return
	case len(classes) > 1:
		var names []string
		for _, cl := range classes {
			names = append(names, cl.Item.Name)
		}
		c.add(subjects, finding(230, RuleUCFile, m.Pos(classes[1].Item.Pos),
			m.Path+" declares "+itoa(len(classes))+" classes: "+strings.Join(names, ", ")+".",
			"One use case per file, named after it. It is the unit this document is organised around and the unit a reviewer is asked to read; a file holding several cannot be reviewed as any of them.",
			"Move every class but "+want+" into a file of its own."))
		return
	}

	cl := classes[0]
	at := m.Pos(cl.Item.Pos)
	if !cl.Item.Default {
		c.add(subjects, finding(230, RuleUCFile, at,
			cl.Item.Name+" is not the default export of its file.",
			"Every use case is imported the same way, by default, so that a reader can tell from an import alone which file holds which use case.",
			"Write export default class "+cl.Item.Name+"."))
	}
	if !strings.HasSuffix(cl.Item.Name, "UseCase") || cl.Item.Name != want {
		c.add(subjects, finding(231, RuleUCFile, at,
			"the use case "+cl.Item.Name+" is declared in "+path.Base(m.Path)+".",
			"A use case is named <Action>UseCase and lives in a file of the same name, so that the name a screen constructs is the name of the file that implements it.",
			"Name the class and the file alike, ending in UseCase."))
	}

	c.useCaseShape(cl, subjects)
	c.useCaseDeps(m, subjects)
	c.useCaseState(m, cl, subjects)
	c.useCaseAuthz(cl, subjects)
}

// isFunctionValue reports whether an initializer is an arrow function or a
// function expression, which makes a property a method in everything but name.
func isFunctionValue(f *syntax.File, r syntax.Range) bool {
	toks := f.Tokens(r)
	if len(toks) == 0 {
		return false
	}
	i := 0
	if toks[0].Is("async") {
		i++
	}
	if i < len(toks) && toks[i].Is("function") {
		return true
	}
	if i < len(toks) && toks[i].Is("(") {
		end := f.GroupEnd(r.From+i) - r.From
		for j := end; j < len(toks); j++ {
			if toks[j].Is("=>") {
				return true
			}
			if toks[j].Is("{") || toks[j].Is("=") {
				break
			}
		}
		return false
	}
	return i+1 < len(toks) && toks[i].Kind == syntax.Ident && toks[i+1].Is("=>")
}

// publicMembers lists what a class offers to its callers.
func publicMembers(cl classOf) []*syntax.Member {
	var out []*syntax.Member
	for _, mem := range cl.Item.Members {
		switch {
		case mem.Kind == syntax.MemberConstructor:
			for _, prm := range mem.Params {
				if prm.Access == "public" {
					out = append(out, &syntax.Member{Kind: syntax.MemberProperty, Name: prm.Name, Pos: prm.Pos})
				}
			}
		case mem.Kind == syntax.MemberStatic || mem.Kind == syntax.MemberIndex:
		case mem.Private():
		default:
			out = append(out, mem)
		}
	}
	return out
}

func (c *checker) useCaseShape(cl classOf, subjects []string) {
	var execute *syntax.Member
	var extra []string
	for _, mem := range publicMembers(cl) {
		method := mem.Kind == syntax.MemberMethod || mem.Kind == syntax.MemberProperty && isFunctionValue(cl.File, mem.Init)
		if mem.Name == "execute" && method && !mem.Static && execute == nil {
			execute = mem
			continue
		}
		extra = append(extra, mem.Name)
	}
	at := cl.Module.Pos(cl.Item.Pos)
	if execute == nil {
		c.add(subjects, finding(233, RuleUCShape, at,
			cl.Item.Name+" has no public execute method.",
			"Every use case is called the same way, so that a screen, a test and a reader know without looking how to run one. A use case reached through a method of its own choosing is a service with a misleading name.",
			"Declare execute(…) and make everything else private."))
	}
	if len(extra) > 0 {
		c.add(subjects, finding(233, RuleUCShape, at,
			cl.Item.Name+" offers more than execute: "+strings.Join(extra, ", ")+".",
			"A use case is one operation somebody asked for. A second public method is a second operation, with its own requirement and its own permission, hiding inside the first.",
			"Make "+strings.Join(extra, ", ")+" private, or split each into a use case of its own."))
	}
}

func (c *checker) useCaseDeps(m *Module, subjects []string) {
	for _, ref := range m.Imports {
		var what string
		switch {
		case ref.Package != "":
			continue // reported as K6-DOMAIN-NO-FRAMEWORK
		case ref.Target == nil:
			what = ref.Local
		default:
			switch c.p.Layer(ref.Target) {
			case LayerService, LayerType:
				continue
			case LayerUI, LayerAdapter, LayerEntry:
				continue // reported as K6-DOMAIN-NO-ADAPTER-IMPORT
			case LayerUseCase:
				what = "the use case " + ref.Target.Path
			default:
				what = c.p.Layer(ref.Target).String() + " " + ref.Target.Path
			}
		}
		c.add(subjects, finding(234, RuleUCDeps, ref.Pos,
			path.Base(m.Path)+" imports "+what+".",
			"A use case depends on services, which it is handed through its constructor, and on domain types. Anything else is reached around the constructor, where a test cannot replace it and a reader cannot see it.",
			"Take what is needed as a service in the constructor, and let the caller pass it in."))
	}
}

func (c *checker) useCaseState(m *Module, cl classOf, subjects []string) {
	for _, it := range m.Items() {
		if it.Kind == syntax.ItemVar && it.VarKind != "const" {
			c.add(subjects, finding(235, RuleUCStateless, m.Pos(it.Pos),
				path.Base(m.Path)+" keeps state in a module level "+it.VarKind+".",
				"A use case is constructed for the call that needs it and holds nothing between calls. State outside the instance is shared by every caller and survives every test.",
				"Keep the value in a service, which owns state on purpose, and pass the service in."))
		}
	}
	for _, mem := range cl.Item.Members {
		if mem.Static && mem.Kind == syntax.MemberProperty && !mem.Readonly {
			c.add(subjects, finding(235, RuleUCStateless, cl.Module.Pos(mem.Pos),
				cl.Item.Name+"."+mem.Name+" is static and writable.",
				"A use case is constructed for the call that needs it and holds nothing between calls. A static field is shared by every caller and survives every test.",
				"Keep the value in a service, or make it static readonly if it is a constant."))
		}
	}
}

// consultsPermission reports whether a body asks about a permission: it reads
// a field of the permissions, or calls audit.
func consultsPermission(toks []syntax.Token) bool {
	for i := 1; i < len(toks); i++ {
		prev, t := toks[i-1], toks[i]
		if !(prev.Is(".") || prev.Is("?.")) || t.Kind != syntax.Ident {
			continue
		}
		if t.Text == "permissions" {
			return true
		}
		if t.Text == "audit" && i+1 < len(toks) && toks[i+1].Is("(") {
			return true
		}
	}
	return false
}

func (c *checker) useCaseAuthz(cl classOf, subjects []string) {
	for _, mem := range cl.Item.Members {
		if mem.Name != "execute" || mem.Private() {
			continue
		}
		body := mem.Body
		if body.Empty() {
			body = mem.Init
		}
		if consultsPermission(cl.File.Tokens(body)) {
			return
		}
		c.add(subjects, finding(236, RuleUCAuthz, cl.Module.Pos(mem.Pos),
			cl.Item.Name+".execute consults no permission.",
			"Authorisation placed anywhere else is authorisation that a second caller can skip. The screen hiding a button is a convenience; the use case refusing the call is the protection.",
			"Ask the AuthService before doing anything: read userInfo.permissions.<permission>, or call authService.audit('<permission>'). If the operation is genuinely open to every signed in user, waive the rule with the reason."))
		return
	}
}

func (c *checker) useCaseElsewhere(m *Module) {
	for _, cl := range m.classes() {
		if !strings.HasSuffix(cl.Item.Name, "UseCase") {
			continue
		}
		c.add([]string{m.ID, cl.Name()}, finding(232, RuleUCFile, m.Pos(cl.Item.Pos),
			cl.Item.Name+" is declared in "+m.Path+", outside "+c.p.srcPath(dirUseCases)+".",
			"Use cases are recognised by where they live. One outside that directory is invisible to every rule that holds use cases accountable, including the one that asks which requirement it serves.",
			"Move it to "+c.p.srcPath(dirUseCases)+"/"+cl.Item.Name+".ts."))
	}
}

// --- K6: directions ---

func (c *checker) domainImports(m *Module, subjects []string) {
	for _, ref := range m.Imports {
		switch {
		case ref.Package != "":
			c.add(subjects, finding(241, RuleDomainNoFramework, ref.Pos,
				path.Base(m.Path)+" is domain code and imports the package "+ref.Package+".",
				"The domain says what the system does in its own terms. A library imported here decides how, and the domain can then no longer be tested, reused or understood without it.",
				"Declare a port in "+c.p.srcPath(dirPorts)+" for what is needed, implement it with "+ref.Package+" in "+c.p.srcPath(dirAdapters)+"/<technology>, and pass it in from "+c.p.srcPath(fileEntry)+"."))
		case ref.Target != nil:
			switch l := c.p.Layer(ref.Target); l {
			case LayerUI, LayerAdapter, LayerEntry:
				c.add(subjects, finding(240, RuleDomainNoAdapter, ref.Pos,
					path.Base(m.Path)+" is domain code and imports the "+l.String()+" "+ref.Target.Path+".",
					"Dependencies point inwards. An adapter exists to be replaceable, and a domain that imports one has made it permanent.",
					"Depend on a port instead, and let "+c.p.srcPath(fileEntry)+" decide which adapter implements it."))
			}
		}
	}
}

func (c *checker) domainGlobals(m *Module, subjects []string) {
	for _, f := range m.Scripts {
		toks := f.Toks
		for i, t := range toks {
			name := ""
			switch {
			case t.Kind == syntax.Ident && browserGlobals[t.Text]:
				if i > 0 && (toks[i-1].Is(".") || toks[i-1].Is("?.")) {
					continue
				}
				if i+1 < len(toks) && toks[i+1].Is(":") {
					continue
				}
				name = t.Text
			case t.Is("import") && i+2 < len(toks) && toks[i+1].Is(".") && toks[i+2].Is("meta"):
				name = "import.meta"
			default:
				continue
			}
			c.add(subjects, finding(242, RuleDomainNoFramework, m.Pos(t.Pos),
				path.Base(m.Path)+" is domain code and uses "+name+".",
				"The domain says what the system does, not where it runs. Reaching the browser directly ties every rule in it to one platform and to a test environment that imitates it.",
				"Declare a port for what is needed and implement it in "+c.p.srcPath(dirAdapters)+"/<technology>."))
		}
	}
}

func (c *checker) service(m *Module, subjects []string) {
	for _, ref := range m.Imports {
		if ref.Target == nil || c.p.Layer(ref.Target) != LayerUseCase {
			continue
		}
		c.add(subjects, finding(247, RuleServiceNoUseCase, ref.Pos,
			path.Base(m.Path)+" is a service and imports the use case "+ref.Target.Path+".",
			"Use cases are built on services, not the other way round. A service that runs a use case has made the order of the layers circular, and a use case can then no longer be understood from what it is handed.",
			"Move the shared logic into a service both can use."))
	}
}

func (c *checker) types(m *Module, subjects []string) {
	for _, ref := range m.Imports {
		if ref.Target != nil && c.p.Layer(ref.Target) == LayerType {
			continue
		}
		what := ref.Package
		if what == "" && ref.Target != nil {
			what = ref.Target.Path
		}
		if what == "" {
			what = ref.Local
		}
		c.add(subjects, finding(243, RuleTypesPure, ref.Pos,
			path.Base(m.Path)+" is a domain type and imports "+what+".",
			"Domain types are the vocabulary every other layer is written in. A type that depends on anything but other types drags that dependency into every layer that names it.",
			"Import only from "+c.p.srcPath(dirTypes)+"."))
	}
	for _, it := range m.Items() {
		switch it.Kind {
		case syntax.ItemImport, syntax.ItemInterface, syntax.ItemEnum, syntax.ItemType, syntax.ItemExportList:
			continue
		}
		c.add(subjects, finding(244, RuleTypesPure, m.Pos(it.Pos),
			path.Base(m.Path)+" is a domain type and declares a "+it.Kind.String()+".",
			"Domain types describe data. Logic placed here is logic no use case owns and no permission guards.",
			"Keep interfaces, enums and type aliases here, and move the logic into a service."))
	}
}

func (c *checker) port(m *Module, subjects []string) {
	for _, ref := range m.Imports {
		if ref.Target != nil {
			if l := c.p.Layer(ref.Target); l == LayerType || l == LayerPort {
				continue
			}
		}
		what := ref.Package
		if what == "" && ref.Target != nil {
			what = ref.Target.Path
		}
		if what == "" {
			what = ref.Local
		}
		c.add(subjects, finding(245, RulePortsPure, ref.Pos,
			path.Base(m.Path)+" is a port and imports "+what+".",
			"A port is stated in the domain's terms so that any adapter can implement it. A port that names a technology can only be implemented by that technology.",
			"Speak in domain types only."))
	}
	var interfaces []itemOf
	for _, it := range m.Items() {
		switch it.Kind {
		case syntax.ItemImport, syntax.ItemType, syntax.ItemExportList:
			continue
		case syntax.ItemInterface:
			interfaces = append(interfaces, it)
			continue
		}
		c.add(subjects, finding(245, RulePortsPure, m.Pos(it.Pos),
			path.Base(m.Path)+" is a port and declares a "+it.Kind.String()+".",
			"A port is a promise, not an implementation. Code placed beside it is code no adapter can replace.",
			"Keep only the interface here; implement it in "+c.p.srcPath(dirAdapters)+"/<technology>."))
	}
	want := baseName(m.Path)
	if len(interfaces) != 1 || interfaces[0].Name != want || !interfaces[0].Exported || !strings.HasSuffix(want, "Repository") {
		c.add(subjects, finding(246, RulePortsPure, ir.Position{File: m.Abs, Line: 1},
			path.Base(m.Path)+" does not declare exactly one exported interface named "+want+".",
			"One port per file, named <Thing>Repository after the file, so that the adapter implementing it can be named after it and found by name.",
			"Declare export interface "+want+" { … } and nothing else; the name ends in Repository."))
	}
}

// portOf resolves the name of an implemented interface to the port it names.
func (c *checker) portOf(m *Module, name string) (string, bool) {
	head, _, _ := strings.Cut(name, ".")
	b, ok := m.bindings()[head]
	if !ok || b.ref.Target == nil || c.p.Layer(b.ref.Target) != LayerPort {
		return "", false
	}
	if b.name == "default" {
		return b.ref.Target.defaultExport(), true
	}
	if b.name == "*" {
		_, rest, _ := strings.Cut(name, ".")
		return rest, true
	}
	return b.name, true
}

func (c *checker) adapter(m *Module, subjects []string) {
	tech := c.p.technology(m)
	for _, ref := range m.Imports {
		if ref.Target == nil {
			continue
		}
		l := c.p.Layer(ref.Target)
		switch {
		case l == LayerPort || l == LayerType:
			continue
		case l == LayerAdapter && c.p.technology(ref.Target) == tech:
			continue
		case l == LayerOutside || l == LayerVocabulary:
			continue
		}
		c.add(subjects, finding(251, RuleAdapterIsolated, ref.Pos,
			path.Base(m.Path)+" is an adapter and imports the "+l.String()+" "+ref.Target.Path+".",
			"An adapter implements a port and knows nothing else of the application. One that reaches into services, the locator or another adapter has hidden dependencies the composition root cannot see and a test cannot replace.",
			"Take what is needed through the constructor — a value, or a function such as accessToken: () => string — and let "+c.p.srcPath(fileEntry)+" pass it in."))
	}
	for _, cl := range m.classes() {
		if !cl.Item.Exported {
			continue
		}
		var ports []string
		for _, impl := range cl.Item.Implements {
			if port, ok := c.portOf(m, impl); ok {
				ports = append(ports, port)
			}
		}
		subj := append([]string{cl.Name()}, subjects...)
		at := m.Pos(cl.Item.Pos)
		if len(ports) != 1 {
			c.add(subj, finding(248, RuleAdapterImplements, at,
				cl.Item.Name+" implements "+plural(len(ports), "port", "ports")+".",
				"An adapter is the other side of exactly one port. Without one it is an implementation nothing in the domain asked for; with several it cannot be replaced for one of them without the others.",
				"Implement one interface from "+c.p.srcPath(dirPorts)+": class <Port>Impl implements <Port>."))
			continue
		}
		if cl.Item.Name != ports[0]+"Impl" {
			c.add(subj, finding(249, RuleAdapterImplements, at,
				cl.Item.Name+" implements "+ports[0]+" and is not named "+ports[0]+"Impl.",
				"The name of an adapter says which port it serves, so that the composition root reads as a list of decisions.",
				"Rename it to "+ports[0]+"Impl."))
		}
	}
}

func (c *checker) adapterImports(m *Module, subjects []string) {
	for _, ref := range m.Imports {
		if ref.Target == nil || c.p.Layer(ref.Target) != LayerAdapter {
			continue
		}
		c.add(subjects, finding(250, RuleAdapterWiredInMain, ref.Pos,
			path.Base(m.Path)+" imports the adapter "+ref.Target.Path+".",
			"A port exists so the choice of technology is made once, where the application is assembled. An import anywhere else makes that choice permanent and invisible.",
			"Depend on the port, and let "+c.p.srcPath(fileEntry)+" construct the adapter."))
	}
}

// locatorNames returns the local names a module gives the service locator.
func (c *checker) locatorNames(m *Module) map[string]bool {
	out := map[string]bool{}
	for name, b := range m.bindings() {
		if b.ref.Target != nil && c.p.Layer(b.ref.Target) == LayerLocator {
			out[name] = true
		}
	}
	return out
}

func (c *checker) locatorInit(m *Module, subjects []string) {
	names := c.locatorNames(m)
	if c.p.Layer(m) == LayerLocator {
		return
	}
	for _, f := range m.Scripts {
		toks := f.Toks
		for i := 0; i+3 < len(toks); i++ {
			if toks[i].Kind != syntax.Ident || !names[toks[i].Text] || !toks[i+1].Is(".") || !toks[i+2].Is("init") || !toks[i+3].Is("(") {
				continue
			}
			c.add(subjects, finding(255, RuleLocatorInitInMain, m.Pos(toks[i].Pos),
				path.Base(m.Path)+" initialises the service locator.",
				"The locator is filled once, where the application is assembled. A second place that fills it decides which services the application runs with somewhere nobody looks.",
				"Call ServiceLocator.init only in "+c.p.srcPath(fileEntry)+"."))
		}
	}
}

var templateLocator = regexp.MustCompile(`\bServiceLocator\b`)

func (c *checker) ui(m *Module, subjects []string) {
	for _, ref := range m.Imports {
		if ref.Target == nil {
			continue
		}
		switch c.p.Layer(ref.Target) {
		case LayerPort:
			c.add(subjects, finding(252, RuleUINoPort, ref.Pos,
				path.Base(m.Path)+" is user interface and imports the port "+ref.Target.Path+".",
				"The user interface reaches the domain through use cases. A screen that holds a port reads or writes data with nothing accountable in between.",
				"Call a use case that does it."))
		case LayerService:
			c.add(subjects, finding(253, RuleUIViaUseCase, ref.Pos,
				path.Base(m.Path)+" is user interface and imports the service "+ref.Target.Path+".",
				"The user interface reaches the domain through use cases, so that every operation a screen performs has a name, a requirement and a permission check. A service called from a screen has none of the three.",
				"Write a use case for the operation and call that; take its services from the ServiceLocator."))
		}
	}

	names := c.locatorNames(m)
	useCases := map[string]bool{}
	for name, b := range m.bindings() {
		if b.ref.Target != nil && c.p.Layer(b.ref.Target) == LayerUseCase {
			useCases[name] = true
		}
	}
	for _, f := range m.Scripts {
		c.locatorOutsideUseCase(m, f, names, useCases, subjects)
	}
	for _, b := range m.Templates {
		for _, loc := range templateLocator.FindAllStringIndex(m.Src[b.From:b.To], -1) {
			off := b.From + loc[0]
			c.add(subjects, finding(254, RuleUIViaUseCase, ir.Position{File: m.Abs, Line: syntax.LineOf(m.Src, off)},
				path.Base(m.Path)+" reaches the ServiceLocator from its template.",
				"The user interface reaches the domain through use cases, so that every operation a screen performs has a name, a requirement and a permission check. A service called from a template has none of the three.",
				"Call a use case in the script and bind its result."))
		}
	}
}

// locatorOutsideUseCase finds every use of the locator that does not hand a
// service to a use case being constructed.
func (c *checker) locatorOutsideUseCase(m *Module, f *syntax.File, names, useCases map[string]bool, subjects []string) {
	type frame struct {
		end     int
		useCase bool
	}
	var stack []frame
	importing := map[int]bool{}
	for _, it := range f.Items {
		if it.Kind == syntax.ItemImport {
			for i := it.Range.From; i < it.Range.To; i++ {
				importing[i] = true
			}
		}
	}
	toks := f.Toks
	for i, t := range toks {
		for len(stack) > 0 && i >= stack[len(stack)-1].end {
			stack = stack[:len(stack)-1]
		}
		if t.Is("(") {
			uc := i >= 2 && toks[i-2].Is("new") && toks[i-1].Kind == syntax.Ident && useCases[toks[i-1].Text]
			stack = append(stack, frame{end: f.GroupEnd(i), useCase: uc})
			continue
		}
		if t.Kind != syntax.Ident || !names[t.Text] || importing[i] || i > 0 && toks[i-1].Is(".") {
			continue
		}
		inside := false
		for _, fr := range stack {
			if fr.useCase {
				inside = true
				break
			}
		}
		if inside {
			continue
		}
		c.add(subjects, finding(254, RuleUIViaUseCase, m.Pos(t.Pos),
			path.Base(m.Path)+" uses the ServiceLocator other than to construct a use case.",
			"The user interface reaches the domain through use cases, so that every operation a screen performs has a name, a requirement and a permission check. A service called directly from a screen has none of the three.",
			"Write a use case for the operation and call new <Name>UseCase(ServiceLocator.get…()).execute(…)."))
	}
}

// --- K7: views and translations ---

func (c *checker) routes() {
	for _, r := range c.p.Routes() {
		if !c.p.Layout.InScope(r.Router.Dir) {
			continue
		}
		subjects := []string{r.Router.ID}
		pos := r.Router.Pos(r.Pos)
		if r.Component == nil {
			if r.Unresolved {
				c.add(subjects, finding(257, RuleRouteTargetsView, pos,
					"the route "+r.Path+" shows a component this run cannot resolve.",
					"A route is an address the application answers on, and what it shows has to be a view somebody can find and review.",
					"Name an imported view: component: "+"<Name>View."))
			}
			continue
		}
		rel, _ := c.p.underSrc(r.Component)
		if !under(rel, dirViews) || !strings.HasSuffix(r.Component.Path, "View.vue") {
			c.add(subjects, finding(257, RuleRouteTargetsView, pos,
				"the route "+r.Path+" shows "+r.Component.Path+", which is not a view.",
				"A route shows a view, and a view is what a route shows. Kept apart from the components it is built of, a view is the one place a reader looks to learn what an address does.",
				"Wrap the component in "+c.p.srcPath(dirViews)+"/<Name>View.vue and route to that."))
		}
	}
	for _, m := range c.p.Modules {
		rel, ok := c.p.underSrc(m)
		if !ok || !under(rel, dirViews) || !c.p.Layout.InScope(m.Dir) {
			continue
		}
		if m.Vue && strings.HasSuffix(m.Path, "View.vue") {
			continue
		}
		c.add([]string{m.ID}, finding(258, RuleRouteTargetsView, ir.Position{File: m.Abs, Line: 1},
			m.Path+" is in the view directory and is not a <Name>View.vue component.",
			"The view directory holds what routes show, named so that the router reads as a table of contents. Anything else there is a component in the wrong place.",
			"Rename it to <Name>View.vue, or move it to "+c.p.srcPath(dirUI)+"/components."))
	}
}

var templateKey = regexp.MustCompile(`(?:^|[^A-Za-z0-9_$.])\$?t\(\s*(['"])([^'"]+)['"]`)

func (c *checker) i18n() {
	locales := c.p.Locales()
	if len(locales) == 0 {
		return
	}
	union := map[string]bool{}
	for _, l := range locales {
		for k := range l.Keys {
			union[k] = true
		}
	}
	for _, l := range locales {
		var missing []string
		for k := range union {
			if !l.Keys[k] {
				missing = append(missing, k)
			}
		}
		if len(missing) == 0 {
			continue
		}
		sort.Strings(missing)
		shown := missing
		if len(shown) > 10 {
			shown = append(append([]string(nil), shown[:10]...), "…")
		}
		c.add([]string{l.Path}, finding(259, RuleI18nKeyParity, ir.Position{File: l.Abs, Line: 1},
			l.Path+" lacks "+plural(len(missing), "key", "keys")+" the other catalogues have: "+strings.Join(shown, ", ")+".",
			"A text missing from one language falls back to another without a word. The screen looks finished to everyone who tests it in the language that has it.",
			"Add the translations."))
	}

	for _, m := range c.p.Modules {
		if c.p.Layer(m) != LayerUI || !c.p.Layout.InScope(m.Dir) {
			continue
		}
		for _, use := range c.p.keysUsed(m) {
			if union[use.key] {
				continue
			}
			c.add([]string{m.ID}, finding(260, RuleI18nKeyExists, use.pos,
				path.Base(m.Path)+" uses the text "+use.key+", which no catalogue defines.",
				"A missing key is shown to the user as the key itself. Nothing fails, which is why it has to be checked.",
				"Add "+use.key+" to every catalogue under "+c.p.srcPath(dirUI)+"/**/locales."))
		}
	}
}

type keyUse struct {
	key string
	pos ir.Position
}

// keysUsed returns the literal translation keys a component or module uses.
func (p *Project) keysUsed(m *Module) []keyUse {
	var out []keyUse
	for _, f := range m.Scripts {
		toks := f.Toks
		for i := 0; i+2 < len(toks); i++ {
			t := toks[i]
			if t.Kind != syntax.Ident || t.Text != "t" && t.Text != "$t" || !toks[i+1].Is("(") || toks[i+2].Kind != syntax.String {
				continue
			}
			if i > 0 && toks[i-1].Is(".") && !(i > 1 && toks[i-2].Is("global")) {
				continue
			}
			out = append(out, keyUse{toks[i+2].Text, m.Pos(toks[i+2].Pos)})
		}
	}
	for _, b := range m.Templates {
		text := m.Src[b.From:b.To]
		for _, loc := range templateKey.FindAllStringSubmatchIndex(text, -1) {
			off := b.From + loc[4]
			out = append(out, keyUse{text[loc[4]:loc[5]], ir.Position{File: m.Abs, Line: syntax.LineOf(m.Src, off)}})
		}
	}
	return out
}

// --- K8: the entry point ---

func (c *checker) entry() {
	main := c.p.Module(c.p.srcPath(fileEntry))
	if main == nil {
		for _, b := range c.p.Bindings(&diag.Set{}) {
			for _, a := range b.Assertions {
				if a.Kind == ir.AssertWaive && a.Rule == RuleMainExists {
					return
				}
			}
		}
		c.out.Add(finding(261, RuleMainExists, ir.Position{File: c.p.Root}, "the project has no "+c.p.srcPath(fileEntry)+".",
			"The entry point is where the application is assembled: the one place adapters are chosen and the locator is filled. Without it every statement about wiring is about nothing.",
			"Add "+c.p.srcPath(fileEntry)+", or waive the rule if this project is a library."))
		return
	}
	if !c.p.Layout.InScope(main.Dir) {
		return
	}
	var missing []string
	has := func(seq ...string) bool {
		for _, f := range main.Scripts {
			for i := 0; i+len(seq) <= len(f.Toks); i++ {
				ok := true
				for j, s := range seq {
					if !f.Toks[i+j].Is(s) {
						ok = false
						break
					}
				}
				if ok {
					return true
				}
			}
		}
		return false
	}
	if !has("createApp", "(") {
		missing = append(missing, "createApp(…)")
	}
	if !has(".", "mount", "(") {
		missing = append(missing, ".mount(…)")
	}
	if !has("ServiceLocator", ".", "init", "(") {
		missing = append(missing, "ServiceLocator.init(…)")
	}
	if len(missing) > 0 {
		c.add([]string{main.ID}, finding(262, RuleMainComposition, ir.Position{File: main.Abs, Line: 1},
			main.Path+" does not assemble the application: "+strings.Join(missing, ", ")+" is missing.",
			"The entry point is the composition root. It creates the application, fills the locator with services built over their adapters, and mounts it; done anywhere else, the wiring is impossible to find and impossible to review.",
			"Do all three in "+main.Path+"."))
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return itoa(n) + " " + many
}
