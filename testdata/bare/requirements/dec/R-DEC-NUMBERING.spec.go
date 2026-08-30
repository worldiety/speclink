package dec

import "github.com/worldiety/speclink/spec"

// RDecNumbering fixes where business numbers come from.
var RDecNumbering = spec.Requirement{
	ID:           "R-DEC-NUMBERING",
	Kind:         spec.Decision,
	Discipline:   spec.Technical,
	Status:       spec.Normative,
	Title:        "Central number registry",
	Text:         "Business numbers MUST be drawn from one central, gapless registry.",
	Rationale:    "A number drawn twice cannot be repaired after the fact.",
	Consequences: "The registry is a single writer, so it is the one part of the system that cannot be scaled by running a second copy, and every context that needs a number waits on it.",
	Sources:      []spec.Source{{Extern: "GoBD Rz. 36"}},
}
