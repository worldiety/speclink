package check

import "testing"

// The How line of a lifecycle finding carries a state rather than an ellipsis,
// and the state is guessed from the event's own name. A guess in a diagnostic
// costs a word of editing when it is wrong, which is cheaper than a placeholder
// the reader has to decode first — but only if it is usually right.
func TestSuggestedStateIsThePastParticiple(t *testing.T) {
	for _, tc := range []struct{ event, want string }{
		{"QuoteSubmitted", "submitted"},
		{"QuoteWithdrawn", "withdrawn"},
		{"InvoiceSettled", "settled"},
		// A run of capitals is one word, so the participle is still the last
		// one rather than a single letter.
		{"QuoteIDAssigned", "assigned"},
		// Nothing sensible to say about a single word; the ellipsis is honest.
		{"Submitted", "submitted"},
		{"", "…"},
	} {
		if got := suggestState(tc.event); got != tc.want {
			t.Errorf("suggestState(%q) = %q, want %q", tc.event, got, tc.want)
		}
	}
}

func TestSplitWordsKeepsAcronymsTogether(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want []string
	}{
		{"QuoteSubmitted", []string{"Quote", "Submitted"}},
		{"QuoteIDAssigned", []string{"Quote", "ID", "Assigned"}},
		{"ID", []string{"ID"}},
		{"quote", []string{"quote"}},
	} {
		got := splitWords(tc.in)
		if len(got) != len(tc.want) {
			t.Errorf("splitWords(%q) = %v, want %v", tc.in, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("splitWords(%q) = %v, want %v", tc.in, got, tc.want)
				break
			}
		}
	}
}
