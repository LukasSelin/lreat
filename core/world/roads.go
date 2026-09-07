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
// that marking a tile keeps from one tick to the next. Together they give the
// map a memory about a hundred and forty ticks long: long enough that a route
// walked daily stands out from one walked once, short enough that a way people
// have stopped using stops asking to be paved.
const (
	Wear = 1
	Fade = 0.995
)

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

// Busiest returns the most walked-on tile within radius of from that a road
// could be laid on, and how worn it is. Ties go to the tile nearest the top
// left, so that two agents reading the same ground reach for the same spot.
func (g *Grid) Busiest(from entity.Pos, radius int) (entity.Pos, float64, bool) {
	var best entity.Pos
	var worn float64
	found := false
	for y := from.Y - radius; y <= from.Y+radius; y++ {
		for x := from.X - radius; x <= from.X+radius; x++ {
			p := entity.Pos{X: x, Y: y}
			if !g.In(p) {
				continue
			}
			t := g.At(p)
			if !t.Pavable() || t.Traffic <= worn {
				continue
			}
			best, worn, found = p, t.Traffic, true
		}
	}
	return best, worn, found
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
	}
	t.Structure = Road
	t.Traffic = 0 // the ground is no longer asking for a road; it has one
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
