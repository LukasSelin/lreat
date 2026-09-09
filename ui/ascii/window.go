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
// looked at and how much of it fits - and the drawing of one.
//
// It goes here rather than in the terminal viewer because the seam is a
// question about the land and not about the screen. On a globe the column
// east of the last one is the first one, and a window straddling that join
// has to read the tiles round it the way the world does or the whole edge of
// the map is drawn as a cliff that is not there.

// Window is the part of the map being drawn: the tile at its top-left
// corner and how many tiles across and down. On a globe X may lie anywhere,
// before the map or past it, and the columns come round; Y is a row of the
// map and rows above the north pole or below the south are simply blank.
type Window struct {
	X, Y, W, H int
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

// Tile is the position the window's cell at x, y is of, and whether that
// cell is over the map at all. Columns come round on a globe; rows do not.
func (win Window) Tile(m *observe.MapView, x, y int) (entity.Pos, bool) {
	if x < 0 || x >= win.W || y < 0 || y >= win.H {
		return entity.Pos{}, false
	}
	p := entity.Pos{X: win.X + x, Y: win.Y + y}
	g := gridOf(m)
	if !g.In(p) {
		return entity.Pos{}, false
	}
	return g.Norm(p), true
}

// Screen is where a position falls in the window, and whether it falls in it
// at all. It is Tile read the other way round: the viewer needs it to put a
// figure where the ground under it is being drawn, and to say which figure a
// click landed on.
func (win Window) Screen(m *observe.MapView, p entity.Pos) (x, y int, ok bool) {
	y = p.Y - win.Y
	if y < 0 || y >= win.H {
		return 0, 0, false
	}
	x = p.X - win.X
	if m.Wrap {
		// The window's corner can stand anywhere, so a tile may be reached
		// by going east or by going west; take whichever lands in the
		// window. A window as wide as the map has both, and either will do.
		x = ((x % m.W) + m.W) % m.W
	}
	if x < 0 || x >= win.W {
		return 0, 0, false
	}
	return x, y, true
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
	rows := make([][]Cell, win.H)
	for j := 0; j < win.H; j++ {
		rows[j] = make([]Cell, win.W)
		for i := 0; i < win.W; i++ {
			p, on := win.Tile(m, i, j)
			if !on {
				rows[j][i] = blank
				continue
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
	// activity sets the colour.
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
