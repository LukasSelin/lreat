package action

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/world"
)

func TestRestAndMealsStepUnderARoofThatIsRightThere(t *testing.T) {
	w, a, _ := village(t)
	a.Inventory[entity.Food] = 2
	for _, d := range []*Def{Rest, Eat} {
		if p, _ := d.Target(a, w); p != a.Pos {
			t.Fatalf("%s with no roof near should happen here, got %v", d.Name, p)
		}
	}
	a.Home, a.HasHome = entity.Pos{X: 18, Y: 11}, true
	w.Grid.At(a.Home).Structure = world.House
	for _, d := range []*Def{Rest, Eat} {
		if p, _ := d.Target(a, w); p != a.Home {
			t.Fatalf("%s with home a step away should happen at home, got %v", d.Name, p)
		}
	}
	a.Pos = entity.Pos{X: 18, Y: 14}
	if p, _ := Rest.Target(a, w); p != a.Pos {
		t.Fatal("a rest is not worth even a short walk home")
	}
}

func TestTemperamentPicksBetweenHearthAndTavern(t *testing.T) {
	w, a, _ := village(t)
	a.Home, a.HasHome = entity.Pos{X: 18, Y: 11}, true
	w.Grid.At(a.Home).Structure = world.House
	tavern := entity.Pos{X: 19, Y: 11}
	w.Grid.At(tavern).Structure = world.Tavern
	a.Temperament.Warmth = 0.9
	if p, _ := Rest.Target(a, w); p != tavern {
		t.Fatalf("the warm rest at the tavern, got %v", p)
	}
	a.Temperament.Warmth = 0.1
	if p, _ := Rest.Target(a, w); p != a.Home {
		t.Fatalf("the cool rest at home, got %v", p)
	}
}

func TestARoofRestsBetterAndATableCheers(t *testing.T) {
	w, a, o := village(t)
	a.Needs[need.Physiological] = 0.5
	Rest.Apply(a, w)
	outside := a.Needs[need.Physiological] - 0.5
	a.Home, a.HasHome = a.Pos, true
	w.Grid.At(a.Home).Structure = world.House
	a.Needs[need.Physiological] = 0.5
	Rest.Apply(a, w)
	if !(a.Needs[need.Physiological]-0.5 > outside) {
		t.Fatal("a rest at home should restore more than one in the open")
	}

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
