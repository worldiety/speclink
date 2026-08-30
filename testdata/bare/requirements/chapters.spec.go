package requirements

import "github.com/worldiety/speclink/spec"

// Warum opens the document with the reason the application exists, which is
// the one thing no model in it can state.
var Warum = spec.Chapter{
	ID:  "warum",
	Doc: "doc/warum.md",
	At:  spec.Beginning,
}
