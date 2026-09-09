package world

import (
	"math"
	"sort"

	"lreat/core/entity"
)

// Weathering: the land does not hold still.
//
// An age of weather - a decade of it - strips soil off the ground in
// proportion to how much water crosses it, how steeply it lies, and how
// little is holding it down; carries what it strips downhill; and lays it
// down again where the water slows. The
// heights change, so the drainage is worked out again, so the rivers are where
// the new ground sends them. Nothing is moved by hand.
//
// What makes this worth having is not that hills wear down. It is that how
// fast they wear down is partly the settlement's doing. Woods hold a hillside
// together and a ploughed field does not, so a people that clears its slopes
// to farm them washes those slopes into its own river, silts its own valley,
// and finds the soil it depended on in a different place from where it left
// it. Nobody decides that; it falls out of where they chose to put their
// fields.

// Wash is how much soil an age of weather - a decade of it; see
// system.ErodeEvery - takes off a tile, given the water crossing it and the
// steepness of it. What the water can lift goes as the root of how much of it
// there is rather than in proportion: taken in proportion, the valley floor
// carries so much of the map's water that it scoured itself out instead of
// silting up, which is the opposite of what a flood plain is. The root is the
// usual reading, and with it the channel still cuts down while the ground
// beside it fills.
//
// The size of it is what makes the ground move at the speed ground moves: a
// ploughed slope loses a few centimetres of soil a decade and a wooded one a
// few millimetres, so a hillside farmed hard is worn out in a century or two
// and one left standing keeps what it has for longer than anybody watching it
// will be alive.
const Wash = 12

// Settle is the share of what the water is carrying that it puts down on
// gentle ground each tile it crosses. Steep ground keeps its load moving.
const Settle = 0.35

// SettleSlope is the slope above which water carries everything it has and
// lays down nothing.
const SettleSlope = 0.12

// Overbank is the share of what a river lays down that it lays down outside
// its own channel, on the low ground either side.
const Overbank = 0.7

// hold is how much of the soil on a tile stays put, by what is growing or
// standing on it. Woods are what hold a hillside together; a ploughed field
// is bare earth by another name; and a roof or a road takes the ground it
// covers out of the weather altogether.
func hold(t *Tile) float64 {
	if t.Structure != None {
		return 0
	}
	return t.Terrain.Hold()
}

// Erode weathers the map by one age and works the drainage out again. It is
// the one thing that changes the shape of the land after the map is made, and
// everything the shape decides - where the rivers run, what the soil will
// hold, how dear it is to walk - follows from it without being told to.
func (w *World) Erode() {
	g := w.Grid
	n := len(g.Tiles)
	// The whole ground moves at once, so the ground asleep is brought up to
	// date first; what the age does to it is done to it as it now stands.
	w.CatchUpAll()

	// Highest ground first, so that what a tile sheds is in the water before
	// the tile below it is asked what the water is carrying.
	order := make([]int32, n)
	for i := range order {
		order[i] = int32(i)
	}
	sort.Slice(order, func(a, b int) bool {
		ha, hb := g.Tiles[order[a]].Height, g.Tiles[order[b]].Height
		if ha != hb {
			return ha > hb
		}
		return order[a] < order[b] // ties by position, so an age repeats
	})

	load := make([]float64, n)   // soil in the water leaving each tile
	change := make([]float64, n) // metres gained or lost
	for _, i := range order {
		t := &g.Tiles[i]
		p := entity.Pos{X: int(i) % g.W, Y: int(i) / g.W}
		slope := g.Slope(p)

		// What the water lays down here: more of it the gentler the ground.
		if load[i] > 0 {
			settled := load[i] * Settle * clamp01(1-slope/SettleSlope)
			load[i] -= settled
			if t.Wet() {
				// A river in flood puts most of its silt over the bank. That
				// is what a flood plain is: not ground the river spared, but
				// ground the river made. Without it the silt stays in the
				// channel, the bed rises, and the good land beside it slowly
				// washes away instead of being fed.
				var bank []int
				for _, off := range dirs {
					c := entity.Pos{X: p.X + off.X, Y: p.Y + off.Y}
					if !g.In(c) {
						continue
					}
					if b := g.At(c); !b.Wet() && b.Drain < FloodDepth {
						bank = append(bank, g.Index(c))
					}
				}
				if len(bank) > 0 {
					over := settled * Overbank
					for _, j := range bank {
						change[j] += over / float64(len(bank))
					}
					settled -= over
				}
			}
			change[i] += settled
		}
		// What it takes away.
		stripped := Wash * math.Sqrt(t.Flow) * slope * hold(t)
		change[i] -= stripped
		load[i] += stripped

		a := g.Aspect(p)
		if a == (entity.Pos{}) {
			continue // the water and everything in it leaves the map here
		}
		down := int32(g.Index(entity.Pos{X: p.X + a.X, Y: p.Y + a.Y}))
		load[down] += load[i]
	}

	for i := range g.Tiles {
		t := &g.Tiles[i]
		t.Height = math.Max(0, t.Height+change[i])
		// Soil goes with the ground it was in. What washes off a slope is
		// what that slope could have grown; what lands on the flat is what
		// makes a flood plain worth farming.
		if !t.Wet() {
			t.Rich = clamp01(t.Rich + change[i]/SoilDepth)
			t.Fertility = math.Min(t.Fertility, t.Rich)
		}
	}

	g.fill()
	g.drain()
	g.carve(w.RNG)
	g.height()
	g.resoil()
	// The ground has moved, so the tree line has moved with it: what was a
	// dry shoulder may now be damp enough to hold a wood, and what the water
	// has cut into may not.
	g.readWoods()
	g.Recount() // the water has moved, and the woods with it
}

// SoilDepth is how many metres of ground make the difference between land
// that will grow anything and land that will grow nothing. A settlement can
// strip a hillside of it in a few lifetimes of hard farming.
const SoilDepth = 3.0

// resoil lets ground that the moving water has made better become better:
// a flat newly within reach of the flood comes up toward what such ground
// holds, a little each age rather than overnight.
//
// It only ever raises. What lowers soil is the weather taking it away, above,
// and nothing else should: a field somebody has cut a channel to holds more
// than the bare ground around it would, and that is the whole point of having
// dug it. An earlier version pulled every tile toward what its drainage alone
// would give, which quietly undid irrigation every age and cost the
// settlements that had invested in it dearly.
func (g *Grid) resoil() {
	const toward = 0.08
	for i := range g.Tiles {
		t := &g.Tiles[i]
		if t.Wet() {
			continue
		}
		p := entity.Pos{X: i % g.W, Y: i / g.W}
		if can := g.SoilAt(p); can > t.Rich {
			t.Rich += toward * (can - t.Rich)
		}
		t.Fertility = math.Min(t.Fertility, t.Rich)
	}
}
