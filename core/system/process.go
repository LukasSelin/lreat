package system

import (
	"lreat/core/ontology"
	"lreat/core/world"
)

// Running the world's half of the ontology.
//
// What happens on its own is stated in core/ontology as processes and
// transforms; this is what carries them out. It is three small runners
// rather than one phase on purpose. Growing has to happen before anybody
// decides what to do about it, spoiling after the market has been traded in,
// and ruin after the dead have been counted - and the order the world's
// chance is drawn in is the order every settlement's history depends on, so
// the runners go where the code they replace already stood rather than
// gathering into a phase of their own. See system.Step, which is unchanged.

// grown puts back what a tick of growing weather puts back, up to what the
// stand's age accounts for. Age bounds what a stand grows into, and nothing
// else: it never takes away what is already standing, so a wood is only ever
// held back from filling out, never thinned by the calendar.
func grown(have, ceiling, by float64) float64 {
	if have >= ceiling {
		return have
	}
	return min(ceiling, have+by)
}

// ripen advances what is growing on one tile by k of growing weather: the
// stand gets that much older, and whatever it is coming on toward fills a
// little further, bounded by the age it has had. A wood does both of these
// twice over, on two clocks - the brush under it within a few years, the
// timber over a lifetime - which is why the age is the tile's and the
// filling is the process's.
//
// A process whose yield the ground keeps no count of only ages the tile. A
// field is the case: a crop is cut once and wholly rather than drawn down,
// so what a strip has to give is read off how far along it is and there is
// no stock to put back.
func ripen(t *world.Tile, k float64) {
	ps := ontology.Growing(world.ClassOf(t))
	if len(ps) == 0 {
		return
	}
	// A stand ages by the weather it gets, not by the calendar: what a
	// winter gives it is nothing, and that is the same clock everything
	// else growing keeps.
	t.Age += k
	for _, p := range ps {
		if p.Rate == 0 {
			continue
		}
		if s, ok := world.Stock(t, p.Yields); ok {
			*s = grown(*s, t.Along(p), p.Rate*k)
		}
	}
}
