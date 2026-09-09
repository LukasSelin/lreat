package world

import (
	"math"

	"lreat/core/entity"
)

// GenerateTerrain raises the ground, lets the water find its way down it, and
// reads everything else off what that leaves: the woods where it is damp and
// not too steep, the outcrops where it is high and bare, the good soil on the
// valley floors and the sunny slopes. Nothing here is drawn on top of the
// land; see relief.go for the shape of it.
func (w *World) GenerateTerrain(width, height int) {
	w.Generate(Config{Width: width, Height: height})
}

// Generate is GenerateTerrain on the given terms.
func (w *World) Generate(cfg Config) {
	width, height := cfg.Width, cfg.Height
	g := NewGrid(width, height)
	g.Wrap = cfg.Wrap

	w.raise(g)
	g.fill()
	g.drain()
	g.carve(w.RNG)
	g.height()

	// Woods stand where the ground is damp enough to grow them and gentle
	// enough to hold soil: the valley sides above the flood, not the crown of
	// the ridge and not the bed of the river. Each tile is scored on how well
	// it suits trees, with a little luck thrown in so that two maps with the
	// same bones are not the same map, and the best of them are wooded. The
	// share is fixed rather than the score, because how wet a map is depends
	// on the shape of it and a settlement needs roughly the same timber
	// whatever ground it was given.
	g.readWoods()
	wooded := make([]float64, 0, len(g.Tiles))
	score := make([]float64, len(g.Tiles))
	for i := range g.Tiles {
		score[i] = g.WoodsAt(entity.Pos{X: i % width, Y: i / width}) + 0.35*w.RNG.Float64()
		if g.Tiles[i].Terrain == Grass {
			wooded = append(wooded, score[i])
		}
	}
	treeLine := quantile(wooded, 1-forestShare)
	for i := range g.Tiles {
		t := &g.Tiles[i]
		p := entity.Pos{X: i % width, Y: i / width}
		if t.Terrain != Grass || score[i] < treeLine || g.TooSteep(p) {
			continue
		}
		t.Terrain = Forest
		t.Wood = 0.6 + 0.4*w.RNG.Float64()
		t.Wild = 0.6 + 0.4*w.RNG.Float64()
		t.Standing() // the woods a map is made with are old woods
	}

	// Outcrops are where the soil has gone: high, steep ground the water runs
	// off rather than soaks into. Scored and shared the same way, because an
	// outcrop is a comparison with the rest of the map, not a measurement.
	heights := make([]float64, len(g.Tiles))
	slopes := make([]float64, len(g.Tiles))
	for i := range g.Tiles {
		heights[i] = g.Tiles[i].Height
		slopes[i] = g.Slope(entity.Pos{X: i % width, Y: i / width})
	}
	highAt := quantile(heights, 0.6)
	bare := make([]float64, len(g.Tiles))
	open := make([]float64, 0, len(g.Tiles))
	for i := range g.Tiles {
		bare[i] = clamp01(slopes[i]/math.Max(1e-12, g.steepAt)) +
			clamp01((heights[i]-highAt)/math.Max(1e-12, Relief-highAt))
		if g.Tiles[i].Terrain == Grass {
			open = append(open, bare[i])
		}
	}
	stoneLine := quantile(open, 1-rockShare)
	for i := range g.Tiles {
		if t := &g.Tiles[i]; t.Terrain == Grass && bare[i] >= stoneLine {
			t.Terrain = Rock
		}
	}

	// Good soil is where the water has been and stopped: the flat of a valley,
	// damp from what drains through it, facing the sun. See Grid.SoilAt, which
	// is the same reading the weather takes every age.
	for i := range g.Tiles {
		t := &g.Tiles[i]
		if t.Terrain == Water {
			continue
		}
		t.Fertility = g.SoilAt(entity.Pos{X: i % width, Y: i / width})
		t.Rich = t.Fertility
	}
	g.Recount()
	w.Forest0 = g.Forest()

	// The market goes where a settlement would put it: dry, gentle ground
	// beside the largest water near the middle of the map, which after the
	// draining is a place the land chose rather than one picked in advance.
	center := entity.Pos{X: width / 2, Y: height / 2}
	best, bestScore := entity.Pos{}, math.Inf(-1)
	for i := range g.Tiles {
		p := entity.Pos{X: i % width, Y: i / width}
		if g.Tiles[i].Terrain == Water {
			continue
		}
		if !g.HasNeighbor(p, func(t *Tile) bool { return t.Terrain == Water }) {
			continue
		}
		// What founds a market: good soil, the flat of the valley rather than
		// the bank above it, level ground to stand on, and somewhere near the
		// middle of the country. Weighted this way round because a settlement
		// picks its ground first and its distance from anywhere second - the
		// other way round put markets on dry hillsides, and the settlements
		// that grew there starved.
		t := &g.Tiles[i]
		score := 2*t.Fertility +
			1.2*clamp01(1-t.Drain/FloodDepth) -
			4*g.Slope(p) -
			0.03*float64(g.Dist(p, center))
		if score > bestScore {
			best, bestScore = p, score
		}
	}
	mp := best
	if math.IsInf(bestScore, -1) {
		mp, _ = g.Nearest(center, width+height, func(_ entity.Pos, t *Tile) bool { return t.Terrain == Grass })
	}
	if g.At(mp).Terrain != Grass {
		g.Turn(mp, Grass)
	}
	g.Build(mp, Market)
	w.Grid = g
	w.MarketPos = mp
	w.FoundMarket(mp)
}
