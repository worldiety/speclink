package ts

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang/ts/syntax"
)

// ConstructUseCase is a class in the use case directory: one operation
// somebody asked for, and the unit a requirement is written about.
var ConstructUseCase = ir.NewConstructKind("use case", "a use case", ir.NeedsRequirement(), ir.PerformsWork())

// classOf is a class declaration together with where it was found.
type classOf struct {
	Module *Module
	File   *syntax.File
	Item   *syntax.Item
}

// Name qualifies the class the way constructs and bindings spell it.
func (c classOf) Name() string { return c.Module.ID + "." + c.Item.Name }

// classes returns the named classes of a module.
func (m *Module) classes() []classOf {
	var out []classOf
	for _, it := range m.Items() {
		if it.Kind == syntax.ItemClass && it.Name != "" {
			out = append(out, classOf{m, it.File, it.Item})
		}
	}
	return out
}

// useCases returns every class in the use case directory.
func (p *Project) useCases() []classOf {
	var out []classOf
	for _, m := range p.Modules {
		if p.Layer(m) == LayerUseCase {
			out = append(out, m.classes()...)
		}
	}
	return out
}

// Constructs recognises the use cases.
//
// Only use cases. A service, a port and an adapter are structure, covered
// through the use cases that reach them, and demanding a requirement for each
// would only produce noise. What the style says carries business meaning is
// the use case, and that is the set forward coverage holds accountable.
func (p *Project) Constructs() []ir.Construct {
	var out []ir.Construct
	for _, c := range p.useCases() {
		src := c.File.Source(c.Item.Range)
		sum := sha256.Sum256([]byte(src))
		out = append(out, ir.Construct{
			Kind:        ConstructUseCase,
			Name:        c.Name(),
			Package:     c.Module.Dir,
			Evidence:    "is a class in " + p.srcPath(dirUseCases),
			Pos:         c.Module.Pos(c.Item.Pos),
			EndLine:     c.File.EndLine(c.Item.Range),
			Fingerprint: hex.EncodeToString(sum[:]),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
