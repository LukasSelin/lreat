package world

import (
	"math"
	"sort"

	"lreat/core/entity"
)

// The shape of the land, and the water that shape decides.
//
// Everything in this file happens once, before anybody is born, but it is
// written to be run again: heights are a field, drainage is derived from
// heights, and the rivers are derived from drainage. Nothing about the water
// is drawn by hand, so if the ground ever moves - silt, erosion, a dam - the
// rivers follow it by being recomputed rather than by being redrawn.
//
// The order is the one a landscape actually obeys. Raise the ground; fill the
// hollows that have no outlet, because a hollow either fills until it spills
// or it is a lake; send every tile's water to its lowest neighbour; add up
// what passes through each tile; and call the tiles that carry enough of it a
// river. Fertility, woods and outcrops then read off the finished land
// instead of being scattered over it.

// TileSpan is how wide a tile is on the ground, in metres. It is what turns a
// difference in height into a slope, and so the only reason heights and
// distances can be spoken of in the same breath.
const TileSpan = 25.0

// Relief is the height in metres between the lowest ground a map can have and
// the highest. Sixty metres over eighty tiles is a river valley with sides to
// it, not a mountain range: enough that walking uphill is felt and that water
// knows where to go.
const Relief = 60.0

// waterShare is how much of a map ends up as watercourse. The threshold that
// achieves it is read off each map's own drainage rather than fixed, because
// how much water a given amount of falling ground gathers varies enormously
// with the shape of it: over a handful of seeds the heaviest-draining tile
// carried anywhere from a fifth of the map to four fifths. A fixed cutoff
// gives one map a river and the next a puddle. A share gives every map a
// river of its own size.
const waterShare = 0.045

// rockShare and rockSteep are how much of a map is bare stone, and how steep
// ground has to be, in its own terms, to be a candidate for it.
const (
	rockShare   = 0.015
	forestShare = 0.13
)

// FloodDepth is how far above its river ground stops being valley floor, in
// metres. Below it the soil is what the water left; above it the ground is
// what the weather gives it.
const FloodDepth = 14.0

// SoilAt is what the land at p will hold: good on the damp flat of a valley
// facing the sun, poor on a steep dry hillside. It is read off the drainage
// rather than stored, so that when the ground moves the soil that the ground
// can carry moves with it. The map is made with it and every age of weather
// pulls the soil that is actually there toward it.
func (g *Grid) SoilAt(p entity.Pos) float64 {
	t := g.At(p)
	damp := clamp01(1 - t.Drain/FloodDepth)
	steep := clamp01(g.Slope(p) / 0.25)
	return clamp01(0.15 + 0.85*damp*(1-0.7*steep)*(0.75+0.5*g.Sunlight(p)))
}

// clamp01 holds a share inside [0,1].
func clamp01(v float64) float64 { return math.Max(0, math.Min(1, v)) }

// quantile returns the value at f through a sorted copy of v. It is how the
// generator turns "the steepest tenth" or "the wettest twentieth" into a
// number for this particular map.
func quantile(v []float64, f float64) float64 {
	c := append([]float64(nil), v...)
	sort.Float64s(c)
	i := int(f * float64(len(c)-1))
	return c[i]
}

// Height is the height of a tile in metres. Off the map it is the sea the
// water eventually reaches, which is what makes every hollow drain somewhere.
func (g *Grid) Height(p entity.Pos) float64 {
	if !g.In(p) {
		return -1
	}
	return g.At(p).Height
}

// Slope is how steeply the ground falls away from a tile: the greatest drop
// to any neighbour, as a rise over a run. A tenth is a gentle hill, a half is
// ground you would not plough.
func (g *Grid) Slope(p entity.Pos) float64 {
	h := g.Height(p)
	steepest := 0.0
	for _, off := range dirs {
		q := entity.Pos{X: p.X + off.X, Y: p.Y + off.Y}
		if !g.In(q) {
			continue
		}
		run := TileSpan
		if off.X != 0 && off.Y != 0 {
			run *= math.Sqrt2
		}
		if d := (h - g.Height(q)) / run; d > steepest {
			steepest = d
		}
	}
	return steepest
}

// Aspect is the way a slope faces: the step toward the lowest neighbour, which
// is the way water leaves and the way the ground looks. The zero step means
// level ground, or a hollow with nowhere lower to go.
func (g *Grid) Aspect(p entity.Pos) entity.Pos {
	best, lowest := entity.Pos{}, g.Height(p)
	for _, off := range dirs {
		q := entity.Pos{X: p.X + off.X, Y: p.Y + off.Y}
		if g.In(q) && g.Height(q) < lowest {
			best, lowest = off, g.Height(q)
		}
	}
	return best
}

// Sunlight is how much of the day's warmth a tile's face catches, in [0,1].
// North is up the map, so ground that falls away southward looks at the sun
// and ground that falls away northward stands in its own shadow. Level ground
// is halfway between. Steep ground makes more of whichever it is.
func (g *Grid) Sunlight(p entity.Pos) float64 {
	a := g.Aspect(p)
	if a == (entity.Pos{}) {
		return 0.5
	}
	// a.Y is positive southward, and a southward-facing slope is the sunny one.
	lean := float64(a.Y) / math.Sqrt(float64(a.X*a.X+a.Y*a.Y))
	return 0.5 + 0.5*lean*math.Min(1, g.Slope(p)/0.3)
}

// raise builds the height field: several lattices of random corners, each
// half the width and half the height of the last, added together. The coarse
// ones are the valley and the ridge; the fine ones are the unevenness that
// keeps a hillside from being a ramp.
func (w *World) raise(g *Grid) {
	h := make([]float64, len(g.Tiles))
	amp, span, total := 1.0, float64(max(g.W, g.H))/2, 0.0
	for octave := 0; octave < 5 && span >= 2; octave++ {
		lattice := w.lattice(g, span)
		for i := range h {
			h[i] += amp * lattice[i]
		}
		total += amp
		amp, span = amp/2, span/2
	}
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, v := range h {
		lo, hi = math.Min(lo, v), math.Max(hi, v)
	}
	scale := Relief / math.Max(1e-9, hi-lo)
	for i := range g.Tiles {
		g.Tiles[i].Height = (h[i] - lo) * scale
	}
	_ = total
}

// lattice is one octave: random corners span tiles apart, smoothly blended.
func (w *World) lattice(g *Grid, span float64) []float64 {
	cols, rows := int(float64(g.W)/span)+2, int(float64(g.H)/span)+2
	corner := make([]float64, cols*rows)
	for i := range corner {
		corner[i] = w.RNG.Float64()
	}
	out := make([]float64, len(g.Tiles))
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			fx, fy := float64(x)/span, float64(y)/span
			cx, cy := int(fx), int(fy)
			tx, ty := smooth(fx-float64(cx)), smooth(fy-float64(cy))
			at := func(dx, dy int) float64 { return corner[(cy+dy)*cols+(cx+dx)] }
			top := at(0, 0)*(1-tx) + at(1, 0)*tx
			bot := at(0, 1)*(1-tx) + at(1, 1)*tx
			out[y*g.W+x] = top*(1-ty) + bot*ty
		}
	}
	return out
}

// smooth is the ease that turns a lattice of corners into hills rather than
// facets: flat where it meets a corner, steepest halfway between.
func smooth(t float64) float64 { return t * t * (3 - 2*t) }

// fill raises every hollow to the level at which it would spill, so that all
// ground drains somewhere and water is never asked to run uphill. It works
// inward from the edges of the map, always from the lowest ground reached so
// far, which is the order water itself would fill a landscape in.
func (g *Grid) fill() {
	filled := make([]float64, len(g.Tiles))
	done := make([]bool, len(g.Tiles))
	q := &heightQueue{}
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			if x != 0 && y != 0 && x != g.W-1 && y != g.H-1 {
				continue
			}
			i := y*g.W + x
			filled[i], done[i] = g.Tiles[i].Height, true
			q.push(heightNode{h: filled[i], idx: int32(i)})
		}
	}
	// A hair of fall per tile, so that a filled flat still has a direction to
	// send its water and does not become a puddle with no outlet.
	const seep = 1e-4
	for q.len() > 0 {
		n := q.pop()
		p := entity.Pos{X: int(n.idx) % g.W, Y: int(n.idx) / g.W}
		for _, off := range dirs {
			c := entity.Pos{X: p.X + off.X, Y: p.Y + off.Y}
			if !g.In(c) {
				continue
			}
			j := c.Y*g.W + c.X
			if done[j] {
				continue
			}
			filled[j] = math.Max(g.Tiles[j].Height, filled[n.idx]+seep)
			done[j] = true
			q.push(heightNode{h: filled[j], idx: int32(j)})
		}
	}
	for i := range g.Tiles {
		g.Tiles[i].Height = filled[i]
	}
}

// drain sends every tile's water to its lowest neighbour and adds up what
// passes through, so that Flow is the share of the map draining through each
// tile. Tiles are settled from the highest down, which is the only order in
// which a tile's own total is complete before it is passed on.
func (g *Grid) drain() {
	n := len(g.Tiles)
	order := make([]int32, n)
	for i := range order {
		order[i] = int32(i)
		g.Tiles[i].Flow = 1 / float64(n)
	}
	sort.Slice(order, func(a, b int) bool {
		ha, hb := g.Tiles[order[a]].Height, g.Tiles[order[b]].Height
		if ha != hb {
			return ha > hb
		}
		return order[a] < order[b] // ties settled by position, so runs repeat
	})
	for _, i := range order {
		p := entity.Pos{X: int(i) % g.W, Y: int(i) / g.W}
		a := g.Aspect(p)
		if a == (entity.Pos{}) {
			continue // the edge of the map: the water leaves here
		}
		down := entity.Pos{X: p.X + a.X, Y: p.Y + a.Y}
		g.At(down).Flow += g.Tiles[i].Flow
	}
}

// carve puts the water where the flow says it goes: the wettest waterShare of
// the map is river, and the heaviest of it spreads onto the lower bank beside
// it, as a river does. Ground the water has left goes back to grass.
//
// It runs at the making of the map and again after every age of weather, so a
// river can take a course it did not have. It will not run through anything
// anybody has built or claimed: a settlement embanks what it stands on, and a
// river that swallowed the market would be the end of a run rather than an
// event in it.
func (g *Grid) carve(rng interface{ Float64() float64 }) {
	flows := make([]float64, len(g.Tiles))
	for i := range g.Tiles {
		flows[i] = g.Tiles[i].Flow
	}
	cut := quantile(flows, 1-waterShare)
	big := quantile(flows, 1-waterShare/4)

	// A channel does not flicker. Ground becomes river when the water really
	// gathers there, and stops being river only when the water has largely
	// gone - not the moment it dips below the line. Without that hysteresis a
	// settlement wipes out its own river: it holds the ground the shifting
	// channel wants, so the new course cannot form, while the old one dries
	// the instant it falls under the threshold.
	wet := make([]bool, len(g.Tiles))
	for i := range g.Tiles {
		if g.Tiles[i].Terrain == Water {
			wet[i] = g.Tiles[i].Flow >= cut/2
		} else {
			wet[i] = g.Tiles[i].Flow >= cut
		}
	}
	// The great rivers spread onto whatever beside them is no higher.
	for i := range g.Tiles {
		if g.Tiles[i].Flow < big {
			continue
		}
		p := entity.Pos{X: i % g.W, Y: i / g.W}
		for _, off := range dirs {
			c := entity.Pos{X: p.X + off.X, Y: p.Y + off.Y}
			if g.In(c) && g.Height(c) <= g.Height(p)+1 {
				wet[c.Y*g.W+c.X] = true
			}
		}
	}
	for i := range g.Tiles {
		t := &g.Tiles[i]
		held := t.Structure != None || t.Owner != 0
		switch {
		case wet[i] && t.Terrain != Water && !held:
			t.Terrain, t.Wood, t.Wild = Water, 0, 0
			t.Fish = 0.7 + 0.3*rng.Float64()
		case !wet[i] && t.Terrain == Water:
			t.Terrain, t.Fish = Grass, 0
		}
	}
}

// height reads how far a tile stands above the water it drains into, in
// metres, and writes it into Drain. It is the truest thing the land can say
// about how wet a place is, and much truer than how much water passes through
// it: a tile on a valley floor beside the river carries hardly any flow of its
// own and is still a water meadow, while a tile halfway up a hillside may
// carry a whole gully's worth and still be dry as a bone. Flood plains sit at
// nothing, terraces a few metres up, hillsides tens.
//
// It is computed by following each tile's water down to the river it joins and
// adding up the fall on the way. Tiles are settled lowest first, so the tile
// downstream is always finished before the one that drains into it.
func (g *Grid) height() {
	n := len(g.Tiles)
	order := make([]int32, n)
	lowest := math.Inf(1)
	for i := range order {
		order[i] = int32(i)
		lowest = math.Min(lowest, g.Tiles[i].Height)
	}
	sort.Slice(order, func(a, b int) bool {
		ha, hb := g.Tiles[order[a]].Height, g.Tiles[order[b]].Height
		if ha != hb {
			return ha < hb
		}
		return order[a] < order[b]
	})
	for _, i := range order {
		t := &g.Tiles[i]
		if t.Terrain == Water {
			t.Drain = 0
			continue
		}
		p := entity.Pos{X: int(i) % g.W, Y: int(i) / g.W}
		a := g.Aspect(p)
		if a == (entity.Pos{}) {
			t.Drain = t.Height - lowest // the water leaves the map here
			continue
		}
		down := g.At(entity.Pos{X: p.X + a.X, Y: p.Y + a.Y})
		t.Drain = t.Height - down.Height + down.Drain
	}
}

// heightNode and heightQueue are a smallest-first heap of tiles by height,
// for filling hollows.
type heightNode struct {
	h   float64
	idx int32
}

type heightQueue []heightNode

func (q *heightQueue) len() int { return len(*q) }

func (q *heightQueue) push(n heightNode) {
	*q = append(*q, n)
	i := len(*q) - 1
	for i > 0 {
		p := (i - 1) / 2
		if !lower((*q)[i], (*q)[p]) {
			break
		}
		(*q)[i], (*q)[p] = (*q)[p], (*q)[i]
		i = p
	}
}

func (q *heightQueue) pop() heightNode {
	old := *q
	top := old[0]
	last := len(old) - 1
	old[0] = old[last]
	old = old[:last]
	*q = old
	i := 0
	for {
		l, best := 2*i+1, i
		if l < len(old) && lower(old[l], old[best]) {
			best = l
		}
		if r := l + 1; r < len(old) && lower(old[r], old[best]) {
			best = r
		}
		if best == i {
			break
		}
		old[i], old[best] = old[best], old[i]
		i = best
	}
	return top
}

func lower(a, b heightNode) bool {
	if a.h != b.h {
		return a.h < b.h
	}
	return a.idx < b.idx
}
