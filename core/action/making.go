package action

import (
	"lreat/core/entity"
	"lreat/core/habit"
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
	// taken from what is left after a house's worth. It has to follow the
	// price of a frame and not sit at some remembered number, or a person
	// saving up for walls burns the walls a meal at a time and never gets
	// past the third length of timber.
	cookReserve = raisingTimber
	// quarryWear is how much of a tool cutting stone uses up.
	quarryWear = 0.15
	// granaryStone and granaryWood are what a granary is built of, and
	// granaryKeeping what each one leaves of the market's spoilage. A
	// granary is a cold store as much as a dry one: it does for the food
	// it holds in August what the weather does for it in January, which is
	// why it earns its stone in a world that now has an August.
	granaryStone   = 3
	granaryWood    = 2
	granaryKeeping = 0.4
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

var Cook = product("make/provision+timber>meal@hearth")

func quarryYield(a *entity.Agent) float64 { return 0.5 + a.Skills[entity.Building] }

var Quarry = take("take/stone@outcrop")

var BuildGranary = raise("raise/timber+stone>granary@open")

var Smelt = product("make/stone+timber>tool@forge")

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
