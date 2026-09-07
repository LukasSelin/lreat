package system

import (
	"lreat/core/clock"
	"lreat/core/entity"
	"lreat/core/world"
)

const (
	// regrowth is how much standing timber a forest tile recovers per tick.
	regrowth = 0.0004
	// wildRegrowth is how much wild food a forest tile recovers per tick,
	// fishRegrowth how much a water tile does, and fallow how much worn
	// fertility a field recovers per tick toward what the land can hold.
	wildRegrowth = 0.0012
	fishRegrowth = 0.0012
	fallow       = 0.0006
	// reseedSamples is how many random tiles per tick are checked for
	// spontaneous reforestation from a wooded neighbor.
	reseedSamples = 3
	reseedChance  = 0.3
	// ErodeEvery is how long an age of weather covers. The land is worked
	// over as a whole rather than a little each day, because the drainage
	// has to be recomputed for the whole map at once for the rivers to be
	// anywhere sensible, and because a hillside does not lose its soil
	// evenly - it loses it in the winters. Once a season, then.
	ErodeEvery = clock.Season
)

func isForest(t *world.Tile) bool { return t.Terrain == world.Forest }

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

// Land lets forests regrow and slowly reclaim unclaimed grass beside them,
// weathers the ground every so often so that the hills wear into the valleys
// and the rivers go where the new heights send them,
// lets wild food and fish come back, and lets a worn field recover toward
// what the land can hold. Clearing land is therefore reversible only where
// nobody has settled, so the built footprint of a settlement persists while
// the wild edges shift; and taking too much from any one place leaves it
// poorer for a long while, which is what makes a settlement move on to
// other ways of living.
func Land(w *world.World) {
	g := w.Grid
	g.Weather()
	if w.Tick%ErodeEvery == 0 {
		w.Erode()
	}
	// Green things keep the season's hours. In the cold half of the year
	// nothing regrows, so what a summer left standing is what a winter has
	// to live on; over a whole year the growth is what it was before the
	// seasons existed. Fish and fallow follow the same clock: the water
	// under ice gives back nothing, and worn ground rests until it thaws.
	k := w.Mods.Regrowth * w.Climate.Growth()
	for i := range g.Tiles {
		t := &g.Tiles[i]
		// A stand ages by the weather it gets, not by the calendar: what a
		// winter gives it is nothing, and that is the same clock everything
		// else growing keeps.
		if t.Alive() {
			t.Age += k
		}
		switch t.Terrain {
		case world.Forest:
			// What a wood holds is bounded by how long it has stood. The
			// timber comes on over a lifetime and the brush under it within
			// a few years, so a young stand feeds a forager long before it
			// is worth felling.
			t.Wood = grown(t.Wood, t.Grown(world.TimberAge), regrowth*k)
			t.Wild = grown(t.Wild, t.Grown(world.BrushAge), wildRegrowth*k)
		case world.Water:
			t.Fish = min(1, t.Fish+fishRegrowth*k)
		case world.Field:
			t.Fertility = min(t.Rich, t.Fertility+fallow*k)
		}
	}
	for k := 0; k < reseedSamples; k++ {
		p := entity.Pos{X: w.RNG.IntN(g.W), Y: w.RNG.IntN(g.H)}
		t := g.At(p)
		if !t.Buildable() || !g.HasNeighbor(p, isForest) {
			continue
		}
		// Seed falls everywhere and takes where the ground will hold it. A
		// wood that spread wherever a seed landed closed over the whole map
		// and left nowhere to break a field; a wood that stops at the tree
		// line comes back over what was cleared and no further.
		if !g.HoldsWood(p) {
			continue
		}
		// Seed falls in the growing season, not on frozen ground.
		if w.RNG.Float64() < reseedChance*w.Climate.Growth() {
			t.Terrain, t.Wood, t.Wild = world.Forest, 0, 0
			t.Sow() // a seedling wood, with nothing on it yet
		}
	}
}
