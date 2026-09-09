package world

import (
	"lreat/core/entity"
	"lreat/core/ontology"
)

// The map in pieces. A chunk is a square of the ground with a few counts
// kept beside it - what stands on it, what grows on it, whether anybody
// has ever walked it - so that a question about the whole map is a sum
// over the chunks rather than a walk over every tile, and so that the
// passes over the ground that every day makes can pass over the ground
// where something is happening and leave the rest. The tiles themselves
// stay in one row-major slice: a chunk is a reading of the ground, not
// where the ground is kept.
//
// The counts are kept by the writes that change them - Build, Turn, Claim
// - rather than taken afresh, because they are read within the tick they
// change in: how many granaries stand is asked after the day's building
// and before the market's food spoils, and a count a tick behind would
// spoil a different amount.

// ChunkSide is how many tiles a chunk is across and down. Every distance
// anything on the map is looked for over is shorter than this, so the
// three chunks by three around a position hold everything within reach.
const ChunkSide = 64

// Chunk is one square of the map and what is known about it.
type Chunk struct {
	X0, Y0, W, H int
	// Built and Owned are how many tiles here have something standing on
	// them, and how many are somebody's. Ground with neither is ground
	// nobody keeps and nothing falls down on.
	Built, Owned                                               int
	Houses, Fields, Forest, Roads, Granaries, Markets, Taverns int
	// Trodden is whether anybody has crossed this chunk since it was last
	// woken, and Trod the day it was last known to have been. Wear keeps
	// ground awake for a while after the last crossing - see active.go -
	// and ground nobody has walked has no wear to fade and no case for a
	// road.
	Trodden bool
	Trod    int
	// Grown is the growing weather the world had had when this chunk was
	// last passed over, and Weathered the day; see active.go.
	Grown     float64
	Weathered int
	// Height is the mean height of the ground here, in metres. It is what
	// the growing weather of a sleeping chunk is read at - the weather goes
	// by latitude and by height, and a chunk is the finest the sleeping
	// ground is reckoned by, so a chunk that is mostly mountain grows like
	// a mountain. Taken again whenever the ground is counted, because the
	// weather moves the ground. See World.Rates.
	Height float64
}

// layChunks divides the grid into chunks. Chunks along the east and south
// edges are whatever is left over.
func (g *Grid) layChunks() {
	g.CW, g.CH = (g.W+ChunkSide-1)/ChunkSide, (g.H+ChunkSide-1)/ChunkSide
	g.Chunks = make([]Chunk, g.CW*g.CH)
	for cy := 0; cy < g.CH; cy++ {
		for cx := 0; cx < g.CW; cx++ {
			c := &g.Chunks[cy*g.CW+cx]
			c.X0, c.Y0 = cx*ChunkSide, cy*ChunkSide
			c.W, c.H = min(ChunkSide, g.W-c.X0), min(ChunkSide, g.H-c.Y0)
			c.Trod = -1
		}
	}
}

// ChunkOf is the chunk the tile kept at i is in.
func (g *Grid) ChunkOf(i int) int {
	return (i/g.W/ChunkSide)*g.CW + (i%g.W)/ChunkSide
}

// ChunkAt is the chunk p is in. The caller must check In first.
func (g *Grid) ChunkAt(p entity.Pos) *Chunk {
	return &g.Chunks[g.ChunkOf(g.Index(p))]
}

// count adds d to the chunk's tally of whatever t has on it.
func (c *Chunk) count(t *Tile, d int) {
	switch t.Structure {
	case House:
		c.Houses += d
	case Road:
		c.Roads += d
	case Granary:
		c.Granaries += d
	case Market:
		c.Markets += d
	case Tavern:
		c.Taverns += d
	}
	if t.Structure != None {
		c.Built += d
	}
	switch {
	case t.Is(ontology.Field):
		c.Fields += d
	case t.Is(ontology.Wood):
		c.Forest += d
	}
	if t.Owner != 0 {
		c.Owned += d
	}
}

// Build puts s on p, or takes down what is there when s is None.
func (g *Grid) Build(p entity.Pos, s Structure) {
	i := g.Index(p)
	t, c := &g.Tiles[i], &g.Chunks[g.ChunkOf(i)]
	c.count(t, -1)
	lent, deep := t.lends(), t.Deep()
	t.Structure = s
	c.count(t, 1)
	g.relend(i, lent)
	if t.Deep() != deep {
		g.wet()
	}
}

// Turn makes the ground at p into terrain tr. What stood or grew on it is
// the caller's to settle.
func (g *Grid) Turn(p entity.Pos, tr Terrain) {
	i := g.Index(p)
	t, c := &g.Tiles[i], &g.Chunks[g.ChunkOf(i)]
	c.count(t, -1)
	deep := t.Deep()
	t.Terrain = tr
	if tr != Field {
		t.Fenced = false // a hedge stands round a field and nothing else
	}
	c.count(t, 1)
	if t.Deep() != deep {
		g.wet()
	}
}

// Claim makes p somebody's, or nobody's when id is zero.
func (g *Grid) Claim(p entity.Pos, id entity.ID) {
	i := g.Index(p)
	t, c := &g.Tiles[i], &g.Chunks[g.ChunkOf(i)]
	c.count(t, -1)
	lent := t.lends()
	t.Owner = id
	c.count(t, 1)
	g.relend(i, lent)
}

// relend settles the neighbours' lender counts after tile i has changed,
// given whether it lent before.
func (g *Grid) relend(i int, lent bool) {
	if now := g.Tiles[i].lends(); now != lent {
		if now {
			g.lend(i, 1)
		} else {
			g.lend(i, -1)
		}
	}
}

// Recount takes every chunk's counts afresh from the ground, for after the
// ground has been made or remade wholesale.
func (g *Grid) Recount() {
	for i := range g.Chunks {
		c := &g.Chunks[i]
		c.Built, c.Owned = 0, 0
		c.Houses, c.Fields, c.Forest, c.Roads, c.Granaries, c.Markets, c.Taverns = 0, 0, 0, 0, 0, 0, 0
		c.Height = 0
	}
	if len(g.lenders) != len(g.Tiles) {
		g.lenders = make([]uint8, len(g.Tiles))
	}
	clear(g.lenders)
	g.wet()
	for i := range g.Tiles {
		g.Chunks[g.ChunkOf(i)].Height += g.Tiles[i].Height
		g.Chunks[g.ChunkOf(i)].count(&g.Tiles[i], 1)
		if g.Tiles[i].lends() {
			g.lend(i, 1)
		}
	}
	for i := range g.Chunks {
		if c := &g.Chunks[i]; c.W*c.H > 0 {
			c.Height /= float64(c.W * c.H)
		}
	}
}

// Houses is how many houses stand on the map, and the rest likewise. Each
// is a sum over the chunks rather than a walk over the tiles.
func (g *Grid) Houses() int { return g.sum(func(c *Chunk) int { return c.Houses }) }

// Fields is how many tiles are under the plough.
func (g *Grid) Fields() int { return g.sum(func(c *Chunk) int { return c.Fields }) }

// Forest is how many tiles are wooded.
func (g *Grid) Forest() int { return g.sum(func(c *Chunk) int { return c.Forest }) }

// Roads is how many tiles are paved.
func (g *Grid) Roads() int { return g.sum(func(c *Chunk) int { return c.Roads }) }

// Granaries is how many granaries stand.
func (g *Grid) Granaries() int { return g.sum(func(c *Chunk) int { return c.Granaries }) }

func (g *Grid) sum(of func(*Chunk) int) int {
	n := 0
	for i := range g.Chunks {
		n += of(&g.Chunks[i])
	}
	return n
}

// lends reports whether a tile lends its wear to the tiles beside it: it
// is neither open ground nor a road, so the errands in and out of it are
// walked on the ground around it. See Draw.
func (t *Tile) lends() bool { return !t.Pavable() && t.Structure != Road }

// lend adds d to the lender count of each of i's eight neighbours.
func (g *Grid) lend(i int, d int8) {
	p := g.PosOf(i)
	for _, off := range dirs {
		q := entity.Pos{X: p.X + off.X, Y: p.Y + off.Y}
		if g.In(q) {
			g.lenders[g.Index(q)] = uint8(int8(g.lenders[g.Index(q)]) + d)
		}
	}
}

// Lenders is how many of p's neighbours lend it their wear.
func (g *Grid) Lenders(i int) int { return int(g.lenders[i]) }
