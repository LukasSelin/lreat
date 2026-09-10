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

	valley := &observe.MapView{W: 8, H: 8, Tiles: make([]world.Tile, 64)}
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
	m := &observe.MapView{W: 40, H: 10, Tiles: make([]world.Tile, 400), Agents: []observe.Mark{
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

// tiled is a map of plain grass, for putting one thing on and asking whether
// the drawing kept it.
func tiled(w, h int) *observe.MapView {
	m := &observe.MapView{W: w, H: h, Tiles: make([]world.Tile, w*h)}
	for i := range m.Tiles {
		m.Tiles[i].Terrain = world.Grass
	}
	return m
}

// A scaled window covers more ground than it has cells, and reads back both
// ways: every tile of a block answers with the cell its block is drawn in.
func TestAScaledWindowCoversABlockToTheCell(t *testing.T) {
	m := tiled(200, 100)
	win := Window{X: 40, Y: 20, W: 10, H: 5, Z: 8}
	if w, h := win.Tiles(); w != 80 || h != 40 {
		t.Fatalf("a window of 10x5 cells at 8 tiles to the cell covers %dx%d tiles, want 80x40", w, h)
	}
	for j := 0; j < 8; j++ {
		for i := 0; i < 8; i++ {
			p := entity.Pos{X: win.X + 16 + i, Y: win.Y + 8 + j}
			x, y, ok := win.Screen(m, p)
			if !ok || x != 2 || y != 1 {
				t.Fatalf("tile %v of the block at cell 2,1 is drawn at %d,%d (%v)", p, x, y, ok)
			}
		}
	}
	if p, on := win.Tile(m, 2, 1); !on || (p != entity.Pos{X: 56, Y: 28}) {
		t.Fatalf("cell 2,1 is the block cornered at %v (%v), want 56,28", p, on)
	}
}

// The point of the whole thing: zoomed out, the map still shows what a map
// is for. A river is one tile wide and a settlement a dozen across, and a
// cell that averaged its block would have neither.
func TestZoomingOutKeepsWhatAMapIsFor(t *testing.T) {
	m := tiled(160, 80)
	for y := 0; y < m.H; y++ {
		m.At(entity.Pos{X: 71, Y: y}).Terrain = world.Water
	}
	m.At(entity.Pos{X: 8, Y: 8}).Structure = world.House
	rows := RenderWindow(m, Settlement, Window{W: 20, H: 10, Z: 8})
	if got := rows[3][8].Ch; got != '~' {
		t.Fatalf("a river a tile wide vanished at eight tiles to the cell: cell 8,3 drew %q", got)
	}
	if got := rows[1][1].Ch; got != '#' {
		t.Fatalf("a house vanished at eight tiles to the cell: cell 1,1 drew %q", got)
	}
	// And the country around them is still the country: ground with nothing
	// on it is drawn as the ground it is.
	if got := rows[5][3].Ch; got == '~' || got == '#' {
		t.Fatalf("open ground at cell 3,5 drew %q", got)
	}
}

// A reading is the land alone, so nothing built competes for the cell — but
// the water does, because a river is how a reader of any of them finds their
// way about.
func TestAScaledReadingKeepsTheWater(t *testing.T) {
	m := tiled(160, 80)
	for y := 0; y < m.H; y++ {
		m.At(entity.Pos{X: 71, Y: y}).Terrain = world.Water
	}
	rows := RenderWindow(m, Soil, Window{W: 20, H: 10, Z: 8})
	if got := rows[3][8].Ch; got != '~' {
		t.Fatalf("the soil reading lost the river at eight tiles to the cell: cell 8,3 drew %q", got)
	}
}

// Everybody in the block is somebody in the cell: zoomed out, what the map
// is asked is where the people are, not which tile each is standing on.
func TestAScaledWindowDrawsTheFiguresInTheBlock(t *testing.T) {
	m := tiled(160, 80)
	m.Agents = []observe.Mark{{ID: 1, Pos: entity.Pos{X: 37, Y: 21}, Action: "farm"}}
	rows := RenderWindow(m, Settlement, Window{W: 20, H: 10, Z: 8})
	if rows[2][4].Ch != '@' {
		t.Fatalf("the figure at 37,21 was not drawn in the cell its block is drawn in")
	}
}

// A scaled window over the seam reads round it, the same as an unscaled one:
// the block east of the last column starts at the first.
func TestAScaledWindowReadsRoundTheSeam(t *testing.T) {
	m := globeView(t)
	win := Window{X: m.W - 8, Y: 8, W: 6, H: 4, Z: 4}
	rows := RenderWindow(m, Height, win)
	for y := range rows {
		for x, c := range rows[y] {
			if c == blank {
				t.Fatalf("the scaled window across the seam has a hole in it at %d,%d", x, y)
			}
		}
	}
	// The cells past the seam are the map's own western ground, drawn at the
	// same scale from the same corner.
	west := RenderWindow(m, Height, Window{X: 0, Y: 8, W: 4, H: 4, Z: 4})
	for y := range west {
		for x := range west[y] {
			if got, want := rows[y][2+x], west[y][x]; got != want {
				t.Fatalf("past the seam at %d,%d is %v, want the map's own %v", x, y, got, want)
			}
		}
	}
}

// The unscaled window is the window everything here was written for, and a
// scale of one changes nothing about it.
func TestAScaleOfOneIsTheOldWindow(t *testing.T) {
	m := globeView(t)
	plain := RenderWindow(m, Settlement, Window{X: 10, Y: 6, W: 12, H: 8})
	one := RenderWindow(m, Settlement, Window{X: 10, Y: 6, W: 12, H: 8, Z: 1})
	for y := range plain {
		for x := range plain[y] {
			if plain[y][x] != one[y][x] {
				t.Fatalf("a scale of one drew %v at %d,%d, want the unscaled %v", one[y][x], x, y, plain[y][x])
			}
		}
	}
}

// A corner of sea is not a sea. Without a bar on how much water it takes,
// one tile of coast in a block of sixty-four turns the whole cell to open
// water and every continent on the map comes out a size too small.
func TestACornerOfWaterDoesNotDrownTheBlock(t *testing.T) {
	m := tiled(160, 80)
	m.At(entity.Pos{X: 3, Y: 3}).Terrain = world.Water
	m.At(entity.Pos{X: 4, Y: 3}).Terrain = world.Water
	rows := RenderWindow(m, Settlement, Window{W: 20, H: 10, Z: 8})
	if got := rows[0][0].Ch; got == '~' {
		t.Fatal("two tiles of water in a block of sixty-four drew the cell as open sea")
	}
	// A river crossing the block is the amount of water that does speak: it
	// is a tile wide and as long as the block is.
	for y := 0; y < 8; y++ {
		m.At(entity.Pos{X: 11, Y: y}).Terrain = world.Water
	}
	rows = RenderWindow(m, Settlement, Window{W: 20, H: 10, Z: 8})
	if got := rows[0][1].Ch; got != '~' {
		t.Fatalf("a river across the block drew %q, want the water", got)
	}
}
