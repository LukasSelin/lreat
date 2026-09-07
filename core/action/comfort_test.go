package action

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/world"
)

func TestRestAndMealsHappenWhereTheAgentStands(t *testing.T) {
	w, a, _ := village(t)
	a.Inventory[entity.Food] = 2
	a.Home, a.HasHome = entity.Pos{X: 18, Y: 11}, true
	w.Grid.At(a.Home).Structure = world.House
	for _, d := range []*Def{Rest, Eat} {
		if p, _ := d.Target(a, w); p != a.Pos {
			t.Fatalf("%s should happen here even with home a step away, got %v", d.Name, p)
		}
	}
}

func TestATableInCompanyCheers(t *testing.T) {
	w, a, o := village(t)
	tavern := entity.Pos{X: 24, Y: 10}
	w.Grid.At(tavern).Structure = world.Tavern
	a.Pos, o.Pos = entity.Pos{X: 24, Y: 11}, entity.Pos{X: 25, Y: 11}
	a.Inventory[entity.Food] = 1
	b0 := a.Needs[need.Belonging]
	Eat.Apply(a, w)
	if !(a.Needs[need.Belonging] > b0) {
		t.Fatal("a meal in company at the tavern should be a little belonging")
	}
}
