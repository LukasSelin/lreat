// Package ascii turns a map snapshot into glyphs, Dwarf Fortress style.
//
// It knows nothing about terminals. Cells carry a symbolic color that the
// terminal viewer maps to real colors and that the headless command ignores.
package ascii

import (
	"strings"

	"lreat/core/observe"
	"lreat/core/world"
)

// Color is a symbolic color for a cell.
type Color uint8

const (
	Default Color = iota
	Water
	Grass
	ForestRich
	ForestPoor
	Field
	House
	Market
	AgentFood
	AgentBuild
	AgentTrade
	AgentGuard
	AgentSocial
	AgentStudy
	AgentIdle
)

// Cell is one rendered glyph.
type Cell struct {
	Ch    rune
	Color Color
}

// Render draws the map as rows of cells. Agents draw over tiles; where
// several agents share a tile the first one's activity sets the color.
func Render(m *observe.MapView) [][]Cell {
	rows := make([][]Cell, m.H)
	for y := 0; y < m.H; y++ {
		rows[y] = make([]Cell, m.W)
		for x := 0; x < m.W; x++ {
			rows[y][x] = tileCell(&m.Tiles[y*m.W+x])
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

func tileCell(t *world.Tile) Cell {
	switch t.Structure {
	case world.House:
		return Cell{Ch: '#', Color: House}
	case world.Market:
		return Cell{Ch: 'M', Color: Market}
	}
	switch t.Terrain {
	case world.Water:
		return Cell{Ch: '~', Color: Water}
	case world.Forest:
		if t.Wood >= 0.5 {
			return Cell{Ch: 'T', Color: ForestRich}
		}
		return Cell{Ch: 't', Color: ForestPoor}
	case world.Field:
		return Cell{Ch: '"', Color: Field}
	}
	return Cell{Ch: '.', Color: Grass}
}

// AgentColor maps an action to the color of the agent doing it.
func AgentColor(action string) Color {
	switch action {
	case "farm", "forage", "eat", "buy food":
		return AgentFood
	case "gather wood", "build shelter", "craft":
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
