package system

import (
	"math"
	"testing"

	"lreat/core/action"
	"lreat/core/entity"
	"lreat/core/habit"
	"lreat/core/need"
)

func TestFoodRemembersWhoGrewIt(t *testing.T) {
	w := fitWorld(21)
	a := w.SpawnAt("a", need.Neutral(), w.MarketPos)
	a.Inventory[entity.Food] = 0
	// Forage at the spot, faking the yield so no forest is needed.
	p := Commit(a, w, action.Forage, a.Pos)
	a.Inventory[entity.Food] += 2.4
	a.Plan.Remaining = 0
	Learn(a, w, p)
	if len(a.Larder) != 2 {
		t.Fatalf("larder holds %d entries for 2.4 food, want 2", len(a.Larder))
	}
	if a.Larder[0].Index != action.Index(action.Forage) || a.Larder[0].Situation != p.Situation {
		t.Fatal("larder entry does not name the forage and its moment")
	}
}

func TestAMealThanksTheField(t *testing.T) {
	w := fitWorld(22)
	a := w.SpawnAt("a", need.Neutral(), w.MarketPos)
	action.Imprint(a)
	farm := action.Index(action.Farm)
	// A remembered farm, taken in a moment quite unlike the farm prior.
	moment := habit.Signature{habit.Hunger: -0.9, habit.Food: 0.9, habit.Near: 1}
	a.Larder = []habit.Step{{Index: farm, Situation: moment}}
	before := a.Habits[farm]

	// A hungry meal: better than eating usually is, so the field is thanked.
	a.Needs[need.Physiological] = 0.05
	a.Inventory[entity.Food] = 1
	Commit(a, w, action.Eat, a.Pos)
	Act(w)
	if len(a.Larder) != 0 {
		t.Fatal("eating should consume the remembered unit")
	}
	after := a.Habits[farm]
	if !(math.Abs(after[habit.Food]-moment[habit.Food]) < math.Abs(before[habit.Food]-moment[habit.Food])) {
		t.Fatalf("a good meal should pull the farm habit toward the moment it was grown in: %v -> %v", before[habit.Food], after[habit.Food])
	}

	// A meal at a full belly, from an agent used to hungry meals, is a
	// disappointment, and the field learns not to be worked in such moments.
	a.Larder = []habit.Step{{Index: farm, Situation: moment}}
	a.Baselines[action.Index(action.Eat)] = 0.3
	before = a.Habits[farm]
	a.Needs[need.Physiological] = 0.95
	a.Inventory[entity.Food] = 1
	Commit(a, w, action.Eat, a.Pos)
	Act(w)
	after = a.Habits[farm]
	if !(math.Abs(after[habit.Food]-moment[habit.Food]) > math.Abs(before[habit.Food]-moment[habit.Food])) {
		t.Fatalf("a needless meal should push the farm habit away from the moment: %v -> %v", before[habit.Food], after[habit.Food])
	}
}

func TestRisingSafetyThanksTheRoofAndTheWatch(t *testing.T) {
	w := fitWorld(23)
	a := w.SpawnAt("a", need.Neutral(), w.MarketPos)
	action.Imprint(a)
	build, guard := action.Index(action.BuildShelter), action.Index(action.Guard)
	moment := habit.Signature{habit.Unsafe: 0.9, habit.Wood: -0.9, habit.Near: 1}
	a.Roof, a.HasRoof = habit.Step{Index: build, Situation: moment}, true
	a.Watch, a.HasWatch = habit.Step{Index: guard, Situation: moment}, true
	b0, g0 := a.Habits[build], a.Habits[guard]

	a.Needs[need.Safety] = 0.2
	p := Commit(a, w, action.Rest, a.Pos)
	a.Needs[need.Safety] = 0.4 // the roof and the order did their work meanwhile
	a.Plan.Remaining = 0
	Learn(a, w, p)
	for _, c := range []struct {
		name   string
		before habit.Signature
		after  habit.Signature
	}{{"roof", b0, a.Habits[build]}, {"watch", g0, a.Habits[guard]}} {
		if !(math.Abs(c.after[habit.Wood]-moment[habit.Wood]) < math.Abs(c.before[habit.Wood]-moment[habit.Wood])) {
			t.Fatalf("rising safety should thank the %s: %v -> %v", c.name, c.before[habit.Wood], c.after[habit.Wood])
		}
	}
}

func TestAWatchIsRemembered(t *testing.T) {
	w := fitWorld(24)
	a := w.SpawnAt("a", need.Neutral(), w.MarketPos)
	w.SpawnAt("b", need.Neutral(), w.MarketPos)
	p := Commit(a, w, action.Guard, w.MarketPos)
	for a.Plan != nil {
		Act(w)
	}
	if !a.HasWatch || a.Watch.Index != p.Index {
		t.Fatal("standing guard should be remembered as the last watch")
	}
}
