package system

import (
	"math"
	"testing"

	"lreat/core/action"
	"lreat/core/belief"
	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/observe"
	"lreat/core/world"
)

func populate(w *world.World, n int) {
	for i := 0; i < n; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
}

func TestDeterminism(t *testing.T) {
	a, b := world.New(42), world.New(42)
	populate(a, 15)
	populate(b, 15)
	Run(a, 1500)
	Run(b, 1500)

	sa, sb := observe.Take(a), observe.Take(b)
	if !equalSnapshots(sa, sb) {
		t.Fatalf("same seed diverged:\n%+v\n%+v", sa, sb)
	}
	if math.IsNaN(sa.FoodPrice) || math.IsNaN(sa.WealthGini) {
		t.Fatalf("NaN leaked into state: %+v", sa)
	}
	if a.Log.Len() != b.Log.Len() {
		t.Fatalf("event counts differ: %d vs %d", a.Log.Len(), b.Log.Len())
	}
	la, lb := a.Log.All(), b.Log.All()
	for i := range la {
		if la[i] != lb[i] {
			t.Fatalf("event %d differs: %+v vs %+v", i, la[i], lb[i])
		}
	}
	for i := range sa.Map.Tiles {
		if sa.Map.Tiles[i] != sb.Map.Tiles[i] {
			t.Fatalf("map tile %d differs", i)
		}
	}
}

func equalSnapshots(a, b observe.Snapshot) bool {
	return a.Tick == b.Tick && a.Population == b.Population && a.MeanNeeds == b.MeanNeeds &&
		a.WealthGini == b.WealthGini && a.Knowledge == b.Knowledge && len(a.Techs) == len(b.Techs) &&
		a.Houses == b.Houses && a.Fields == b.Fields
}

func TestHungryAgentSeeksFood(t *testing.T) {
	w := world.New(1)
	populate(w, 3)
	a := w.Agents[0]
	a.Needs = need.Levels{0.05, 0.9, 0.9, 0.9, 0.9}
	a.Inventory[entity.Food] = 2
	if d, _ := Choose(a, w); d != action.Eat {
		t.Fatalf("starving agent with food chose %q, want eat", d.Name)
	}
	a.Inventory[entity.Food] = 0
	// Honest enough not to take the shortcut; theft is covered separately.
	a.Norms[belief.Honesty] = 1
	d, _ := Choose(a, w)
	if d != action.Forage && d != action.Farm {
		t.Fatalf("starving agent without food chose %q, want forage or farm", d.Name)
	}
}

func TestSatedAgentClimbsThePyramid(t *testing.T) {
	w := world.New(1)
	populate(w, 3)
	a := w.Agents[0]
	a.Needs = need.Levels{1, 1, 1, 1, 0.1}
	a.Inventory[entity.Food] = 4
	if d, _ := Choose(a, w); d != action.Study {
		t.Fatalf("agent with every lower tier met chose %q, want study", d.Name)
	}
}

func TestLonelyButHungryAgentStillEats(t *testing.T) {
	// The leak lets belonging register, but hunger must win when it is severe.
	w := world.New(1)
	populate(w, 3)
	a := w.Agents[0]
	a.Needs = need.Levels{0.1, 0.8, 0.0, 0.8, 0.8}
	a.Inventory[entity.Food] = 2
	if d, _ := Choose(a, w); d != action.Eat {
		t.Fatalf("hungry lonely agent chose %q, want eat", d.Name)
	}
}

func TestDistanceDiscountsActions(t *testing.T) {
	// Two identical agents want wood; the one next to a forest should score
	// gathering higher than the one far from it.
	w := world.NewSized(1, 60, 20)
	for i := range w.Grid.Tiles {
		w.Grid.Tiles[i] = world.Tile{Terrain: world.Grass, Fertility: 0.5}
	}
	w.Grid.At(entity.Pos{X: 0, Y: 10}).Terrain = world.Forest
	w.Grid.At(entity.Pos{X: 0, Y: 10}).Wood = 1
	near := w.SpawnAt("near", need.Neutral(), entity.Pos{X: 1, Y: 10})
	far := w.SpawnAt("far", need.Neutral(), entity.Pos{X: 59, Y: 10})
	for _, a := range []*entity.Agent{near, far} {
		a.Needs = need.Levels{1, 0.2, 1, 1, 1}
	}
	urg := need.Urgencies(near.Needs)
	target, ok := action.GatherWood.Target(near, w)
	if !ok {
		t.Fatal("no forest found")
	}
	sNear := Score(near, action.GatherWood.Expect(near, w, target), urg) / float64(action.GatherWood.Ticks+entity.Dist(near.Pos, target))
	sFar := Score(far, action.GatherWood.Expect(far, w, target), urg) / float64(action.GatherWood.Ticks+entity.Dist(far.Pos, target))
	if sNear <= sFar {
		t.Fatalf("near score %.4f should beat far score %.4f", sNear, sFar)
	}
}

func TestAgentsWalkBeforeActing(t *testing.T) {
	w := world.New(1)
	a := w.Spawn("a", need.Neutral())
	target := entity.Pos{X: a.Pos.X + 5, Y: a.Pos.Y}
	a.Plan = &entity.Plan{Action: "rest", Target: target, Remaining: 1, Total: 1}
	for i := 0; i < 5; i++ {
		Act(w)
		if a.Plan == nil {
			t.Fatalf("plan completed after %d steps, before arriving", i+1)
		}
	}
	if a.Pos != target {
		t.Fatalf("agent at %v, want %v", a.Pos, target)
	}
	Act(w)
	if a.Plan != nil {
		t.Fatal("plan should complete once at target")
	}
}

func TestCityDevelopsWithoutAPlayer(t *testing.T) {
	w := world.New(7)
	populate(w, 20)
	Run(w, 6000)
	if len(w.Agents) == 0 {
		t.Fatal("everyone starved")
	}
	if len(w.Techs()) == 0 {
		t.Fatalf("no discoveries in 6000 ticks; knowledge=%.1f", w.Knowledge)
	}
	s := observe.Take(w)
	if s.Houses == 0 || s.Fields == 0 {
		t.Fatalf("settlement left no footprint: houses=%d fields=%d", s.Houses, s.Fields)
	}
}
