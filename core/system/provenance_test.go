package system

import (
	"math"
	"testing"

	"lreat/core/action"
	"lreat/core/entity"
	"lreat/core/habit"
	"lreat/core/need"
	"lreat/core/world"
)

// moved reports how far h moved toward or away from s on one coordinate:
// positive is toward.
func moved(before, after, s habit.Signature, k int) float64 {
	return math.Abs(before[k]-s[k]) - math.Abs(after[k]-s[k])
}

func TestFoodRemembersWhoGrewIt(t *testing.T) {
	w := fitWorld(21)
	a := w.SpawnAt("a", need.Neutral(), w.MarketPos)
	a.Inventory[entity.Food] = 0
	// Forage at the spot, faking the yield so no forest is needed.
	p := Commit(a, w, action.Forage, a.Pos)
	a.Inventory[entity.Food] += 2.4
	a.Plan.Remaining = 0
	Learn(a, w, p)
	if len(a.Larder) != 1 {
		t.Fatalf("larder holds %d harvests, want 1", len(a.Larder))
	}
	h := a.Larder[0]
	if h.Index != action.Index(action.Forage) || h.Situation != p.Situation || math.Abs(h.Left-2.4) > 1e-9 {
		t.Fatalf("harvest does not name the forage, its moment, and its yield: %+v", h)
	}
}

// eat has the agent eat one unit at the given fullness, through the builder,
// so the meal's reward reaches the larder.
func eat(w *world.World, a *entity.Agent, phys float64) {
	a.Needs[need.Physiological] = phys
	a.Inventory[entity.Food] = 1
	Commit(a, w, action.Eat, a.Pos)
	Act(w)
}

func TestAHarvestIsJudgedByAllItFed(t *testing.T) {
	w := fitWorld(22)
	a := w.SpawnAt("a", need.Neutral(), w.MarketPos)
	action.Imprint(a)
	farm := action.Index(action.Farm)
	moment := habit.Signature{habit.Hunger: -0.9, habit.Food: 0.9, habit.Near: 1}

	// One unit, eaten hungry, by an agent that expects nothing of a
	// harvest: the field is thanked, and the expectation rises.
	a.Larder = []habit.Harvest{{Step: habit.Step{Index: farm, Situation: moment}, Left: 1}}
	before := a.Habits[farm]
	eat(w, a, 0.05)
	if len(a.Larder) != 0 {
		t.Fatal("the last unit should settle the harvest")
	}
	if !(moved(before, a.Habits[farm], moment, habit.Food) > 0) {
		t.Fatal("a harvest that fed a hungry meal should pull the farm habit toward its moment")
	}
	if !(a.Harvest > 0) {
		t.Fatal("the sense of what a harvest brings should rise")
	}

	// A unit that fed a full belly, from an agent used to harvests that
	// brought more, is a poor harvest: the field is pushed away.
	a.Larder = []habit.Harvest{{Step: habit.Step{Index: farm, Situation: moment}, Left: 1}}
	a.Harvest = 0.3
	before = a.Habits[farm]
	eat(w, a, 0.95)
	if !(moved(before, a.Habits[farm], moment, habit.Food) < 0) {
		t.Fatal("a harvest that fed a needless meal should push the farm habit away from its moment")
	}
}

func TestABiggerHarvestIsThankedMore(t *testing.T) {
	w := fitWorld(23)
	a := w.SpawnAt("a", need.Neutral(), w.MarketPos)
	action.Imprint(a)
	farm, forage := action.Index(action.Farm), action.Index(action.Forage)
	moment := habit.Signature{habit.Hunger: 0.2, habit.Food: -0.4, habit.Near: 1}
	a.Harvest = 0.2 // used to harvests worth about one hungry meal
	// Start both habits from the same place, so the only difference between
	// them is what their harvests brought.
	a.Habits[forage] = a.Habits[farm]

	// The forest brought one unit; the field brought three. Each is eaten
	// hungry.
	a.Larder = []habit.Harvest{
		{Step: habit.Step{Index: forage, Situation: moment}, Left: 1},
		{Step: habit.Step{Index: farm, Situation: moment}, Left: 3},
	}
	f0, g0 := a.Habits[farm], a.Habits[forage]
	for i := 0; i < 4; i++ {
		eat(w, a, 0.05)
	}
	if len(a.Larder) != 0 {
		t.Fatalf("both harvests should have settled, %d left", len(a.Larder))
	}
	farmMoved := moved(f0, a.Habits[farm], moment, habit.Food)
	forageMoved := moved(g0, a.Habits[forage], moment, habit.Food)
	if !(farmMoved > forageMoved) {
		t.Fatalf("the field that fed three meals should be thanked more than the forest that fed one: farm %v, forage %v", farmMoved, forageMoved)
	}
	if !(farmMoved > 0) {
		t.Fatal("a harvest above expectation should be thanked")
	}
}

func TestRisingSafetyThanksTheRoofAndTheWatch(t *testing.T) {
	w := fitWorld(24)
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
	if !(moved(b0, a.Habits[build], moment, habit.Wood) > 0) {
		t.Fatal("rising safety should thank the roof")
	}
	if !(moved(g0, a.Habits[guard], moment, habit.Wood) > 0) {
		t.Fatal("rising safety should thank the watch")
	}
}

func TestAWatchIsRemembered(t *testing.T) {
	w := fitWorld(25)
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
