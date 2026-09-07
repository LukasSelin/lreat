package action

import (
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/habit"
	"lreat/core/need"
	"lreat/core/world"
)

// Making and keeping: the longer chain that turns what the land gives into
// things that last. Raw food spoils and feeds a little; cooked it keeps and
// feeds more. Stone is cut from outcrops and makes a house that stands and a
// granary that keeps the market's food. Metal, once known, makes tools that
// are worth something. Each step is far out of reach until the settlement
// has learned it.

const (
	// A cooking is a batch: cookBatch units of food over cookFuel of wood
	// make cookBatch meals. Fuel is dear on purpose. When a meal cost half a
	// unit of wood, four meals cost a house, and settlements that learned to
	// cook stopped building and stopped having children; measured over 96
	// seeds that took survivors from 86 to 68. A batch on a fifth of a unit
	// is what a hearth actually burns.
	cookBatch = 2.0
	cookFuel  = 0.2
	// cookReserve is the wood an agent keeps back from the hearth: fuel is
	// taken from what is left after a house's worth.
	cookReserve = 2.0
	// quarryWear is how much of a tool cutting stone uses up.
	quarryWear = 0.15
	// granaryStone and granaryWood are what a granary is built of, and
	// granaryKeeping how much of the market's spoilage each one stops.
	granaryStone   = 3
	granaryWood    = 2
	granaryKeeping = 0.5
	// granaryRadius is how far from the market a granary may stand.
	granaryRadius = 6
	// smeltStone and smeltWood are what a smelting takes, and smeltYield
	// how many tools' worth of metal it gives before quality.
	smeltStone = 1
	smeltWood  = 1
	smeltYield = 2
	// stoneHouse is how much better a house built with a stone in it is,
	// and toolFarming how much better a field worked with a tool is;
	// farmWearTool is what the tool loses to it.
	stoneHouse   = 1.5
	toolFarming  = 1.25
	farmWearTool = 0.02
)

func isRock(_ entity.Pos, t *world.Tile) bool { return t.Terrain == world.Rock }

// rockNear reports whether there is stone to cut within reach of p.
func rockNear(w *world.World, p entity.Pos, radius int) bool {
	_, ok := w.Grid.Nearest(p, radius, isRock)
	return ok
}

var Cook = &Def{
	Name: "cook", Ticks: 2, Target: hearth,
	Available: func(a *entity.Agent, w *world.World) bool {
		return a.Inventory[entity.Food] >= cookBatch && a.Inventory[entity.Wood] >= cookReserve+cookFuel &&
			hasPlace(hearth)(a, w) && known(a, w, "pottery", "cook")
	},
	Expect: func(a *entity.Agent, _ *world.World, _ entity.Pos) need.Levels {
		// A cooked meal feeds more than a raw one; that difference is what
		// cooking is worth today.
		return need.Levels{need.Physiological: (mealNourish - nourished) * cookBatch * foodValue(a) * 4, need.Esteem: 0.02}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		a.Inventory[entity.Food] -= cookBatch
		a.Inventory[entity.Wood] -= cookFuel
		a.Inventory[entity.Meals] += cookBatch
		a.AddSkill(entity.Crafting, 0.005)
		a.Needs.Add(need.Esteem, 0.02)
	},
}

func quarryYield(a *entity.Agent) float64 { return 0.5 + a.Skills[entity.Building] }

var Quarry = &Def{
	Name: "quarry", Ticks: 3,
	Available: func(a *entity.Agent, w *world.World) bool {
		return a.Inventory[entity.Tools] >= 0.5 && known(a, w, "quarrying", "quarry")
	},
	Target: func(a *entity.Agent, w *world.World) (entity.Pos, bool) {
		return w.Grid.Nearest(a.Pos, searchRadius, isRock)
	},
	Expect: func(a *entity.Agent, _ *world.World, _ entity.Pos) need.Levels {
		// Stone is for building; it is worth the safety of the house it
		// will go into, if the agent lacks one.
		return need.Levels{need.Safety: 0.06 * (1 - a.Shelter) * quarryYield(a), need.Esteem: 0.03}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		if w.Grid.At(a.Pos).Terrain != world.Rock {
			return
		}
		a.Inventory[entity.Stone] += quarryYield(a)
		a.Inventory[entity.Tools] = max(0, a.Inventory[entity.Tools]-quarryWear)
		a.AddSkill(entity.Building, 0.01)
		a.Needs.Add(need.Esteem, 0.03)
	},
}

// granarySite is a plot beside the market. Like a house it wants its own
// ground around it: a granary the carts cannot get round is no use to the
// market it keeps food for.
func granarySite(_ *entity.Agent, w *world.World) (entity.Pos, bool) {
	if p, ok := w.Grid.Nearest(w.MarketPos, granaryRadius, func(p entity.Pos, _ *world.Tile) bool {
		return w.Grid.RoomToBuild(p)
	}); ok {
		return p, true
	}
	return w.Grid.Nearest(w.MarketPos, granaryRadius, func(_ entity.Pos, t *world.Tile) bool { return t.Buildable() })
}

var BuildGranary = &Def{
	Name: "build granary", Ticks: 4, Target: granarySite,
	Available: func(a *entity.Agent, w *world.World) bool {
		return a.Inventory[entity.Stone] >= granaryStone && a.Inventory[entity.Wood] >= granaryWood && known(a, w, "masonry", "build granary")
	},
	Expect: func(_ *entity.Agent, w *world.World, _ entity.Pos) need.Levels {
		// A granary is a public good: the market keeps what it holds. Its
		// worth to the builder is standing, and a little safety in a
		// settlement that will not run short.
		return need.Levels{need.Esteem: 0.2, need.Safety: 0.05 * w.Mods.Keeping}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		t := w.Grid.At(a.Pos)
		if !t.Buildable() {
			return
		}
		t.Structure = world.Granary
		a.Inventory[entity.Stone] -= granaryStone
		a.Inventory[entity.Wood] -= granaryWood
		w.Mods.Keeping *= granaryKeeping
		a.Reputation += 0.2
		a.AddSkill(entity.Building, 0.03)
		a.Needs.Add(need.Esteem, 0.2)
		w.Emit(event.Built, a.ID, 0, "%s built a granary", a.Name)
	},
}

var Smelt = &Def{
	Name: "smelt", Ticks: 3, Target: forge,
	Available: func(a *entity.Agent, w *world.World) bool {
		return a.Inventory[entity.Stone] >= smeltStone && a.Inventory[entity.Wood] >= smeltWood &&
			hasPlace(forge)(a, w) && known(a, w, "metallurgy", "smelt")
	},
	Expect: func(a *entity.Agent, w *world.World, _ entity.Pos) need.Levels {
		q := craftQuality(a, w) * smeltYield
		return need.Levels{need.Esteem: 0.15 * q, need.Safety: 0.03 * q}
	},
	Apply: func(a *entity.Agent, w *world.World) {
		q := craftQuality(a, w) * smeltYield
		a.Inventory[entity.Stone] -= smeltStone
		a.Inventory[entity.Wood] -= smeltWood
		a.Inventory[entity.Tools] += q
		a.Reputation += 0.08 * q
		a.AddSkill(entity.Crafting, 0.02)
		a.Needs.Add(need.Esteem, 0.15*q)
	},
}

// Reach at birth for making and keeping. All begin far off.
const (
	reachCook    = 0.4
	reachQuarry  = 0.4
	reachGranary = 0.2
	reachSmelt   = 0.2
)

func init() {
	// Cooking belongs to the moment of surplus: a full larder, wood to
	// spare beyond a house's worth, and no hunger to speak of. It is a
	// keeping act, not a way of eating.
	seed(Cook, reachCook, habit.Signature{
		habit.Hunger: -0.5, habit.Food: 0.9, habit.Wood: 0.8, habit.Industry: 0.3, habit.Near: 0.5,
	})
	// Quarrying belongs to the unsheltered with a tool in hand and stone
	// near.
	seed(Quarry, reachQuarry, habit.Signature{
		habit.Unsafe: 0.5, habit.Shelter: -0.6, habit.Industry: 0.5, habit.Near: 0.5, habit.Skill: 0.3,
	})
	Quarry.Skilled = uses(entity.Building)
	// A granary belongs to those with standing to win and stone to spare,
	// in a settlement they mean to stay in.
	seed(BuildGranary, reachGranary, habit.Signature{
		habit.Unproven: 0.7, habit.Wood: 0.5, habit.Charity: 0.4, habit.Tradition: 0.4, habit.Skill: 0.4,
	})
	BuildGranary.Skilled = uses(entity.Building)
	// Smelting belongs to the skilled crafter with stone and fuel.
	seed(Smelt, reachSmelt, habit.Signature{
		habit.Unproven: 0.6, habit.Wood: 0.5, habit.Industry: 0.5, habit.Skill: 0.6,
	})
	Smelt.Skilled = uses(entity.Crafting)
}
