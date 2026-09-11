package main

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"lreat/core/entity"
	"lreat/core/observe"
	"lreat/core/world"
)

// ground is a map of the given shape with nobody on it: enough for the
// window to be asked where it is looking, and made in a moment, which a
// world a thousand tiles round is not.
func ground(w, h int, wrap bool) *observe.MapView {
	return &observe.MapView{W: w, H: h, Wrap: wrap, Tiles: make([]world.Tile, w*h), Layers: world.NewLayers(w * h)}
}

// watching is a view of a map, on a screen of the given size, with the
// camera already put where the first frame would put it.
func watching(t *testing.T, m *observe.MapView, sw, sh int) (*view, tcell.SimulationScreen) {
	t.Helper()
	sc := tcell.NewSimulationScreen("UTF-8")
	if err := sc.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(sc.Fini)
	sc.SetSize(sw, sh)
	v := &view{screen: sc, snap: &observe.Snapshot{Map: m}}
	v.draw()
	return v, sc
}

// look presses the keys that move the window.
func look(v *view, keys ...rune) {
	for _, r := range keys {
		v.handleKey(nil, tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
	}
}

// A map the terminal holds whole has nowhere else to look, and the keys that
// would look there do nothing. This is the behaviour every valley run has
// always had and the one that must not change.
func TestAMapThatFitsHasNowhereToLook(t *testing.T) {
	m := ground(40, 12, false)
	v, sc := watching(t, m, 120, 40)
	look(v, 'l', 'l', 'j', 'j')
	if v.cam.x != 0 || v.cam.y != 0 {
		t.Fatalf("the window moved to %d,%d on a map that fits the screen", v.cam.x, v.cam.y)
	}
	if strings.Contains(screenText(sc), "looking at") {
		t.Fatal("a map drawn whole says where it is looking; there is only one answer")
	}
}

// On a map larger than the screen the keys move the window, and the view
// says where it has got to: on a globe every view looks alike, and a reader
// who cannot say which corner of the world is on the screen cannot come back
// to it either.
func TestLookingAroundMovesTheWindow(t *testing.T) {
	m := ground(400, 200, false)
	v, sc := watching(t, m, 120, 40)
	at := v.cam.x
	look(v, 'l')
	if v.cam.x <= at {
		t.Fatalf("l left the window at %d, want it further east than %d", v.cam.x, at)
	}
	east := v.cam.x
	look(v, 'h')
	if v.cam.x != at {
		t.Fatalf("h from %d landed at %d, want back at %d", east, v.cam.x, at)
	}
	down := v.cam.y
	look(v, 'j')
	if v.cam.y <= down {
		t.Fatalf("j left the window at row %d, want it further south than %d", v.cam.y, down)
	}
	if !strings.Contains(screenText(sc), "looking at") {
		t.Fatal("a map larger than the screen does not say which part of it is on the screen")
	}
}

// A valley has edges and the view stops at them: there is nothing past the
// last column to show.
func TestTheViewStopsAtAValleysEdge(t *testing.T) {
	m := ground(400, 200, false)
	v, _ := watching(t, m, 120, 40)
	v.cam.x, v.cam.y = 0, 0
	look(v, 'h', 'k')
	if v.cam.x != 0 || v.cam.y != 0 {
		t.Fatalf("the window went off the top-left of a valley, to %d,%d", v.cam.x, v.cam.y)
	}
	for i := 0; i < 40; i++ {
		look(v, 'l')
	}
	mw, _ := mapArea(m, 120, 40, v.cam.scale())
	if v.cam.x != m.W-mw {
		t.Fatalf("the window stopped at %d going east, want the last screenful at %d", v.cam.x, m.W-mw)
	}
}

// A globe has no east edge to stop at. Looking west from the seam comes out
// on the far side of the map, which is the same going-round the walkers on it
// do. See world.Grid.Norm.
func TestTheViewComesRoundAGlobe(t *testing.T) {
	m := ground(512, 200, true)
	v, _ := watching(t, m, 120, 40)
	v.cam.x = 0
	look(v, 'h')
	if v.cam.x <= m.W/2 {
		t.Fatalf("looking west from the seam landed at %d, want it round on the far side of the map", v.cam.x)
	}
	v.cam.x = m.W - 1
	look(v, 'l')
	if v.cam.x >= m.W-1 {
		t.Fatalf("looking east over the seam landed at %d, want it round at the start", v.cam.x)
	}
}

// The rows stop at the poles even on a globe: the map is a cylinder and not
// a sphere, and there is nothing above the top row.
func TestTheViewStopsAtThePoles(t *testing.T) {
	m := ground(512, 200, true)
	v, _ := watching(t, m, 120, 40)
	v.cam.y = 0
	look(v, 'k', 'k')
	if v.cam.y != 0 {
		t.Fatalf("the window went above the north pole, to row %d", v.cam.y)
	}
}

// C is the way back. On a globe the settlement is a fraction of a per cent of
// the map and the only part of it anybody is watching; without a way back,
// one pan too far east is a run abandoned.
func TestTheSettlementIsAlwaysOnePressAway(t *testing.T) {
	m := ground(512, 200, true)
	m.Market = entity.Pos{X: 300, Y: 120}
	v, _ := watching(t, m, 120, 40)
	look(v, 'l', 'l', 'j', 'j', 'c')
	mw, mh := mapArea(m, 120, 40, v.cam.scale())
	if _, _, ok := v.cam.window(m, mw, mh).Screen(m, m.Market); !ok {
		t.Fatalf("c left the settlement off the screen: the window is at %d,%d", v.cam.x, v.cam.y)
	}
}

// Picking a figure out of the crowd is no use on a map larger than the
// screen if the figure is a hundred tiles away, so the window can be told to
// keep whoever is followed on it. Looking around by hand lets go again: a
// view snatched back on the next tick is not a view anybody can use.
func TestFollowingKeepsTheFigureOnTheScreen(t *testing.T) {
	m := ground(512, 200, true)
	m.Agents = []observe.Mark{{ID: 3, Pos: entity.Pos{X: 400, Y: 150}, Action: "farm"}}
	v, _ := watching(t, m, 120, 40)
	v.sel = 3
	v.cam.lock = true
	v.draw()
	mw, mh := mapArea(m, 120, 40, v.cam.scale())
	if _, _, ok := v.cam.window(m, mw, mh).Screen(m, m.Agents[0].Pos); !ok {
		t.Fatal("the figure being followed is off the screen")
	}
	look(v, 'h')
	if v.cam.lock {
		t.Fatal("looking around by hand did not let go of the figure being followed")
	}
	look(v, 'f')
	if !v.cam.lock {
		t.Fatal("f did not take the figure up again")
	}
}

// A click is in screen cells and the figures are on the ground. On a globe
// the cell in the corner of the screen is not tile 0,0 and need not even be
// east of it, so the click goes back through the window before anybody is
// picked out by it.
func TestClickingPicksAFigureThroughTheWindow(t *testing.T) {
	m := ground(512, 200, true)
	m.Agents = []observe.Mark{{ID: 9, Pos: entity.Pos{X: 260, Y: 100}, Action: "farm"}}
	v, _ := watching(t, m, 120, 40)
	v.cam.x, v.cam.y, v.cam.placed = 250, 95, true
	v.draw()
	v.handleMouse(tcell.NewEventMouse(10, 5, tcell.Button1, tcell.ModNone))
	if v.sel != 9 {
		t.Fatalf("clicking the figure at 260,100 picked %d, want 9", v.sel)
	}
	v.sel = 0
	v.handleMouse(tcell.NewEventMouse(0, 0, tcell.Button1, tcell.ModNone))
	if v.sel != 0 {
		t.Fatal("clicking empty ground picked somebody")
	}
}

// The first frame looks at the settlement rather than at the corner of the
// world, which on a globe is a thousand tiles of sea away from anybody.
func TestTheFirstFrameLooksAtTheSettlement(t *testing.T) {
	m := ground(512, 200, true)
	m.Market = entity.Pos{X: 88, Y: 140}
	v, _ := watching(t, m, 120, 40)
	mw, mh := mapArea(m, 120, 40, v.cam.scale())
	if _, _, ok := v.cam.window(m, mw, mh).Screen(m, m.Market); !ok {
		t.Fatalf("the first frame is looking at %d,%d, with the settlement off the screen", v.cam.x, v.cam.y)
	}
}

// A window with no room for a map says so, rather than drawing a strip of
// ground that says nothing. What it will not do any more is refuse a map
// merely for being larger than the screen.
func TestATerminalWithNoRoomForAMapSaysSo(t *testing.T) {
	_, sc := watching(t, ground(400, 200, false), panelWidth+minMapW-1, 40)
	if !strings.Contains(screenText(sc), "terminal too small") {
		t.Fatal("a terminal with no room for a map drew one anyway")
	}
	_, wide := watching(t, ground(400, 200, false), 120, 40)
	if strings.Contains(screenText(wide), "terminal too small") {
		t.Fatal("a map larger than the screen was refused instead of being looked at through a window")
	}
}

// The whole way through, on real ground: a wrapped world, the snapshot taken
// off it, and the window drawn from that. What this catches that the tests
// above cannot is the shape of the world going missing somewhere between the
// grid and the glyphs — a globe drawn as a valley looks perfectly reasonable
// until the seam is on the screen.
func TestAWrappedWorldIsWatchedThroughAWindow(t *testing.T) {
	w := world.NewWith(3, world.Config{Width: 128, Height: 64, Wrap: true, SeaShare: 0.3, Settlements: 1})
	for i := 0; i < 5; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	s := observe.Take(w)
	if !s.Map.Wrap {
		t.Fatal("the snapshot of a globe does not say the world wraps")
	}
	sc := tcell.NewSimulationScreen("UTF-8")
	if err := sc.Init(); err != nil {
		t.Fatal(err)
	}
	defer sc.Fini()
	sc.SetSize(100, 30)
	v := &view{screen: sc, snap: &s}
	v.draw()
	if text := screenText(sc); !strings.Contains(text, "looking at") {
		t.Fatalf("a world larger than the screen was drawn without saying where:\n%s", text)
	}
	// Straddle the seam and draw again: the column east of the last one is
	// the first one, and there is ground in every cell of the window.
	mw, mh := mapArea(s.Map, 100, 30, v.cam.scale())
	v.cam.x, v.cam.y, v.cam.placed = s.Map.W-mw/2, 10, true
	v.draw()
	cells, width, _ := sc.GetContents()
	for y := 0; y < mh; y++ {
		for x := 0; x < mw; x++ {
			if r := cells[y*width+x].Runes[0]; r == ' ' || r == 0 {
				t.Fatalf("the window across the seam has a hole in it at %d,%d", x, y)
			}
		}
	}
}

// arrows presses arrow and page keys.
func arrows(v *view, keys ...tcell.Key) {
	for _, k := range keys {
		v.handleKey(nil, tcell.NewEventKey(k, 0, tcell.ModNone))
	}
}

// The arrows look around a map larger than the screen. They are the keys a
// reader reaches for without being told, and reaching for hjkl instead is
// reaching for a key nobody knows about to work the one control a globe
// cannot be seen without.
func TestTheArrowsLookAroundABigMap(t *testing.T) {
	m := ground(512, 200, true)
	v, sc := watching(t, m, 120, 40)
	at := v.cam.x
	arrows(v, tcell.KeyRight)
	if v.cam.x <= at {
		t.Fatalf("right left the window at %d, want it further east than %d", v.cam.x, at)
	}
	arrows(v, tcell.KeyLeft)
	if v.cam.x != at {
		t.Fatalf("left landed at %d, want back at %d", v.cam.x, at)
	}
	down := v.cam.y
	arrows(v, tcell.KeyDown)
	if v.cam.y <= down {
		t.Fatalf("down left the window at row %d, want it further south than %d", v.cam.y, down)
	}
	if v.focus != 0 {
		t.Fatal("an arrow opened a graph out on a map that was being looked around")
	}
	// The graphs are still reachable, by the keys the panel names.
	arrows(v, tcell.KeyPgDn)
	if v.focus == 0 {
		t.Fatal("pgdn did not step onto a graph")
	}
	if !strings.Contains(screenText(sc), "pgup/pgdn") {
		t.Fatal("the panel does not say which keys the graphs answer to")
	}
}

// On a map drawn whole the arrows are what they always were. This is every
// valley run there has ever been and the behaviour that must not move.
func TestTheArrowsStillStepTheGraphsOnAMapThatFits(t *testing.T) {
	m := ground(40, 12, false)
	v, sc := watching(t, m, 120, 40)
	arrows(v, tcell.KeyDown)
	if v.focus == 0 {
		t.Fatal("down did not open a graph out on a map the screen holds whole")
	}
	if v.cam.x != 0 || v.cam.y != 0 {
		t.Fatalf("an arrow moved the window to %d,%d on a map that fits", v.cam.x, v.cam.y)
	}
	arrows(v, tcell.KeyUp)
	if v.focus != 0 {
		t.Fatal("up did not step back off the graph")
	}
	if !strings.Contains(screenText(sc), "up/down") {
		t.Fatal("the panel does not say the arrows step the graphs")
	}
}

// The pages of graphs have no map to look at, so the arrows step graphs
// there whatever the world is.
func TestTheArrowsStepTheGraphsOnThePages(t *testing.T) {
	for _, page := range []rune{'d', 'w'} {
		m := ground(512, 200, true)
		v, _ := watching(t, m, 120, 40)
		look(v, page)
		at := v.cam.x
		arrows(v, tcell.KeyDown)
		if v.focus == 0 {
			t.Fatalf("down opened no graph on the %c page", page)
		}
		if v.cam.x != at {
			t.Fatalf("down moved the map window while the %c page was up", page)
		}
	}
}
