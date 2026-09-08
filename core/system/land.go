package system

import (
	"lreat/core/clock"
	"lreat/core/entity"
	"lreat/core/ontology"
	"lreat/core/world"
)

const (
	// How much a water tile's fish come back per tick, and how much worn
	// fertility a field recovers per tick toward what the land can hold.
	// Neither of these is a process: a shoal is a stock that replenishes,
	// not a crop that has to come on, and worn soil is resting rather than
	// growing. What a wood and a field have standing on them is in
	// ontology.Processes; see ripen.
	fishRegrowth = 0.0012
	fallow       = 0.0006
	// reseedSamples is how many random tiles per tick are checked for
	// spontaneous reforestation from a wooded neighbor.
	reseedSamples = 3
	reseedChance  = 0.3
	// ErodeEvery is how long an age of weather covers. An age is a decade,
	// not a season: the land is not weather, and a hillside that moved every
	// year would be a different thing altogether - a settlement would watch
	// its own river wander in the time it took to raise a barn. What the
	// ground does, it does over decades and centuries, so a run of a few
	// generations sees the slopes it cleared begin to go and a run of
	// several sees the valley they went into.
	//
	// It is also why the land is worked over as a whole rather than a little
	// each day: the drainage has to be recomputed for the whole map at once
	// for the rivers to be anywhere sensible, and doing that every day would
	// buy nothing but a slower run.
	ErodeEvery = 10 * clock.Year
)

func isForest(t *world.Tile) bool { return t.Is(ontology.Wood) }

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
		ripen(t, k)
		switch t.Terrain {
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
