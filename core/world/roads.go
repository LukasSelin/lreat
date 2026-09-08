package world

import (
	"math"
	"sort"

	"lreat/core/entity"
	"lreat/core/event"
)

// Roads are the settlement's first piece of shared infrastructure: the only
// thing it builds that nobody lives in, farms, or sells, and that pays back
// only by being walked on. Everything here is the material — how a road is
// laid, where a sensible one runs, and how the ground remembers being walked
// on. Nothing here decides that a road ought to be laid: agents do that for
// themselves, in action.Pave, by recognising worn ground as calling for one.

// Wear is how much one crossing marks the ground, and Fade is the share of
// that marking a tile keeps from one day to the next. Together they give the
// map a memory about a hundred and forty days long - a season and a half:
// long enough that a route walked daily stands out from one walked once,
// short enough that a way people have stopped using stops asking to be
// paved.
const (
	Wear = 1
	Fade = 0.995
)

// Walked is how little wear a road may carry and still count as somebody's
// way. Below it the road is not kept - see ontology.Transforms, where what
// becomes of an unkept road is stated with everything else that falls down on
// its own - and the grass starts closing over it.
//
// It is low, at about one crossing every forty ticks. The live streets of a
// settlement sit around forty of wear and the quietest lane in it around ten,
// so this takes a way nobody comes down at all and leaves everything anybody
// still uses. What it is not is a measure of how busy a road ought to be: a
// back lane to one house is worth keeping, and does not have to earn its
// keep against the market square.
const Walked = 5

// Tread records that somebody crossed this tile.
func (g *Grid) Tread(p entity.Pos) {
	if g.In(p) {
		g.At(p).Traffic += Wear
	}
}

// Weather fades every tile's wear by one tick's worth.
func (g *Grid) Weather() {
	for i := range g.Tiles {
		if g.Tiles[i].Traffic > 0 {
			g.Tiles[i].Traffic *= Fade
		}
	}
}

// Draw is the case for laying a road on p: what people walk here, plus a
// share of what they walk on the ground beside it that could never be a
// street anyway, and nothing at all where the streets already run past.
//
// Most of the traffic a street carries is not on the street. It is on the
// houses and fields the street runs between, and those are never paved, so
// read on its own the gap beside a thronged doorway looks like empty ground.
// In a close-built settlement that leaves nowhere at all worth paving, which
// is what kept the value rule from laying a single length of road.
//
// What a doorway lends, though, it lends once. Lent whole to every gap around
// it, one busy house argued for eight streets as loudly as for one, and since
// paving a gap took nothing off what the house had to lend, the other seven
// went on arguing just as loudly afterwards. Settlements came out ringed in
// pavement, a road on every side of every house and nobody the quicker for
// seven of them. So the wear is divided among the ways out of the building,
// and a building with a street already outside it lends nothing at all: its
// errands have their road, and counting them again only buys a second one.
//
// Only ground that cannot be paved lends its wear. Open ground speaks for
// itself: were it to lend as well, every tile near a busy one would read as
// busy, and paving would come out in patches instead of the lines a road
// wants. Roads lend nothing either - traffic already on a street is already
// served, and counting it would pave the settlement outward from its first
// road until the ground ran out.
//
// Ground the streets already run past has no case at all: see Served. That
// is asked here rather than by the callers so that reading the whole map at
// once and reading one neighbourhood by hand cannot disagree about it.
//
// What comes out is not the wear but the wear weighed by what a road there
// would save, which is Saving. Two tiles walked alike do not make an equal
// case for paving if one is meadow and the other a ford.
var ProfDraw, ProfVisit int64

func (g *Grid) Draw(p entity.Pos) float64 {
	ProfDraw++
	if !g.In(p) || g.Served(p) {
		return 0
	}
	i := p.Y*g.W + p.X
	d := g.Tiles[i].Traffic
	// Away from the edge the eight neighbours are eight fixed steps along
	// the tile slice, in the same order dirs walks them, so the case for a
	// road adds up the same way without asking the map where it is eight
	// times over. This is read for every tile in sight of anybody holding
	// timber, which is often enough for that to matter.
	if p.X > 0 && p.Y > 0 && p.X < g.W-1 && p.Y < g.H-1 {
		w := g.W
		for _, o := range [8]int{-w - 1, -w, -w + 1, -1, 1, w - 1, w, w + 1} {
			t := &g.Tiles[i+o]
			if t.Pavable() || t.Structure == Road {
				continue
			}
			if ways := g.ways(entity.Pos{X: (i + o) % w, Y: (i + o) / w}); ways > 0 {
				d += t.Traffic / float64(ways)
			}
		}
		return d * g.Saving(p)
	}
	for _, off := range dirs {
		q := entity.Pos{X: p.X + off.X, Y: p.Y + off.Y}
		if !g.In(q) {
			continue
		}
		t := g.At(q)
		if t.Pavable() || t.Structure == Road {
			continue
		}
		if ways := g.ways(q); ways > 0 {
			d += t.Traffic / float64(ways)
		}
	}
	return d * g.Saving(p)
}

// ways is how many gaps a building's traffic could leave by, and 0 once one
// of them is a street. It is the divisor in Draw: what the errands in and out
// of one door are worth is shared among the ground they could be walked on,
// and spent entirely once any of that ground is paved.
func (g *Grid) ways(p entity.Pos) int {
	n := 0
	for _, off := range dirs {
		q := entity.Pos{X: p.X + off.X, Y: p.Y + off.Y}
		if !g.In(q) {
			continue
		}
		switch t := g.At(q); {
		case t.Structure == Road:
			return 0
		case t.Pavable():
			n++
		}
	}
	return n
}

// Served reports whether p is already on the street: whether the ways it
// touches run past it anyway, so that paving it would widen what is there
// rather than carry it anywhere new.
//
// This is the difference between a road and a paved field. Ground beside a
// street is busy precisely because the street is there - a road is the
// cheapest going on the map, so every route that can bends onto it, and the
// tiles alongside carry the traffic that funnels on and off. Read as bare
// wear that is a standing case for paving the tile next to a road, and then
// the tile next to that, for as long as anybody keeps walking. Left alone it
// is what turns a settlement's streets into its ground: over twenty thousand
// ticks the settlements grew a thousand tiles of road around a single house,
// nine in ten of them touching no building at all.
//
// A tile with no road beside it is unserved: there is no street here yet. So
// is a tile beside a single road, or beside two stubs that do not otherwise
// meet - the first carries a way onward, the second joins two ways into one,
// and both leave the settlement somewhere it could not go before. What is
// refused is only the tile whose roads already reach each other without it.
func (g *Grid) Served(p entity.Pos) bool {
	// Kept in an array rather than a slice: this is asked of every tile the
	// settlement could pave, on every tick anybody thinks about paving.
	var near [8]entity.Pos
	n := 0
	for _, off := range dirs {
		q := entity.Pos{X: p.X + off.X, Y: p.Y + off.Y}
		if g.In(q) && g.At(q).Structure == Road {
			near[n] = q
			n++
		}
	}
	if n < 2 {
		return false // nothing to widen: no street here, or an end to carry on
	}
	// Are they all one way already? Walk the roads that touch p, stepping
	// only between those that touch each other, and see whether that reaches
	// them all. If it does not, p is where two ways would be joined.
	var seen [8]bool
	seen[0] = true
	reached := 1
	for grew := true; grew; {
		grew = false
		for i := 0; i < n; i++ {
			if !seen[i] {
				continue
			}
			for j := 0; j < n; j++ {
				if seen[j] || entity.Dist(near[i], near[j]) > 1 {
					continue
				}
				seen[j], reached, grew = true, reached+1, true
			}
		}
	}
	return reached == n
}

// Busiest returns the tile within radius of from where a road would serve the
// most traffic, and how strong the case for it is. Only ground a road could
// actually be laid on is offered, and only what ok admits - a caller short of
// timber can rule out the water, which costs more to carry a road over. The
// case is read from the whole neighbourhood: see Draw, and ground the streets
// already run past is passed over: see Served. Ties go to the tile nearest the
// top left, so two agents reading the same ground reach for the same spot.
func (g *Grid) Busiest(from entity.Pos, radius int, ok func(*Tile) bool) (entity.Pos, float64, bool) {
	var best entity.Pos
	var worn float64
	found := false
	y0, y1 := max(0, from.Y-radius), min(g.H-1, from.Y+radius)
	x0, x1 := max(0, from.X-radius), min(g.W-1, from.X+radius)
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			t := &g.Tiles[y*g.W+x]
			if !t.Pavable() || (ok != nil && !ok(t)) {
				continue
			}
			p := entity.Pos{X: x, Y: y}
			if d := g.Draw(p); d > worn {
				best, worn, found = p, d, true
			}
		}
	}
	return best, worn, found
}

// Pick is where a walk over the ground would lay a road, and how strong the
// case for laying it there is. Found is false when the walk was offered
// nothing at all - either no ground it could be laid on, or none that anybody
// has ever walked.
type Pick struct {
	Pos   entity.Pos
	Worn  float64
	Found bool
}

// Ways is the case for a road, read off the whole map at once.
//
// Everybody who thinks about roads on a tick asks the same question of the
// same ground - what near me is most walked on, and could be paved - and the
// ground does not change while they are asking, because deciding only reads
// the world. So the case for each tile is worked out once for the whole
// settlement, and what is left to each of them is a look over its own
// neighbourhood. Nine people thinking about roads on a tick read the ground
// between them twice over rather than nine times.
type Ways struct {
	g *Grid
	// stamp is the tick this was read plus one, so that a reading nobody has
	// taken is never mistaken for one taken at the first tick.
	stamp int
	// draw is the case for a road on each tile, and nothing on ground no road
	// could be laid on. See Draw.
	draw []float64
}

// readWays takes the reading, into the buffer of the last one where it fits.
// tick is the tick it is a reading of.
func (g *Grid) readWays(y *Ways, tick int) *Ways {
	if y == nil {
		y = &Ways{}
	}
	if len(y.draw) != len(g.Tiles) {
		y.draw = make([]float64, len(g.Tiles))
	}
	y.g, y.stamp = g, tick+1
	for i := range g.Tiles {
		y.draw[i] = 0
		if g.Tiles[i].Pavable() {
			y.draw[i] = g.Draw(entity.Pos{X: i % g.W, Y: i / g.W})
		}
	}
	return y
}

// Busiest is the busiest ground within radius of from that a road could be
// laid on: the best of the dry ground and the best of the water, each with
// the case for it. They are kept apart because a road over water is a bridge
// and costs more timber, so somebody may be able to afford the one and not
// the other.
//
// Ground nobody has walked is passed over: its case is nothing, and nothing
// never wins. Ties go to the tile earliest in row-major order, which is the
// one somebody walking the neighbourhood would have come to first.
func (y *Ways) Busiest(from entity.Pos, radius int) (dry, wet Pick) {
	g := y.g
	x0, x1 := max(0, from.X-radius), min(g.W-1, from.X+radius)
	for row := max(0, from.Y-radius); row <= min(g.H-1, from.Y+radius); row++ {
		base := row * g.W
		for x := x0; x <= x1; x++ {
			d := y.draw[base+x]
			if d <= 0 {
				continue
			}
			if g.Tiles[base+x].Terrain == Water {
				if d > wet.Worn {
					wet = Pick{Pos: entity.Pos{X: x, Y: row}, Worn: d, Found: true}
				}
			} else if d > dry.Worn {
				dry = Pick{Pos: entity.Pos{X: x, Y: row}, Worn: d, Found: true}
			}
		}
	}
	return dry, wet
}

// Pave lays a road on one tile and reports whether it took. Woods in the way
// are cleared, since a road through a forest is a road, not a forest.
func (g *Grid) Pave(p entity.Pos) bool {
	if !g.In(p) {
		return false
	}
	t := g.At(p)
	if !t.Pavable() {
		return false
	}
	if t.Terrain == Forest {
		t.Terrain, t.Wood = Grass, 0
		t.Sow()
	}
	// Over water the road is a bridge, so the water stays: the fish go on
	// swimming under it and the tile is still a river to look at.
	t.Structure = Road
	// The wear stays. On open ground it was the ground asking for a road,
	// and nothing reads it that way any more - a road is not Pavable, and
	// lends nothing to its neighbours, so it can no longer ask for anything.
	// What it is now is the road's keep: the feet that made the case for the
	// way are the same feet that hold the grass off it, and a road starts
	// life with a good deal of that credit behind it. See Walked.
	return true
}

// PaveRoute lays a road along the cheapest walking route between two tiles
// and returns how many tiles it laid. The ends are left as they are, so a
// route may be run between two houses, or from a house to the market, without
// paving over either. Because the route is the one a walker would take, the
// road ends up where the traffic already is: it follows the ground rather
// than fighting it, rounding thickets and keeping off the water.
func (w *World) PaveRoute(from, to entity.Pos) int {
	laid := 0
	for _, p := range w.Grid.Routes(from).Path(to) {
		if p == to {
			break
		}
		if w.Grid.Pave(p) {
			laid++
		}
	}
	return laid
}

// PaveStreets extends the settlement's road network to reach every house.
// The market is the network's root; each house is joined to whichever part of
// the network it can reach most cheaply, nearest houses first, so the streets
// grow outward from the middle instead of each house running its own long
// track to the centre. Running it again connects whatever has been built
// since and leaves the existing streets alone. It returns the tiles laid.
func (w *World) PaveStreets() int {
	network := []entity.Pos{}
	var houses []entity.Pos
	for i := range w.Grid.Tiles {
		p := entity.Pos{X: i % w.Grid.W, Y: i / w.Grid.W}
		switch w.Grid.Tiles[i].Structure {
		case Market, Road:
			network = append(network, p)
		case House:
			houses = append(houses, p)
		}
	}
	if len(network) == 0 || len(houses) == 0 {
		return 0
	}

	// Nearest first, by the walk to the market rather than by the crow's
	// flight, and ties broken by position so that runs repeat.
	fromMarket := w.Grid.Routes(w.MarketPos)
	sort.Slice(houses, func(i, j int) bool {
		ci, cj := fromMarket.Cost(houses[i]), fromMarket.Cost(houses[j])
		if ci != cj {
			return ci < cj
		}
		if houses[i].Y != houses[j].Y {
			return houses[i].Y < houses[j].Y
		}
		return houses[i].X < houses[j].X
	})

	laid := 0
	for _, h := range houses {
		f := w.Grid.Routes(h)
		best, bestCost := entity.Pos{}, math.Inf(1)
		for _, n := range network {
			if c := f.Cost(n); c < bestCost {
				best, bestCost = n, c
			}
		}
		if math.IsInf(bestCost, 1) {
			continue
		}
		for _, p := range f.Path(best) {
			if p == best {
				break
			}
			if w.Grid.Pave(p) {
				network = append(network, p)
				laid++
			}
		}
	}
	if laid > 0 {
		w.Emit(event.Built, 0, 0, "%d tiles of road were laid through the settlement", laid)
	}
	return laid
}
