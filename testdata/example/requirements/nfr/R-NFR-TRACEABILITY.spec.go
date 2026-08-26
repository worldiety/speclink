// Package nfr holds the cross cutting quality goals.
package nfr

import "github.com/worldiety/speclink/spec"

// RNfrTraceability is the requirement the envelope fields exist for.
//
// Every stored message carries fields that answer no business question on their
// own — which thing it is about, who caused it, when. They are not thereby self
// evident: they are the audit trail, and they exist because somebody has to be
// able to reconstruct afterwards what happened. Naming that requirement at each
// of them is what keeps the answer available when it is needed.
var RNfrTraceability = spec.Requirement{
	ID:         "R-NFR-TRACEABILITY",
	Kind:       spec.NonFunctional,
	Discipline: spec.Mixed,
	Status:     spec.Normative,
	Title:      "Stored facts are attributable",
	Text:       "Every stored fact MUST identify the thing it is about, so that the sequence of events on any object can be reconstructed.",
	Sources: []spec.Source{
		{Extern: "GoBD Rz. 36"},
	},
}
