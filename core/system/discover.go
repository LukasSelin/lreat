package system

import (
	"lreat/core/action"
	"lreat/core/clock"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/habit"
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
	// Signature is the kind of moment this discovery belongs to, in the
	// same space agents recognise their own moments in, and Cost is how
	// many days of that moment it takes to arrive at. Together they are
	// the difference between a technology a settlement is handed the
	// instant its conditions read true and one it works its way to.
	//
	// The two halves are direction and length, which is the whole of the
	// idea. Direction says whether this settlement is working on the thing
	// at all: a people who are cold and fed are pointed somewhere else than
	// a people who are warm and starving, and neither is pointed where a
	// comfortable people are. Length says how fast, because the projection
	// of a person's situation onto the discovery's direction is how hard
	// that moment is actually pressing on them - and a settlement with
	// nothing much wrong with it has a short situation vector in every
	// direction, so it makes no progress toward anything. That a
	// prosperous people stagnate is not a rule here. It is what adding
	// nearly zero for sixty years comes to.
	//
	// Zero Cost is the old behaviour: the moment the Condition reads true,
	// it is had. The skill-gated half of the catalog is still that way,
	// because wanting a thing badly is not how anybody comes by a master
	// mason.
	Signature habit.Signature
	Cost      float64
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
		// A hungry person standing by water. It is the plainest moment in
		// the catalog and it wants nothing else in it.
		Signature: habit.Signature{habit.Hunger: 1},
		Cost:      pressYear,
		Condition: func(w *world.World) bool { return waterNear(w) && forestThin(w) },
		Effect:    func(w *world.World) { w.Mods.FishYield *= 1.5 },
		Text:      "with the woods picked thin, people turned to the river",
		Opens:     []string{"fish"},
	},
	{
		Tech: "trapping", Knowledge: 0,
		// Hungry, and of a mind to make something rather than go and look.
		// A snare is patience, so this one asks for industry in the person
		// as well as want in the belly.
		Signature: habit.Signature{habit.Hunger: 1, habit.Industry: 0.5},
		Cost:      1.5 * pressYear,
		Condition: func(w *world.World) bool { return forestThin(w) && tooled(w) >= 2 },
		Effect:    func(w *world.World) { w.Mods.HuntYield *= 1.5 },
		Text:      "hunters learned to set snares",
		Opens:     []string{"hunt"},
	},
	{
		Tech: "irrigation", Knowledge: 10, Craft: craft(entity.Farming),
		// Hungry and industrious both, and more of the second than fishing
		// wants: nobody cuts a channel on the day they are hungry, they cut
		// it in the year they have been.
		Signature: habit.Signature{habit.Hunger: 1, habit.Industry: 0.8},
		Cost:      2 * pressYear,
		Condition: func(w *world.World) bool { return w.Has("agriculture") && fieldsDry(w) },
		Effect:    func(w *world.World) { w.Mods.FarmYield *= 1.2 },
		Text:      "with the fields out of the river's reach, farmers cut channels to them",
		Opens:     []string{"irrigate"},
	},
	{
		Tech: "forestry", Knowledge: 10,
		// Planting a wood is the longest thought anybody here has: it is
		// work whose good falls to somebody else, so it belongs to the
		// industrious and the traditional rather than to the hungry.
		Signature: habit.Signature{habit.Industry: 1, habit.Tradition: 0.6, habit.Curious: 0.3},
		Cost:      2 * pressYear,
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
		// Somebody with the leisure to wonder about a thing. Pottery is the
		// one here that answers plenty rather than want, and its moment is
		// curiosity rather than hunger - which is also why a settlement
		// scraping by never gets it however much food happens to be on the
		// market shelf that week.
		Signature: habit.Signature{habit.Curious: 1, habit.Industry: 0.4},
		Cost:      1.5 * pressYear,
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
		// Lonely, and among people while being it - which is the particular
		// misery a tavern answers and is not the same as being alone. A
		// settlement of hermits never builds one however much grain it has.
		Signature: habit.Signature{habit.Lonely: 1, habit.Company: 0.5},
		Cost:      pressYear,
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
	// What a settlement gets to only by having been somewhere a long time.
	// Each of these asks for masters, and a master is three times the
	// labour an ordinary hand is - so these are the far end of a run rather
	// than the middle of one, and a settlement that never made anybody
	// really good at anything never sees them at all. That is the point of
	// them: the catalog used to be a thing every settlement held entire by
	// year fifteen, and the top of it should be somewhere most of them
	// never get.
	{
		// Weaving answers the winter, and only a settlement that has
		// actually been caught out in one thinks of it. The condition is
		// the moment rather than the climate: it must be biting now and a
		// quarter of the people must be standing in it under a poor roof.
		// A settlement that housed itself early never learns to weave,
		// which is right - it solved the same problem another way.
		Tech: "weaving", Knowledge: 20, Craft: craft(entity.Crafting),
		Condition: func(w *world.World) bool {
			return exposed(w) && adept(w, entity.Crafting, entity.Apprentice, 30) >= 2
		},
		Effect: func(w *world.World) { w.Mods.Warmth *= 0.6 },
		Text:   "caught out in a hard winter, they learned to weave",
	},
	{
		// The arch is masonry's second thought, and it needs the stone
		// under it: quarrying first, then masons who have raised eighty
		// things and are masters of it.
		Tech: "the arch", Knowledge: 80, Craft: craft(entity.Building),
		Condition: func(w *world.World) bool {
			return w.Has("quarrying") && adept(w, entity.Building, entity.Master, 80) >= 2
		},
		Effect: func(w *world.World) {
			w.Mods.BuildEfficiency *= 1.3
			w.Mods.ShelterDecay *= 0.7
		},
		Text: "masons learned to turn an arch, and built to last",
	},
	{
		// Husbandry is what a people who have trapped out their woods do
		// next: keep the animal rather than hunt it. What separates it from
		// trapping is that a kept beast needs somewhere to be kept, so it
		// asks for builders and not only for tools - without that it fired
		// on the same day trapping did on eleven seeds of twelve, which is
		// not a second technology but a longer sentence about the first.
		Tech: "husbandry", Knowledge: 40,
		Condition: func(w *world.World) bool {
			return w.Has("trapping") && forestThin(w) && tooled(w) >= 3 &&
				adept(w, entity.Building, entity.Journeyman, 40) >= 2
		},
		Effect: func(w *world.World) { w.Mods.HuntYield *= 1.6 },
		Text:   "with the woods trapped out, they began to keep the beasts instead",
	},
	{
		// Medicine is writing's, because it is the one thing here that has
		// to be written down to be kept: what somebody worked out about a
		// fever is no use to the settlement if it dies with them.
		Tech: "medicine", Knowledge: 150, Craft: craft(entity.Scholarship),
		Condition: func(w *world.World) bool {
			return w.Has("writing") && adept(w, entity.Scholarship, entity.Master, 40) >= 1
		},
		Effect: func(w *world.World) { w.Mods.Healing *= 2 },
		Text:   "scholars wrote down what was known of tending the sick",
	},
	{
		// The plough is the last of them and the deepest in: it takes the
		// metal, and it takes three farmers who have each worked ground
		// three hundred times and are masters of it. Two at a hundred and
		// fifty was no bar at all - a settlement with metal already had
		// them, so the plough arrived on the same day metallurgy did on
		// six seeds of seven. Three hundred is past what any but a
		// seriously farming people reach.
		Tech: "the plough", Knowledge: 120, Craft: craft(entity.Farming),
		Condition: func(w *world.World) bool {
			return w.Has("metallurgy") && adept(w, entity.Farming, entity.Master, 300) >= 3
		},
		Effect: func(w *world.World) { w.Mods.FarmYield *= 1.4 },
		Text:   "smiths and farmers between them worked out the plough",
	},
}

// exposedShare is how much of a settlement must be out in the weather
// before the weather is something it is trying to solve, and coldBites how
// hard the cold must press to count as being out in it at all.
const (
	coldBites    = 0.5
	exposedShare = 0.25
	poorRoof     = 0.5
)

// exposed reports whether the cold is biting now and enough of the
// settlement is standing in it under a poor roof to be feeling it. It is a
// pressure of the moment like the rest of them, not a fact about the map:
// the same settlement is exposed in February and not in June.
func exposed(w *world.World) bool {
	if len(w.Agents) == 0 || w.Climate.Chill() < coldBites {
		return false
	}
	n := 0
	for _, a := range w.Agents {
		if a.Shelter < poorRoof {
			n++
		}
	}
	return float64(n) >= exposedShare*float64(len(w.Agents))
}

// Pressures a settlement can be under, which is what the later discoveries
// answer. None of them is a number anyone in the settlement reads; they are
// the state of the land around the market, and a settlement that has not
// worn its land finds no need to learn these things.

// nearMarket is how far around the market the land is looked at.
const nearMarket = 20

// meanOf averages f over the tiles near the market that ok picks out. Both
// are handed the tile and where it is kept, for the readings beside the map.
func meanOf(w *world.World, ok func(int, *world.Tile) bool, f func(int, *world.Tile) float64) (float64, int) {
	var sum float64
	n := 0
	// Only the window is walked, in the order a walk over the whole map
	// would have come to its tiles, so the sum lands in the same order.
	g, m := w.Grid, w.MarketPos
	x0, x1 := g.Columns(m.X, nearMarket)
	for y := max(0, m.Y-nearMarket); y <= min(g.H-1, m.Y+nearMarket); y++ {
		for dx := x0; dx <= x1; dx++ {
			p := g.Norm(entity.Pos{X: m.X + dx, Y: y})
			i := g.Index(p)
			t := &g.Tiles[i]
			if ok(i, t) {
				sum += f(i, t)
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
				if i := w.Grid.Index(p); w.Grid.Tiles[i].Is(ontology.Wood) {
					wild += w.Grid.Wild[i]
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
	_, n := meanOf(w, func(_ int, t *world.Tile) bool { return t.Is(ontology.Water) }, func(int, *world.Tile) float64 { return 0 })
	return n > 0
}

// driestField is the dampness below which the ground a settlement is farming
// is out of the river's reach. Damp runs from 1 on the valley floor to 0 at
// FloodDepth above it; settlements pick good land and farm at a median of
// 0.64 to 0.83, so this is the ground a people have been pushed onto rather
// than the ground they would choose.
const driestField = 0.65

// fieldsDry reports whether the fields near the market stand too far above
// the water to be fed by it.
//
// This asked whether the fields had gone poor, and it was never once true in
// sixty years on any seed. That was not a threshold set wrong: fields do not
// wear. Fertility is restored toward Rich by Fallow faster than farming
// takes it, so the mean fertility of a settlement's fields sits between 0.94
// and 1.00 of what the ground can hold, for the whole of a run. A condition
// asking for 0.45 was asking for a thing this world does not do, and
// irrigation was therefore unreachable from the day it was written.
//
// That fields do not wear may itself be worth changing - land that pushes
// back is most of what makes a settlement move on - but it is a change to
// what the ground does and not to what a discovery asks of it, and it wants
// a batch of its own.
//
// So irrigation answers the other thing irrigation is actually for. A field
// high above the river is dry whatever its soil is worth, and cutting a
// channel to it is the answer to that and to nothing else. It is a fact
// about where a settlement was pushed to farm rather than about how hard it
// has farmed, which also makes it a pressure that can come and go as the
// holding spreads uphill.
func fieldsDry(w *world.World) bool {
	damp, n := meanOf(w, func(_ int, t *world.Tile) bool { return t.Is(ontology.Field) },
		func(_ int, t *world.Tile) float64 { return max(0, min(1, 1-t.Drain/world.FloodDepth)) })
	return n >= 3 && damp < driestField
}

// forestGone reports whether the settlement has cleared much of the forest
// it was founded among.
// clearedShare is how bare a settlement's own ground must be beside the
// country at large before it counts as having cleared the woods it was
// founded among. Seven tenths: the settlements probed sat between a third
// and one and a fifth of the ambient, so this picks out the ones that have
// actually eaten into their surroundings and leaves the ones that settled
// somewhere sparse and changed nothing.
const clearedShare = 0.7

// forestGone reports whether the settlement has cleared the forest it was
// founded among.
//
// It used to ask whether the map had lost two fifths of its forest, which
// one settlement cannot do and never did: over sixty years the whole-map
// share bottomed out between 0.61 and 1.00 against a bar of 0.60, so
// forestry was as good as unreachable. Every other pressure in this file
// reads the ground near the market, for the reason usedForest gives - the
// woods a settlement lives off are the few tiles nearest it, and the country
// beyond stays full whatever it does. This one was reading the country.
//
// It reads the settlement's own ground against the country now, which is
// scale-free: a people who have cleared their surroundings stand out however
// wooded or bare the map they were dropped on, and a people who settled in a
// clearing and cut nothing do not.
func forestGone(w *world.World) bool {
	if len(w.Grid.Tiles) == 0 {
		return false
	}
	ambient := float64(w.Grid.Forest()) / float64(len(w.Grid.Tiles))
	if ambient < 0.01 {
		return false // no woods anywhere; nothing to have cleared
	}
	near, n := meanOf(w, func(int, *world.Tile) bool { return true }, func(_ int, t *world.Tile) float64 {
		if t.Is(ontology.Wood) {
			return 1
		}
		return 0
	})
	return n > 0 && near < clearedShare*ambient
}

// rockNear reports whether there is stone to cut near the market.
func rockNear(w *world.World) bool {
	_, ok := w.Grid.Nearest(w.MarketPos, nearMarket, func(_ entity.Pos, t *world.Tile) bool { return t.Affords(ontology.Stone) })
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

// pressYear is a year of one person being wholly in a moment: a projection
// of one, every day, for three hundred and sixty days. Nobody is ever wholly
// in a moment, so a Cost of one pressYear is a good deal more than a year -
// which is the point of measuring in pressure rather than in time.
const pressYear = float64(clock.Year)

// pressing is how hard the settlement's most affected person is feeling a
// moment, and who that is.
//
// It is the projection of that person's own situation onto the discovery's
// direction, which is one operation that says both of the things wanted:
// how well the moment matches, and how much of a moment it is. The maximum
// over everybody rather than the mean or the sum, because a thing is worked
// out by the person it is happening to hardest, and a settlement of five
// hundred comfortable people with one desperate one in it is a settlement
// where somebody is about to think of something. A mean would drown them.
func pressing(w *world.World, d *Discovery) (float64, *entity.Agent) {
	dir := habit.Unit(d.Signature)
	if habit.Norm(dir) < habit.Epsilon {
		return 0, nil
	}
	best, by := 0.0, (*entity.Agent)(nil)
	for _, a := range w.Agents {
		// Ties go to the earliest born, which is what keeps a run
		// reproducible: agents are walked in birth order and a later one
		// has to beat what it finds, not equal it.
		if p := habit.Dot(action.Shared(a, w), dir); p > best {
			best, by = p, a
		}
	}
	return best, by
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
		// The condition says the settlement is in a position to work the
		// thing out. What decides when it does is how hard somebody is
		// actually feeling the want, day after day, until it comes to
		// enough. A Cost of zero is the old behaviour and is had at once.
		var by *entity.Agent
		if d.Cost > 0 {
			push, who := pressing(w, &d)
			if w.Press(d.Tech, push) < d.Cost {
				continue
			}
			by = who
		}
		w.Unlock(d.Tech)
		d.Effect(w)
		w.Room()
		for _, name := range d.Opens {
			if i := action.Index(action.ByName(name)); i >= 0 {
				w.ReachFloor[i] = max(w.ReachFloor[i], action.Opened)
			}
		}
		// Naming whoever was most in the moment for it is most of what
		// this change buys anybody watching: a technology stops being
		// something that happened to a settlement and becomes something
		// that happened to a person in one.
		if by != nil {
			w.Emit(event.Discovered, by.ID, 0, "%s: %s (%s worst of all)", d.Tech, d.Text, by.Name)
		} else {
			w.Emit(event.Discovered, 0, 0, "%s: %s", d.Tech, d.Text)
		}
	}
}
