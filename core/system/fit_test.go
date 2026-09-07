package system

import (
	"math"
	"testing"

	"lreat/core/action"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/habit"
	"lreat/core/need"
	"lreat/core/observe"
	"lreat/core/world"
)

// fitWorld is a seeded world whose agents choose by recognition, which is
// the default; valueWorld is one whose agents choose by expected value, the
// original rule. The tests written for that rule use valueWorld so both
// rules stay covered.
func fitWorld(seed uint64) *world.World {
	w := world.New(seed)
	w.Rules.Fit = true
	return w
}

func valueWorld(seed uint64) *world.World {
	w := world.New(seed)
	w.Rules.Fit = false
	return w
}

func TestFitModeIsDeterministic(t *testing.T) {
	a, b := fitWorld(42), fitWorld(42)
	populate(a, 15)
	populate(b, 15)
	Run(a, 1500)
	Run(b, 1500)
	if !equalSnapshots(observe.Take(a), observe.Take(b)) {
		t.Fatal("same seed diverged under fit-based choice")
	}
	if a.Log.Len() != b.Log.Len() {
		t.Fatalf("event counts differ: %d vs %d", a.Log.Len(), b.Log.Len())
	}
	for i, x := range a.Agents {
		y := b.Agents[i]
		if x.Habits != y.Habits || x.Reach != y.Reach || x.Baseline != y.Baseline || x.Baselines != y.Baselines {
			t.Fatalf("agent %d learned differently in two identical worlds", x.ID)
		}
		for _, h := range x.Habits[:action.Count] {
			for _, v := range h {
				if math.IsNaN(v) {
					t.Fatalf("NaN in a habit of agent %d", x.ID)
				}
			}
		}
	}
}

func TestAGoodOutcomePullsTheHabitTowardTheMoment(t *testing.T) {
	w := fitWorld(3)
	a := w.SpawnAt("a", need.Neutral(), w.MarketPos)
	a.Needs[need.Physiological] = 0.05
	a.Inventory[entity.Food] = 4 // a full larder: the food coordinate reads +1
	p := Commit(a, w, action.Eat, a.Pos)
	if p.Situation[habit.Food] != 1 {
		t.Fatalf("food coordinate = %v, want 1", p.Situation[habit.Food])
	}
	before := a.Habits[action.Index(action.Eat)]
	Act(w)
	after := a.Habits[action.Index(action.Eat)]
	if !(after[habit.Food] > before[habit.Food]) {
		t.Fatalf("eating well should teach that full larders call for eating: %v -> %v", before[habit.Food], after[habit.Food])
	}
	if a.Baseline <= 0 {
		t.Fatalf("baseline should rise after a good outcome, got %v", a.Baseline)
	}
}

func TestAFailedPlanIsALesson(t *testing.T) {
	w := fitWorld(4)
	a := w.SpawnAt("a", need.Neutral(), w.MarketPos)
	a.Needs[need.Physiological] = 0.05
	a.Inventory[entity.Food] = 1 // the food coordinate reads -0.5
	p := Commit(a, w, action.Eat, a.Pos)
	a.Inventory[entity.Food] = 0 // someone took it while the plan was pending
	// An agent used to eating well finds a fruitless meal disappointing.
	// Within a bare Act nothing decays, so without that expectation the
	// outcome would be exactly nothing and carry no lesson either way.
	a.Baselines[p.Index] = 0.1
	before := a.Habits[p.Index]
	Act(w)
	if a.Plan != nil {
		t.Fatal("plan should have ended")
	}
	if a.Inventory[entity.Food] != 0 || a.Needs[need.Physiological] > 0.05 {
		t.Fatal("eat should have failed")
	}
	after := a.Habits[p.Index]
	s := p.Situation[habit.Food]
	if !(math.Abs(after[habit.Food]-s) > math.Abs(before[habit.Food]-s)) {
		t.Fatalf("a failed plan should push the habit away from the moment: %v -> %v (moment %v)", before[habit.Food], after[habit.Food], s)
	}
}

func TestPlansMadeOutsideTheBuilderTeachNothing(t *testing.T) {
	w := fitWorld(5)
	a := w.SpawnAt("a", need.Neutral(), w.MarketPos)
	action.Imprint(a)
	before := a.Habits
	Learn(a, w, &entity.Plan{Action: "rest"})
	if a.Habits != before {
		t.Fatal("a bare plan should carry no lesson")
	}
}

func TestReachGrowsByDoing(t *testing.T) {
	w := fitWorld(6)
	a := w.SpawnAt("a", need.Neutral(), w.MarketPos)
	Commit(a, w, action.Study, a.Pos)
	r0 := a.Reach[action.Index(action.Study)]
	for a.Plan != nil {
		Act(w)
	}
	if !(a.Reach[action.Index(action.Study)] > r0) {
		t.Fatal("studying should bring study further into reach")
	}
}

func TestIntensitySharpensWithNeed(t *testing.T) {
	w := fitWorld(7)
	a := w.SpawnAt("a", need.Neutral(), w.MarketPos)
	a.Needs = need.Levels{0.95, 0.95, 0.95, 0.95, 0.95}
	calm := Intensity(a)
	a.Needs = need.Levels{0.05, 0.3, 0.5, 0.3, 0.2}
	pressed := Intensity(a)
	if !(pressed > calm) {
		t.Fatalf("intensity calm=%v pressed=%v", calm, pressed)
	}
}

// Liveness under recognition. These mirror the value-mode tests and are the
// first thing to look at after any change to the priors or the learning
// constants. They log the same tallies so the two modes can be compared.
//
// They run long enough for the founders to die of old age, so passing means
// the settlement replaced itself. Before credit followed provenance it did
// not: it lived at subsistence, never reached the safety a birth needs,
// and was extinct within a generation. See docs/action-space.md.

func TestCityDevelopsByRecognition(t *testing.T) {
	w := fitWorld(7)
	populate(w, 20)
	Run(w, 6000)
	died := 0
	for _, e := range w.Log.All() {
		if e.Kind == event.Died {
			died++
		}
	}
	s := observe.Take(w)
	t.Logf("pop %d, died %d, houses %d, fields %d, techs %v, needs %v", s.Population, died, s.Houses, s.Fields, w.Techs(), s.MeanNeeds)
	if len(w.Agents) == 0 {
		t.Fatal("everyone starved")
	}
	if s.Houses == 0 || s.Fields == 0 {
		t.Fatalf("settlement left no footprint: houses=%d fields=%d", s.Houses, s.Fields)
	}
	if len(w.Techs()) == 0 {
		t.Fatalf("no discoveries in 6000 ticks; knowledge=%.1f", w.Knowledge)
	}
	if len(w.Agents) < 20 {
		t.Fatalf("the settlement did not replace its founders: %d alive after %d deaths", len(w.Agents), s.Deaths)
	}
}

func TestSocialLifeEmergesByRecognition(t *testing.T) {
	w := fitWorld(31)
	for i := 0; i < 25; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	Run(w, 6000)
	counts := map[event.Kind]int{}
	for _, e := range w.Log.All() {
		counts[e.Kind]++
	}
	t.Logf("requested %d, fulfilled %d, unmet %d, stolen %d, given %d, avenged %d",
		counts[event.Requested], counts[event.Fulfilled], counts[event.Unmet],
		counts[event.Stolen], counts[event.Given], counts[event.Avenged])
	if counts[event.Requested] == 0 {
		t.Fatal("no request was ever posted")
	}
	if counts[event.Fulfilled] == 0 {
		t.Fatal("no request was ever fulfilled")
	}
	if counts[event.Stolen] == 0 && counts[event.Given] == 0 {
		t.Fatal("the moral layer never engaged")
	}
	if honest := counts[event.Fulfilled] + counts[event.Given]; counts[event.Stolen] > honest*20 {
		t.Fatalf("theft dwarfs honest dealing: %d stolen vs %d fulfilled or given", counts[event.Stolen], honest)
	}
}

func TestFeudsFormByRecognition(t *testing.T) {
	w := fitWorld(1)
	for i := 0; i < 20; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	Run(w, 4000)
	var avenged int
	for _, e := range w.Log.All() {
		if e.Kind == event.Avenged {
			avenged++
		}
	}
	s := observe.Take(w)
	t.Logf("avenged %d, feuds %d", avenged, s.Feuds)
	if avenged == 0 {
		t.Fatal("nobody ever got even")
	}
}
