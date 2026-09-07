package action

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/world"
)

func TestAHomelessCrafterWorksAtTheMarketNotInTheWoods(t *testing.T) {
	w, a, _ := village(t)
	a.Inventory[entity.Wood] = 1
	p, ok := Craft.Target(a, w)
	if !ok || p != w.MarketPos {
		t.Fatalf("without a house the bench is a market stall, got %v %v", p, ok)
	}
	a.Pos = entity.Pos{X: 59, Y: 0} // far beyond the settlement
	if Craft.Available(a, w) || Study.Available(a, w) {
		t.Fatal("far from any settlement there is no bench and no desk")
	}
	a.Pos = entity.Pos{X: 18, Y: 10}
	a.Home, a.HasHome = entity.Pos{X: 18, Y: 8}, true
	if p, _ := Craft.Target(a, w); p != a.Home {
		t.Fatal("with a house the bench is at home")
	}
}

func TestCookingNeedsAHearth(t *testing.T) {
	w, a, _ := village(t)
	w.Unlock("pottery")
	a.Inventory[entity.Food], a.Inventory[entity.Wood] = 3, 3
	if Cook.Available(a, w) {
		t.Fatal("no house and no tavern is no hearth")
	}
	tavern := entity.Pos{X: 22, Y: 8}
	w.Grid.At(tavern).Structure = world.Tavern
	if !Cook.Available(a, w) {
		t.Fatal("the tavern's hearth should serve the homeless")
	}
	if p, _ := Cook.Target(a, w); p != tavern {
		t.Fatalf("cooking should happen at the tavern, got %v", p)
	}
}

func TestSmeltingNeedsAForgeOfOnesOwn(t *testing.T) {
	w, a, _ := village(t)
	w.Unlock("metallurgy")
	a.Inventory[entity.Stone], a.Inventory[entity.Wood] = 1, 1
	if Smelt.Available(a, w) {
		t.Fatal("nobody smelts on the common ground")
	}
	a.Home, a.HasHome = entity.Pos{X: 18, Y: 8}, true
	if !Smelt.Available(a, w) {
		t.Fatal("a house is a forge")
	}
}

func TestGuardingNeedsSomebodyAtTheMarket(t *testing.T) {
	w, a, o := village(t)
	if !Guard.Available(a, w) {
		t.Fatal("with somebody at the market there is something to guard")
	}
	o.Pos = entity.Pos{X: 30, Y: 10}
	if Guard.Available(a, w) {
		t.Fatal("an empty market is nothing to guard")
	}
	o.Pos = entity.Pos{X: 21, Y: 10}
	a.Pos = entity.Pos{X: 59, Y: 0}
	if Guard.Available(a, w) {
		t.Fatal("a market beyond the settlement's reach is not one's own to guard")
	}
}
