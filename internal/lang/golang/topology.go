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
func (p *Package) ReadTopology(out *diag.Set) ([]ir.Participant, []ir.Channel, []ir.Message) {
	var (
		parts    []ir.Participant
		channels []ir.Channel
		messages []ir.Message
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
				case p.isSpecType(lit, "Message"):
					messages = append(messages, p.readMessage(vs, lit))
				}
			}
		}
	}
	return parts, channels, messages
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
		case "Envelope":
			if t := p.pkg.TypesInfo.TypeOf(value); t != nil {
				ch.Envelope = p.wireShape(t)
			}
		case "Messages":
			ch.MessageRefs = p.identList(value)
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

// readMessage reads one message declaration.
func (p *Package) readMessage(vs *ast.ValueSpec, lit *ast.CompositeLit) ir.Message {
	m := ir.Message{
		GoIdent: p.PkgPath() + "." + vs.Names[0].Name,
		Pos:     p.pos(vs.Pos()),
	}
	for key, value := range p.fieldsOfLit(lit) {
		switch key {
		case "Payload":
			// The declaration gives a zero value, so the type of the
			// expression is the payload. Read the same way a request body is,
			// so a change inside a nested struct reaches the string a rule
			// compares.
			if t := p.pkg.TypesInfo.TypeOf(value); t != nil {
				m.Payload = p.wireShape(t)
				m.PayloadType = typeName(t)
			}
		case "Ack":
			if t := p.pkg.TypesInfo.TypeOf(value); t != nil {
				m.AckType = typeName(t)
			}
		case "From":
			m.From, _ = p.stringArg(value)
		case "To":
			m.To, _ = p.stringArg(value)
		case "Purpose":
			m.Purpose, _ = p.stringArg(value)
		case "Trigger":
			m.Trigger, _ = p.stringArg(value)
		case "Repeatable":
			m.Repeatable = p.readAnswer(value)
		case "Satisfies":
			m.Satisfies = p.identList(value)
		case "Topics":
			m.Topics = p.identList(value)
		}
	}
	return m
}

// readAnswer resolves a spec.Answer constant by the name it is written with.
//
// A computed value leaves it unanswered rather than guessed at, and unanswered
// is itself reported: a promise about redelivery worked out at run time is one
// nobody can read from the source.
func (p *Package) readAnswer(e ast.Expr) ir.Answer {
	sel, ok := e.(*ast.SelectorExpr)
	if !ok {
		return ir.Unanswered
	}
	a, _ := ir.AnswerOf(sel.Sel.Name)
	return a
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
	var declared []ir.Message

	for _, p := range m.Measured {
		parts, channels, messages := p.ReadTopology(out)
		t.Participants = append(t.Participants, parts...)
		t.Channels = append(t.Channels, channels...)
		declared = append(declared, messages...)

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
	resolveMessages(&t, declared)
	return t
}

// resolveMessages attaches the declared messages to the channels that list
// them.
//
// The second pass exists for the reason the requirement tree has one: a message
// may be declared after the channel naming it, and the order of files must not
// decide what a channel is said to carry. A reference resolving to nothing is
// left for the checks to report, with the channel it was named on.
func resolveMessages(t *ir.Topology, declared []ir.Message) {
	byIdent := make(map[string]ir.Message, len(declared))
	for _, m := range declared {
		byIdent[m.GoIdent] = m
	}
	for i := range t.Channels {
		for _, ref := range t.Channels[i].MessageRefs {
			if m, ok := byIdent[ref]; ok {
				t.Channels[i].Messages = append(t.Channels[i].Messages, m)
			}
		}
	}
	t.DeclaredMessages = declared
}

// firstPos points at the head of a package, for a finding about the package
// rather than about anything in it.
func (p *Package) firstPos() ir.Position {
	if len(p.pkg.Syntax) == 0 {
		return ir.Position{}
	}
	return p.pos(p.pkg.Syntax[0].Package)
}
