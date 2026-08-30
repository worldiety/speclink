// Package topology names the world around this system and every way across it.
//
// None of it is inferable. No Go module states that a sales clerk exists, that
// the file store is a place outside this program's control, or that the channel
// to it carries quotation data with no protection at all because it never
// leaves the machine. That is not missing from the code — it is knowledge the
// code cannot hold, so it is written down and then held against what the code
// does show.
package topology

import (
	"example.com/bare/app/sales/adapter/fs"
	"example.com/bare/requirements/dec"
	"example.com/bare/requirements/fun/quote"
	"github.com/worldiety/speclink/spec"
)

// Vertrieb is the only person this skeleton knows.
var Vertrieb = spec.Actor{
	ID:   "vertrieb",
	Name: "Vertrieb",
	Role: "Legt Angebote an, gibt sie ab und schlägt sie nach. Handelt unter einer Kennung mit Berechtigungen.",
}

// Dateiablage is outside the program even though it is on the same machine.
//
// The distinction that matters is not distance but responsibility: what is
// written there outlives the process, is readable by anything with the same
// rights, and is not this program's to guarantee.
var Dateiablage = spec.Foreign{
	ID:   "dateiablage",
	Name: "Dateiablage",
	Role: "Verzeichnis im Dateisystem. Hält die abgelegten Angebote; Zugriffsschutz und Sicherung liegen beim Betrieb, nicht bei dieser Anwendung.",
}

// Selbstbedienung is how a person reaches the system.
var Selbstbedienung = spec.Channel{
	From:     "vertrieb",
	To:       "app/sales/rest",
	Label:    "Selbstbedienung",
	Protocol: "HTTP",
	Data:     "Angebotsdaten, Kennung des Handelnden",
	Auth:     "Kennung mit Berechtigungen, geprüft im Use Case",
	Crypto:   "entfällt, der Platzhalter terminiert kein TLS",
	Satisfies: []spec.Requirement{
		quote.RQuoteSubmit,
		quote.RQuoteLookup,
	},
}

// Angebotsablage is the one place this system writes something that outlives it.
var Angebotsablage = spec.Channel{
	From:     "app/sales/adapter/fs",
	To:       "dateiablage",
	Label:    "Angebotsablage",
	Protocol: "Dateisystem",
	Data:     "Abgelegte Angebote als aktueller Stand",
	Auth:     "Rechte des Prozessbenutzers",
	Crypto:   "entfällt, lokal",
	Satisfies: []spec.Requirement{
		dec.RDecQuoteState,
	},
	// The shape that outlives the process. Naming it here is what lets a run
	// notice that the thing on disk stopped being what this system reads.
	Contract: fs.QuoteStore{},
}
