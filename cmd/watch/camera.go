package main

import (
	"lreat/core/entity"
	"lreat/core/observe"
	"lreat/ui/ascii"
)

// Until there was a globe the map was always drawn whole: the ground was
// sized to the terminal at founding and every tile of it was on the screen
// at once, so where to look was not a question anybody could ask. A globe is
// a thousand tiles round and five hundred down and no terminal is either,
// which makes the screen a window onto the world rather than the world
// itself — and the window has to be movable, or fifteen sixteenths of the
// map is ground nobody can ever see.
//
// The camera is that window's corner, and how much ground a cell of it
// stands for. Both are kept in tiles rather than in screen cells because
// what they are about is the ground: pan east far enough on a globe and you
// come back to where you started, and that is the world's arithmetic and not
// the terminal's. See ascii.Window, which does the drawing, and Grid.Norm,
// which is the same going-round one layer down.
type camera struct {
	// x, y is the tile drawn at the top-left corner of the map area. On a
	// globe x may sit anywhere and the columns come round; y is always a row
	// of the map.
	x, y int
	// z is the scale: how many tiles square the ground under one cell is.
	// Panning is how to see somewhere else and this is how to see more at
	// once — the two questions a map larger than the screen raises, and
	// panning only answers the first. A continent is four hundred tiles
	// across; at a tile to the cell it can only ever be crossed a screenful
	// at a time, and where the coast runs or where the mountains stand is
	// not a thing anybody can be told a screenful at a time.
	//
	// Zero is the tile-to-a-cell map every valley run has always had. See
	// scale, which reads it that way.
	z int
	// lock keeps the figure being followed in the window, so that picking
	// somebody out of the crowd on a map larger than the screen goes to them
	// rather than merely naming them in the panel. Panning by hand lets go:
	// somebody who has just looked somewhere deliberately does not want the
	// view snatched back on the next tick.
	lock bool
	// placed says the camera has been put somewhere. Until it has, the first
	// frame centres it on the settlement, which is the one place on a globe
	// where there is anything to watch.
	placed bool
}

// scale is how many tiles a cell stands for, with the camera nobody has
// zoomed read as the tile-to-a-cell one it is.
func (c *camera) scale() int {
	if c.z < 1 {
		return 1
	}
	return c.z
}

// cells is how many cells a run of n tiles takes at scale z, the last of
// them part full: a map of a thousand tiles at three to the cell wants three
// hundred and thirty-four, and the ground in the last one is real ground.
func cells(n, z int) int { return (n + z - 1) / z }

// mapArea is how much of the map the terminal can show: the panel stands
// beside it and the legend, the activity graph and its span run under it, and
// a map that takes fewer cells than that is drawn at its own size rather than
// stretched. This is fitMap's reckoning again — see menu.go — but made every
// frame, against the map as it is rather than against the map to come, and
// at the scale it is being drawn at: zooming out is the map wanting fewer
// cells and not the terminal offering more.
func mapArea(m *observe.MapView, sw, sh, z int) (w, h int) {
	return min(cells(m.W, z), sw-panelWidth), min(cells(m.H, z), sh-graphHeight-3)
}

// The least map worth drawing. Below this the panel and the graph have eaten
// the screen and what is left says nothing about the ground, so the view says
// so in words instead.
const (
	minMapW = 20
	minMapH = 8
)

// window is the part of the map to draw, with the camera brought back onto
// the map first.
func (c *camera) window(m *observe.MapView, w, h int) ascii.Window {
	c.clamp(m, w, h)
	return ascii.Window{X: c.x, Y: c.y, W: w, H: h, Z: c.scale()}
}

// clamp holds the camera where there is ground to see. Rows stop at the
// poles: scrolling past the top of the map would show a band of nothing and
// there is nothing above the north pole to show. Columns stop at the edges of
// a valley for the same reason, and on a globe they do not stop at all —
// they come round, which is what a globe is.
//
// What w and h are counted in is cells, and what the camera is kept in is
// tiles, so the width of the window is the one multiplied by the scale: at
// eight tiles to the cell a screen eighty cells across is looking at six
// hundred and forty tiles of ground.
func (c *camera) clamp(m *observe.MapView, w, h int) {
	z := c.scale()
	if m.Wrap {
		c.x = ((c.x % m.W) + m.W) % m.W
	} else {
		c.x = clampInt(c.x, 0, max(0, m.W-w*z))
	}
	c.y = clampInt(c.y, 0, max(0, m.H-h*z))
}

// center puts a position in the middle of the window.
func (c *camera) center(m *observe.MapView, p entity.Pos, w, h int) {
	z := c.scale()
	c.x, c.y, c.placed = p.X-w*z/2, p.Y-h*z/2, true
	c.clamp(m, w, h)
}

// pan moves the window by whole cells — a cell of ground being what the eye
// is moving over, whatever the scale says that is — and gives up following
// whoever was being followed, because the two are the same control asked for
// opposite things.
func (c *camera) pan(m *observe.MapView, dx, dy, w, h int) {
	z := c.scale()
	c.x, c.y = c.x+dx*z, c.y+dy*z
	c.lock, c.placed = false, true
	c.clamp(m, w, h)
}

// step is how far one press of a pan key moves the view: a third of what is
// on the screen, so that most of what was being looked at is still there
// afterwards. Panning a tile at a time over a map a thousand round is not
// travelling, and panning a whole screen at a time loses the place.
func step(n int) int { return max(1, n/3) }

// zooms are the scales the map is drawn at, each twice the last. Doubling is
// what makes a scale readable: a step that changes the picture by a fifth is
// a step nobody can see they have taken, and the eye can hold "twice as much
// ground" from one frame to the next.
var zooms = [...]int{1, 2, 4, 8, 16, 32}

// zoomed is the scale one step out or in from the one being used, and
// whether there is such a step. Zooming out stops when the whole map is on
// the screen: past that the window is showing the world with room to spare
// and there is nothing further out to see.
func zoomed(m *observe.MapView, sw, sh, z, by int) (int, bool) {
	at := 0
	for i, s := range zooms {
		if s == z {
			at = i
		}
	}
	next := at + by
	if next < 0 || next >= len(zooms) {
		return z, false
	}
	if by > 0 {
		// Already holding the whole world: there is no further out.
		if w, h := mapArea(m, sw, sh, z); w*z >= m.W && h*z >= m.H {
			return z, false
		}
		// And no scale so far out that the map it leaves is too small to
		// read, which is the same rule the view has for a terminal with no
		// room for a map.
		if w, h := mapArea(m, sw, sh, zooms[next]); w < minMapW || h < minMapH {
			return z, false
		}
	}
	return zooms[next], true
}

// looking is whether the arrows are for looking around at the moment: there
// is a map on the page, and more of it than the screen is holding.
//
// The arrows were the graph stepper everywhere and hjkl were the only way to
// move over a globe, which is a reach for a key nobody knows about to work
// the one control the world cannot be seen without. So the arrows are given
// to whichever of the two the page actually has: over a map with somewhere
// to look they look, and everywhere else — a valley drawn whole, a globe
// zoomed out until it is drawn whole, the vitals, the world — they step the
// graphs exactly as they always did. What holds throughout is the pair of
// keys the graphs answer to in either case: see keyed in focus.go.
func (v *view) looking() bool {
	if v.vitals || v.world || v.snap == nil || v.screen == nil {
		return false
	}
	sw, sh := v.screen.Size()
	z := v.cam.scale()
	w, h := mapArea(v.snap.Map, sw, sh, z)
	if w < minMapW || h < minMapH {
		return false
	}
	return w < cells(v.snap.Map.W, z) || h < cells(v.snap.Map.H, z)
}
