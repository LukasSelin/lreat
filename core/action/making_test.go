package action

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/ontology"
	"lreat/core/world"
)

// workshop is the shore with an outcrop added and every making tech known.
func workshop(t *testing.T) (*world.World, *entity.Agent) {
	t.Helper()
	w, a := shore(t)
	for y := 3; y < 6; y++ {
		w.Grid.At(entity.Pos{X: 2, Y: y}).Terrain = world.Rock
	}
	for _, tech := range []world.Tech{"pottery", "quarrying", "masonry", "metallurgy"} {
		w.Unlock(tech)
	}
	// A house: a hearth, a forge, and a bench.
	a.Home, a.HasHome = entity.Pos{X: 4, Y: 1}, true
	w.Grid.At(a.Home).Structure, w.Grid.At(a.Home).Owner = world.House, a.ID
	return w, a
}

func TestCookingTurnsFoodAndFuelIntoMeals(t *testing.T) {
	w, a := workshop(t)
	spare := cookReserve + cookFuel
	a.Inventory[entity.Food], a.Inventory[entity.Wood] = 3, spare
	if !run(w, a, Cook) {
		t.Fatal("cooking should be possible with a batch of food and fuel to spare")
	}
	if a.Inventory[entity.Meals] != cookBatch || a.Inventory[entity.Food] != 3-cookBatch || a.Inventory[entity.Wood] != spare-cookFuel {
		t.Fatalf("after cooking: meals %v food %v wood %v", a.Inventory[entity.Meals], a.Inventory[entity.Food], a.Inventory[entity.Wood])
	}
	a.Inventory[entity.Wood] = cookReserve
	if Cook.Available(a, w) {
		t.Fatal("cooking should not burn the wood a house is built of")
	}
}

func TestAHungryAgentEatsMealsFirstAndTheyFeedMore(t *testing.T) {
	w, a := workshop(t)
	a.Needs[need.Physiological] = 0.1
	a.Inventory[entity.Food], a.Inventory[entity.Meals] = 1, 1
	meals, raw, restores := helping(a)
	if meals != 1 || raw <= 0 {
		t.Fatalf("a hungry agent should eat the meal and then some raw food: meals %v raw %v", meals, raw)
	}
	if !(restores > (meals+raw)*nourished) {
		t.Fatal("a cooked meal should feed more than the same raw food")
	}
	Eat.Apply(a, w)
	if a.Inventory[entity.Meals] != 0 {
		t.Fatal("the meal should be eaten")
	}
}

func TestQuarryingCutsStoneWithATool(t *testing.T) {
	w, a := workshop(t)
	if Quarry.Available(a, w) {
		t.Fatal("quarrying without a tool should not be possible")
	}
	a.Inventory[entity.Tools] = 1
	if !run(w, a, Quarry) {
		t.Fatal("quarrying should be possible with a tool and an outcrop")
	}
	if a.Inventory[entity.Stone] <= 0 || a.Inventory[entity.Tools] >= 1 {
		t.Fatalf("quarrying should give stone and wear the tool: stone %v tools %v", a.Inventory[entity.Stone], a.Inventory[entity.Tools])
	}
	if w.Grid.At(a.Pos).Terrain != world.Rock {
		t.Fatal("the agent should stand on the outcrop")
	}
}

func TestAGranaryKeepsTheMarketsFood(t *testing.T) {
	w, a := workshop(t)
	a.Inventory[entity.Stone], a.Inventory[entity.Wood] = granaryStone, granaryWood
	before := w.Mods.Keeping
	if !run(w, a, BuildGranary) {
		t.Fatal("a granary should be buildable with stone and wood beside the market")
	}
	if w.Grid.At(a.Pos).Structure != world.Granary {
		t.Fatal("a granary should stand where it was built")
	}
	if !(w.Mods.Keeping < before) {
		t.Fatal("a granary should keep the market's food")
	}
	if a.Inventory[entity.Stone] != 0 || a.Inventory[entity.Wood] != 0 {
		t.Fatal("a granary should cost its stone and wood")
	}
}

func TestSmeltingMakesMoreThanCrafting(t *testing.T) {
	w, a := workshop(t)
	a.Inventory[entity.Stone], a.Inventory[entity.Wood] = 1, 2
	if !run(w, a, Smelt) {
		t.Fatal("smelting should be possible with stone and fuel")
	}
	smelted := a.Inventory[entity.Tools]
	a.Inventory[entity.Tools] = 0
	Craft.Apply(a, w)
	if !(smelted > a.Inventory[entity.Tools]) {
		t.Fatalf("smelting should give more tools than crafting: %v vs %v", smelted, a.Inventory[entity.Tools])
	}
}

func TestAToolMakesTheFieldGoFurtherAndWears(t *testing.T) {
	w, a := workshop(t)
	run(w, a, Clear)
	season(w, ontology.Crop.Full())
	run(w, a, Farm)
	bare := a.Inventory[entity.Food]
	a.Inventory[entity.Food] = 0
	a.Inventory[entity.Tools] = 1
	season(w, ontology.Crop.Full())
	Farm.Apply(a, w)
	if !(a.Inventory[entity.Food] > bare-2) { // the first farm claimed the field from 2 food
		t.Fatalf("a tool should make the field go further: %v with, %v without", a.Inventory[entity.Food], bare-2)
	}
	if !(a.Inventory[entity.Tools] < 1) {
		t.Fatal("the tool should wear")
	}
}

func TestStoneMakesABetterHouse(t *testing.T) {
	w, a := workshop(t)
	a.Inventory[entity.Wood] = raisingTimber
	run(w, a, BuildShelter)
	wood := a.Shelter
	b := w.SpawnAt("b", need.Neutral(), entity.Pos{X: 4, Y: 4})
	b.Skills[entity.Building] = a.Skills[entity.Building]
	b.Inventory[entity.Wood], b.Inventory[entity.Stone] = raisingTimber, 1
	run(w, b, BuildShelter)
	if !(b.Shelter > wood) {
		t.Fatalf("a stone house should be better: %v vs %v", b.Shelter, wood)
	}
	if b.Inventory[entity.Stone] != 0 {
		t.Fatal("the stone should be in the walls")
	}
}
