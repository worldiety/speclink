package sales

import (
	"go.wdy.de/nago/application/evs"
)

// Stats is a read model that is bound to no requirement, so K1-CONSTRUCT-UNBOUND
// must fire. A projection is aggregate crossing, so nothing covers it
// transitively: if it is not stated why it exists, nobody can tell whether it
// should.
type Stats struct {
	Count int
}

// Clone returns a deep copy, as every projection state must.
func (s *Stats) Clone() *Stats {
	if s == nil {
		return nil
	}
	c := *s
	return &c
}

// Counted is the fact the read model folds.
type Counted struct{}

// Discriminator is the stable serialisation tag.
func (Counted) Discriminator() evs.Discriminator { return "bad.sales.counted.v1" }

func newStats(src evs.Source) *evs.Singleton[*Stats] {
	p := evs.NewSingleton[*Stats](src, evs.ProjectionOptions{})
	evs.Project(p,
		func(Counted) evs.Unit { return evs.TheUnit() },
		func(s *Stats, e Counted) { s.Count++ },
	)
	return p
}
