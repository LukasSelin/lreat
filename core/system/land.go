package system

import (
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
	// ErodeEvery is how many ticks an age of weather covers. The land is
	// worked over as a whole rather than a little each tick, because the
	// drainage has to be recomputed for the whole map at once for the rivers
	// to be anywhere sensible, and because a hillside does not lose its soil
	// evenly - it loses it in the winters.
	ErodeEvery = 120
)

func isForest(t *world.Tile) bool { return t.Terrain == world.Forest }

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
		switch t.Terrain {
		case world.Forest:
			t.Wood = min(1, t.Wood+regrowth*k)
			t.Wild = min(1, t.Wild+wildRegrowth*k)
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
		// Seed falls in the growing season, not on frozen ground.
		if w.RNG.Float64() < reseedChance*w.Climate.Growth() {
			t.Terrain, t.Wood, t.Wild = world.Forest, 0.2, 0.3
		}
	}
}
