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
//
// Water stays where it was, and the reason is worth recording. A settlement
// grows on both banks, because the ground worth farming is the ground near
// the river, so a quarter of its people were spending their lives wading. The
// obvious fix was to make the water dearer. It was tried at 5, 7 and 9 and it
// was the wrong fix: a river nobody can afford to cross is a river nobody
// wears a ford in, and a ford nobody wears is a ford nobody bridges. Dearer
// water cut the wading barely at all and cost up to a quarter of the
// population. What answers a river is a bridge, and the cheapest water is
// what gets one built.
var moveCost = [TerrainCount]float64{
	Grass:  1,
	Field:  1.3,
	Forest: 2.2,
	Water:  3.5,
	Rock:   1.8,
}

// structureCost is the effort of entering a tile that has been built on, and
// it overrides the terrain underneath. A market is a trodden square, as easy
// as open grass. A house is a wall and a hearth rather than a thoroughfare,
// so crossing one is slower than walking round it. A road is the only thing
// built purely to be walked on, and it is the fastest ground on the map.
//
// The house toll is deliberately mild. Recognition reads distance straight
// off the ground - a costly walk makes an errand read as a poor fit, rather
// than merely dividing its worth as the value rule does - so a settlement
// that grows dear to cross degrades the very judgement its people make, and
// the worse it gets the more it builds. At 1.6 that loop was enough to stop
// a settlement replacing its founders. Roads are the answer to it; until
// somebody decides to lay them, the toll stays where a growing city can
// carry it.
var structureCost = [...]float64{
	None:    0, // unused: terrain decides
	House:   1.3,
	Granary: 1.3,
	Tavern:  1.2,
	Market:  1,
	Road:    0.5,
}

// Saving is what a road laid on p would take off each crossing of it, as a
// multiple of what a road on grass takes off. It is the other half of what a
// length of road is worth, and until it was asked the only half being read
// was how many people walk there.
//
// A road costs the same timber wherever it goes but does not save the same
// amount. On grass it turns a 1 into a 0.5; through a wood a 2.2, and over
// water a 3.5, and that last is why a bridge is worth six lengths of ordinary
// street to the people who cross it. Read on wear alone a settlement paves
// the flat ground it was already crossing easily and leaves the marsh, the
// thicket and the river - the places where the going is dear, which is to say
// the places where a road is the whole point. Weighing the wear by this is
// what lets a crossing win on its merits: it earned an exception before,
// because the bar it could not clear was written for a lane.
//
// Grass is the unit, so open ground reads exactly as worn as it is walked.
// The bodily saving is not in here: a road is easier underfoot as well as
// quicker, but roadDrain is the same wherever the road lies, so it says
// nothing about which ground is worth paving. Nor is the climb, which a road
// does not flatten - see StepCost, where the slope is added whatever is
// built on the tile.
func (g *Grid) Saving(p entity.Pos) float64 {
	if !g.In(p) {
		return 0
	}
	paved := structureCost[Road]
	return (moveCost[g.At(p).Terrain] - paved) / (moveCost[Grass] - paved)
}

// roadDrain is how much of the ordinary bodily cost a tick of walking on a
// road exacts. Paving saves more than the time it saves: a firm level surface
// is easier underfoot as well as quicker, so a road is cheap twice over. This
// is what makes a street worth the labour of laying it, and what makes the
// shape of a settlement's roads show up in how tired its people are.
const roadDrain = 0.7

// SwimLoad is the most a walker may be carrying and still take to the water.
// It is not a heavy pack; it is nothing at all, near enough. People are poor
// swimmers with both arms free, and a person holding a sack of grain over a
// river is a person drowning: what they do in life is put the sack down or
// walk to the bridge. So the water is not dear to a laden walker, it is shut,
// and the threshold is here only so that a crumb left in a pocket does not
// count as cargo.
//
// This is what makes a bridge worth its timber to somebody who already lives
// beside a ford. Wading was always slow; now it is the difference between
// carrying the harvest home and not carrying it at all, and the far bank is
// only part of the settlement for as long as the crossing stands.
const SwimLoad = 0.1

// Carrying tells the router how much the walker it is about to route for is
// holding, so that a laden walker is routed round open water instead of
// through it. It holds for the next route this router runs and no longer,
// which is what keeps a load from leaking into somebody else's journey.
func (r *Router) Carrying(load float64) *Router {
	r.load = load
	return r
}

// Carrying routes on the grid's own router, for callers working one at a
// time.
func (g *Grid) Carrying(load float64) *Router {
	return g.ownRouter().Carrying(load)
}

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

// Climb and Descend are the ticks a metre of rise and a metre of fall add to
// a step. Going up is what costs: a steep tile on this map rises seven metres
// or so, which is most of another tile's walking on top of the ground itself,
// and that is what makes a route round the shoulder of a hill cheaper than a
// route over it. Coming down is charged a little too, because a walker picks
// their way down a bank rather than running at it, and because a step that
// cost nothing downhill would make a zigzag look free.
const (
	Climb   = 0.10
	Descend = 0.02
)

// StepCost is the effort of moving from one tile to the next: the ground
// being entered, plus the climb or the descent into it, plus the fence
// between them if there is one. It is what routing costs a journey by, so
// agents round a hill rather than going over it, walk round a hedged holding
// rather than through the corn, and so the ways they wear - and the roads
// they lay on those ways - follow the contours and the valley floors the way
// real ones do.
func (g *Grid) StepCost(from, to entity.Pos) float64 {
	return g.StepCostFor(from, to, 0)
}

// StepCostFor is StepCost for a named walker, whose own fields are theirs to
// walk into. Everybody else's cost them the climb over the fence.
func (g *Grid) StepCostFor(from, to entity.Pos, holder entity.ID) float64 {
	c := g.MoveCost(to)
	if math.IsInf(c, 1) || !g.In(from) {
		return c
	}
	c += g.fenceCost(int32(from.Y*g.W+from.X), int32(to.Y*g.W+to.X), holder)
	if d := g.Height(to) - g.Height(from); d > 0 {
		return c + Climb*d
	} else {
		return c - Descend*d
	}
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
	return r.route(&r.scratch, from, to, true, g.Toward(from, to)).Step(to)
}

// StepToward routes on the grid's own router, for callers working one at a
// time.
func (g *Grid) StepToward(from, to entity.Pos) entity.Pos {
	return g.ownRouter().StepToward(from, to)
}

// Path is the cheapest way from one tile to another, from excluded and to
// included. It is what an agent is given to walk when it settles on a plan,
// so that the way is worked out once rather than re-asked at every step.
func (r *Router) Path(from, to entity.Pos) []entity.Pos {
	g := r.g
	if from == to || !g.In(to) {
		return nil
	}
	return r.route(&r.scratch, from, to, true, g.Toward(from, to)).Path(to)
}

// Path routes on the grid's own router, for callers working one at a time.
func (g *Grid) Path(from, to entity.Pos) []entity.Pos {
	return g.ownRouter().Path(from, to)
}

// TravelCost is the ticks of walking from one tile to another along the route
// the agent would actually take. Deciding uses it in place of raw distance,
// so a target across the water is judged as far as the wading makes it, and
// one along a street as near as the paving makes it.
func (r *Router) TravelCost(from, to entity.Pos) float64 {
	g := r.g
	if from == to {
		r.load = 0
		return 0
	}
	if !g.In(to) {
		r.load = 0
		return math.Inf(1)
	}
	if r.surveyed && from == r.spreadFrom && (r.load > SwimLoad) == r.spreadLaden && r.holder == r.spreadHolder {
		r.load = 0
		return r.fromSurvey(to)
	}
	return r.route(&r.scratch, from, to, true, g.Toward(from, to)).Cost(to)
}

// TravelCost routes on the grid's own router, for callers working one at a
// time.
func (g *Grid) TravelCost(from, to entity.Pos) float64 {
	return g.ownRouter().TravelCost(from, to)
}
