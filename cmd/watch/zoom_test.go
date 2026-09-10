package main

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"lreat/core/entity"
	"lreat/core/observe"
)

// Panning answers where else to look and only zooming answers what shape the
// place is. Out, a cell stands for more ground; in, it stands for less, down
// to the tile-for-a-cell map the settlement is watched on.
func TestZoomingChangesHowMuchGroundACellStandsFor(t *testing.T) {
	m := ground(512, 200, true)
	v, _ := watching(t, m, 120, 40)
	if v.cam.scale() != 1 {
		t.Fatalf("the first frame is drawn at %d tiles to the cell, want the tile-for-a-cell map", v.cam.scale())
	}
	look(v, 'z')
	if v.cam.scale() != 2 {
		t.Fatalf("z left the map at %d tiles to the cell, want twice as much ground", v.cam.scale())
	}
	look(v, 'z', 'z')
	if v.cam.scale() != 8 {
		t.Fatalf("three steps out is %d tiles to the cell, want 8", v.cam.scale())
	}
	look(v, 'Z', 'Z', 'Z')
	if v.cam.scale() != 1 {
		t.Fatalf("back in three steps is %d tiles to the cell, want 1", v.cam.scale())
	}
	look(v, 'Z')
	if v.cam.scale() != 1 {
		t.Fatalf("Z past the tile-for-a-cell map left the scale at %d", v.cam.scale())
	}
}

// What the eye is holding when the key goes down is the middle of the
// screen. A zoom that moves it is a zoom that has to be panned back
// afterwards.
func TestZoomingKeepsTheMiddleOfTheViewWhereItIs(t *testing.T) {
	m := ground(512, 200, true)
	m.Market = entity.Pos{X: 300, Y: 120}
	v, _ := watching(t, m, 120, 40)
	mid := func() entity.Pos {
		z := v.cam.scale()
		w, h := mapArea(m, 120, 40, z)
		return entity.Pos{X: v.cam.x + w*z/2, Y: v.cam.y + h*z/2}
	}
	before := mid()
	look(v, 'z')
	after := mid()
	// Within a cell of where it was: the corner is a whole block of ground
	// now and the middle cannot land between two of them.
	if d := after.X - before.X; d > v.cam.scale() || d < -v.cam.scale() {
		t.Fatalf("zooming out moved the middle of the view from %v to %v", before, after)
	}
	if d := after.Y - before.Y; d > v.cam.scale() || d < -v.cam.scale() {
		t.Fatalf("zooming out moved the middle of the view from %v to %v", before, after)
	}
}

// Zooming out stops when the whole world is on the screen. Past that the
// window is holding the map with room to spare and there is nothing further
// out to see.
func TestZoomingOutStopsAtTheWholeWorld(t *testing.T) {
	m := ground(512, 200, true)
	v, sc := watching(t, m, 120, 40)
	for i := 0; i < 10; i++ {
		look(v, 'z')
	}
	z := v.cam.scale()
	w, h := mapArea(m, 120, 40, z)
	if w*z < m.W || h*z < m.H {
		t.Fatalf("zoomed all the way out at %d tiles to the cell the window holds %dx%d tiles of %dx%d", z, w*z, h*z, m.W, m.H)
	}
	if before := z; func() int { look(v, 'z'); return v.cam.scale() }() != before {
		t.Fatalf("zooming out past the whole world went to %d tiles to the cell", v.cam.scale())
	}
	// And with the world drawn whole there is nowhere left to look, so the
	// arrows go back to the graphs they answer to everywhere else.
	if v.looking() {
		t.Fatal("the arrows are still looking around a map the screen is holding whole")
	}
	if !strings.Contains(screenText(sc), "tiles to the cell") {
		t.Fatal("a map drawn at more than a tile to the cell does not say so")
	}
}

// A valley is drawn whole at the only scale it has, and the keys that would
// change that have nowhere to go. This is every valley run there has ever
// been and the behaviour that must not move.
func TestAMapThatFitsHasNoScaleToChange(t *testing.T) {
	m := ground(40, 12, false)
	v, sc := watching(t, m, 120, 40)
	look(v, 'z', 'z')
	if v.cam.scale() != 1 {
		t.Fatalf("z drew a map that fits at %d tiles to the cell", v.cam.scale())
	}
	if text := screenText(sc); strings.Contains(text, "tiles to the cell") || strings.Contains(text, "z/Z zoom") {
		t.Fatal("a map drawn whole offers a scale to change; there is only one")
	}
}

// Zoomed out, a cell is a block of ground and a click on it picks whoever is
// in the block: what a map at that scale is asked is where the people are.
func TestClickingPicksAFigureOutOfTheBlock(t *testing.T) {
	m := ground(512, 200, true)
	m.Agents = []observe.Mark{{ID: 9, Pos: entity.Pos{X: 261, Y: 103}, Action: "farm"}}
	v, _ := watching(t, m, 120, 40)
	v.cam.z = 4
	v.cam.x, v.cam.y, v.cam.placed = 200, 40, true
	v.draw()
	// 261,103 falls in the block cornered at 260,100, which with the window
	// standing at 200,40 is cell 15,15.
	v.handleMouse(tcell.NewEventMouse(15, 15, tcell.Button1, tcell.ModNone))
	if v.sel != 9 {
		t.Fatalf("clicking the block holding the figure at 261,103 picked %d, want 9", v.sel)
	}
}

// The wheel zooms, which is the one thing a wheel does on every map anybody
// has ever used.
func TestTheWheelZooms(t *testing.T) {
	m := ground(512, 200, true)
	v, _ := watching(t, m, 120, 40)
	v.handleMouse(tcell.NewEventMouse(10, 5, tcell.WheelDown, tcell.ModNone))
	if v.cam.scale() != 2 {
		t.Fatalf("the wheel down left the map at %d tiles to the cell, want 2", v.cam.scale())
	}
	v.handleMouse(tcell.NewEventMouse(10, 5, tcell.WheelUp, tcell.ModNone))
	if v.cam.scale() != 1 {
		t.Fatalf("the wheel up left the map at %d tiles to the cell, want 1", v.cam.scale())
	}
}

// Zooming out to see the coast and having the view snatched back to a figure
// on the next tick is the view refusing the question, so it lets go of
// whoever is being followed the way panning does.
func TestZoomingLetsGoOfWhoeverIsFollowed(t *testing.T) {
	m := ground(512, 200, true)
	m.Agents = []observe.Mark{{ID: 3, Pos: entity.Pos{X: 400, Y: 150}, Action: "farm"}}
	v, _ := watching(t, m, 120, 40)
	v.sel, v.cam.lock = 3, true
	v.draw()
	look(v, 'z')
	if v.cam.lock {
		t.Fatal("zooming out did not let go of the figure being followed")
	}
}
