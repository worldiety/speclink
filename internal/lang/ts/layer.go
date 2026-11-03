package ts

import (
	"path"
	"strings"
)

// Layer is the place a module has in the hex1 style.
//
// The places are the style's, not the framework's. Nothing in Vue says a use
// case is a class in domain/usecases or that a port is an interface in
// domain/repositories; that is an architecture, and it is selectable because a
// project can reasonably follow a different one.
type Layer int

const (
	// LayerNone is a module below the source root that belongs to no layer.
	LayerNone Layer = iota
	LayerUseCase
	LayerService
	// LayerLocator is the service locator, the registry the user interface
	// takes its services from to build a use case.
	LayerLocator
	LayerPort
	LayerType
	LayerUI
	// LayerAdapter is a driven adapter: the other side of a port.
	LayerAdapter
	LayerEntry
	// LayerVocabulary is the project's speclink.ts.
	LayerVocabulary
	// LayerSpec is a requirement or annotation file.
	LayerSpec
	LayerTest
	// LayerOutside is a module outside the source root: build configuration
	// and the requirement tree.
	LayerOutside
)

func (l Layer) String() string {
	switch l {
	case LayerUseCase:
		return "use case"
	case LayerService:
		return "service"
	case LayerLocator:
		return "service locator"
	case LayerPort:
		return "port"
	case LayerType:
		return "domain type"
	case LayerUI:
		return "user interface"
	case LayerAdapter:
		return "adapter"
	case LayerEntry:
		return "entry point"
	case LayerVocabulary:
		return "speclink vocabulary"
	case LayerSpec:
		return "specification"
	case LayerTest:
		return "test"
	case LayerOutside:
		return "outside the source root"
	}
	return "no layer"
}

// domain reports whether a layer is part of the domain.
func (l Layer) domain() bool {
	switch l {
	case LayerUseCase, LayerService, LayerLocator, LayerPort, LayerType:
		return true
	}
	return false
}

// The directories of the style, relative to the source root.
const (
	dirUseCases = "domain/usecases"
	dirServices = "domain/services"
	dirPorts    = "domain/repositories"
	dirTypes    = "domain/types"
	dirUI       = "adapters/ui"
	dirAdapters = "adapters"
	dirViews    = "adapters/ui/views"
	fileEntry   = "main.ts"
	fileLocator = "ServiceLocator.ts"
)

func (p *Project) srcRoot() string {
	if p.Layout.SrcRoot == "" {
		return "src"
	}
	return p.Layout.SrcRoot
}

// underSrc returns the path of a module relative to the source root.
func (p *Project) underSrc(m *Module) (string, bool) {
	root := p.srcRoot()
	if root == "." {
		return m.Path, true
	}
	return strings.CutPrefix(m.Path, root+"/")
}

// Layer classifies a module.
func (p *Project) Layer(m *Module) Layer {
	base := path.Base(m.Path)
	switch {
	case p.isTest(m):
		return LayerTest
	case strings.HasSuffix(base, ".req.ts") || strings.HasSuffix(base, ".annotation.ts"):
		return LayerSpec
	case base == VocabularyFile:
		return LayerVocabulary
	}
	rel, ok := p.underSrc(m)
	if !ok {
		return LayerOutside
	}
	switch {
	case rel == fileEntry:
		return LayerEntry
	case under(rel, dirUseCases):
		return LayerUseCase
	case rel == dirServices+"/"+fileLocator:
		return LayerLocator
	case under(rel, dirServices):
		return LayerService
	case under(rel, dirPorts):
		return LayerPort
	case under(rel, dirTypes):
		return LayerType
	case under(rel, dirUI):
		return LayerUI
	case under(rel, dirAdapters) && strings.Count(strings.TrimPrefix(rel, dirAdapters+"/"), "/") >= 1:
		return LayerAdapter
	}
	return LayerNone
}

// technology returns the directory an adapter lives in, http or localstorage.
func (p *Project) technology(m *Module) string {
	rel, _ := p.underSrc(m)
	rest := strings.TrimPrefix(rel, dirAdapters+"/")
	if i := strings.IndexByte(rest, '/'); i >= 0 {
		return rest[:i]
	}
	return ""
}

func (p *Project) isTest(m *Module) bool {
	base := path.Base(m.Path)
	if strings.Contains(base, ".test.") || strings.Contains(base, ".spec.") {
		return true
	}
	for _, r := range p.testRoots() {
		if r == "." || under(m.Path, r) {
			return true
		}
	}
	return false
}

func (p *Project) testRoots() []string {
	if len(p.Layout.TestRoots) == 0 {
		return []string{"tests"}
	}
	return p.Layout.TestRoots
}

func under(rel, dir string) bool { return strings.HasPrefix(rel, dir+"/") }

// srcPath spells a style directory as the project has it, for messages.
func (p *Project) srcPath(dir string) string {
	if p.srcRoot() == "." {
		return dir
	}
	return p.srcRoot() + "/" + dir
}
