// Package requirements holds the themes the requirement tree is grouped under.
//
// A theme orders from inside; a standard imposes from outside. Both head a
// chapter, and keeping them apart is what lets the one be asked "which of these
// does nothing answer" while the other is simply how the people who built this
// think about it.
package requirements

import "github.com/worldiety/speclink/spec"

// Zugriff gathers everything about who may do what.
var Zugriff = spec.Topic{
	ID:          "T-ZUGRIFF",
	Title:       "Zugriff und Berechtigung",
	Description: "Wer eine fachliche Handlung auslösen darf und wo das geprüft wird. Nicht hier: wie ein Handelnder erkannt wird — das entscheidet die Präsentation.",
}

// Ablage gathers the rulings about how data is kept.
var Ablage = spec.Topic{
	ID:          "T-ABLAGE",
	Title:       "Ablage der Geschäftsdaten",
	Description: "Wie Geschäftsdaten abgelegt werden und was daraus über ihre Lesbarkeit folgt. Nicht hier: welche Felder es gibt — das steht an der Anforderung selbst.",
}
