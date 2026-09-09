// Package ascii turns a map snapshot into glyphs, Dwarf Fortress style.
//
// It knows nothing about terminals. Cells carry a symbolic color that the
// terminal viewer maps to real colors and that the headless command ignores.
package ascii

import (
	"strings"

	"lreat/core/entity"
	"lreat/core/observe"
	"lreat/core/world"
)

// Color is a symbolic color for a cell.
type Color uint8

const (
	Default Color = iota
	Water
	Field
	FieldFenced
	House
	Market
	Road
	Rock
	RockHigh
	Granary
	Tavern
	AgentFood
	AgentBuild
	AgentTrade
	AgentGuard
	AgentSocial
	AgentStudy
	AgentIdle
	// Ground and Wood are open country and woodland, each in Bands steps
	// from the valley floor to the skyline. They must stay contiguous and in
	// order; Ground and Wood below index them.
	Ground0
	Ground1
	Ground2
	Ground3
	Ground4
	Ground5
	Wood0
	Wood1
	Wood2
	Wood3
	Wood4
	Wood5
	// The ramps the readings are drawn in: how wet the ground is, what it
	// will grow, and how hard it is walked. Contiguous and in order, like the
	// two above; see the Ramp vars below.
	Wet0
	Wet1
	Wet2
	Wet3
	Wet4
	Wet5
	Crop0
	Crop1
	Crop2
	Crop3
	Crop4
	Crop5
	Worn0
	Worn1
	Worn2
	Worn3
	Worn4
	Worn5
	Shoal0
	Shoal1
	Shoal2
	Shoal3
	Shoal4
	Shoal5
	// Held are the colours a holding is drawn in. Unlike every ramp above
	// them these are not one colour getting stronger: they stand for
	// different owners and not for more and less of one thing, so they are
	// six hues as unlike each other as the terminal affords.
	Held0
	Held1
	Held2
	Held3
	Held4
	Held5
	// Bare is ground nobody has claimed, and the quietest thing on the map:
	// on a reading about what is held, what is not held should recede.
	Bare
)

// Grass and ForestRich name the middle of each ramp, for callers that want
// the colour of open country or of woodland without meaning any particular
// height - a legend, or a line on a graph.
const (
	Grass      = Ground1
	ForestRich = Wood1
)

// Bands is how many steps of colour the map is drawn in from the valley floor
// to the top of the skyline. Six, because the eye can tell six shades of one
// hue apart at a glance and cannot tell twelve.
const Bands = 6

// The ramps, each low to high. Ground and Wood draw the settlement view as
// well as their own readings; the rest are only ever a reading.
var (
	Ground = [Bands]Color{Ground0, Ground1, Ground2, Ground3, Ground4, Ground5}
	Wood   = [Bands]Color{Wood0, Wood1, Wood2, Wood3, Wood4, Wood5}
	Wet    = [Bands]Color{Wet0, Wet1, Wet2, Wet3, Wet4, Wet5}
	Crop   = [Bands]Color{Crop0, Crop1, Crop2, Crop3, Crop4, Crop5}
	Worn   = [Bands]Color{Worn0, Worn1, Worn2, Worn3, Worn4, Worn5}
	Shoal  = [Bands]Color{Shoal0, Shoal1, Shoal2, Shoal3, Shoal4, Shoal5}
	Held   = [Bands]Color{Held0, Held1, Held2, Held3, Held4, Held5}
)

// band is which step of the ramp a height falls on. The valley's own relief
// takes the lower half of the ramp and everything standing above the valley
// takes the upper half, which is the same split the land is built on: see
// world.Relief and world.Grid.UplandRise. The upper half is measured against
// the map's own high country rather than a fixed height, because how far the
// mountains rise depends on how wide the map is.
//
// A single ramp stretched over the whole range was the first try and it was
// no good. The lowland is a fifth of the height of the map and a settlement
// spends its whole life on it, so it came out as one flat colour - which is
// the complaint this is all meant to answer, moved from the ground to the
// picture of it.
func band(g *world.Grid, h float64) int {
	half := Bands / 2
	if h < world.Relief {
		return min(int(h/world.Relief*float64(half)), half-1)
	}
	up := (h - world.Relief) / g.UplandRise()
	return min(half+int(up*float64(Bands-half)), Bands-1)
}

// Cell is one rendered glyph.
type Cell struct {
	Ch    rune
	Color Color
}

// Render draws the map as rows of cells. Agents draw over tiles; where
// several agents share a tile the first one's activity sets the color.
func Render(m *observe.MapView) [][]Cell {
	// The snapshot's tiles read as a grid, so that the drawing can ask the
	// land the same questions the world asks it - chiefly how steeply a tile
	// falls away, which is a reading of the eight tiles round it and not of
	// the tile itself.
	g := &world.Grid{W: m.W, H: m.H, Tiles: m.Tiles}
	rows := make([][]Cell, m.H)
	for y := 0; y < m.H; y++ {
		rows[y] = make([]Cell, m.W)
		for x := 0; x < m.W; x++ {
			rows[y][x] = tileCell(g, entity.Pos{X: x, Y: y})
		}
	}
	seen := make(map[[2]int]bool, len(m.Agents))
	for _, a := range m.Agents {
		key := [2]int{a.Pos.X, a.Pos.Y}
		if seen[key] {
			continue
		}
		seen[key] = true
		rows[a.Pos.Y][a.Pos.X] = Cell{Ch: '@', Color: AgentColor(a.Action)}
	}
	return rows
}

// Lines renders the map as plain strings, for logs and tests.
func Lines(m *observe.MapView) []string { return linesOf(Render(m)) }

// linesOf is the glyphs of a drawn map, a string to the row.
func linesOf(cells [][]Cell) []string {
	out := make([]string, len(cells))
	for y, row := range cells {
		var b strings.Builder
		for _, c := range row {
			b.WriteRune(c.Ch)
		}
		out[y] = b.String()
	}
	return out
}

// tileCell is one tile: the glyph says what the ground is made of and how it
// lies, and the colour says how high it stands. Splitting them this way is
// what puts the shape of the country into the picture at all. Drawn with the
// colour saying what the tile was made of, every wood on the map was the same
// green whether it stood in the flood plain or on a shoulder eight hundred
// feet up, and the only relief anywhere in the drawing was three shades of
// open grass - which on a map that is two thirds trees, water and buildings
// is no relief at all.
func tileCell(g *world.Grid, p entity.Pos) Cell {
	t := g.At(p)
	switch t.Structure {
	case world.House:
		return Cell{Ch: '#', Color: House}
	case world.Market:
		return Cell{Ch: 'M', Color: Market}
	case world.Road:
		return Cell{Ch: '+', Color: Road}
	case world.Granary:
		return Cell{Ch: 'G', Color: Granary}
	case world.Tavern:
		return Cell{Ch: '&', Color: Tavern}
	}
	return ground[t.Terrain](scene{g: g, p: p, t: t, b: band(g, t.Height)})
}

// scene is what drawing one tile needs: the tile, where it is, the grid it
// sits in for the questions that are about its neighbours, and which height
// band it falls in.
type scene struct {
	g *world.Grid
	p entity.Pos
	t *world.Tile
	b int
}

// ground draws each kind of ground, a row apiece. It is a table rather than a
// switch because a switch has a default and this should not: a terrain the
// renderer has not been told about is a nil to trip over in a test, where a
// default is a tile quietly drawn as open grass on every map from then on.
var ground = [world.TerrainCount]func(scene) Cell{
	world.Water: func(scene) Cell { return Cell{Ch: '~', Color: Water} },

	// A wood is drawn by how much of it is left to cut, and coloured by how
	// high it stands.
	world.Forest: func(s scene) Cell {
		if s.t.Wood >= 0.5 {
			return Cell{Ch: 'T', Color: Wood[s.b]}
		}
		return Cell{Ch: 't', Color: Wood[s.b]}
	},

	// A hedged holding is drawn as hedged. It is the one thing on the map
	// that changes how a journey goes without anything being built on a
	// tile, so it has to be visible or the ways people take round it look
	// like nothing at all.
	world.Field: func(s scene) Cell {
		if s.t.Fenced {
			return Cell{Ch: '=', Color: FieldFenced}
		}
		return Cell{Ch: '"', Color: Field}
	},

	// An outcrop on the valley floor is a boulder field and one on the
	// skyline is a crag, and they should not be the same grey.
	world.Rock: func(s scene) Cell {
		if s.b >= Bands/2 {
			return Cell{Ch: '^', Color: RockHigh}
		}
		return Cell{Ch: '^', Color: Rock}
	},

	// Open ground. Ground that falls away fast enough to be felt is drawn as
	// a hillside whatever else is true of it, so that the shape of the
	// country survives being printed without colour - which is how the
	// headless command and every test that reads a map see it. Otherwise the
	// glyph is how wet the ground is: the water meadows of the valley floor,
	// the ordinary ground of the terraces, and the dry slopes above.
	world.Grass: func(s scene) Cell {
		if s.g.Slope(s.p) >= hillside {
			return Cell{Ch: 'n', Color: Ground[s.b]}
		}
		switch {
		case s.t.Drain < world.FloodDepth/3:
			return Cell{Ch: ',', Color: Ground[s.b]}
		case s.t.Drain > world.FloodDepth*2:
			return Cell{Ch: '`', Color: Ground[s.b]}
		}
		return Cell{Ch: '.', Color: Ground[s.b]}
	},
}

// hillside is the slope at which open ground stops being ground you walk over
// and starts being ground you walk up: one in five, which over a
// five-and-twenty metre tile is five metres of climbing, and about half
// another tile's walking on top of the tile itself. See world.Climb.
const hillside = 0.2

// AgentColor maps an action to the color of the agent doing it.
func AgentColor(action string) Color {
	switch action {
	case "farm", "forage", "eat", "buy food", "fish", "hunt", "cook":
		return AgentFood
	case "gather wood", "build shelter", "craft", "irrigate", "plant trees", "lay road", "quarry", "build granary", "smelt", "build tavern":
		return AgentBuild
	case "sell":
		return AgentTrade
	case "guard":
		return AgentGuard
	case "socialize", "teach":
		return AgentSocial
	case "study":
		return AgentStudy
	}
	return AgentIdle
}

// Group is a kind of work, the unit the activity graph stacks by. Which
// action an agent picks flips from tick to tick — farm, forage, eat — but
// the kind of work it belongs to holds still long enough to read.
type Group struct {
	Name  string
	Color Color
}

// Groups are every kind of work, in the order the graph stacks them. The
// order is fixed so a band stays where it was between frames.
var Groups = [...]Group{
	{"food", AgentFood},
	{"build", AgentBuild},
	{"trade", AgentTrade},
	{"guard", AgentGuard},
	{"social", AgentSocial},
	{"study", AgentStudy},
	{"other", AgentIdle},
}

// GroupOf returns the index in Groups of the kind of work an action is.
func GroupOf(action string) int {
	c := AgentColor(action)
	for i, g := range Groups {
		if g.Color == c {
			return i
		}
	}
	return len(Groups) - 1
}
