package action

import (
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/habit"
	"lreat/core/need"
	"lreat/core/ontology"
	"lreat/core/world"
)

// The land's answers: what a settlement can do once the forest it was
// founded among has been picked thin, its fields have gone poor, or its woods
// have been cleared. Each is far out of reach until the settlement, under
// that pressure, discovers it; then it is a way of living the land can carry
// where the old one could not.

const (
	// fishTake is how much of a water tile's fish one fishing consumes, and
	// huntTake how much of a forest's wild food one hunt does. Hunting takes
	// more than foraging and gives more, which is why a forest hunted hard
	// empties fast.
	fishTake = 0.15
	huntTake = 0.3
	// toolWear is how much of a tool a hunt uses up.
	toolWear = 0.1
	// irrigationCost is the wood a channel takes, and irrigationGain how
	// much more a field can hold once watered.
	irrigationCost = 1
	irrigationGain = 0.25
	// waterReach is how far from water a field can be irrigated.
	waterReach = 8
	// plantRadius is how far from home an agent will go to plant.
	plantRadius = 10
)

func isWater(_ entity.Pos, t *world.Tile) bool { return t.Is(ontology.Water) }

// waterKinds is the ground that is water, asked of the trees.
var waterKinds = world.KindsOf(ontology.Water)

// known reports whether an agent can attempt one of the land's answers: the
// settlement has discovered it, or the agent has come near enough on its
// own, through study or teaching, to try it before anyone else has. The
// value rule has no reach gate of its own, so this is what keeps a starving
// value-mode agent from walking to a river nobody has fished.
func known(a *entity.Agent, w *world.World, tech world.Tech, name string) bool {
	if w.Has(tech) {
		return true
	}
	Imprint(a)
	return a.Reach[Index(ByName(name))] >= Opened
}

// bank is a walkable tile beside water with fish in it.
func bank(w *world.World) func(p entity.Pos, t *world.Tile) bool {
	return func(p entity.Pos, t *world.Tile) bool {
		if t.Is(ontology.Water) {
			return false
		}
		return bestWater(w, p) != nil
	}
}

// bestWater is the water tile beside p with the most fish, or nil.
func bestWater(w *world.World, p entity.Pos) *world.Tile {
	var best *world.Tile
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			q := entity.Pos{X: p.X + dx, Y: p.Y + dy}
			if !w.Grid.In(q) {
				continue
			}
			t := w.Grid.At(q)
			if t.Offers(ontology.Fish) >= 0.2 && (best == nil || t.Fish > best.Fish) {
				best = t
			}
		}
	}
	return best
}

func fishYield(a *entity.Agent, w *world.World, fish float64) float64 {
	return (0.3 + 0.7*fish) * (0.7 + 0.6*a.Skills[entity.Fishing]) * w.Mods.FishYield
}

var Fish = mover("take/fish@water")

func huntYield(w *world.World, wild float64) float64 {
	return (0.4 + 1.6*wild) * w.Mods.HuntYield
}

var Hunt = mover("take/game@wood")

// nearWater reports whether water lies within reach of p.
func nearWater(w *world.World, p entity.Pos) bool {
	_, ok := w.Grid.NearestOfKind(p, waterReach, waterKinds, isWater)
	return ok
}

// thirsty is the strip of a holding most worth cutting a channel to: the
// poorest ground the farmer holds that water can be brought to. A holding is
// watered a strip at a time, the way it was broken.
func thirsty(a *entity.Agent, w *world.World) (entity.Pos, bool) {
	var best entity.Pos
	rich := 1.0
	for _, p := range a.Parcel {
		if r := w.Grid.Rich[w.Grid.Index(p)]; r < rich && nearWater(w, p) {
			best, rich = p, r
		}
	}
	return best, rich < 1
}

var Irrigate = &Def{
	Name: "irrigate", Ticks: 4, Supply: supply(nil, []*ontology.Class{ontology.Timber}, false),
	Available: func(a *entity.Agent, w *world.World) bool {
		if !a.HasField || a.Inventory[entity.Wood] < irrigationCost || !known(a, w, "irrigation", "irrigate") {
			return false
		}
		_, ok := thirsty(a, w)
		return ok
	},
	Target: func(a *entity.Agent, w *world.World) (entity.Pos, bool) { return thirsty(a, w) },
	Expect: func(a *entity.Agent, w *world.World, _ entity.Pos) need.Levels {
		// Worth what the extra fertility will grow, over a few harvests.
		return need.Levels{need.Physiological: foodValue(a) * 0.7 * irrigationGain * (0.8 + 2*a.Skills[entity.Farming]) * w.Mods.FarmYield * 3, need.Esteem: 0.03}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		t := w.Grid.At(a.Pos)
		if !t.Is(ontology.Field) || t.Owner != a.ID {
			return
		}
		a.Inventory[entity.Wood] -= irrigationCost
		i := w.Grid.Index(a.Pos)
		w.Grid.Rich[i] = min(1, w.Grid.Rich[i]+irrigationGain)
		w.Grid.Fertility[i] = min(w.Grid.Rich[i], w.Grid.Fertility[i]+irrigationGain)
		a.Learn(entity.Farming, 0.01)
		a.Needs.Add(need.Esteem, 0.03)
		w.Emit(event.Built, a.ID, 0, "%s cut a channel to the field", a.Name)
	},
}

var PlantTrees = &Def{
	Name: "plant trees", Ticks: 2, Supply: supply([]*ontology.Class{ontology.Timber}, nil, false),
	Available: func(a *entity.Agent, w *world.World) bool { return known(a, w, "forestry", "plant trees") },
	Target: func(a *entity.Agent, w *world.World) (entity.Pos, bool) {
		anchor := a.Pos
		if a.HasHome {
			anchor = a.Home
		}
		return w.Grid.Nearest(anchor, plantRadius, func(p entity.Pos, t *world.Tile) bool {
			return t.Buildable() && w.Grid.HoldsWood(p)
		})
	},
	Expect: func(*entity.Agent, *world.World, entity.Pos) need.Levels {
		// Nothing for the planter today; the settlement gathers there later.
		return need.Levels{need.Esteem: 0.02, need.Actualization: 0.02}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		t := w.Grid.At(a.Pos)
		// Ground that will not hold a wood is ground where the planting comes
		// to nothing, however well meant: nobody grows a forest on a dry
		// shoulder by wanting one there.
		if !t.Buildable() || !w.Grid.HoldsWood(a.Pos) {
			return
		}
		// What is planted is a planting: no timber and nothing to forage for
		// years yet. The good of it goes to whoever is here when it is grown,
		// which is the whole of what the act is for.
		w.Grid.Turn(a.Pos, world.Forest)
		t.Wood, t.Wild = 0, 0
		w.Grid.Sow(w.Grid.Index(a.Pos))
		a.Needs.Add(need.Esteem, 0.02)
		a.Needs.Add(need.Actualization, 0.02)
	},
}

// Reach at birth for the land's answers. All begin far off; the discoveries
// that answer each pressure open them to everyone.
const (
	reachFish   = 0.4
	reachHunt   = 0.5
	reachWater  = 0.3
	reachForest = 0.3
)

func init() {
	// Fishing belongs to the hungry moment by the water, and to those who
	// have learned it.
	seed(Fish, reachFish, habit.Signature{
		habit.Hunger: 0.7, habit.Near: 0.6, habit.Skill: 0.3, habit.Lack: 0.6,
	})
	Fish.Skilled = uses(entity.Fishing)
	// Hunting belongs to the hungry moment with a tool in hand.
	seed(Hunt, reachHunt, habit.Signature{
		habit.Hunger: 0.8, habit.Near: 0.4, habit.Lack: 0.7,
	})
	// Irrigating belongs to the industrious farmer with wood to spare.
	seed(Irrigate, reachWater, habit.Signature{
		habit.Industry: 0.7, habit.Near: 0.5, habit.Skill: 0.4, habit.Stock: 0.5,
	})
	Irrigate.Skilled = uses(entity.Farming)
	// Planting belongs to a moment with no wood and a mind for those who
	// come after: it feeds nobody today.
	seed(PlantTrees, reachForest, habit.Signature{
		habit.Industry: 0.4, habit.Charity: 0.3, habit.Tradition: 0.4, habit.Near: 0.5, habit.Lack: 0.5,
	})
}
