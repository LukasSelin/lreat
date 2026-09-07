package system

import (
	"lreat/core/action"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/world"
)

// Discovery is a technology waiting to be noticed. Knowledge is the raw
// threshold; Condition is the social shape the settlement must have. Nobody
// chooses these. They happen to a city that has become a certain kind of place.
type Discovery struct {
	Tech      world.Tech
	Knowledge float64
	Condition func(w *world.World) bool
	Effect    func(w *world.World)
	Text      string
	// Opens names actions the discovery puts within everyone's reach.
	Opens []string
}

// skilled counts agents with at least level in a skill.
func skilled(w *world.World, s entity.Skill, level float64) int {
	n := 0
	for _, a := range w.Agents {
		if a.Skills[s] >= level {
			n++
		}
	}
	return n
}

// Discoveries is the catalog, checked in order every tick.
var Discoveries = []Discovery{
	{
		Tech: "agriculture", Knowledge: 15,
		Condition: func(w *world.World) bool { return skilled(w, entity.Farming, 0.2) >= 2 },
		Effect:    func(w *world.World) { w.Mods.FarmYield *= 1.8 },
		Text:      "farmers learned to rotate their fields",
	},
	{
		Tech: "masonry", Knowledge: 40,
		Condition: func(w *world.World) bool { return skilled(w, entity.Building, 0.3) >= 2 },
		Effect: func(w *world.World) {
			w.Mods.BuildEfficiency *= 1.6
			w.Mods.ShelterDecay *= 0.5
		},
		Text:  "builders began working in stone",
		Opens: []string{"craft"},
	},
	{
		Tech: "writing", Knowledge: 90,
		Condition: func(w *world.World) bool { return skilled(w, entity.Scholarship, 0.3) >= 3 },
		Effect:    func(w *world.World) { w.Mods.StudyRate *= 2 },
		Text:      "scholars started keeping written records",
		Opens:     []string{"study", "teach"},
	},
	// The land's answers. Each is learned only by a settlement under the
	// pressure it answers, so different worlds take different paths: one
	// with a thin forest by a river learns to fish, one that has cleared its
	// woods learns to plant them, one whose fields have gone poor learns to
	// bring water to them. The first two take no learning at all, only
	// need: a hungry people by a river will fish. The later two take some.
	{
		Tech: "fishing", Knowledge: 0,
		Condition: func(w *world.World) bool { return waterNear(w) && forestThin(w) },
		Effect:    func(w *world.World) { w.Mods.FishYield *= 1.5 },
		Text:      "with the woods picked thin, people turned to the river",
		Opens:     []string{"fish"},
	},
	{
		Tech: "trapping", Knowledge: 0,
		Condition: func(w *world.World) bool { return forestThin(w) && tooled(w) >= 2 },
		Effect:    func(w *world.World) { w.Mods.HuntYield *= 1.5 },
		Text:      "hunters learned to set snares",
		Opens:     []string{"hunt"},
	},
	{
		Tech: "irrigation", Knowledge: 10,
		Condition: func(w *world.World) bool { return w.Has("agriculture") && fieldsWorn(w) },
		Effect:    func(w *world.World) { w.Mods.FarmYield *= 1.2 },
		Text:      "with the fields gone poor, farmers cut channels from the river",
		Opens:     []string{"irrigate"},
	},
	{
		Tech: "forestry", Knowledge: 10,
		Condition: forestGone,
		Effect:    func(w *world.World) { w.Mods.Regrowth *= 2 },
		Text:      "with the woods cleared, people began to plant them",
		Opens:     []string{"plant trees"},
	},
	{
		Tech: "metallurgy", Knowledge: 200,
		Condition: func(w *world.World) bool {
			return w.Has("writing") && skilled(w, entity.Crafting, 0.4) >= 2 && w.Market.Stock[entity.Tools] >= 5
		},
		Effect: func(w *world.World) {
			w.Mods.CraftQuality *= 2
			w.Mods.FarmYield *= 1.3
		},
		Text:  "smiths learned to work metal",
		Opens: []string{"craft", "guard"},
	},
}

// Pressures a settlement can be under, which is what the later discoveries
// answer. None of them is a number anyone in the settlement reads; they are
// the state of the land around the market, and a settlement that has not
// worn its land finds no need to learn these things.

// nearMarket is how far around the market the land is looked at.
const nearMarket = 20

// meanOf averages f over the tiles near the market that ok picks out.
func meanOf(w *world.World, ok func(*world.Tile) bool, f func(*world.Tile) float64) (float64, int) {
	var sum float64
	n := 0
	for y := 0; y < w.Grid.H; y++ {
		for x := 0; x < w.Grid.W; x++ {
			p := entity.Pos{X: x, Y: y}
			if entity.Dist(p, w.MarketPos) > nearMarket {
				continue
			}
			t := w.Grid.At(p)
			if ok(t) {
				sum += f(t)
				n++
			}
		}
	}
	if n == 0 {
		return 0, 0
	}
	return sum / float64(n), n
}

// usedForest is how many of the forest tiles nearest the market count as
// the forest a settlement lives off. Foragers go to the nearest forest, so
// it is these few tiles that are picked bare while the woods beyond stay
// full, and it is these that say whether the settlement feels the land
// pushing back.
const usedForest = 24

// forestThin reports whether the forest the settlement lives off has little
// left to give: it is being foraged faster than it comes back.
func forestThin(w *world.World) bool {
	var wild float64
	n := 0
	// Ring outward from the market, nearest tiles first, until enough forest
	// has been seen.
	for r := 0; r <= nearMarket && n < usedForest; r++ {
		for dy := -r; dy <= r && n < usedForest; dy++ {
			for dx := -r; dx <= r && n < usedForest; dx++ {
				if max(abs(dx), abs(dy)) != r {
					continue
				}
				p := entity.Pos{X: w.MarketPos.X + dx, Y: w.MarketPos.Y + dy}
				if !w.Grid.In(p) {
					continue
				}
				if t := w.Grid.At(p); t.Terrain == world.Forest {
					wild += t.Wild
					n++
				}
			}
		}
	}
	return n == 0 || wild/float64(n) < 0.5
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// waterNear reports whether there is water to fish near the market.
func waterNear(w *world.World) bool {
	_, n := meanOf(w, func(t *world.Tile) bool { return t.Terrain == world.Water }, func(*world.Tile) float64 { return 0 })
	return n > 0
}

// fieldsWorn reports whether the fields near the market have gone poor.
func fieldsWorn(w *world.World) bool {
	fert, n := meanOf(w, func(t *world.Tile) bool { return t.Terrain == world.Field }, func(t *world.Tile) float64 { return t.Fertility })
	return n >= 3 && fert < 0.45
}

// forestGone reports whether the settlement has cleared much of the forest
// it was founded among.
func forestGone(w *world.World) bool {
	if w.Forest0 == 0 {
		return false
	}
	now := w.Grid.Count(func(t *world.Tile) bool { return t.Terrain == world.Forest })
	return float64(now) < 0.6*float64(w.Forest0)
}

// tooled counts agents holding at least a tool's worth of tools.
func tooled(w *world.World) int {
	n := 0
	for _, a := range w.Agents {
		if a.Inventory[entity.Tools] >= 0.5 {
			n++
		}
	}
	return n
}

// Discover unlocks any technology whose conditions the world now meets.
func Discover(w *world.World) {
	for _, d := range Discoveries {
		if w.Has(d.Tech) || w.Knowledge < d.Knowledge {
			continue
		}
		if d.Condition != nil && !d.Condition(w) {
			continue
		}
		w.Unlock(d.Tech)
		d.Effect(w)
		for _, name := range d.Opens {
			if i := action.Index(action.ByName(name)); i >= 0 {
				w.ReachFloor[i] = max(w.ReachFloor[i], action.Opened)
			}
		}
		w.Emit(event.Discovered, 0, 0, "%s: %s", d.Tech, d.Text)
	}
}
