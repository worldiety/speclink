package ir

// Waivers records the exemptions declared through spec.Waive.
//
// It lives here rather than next to one family of checks because a waiver is a
// statement about a construct, not about the checker that happens to read it.
// Keeping it in one place is also the only way the promise made by spec.Waive —
// that it is the single escape hatch of the tool — can hold across every rule.
type Waivers map[waiverKey]string

// waiverKey identifies a waiver: a rule, optionally narrowed to one target.
type waiverKey struct {
	target string
	rule   string
}

// CollectWaivers gathers the waivers of every binding.
func CollectWaivers(bindings []Binding) Waivers {
	out := Waivers{}
	for _, b := range bindings {
		for _, a := range b.Assertions {
			if a.Kind != AssertWaive {
				continue
			}
			out[waiverKey{target: b.Target.String(), rule: a.Rule}] = a.Text
		}
	}
	return out
}

// Has reports whether the rule is waived for that target. An empty target asks
// for a waiver that was declared without narrowing to one construct.
func (w Waivers) Has(target, rule string) bool {
	_, ok := w[waiverKey{target: target, rule: rule}]
	return ok
}

// Reason returns the justification given for a waiver, empty when none applies.
func (w Waivers) Reason(target, rule string) string {
	return w[waiverKey{target: target, rule: rule}]
}
