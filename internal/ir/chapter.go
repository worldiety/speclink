package ir

// Chapter is a place in the document filled with written prose.
type Chapter struct {
	ID      string
	GoIdent string
	Doc     string
	At      Place
	Pos     Position
}

// Place names a point in the generated document. The values mirror spec.Place.
type Place int

const (
	PlaceBeginning Place = iota + 1
	PlaceBeforeArchitecture
	PlaceBeforeComposition
	PlaceBeforeBoundary
	PlaceBeforeSurface
	PlaceBeforeProcesses
	PlaceBeforeRequirements
	PlaceAppendix
)

// PlaceOf maps the name of a spec.Place constant onto the value.
//
// By name rather than by the integer, because the integer is an implementation
// detail of the spec package and reading it here would make the two orders
// have to agree forever without anything checking that they do.
func PlaceOf(ident string) (Place, bool) {
	switch ident {
	case "Beginning":
		return PlaceBeginning, true
	case "BeforeArchitecture":
		return PlaceBeforeArchitecture, true
	case "BeforeComposition":
		return PlaceBeforeComposition, true
	case "BeforeBoundary":
		return PlaceBeforeBoundary, true
	case "BeforeSurface":
		return PlaceBeforeSurface, true
	case "BeforeProcesses":
		return PlaceBeforeProcesses, true
	case "BeforeRequirements":
		return PlaceBeforeRequirements, true
	case "Appendix":
		return PlaceAppendix, true
	}
	return 0, false
}
