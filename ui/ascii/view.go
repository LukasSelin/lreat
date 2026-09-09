package ascii

import (
	"lreat/core/entity"
	"lreat/core/observe"
	"lreat/core/world"
)

// The map has more than one thing to say about the same ground, and only one
// of them fits on a tile at a time.
//
// The settlement view is what is there: trees, water, roofs, people. It is
// the one to watch a run on. The rest are readings - how high the ground
// stands, how wet it is, what it will grow, what is standing on it, where
// people actually walk - and each is the one question the land is being asked
// answered over the whole map at once.
//
// They are readings and not decorations, which is why they are here rather
// than in the terminal viewer: every one of them is a number the simulation
// already keeps and already acts on. Soil is what a settler weighs when
// choosing where to break a field; drainage is what decides that soil; wear
// is what somebody reads before laying a road. Drawing them is the only way
// to see the world the way the people in it do.

// View is which reading of the land the map is drawn as.
type View uint8

const (
	Settlement View = iota
	Height
	Moisture
	Soil
	Woods
	Wear
	Fish
	Holdings
)

// Legend is how a reading's key should be read. A Scale runs from one end to
// the other and the question is how much; a Key is a set of colours standing
// for different things and the question is which. Drawing the second as
// though it were the first is how a map comes to say that one farmer is more
// than another.
type Legend uint8

const (
	Scale Legend = iota
	Key
)

// Reading is one view: what to call it, what the shading means at each end,
// and how to draw a tile under it.
type Reading struct {
	Name string
	// Low and High name the two ends of the ramp, for the legend. A reader
	// who cannot tell which end is which is looking at a pattern rather than
	// a map.
	Low, High string
	// Ramp is the colours the reading is drawn in, so that a legend can show
	// the same scale the map is using rather than a guess at it.
	Ramp [Bands]Color
	// Legend says whether the ramp is a scale or a key. Says is what a key
	// means, standing where Low and High stand on a scale.
	Legend Legend
	Says   string
	draw   func(scene) Cell
}

// RampOf is the scale a view is drawn in, for a legend to show.
func RampOf(v View) [Bands]Color {
	if int(v) >= len(Views) {
		return Ground
	}
	return Views[v].Ramp
}

// Views are the readings in the order the key cycles them, the settlement
// first because it is the one to come back to.
var Views = [...]Reading{
	Settlement: {Name: "settlement", Low: "", High: "", draw: nil},

	Height: {Ramp: Ground, Name: "elevation", Low: "valley floor", High: "the tops",
		draw: func(s scene) Cell { return shade(s, Ground, float64(s.b)/float64(Bands-1)) }},

	// How wet the ground is, which on this map is how far it stands above
	// the water it drains into: a tile on the valley floor beside the river
	// is a water meadow, and one a few metres up the bank is not. See
	// world.Grid.height, which is where Drain comes from.
	Moisture: {Ramp: Wet, Name: "moisture", Low: "dry", High: "wet",
		draw: func(s scene) Cell {
			if s.t.Terrain == world.Water {
				return Cell{Ch: '~', Color: Wet[Bands-1]}
			}
			return shade(s, Wet, 1-clamp(s.t.Drain/(3*world.FloodDepth)))
		}},

	// What the ground will grow. This is the fertility a field on the tile
	// would have, which is the number a settler actually weighs when choosing
	// where to break one - so the bright ground on this view is where the
	// settlement ought to be farming, and a run where the fields are not on
	// it is a run worth asking about.
	Soil: {Ramp: Crop, Name: "soil", Low: "barren", High: "good ground",
		draw: func(s scene) Cell {
			if s.t.Terrain == world.Water {
				return Cell{Ch: '~', Color: Water}
			}
			return shade(s, Crop, clamp(s.t.Rich))
		}},

	// What is standing: timber to cut and wild food to gather. Both come off
	// the same tiles and neither is visible on the settlement view beyond
	// whether a wood is 'T' or 't'.
	Woods: {Ramp: Wood, Name: "woods", Low: "bare", High: "deep wood",
		draw: func(s scene) Cell {
			if s.t.Terrain == world.Water {
				return Cell{Ch: '~', Color: Water}
			}
			return shade(s, Wood, clamp((s.t.Wood+s.t.Wild)/2))
		}},

	// Where people actually walk. Nobody plans this and nothing draws it: it
	// is worn into the ground a crossing at a time, and it is what somebody
	// deciding to lay a road is reading. A settlement's roads should sit on
	// its bright ground, and where they do not, either the road was laid too
	// early or the errands have moved since.
	// What the water has in it. Fish is the one thing a settlement lives on
	// that is not on the land at all, so this reading is mostly a picture of
	// the rivers: the ground is drawn at its quietest and only the water
	// carries anything. A river that has been fished out shows here as a
	// channel gone dark, which is the same channel on every other view.
	Fish: {Ramp: Shoal, Name: "fish", Low: "fished out", High: "teeming",
		draw: func(s scene) Cell {
			if s.t.Terrain != world.Water {
				return Cell{Ch: '.', Color: Bare}
			}
			return shade(s, Shoal, clamp(s.t.Fish))
		}},

	// Who holds what. This is the one reading that is not a quantity: a
	// holding is somebody's, and the colours stand for different owners
	// rather than for more and less of one thing. Six of them for however
	// many settlers there are, so two holdings across the map from each
	// other may share a colour - what this is for is seeing the shape of a
	// holding and how the claimed ground sits against the unclaimed, not
	// telling one farmer from another by eye.
	Holdings: {Ramp: Held, Name: "holdings", Legend: Key,
		Says: "a colour to a holder, unclaimed ground dim",
		draw: func(s scene) Cell {
			if s.t.Owner == 0 {
				if s.t.Terrain == world.Water {
					return Cell{Ch: '~', Color: Bare}
				}
				return Cell{Ch: '.', Color: Bare}
			}
			return Cell{Ch: '#', Color: Held[int(s.t.Owner)%Bands]}
		}},

	Wear: {Ramp: Worn, Name: "wear", Low: "untrodden", High: "a thoroughfare",
		draw: func(s scene) Cell {
			return shade(s, Worn, clamp(s.t.Traffic/wornEnough))
		}},
}

// wornEnough is the traffic at which ground reads as a way people take, and
// so the top of the wear ramp. Above it the shading would say nothing more.
const wornEnough = 200

// shades are the six steps a reading is drawn in, thin to solid. They rise in
// ink as well as in colour so that a reading survives being printed without
// any colour at all, which is how the headless command and every test that
// reads a map see it.
var shades = [Bands]rune{'.', ':', '-', '=', '+', '#'}

// shade draws one tile of a reading: v in [0,1] picks the same step of the
// glyph ramp and of the colour ramp, so the two say the same thing twice.
func shade(_ scene, ramp [Bands]Color, v float64) Cell {
	i := int(clamp(v) * float64(Bands))
	if i >= Bands {
		i = Bands - 1
	}
	return Cell{Ch: shades[i], Color: ramp[i]}
}

// clamp holds a reading inside [0,1].
func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// RenderView draws the map under one reading. The settlement view is the
// ordinary map with everybody on it; a reading is the land alone, with the
// water left in because a river is how anybody finds their way around a map,
// and the market left in because it is where the settlement is.
func RenderView(m *observe.MapView, view View) [][]Cell {
	if int(view) >= len(Views) || Views[view].draw == nil {
		return Render(m)
	}
	g := &world.Grid{W: m.W, H: m.H, Tiles: m.Tiles}
	rows := make([][]Cell, m.H)
	for y := 0; y < m.H; y++ {
		rows[y] = make([]Cell, m.W)
		for x := 0; x < m.W; x++ {
			p := entity.Pos{X: x, Y: y}
			t := g.At(p)
			rows[y][x] = Views[view].draw(scene{g: g, p: p, t: t, b: band(g, t.Height)})
		}
	}
	if m.Market.X != 0 || m.Market.Y != 0 {
		rows[m.Market.Y][m.Market.X] = Cell{Ch: 'M', Color: Market}
	}
	return rows
}

// LinesView renders one reading as plain strings, for logs and tests.
func LinesView(m *observe.MapView, view View) []string {
	return linesOf(RenderView(m, view))
}
