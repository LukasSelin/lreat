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
	// abandonChance is the per-tick chance that a strip of field nobody is
	// left to work goes back to grass. A holding is eight times the ground a
	// house stands on, so what the dead hold on to matters eight times as
	// much: without this the map fills with the fields of people who died a
	// thousand ticks ago and the living have nowhere to plough.
	abandonChance = 0.002
	// reseedSamples is how many random tiles per tick are checked for
	// spontaneous reforestation from a wooded neighbor.
	reseedSamples = 3
	reseedChance  = 0.3
)

func isForest(t *world.Tile) bool { return t.Terrain == world.Forest }

// Land lets forests regrow and slowly reclaim unclaimed grass beside them,
// lets wild food and fish come back, and lets a worn field recover toward
// what the land can hold. Clearing land is therefore reversible only where
// nobody has settled, so the built footprint of a settlement persists while
// the wild edges shift; and taking too much from any one place leaves it
// poorer for a long while, which is what makes a settlement move on to
// other ways of living.
// working is the set of agents still alive to hold land.
func working(w *world.World) map[entity.ID]bool {
	held := make(map[entity.ID]bool, len(w.Agents))
	for _, a := range w.Agents {
		held[a.ID] = true
	}
	return held
}

func Land(w *world.World) {
	g := w.Grid
	g.Weather()
	k := w.Mods.Regrowth
	held := working(w)
	for i := range g.Tiles {
		t := &g.Tiles[i]
		switch t.Terrain {
		case world.Forest:
			t.Wood = min(1, t.Wood+regrowth*k)
			t.Wild = min(1, t.Wild+wildRegrowth*k)
		case world.Water:
			t.Fish = min(1, t.Fish+fishRegrowth*k)
		case world.Field:
			t.Fertility = min(t.Rich, t.Fertility+fallow)
			// A field is only a field while somebody works it. Left alone,
			// the furrows close over and the ground is common again.
			if !held[t.Owner] && w.RNG.Float64() < abandonChance {
				t.Terrain, t.Owner = world.Grass, 0
			}
		}
	}
	for k := 0; k < reseedSamples; k++ {
		p := entity.Pos{X: w.RNG.IntN(g.W), Y: w.RNG.IntN(g.H)}
		t := g.At(p)
		if !t.Buildable() || !g.HasNeighbor(p, isForest) {
			continue
		}
		if w.RNG.Float64() < reseedChance {
			t.Terrain, t.Wood, t.Wild = world.Forest, 0.2, 0.3
		}
	}
}
