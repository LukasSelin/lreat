package world

import (
	"math"
	"math/rand/v2"

	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
)

// Islands: the day's acting cut into the pieces of the world that cannot
// touch each other, so that each piece can act on a goroutine of its own.
//
// Everything a person does in a day happens near where they stand: they
// walk a few tiles, take from the ground under them, build beside them,
// deal with whoever is within a few tiles, and route within a window. So
// two people far enough apart cannot read or write the same ground, the
// same neighbour or the same stock today, whatever either of them does -
// and a whole party far from another party is, for the day, another world.
// That is what an island is: the peopled chunks, joined wherever two lie
// near enough that somebody on one could reach ground somebody on the
// other could reach. What the islands do, they do at the same time; what
// any of them does to the settlement as a whole - the market, the public
// order, what is known, what is on the board, what happened - is taken
// down separately and put together afterwards, island by island in a
// fixed order, which is what keeps a run the same however many goroutines
// it ran on.
//
// A world with one island - the default valley, and a globe whose people
// all live in one place - is acted on directly, in the one order it always
// was, and comes out bit for bit as before. Only a world whose people live
// apart takes the islands' order, and then the same order every time.

// IslandReach is the furthest anything a person does in a day reads or
// writes from where they stand: a route is worked out within Window of
// them, and nobody is looked for further than NearbyLimit. Every other
// radius an act uses is shorter; a test in package action holds them to it.
const IslandReach = Window + NearbyLimit

// islandApart is how many chunks apart two peopled chunks have to be for
// nobody on one to reach ground anybody on the other reaches. Two people
// in chunks d apart stand at least (d-1)*ChunkSide tiles apart, and each
// reaches IslandReach; chunks nearer than this are one island.
const islandApart = (IslandReach+ChunkSide-1)/ChunkSide + 1

// Island is one piece of the day's population: the chunks its people
// stand on and the people themselves, in agent order.
type Island struct {
	// Key is the lowest of its chunks, which names the island for the day:
	// the islands are worked in Key order, and an island's own chance is
	// the stream of its Key.
	Key    int
	Chunks []int
	Agents []*entity.Agent
}

// isles is the working memory of cutting the population into islands and
// acting on them, kept between ticks so that a day allocates nothing for it
// once the map is known.
type isles struct {
	parent  []int32
	peopled []bool
	at      []int32 // by chunk: the island it is in, or -1
	out     []Island
	views   []*isleView
	streams map[int]*rand.Rand
	// slack is the landmarks' slack as it stood when the views were made,
	// so that what each view added can be read off as a difference after
	// the others have already been put in. See Landmarks.
	slack []float64
}

// IsleCount is how the last day's acting was cut, for anybody timing a run.
type IsleCount struct {
	Islands int
	// Largest is how much of the population the biggest island held, as a
	// share: the acting can go no faster than that island alone.
	Largest float64
}

// Islands cuts today's population into islands, in Key order. It reads
// where everybody stands and nothing else, and draws no chance.
func (w *World) Islands() []Island {
	g := w.Grid
	n := len(g.Chunks)
	s := &w.isles
	if len(s.parent) != n {
		s.parent = make([]int32, n)
		s.peopled = make([]bool, n)
		s.at = make([]int32, n)
	}
	for c := range s.parent {
		s.parent[c] = int32(c)
		s.peopled[c] = false
		s.at[c] = -1
	}
	for _, a := range w.Agents {
		s.peopled[g.ChunkOf(g.Index(a.Pos))] = true
	}
	var find func(c int32) int32
	find = func(c int32) int32 {
		for s.parent[c] != c {
			s.parent[c] = s.parent[s.parent[c]]
			c = s.parent[c]
		}
		return c
	}
	union := func(a, b int32) {
		ra, rb := find(a), find(b)
		if ra == rb {
			return
		}
		if ra < rb { // the root is the lowest chunk, which is the Key
			s.parent[rb] = ra
		} else {
			s.parent[ra] = rb
		}
	}
	for c := 0; c < n; c++ {
		if !s.peopled[c] {
			continue
		}
		cx, cy := c%g.CW, c/g.CW
		for dy := -(islandApart - 1); dy <= islandApart-1; dy++ {
			y := cy + dy
			if y < 0 || y >= g.CH {
				continue
			}
			for dx := -(islandApart - 1); dx <= islandApart-1; dx++ {
				x := cx + dx
				if g.Wrap {
					x = ((x % g.CW) + g.CW) % g.CW
				} else if x < 0 || x >= g.CW {
					continue
				}
				if d := y*g.CW + x; d != c && s.peopled[d] {
					union(int32(c), int32(d))
				}
			}
		}
	}
	// One island per root, in Key order: the roots are the lowest chunks
	// of their islands and the chunks are walked in order, so the islands
	// come out in order of their keys.
	out := s.out[:0]
	for c := 0; c < n; c++ {
		if !s.peopled[c] {
			continue
		}
		r := int(find(int32(c)))
		if s.at[r] < 0 {
			s.at[r] = int32(len(out))
			if len(out) < cap(out) {
				out = out[:len(out)+1]
				out[len(out)-1].Key = r
				out[len(out)-1].Chunks = out[len(out)-1].Chunks[:0]
				out[len(out)-1].Agents = out[len(out)-1].Agents[:0]
			} else {
				out = append(out, Island{Key: r})
			}
		}
		k := s.at[r]
		s.at[c] = k
		out[k].Chunks = append(out[k].Chunks, c)
	}
	for _, a := range w.Agents {
		k := s.at[g.ChunkOf(g.Index(a.Pos))]
		out[k].Agents = append(out[k].Agents, a)
	}
	s.out = out
	return out
}

// EachIsland runs f over every island of today's population. Where there is
// one, f runs on the world itself and in the one order there always was;
// where there are several, each runs on a goroutine of its own over a view
// of the world that keeps what it does to the settlement as a whole apart
// from the others', and what the views did is put together afterwards, in
// Key order. See view and merge for what is kept apart and how it is put
// together; see Islands for why nothing else has to be.
//
// f may do anything a day's acting does: move its people, change the
// ground under them, trade, post and answer, tell what happened. It must
// not bring anybody into the world or take anybody out of it, and it must
// not unlock anything: those are the settlement's to do, one at a time,
// and a view that does them is stopped.
func (w *World) EachIsland(f func(w *World, is Island)) {
	isles := w.Islands()
	w.Isles = IsleCount{Islands: len(isles)}
	for _, is := range isles {
		w.Isles.Largest = math.Max(w.Isles.Largest, float64(len(is.Agents))/float64(len(w.Agents)))
	}
	if len(isles) <= 1 {
		for _, is := range isles {
			f(w, is)
		}
		return
	}
	// What the views must not mend for themselves is mended here first, on
	// the one goroutine: the water's labels, which routing reads.
	w.Grid.Regions()
	before, ground := *w, *w.Grid
	w.isles.slack = append(w.isles.slack[:0], w.Grid.landmarks.slack...)
	ground.landmarks.slack = w.isles.slack
	views := make([]*isleView, len(isles))
	for k, is := range isles {
		views[k] = w.view(k, is)
	}
	InParallel(len(isles), WorkersFor(len(isles)), func(k, _ int) {
		f(&views[k].w, isles[k])
	})
	for k := range isles {
		w.merge(views[k], &before, &ground)
	}
}

// isleView is one island's view of the world for a day: a copy of the
// world and of the map that shares the ground, the people and the file with
// every other view, and keeps its own of everything a day's acting changes
// about the settlement as a whole. The policy for every field of World and
// Grid is written out in islandPolicy, and a test holds the structs to it.
type isleView struct {
	w      World
	g      Grid
	router *Router
	// routers is the view's pool of routers between days, so that acting on
	// an island allocates none.
	routers  []*Router
	requests []*entity.Request
	markets  []entity.Pos
	// slack is the view's own copy of the landmarks' slack: what its people
	// build today is added up here, and put in with the rest afterwards.
	slack []float64
}

// view makes island k's view of the world for today.
func (w *World) view(k int, is Island) *isleView {
	s := &w.isles
	for len(s.views) <= k {
		s.views = append(s.views, &isleView{})
	}
	v := s.views[k]
	v.g = *w.Grid
	v.g.islanded = true
	v.slack = append(v.slack[:0], w.Grid.landmarks.slack...)
	v.g.landmarks.slack = v.slack
	if v.router == nil {
		v.router = &Router{}
	}
	v.router.g = &v.g
	v.router.Reset()
	v.router.holder = 0
	v.g.router = v.router
	for _, r := range v.routers {
		r.g = &v.g
		r.Reset()
	}
	v.w = *w
	v.w.Grid = &v.g
	v.w.RNG = w.streamFor(is)
	v.w.Log = event.NewLog(math.MaxInt / 2)
	v.w.Chronicle = nil
	v.requests = append(v.requests[:0], w.Requests...)
	v.w.Requests = v.requests
	v.markets = append(v.markets[:0], w.markets...)
	v.w.markets = v.markets
	v.w.routers = v.routers
	v.w.thoughts = nil
	return v
}

// streamFor is the chance an island draws on. The island the principal
// square stands on draws the world's own, so that the settlement's history
// goes on being drawn from the one stream it always was; any other island
// draws a stream of its own, seeded from the world and its Key, made the
// first time that Key is an island and kept.
func (w *World) streamFor(is Island) *rand.Rand {
	g := w.Grid
	market := g.ChunkOf(g.Index(g.Norm(w.MarketPos)))
	for _, c := range is.Chunks {
		if c == market {
			return w.RNG
		}
	}
	s := &w.isles
	if s.streams == nil {
		s.streams = map[int]*rand.Rand{}
	}
	r, ok := s.streams[is.Key]
	if !ok {
		r = rand.New(rand.NewPCG(w.seed, uint64(is.Key)*0x9E3779B97F4A7C15+7))
		s.streams[is.Key] = r
	}
	return r
}

// merge puts what a view did to the settlement as a whole into the world:
// what it told, in the order it told it; what it closed on the board and
// founded; how much it moved the order, the knowledge and the market; and
// whether it moved the water. before and ground are the world and the map
// as they stood when the views were made, so that what each view changed
// is read off as a difference and not as a state.
func (w *World) merge(v *isleView, before *World, ground *Grid) {
	iw, ig := &v.w, &v.g
	switch {
	case len(iw.Agents) != len(before.Agents) || iw.nextID != before.nextID:
		panic("world: an island brought somebody into the world or took somebody out; that is the settlement's to do")
	case iw.Deaths != before.Deaths:
		panic("world: an island counted a death; that is the settlement's to do")
	case len(iw.techs) != len(before.techs):
		panic("world: an island unlocked something; that is the settlement's to do")
	case iw.MarketPos != before.MarketPos:
		panic("world: an island moved the principal square; that is the settlement's to do")
	}
	for _, e := range iw.Log.All() {
		w.Log.Append(e)
		w.note(e)
	}
	for _, r := range before.Requests {
		if iw.HasRequest(r.ID) {
			continue
		}
		w.Close(r.ID)
	}
	for _, q := range iw.markets[len(before.markets):] {
		w.FoundMarket(q)
	}
	w.Safety = need.Clamp(w.Safety + (iw.Safety - before.Safety))
	w.Knowledge += iw.Knowledge - before.Knowledge
	for good := range w.Market.Stock {
		w.Market.Stock[good] += iw.Market.Stock[good] - before.Market.Stock[good]
		w.Market.Price[good] += iw.Market.Price[good] - before.Market.Price[good]
	}
	w.Grid.waters += ig.waters - ground.waters
	if ig.regionsStale {
		w.Grid.regionsStale = true
	}
	// What the island built that made its ground cheaper goes on the
	// tables' slack; islands are chunks apart, so no chunk hears from two.
	for c := range w.Grid.landmarks.slack {
		w.Grid.landmarks.slack[c] += ig.landmarks.slack[c] - ground.landmarks.slack[c]
	}
	if !ig.landmarks.ladenOK {
		w.Grid.landmarks.ladenOK = false
	}
	if ig.landmarks.stale {
		w.Grid.landmarks.stale = true
	}
	v.routers = iw.routers
}

// HasRequest reports whether a request is still on the board.
func (w *World) HasRequest(id entity.RequestID) bool {
	for _, r := range w.Requests {
		if r.ID == id {
			return true
		}
	}
	return false
}

// islandPolicy says, for every field of World and of Grid, what a view does
// with it: "shared" for what the views read and write side by side because
// islands never touch the same of it (the ground, the people, the file);
// "read" for what no day's acting changes; "own" for what a view keeps its
// own of and merge puts together; "scratch" for working memory a view has
// its own of and nothing reads back; "kept" for what a view must leave as
// it found it, which merge checks. A field of either struct not named here
// fails TestEveryFieldHasAnIslandPolicy: adding one is deciding what an
// island does with it.
var islandPolicy = map[string]string{
	// World
	"Tick": "read", "RNG": "own", "Agents": "kept", "Grid": "own",
	"MarketPos": "kept", "markets": "own", "Market": "own", "Climate": "read",
	"Safety": "own", "Knowledge": "own", "Mods": "read", "Rules": "read",
	"Log": "own", "Requests": "own", "Forest0": "read", "ReachFloor": "read",
	"Deaths": "kept", "Vitals": "read", "Chronicle": "own", "Choices": "read",
	"Entropy": "read", "watched": "read", "thoughts": "scratch", "techs": "kept",
	"pressed": "read", "nextID": "kept", "nextReqID": "read", "index": "shared",
	"Growing": "read", "swept": "read", "rates": "read", "Config": "read",
	"Stuck": "read", "Awake": "read", "routers": "own", "ways": "read",
	"ruin": "read", "seed": "read", "isles": "read", "Isles": "read",
	// Grid
	"W": "read", "H": "read", "Wrap": "read", "Tiles": "shared", "Layers": "shared",
	"CW": "read", "CH": "read", "Chunks": "shared", "PW": "read", "PH": "read",
	"patches": "shared", "Active": "read", "lenders": "shared", "steepAt": "read",
	"steepLine": "read", "woodsLine": "read", "woodsRead": "read", "holds": "read",
	"seam": "read", "seamQueue": "read", "hot": "read", "frost": "read", "sea": "read",
	"regions": "read", "regionStack": "read", "regionsStale": "own", "waters": "own",
	"fenceSeen": "read", "fenceGen": "read", "fenceBlock": "read", "fenceStack": "read",
	"fencePatches": "read", "fenceFields": "read", "router": "scratch", "islanded": "read",
	"landmarks": "own",
}

// sameGround reports whether two grids are the same map: the one map, or a
// view of it that shares its ground.
func sameGround(a, b *Grid) bool {
	return a == b || (a != nil && b != nil && len(a.Tiles) > 0 && len(a.Tiles) == len(b.Tiles) && &a.Tiles[0] == &b.Tiles[0])
}
