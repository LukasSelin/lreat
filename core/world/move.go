package world

import (
	"math"

	"lreat/core/entity"
)

// moveCost is the effort of entering a tile of each terrain, measured in
// ticks. Grass is the unit. Ground that fights back costs more of both things
// an agent has to spend: time, because a slow tile is several ticks of walking
// instead of one, and body, because the exertion drains the physiological
// tier in proportion. Terrain is therefore not decoration; it is a standing
// tax on every plan that crosses it, and the map shapes where people settle,
// what they walk to, and which side of the river they give up on.
var moveCost = [...]float64{
	Grass:  1,
	Field:  1.3,
	Forest: 2.2,
	Water:  3.5,
}

// MoveCost returns the ticks of effort needed to enter p. Anything built on a
// tile has cleared and trodden it, so structures are as cheap as open grass.
// Tiles off the map are infinitely expensive, which keeps agents inside it.
func (g *Grid) MoveCost(p entity.Pos) float64 {
	if !g.In(p) {
		return math.Inf(1)
	}
	t := g.At(p)
	if t.Structure != None {
		return moveCost[Grass]
	}
	return moveCost[t.Terrain]
}

// StepToward returns the tile an agent at from should enter next on its way to
// to: of the neighbours that close the distance, the cheapest to walk into.
// It is greedy rather than a search, so agents slip around a thicket that is
// in the way but still ford a river that lies square across the route. Ties
// keep the straight-line step, and the scan order is fixed, so runs repeat.
func (g *Grid) StepToward(from, to entity.Pos) entity.Pos {
	if from == to {
		return from
	}
	best := entity.StepToward(from, to)
	bestCost := g.MoveCost(best)
	d := entity.Dist(from, to)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			p := entity.Pos{X: from.X + dx, Y: from.Y + dy}
			if p == best || !g.In(p) || entity.Dist(p, to) >= d {
				continue
			}
			if c := g.MoveCost(p); c < bestCost {
				best, bestCost = p, c
			}
		}
	}
	return best
}

// TravelCost estimates the ticks of walking from one tile to another by
// costing the route the agent would actually take. Deciding uses it in place
// of raw distance, so a target across the water is judged as far as the wading
// makes it, not as near as the crow flies.
func (g *Grid) TravelCost(from, to entity.Pos) float64 {
	var total float64
	p := from
	for i := entity.Dist(from, to); i > 0 && p != to; i-- {
		p = g.StepToward(p, to)
		total += g.MoveCost(p)
	}
	return total
}
