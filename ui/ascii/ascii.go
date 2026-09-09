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

// Ground and Wood are the two ramps, low to high.
var (
	Ground = [Bands]Color{Ground0, Ground1, Ground2, Ground3, Ground4, Ground5}
	Wood   = [Bands]Color{Wood0, Wood1, Wood2, Wood3, Wood4, Wood5}
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
func Lines(m *observe.MapView) []string {
	cells := Render(m)
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
	b := band(g, t.Height)
	switch t.Terrain {
	case world.Water:
		return Cell{Ch: '~', Color: Water}
	case world.Forest:
		if t.Wood >= 0.5 {
			return Cell{Ch: 'T', Color: Wood[b]}
		}
		return Cell{Ch: 't', Color: Wood[b]}
	case world.Field:
		// A hedged holding is drawn as hedged. It is the one thing on the map
		// that changes how a journey goes without anything being built on a
		// tile, so it has to be visible or the ways people take round it look
		// like nothing at all.
		if t.Fenced {
			return Cell{Ch: '=', Color: FieldFenced}
		}
		return Cell{Ch: '"', Color: Field}
	case world.Rock:
		// An outcrop on the valley floor is a boulder field and one on the
		// skyline is a crag, and they should not be the same grey.
		if b >= Bands/2 {
			return Cell{Ch: '^', Color: RockHigh}
		}
		return Cell{Ch: '^', Color: Rock}
	}
	// Open ground. Ground that falls away fast enough to be felt is drawn as
	// a hillside whatever else is true of it, so that the shape of the
	// country survives being printed without colour - which is how the
	// headless command and every test that reads a map see it.
	if g.Slope(p) >= hillside {
		return Cell{Ch: 'n', Color: Ground[b]}
	}
	// Otherwise the glyph is how wet the ground is: the water meadows of the
	// valley floor, the ordinary ground of the terraces, and the dry slopes
	// above.
	switch {
	case t.Drain < world.FloodDepth/3:
		return Cell{Ch: ',', Color: Ground[b]}
	case t.Drain > world.FloodDepth*2:
		return Cell{Ch: '`', Color: Ground[b]}
	}
	return Cell{Ch: '.', Color: Ground[b]}
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
