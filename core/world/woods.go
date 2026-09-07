package world

import "lreat/core/entity"

// woodsShare is how much of a map's dry land will hold a wood. It is the
// answer to a settlement watching the forest close over it: with nothing to
// say where trees can stand, every tile a seed fell on became one, and after
// six thousand ticks the woods covered three quarters of the map and the
// fields had gone from forty strips to one, because a field can only be
// broken on open ground and there was none left.
//
// A tree line is what the land actually has. Woods stand on the damp gentle
// ground and stop where the ground turns dry, steep or high, and that line
// is read off the map the same way the founding woods were placed: the
// wettest, gentlest fifth of the land will hold a wood, and nothing else
// will. The founding woods take a little over half of that, so a settlement
// still watches the forest creep back over what it cleared - just not over
// the ground it farms.
const woodsShare = 0.2

// woodsSteep is how much of a map is too steep to hold a wood at all. Slope
// enters the reading below as a discount - steep ground suits trees less than
// flat ground of the same dampness - and a discount is not a limit: a damp
// enough bank came out above the tree line however sharply it fell away, so
// woods crept up the sides of the gullies the water had just cut. A tree
// needs ground to stand its roots in, and ground that steep is on its way
// downhill. The steepest fifth of the map holds nothing, whatever else is
// true of it.
const woodsSteep = 0.2

// WoodsAt is how well a tile's ground suits trees: damp enough to grow them
// and gentle enough to hold the soil they grow in. It is the reading the map
// was made with, and it moves when the weather moves the ground under it.
func (g *Grid) WoodsAt(p entity.Pos) float64 {
	if !g.In(p) || g.TooSteep(p) {
		return 0
	}
	t := g.At(p)
	damp := clamp01(1 - t.Drain/(2*FloodDepth))
	steep := clamp01(g.Slope(p) / max(1e-12, g.steepAt))
	return damp * (1 - 0.6*steep)
}

// TooSteep reports whether the ground falls away too fast for a wood to hold
// on it. It is a comparison with the rest of the map, like everything else
// read off the land: what counts as a bank on a map of hills is a cliff on a
// map of water meadows.
func (g *Grid) TooSteep(p entity.Pos) bool {
	if !g.In(p) {
		return true
	}
	if !g.woodsRead {
		g.readWoods()
	}
	// Steeper than the line, not at it: on ground with no slope in it at all
	// the line is zero, and flat ground is the last thing that should read as
	// too steep to hold a tree.
	return g.Slope(p) > g.steepLine
}

// HoldsWood reports whether trees will take on this tile: whether its ground
// is above the tree line. What is already wooded is left alone - a standing
// wood is a fact about the map, however it got there - so this is asked of
// ground that is about to become a wood, by seeding or by planting, and not
// of ground that is one.
func (g *Grid) HoldsWood(p entity.Pos) bool {
	if !g.In(p) {
		return false
	}
	if !g.woodsRead {
		g.readWoods()
	}
	return !g.TooSteep(p) && g.WoodsAt(p) >= g.woodsLine
}

// readWoods takes the map's measure of itself: how steep its steep ground is,
// and where the tree line falls on it. Both are comparisons with the rest of
// the map rather than fixed numbers, because a fixed cutoff gives one map a
// forest and the next a heath. It is run when the land is made and again
// whenever the weather has moved it.
func (g *Grid) readWoods() {
	slopes := make([]float64, len(g.Tiles))
	for i := range g.Tiles {
		slopes[i] = g.Slope(entity.Pos{X: i % g.W, Y: i / g.W})
	}
	g.steepAt = quantile(slopes, 0.9)
	g.steepLine = quantile(slopes, 1-woodsSteep)
	g.woodsRead = true // the slope readings are in; WoodsAt may be asked now

	suits := make([]float64, 0, len(g.Tiles))
	for i := range g.Tiles {
		if g.Tiles[i].Terrain == Water {
			continue // the river is not ground trees might have had
		}
		suits = append(suits, g.WoodsAt(entity.Pos{X: i % g.W, Y: i / g.W}))
	}
	if len(suits) == 0 {
		return
	}
	g.woodsLine = quantile(suits, 1-woodsShare)
}
