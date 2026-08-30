package golang

import (
	"go/ast"
	"go/token"
	"sort"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// ReadTopology extracts the actors, foreign systems and channels declared in a
// package's *.topology.go files.
func (p *Package) ReadTopology(out *diag.Set) ([]ir.Participant, []ir.Channel) {
	var (
		parts    []ir.Participant
		channels []ir.Channel
	)

	for _, f := range p.topologyFiles {
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.VAR {
				continue
			}
			for _, s := range gd.Specs {
				vs, ok := s.(*ast.ValueSpec)
				if !ok || len(vs.Names) != 1 || len(vs.Values) != 1 {
					continue
				}
				lit, ok := vs.Values[0].(*ast.CompositeLit)
				if !ok {
					continue
				}
				switch {
				case p.isSpecType(lit, "Actor"):
					parts = append(parts, p.readParticipant(ir.ParticipantActor, vs, lit))
				case p.isSpecType(lit, "Foreign"):
					parts = append(parts, p.readParticipant(ir.ParticipantForeign, vs, lit))
				case p.isSpecType(lit, "Channel"):
					channels = append(channels, p.readChannel(vs, lit))
				}
			}
		}
	}
	return parts, channels
}

func (p *Package) readParticipant(kind ir.ParticipantKind, vs *ast.ValueSpec, lit *ast.CompositeLit) ir.Participant {
	part := ir.Participant{Kind: kind, Pos: p.pos(vs.Pos())}
	for key, value := range p.fieldsOfLit(lit) {
		switch key {
		case "ID":
			part.ID, _ = p.stringArg(value)
		case "Name":
			part.Name, _ = p.stringArg(value)
		case "Role":
			part.Role, _ = p.stringArg(value)
		case "Satisfies":
			part.Satisfies = p.identList(value)
		case "Topics":
			part.Topics = p.identList(value)
		}
	}
	return part
}

func (p *Package) readChannel(vs *ast.ValueSpec, lit *ast.CompositeLit) ir.Channel {
	ch := ir.Channel{Pos: p.pos(vs.Pos())}
	for key, value := range p.fieldsOfLit(lit) {
		switch key {
		case "From":
			ch.From, _ = p.stringArg(value)
		case "To":
			ch.To, _ = p.stringArg(value)
		case "Label":
			ch.Label, _ = p.stringArg(value)
		case "Protocol":
			ch.Protocol, _ = p.stringArg(value)
		case "Data":
			ch.Data, _ = p.stringArg(value)
		case "Auth":
			ch.Auth, _ = p.stringArg(value)
		case "Crypto":
			ch.Crypto, _ = p.stringArg(value)
		case "Satisfies":
			ch.Satisfies = p.identList(value)
		case "Topics":
			ch.Topics = p.identList(value)
		case "Contract":
			// The declaration gives a zero value, so the type of the
			// expression is the contract. Read the same way a request body
			// is, so that a change inside a nested struct reaches the string
			// a rule compares.
			if t := p.pkg.TypesInfo.TypeOf(value); t != nil {
				ch.Contract = p.wireShape(t)
			}
		}
	}
	return ch
}

// fieldsOfLit returns the keyed fields of a composite literal.
func (p *Package) fieldsOfLit(lit *ast.CompositeLit) map[string]ast.Expr {
	out := map[string]ast.Expr{}
	for _, el := range lit.Elts {
		kv, ok := el.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		if key, ok := kv.Key.(*ast.Ident); ok {
			out[key.Name] = kv.Value
		}
	}
	return out
}

// Topology collects the declarations and the facts they are held against.
//
// Packages and Adapters come from the measured set rather than from everything
// loaded, so that an endpoint naming a package the scope left out is treated as
// unmeasured rather than as a mistake.
func (m *Model) Topology(out *diag.Set) ir.Topology {
	t := ir.Topology{Packages: map[string]bool{}}

	for _, p := range m.Measured {
		parts, channels := p.ReadTopology(out)
		t.Participants = append(t.Participants, parts...)
		t.Channels = append(t.Channels, channels...)

		rel := p.relDir(m.Root)
		if rel == "" {
			continue
		}
		t.Packages[rel] = true
		// An adapter is where the system touches something outside. That is
		// the architecture's own statement, read from the layout rather than
		// guessed from an import: a framework import crosses no boundary, and
		// a heuristic over import paths would call every one of them a channel.
		if isAdapterPackage(rel) {
			t.Adapters = append(t.Adapters, ir.Adapter{Dir: rel, Pkg: p.PkgPath(), Pos: p.firstPos()})
		}
	}

	sort.Slice(t.Adapters, func(i, j int) bool { return t.Adapters[i].Dir < t.Adapters[j].Dir })
	return t
}

// firstPos points at the head of a package, for a finding about the package
// rather than about anything in it.
func (p *Package) firstPos() ir.Position {
	if len(p.pkg.Syntax) == 0 {
		return ir.Position{}
	}
	return p.pos(p.pkg.Syntax[0].Package)
}
