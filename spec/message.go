package spec

// Message is one kind of thing that crosses a channel.
//
// # Why a channel needs more than one
//
// A channel says what crosses it in the words of a data protection register,
// and names the one shape it carries. That is enough for a boundary that
// carries a single payload — a file store, a payment gateway — and it is not a
// protocol. A control channel carries a dozen kinds of message in both
// directions, each with its own trigger and its own answer, and describing that
// as one contract loses everything a reader needs in order to speak it.
//
// # What is declared here, and what is not
//
// The shape is not declared. It follows from Payload, which is a Go type, and
// so do the field names, their nesting, which of them may be omitted and what
// a generated schema says. Writing the fields down again here would be the
// same fact in two places, and the copy is the one that rots.
//
// What remains is what no type can carry: which way it goes, what it is for,
// when it is sent, whether it may be sent twice, and what answers it.
type Message struct {
	// Payload is the Go type whose shape crosses, given as a zero value:
	//
	//	Payload: proto.Hello{}
	Payload any

	// From and To name the two ends, and must be the ends of the channel that
	// carries this message. A control channel is used in both directions, so
	// the direction is stated per message rather than taken from the channel.
	From, To string

	// Purpose says what this message is for, in one or two sentences.
	Purpose string
	// Trigger says when it is sent. A message whose moment is not written down
	// is one whose absence nobody can judge: silence on a channel is either
	// correct or a fault, and only the trigger says which.
	Trigger string

	// Repeatable says whether the far end may receive this message twice
	// without a second effect.
	//
	// # Why it is not a bool
	//
	// Because the zero value of a bool would make "nobody said" and "no" the
	// same answer, and they are opposite instructions to whoever implements
	// the far end. One means look it up; the other means guard against
	// duplicates. A channel that can drop and reconnect makes this the
	// difference between a safe retry and a second charge.
	//
	// It is also a promise, so it is recorded and a change to it is reported:
	// turning a repeatable message into one that is not breaks every sender
	// that was relying on being able to send it again.
	Repeatable Answer

	// Ack names the payload of the message that answers this one, where one
	// does. Left out where the message is not answered.
	Ack any

	// Satisfies names the requirements this message answers to.
	Satisfies []Requirement
	// Topics are the themes it belongs to. Optional, as everywhere else.
	Topics []Topic
}

// Answer is a three valued yes or no whose zero value is neither.
//
// It exists for the facts where not having been asked is a third state that
// matters. A bool would collapse it into the negative, which is the direction
// this kind of question must never fail in: an unanswered question read as a
// "no" is a decision nobody made, presented as one somebody did.
type Answer int

const (
	// Unanswered is the zero value: nobody has stated this.
	Unanswered Answer = iota
	// Yes states it affirmatively.
	Yes
	// No states it negatively, which is a statement and not an absence.
	No
)

func (a Answer) String() string {
	switch a {
	case Yes:
		return "yes"
	case No:
		return "no"
	}
	return "unanswered"
}
