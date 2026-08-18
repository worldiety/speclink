package person

import (
	uient "go.wdy.de/nago/application/ent/ui"
)

// Pages reaches for the generic CRUD user interface. Its look and feel is the
// reference for a nago screen, but importing it is forbidden by
// K4-NO-GENERIC-CRUD: routes, permissions and read models are assembled at run
// time, so no view can be traced back to a requirement.
var Pages uient.Pages
