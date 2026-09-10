package ascii

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/observe"
	"lreat/core/world"
)

// globeView is a small wrapped world's map: two chunks round and one down,
// which is a globe in every way that matters here and is generated in a
// moment rather than in a minute.
func globeView(t *testing.T) *observe.MapView {
	t.Helper()
	w := world.NewWith(7, world.Config{Width: 128, Height: 64, Wrap: true, SeaShare: 0.3, Settlements: 1})
	w.Spawn("a", need.Neutral())
	s := observe.Take(w)
	if !s.Map.Wrap {
		t.Fatal("a globe's map does not say it wraps")
	}
	return s.Map
}

// A window is the part of the map it names and nothing else: what it draws
// is what the whole map has in the same place.
func TestAWindowIsThePartOfTheMapItNames(t *testing.T) {
	m := globeView(t)
	whole := RenderView(m, Height)
	win := Window{X: 20, Y: 8, W: 16, H: 9}
	part := RenderWindow(m, Height, win)
	if len(part) != win.H || len(part[0]) != win.W {
		t.Fatalf("window drew %dx%d, want %dx%d", len(part[0]), len(part), win.W, win.H)
	}
	for y := 0; y < win.H; y++ {
		for x := 0; x < win.W; x++ {
			if got, want := part[y][x], whole[win.Y+y][win.X+x]; got != want {
				t.Fatalf("window at %d,%d is %v, want %v", x, y, got, want)
			}
		}
	}
}

// A window straddling the seam reads round it. This is the whole reason the
// window knows what shape the world is: drawn without it, the last column of
// a globe is an edge, and a globe has none.
func TestAWindowReadsRoundTheSeam(t *testing.T) {
	m := globeView(t)
	whole := RenderView(m, Height)
	win := Window{X: m.W - 4, Y: 10, W: 8, H: 3}
	part := RenderWindow(m, Height, win)
	for y := 0; y < win.H; y++ {
		for x := 0; x < win.W; x++ {
			wx := (win.X + x) % m.W
			if got, want := part[y][x], whole[win.Y+y][wx]; got != want {
				t.Fatalf("across the seam at %d,%d is %v, want the map's own %v", x, y, got, want)
			}
		}
	}
}

// Panning east all the way round comes back to where it started, which is
// what being a cylinder means.
func TestPanningRoundAGlobeComesBack(t *testing.T) {
	m := globeView(t)
	at := func(x int) [][]Cell { return RenderWindow(m, Height, Window{X: x, Y: 4, W: 6, H: 4}) }
	home, round := at(30), at(30+m.W)
	for y := range home {
		for x := range home[y] {
			if home[y][x] != round[y][x] {
				t.Fatal("a window a whole map east is not the window it started at")
			}
		}
	}
}

// A window over a pole, or off the edge of a valley, draws nothing there
// rather than inventing ground.
func TestAWindowOffTheMapIsBlank(t *testing.T) {
	m := globeView(t)
	rows := RenderWindow(m, Height, Window{X: 0, Y: -2, W: 4, H: 4})
	for y := 0; y < 2; y++ {
		for x := 0; x < 4; x++ {
			if rows[y][x] != blank {
				t.Fatalf("row %d above the north pole drew %v, want nothing", y, rows[y][x])
			}
		}
	}
	if rows[2][0] == blank {
		t.Fatal("the first row of the map drew nothing")
	}

	valley := &observe.MapView{W: 8, H: 8, Tiles: make([]world.Tile, 64), Layers: world.NewLayers(64)}
	off := RenderWindow(valley, Settlement, Window{X: 6, Y: 0, W: 4, H: 1})
	if off[0][1] == blank {
		t.Fatal("the last column of the valley drew nothing")
	}
	if off[0][2] != blank || off[0][3] != blank {
		t.Fatal("a valley has edges: past the last column there should be nothing")
	}
}

// Screen and Tile are the same thing read in opposite directions, which is
// what lets a viewer put a figure where its ground is drawn and say which
// figure a click landed on.
func TestAWindowReadsBothWays(t *testing.T) {
	m := globeView(t)
	win := Window{X: m.W - 3, Y: 5, W: 10, H: 6}
	for y := 0; y < win.H; y++ {
		for x := 0; x < win.W; x++ {
			p, on := win.Tile(m, x, y)
			if !on {
				t.Fatalf("cell %d,%d is over no ground on a globe", x, y)
			}
			bx, by, ok := win.Screen(m, p)
			if !ok || bx != x || by != y {
				t.Fatalf("tile %v reads back at %d,%d (%v), want %d,%d", p, bx, by, ok, x, y)
			}
		}
	}
	if _, _, ok := win.Screen(m, entity.Pos{X: (win.X + 40) % m.W, Y: 5}); ok {
		t.Fatal("a tile well outside the window reads as being in it")
	}
}

// Everybody standing in the window is drawn there, and nobody outside it is.
func TestAWindowDrawsTheFiguresInIt(t *testing.T) {
	m := &observe.MapView{W: 40, H: 10, Tiles: make([]world.Tile, 400), Layers: world.NewLayers(400), Agents: []observe.Mark{
		{ID: 1, Pos: entity.Pos{X: 22, Y: 4}, Action: "farm"},
		{ID: 2, Pos: entity.Pos{X: 2, Y: 4}, Action: "farm"},
	}}
	rows := RenderWindow(m, Settlement, Window{X: 20, Y: 2, W: 10, H: 5})
	if rows[2][2].Ch != '@' {
		t.Fatal("the figure inside the window was not drawn in it")
	}
	for _, row := range rows {
		for x, c := range row {
			if c.Ch == '@' && !(x == 2) {
				t.Fatalf("a figure outside the window was drawn at column %d", x)
			}
		}
	}
}
