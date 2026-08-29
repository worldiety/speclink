// Package processes holds the courses of business.
//
// It sits above the contexts rather than inside one, because a process is
// precisely the thing no single context owns: it names use cases from several
// of them and is the promise that they add up. A context that declared it would
// have to import its neighbour, which the layering rules rightly forbid.
package processes

import (
	"example.com/erp/app/sales"
	"example.com/erp/requirements/fun/quote"
	"github.com/worldiety/speclink/spec"
)

// PQuoteDecision is the course a quote takes from drafting to its outcome.
//
// Three things in it are worth reading rather than skipping. The fork is a
// genuine wait: the clerk is shown the customer overview and the list of quotes
// before deciding, and the decision is not put until both are there. The
// rework branch goes back to the submission, which is why this is a graph and
// not a list of steps. And the two ends are separate on purpose — approved and
// withdrawn are different outcomes, and a process that modelled them as one
// endpoint would have thrown away the distinction anybody cares about.
var PQuoteDecision = spec.Process{
	ID:      "P-QUOTE-DECISION",
	Title:   "Angebot bis zur Entscheidung",
	Purpose: "Ein abgegebenes Angebot wird freigegeben, zurückgezogen oder zur Nachbesserung zurückgereicht.",
	Satisfies: []spec.Requirement{
		quote.RQuoteSubmit,
		quote.RQuoteApprove,
	},
	Nodes: []spec.Node{
		spec.Start("entwurf", "Angebot ist entworfen"),
		spec.Do[sales.SubmitQuote]("abgeben"),
		spec.Emit[sales.QuoteSubmitted]("abgegeben"),

		spec.Fork("aufteilen"),
		spec.Do[sales.FindQuoteOverview]("uebersicht"),
		spec.Do[sales.ListQuotes]("liste"),
		spec.Join("zusammen"),

		spec.Choice("pruefen"),
		spec.Do[sales.ApproveQuote]("freigeben"),
		spec.Do[sales.WithdrawQuote]("zurueckziehen"),
		spec.Emit[sales.QuoteWithdrawn]("zurueckgezogen"),

		spec.End("freigegeben", "freigegeben"),
		spec.End("verworfen", "zurückgezogen"),
	},
	Edges: []spec.Edge{
		{From: "entwurf", To: "abgeben"},
		{From: "abgeben", To: "abgegeben"},
		{From: "abgegeben", To: "aufteilen"},

		{From: "aufteilen", To: "uebersicht"},
		{From: "aufteilen", To: "liste"},
		{From: "uebersicht", To: "zusammen"},
		{From: "liste", To: "zusammen"},
		{From: "zusammen", To: "pruefen"},

		{From: "pruefen", To: "freigeben", When: "angenommen"},
		{From: "pruefen", To: "zurueckziehen", When: "abgelehnt"},
		{From: "pruefen", To: "abgeben", When: "nachzubessern"},

		{From: "freigeben", To: "freigegeben"},
		{From: "zurueckziehen", To: "zurueckgezogen"},
		{From: "zurueckgezogen", To: "verworfen"},
	},
}
