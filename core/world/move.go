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

// structureCost is the effort of entering a tile that has been built on, and
// it overrides the terrain underneath. A market is a trodden square, as easy
// as open grass. A house is a wall and a hearth rather than a thoroughfare,
// so crossing one is slower than walking round it. A road is the only thing
// built purely to be walked on, and it is the fastest ground on the map.
var structureCost = [...]float64{
	None:   0, // unused: terrain decides
	House:  1.6,
	Market: 1,
	Road:   0.5,
}

// roadDrain is how much of the ordinary bodily cost a tick of walking on a
// road exacts. Paving saves more than the time it saves: a firm level surface
// is easier underfoot as well as quicker, so a road is cheap twice over. This
// is what makes a street worth the labour of laying it, and what makes the
// shape of a settlement's roads show up in how tired its people are.
const roadDrain = 0.7

// MoveCost returns the ticks of effort needed to enter p. Tiles off the map
// are infinitely expensive, which keeps agents inside it.
func (g *Grid) MoveCost(p entity.Pos) float64 {
	if !g.In(p) {
		return math.Inf(1)
	}
	t := g.At(p)
	if t.Structure != None {
		return structureCost[t.Structure]
	}
	return moveCost[t.Terrain]
}

// MoveDrain returns how hard on the body a tick of walking into p is, as a
// multiple of the ordinary cost. Only paving changes it.
func (g *Grid) MoveDrain(p entity.Pos) float64 {
	if g.In(p) && g.At(p).Structure == Road {
		return roadDrain
	}
	return 1
}

// StepToward returns the tile an agent at from should enter next on its way
// to to: the first step of the cheapest route there. Because the route is
// costed rather than guessed at, a walker rounds a thicket, fords a river
// only where fording beats going round, and joins a road that runs its way
// even when the road starts off to one side. Ties keep the straight-line
// step, so runs repeat.
func (r *Router) StepToward(from, to entity.Pos) entity.Pos {
	g := r.g
	if from == to || !g.In(to) {
		return from
	}
	stop := int32(to.Y*g.W + to.X)
	return r.route(&r.scratch, from, stop, entity.StepToward(from, to)).Step(to)
}

// StepToward routes on the grid's own router, for callers working one at a
// time.
func (g *Grid) StepToward(from, to entity.Pos) entity.Pos {
	return g.ownRouter().StepToward(from, to)
}

// TravelCost is the ticks of walking from one tile to another along the route
// the agent would actually take. Deciding uses it in place of raw distance,
// so a target across the water is judged as far as the wading makes it, and
// one along a street as near as the paving makes it.
func (r *Router) TravelCost(from, to entity.Pos) float64 {
	g := r.g
	if from == to {
		return 0
	}
	if !g.In(to) {
		return math.Inf(1)
	}
	stop := int32(to.Y*g.W + to.X)
	return r.route(&r.scratch, from, stop, entity.StepToward(from, to)).Cost(to)
}

// TravelCost routes on the grid's own router, for callers working one at a
// time.
func (g *Grid) TravelCost(from, to entity.Pos) float64 {
	return g.ownRouter().TravelCost(from, to)
}
