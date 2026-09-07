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
)

// Structure is what has been built on a tile.
type Structure uint8

const (
	None Structure = iota
	House
	Market
	Road
	Granary // keeps the market's food from spoiling
)

// Tile is one cell of the world. Fertility comes from the river and is worn
// down by farming; Rich is the most it can recover to. Wood is the standing
// timber on a forest tile and is what gathering consumes; Wild is what the
// forest has to give in food, berries and game, and is what foraging and
// hunting consume. Fish is what a water tile has to give. All of them
// regrow, slowly, so the land pushes back against a settlement that takes
// too much and yields to one that leaves it be.
type Tile struct {
	Terrain   Terrain
	Structure Structure
	Owner     entity.ID
	Fertility float64
	Rich      float64
	Wood      float64
	Wild      float64
	Fish      float64
}

// Buildable reports whether a tile is open ground nobody has claimed. A road
// is not buildable: once a way is laid, it stays a way.
func (t *Tile) Buildable() bool {
	return t.Terrain == Grass && t.Structure == None && t.Owner == 0
}

// Pavable reports whether a road may be laid on this tile. Roads go over open
// ground and through woods, which they clear, but they do not cross water,
// take another building's place, or run over land somebody has claimed.
func (t *Tile) Pavable() bool {
	return t.Terrain != Water && t.Structure == None && t.Owner == 0
}

// Grid is the world map, row-major.
type Grid struct {
	W, H  int
	Tiles []Tile

	// router is the working memory the grid's own routing runs on. It serves
	// callers routing one after another; anything routing at the same time as
	// something else needs a Router of its own.
	router *Router
}

// ownRouter is the grid's router, made on first use. It is not safe to reach
// for from two goroutines at once, which is the whole reason Router can be
// held by somebody else.
func (g *Grid) ownRouter() *Router {
	if g.router == nil {
		g.router = &Router{g: g}
	}
	return g.router
}

// NewGrid returns an all-grass grid.
func NewGrid(w, h int) *Grid {
	return &Grid{W: w, H: h, Tiles: make([]Tile, w*h)}
}

// In reports whether p is inside the grid.
func (g *Grid) In(p entity.Pos) bool {
	return p.X >= 0 && p.Y >= 0 && p.X < g.W && p.Y < g.H
}

// At returns the tile at p. The caller must check In first.
func (g *Grid) At(p entity.Pos) *Tile {
	return &g.Tiles[p.Y*g.W+p.X]
}

// Clone returns a deep copy, for snapshots.
func (g *Grid) Clone() *Grid {
	c := &Grid{W: g.W, H: g.H, Tiles: make([]Tile, len(g.Tiles))}
	copy(c.Tiles, g.Tiles)
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
	check := func(p entity.Pos) bool { return g.In(p) && ok(p, g.At(p)) }
	if check(from) {
		return from, true
	}
	for r := 1; r <= maxR; r++ {
		if from.X-r < 0 && from.Y-r < 0 && from.X+r >= g.W && from.Y+r >= g.H {
			break
		}
		for dx := -r; dx <= r; dx++ {
			if p := (entity.Pos{X: from.X + dx, Y: from.Y - r}); check(p) {
				return p, true
			}
			if p := (entity.Pos{X: from.X + dx, Y: from.Y + r}); check(p) {
				return p, true
			}
		}
		for dy := -r + 1; dy <= r-1; dy++ {
			if p := (entity.Pos{X: from.X - r, Y: from.Y + dy}); check(p) {
				return p, true
			}
			if p := (entity.Pos{X: from.X + r, Y: from.Y + dy}); check(p) {
				return p, true
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
