package reqtree

import (
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// RuleTopicDuplicate fires when two themes claim one ID.
const RuleTopicDuplicate = "K19-TOPIC-DUPLICATE"

// ResolveTopics indexes the declared themes and rewrites the references of
// every requirement from Go identifiers into topic IDs.
//
// The same second pass DerivedFrom goes through, and for the same reason: a
// theme may be declared after the requirements that use it, and the order of
// files must not decide whether a document has chapters.
func (t *Tree) ResolveTopics(topics []*ir.Topic, out *diag.Set) {
	t.topics = make(map[string]*ir.Topic, len(topics))
	t.byTopicIdent = make(map[string]*ir.Topic, len(topics))
	byIdent := t.byTopicIdent

	for _, top := range topics {
		if first, dup := t.topics[top.ID]; dup {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 130),
				Pos:  top.Pos,
				Rule: RuleTopicDuplicate,
				What: "topic " + strconvQuote(top.ID) + " is declared twice.",
				Why:  "The ID heads a chapter and appears in every diagnostic about it. Two declarations under one name make both ambiguous.",
				How:  "Rename one of them; the other is at " + first.Pos.String() + ".",
			})
			continue
		}
		t.topics[top.ID] = top
		byIdent[top.GoIdent] = top
	}

	for _, id := range t.sortedIDs() {
		r := t.ByID[id]
		resolved := make([]string, 0, len(r.Topics))
		for _, ref := range r.Topics {
			top, ok := byIdent[ref]
			if !ok {
				// Unreachable for a well formed build, exactly as the same
				// case is for DerivedFrom: naming a topic that does not exist
				// is a Go compile error long before speclink runs. It can only
				// happen when the declaring package was not part of the load,
				// which is a question about the patterns and not about the
				// requirement — so it is phrased as one.
				out.Add(diag.Finding{
					Code: diag.Code(diag.PhaseResolve, 13),
					Pos:  r.Pos,
					What: "Topics of " + r.ID + " references " + ref + ", which is not a known topic.",
					Why:  "A theme that resolves to nothing puts the requirement in no chapter while looking as though it were filed.",
					How:  "Include the declaring package in the analysed patterns, or remove the reference.",
				})
				continue
			}
			resolved = append(resolved, top.ID)
		}
		r.Topics = resolved
	}
}

// Topics returns the declared themes, ordered by ID.
func (t *Tree) Topics() []*ir.Topic {
	out := make([]*ir.Topic, 0, len(t.topics))
	for _, top := range t.topics {
		out = append(out, top)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Topic resolves one theme by ID.
func (t *Tree) Topic(id string) *ir.Topic { return t.topics[id] }

// TopicByGoIdent resolves a theme by the qualified identifier that names it.
//
// Declarations outside the requirement tree — a channel, a participant — name a
// theme the way Go names it, because that is what the compiler checks. They
// need the same lookup the requirements get, rather than a second index that
// could disagree with this one.
func (t *Tree) TopicByGoIdent(ident string) *ir.Topic { return t.byTopicIdent[ident] }

func strconvQuote(s string) string { return `"` + s + `"` }

// lastDotted renders a qualified Go identifier as the name a reader wrote.
func lastDotted(s string) string {
	if i := strings.LastIndexByte(s, '.'); i >= 0 {
		return s[i+1:]
	}
	return s
}

// RuleTopicUnused fires when nothing is filed under a declared theme.
const RuleTopicUnused = "K19-TOPIC-UNUSED"

// CheckTopicsUsed reports a theme nothing belongs to.
//
// It is the only direction worth checking here. There is no such thing as a
// theme nobody has covered — a theme is not an obligation — but an empty one is
// a chapter heading with nothing under it, which reads as a part of the system
// that was left out rather than as a heading somebody stopped using.
// alsoUsed are the Go identifiers naming a theme from outside the requirement
// tree — a channel or a participant. They count as use, because a theme that
// only groups the edge of the system is still a heading with something under
// it, and reporting it as empty would push people to file requirements under it
// that do not belong there.
func (t *Tree) CheckTopicsUsed(alsoUsed []string, out *diag.Set) {
	used := map[string]bool{}
	for _, r := range t.ByID {
		for _, id := range r.Topics {
			used[id] = true
		}
	}
	for _, ident := range alsoUsed {
		if top := t.byTopicIdent[ident]; top != nil {
			used[top.ID] = true
		}
	}

	for _, top := range t.Topics() {
		if used[top.ID] {
			continue
		}
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 131),
			Pos:  top.Pos,
			Rule: RuleTopicUnused,
			What: "nothing is filed under " + strconvQuote(top.ID) + ".",
			Why:  "A theme heads a chapter. An empty one reads as a part of the system that was left out, rather than as a heading somebody stopped using.",
			How:  "File the requirements, channels or participants that belong here, or remove the topic.",
		})
	}
}
