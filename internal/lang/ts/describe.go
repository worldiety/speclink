package ts

import (
	"github.com/worldiety/speclink/internal/ir"
)

// DescribeArchitecture states the shape the project is held to.
//
// Every sentence corresponds to a rule CheckArchitecture runs, with the paths
// the project configured rather than the profile's defaults.
func (p *Project) DescribeArchitecture() ir.Architecture {
	a := ir.Architecture{Style: "hex1"}
	src := p.srcPath

	a.Layers = []ir.Layer{
		{Name: "Use cases", Where: src(dirUseCases), Purpose: "One class per operation somebody asked for, with a single execute method. The unit a requirement is written about."},
		{Name: "Services", Where: src(dirServices), Purpose: "The business logic use cases are built on. They own state and reach the outside only through ports."},
		{Name: "Ports", Where: src(dirPorts), Purpose: "Interfaces stating, in the domain's terms, what the domain needs from the outside."},
		{Name: "Domain types", Where: src(dirTypes), Purpose: "The vocabulary every other layer is written in: interfaces, enums and type aliases."},
		{Name: "User interface", Where: src(dirUI), Purpose: "Views, components, the router and the translations. It reaches the domain through use cases only."},
		{Name: "Adapters", Where: src(dirAdapters) + "/<technology>", Purpose: "The other side of a port: HTTP, local storage, an identity provider. One class per port."},
		{Name: "Entry point", Where: src(fileEntry), Purpose: "Where the application is assembled: adapters are chosen, services built over them and the locator filled."},
	}

	a.Rules = []ir.Rule{
		{ID: RuleUCFile, Statement: "A use case is the default exported class <Name>UseCase of " + src(dirUseCases) + "/<Name>UseCase.ts, alone in its file.", Why: "It is the unit this document is organised around and the unit a reviewer is asked to read."},
		{ID: RuleUCShape, Statement: "A use case offers exactly one public method, execute.", Why: "A second public method is a second operation, with its own requirement and permission, hiding inside the first."},
		{ID: RuleUCDeps, Statement: "A use case imports only services and domain types, and receives its services through the constructor.", Why: "Whatever is reached around the constructor cannot be replaced in a test or seen by a reader."},
		{ID: RuleUCStateless, Statement: "A use case keeps no module level or static state.", Why: "State outside the instance is shared by every caller and survives every test."},
		{ID: RuleUCAuthz, Statement: "A use case consults a permission in execute.", Why: "Authorisation placed anywhere else is authorisation that a second caller can skip."},
		{ID: RuleDomainNoAdapter, Statement: "The domain does not import the user interface, an adapter or the entry point.", Why: "Dependencies point inwards; an adapter imported by the domain is no longer replaceable."},
		{ID: RuleDomainNoFramework, Statement: "The domain imports no package and uses no browser global.", Why: "The domain says what the system does, not with which library or on which platform."},
		{ID: RuleTypesPure, Statement: "Domain types import only domain types and declare no logic.", Why: "Every layer names them, so every dependency they carry is carried everywhere."},
		{ID: RulePortsPure, Statement: "A port file declares one exported interface <Name>Repository, in domain types only.", Why: "A port stated in a technology's terms can only be implemented by that technology."},
		{ID: RuleServiceNoUseCase, Statement: "A service does not import a use case.", Why: "Use cases are built on services, and the reverse makes the layers circular."},
		{ID: RuleAdapterImplements, Statement: "Every exported adapter class implements exactly one port and is named <Port>Impl.", Why: "The composition root then reads as a list of decisions, one per port."},
		{ID: RuleAdapterWiredInMain, Statement: "Only " + src(fileEntry) + " imports an adapter.", Why: "The choice of technology is made once, where the application is assembled."},
		{ID: RuleAdapterIsolated, Statement: "An adapter imports only ports, domain types, packages and its own technology's modules.", Why: "Hidden dependencies of an adapter are invisible to the composition root and to tests."},
		{ID: RuleUINoPort, Statement: "The user interface does not import a port.", Why: "A screen holding a port reads or writes data with nothing accountable in between."},
		{ID: RuleUIViaUseCase, Statement: "The user interface imports no service and uses the ServiceLocator only to construct a use case.", Why: "Every operation a screen performs then has a name, a requirement and a permission check."},
		{ID: RuleLocatorInitInMain, Statement: "Only " + src(fileEntry) + " initialises the ServiceLocator.", Why: "Which services the application runs with is decided in one place."},
		{ID: RuleLayerUnknown, Statement: "Every module below " + p.srcRoot() + " belongs to a layer.", Why: "A module in no layer is held to no rule."},
		{ID: RuleRouteTargetsView, Statement: "A route shows a <Name>View.vue from " + src(dirViews) + ", and that directory holds nothing else.", Why: "The router then reads as a table of contents of the application."},
		{ID: RuleI18nKeyParity, Statement: "Every translation catalogue has the same keys.", Why: "A missing text falls back to another language without a word."},
		{ID: RuleI18nKeyExists, Statement: "Every literal translation key the user interface uses is defined.", Why: "A missing key is shown to the user as the key itself."},
		{ID: RuleMainExists, Statement: "The project has " + src(fileEntry) + ".", Why: "Without an entry point every statement about wiring is about nothing."},
		{ID: RuleMainComposition, Statement: src(fileEntry) + " creates the application, fills the ServiceLocator and mounts the application.", Why: "Wiring done anywhere else is impossible to find and to review."},
	}
	return a
}
