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
// The camera is that window's corner. It is kept in tiles rather than in
// screen cells because what it is about is the ground: pan east far enough
// on a globe and you come back to where you started, and that is the world's
// arithmetic and not the terminal's. See ascii.Window, which does the
// drawing, and Grid.Norm, which is the same going-round one layer down.
type camera struct {
	// x, y is the tile drawn at the top-left corner of the map area. On a
	// globe x may sit anywhere and the columns come round; y is always a row
	// of the map.
	x, y int
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

// mapArea is how much of the map the terminal can show: the panel stands
// beside it and the legend, the activity graph and its span run under it, and
// a map smaller than what is left is drawn at its own size rather than
// stretched. This is fitMap's reckoning again — see menu.go — but made every
// frame and against the map as it is rather than against the map to come.
func mapArea(m *observe.MapView, sw, sh int) (w, h int) {
	return min(m.W, sw-panelWidth), min(m.H, sh-graphHeight-3)
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
	return ascii.Window{X: c.x, Y: c.y, W: w, H: h}
}

// clamp holds the camera where there is ground to see. Rows stop at the
// poles: scrolling past the top of the map would show a band of nothing and
// there is nothing above the north pole to show. Columns stop at the edges of
// a valley for the same reason, and on a globe they do not stop at all —
// they come round, which is what a globe is.
func (c *camera) clamp(m *observe.MapView, w, h int) {
	if m.Wrap {
		c.x = ((c.x % m.W) + m.W) % m.W
	} else {
		c.x = clampInt(c.x, 0, max(0, m.W-w))
	}
	c.y = clampInt(c.y, 0, max(0, m.H-h))
}

// center puts a position in the middle of the window.
func (c *camera) center(m *observe.MapView, p entity.Pos, w, h int) {
	c.x, c.y, c.placed = p.X-w/2, p.Y-h/2, true
	c.clamp(m, w, h)
}

// pan moves the window by whole tiles and gives up following whoever was
// being followed, because the two are the same control asked for opposite
// things.
func (c *camera) pan(m *observe.MapView, dx, dy, w, h int) {
	c.x, c.y = c.x+dx, c.y+dy
	c.lock, c.placed = false, true
	c.clamp(m, w, h)
}

// step is how far one press of a pan key moves the view: a third of what is
// on the screen, so that most of what was being looked at is still there
// afterwards. Panning a tile at a time over a map a thousand round is not
// travelling, and panning a whole screen at a time loses the place.
func step(n int) int { return max(1, n/3) }
