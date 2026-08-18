// Package dec holds the cross cutting architectural decisions.
package dec

import "github.com/worldiety/speclink/spec"

var RDecNumbering = spec.Requirement{
	ID:         "R-DEC-NUMBERING",
	Kind:       spec.Decision,
	Discipline: spec.Technical,
	Status:     spec.Abstract,
	Title:      "Central number registry",
	Text:       "Sequential business numbers MUST be drawn from one central registry.",
	Rationale:  "Gapless numbering is a bookkeeping obligation and cannot be guaranteed if every context counts on its own.",
	Sources: []spec.Source{
		{Extern: "GoBD Rz. 36"},
	},
}
