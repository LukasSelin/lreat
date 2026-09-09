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

// Relief is the fall of the lowland in metres, from the lowest ground a map
// can have to the shoulders of the valley: sixty metres over eighty tiles is
// a river valley with sides to it, enough that walking uphill is felt and
// that water knows where to go. It is what the whole map used to be, and the
// ground a settlement lives on is still made to exactly this figure.
const Relief = 60.0

// Upland is how far the high country stands above the valley it stands in,
// on a map of the default width, and uplandShare is how much of a map it
// covers. Two hundred and sixty metres is a wall rather than a slope: ground
// a route goes round because going over it costs twenty tiles of climbing,
// and that is the whole difference between a map with somewhere on it and a
// map without.
//
// It is quoted at a width and scaled to the map being made - see
// Grid.UplandRise - because a height spread over more ground is a gentler
// thing. Held at a fixed two hundred and sixty metres, a map three times as
// wide put the same mountains over three times the distance and the steep
// tenth of the ground went from a slope of 0.45 to one of 0.22: the peaks
// were still the same height and there was nothing steep anywhere, which is
// the gentle bowl this was all meant to stop being. Scaled, that tenth holds
// between 0.45 and 0.42 from eighty tiles wide to two hundred and forty, and
// a bigger map is a bigger country at the same ruggedness rather than the
// same country drawn larger.
//
// Relief is deliberately not scaled with it. The lowland is where a
// settlement lives, and what makes it liveable is measured in metres and not
// in tiles: FloodDepth says the valley floor is the ground within fourteen
// metres of its river, and the soil reads off that. Stretching the lowland to
// match a wider map would put most of it above the flood and take its soil
// down to the floor of 0.15, which is the thing that went wrong when the
// valley and the mountains were one field scaled together.
//
// The share is what keeps a map habitable, and it is a share rather than a
// height for the reason everything else here is: how much of a map comes out
// above a fixed line depends entirely on the shape of that map, and what is
// wanted is that every map has both a lowland to live in and a skyline behind
// it. See waterShare, cut the same way and for the same reason.
const (
	Upland      = 260.0
	uplandShare = 0.22
	// uplandMass is how much of the rise is the bulk of the high country and
	// how much is the ridges standing on it.
	uplandMass = 0.45
	// uplandSpan is the map width Upland is quoted at, which is the width a
	// settlement is founded on unless somebody says otherwise.
	uplandSpan = DefaultWidth
)

// Span is how many tiles across the map is at its widest. It is what the
// shape of the land is measured in: the octaves start at half of it, the high
// country is masked at half of it, and the mountains rise in proportion to it.
func (g *Grid) Span() int { return max(g.W, g.H) }

// UplandRise is how far this map's high country stands above its valley.
func (g *Grid) UplandRise() float64 { return Upland * float64(g.Span()) / uplandSpan }

// Skyline is the top of the map: the valley's own relief plus the high
// country standing on it.
func (g *Grid) Skyline() float64 { return Relief + g.UplandRise() }

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
// facing the sun, over a mixture that keeps what it is given; poor on a steep
// dry hillside, and poor on sand however well it lies. It is read off the
// drainage and the soil's own make-up rather than stored, so that when the
// ground moves the soil that the ground can carry moves with it. The map is
// made with it and every age of weather pulls the soil that is actually
// there toward it.
//
// The mixture enters as a multiplier and not as a term of its own, because
// that is what it is: a loam on a dry shoulder is still a dry shoulder, and
// the best-lying ground in the valley grows little if it is sand that will
// not hold water or clay that will not give it up. It is centred on a middling
// loam, so that a map's soils average to what they averaged before there was
// any such thing as a mixture.
func (g *Grid) SoilAt(p entity.Pos) float64 {
	t := g.At(p)
	damp := clamp01(1 - t.Drain/FloodDepth)
	steep := clamp01(g.Slope(p) / 0.25)
	lie := damp * (1 - 0.7*steep) * (0.75 + 0.5*g.Sunlight(p))
	return clamp01(0.15 + 0.85*lie*(0.75+0.5*t.Loam()))
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

// raise builds the height field. The land is made of two things, because a
// country is: the lie of it - a broad smooth swell, the valley and its
// shoulders - and the high country standing on part of that.
//
// They are made differently because they are different. The lowland is a sum
// of octaves, each half the span and half the height of the last, which is
// the shape gentle ground has: swells with no edge to them. The high ground
// is the same octaves folded at their middle - the crest of each ridge is
// where the noise crossed its own centre - which is the shape ground has
// where it was pushed up and then cut into rather than laid down: ridges with
// a line along the top, and sides that fall away from them.
//
// A single texture at a single amplitude was what this used to be, and it
// gave every seed the same gentle bowl: nine tiles in ten under a tenth of a
// slope on most maps and under a sixth on all of them, with the highest
// ground only the largest of the same lumps. There was nothing to walk round
// and nothing to look up at.
func (w *World) raise(g *Grid) {
	h := w.relief(g)
	for i := range g.Tiles {
		g.Tiles[i].Height = h[i]
	}
}

// relief is the drawn height field itself, without putting it on the map. It
// is what raise writes down, and it is also what a history is measured
// against: a made world says where its high ground is, and this says how high
// a map's ground is spread. See Grid.normalise in history.go.
func (w *World) relief(g *Grid) []float64 {
	lie := w.fold(g, false)
	crest := w.fold(g, true)

	// Where the high country stands. One lattice far coarser than anything in
	// the octaves above, so that upland is a region of the map rather than a
	// speckle through it.
	where := w.lattice(g, float64(g.Span())/2)
	rise := g.UplandRise()
	h := make([]float64, len(g.Tiles))
	foot, top := quantile(where, 1-uplandShare), quantile(where, 1)
	reach := math.Max(1e-9, top-foot)

	for i := range g.Tiles {
		// The mask is eased rather than cut, so that the mountains have feet:
		// ground just past the line rises a little and is a hill, ground at
		// the top of it rises the whole way. Cut straight, the high country
		// began at a wall with no approach to it.
		m := smooth(clamp01((where[i] - foot) / reach))
		// Part of the rise is the mass of the upland and part of it is the
		// ridges on that mass, because a mountain is both and the two do not
		// peak in the same places. Given entirely to the ridges, a seed whose
		// crests happened to fall away from its high ground came out with no
		// mountains at all - two of the first five did, and were the same
		// gentle bowl the whole thing was meant to stop being.
		h[i] = Relief*lie[i] + rise*m*(uplandMass+(1-uplandMass)*crest[i])
	}
	return h
}

// fold sums the octaves, in [0,1] before it is scaled. Folded, each octave is
// turned inside out at its middle and weighted by how high the coarser ones
// left it, which is what puts the fine detail on the flanks of the big ridges
// instead of spreading it evenly over everything: a mountain gets gullies and
// a plain stays a plain.
func (w *World) fold(g *Grid, ridged bool) []float64 {
	out := make([]float64, len(g.Tiles))
	carry := make([]float64, len(g.Tiles))
	for i := range carry {
		carry[i] = 1
	}
	// The octaves run until they are finer than a tile rather than for a
	// fixed count, so that a bigger map gets more detail rather than the same
	// detail stretched over it. On the default eighty tiles that is five of
	// them, which is what it always was.
	amp, step := 1.0, float64(g.Span())/2
	for step >= 2 {
		lattice := w.lattice(g, step)
		for i := range lattice {
			v := lattice[i]
			if ridged {
				v = 1 - math.Abs(2*v-1)
				v *= carry[i]
				carry[i] = clamp01(0.4 + 0.6*v)
			}
			out[i] += amp * v
		}
		amp, step = amp/2, step/2
	}
	// Stretched to fill [0,1], so that the top of a ridge means the top of
	// the ridge on this map rather than whatever fraction of its octaves
	// happened to agree there. Without it a crest reached a third of the
	// height it was given and every mountain came out a hill.
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, v := range out {
		lo, hi = math.Min(lo, v), math.Max(hi, v)
	}
	span := math.Max(1e-9, hi-lo)
	for i := range out {
		out[i] = (out[i] - lo) / span
	}
	return out
}

// lattice is one octave: random corners span tiles apart, smoothly blended.
func (w *World) lattice(g *Grid, span float64) []float64 {
	cols, rows := int(float64(g.W)/span)+2, int(float64(g.H)/span)+2
	// On a globe the corners go round: a whole number of them fit the
	// width, and the last blends into the first, so that the ground on one
	// side of the seam is the same ground as on the other. The spacing is
	// nudged to make them fit, by less than a corner over the whole width.
	across := span
	if g.Wrap {
		cols = max(1, int(math.Ceil(float64(g.W)/span)))
		across = float64(g.W) / float64(cols)
	}
	corner := make([]float64, cols*rows)
	for i := range corner {
		corner[i] = w.RNG.Float64()
	}
	out := make([]float64, len(g.Tiles))
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			fx, fy := float64(x)/across, float64(y)/span
			cx, cy := int(fx), int(fy)
			tx, ty := smooth(fx-float64(cx)), smooth(fy-float64(cy))
			at := func(dx, dy int) float64 {
				c := cx + dx
				if g.Wrap {
					c %= cols
				}
				return corner[(cy+dy)*cols+c]
			}
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

// Incise is how far the water has cut into the ground it has been running
// over, in metres, along the largest river a map has. It is what makes a
// valley a valley rather than a dip: raised and left alone, a river lies on
// the surface of the country like a line drawn on it, and the ground falls
// away from the water at a slope nobody can see. Cut down, the river sits at
// the bottom of something and the ground beside it is a bank.
//
// Twelve metres, and not more, because of what is beside the channel rather
// than what is in it. FloodDepth says the valley floor is the ground within
// fourteen metres of its river, which is where the soil is and where a
// settlement feeds itself. Cut deeper than that and the river's own banks
// stand above its flood plain: at thirty-four metres the soil on the gentle
// ground of all five seeds tried sat on its floor of 0.15, which is a gorge
// with nothing growing in it and not a valley.
const Incise = 12.0

// incise deepens the ways the water has already found. It runs on the first
// drainage, before the rivers are drawn, so the channels are drawn into
// ground that has been cut rather than onto ground that has not - and the
// heights are settled again afterwards, because ground that has moved drains
// differently.
//
// The cut is charged as the root of how much water crosses a tile, which is
// the usual reading and the one erode.go already takes: a gully cuts nearly
// as deep as the river it feeds, and the difference between a great river and
// a small one is far less than the difference in what they carry.
//
// It is then spread over the ground either side before it is taken off, which
// is what makes this a valley and not a trench. Applied where it was
// computed, the whole depth landed in a channel one tile wide with walls
// standing straight up out of the flood plain, and the map got steeper
// everywhere without looking like anything.
func (g *Grid) incise() {
	most := 0.0
	for i := range g.Tiles {
		most = math.Max(most, g.Tiles[i].Flow)
	}
	if most <= 0 {
		return
	}
	cut := make([]float64, len(g.Tiles))
	for i := range g.Tiles {
		// Charged by the water, and paid by the rock: the same river cuts a
		// gorge through shale and is turned aside by granite.
		cut[i] = Incise * math.Sqrt(g.Tiles[i].Flow/most) / g.Tiles[i].Hard()
	}
	for pass := 0; pass < valleyWidth; pass++ {
		cut = g.spread(cut)
	}
	for i := range g.Tiles {
		g.Tiles[i].Height -= cut[i]
	}
}

// valleyWidth is how far the cut is carried out from the channel, in passes
// of the blur below and so roughly in tiles. Three is a valley a few hundred
// metres across, with sides that can be walked up.
const valleyWidth = 3

// spread is one pass of a blur: every tile becomes the mean of itself and the
// eight around it, with the edge of the map reflecting rather than pulling
// toward nothing.
func (g *Grid) spread(v []float64) []float64 {
	out := make([]float64, len(v))
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			sum, n := 0.0, 0
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					q := entity.Pos{X: x + dx, Y: y + dy}
					if !g.In(q) {
						continue
					}
					sum, n = sum+v[g.Index(q)], n+1
				}
			}
			out[y*g.W+x] = sum / float64(n)
		}
	}
	return out
}

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
			// On a globe only the poles are an edge: water leaves the map
			// there and nowhere else.
			edge := y == 0 || y == g.H-1 || (!g.Wrap && (x == 0 || x == g.W-1))
			i := y*g.W + x
			if !edge && !g.underSea(i) {
				continue
			}
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
			j := g.Index(c)
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
		if g.Tiles[i].Wet() {
			wet[i] = g.Tiles[i].Flow >= cut/2
		} else {
			wet[i] = g.Tiles[i].Flow >= cut
		}
		if g.underSea(i) {
			wet[i] = true
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
				wet[g.Index(c)] = true
			}
		}
	}
	for i := range g.Tiles {
		t := &g.Tiles[i]
		held := t.Structure != None || t.Owner != 0
		switch {
		case wet[i] && !t.Wet() && !held:
			t.Terrain, t.Wood, t.Wild, t.Age = Water, 0, 0, 0
			t.Fish = 0.7 + 0.3*rng.Float64()
		case !wet[i] && t.Wet():
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
		if t.Wet() {
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

// The sea. A valley has none: its water leaves at the edges of the map. A
// globe has no edges but the poles, and a globe with no sea is a globe
// where every river runs to a pole and cuts the country in two from top to
// bottom, which no laden walker can get round. So a share of the lowest
// ground on a globe is put under the sea before the water finds its way
// down, and the sea is where it finds its way to.

// underSea reports whether the tile at i lies at or below sea level. A map
// with no sea has a sea level below all its ground.
func (g *Grid) underSea(i int) bool {
	return g.sea >= 0 && g.Tiles[i].Height <= g.sea
}

// flood puts the lowest share of the ground under the sea, and reads the
// sea level off the ground so that erosion can move the coast.
func (g *Grid) flood(share float64, rng interface{ Float64() float64 }) {
	g.sea = -1
	if share <= 0 {
		return
	}
	heights := make([]float64, len(g.Tiles))
	for i := range g.Tiles {
		heights[i] = g.Tiles[i].Height
	}
	g.sea = quantile(heights, share)
	for i := range g.Tiles {
		if t := &g.Tiles[i]; g.underSea(i) {
			t.Terrain, t.Wood, t.Wild, t.Age = Water, 0, 0, 0
			t.Fish = 0.7 + 0.3*rng.Float64()
		}
	}
}
