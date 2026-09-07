package system

import (
	"lreat/core/entity"
	"lreat/core/world"
)

const (
	// regrowth is how much standing timber a forest tile recovers per tick.
	regrowth = 0.0004
	// reseedSamples is how many random tiles per tick are checked for
	// spontaneous reforestation from a wooded neighbor.
	reseedSamples = 3
	reseedChance  = 0.3
)

func isForest(t *world.Tile) bool { return t.Terrain == world.Forest }

// Land lets forests regrow and slowly reclaim unclaimed grass beside them.
// Clearing land is therefore reversible only where nobody has settled, so
// the built footprint of a settlement persists while the wild edges shift.
func Land(w *world.World) {
	g := w.Grid
	g.Weather()
	for i := range g.Tiles {
		t := &g.Tiles[i]
		if t.Terrain == world.Forest && t.Wood < 1 {
			t.Wood += regrowth
			if t.Wood > 1 {
				t.Wood = 1
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
			t.Terrain, t.Wood = world.Forest, 0.2
		}
	}
}
