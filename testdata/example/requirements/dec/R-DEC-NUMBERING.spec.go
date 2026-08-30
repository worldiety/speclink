// Package dec holds the cross cutting architectural decisions.
package dec

import "github.com/worldiety/speclink/spec"

var RDecNumbering = spec.Requirement{
	ID:           "R-DEC-NUMBERING",
	Kind:         spec.Decision,
	Discipline:   spec.Technical,
	Status:       spec.Abstract,
	Title:        "Central number registry",
	Text:         "Sequential business numbers MUST be drawn from one central registry.",
	Rationale:    "Gapless numbering is a bookkeeping obligation and cannot be guaranteed if every context counts on its own.",
	Consequences: "The registry is a single writer and a shared dependency of every context that issues a document, so it is both the scaling limit and the outage that stops all of them at once.",
	Sources: []spec.Source{
		{Extern: "GoBD Rz. 36"},
	},
}
