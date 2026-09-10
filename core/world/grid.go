package world

import "lreat/core/entity"

// Terrain is what a tile is made of.
type Terrain uint8

const (
	Grass Terrain = iota
	Forest
	Water
	Field
	Rock // an outcrop: stone to cut, nothing to grow
	// TerrainCount is how many kinds of ground there are. It sizes the
	// tables that have to carry a row for each; see kind.go.
	TerrainCount
)

// Structure is what has been built on a tile.
type Structure uint8

const (
	None Structure = iota
	House
	Market
	Road
	Granary // keeps the market's food from spoiling
	Tavern  // where people meet of an evening
)

// Tile is one cell of the world: what the ground is, what stands on it and
// whose it is, and the land itself - its height, its drainage, the rock
// under it and the soil over that. What changes on it by the day - how
// worn it is, how far what grows on it has come, what it has to give - is
// kept beside the map rather than on the tile; see Layers. Wood is the
// standing timber on a forest tile and is what gathering consumes; Wild is
// what the forest has to give in food, berries and game, and is what
// foraging and hunting consume. Fish is what a water tile has to give. All
// of them regrow, slowly, so the land pushes back against a settlement that
// takes too much and yields to one that leaves it be.
type Tile struct {
	Terrain   Terrain
	Structure Structure
	Owner     entity.ID
	Wood      float64
	Wild      float64
	Fish      float64

	// Height is metres above the lowest ground on the map, and Flow is the
	// share of the map whose water drains through this tile. Between them
	// they are the land itself: the rivers, the fertility and the going
	// underfoot are all read off these two rather than drawn on top of them.
	// See relief.go.
	Height float64
	Flow   float64
	// Drain is how far this tile stands above the water it drains into, in
	// metres. It is what makes a valley floor a water meadow and a hillside
	// dry, and it is the ground truth the soil is read from.
	Drain float64

	// Bedrock is the rock under this tile, and Sand and Clay the shares of
	// the soil over it that are one and the other, the rest being silt. The
	// rock never changes; what is made of it moves with every age of
	// weather, sorted by the water that carries it. Between them they are
	// what the ground is made of, and the fertility, the drainage and how
	// fast a hillside comes down are all read off them. See bedrock.go.
	Bedrock Bedrock
	Sand    float64
	Clay    float64

	// Plate is which piece of the crust this tile rides, and Formed the
	// epoch its rock dates from. Both are written by a world made from its
	// own history and are nothing on a world that was drawn; see history.go.
	// They are kept because what a later change wants to ask of a map -
	// where the ore is, where the ground still shakes - is a question about
	// which plate and how old, and neither can be worked out afterwards.
	Plate  uint8
	Formed uint8

	// Fenced is whether this tile lies inside a fence: a strip of a block of
	// worked ground large enough that somebody hedged it. It is not a
	// structure and not a terrain - the ground under it is still field, and
	// the fence itself is the line round the block rather than anything
	// standing on a tile. See fence.go.
	Fenced bool
}

// Buildable reports whether a tile is open ground nobody has claimed. A road
// is not buildable: once a way is laid, it stays a way.
func (t *Tile) Buildable() bool {
	return t.Terrain == Grass && t.Structure == None && t.Owner == 0
}

// Pavable reports whether a road may be laid on this tile. Roads go over open
// ground, through woods, which they clear, over outcrops, which the quarrymen
// go on cutting from underneath, and across water, where the road is a
// bridge. They do not take another building's place or run over land somebody
// has claimed.
func (t *Tile) Pavable() bool {
	return t.Structure == None && t.Owner == 0
}

// Wet reports whether this tile is water rather than ground, whatever has
// been carried over it. It is the question the map-maker asks of water nine
// times over - what will not grow trees, what silt runs off, what nobody
// stands on - and it is not the question of whether a river runs here, which
// is Flow.
func (t *Tile) Wet() bool { return t.Terrain.Wet() }

// Bridged reports whether this tile is a road carried over water.
func (t *Tile) Bridged() bool {
	return t.Structure == Road && t.Wet()
}

// Deep reports whether crossing this tile means swimming: water with nothing
// built over it. A bridge is not deep, because the walker is on the road and
// the water is underneath.
func (t *Tile) Deep() bool {
	return t.Wet() && t.Structure == None
}

// Grid is the world map, row-major. With Wrap the east edge is joined to the
// west and the map is a globe drawn as a cylinder; without it the map is a
// valley with edges. See globe.go.
type Grid struct {
	W, H  int
	Wrap  bool
	Tiles []Tile
	// Layers is the ground that changes by the day, one slice per reading
	// and indexed as Tiles is; see layers.go.
	Layers
	// Chunks is the map in pieces, CW across and CH down. See chunk.go.
	CW, CH int
	Chunks []Chunk
	// patches is the map in smaller pieces, PW across and PH down, each
	// counting what kind of ground its tiles are. It is what a search for
	// ground asks before walking, and it is the only tally of what kind of
	// ground the map holds where. See patch.go.
	PW, PH  int
	patches []patch
	// Active is which chunks are awake today, set by World.Wake. Empty
	// until the first day, when every chunk is read as awake.
	Active []bool
	// lenders is, for each tile, how many of its eight neighbours have
	// something standing on them or are somebody's: the neighbours that
	// lend a tile their wear when the case for a road on it is read. Kept
	// by Build and Claim, so that the reading can pass over the tiles that
	// have no wear of their own and nobody to lend them any. See Draw.
	lenders []uint8

	// steepAt, steepLine and woodsLine are the map's measure of its own
	// ground: what counts as steep on it, the slope above which nothing
	// wooded will hold, and how well a tile must suit trees before one will
	// take there. All are read by readWoods; woodsRead says whether they
	// have been. See woods.go.
	steepAt   float64
	steepLine float64
	woodsLine float64
	woodsRead bool
	// holds is whether trees will take on each tile, read at the same time
	// as the lines above and from the same ground. See readHolds.
	holds []bool

	// seam and seamQueue are the working memory a history's plate boundaries
	// are spread with, kept here so that an epoch allocates nothing. See
	// history.go.
	seam      []seam
	seamQueue []int32

	// hot is where a world's hotspots are: the places fed from below rather
	// than at a plate's edge. Drawn once when a history starts and fixed for
	// the life of the world. See history.go.
	hot []entity.Pos

	// frost is, for each row, the height above which the year on that row
	// never warms past Frost. It is the weather's, not the ground's, but it
	// is kept here because everything that asks it is a question about a
	// tile: it is written once, when the land is made, by the world that
	// knows what climate this map has. See Frozen and Climate.frostline.
	frost []float64

	// sea is the height of the sea, or below zero on a map with none. See
	// flood in relief.go.
	sea float64

	// regions is which laden-walkable ground each tile is part of, and
	// regionsStale whether the water has moved since it was worked out.
	// See region.go.
	regions      []int32
	regionStack  []int32
	regionsStale bool
	// waters counts the times the water has moved, so that an answer
	// about whether there is a way somewhere can be dated. See NoWay.
	waters int
	// fenceSeen, fenceBlock and fenceStack are the working memory the daily
	// walk of the fields runs on, kept here so that reading the enclosures
	// allocates nothing. See fence.go.
	fenceSeen  []bool
	fenceBlock []int32
	fenceStack []int32

	// router is the working memory the grid's own routing runs on. It serves
	// callers routing one after another; anything routing at the same time as
	// something else needs a Router of its own.
	router *Router
}

// ownRouter is the grid's router, made on first use. It is not safe to reach
// for from two goroutines at once, which is the whole reason Router can be
// held by somebody else.
//
// It comes back holding nothing. A Router keeps whose walker it is routing
// for until it is told otherwise - see Holding - and this one is picked up by
// anybody in turn, so a walker's own gates must not be left standing open for
// whoever asks next.
func (g *Grid) ownRouter() *Router {
	if g.router == nil {
		g.router = &Router{g: g}
	}
	g.router.holder = 0
	return g.router
}

// NewGrid returns an all-grass grid.
func NewGrid(w, h int) *Grid {
	g := &Grid{W: w, H: h, Tiles: make([]Tile, w*h), Layers: NewLayers(w * h), lenders: make([]uint8, w*h), sea: -1}
	g.layChunks()
	g.layPatches()
	g.repatch()
	return g
}

// In reports whether p is on the map. On a globe every column is; only a
// row past a pole is off it.
func (g *Grid) In(p entity.Pos) bool {
	if p.Y < 0 || p.Y >= g.H {
		return false
	}
	return g.Wrap || (p.X >= 0 && p.X < g.W)
}

// At returns the tile at p. The caller must check In first.
func (g *Grid) At(p entity.Pos) *Tile {
	return &g.Tiles[g.Index(p)]
}

// Clone returns a deep copy, for snapshots.
func (g *Grid) Clone() *Grid {
	c := &Grid{W: g.W, H: g.H, Wrap: g.Wrap, Tiles: make([]Tile, len(g.Tiles)), Layers: g.Layers.Copy(), sea: g.sea}
	copy(c.Tiles, g.Tiles)
	c.lenders = make([]uint8, len(g.Tiles))
	c.layChunks()
	c.layPatches()
	c.Recount()
	return c
}

// Count returns how many tiles satisfy ok.
func (g *Grid) Count(ok func(*Tile) bool) int {
	n := 0
	for i := range g.Tiles {
		if ok(&g.Tiles[i]) {
			n++
		}
	}
	return n
}

// Nearest finds the closest tile to from, within maxR steps, that satisfies
// ok. It walks square rings outward in a fixed order, so results are
// deterministic and ties resolve the same way every run.
func (g *Grid) Nearest(from entity.Pos, maxR int, ok func(p entity.Pos, t *Tile) bool) (entity.Pos, bool) {
	return g.NearestOfKind(from, maxR, 0, ok)
}

// NearestOfKind is Nearest told what ground could possibly satisfy it. It
// is for a search whose answer can only ever stand on ground of one of
// these kinds - somewhere with timber standing on it is a wood, and
// nothing else is - and it uses that twice.
//
// Once before it starts: no ground of these kinds within the radius means
// no answer within the radius, so there is nothing to walk. And then on
// every tile it would otherwise ask about: a tile whose patch holds none
// of these kinds cannot be the answer, so it is stepped over without the
// tile being read or the predicate being run. The rings are walked in the
// order they always were and the tiles that could match are asked in the
// order they always were, so the answer is the answer Nearest would have
// given. What changes is only how much ground is read to reach it.
//
// The patch a tile is in is worked out once per run of tiles that share
// one, which on a ring is fifteen tiles in sixteen.
//
// A kinds of zero means nothing is known about what could satisfy the
// search, and every tile is asked, as Nearest does. Passing kinds that
// the answer could stand *near* rather than *on* would be wrong: a bank
// is dry ground beside water, and no patch of it need hold any water.
func (g *Grid) NearestOfKind(from entity.Pos, maxR int, kinds KindSet, ok func(p entity.Pos, t *Tile) bool) (entity.Pos, bool) {
	from = g.Norm(from)
	if kinds != 0 && !g.AnyWithin(from, maxR, kinds) {
		return entity.Pos{}, false
	}
	check := func(p entity.Pos) bool { return ok(p, g.At(p)) }
	if kinds != 0 {
		was, held := -1, false
		check = func(p entity.Pos) bool {
			if i := g.patchAt(p); i != was {
				was, held = i, g.patchHolds(i, kinds)
			}
			if !held {
				return false
			}
			return ok(p, g.At(p))
		}
	}
	if g.In(from) && check(from) {
		return from, true
	}
	if g.Wrap {
		return g.nearestRound(from, maxR, check)
	}
	for r := 1; r <= maxR; r++ {
		if from.X-r < 0 && from.Y-r < 0 && from.X+r >= g.W && from.Y+r >= g.H {
			break
		}
		// The ring is clipped to the map before it is walked rather than
		// tile by tile as it is. A search that reaches to the far side of
		// the map spends most of its rings off the edge of it, and asking
		// after each of those tiles in turn was the greater part of the
		// cost of not finding anything.
		x0, x1 := max(-r, -from.X), min(r, g.W-1-from.X)
		top, bottom := from.Y-r >= 0, from.Y+r < g.H
		for dx := x0; dx <= x1; dx++ {
			if top {
				if p := (entity.Pos{X: from.X + dx, Y: from.Y - r}); check(p) {
					return p, true
				}
			}
			if bottom {
				if p := (entity.Pos{X: from.X + dx, Y: from.Y + r}); check(p) {
					return p, true
				}
			}
		}
		y0, y1 := max(-r+1, -from.Y), min(r-1, g.H-1-from.Y)
		left, right := from.X-r >= 0, from.X+r < g.W
		for dy := y0; dy <= y1; dy++ {
			if left {
				if p := (entity.Pos{X: from.X - r, Y: from.Y + dy}); check(p) {
					return p, true
				}
			}
			if right {
				if p := (entity.Pos{X: from.X + r, Y: from.Y + dy}); check(p) {
					return p, true
				}
			}
		}
	}
	return entity.Pos{}, false
}

// HasNeighbor reports whether any of the eight tiles around p satisfies ok.
func (g *Grid) HasNeighbor(p entity.Pos, ok func(*Tile) bool) bool {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			q := entity.Pos{X: p.X + dx, Y: p.Y + dy}
			if g.In(q) && ok(g.At(q)) {
				return true
			}
		}
	}
	return false
}

// Roofed reports whether a tile is a building somebody stands inside: a
// house, the market, a granary, a tavern. A road is not - a way beside a
// door is what a door is for.
func (t *Tile) Roofed() bool {
	switch t.Structure {
	case House, Market, Granary, Tavern:
		return true
	}
	return false
}

// Frozen reports whether the ground here never thaws: high enough, or far
// enough toward the pole, that the year's mean stays under Frost. It is one
// rule where there were two - the poles were bare because they were cold and
// the peaks were green because nobody had told the weather they were high -
// and it is what puts a tree line on a map with mountains on it.
//
// It is a fact about the ground and the latitude, both of which the weather
// wanders around rather than changes, so it is read off a height written down
// when the land was made. Water is not frozen ground: what a frozen sea is
// belongs to the sea, and nothing here has an answer for it yet.
func (g *Grid) Frozen(p entity.Pos) bool {
	if len(g.frost) != g.H || !g.In(p) {
		return false
	}
	t := g.At(p)
	return !t.Wet() && t.Height >= g.frost[p.Y]
}

// RoomToBuild reports whether p is open ground with open ground all round
// it: no building on any of the eight tiles that touch it. Roofs raised
// wherever there was a gap grew into one solid block with no way through
// it, which is a settlement nobody can lay a road in. Kept a tile apart,
// every house keeps its own sides clear, and the gaps between neighbours
// line up into the lanes a road is later laid along.
func (g *Grid) RoomToBuild(p entity.Pos) bool {
	return g.In(p) && g.At(p).Buildable() && !g.HasNeighbor(p, (*Tile).Roofed)
}

// Raze takes down what stands on p and gives the ground back: the tile keeps
// its terrain and loses its building and its owner, and a field goes back to
// grass. The market is the one thing that cannot come down, being the root
// of everything else. A settlement that could only ever add to itself would
// be stuck for good with every choice its founders made on ground they had
// only just arrived on, so what has been built has to be able to go.
func (g *Grid) Raze(p entity.Pos) bool {
	if !g.In(p) {
		return false
	}
	t := g.At(p)
	if t.Structure == Market {
		return false
	}
	if t.Structure == None && t.Owner == 0 {
		return false
	}
	if t.Terrain == Field {
		g.Turn(p, Grass)
		g.Age[g.Index(p)], t.Fenced = 0, false // the crop and the hedge go with the claim
	}
	g.Build(p, None)
	g.Claim(p, 0)
	return true
}

// nearestRound is Nearest on a globe. A ring's top and bottom rows run the
// whole way round once the ring is wider than the map, each column once;
// its sides are still a column each while there is a column that far off,
// and the same column when the map is exactly two rings wide. What is
// walked is walked in a fixed order, so runs repeat.
func (g *Grid) nearestRound(from entity.Pos, maxR int, check func(entity.Pos) bool) (entity.Pos, bool) {
	half := g.W / 2
	for r := 1; r <= maxR; r++ {
		top, bottom := from.Y-r >= 0, from.Y+r < g.H
		if !top && !bottom && r > half {
			break
		}
		x0, x1 := -r, r
		if 2*r+1 >= g.W {
			x0, x1 = -half, g.W-1-half
		}
		for dx := x0; dx <= x1; dx++ {
			x := g.wrapX(from.X + dx)
			if top {
				if p := (entity.Pos{X: x, Y: from.Y - r}); check(p) {
					return p, true
				}
			}
			if bottom {
				if p := (entity.Pos{X: x, Y: from.Y + r}); check(p) {
					return p, true
				}
			}
		}
		if r > half {
			continue
		}
		y0, y1 := max(-r+1, -from.Y), min(r-1, g.H-1-from.Y)
		left, right := g.wrapX(from.X-r), g.wrapX(from.X+r)
		for dy := y0; dy <= y1; dy++ {
			if p := (entity.Pos{X: left, Y: from.Y + dy}); check(p) {
				return p, true
			}
			if right != left {
				if p := (entity.Pos{X: right, Y: from.Y + dy}); check(p) {
					return p, true
				}
			}
		}
	}
	return entity.Pos{}, false
}
