package check

import (
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/baseline"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/reqtree"
)

const (
	// RuleMessageEndUnknown fires when a message names an end that is not one
	// of the two ends of the channel carrying it.
	RuleMessageEndUnknown = "K17-MESSAGE-END-UNKNOWN"
	// RuleMessageIncomplete fires when a message leaves out what it is for,
	// when it is sent, or what it carries.
	RuleMessageIncomplete = "K17-MESSAGE-INCOMPLETE"
	// RuleMessageRepeatUnstated fires when a message does not say whether it
	// may be delivered twice.
	RuleMessageRepeatUnstated = "K17-MESSAGE-REPEAT-UNSTATED"
	// RuleMessageAckUnknown fires when the answering message named by Ack is
	// not carried by the same channel.
	RuleMessageAckUnknown = "K17-MESSAGE-ACK-UNKNOWN"
	// RuleMessageUncarried fires when a message is declared and no channel
	// lists it.
	RuleMessageUncarried = "K17-MESSAGE-UNCARRIED"
	// RuleChannelContractAndMessages fires when a channel states both the
	// short form and a protocol.
	RuleChannelContractAndMessages = "K17-CHANNEL-CONTRACT-AND-MESSAGES"
	// RuleMessageShapeChanged fires when a recorded message no longer carries
	// the structure it was recorded with.
	RuleMessageShapeChanged = "K17-MESSAGE-SHAPE-CHANGED"
	// RuleMessageFieldRemoved fires when a field of a recorded message is gone.
	RuleMessageFieldRemoved = "K17-MESSAGE-FIELD-REMOVED"
	// RuleMessageFieldChanged fires when a field kept its name and changed its
	// structure.
	RuleMessageFieldChanged = "K17-MESSAGE-FIELD-CHANGED"
)

// Messages holds the protocol of every channel that carries one.
//
// # Why a channel needs this at all
//
// Contract answers "what shape crosses here", which is the whole story for a
// boundary carrying one payload and almost none of it for a control channel.
// A dozen kinds of message travel in both directions, each with its own moment
// and its own answer, and folding them into one contract loses everything
// somebody needs in order to speak the protocol.
//
// # What is checked, and what deliberately is not
//
// The shape is not checked here, because it is not declared: it follows from
// the Go type, and the rules that hold a shape to its promise already exist.
// What is checked is the part a type cannot carry — that the direction is one
// the channel actually has, that the answer named is a message this channel
// carries, and that the question about redelivery has been answered at all.
func Messages(tree *reqtree.Tree, t ir.Topology, out *diag.Set) {
	channels := append([]ir.Channel(nil), t.Channels...)
	sort.Slice(channels, func(i, j int) bool { return channels[i].Pos.Less(channels[j].Pos) })

	carried := map[string]bool{}
	for _, c := range channels {
		checkChannelForm(c, out)
		payloads := map[string]bool{}
		for _, m := range c.Messages {
			carried[m.GoIdent] = true
			if m.PayloadType != "" {
				payloads[m.PayloadType] = true
			}
		}
		for _, m := range c.Messages {
			checkMessage(tree, c, m, payloads, out)
		}
	}
	checkUncarried(t, carried, out)
}

// checkChannelForm refuses a channel that states what crosses it twice.
func checkChannelForm(c ir.Channel, out *diag.Set) {
	if c.Contract == nil || len(c.MessageRefs) == 0 {
		return
	}
	out.Add(diag.Finding{
		Code: diag.Code(diag.PhaseSemantic, 200),
		Pos:  c.Pos,
		Rule: RuleChannelContractAndMessages,
		What: "channel " + c.Name() + " states both a contract and a set of messages.",
		Why:  "Both say what crosses here, so a reader has to work out which one is authoritative, and the two drift apart the first time one of them is extended. Contract is the short form for a boundary carrying exactly one shape; a channel carrying a protocol has messages.",
		How:  "Keep Messages and drop Contract, or the other way round if this boundary really carries one shape.",
	})
}

func checkMessage(tree *reqtree.Tree, c ir.Channel, m ir.Message, payloads map[string]bool, out *diag.Set) {
	name := messageName(m)

	var missing []string
	if m.Payload == nil {
		missing = append(missing, "Payload")
	}
	if m.Purpose == "" {
		missing = append(missing, "Purpose")
	}
	if m.Trigger == "" {
		missing = append(missing, "Trigger")
	}
	if len(missing) > 0 {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 201),
			Pos:  m.Pos,
			Rule: RuleMessageIncomplete,
			What: "message " + name + " leaves out " + strings.Join(missing, " and ") + ".",
			Why:  "What crosses, what it is for and when it is sent are what somebody needs in order to speak this protocol. The moment matters as much as the rest: silence on a channel is either correct or a fault, and only the trigger says which.",
			How:  "State the payload type, what the message is for, and the moment it is sent.",
		})
	}

	// The direction is stated per message because a control channel is used
	// both ways, but it has to be a direction this channel actually has.
	for _, end := range []string{m.From, m.To} {
		if end == "" || end == c.From || end == c.To {
			continue
		}
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 202),
			Pos:  m.Pos,
			Rule: RuleMessageEndUnknown,
			What: "message " + name + " names " + quote(end) + ", which is not an end of " + c.Name() + ".",
			Why:  "A message travels along the channel that carries it. An end that is not one of the channel's two is either a typo or a message on the wrong channel, and both read as a route that exists.",
			How:  "Use " + quote(c.From) + " or " + quote(c.To) + ", or list this message on the channel it really crosses.",
		})
	}

	if m.Repeatable == ir.Unanswered {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 203),
			Pos:  m.Pos,
			Rule: RuleMessageRepeatUnstated,
			What: "message " + name + " does not say whether it may be delivered twice.",
			Why:  "The two answers are opposite instructions to whoever implements the far end: one means look the identifier up and ignore what you have already done, the other means guard against duplicates. On a channel that can drop and reconnect, that is the difference between a safe retry and doing the work twice.",
			How:  "Set Repeatable to spec.Yes or spec.No.",
		})
	}

	if m.AckType != "" && !payloads[m.AckType] {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 204),
			Pos:  m.Pos,
			Rule: RuleMessageAckUnknown,
			What: "message " + name + " is answered by " + shortName(m.AckType) + ", which " + c.Name() + " does not carry.",
			Why:  "An answer that travels on no channel is one the sender waits for and never receives. A reference that resolves to nothing looks like a protocol and is not one.",
			How:  "List the answering message on this channel, or correct the type named by Ack.",
		})
	}

	for _, ref := range m.Satisfies {
		if tree.ByGoIdent(ref) == nil {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 201),
				Pos:  m.Pos,
				Rule: RuleMessageIncomplete,
				What: "message " + name + " names " + shortName(ref) + ", which is not a requirement.",
				Why:  "A reference that resolves to nothing looks like traceability and provides none.",
				How:  "Name a declared requirement.",
			})
		}
	}
	checkTopicRefs(tree, m.Topics, "message "+name, m.Pos, out)
}

// checkUncarried reports a message no channel lists.
//
// The same direction as an unused participant, and for the same reason: a
// shape written down and never sent reads as part of a protocol, and somebody
// on the far end will implement it.
func checkUncarried(t ir.Topology, carried map[string]bool, out *diag.Set) {
	declared := append([]ir.Message(nil), t.DeclaredMessages...)
	sort.Slice(declared, func(i, j int) bool { return declared[i].Pos.Less(declared[j].Pos) })

	for _, m := range declared {
		if carried[m.GoIdent] {
			continue
		}
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 205),
			Pos:  m.Pos,
			Rule: RuleMessageUncarried,
			What: "message " + messageName(m) + " is carried by no channel.",
			Why:  "A message nothing sends still reads as part of the protocol, and whoever implements the far end will write code for it. Either a channel is missing it or it is left over from a shape that was dropped.",
			How:  "List it on the channel that carries it, or remove it.",
		})
	}
}

// messageName is what a reader recognises: the payload type where there is
// one, and otherwise the declaration.
func messageName(m ir.Message) string {
	if m.PayloadType != "" {
		return shortName(m.PayloadType)
	}
	return shortName(m.GoIdent)
}

// MessageEvolution holds every recorded message to the shape it was recorded
// with.
//
// # Why a protocol needs this more than a single contract does
//
// Both ends of a control channel are deployed apart and upgraded apart. That
// is the entire difficulty: at any moment one of them is older than the other,
// still sending what it always sent and still expecting what it always
// expected. A field quietly dropped from a message is not found by either
// program's tests, because each is consistent with itself.
func MessageEvolution(channels []ir.Channel, base *baseline.File, out *diag.Set) {
	if base == nil {
		return
	}
	sorted := append([]ir.Channel(nil), channels...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name() < sorted[j].Name() })

	for _, ch := range sorted {
		rec, known := base.Channels[ch.Name()]
		if !known || len(rec.Messages) == 0 {
			continue
		}
		now := map[string]*baseline.Wire{}
		for _, m := range ch.Messages {
			if m.Payload != nil {
				now[m.PayloadType] = wireOf(m.Payload)
			}
		}

		for _, typ := range sortedKeys(rec.Messages) {
			ref := ch.Name() + ", message " + shortName(typ)
			current, still := now[typ]
			if !still {
				out.Add(diag.Finding{
					Code: diag.Code(diag.PhaseSemantic, 210),
					Pos:  ch.Pos,
					Rule: RuleMessageShapeChanged,
					What: ref + " was recorded and the channel no longer carries it.",
					Why:  "The far end was written against a protocol containing it and is still deployed. Removing a message from the declaration removes it from this side only; the other side goes on sending it, and now nothing compares what arrives.",
					How:  "List the message again, or record its withdrawal with freeze so the lock file is where somebody reviews it.",
				})
				continue
			}
			shapeEvolution(ref, ch.Pos, rec.Messages[typ], current, messageRules, out)
		}
	}
}

// RecordMessages writes the message shapes of every channel into the baseline.
func RecordMessages(channels []ir.Channel, base *baseline.File) int {
	changed := 0
	for _, ch := range channels {
		if len(ch.Messages) == 0 {
			continue
		}
		now := map[string]*baseline.Wire{}
		for _, m := range ch.Messages {
			if m.Payload != nil {
				now[m.PayloadType] = wireOf(m.Payload)
			}
		}
		if base.Channels == nil {
			base.Channels = map[string]baseline.Channel{}
		}
		rec := base.Channels[ch.Name()]
		if sameMessages(rec.Messages, now) {
			continue
		}
		rec.Messages = now
		base.Channels[ch.Name()] = rec
		changed++
	}
	return changed
}

func sameMessages(was, now map[string]*baseline.Wire) bool {
	if len(was) != len(now) {
		return false
	}
	for typ, w := range was {
		if !sameWire(w, now[typ]) {
			return false
		}
	}
	return true
}
