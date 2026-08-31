package main

import (
	"strings"
	"testing"
)

// protocolFixture turns the file store channel of a copy into one that carries
// a protocol, so the machinery can be exercised on a real build.
//
// Written here rather than into the fixture itself: that boundary carries one
// shape, which is exactly what Contract is for, and a demonstration
// contradicting its own fixture teaches the wrong thing about when to reach
// for this.
func protocolFixture(t *testing.T, messages string) string {
	t.Helper()
	dir := copyFixture(t, "../../testdata/bare")

	appendTo(t, dir, "app/sales/adapter/fs/quotes.go", `
// StoreAck confirms that a quote was written.
type StoreAck struct {
	ID    string `+"`json:\"id\"`"+`
	Wrote bool   `+"`json:\"wrote\"`"+`
}
`)
	rewrite(t, dir, "topology/boundary.topology.go",
		"\tContract: fs.QuoteStore{},", "\tMessages: []spec.Message{Ablegen, Bestaetigung},")
	appendTo(t, dir, "topology/boundary.topology.go", messages)
	return dir
}

const soundProtocol = `
var Ablegen = spec.Message{
	Payload:    fs.QuoteStore{},
	From:       "app/sales/adapter/fs",
	To:         "dateiablage",
	Purpose:    "Legt den aktuellen Stand eines Angebots ab.",
	Trigger:    "Bei jeder Abgabe eines Angebots.",
	Repeatable: spec.Yes,
	Ack:        fs.StoreAck{},
	Satisfies:  []spec.Requirement{dec.RDecQuoteState},
}

var Bestaetigung = spec.Message{
	Payload:    fs.StoreAck{},
	From:       "dateiablage",
	To:         "app/sales/adapter/fs",
	Purpose:    "Bestätigt die Ablage.",
	Trigger:    "Als Antwort auf jede Ablage.",
	Repeatable: spec.Yes,
	Satisfies:  []spec.Requirement{dec.RDecQuoteState},
}
`

// TestAProtocolIsHeldToItsShape is why messages are recorded at all.
//
// Both ends of a control channel are deployed apart and upgraded apart. At any
// moment one of them is older, still sending what it always sent. A field
// quietly dropped is found by neither program's tests, because each is
// consistent with itself; what breaks is the pair, once the two versions meet.
func TestAProtocolIsHeldToItsShape(t *testing.T) {
	t.Parallel()
	dir := protocolFixture(t, soundProtocol)

	if out, code := runSpeclink(t, "freeze", dir); code != 0 {
		t.Fatalf("freeze failed with %d:\n%s", code, out)
	}
	if out, code := runVerify(t, dir); code != 0 {
		t.Fatalf("the protocol must be clean once recorded:\n%s", summary(out))
	}

	rewrite(t, dir, "app/sales/adapter/fs/quotes.go", "\tWrote bool   `json:\"wrote\"`\n", "")

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a field dropped from a recorded message must be reported:\n%s", summary(out))
	}
	if !strings.Contains(out, "SPEC-V6-211") {
		t.Errorf("expected the removal to be reported, got:\n%s", summary(out))
	}
	// The words matter: this is not a far end that never heard of us, it is
	// the other half of our own protocol, still deployed.
	if !strings.Contains(out, "still deployed and still sends it") {
		t.Errorf("the finding uses the wording of a one sided contract:\n%s", summary(out))
	}
}

// TestMovingFromAContractToAProtocolCanBeRecorded covers a gap in freeze.
//
// The finding for a dropped contract says to record the removal with freeze.
// It could not: freeze skipped every channel that stated no contract, so a
// recorded one was never cleared and the finding never went away.
func TestMovingFromAContractToAProtocolCanBeRecorded(t *testing.T) {
	t.Parallel()
	dir := protocolFixture(t, soundProtocol)

	out, code := runVerify(t, dir)
	if code == 0 || !strings.Contains(out, "SPEC-V6-170") {
		t.Fatalf("expected the dropped contract to be reported first:\n%s", summary(out))
	}
	if out, code := runSpeclink(t, "freeze", dir); code != 0 {
		t.Fatalf("freeze failed with %d:\n%s", code, out)
	}
	if out, code := runVerify(t, dir); code != 0 {
		t.Errorf("freeze did not clear the recorded contract:\n%s", summary(out))
	}
}

// TestAProtocolMustAnswerForItself covers the questions a type cannot answer.
func TestAProtocolMustAnswerForItself(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		edit func(string) string
		want string
	}{
		{
			// Both answers are opposite instructions to the far end: look the
			// identifier up, or guard against duplicates.
			name: "redelivery unstated",
			edit: func(s string) string { return strings.Replace(s, "\tRepeatable: spec.Yes,\n", "", 1) },
			want: "SPEC-V6-203",
		},
		{
			// A message travels along the channel carrying it. An end that is
			// not one of its two reads as a route that exists.
			name: "end not on this channel",
			edit: func(s string) string {
				return strings.Replace(s, `To:         "dateiablage",`, `To:         "objektspeicher",`, 1)
			},
			want: "SPEC-V6-202",
		},
		{
			// An answer travelling on no channel is one the sender waits for
			// and never receives.
			name: "answer not carried",
			edit: func(s string) string {
				return strings.Replace(s, "Ack:        fs.StoreAck{},", "Ack:        fs.Quotes{},", 1)
			},
			want: "SPEC-V6-204",
		},
		{
			// Silence on a channel is either correct or a fault, and only the
			// trigger says which.
			name: "no trigger",
			edit: func(s string) string {
				return strings.Replace(s, "\tTrigger:    \"Bei jeder Abgabe eines Angebots.\",\n", "", 1)
			},
			want: "SPEC-V6-201",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := protocolFixture(t, tc.edit(soundProtocol))

			out, code := runVerify(t, dir)
			if code == 0 {
				t.Fatalf("expected %s, but the run was clean:\n%s", tc.want, summary(out))
			}
			if !strings.Contains(out, tc.want) {
				t.Errorf("expected %s, got:\n%s", tc.want, summary(out))
			}
		})
	}
}

// TestAChannelStatesWhatCrossesOnce refuses the two spellings together.
func TestAChannelStatesWhatCrossesOnce(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")
	appendTo(t, dir, "app/sales/adapter/fs/quotes.go", `
// StoreAck confirms that a quote was written.
type StoreAck struct {
	ID string `+"`json:\"id\"`"+`
}
`)
	rewrite(t, dir, "topology/boundary.topology.go",
		"\tContract: fs.QuoteStore{},",
		"\tContract: fs.QuoteStore{},\n\tMessages: []spec.Message{Ablegen, Bestaetigung},")
	appendTo(t, dir, "topology/boundary.topology.go", soundProtocol)

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("stating what crosses twice must be reported:\n%s", summary(out))
	}
	if !strings.Contains(out, "SPEC-V6-200") {
		t.Errorf("expected the double statement to be reported, got:\n%s", summary(out))
	}
}

// TestTheProtocolReachesTheDocument is what the whole chapter is for.
func TestTheProtocolReachesTheDocument(t *testing.T) {
	t.Parallel()
	dir := protocolFixture(t, soundProtocol)

	out, code := runSpeclink(t, "generate", dir)
	if code != 0 {
		t.Fatalf("generate failed with %d:\n%s", code, out)
	}
	for _, want := range []string{
		"What is spoken on each channel",
		"Als Antwort auf jede Ablage.",
		// A message that nothing answers is a fact a caller needs before it
		// sits waiting, so it is stated rather than left blank.
		"n/a",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the document is missing %q:\n%s", want, out)
		}
	}
}
