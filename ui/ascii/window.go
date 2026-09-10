package ascii

import (
	"lreat/core/entity"
	"lreat/core/observe"
	"lreat/core/world"
)

// A map used to be the same size as the screen it was drawn on, and drawing
// it was a matter of walking every tile it had. A globe is a thousand tiles
// round and no terminal is: what is drawn there is a window onto it, and the
// window moves. This is that window - which corner of the ground is being
// looked at, how much of it fits, and how much ground one cell stands for -
// and the drawing of one.
//
// It goes here rather than in the terminal viewer because the seam is a
// question about the land and not about the screen. On a globe the column
// east of the last one is the first one, and a window straddling that join
// has to read the tiles round it the way the world does or the whole edge of
// the map is drawn as a cliff that is not there.

// Window is the part of the map being drawn: the tile at its top-left
// corner, how many cells across and down, and how many tiles square the
// ground under one cell is. On a globe X may lie anywhere, before the map or
// past it, and the columns come round; Y is a row of the map and rows above
// the north pole or below the south are simply blank.
type Window struct {
	X, Y, W, H int
	// Z is the scale: the side of the block of ground one cell stands for.
	// One is the map drawn tile by tile, which is what it always was and
	// still is at the settlement. Above that a cell is a block and the
	// drawing has to choose which tile of the block to draw - see stands.
	// Zero means one, so that a window nobody has scaled is a window at the
	// scale everything here was written for.
	Z int
}

// Scale is the side of the block one cell stands for, with the unscaled
// window read as the tile-to-a-cell one it is.
func (win Window) Scale() int {
	if win.Z < 1 {
		return 1
	}
	return win.Z
}

// Tiles is how much ground the window covers, which at any scale but the
// first is more than it has cells.
func (win Window) Tiles() (w, h int) {
	z := win.Scale()
	return win.W * z, win.H * z
}

// Whole is the window that holds the entire map, which is what every caller
// wanted back when a map was small enough to have one.
func Whole(m *observe.MapView) Window { return Window{W: m.W, H: m.H} }

// gridOf reads a snapshot's tiles as a grid, so that the drawing can ask the
// land the same questions the world asks it - chiefly how steeply a tile
// falls away, which is a reading of the eight tiles round it and not of the
// tile itself, and on a globe is a reading that crosses the seam.
func gridOf(m *observe.MapView) *world.Grid {
	return &world.Grid{W: m.W, H: m.H, Wrap: m.Wrap, Tiles: m.Tiles}
}

// Tile is the position the window's cell at x, y is of - the corner of the
// block of ground it stands for, which at the first scale is the whole of
// it - and whether that cell is over the map at all. Columns come round on a
// globe; rows do not.
func (win Window) Tile(m *observe.MapView, x, y int) (entity.Pos, bool) {
	if x < 0 || x >= win.W || y < 0 || y >= win.H {
		return entity.Pos{}, false
	}
	z := win.Scale()
	p := entity.Pos{X: win.X + x*z, Y: win.Y + y*z}
	g := gridOf(m)
	if !g.In(p) {
		return entity.Pos{}, false
	}
	return g.Norm(p), true
}

// Screen is where a position falls in the window, and whether it falls in it
// at all. It is Tile read the other way round: the viewer needs it to put a
// figure where the ground under it is being drawn, and to say which figure a
// click landed on. Scaled, a block of tiles answers with the one cell they
// share, which is what a scale is.
func (win Window) Screen(m *observe.MapView, p entity.Pos) (x, y int, ok bool) {
	z := win.Scale()
	tw, th := win.Tiles()
	dy := p.Y - win.Y
	if dy < 0 || dy >= th {
		return 0, 0, false
	}
	dx := p.X - win.X
	if m.Wrap {
		// The window's corner can stand anywhere, so a tile may be reached
		// by going east or by going west; take whichever lands in the
		// window. A window as wide as the map has both, and either will do.
		dx = ((dx % m.W) + m.W) % m.W
	}
	if dx < 0 || dx >= tw {
		return 0, 0, false
	}
	return dx / z, dy / z, true
}

// blank is what is drawn where the window is off the map: over a pole, or
// past the edge of a valley. Nothing, rather than an invented coastline.
var blank = Cell{Ch: ' ', Color: Default}

// RenderWindow draws one window of the map under one reading. The settlement
// view is the ordinary map with everybody on it; a reading is the land alone,
// with the water left in because a river is how anybody finds their way
// around a map, and the market left in because it is where the settlement is.
func RenderWindow(m *observe.MapView, view View, win Window) [][]Cell {
	g := gridOf(m)
	reading := int(view) < len(Views) && Views[view].draw != nil
	z := win.Scale()
	rows := make([][]Cell, win.H)
	for j := 0; j < win.H; j++ {
		rows[j] = make([]Cell, win.W)
		for i := 0; i < win.W; i++ {
			p, on := win.Tile(m, i, j)
			if !on {
				rows[j][i] = blank
				continue
			}
			if z > 1 {
				p = stands(g, p, z, reading)
			}
			t := g.At(p)
			if reading {
				rows[j][i] = Views[view].draw(scene{g: g, p: p, t: t, b: band(g, t.Height)})
				continue
			}
			rows[j][i] = tileCell(g, p)
		}
	}
	if reading {
		if m.Market.X != 0 || m.Market.Y != 0 {
			if x, y, ok := win.Screen(m, m.Market); ok {
				rows[y][x] = Cell{Ch: 'M', Color: Market}
			}
		}
		return rows
	}
	// Agents draw over tiles; where several share a tile the first one's
	// activity sets the colour. Scaled, a cell is a block of ground and the
	// first agent anywhere in it stands for the rest, which is the question
	// a map at that scale is being asked: where the people are, not which
	// tile each of them is standing on.
	seen := make(map[[2]int]bool, len(m.Agents))
	for _, a := range m.Agents {
		x, y, ok := win.Screen(m, a.Pos)
		if !ok {
			continue
		}
		key := [2]int{x, y}
		if seen[key] {
			continue
		}
		seen[key] = true
		rows[y][x] = Cell{Ch: '@', Color: AgentColor(a.Action)}
	}
	return rows
}

// stands is the tile that stands for a block of ground z tiles square: one
// cell, and up to a few hundred tiles under it, so the drawing has to
// choose.
//
// It takes the most telling tile in the block rather than averaging the
// block, because what is worth seeing on a map zoomed out is exactly what an
// average destroys. A river is a tile wide and a settlement a dozen across;
// at eight tiles to the cell the mean of the block holding them is the
// country around them, and the river and the settlement are gone. Ground
// with nothing on it is the common case and any tile of it will do, so the
// middle one does - a sample of a field that changes slowly across a block,
// which is what height and soil and moisture are.
func stands(g *world.Grid, corner entity.Pos, z int, reading bool) entity.Pos {
	middle := entity.Pos{X: corner.X + z/2, Y: corner.Y + z/2}
	if !g.In(middle) {
		middle = corner
	}
	middle = g.Norm(middle)

	built, wet := middle, middle
	rank, wets := 0, 0
	for j := 0; j < z; j++ {
		for i := 0; i < z; i++ {
			q := entity.Pos{X: corner.X + i, Y: corner.Y + j}
			if !g.In(q) {
				continue
			}
			q = g.Norm(q)
			t := g.At(q)
			if !reading {
				if r := builtRank(t); r > rank {
					built, rank = q, r
				}
			}
			if t.Terrain == world.Water || (!reading && t.Terrain == world.Ice) {
				if wets == 0 {
					wet = q
				}
				wets++
			}
		}
	}
	switch {
	case rank > 0:
		// What people have built stands whatever else is in the block.
		// There is little of it and it is the whole reason for watching.
		return built
	case wets >= z:
		// Water enough to be a feature of the block rather than a corner of
		// one: a river crossing it is a tile wide and z tiles long, and that
		// is the amount asked for. The bar matters both ways - without one a
		// single tile of sea turns every coastal block into open water and
		// the continents come out a size too small, and with it too high the
		// rivers go, and the rivers are what a reader steers by.
		return wet
	}
	return middle
}

// builtRank is what a tile has been built on, in the order the eye should
// be given it when several of them are in the same block. A reading has no
// buildings on it - the land alone is the point of one - and the market is
// put back afterwards by the drawing, as it is at every scale.
func builtRank(t *world.Tile) int {
	switch t.Structure {
	case world.Market:
		return 6
	case world.Tavern:
		return 5
	case world.Granary:
		return 4
	case world.House:
		return 3
	case world.Road:
		return 2
	}
	return 0
}
