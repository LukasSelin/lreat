package world

import "lreat/core/entity"

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
	// Trodden is whether anybody has ever crossed this chunk. Ground
	// nobody has walked has no wear to fade and no case for a road.
	Trodden bool
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
	switch t.Terrain {
	case Field:
		c.Fields += d
	case Forest:
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
	t.Structure = s
	c.count(t, 1)
}

// Turn makes the ground at p into terrain tr. What stood or grew on it is
// the caller's to settle.
func (g *Grid) Turn(p entity.Pos, tr Terrain) {
	i := g.Index(p)
	t, c := &g.Tiles[i], &g.Chunks[g.ChunkOf(i)]
	c.count(t, -1)
	t.Terrain = tr
	c.count(t, 1)
}

// Claim makes p somebody's, or nobody's when id is zero.
func (g *Grid) Claim(p entity.Pos, id entity.ID) {
	i := g.Index(p)
	t, c := &g.Tiles[i], &g.Chunks[g.ChunkOf(i)]
	c.count(t, -1)
	t.Owner = id
	c.count(t, 1)
}

// Recount takes every chunk's counts afresh from the ground, for after the
// ground has been made or remade wholesale.
func (g *Grid) Recount() {
	for i := range g.Chunks {
		c := &g.Chunks[i]
		c.Built, c.Owned = 0, 0
		c.Houses, c.Fields, c.Forest, c.Roads, c.Granaries, c.Markets, c.Taverns = 0, 0, 0, 0, 0, 0, 0
	}
	for i := range g.Tiles {
		g.Chunks[g.ChunkOf(i)].count(&g.Tiles[i], 1)
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
