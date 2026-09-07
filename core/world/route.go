package world

import (
	"math"

	"lreat/core/entity"
)

// dirs are the eight steps out of a tile, in a fixed order. Everything that
// scans neighbours walks them in this order so that runs repeat exactly.
var dirs = [8]entity.Pos{
	{X: -1, Y: -1}, {X: 0, Y: -1}, {X: 1, Y: -1},
	{X: -1, Y: 0}, {X: 1, Y: 0},
	{X: -1, Y: 1}, {X: 0, Y: 1}, {X: 1, Y: 1},
}

// tie is how much two route costs may differ and still count as equal. Costs
// are small sums of tile costs, so anything under this is float noise.
const tie = 1e-9

// Routes is the cheapest walking cost from one tile out over the map, on the
// ground as it currently stands. It is what lets an agent use a road: a route
// is chosen by what it costs to walk, so a paved way three tiles off the
// straight line wins whenever the paving saves more than the detour spends.
//
// Only the tiles the search reached carry a cost; gen and seen say which
// those are, so the same buffers can serve search after search without being
// cleared each time.
type Routes struct {
	g    *Grid
	from entity.Pos
	gen  int32
	seen []int32
	cost []float64
	prev []int32
	rank []int8 // rank of the first step of this tile's route, for tie-breaks
}

// Router is the working memory one line of routing runs on: the frontier its
// searches build, and the result buffers a search reads its answer straight
// out of. Keeping it here rather than on the Grid means pathing every agent on
// every tick allocates nothing while still leaving the map itself something
// several goroutines can read at once, each routing on a Router of its own.
//
// A Router may not be shared. The Grid it points at may.
type Router struct {
	g        *Grid
	frontier []routeNode
	scratch  Routes
	// load is what the walker of the next route is carrying, set by Carrying
	// and spent by the search that follows it.
	load float64
}

// Router returns a router over this grid, for a caller that needs its own.
func (g *Grid) Router() *Router { return &Router{g: g} }

// offMap is a tile no search will meet, for when there is no step to prefer.
var offMap = entity.Pos{X: -1, Y: -1}

// Routes computes the cheapest way from one tile to every tile on the map,
// in a result the caller may keep.
func (r *Router) Routes(from entity.Pos) *Routes {
	return r.route(&Routes{}, from, -1, offMap)
}

// Routes computes the cheapest way from one tile to every tile on the map,
// in a result the caller may keep. It routes on the grid's own router, so it
// is for callers working one at a time.
func (g *Grid) Routes(from entity.Pos) *Routes {
	return g.ownRouter().Routes(from)
}

// minMoveCost is the cheapest a tile can be to enter. Routing to a known
// destination uses it to see how far the destination could possibly still be,
// which keeps the search from spreading over ground that cannot be on the
// way. It must never overstate the cheapest tile, or routes stop being the
// cheapest ones.
const minMoveCost = 0.5

// routeNode is one entry of the frontier. Ordering is by cost, then by the
// rank of the route's first step, then by tile, so that equal-cost routes
// resolve the same way on every run.
type routeNode struct {
	cost float64
	rank int8
	idx  int32
}

func before(a, b routeNode) bool {
	if a.cost != b.cost {
		return a.cost < b.cost
	}
	if a.rank != b.rank {
		return a.rank < b.rank
	}
	return a.idx < b.idx
}

// route runs Dijkstra out from `from`. It stops early once tile index `stop`
// is settled, if stop is not negative. `prefer`, when it is a neighbour of
// from, is the first step that wins ties, which keeps a walker on the
// straight line when going around costs exactly as much as going through.
//
// The frontier is a hand-rolled binary heap of concrete nodes rather than a
// container/heap: this runs for every agent on every tick, and boxing each
// node into an interface would cost more than the search itself.
func (r *Router) route(f *Routes, from entity.Pos, stop int32, prefer entity.Pos) *Routes {
	g := r.g
	// A load is given to one journey and does not outlive it.
	laden := r.load > SwimLoad
	r.load = 0
	n := len(g.Tiles)
	if len(f.seen) != n {
		f.seen = make([]int32, n)
		f.cost = make([]float64, n)
		f.prev = make([]int32, n)
		f.rank = make([]int8, n)
		f.gen = 0
	}
	f.g, f.from, f.gen = g, from, f.gen+1
	if !g.In(from) {
		return f
	}
	src := int32(from.Y*g.W + from.X)
	f.seen[src], f.cost[src], f.prev[src], f.rank[src] = f.gen, 0, -1, -1

	// A destination gives the search a direction: a tile is worth opening
	// only by what the route through it has cost so far plus the least the
	// rest could cost. Without one the search simply spreads outward.
	var sx, sy int
	var guided bool
	if stop >= 0 {
		sx, sy, guided = int(stop)%g.W, int(stop)/g.W, true
	}
	toGo := func(x, y int) float64 {
		if !guided {
			return 0
		}
		dx, dy := x-sx, y-sy
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}
		if dy > dx {
			dx = dy
		}
		return minMoveCost * float64(dx)
	}

	q := append(r.frontier[:0], routeNode{rank: -1, idx: src})
	for len(q) > 0 {
		top := q[0]
		last := len(q) - 1
		q[0] = q[last]
		q = q[:last]
		siftDown(q, 0)

		px, py := int(top.idx)%g.W, int(top.idx)/g.W
		here := f.cost[top.idx]
		if top.cost > here+toGo(px, py)+tie {
			continue // a cheaper route here turned up after this was queued
		}
		if top.idx == stop {
			break
		}
		for d := range dirs {
			cx, cy := px+dirs[d].X, py+dirs[d].Y
			if cx < 0 || cy < 0 || cx >= g.W || cy >= g.H {
				continue
			}
			j := int32(cy*g.W + cx)
			t := &g.Tiles[j]
			// Open water is not dear to a laden walker, it is shut. Two
			// things are still allowed through it. The end of the journey
			// itself, because somebody may wade in from the bank to fish, or
			// to stand in the river and build the bridge that ends the
			// exception. And a step out of water into water, because a walker
			// the river has risen under, or whose bridge has gone, has to be
			// able to get out of it; what is forbidden is walking in, not
			// being in.
			if laden && t.Deep() && j != stop && !g.Tiles[top.idx].Deep() {
				continue
			}
			step := moveCost[t.Terrain]
			if t.Structure != None {
				step = structureCost[t.Structure]
			}
			// The climb into the tile, which is what makes a route follow a
			// contour rather than go straight over the hill in the way.
			if d := t.Height - g.Tiles[top.idx].Height; d > 0 {
				step += Climb * d
			} else {
				step -= Descend * d
			}
			cost := here + step
			rank := top.rank
			if top.idx == src {
				rank = int8(d)
				if cx == prefer.X && cy == prefer.Y {
					rank = -1
				}
			}
			if f.seen[j] == f.gen && cost > f.cost[j]-tie && (cost > f.cost[j]+tie || rank >= f.rank[j]) {
				continue
			}
			f.seen[j], f.cost[j], f.prev[j], f.rank[j] = f.gen, cost, top.idx, rank
			q = append(q, routeNode{cost: cost + toGo(cx, cy), rank: rank, idx: j})
			siftUp(q, len(q)-1)
		}
	}
	r.frontier = q[:0]
	return f
}

func siftUp(q []routeNode, i int) {
	for i > 0 {
		p := (i - 1) / 2
		if !before(q[i], q[p]) {
			return
		}
		q[i], q[p] = q[p], q[i]
		i = p
	}
}

func siftDown(q []routeNode, i int) {
	for {
		l, best := 2*i+1, i
		if l < len(q) && before(q[l], q[best]) {
			best = l
		}
		if r := l + 1; r < len(q) && before(q[r], q[best]) {
			best = r
		}
		if best == i {
			return
		}
		q[i], q[best] = q[best], q[i]
		i = best
	}
}

// Cost is the ticks of walking from the origin to p, or +Inf if there is no
// way there.
func (f *Routes) Cost(p entity.Pos) float64 {
	if !f.g.In(p) {
		return math.Inf(1)
	}
	i := p.Y*f.g.W + p.X
	if f.seen[i] != f.gen {
		return math.Inf(1)
	}
	return f.cost[i]
}

// Step is the first tile of the cheapest route to p. It returns the origin
// itself when p is the origin or cannot be reached.
func (f *Routes) Step(p entity.Pos) entity.Pos {
	if !f.g.In(p) || p == f.from {
		return f.from
	}
	i := int32(p.Y*f.g.W + p.X)
	if f.seen[i] != f.gen || f.prev[i] < 0 {
		return f.from
	}
	src := int32(f.from.Y*f.g.W + f.from.X)
	for f.prev[i] != src {
		i = f.prev[i]
	}
	return entity.Pos{X: int(i) % f.g.W, Y: int(i) / f.g.W}
}

// Path is the cheapest route to p, origin excluded and p included. It is
// empty when p cannot be reached.
func (f *Routes) Path(p entity.Pos) []entity.Pos {
	if !f.g.In(p) || p == f.from {
		return nil
	}
	i := int32(p.Y*f.g.W + p.X)
	if f.seen[i] != f.gen || f.prev[i] < 0 {
		return nil
	}
	src := int32(f.from.Y*f.g.W + f.from.X)
	var out []entity.Pos
	for i != src {
		out = append(out, entity.Pos{X: int(i) % f.g.W, Y: int(i) / f.g.W})
		i = f.prev[i]
	}
	for l, r := 0, len(out)-1; l < r; l, r = l+1, r-1 {
		out[l], out[r] = out[r], out[l]
	}
	return out
}
