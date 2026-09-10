package system

import (
	"lreat/core/action"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/ontology"
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
	// Craft is the skill this technology's work lives in, and nil for the
	// ones whose work is nobody's craft - a tavern is not a trade and a
	// fish trap is not either. It is what mastery is read against: a
	// settlement has mastered a technology when somebody in it stands in
	// the master tier of its craft, which under entity's ladder is a thing
	// only doing the work can reach. So the date is earned twice over and
	// cannot be got by being told.
	Craft *entity.Skill
}

// craft names the skill a technology's work lives in. It is a function
// because a skill's zero value is farming, and a field left unset would
// quietly claim to be a farm.
func craft(s entity.Skill) *entity.Skill { return &s }

// adept counts the people a settlement can call on for a skill: those who
// have both got good at it and actually done it.
//
// Both halves are needed and neither is the other. Since a lesson stops
// short of the teacher and reading stops sooner still, somebody can stand in
// the journeyman tier having been shown the whole of it and never once had
// their hands on the work - and it was that person the discoveries used to
// be unlocked by. Agriculture asked for two people at a fifth of farming, in
// settlements whose most practised farmer had broken ground five times in
// sixty years; the fields were being invented by people who had never
// really farmed. Practice is the other half: how many times this pair of
// hands has actually done it, which only entity.Learn raises.
//
// The numbers each discovery asks for are read off what settlements
// actually reach rather than chosen: see the tune probe quoted in
// docs/baseline.md.
func adept(w *world.World, s entity.Skill, t entity.Tier, done int) int {
	n := 0
	for _, a := range w.Agents {
		if entity.TierOf(a.Skills[s]) >= t && a.Practice[s] >= done {
			n++
		}
	}
	return n
}

// Discoveries is the catalog, checked in order every tick.
var Discoveries = []Discovery{
	{
		Tech: "agriculture", Knowledge: 15, Craft: craft(entity.Farming),
		// Two people who have worked ground eighty times and are
		// out of the novice tier for it. Of six probed settlements this
		// passes four and refuses two, and the two it refuses are the two
		// that never farmed: their best hands had four and sixteen days on
		// the ground in sixty years. A people who forage do not invent the
		// field, and they no longer get the yield for it.
		Condition: func(w *world.World) bool { return adept(w, entity.Farming, entity.Apprentice, 80) >= 2 },
		Effect:    func(w *world.World) { w.Mods.FarmYield *= 1.8 },
		Text:      "farmers learned to rotate their fields",
	},
	{
		Tech: "masonry", Knowledge: 40, Craft: craft(entity.Building),
		// Two journeymen with forty raisings behind them. Every settlement
		// probed had them, which is the point: masonry opens the craft, the
		// road and the granary, and a gate that shut it would shut most of
		// what comes after. It is a raise on 0.3 of building and not a wall.
		Condition: func(w *world.World) bool { return adept(w, entity.Building, entity.Journeyman, 40) >= 2 },
		Effect: func(w *world.World) {
			w.Mods.BuildEfficiency *= 1.6
			w.Mods.ShelterDecay *= 0.5
		},
		Text:  "builders began working in stone",
		Opens: []string{"craft", "lay road", "build granary"},
	},
	{
		Tech: "writing", Knowledge: 90, Craft: craft(entity.Scholarship),
		// Three journeymen who have each done the work twenty times. Since
		// reading alone stops at the top of the apprentice tier, a
		// journeyman scholar is by construction one who has tutored rather
		// than only studied - so this asks for three people who have taught
		// what they know, which is close to what a written record is for.
		Condition: func(w *world.World) bool { return adept(w, entity.Scholarship, entity.Journeyman, 20) >= 3 },
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
		Tech: "fishing", Knowledge: 0, Craft: craft(entity.Fishing),
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
		Tech: "irrigation", Knowledge: 10, Craft: craft(entity.Farming),
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
	// Making and keeping. Pottery answers a glut: food piling up at the
	// market and spoiling there. Quarrying comes to masons who have stone
	// near.
	{
		Tech: "pottery", Knowledge: 30, Craft: craft(entity.Crafting),
		Condition: func(w *world.World) bool { return w.Market.Stock[entity.Food] >= 8 },
		Effect:    func(w *world.World) { w.Mods.Keeping *= 0.7 },
		Text:      "with food spoiling at the market, potters learned to keep it",
		Opens:     []string{"cook"},
	},
	{
		Tech: "quarrying", Knowledge: 40, Craft: craft(entity.Building),
		Condition: func(w *world.World) bool { return w.Has("masonry") && rockNear(w) },
		Effect:    func(w *world.World) { w.Mods.BuildEfficiency *= 1.2 },
		Text:      "masons learned to cut stone from the outcrops",
		Opens:     []string{"quarry"},
	},
	// Brewing answers a settlement big enough to be lonely in, with food to
	// spare for it.
	{
		Tech: "brewing", Knowledge: 25,
		Condition: func(w *world.World) bool { return len(w.Agents) >= 12 && w.Market.Stock[entity.Food] >= 5 },
		Text:      "with grain to spare, somebody opened a tavern",
		Opens:     []string{"build tavern"},
		Effect:    func(*world.World) {},
	},
	{
		Tech: "metallurgy", Knowledge: 200, Craft: craft(entity.Crafting),
		Condition: func(w *world.World) bool {
			return w.Has("writing") && adept(w, entity.Crafting, entity.Master, 100) >= 2 && w.Market.Stock[entity.Tools] >= 5
		},
		Effect: func(w *world.World) {
			w.Mods.CraftQuality *= 2
			w.Mods.FarmYield *= 1.3
		},
		Text:  "smiths learned to work metal",
		Opens: []string{"craft", "guard", "smelt"},
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
	// Only the window is walked, in the order a walk over the whole map
	// would have come to its tiles, so the sum lands in the same order.
	g, m := w.Grid, w.MarketPos
	x0, x1 := g.Columns(m.X, nearMarket)
	for y := max(0, m.Y-nearMarket); y <= min(g.H-1, m.Y+nearMarket); y++ {
		for dx := x0; dx <= x1; dx++ {
			p := g.Norm(entity.Pos{X: m.X + dx, Y: y})
			t := g.At(p)
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
				if t := w.Grid.At(p); t.Is(ontology.Wood) {
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
	_, n := meanOf(w, func(t *world.Tile) bool { return t.Is(ontology.Water) }, func(*world.Tile) float64 { return 0 })
	return n > 0
}

// fieldsWorn reports whether the fields near the market have gone poor.
func fieldsWorn(w *world.World) bool {
	fert, n := meanOf(w, func(t *world.Tile) bool { return t.Is(ontology.Field) }, func(t *world.Tile) float64 { return t.Fertility })
	return n >= 3 && fert < 0.45
}

// forestGone reports whether the settlement has cleared much of the forest
// it was founded among.
func forestGone(w *world.World) bool {
	if w.Forest0 == 0 {
		return false
	}
	now := w.Grid.Forest()
	return float64(now) < 0.6*float64(w.Forest0)
}

// rockNear reports whether there is stone to cut near the market.
func rockNear(w *world.World) bool {
	_, ok := w.Grid.Nearest(w.MarketPos, nearMarket, func(_ entity.Pos, t *world.Tile) bool { return t.Offers(ontology.Stone) > 0 })
	return ok
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

// Discover unlocks any technology whose conditions the world now meets, and
// notes the ones the settlement has since become master of.
func Discover(w *world.World) {
	for _, d := range Discoveries {
		if w.Has(d.Tech) {
			// Held already: the only thing left to happen to it is that
			// somebody becomes a master of its craft. adept with the master
			// tier asks for exactly that, and asks for no practice beside
			// it because none is needed - being shown a craft stops at the
			// threshold of that tier, so anybody standing in it worked
			// their way there.
			if d.Craft != nil && w.Known(d.Tech).Mastered == 0 && adept(w, *d.Craft, entity.Master, 0) >= 1 {
				w.Master(d.Tech)
				w.Emit(event.Discovered, 0, 0, "%s: the settlement has a master of it", d.Tech)
			}
			continue
		}
		if w.Knowledge < d.Knowledge {
			continue
		}
		if d.Condition != nil && !d.Condition(w) {
			continue
		}
		w.Unlock(d.Tech)
		d.Effect(w)
		w.Room()
		for _, name := range d.Opens {
			if i := action.Index(action.ByName(name)); i >= 0 {
				w.ReachFloor[i] = max(w.ReachFloor[i], action.Opened)
			}
		}
		w.Emit(event.Discovered, 0, 0, "%s: %s", d.Tech, d.Text)
	}
}
