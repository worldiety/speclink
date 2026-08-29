package mem

import "github.com/worldiety/speclink/spec"

// The one adapter that crosses no boundary.
//
// Structurally this is an adapter like any other — it sits under adapter/ and
// implements a port — and the rule that every such place be described by a
// channel cannot see the difference. Here there is genuinely nothing on the far
// side: the map lives in this process and dies with it, so there is no
// protocol, no counterpart and nobody else's responsibility to record.
//
// That is a statement worth making rather than a rule worth weakening. It also
// stops being true the moment somebody puts a database behind this, and then
// the waiver has to be deleted and the channel written.
var _ = spec.ForPackage(
	spec.Waive("K17-ADAPTER-NO-CHANNEL", "The store is a map in this process. Nothing outlives the run, nothing crosses a boundary, and there is no counterpart to name — while the invoicing context is still a draft."),
)
