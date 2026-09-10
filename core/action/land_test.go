package action

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/ontology"
	"lreat/core/world"
)

// shore is a small world with a strip of water down the middle and a forest
// patch to the west, so every land action has somewhere to happen.
func shore(t *testing.T) (*world.World, *entity.Agent) {
	t.Helper()
	w := world.NewSized(3, 12, 6)
	w.Grid.Layers = world.NewLayers(len(w.Grid.Tiles))
	for i := range w.Grid.Tiles {
		w.Grid.Tiles[i] = world.Tile{Terrain: world.Grass}
		w.Grid.Fertility[i], w.Grid.Rich[i] = 0.3, 0.3
	}
	for y := 0; y < w.Grid.H; y++ {
		p := entity.Pos{X: 6, Y: y}
		w.Grid.At(p).Terrain, w.Grid.Fish[w.Grid.Index(p)] = world.Water, 1
	}
	for y := 0; y < 3; y++ {
		p := entity.Pos{X: 1, Y: y}
		i := w.Grid.Index(p)
		w.Grid.At(p).Terrain, w.Grid.Wood[i], w.Grid.Wild[i] = world.Forest, 1, 1
	}
	w.Grid.At(entity.Pos{X: 9, Y: 2}).Structure = world.Market
	w.MarketPos = entity.Pos{X: 9, Y: 2}
	a := w.SpawnAt("a", w.RandomPersonality(), entity.Pos{X: 4, Y: 2})
	for _, tech := range []world.Tech{"fishing", "trapping", "irrigation", "forestry"} {
		w.Unlock(tech)
	}
	return w, a
}

// run does the plan for d at its target: the agent is placed there and the
// act applied, which is what Act does once the walk is over.
// season lets what grows grow. It is the ripening system.Land runs every
// tick, which this package cannot call: a wood that has stood a while, a crop
// that has come on since it was sown.
func season(w *world.World, ticks float64) {
	for i := range w.Grid.Tiles {
		if t := &w.Grid.Tiles[i]; t.Alive() {
			w.Grid.Age[i] += ticks
		}
	}
}

func run(w *world.World, a *entity.Agent, d *Def) bool {
	if !d.Available(a, w) {
		return false
	}
	p, ok := d.Target(a, w)
	if !ok {
		return false
	}
	a.Pos = p
	d.Apply(a, w)
	return true
}

func TestForagingThinsTheForest(t *testing.T) {
	w, a := shore(t)
	before := w.Grid.Wild[w.Grid.Index(entity.Pos{X: 1, Y: 2})]
	if !run(w, a, Forage) {
		t.Fatal("forage should be possible beside a forest")
	}
	wild := w.Grid.Wild[w.Grid.Index(a.Pos)]
	if !(wild < before) {
		t.Fatalf("foraging should take from the forest: %v -> %v", before, wild)
	}
	if a.Inventory[entity.Food] <= 2 {
		t.Fatal("foraging should give food")
	}
	// A picked forest gives less.
	full, thin := forageYield(1), forageYield(0.1)
	if !(thin < 0.6*full) {
		t.Fatalf("a thin forest should give much less: %v vs %v", thin, full)
	}
}

func TestFarmingWearsAFieldAndFallowRestoresIt(t *testing.T) {
	w, a := shore(t)
	if !run(w, a, Clear) {
		t.Fatal("clearing should be possible on open ground")
	}
	if Farm.Available(a, w) {
		t.Fatal("a strip sown this morning should have no crop on it yet")
	}
	season(w, ontology.Crop.Full())
	if !run(w, a, Farm) {
		t.Fatal("farm should be possible on a field in ear")
	}
	f := w.Grid.Index(a.Field)
	rich := w.Grid.Rich[f]
	for i := 0; i < 10; i++ {
		season(w, ontology.Crop.Full())
		Farm.Apply(a, w)
	}
	if !(w.Grid.Fertility[f] < rich) {
		t.Fatalf("ten harvests should wear the field: %v of %v", w.Grid.Fertility[f], rich)
	}
	if w.Grid.Fertility[f] < wornField {
		t.Fatal("a field should not be worn below the floor")
	}
}

func TestFishingTakesFromTheWater(t *testing.T) {
	w, a := shore(t)
	if !run(w, a, Fish) {
		t.Fatal("fish should be possible near water")
	}
	if a.Inventory[entity.Food] <= 2 {
		t.Fatal("fishing should give food")
	}
	if bestWater(w, a.Pos) < 0 {
		t.Fatal("the agent should stand on a bank")
	}
	var least float64 = 2
	for y := 0; y < w.Grid.H; y++ {
		least = min(least, w.Grid.Fish[w.Grid.Index(entity.Pos{X: 6, Y: y})])
	}
	if !(least < 1) {
		t.Fatal("fishing should take from a water tile")
	}
	if a.Skills[entity.Fishing] <= 0 {
		t.Fatal("fishing should teach fishing")
	}
}

func TestHuntingNeedsToolsAndWearsThem(t *testing.T) {
	w, a := shore(t)
	if Hunt.Available(a, w) {
		t.Fatal("hunting without tools should not be possible")
	}
	a.Inventory[entity.Tools] = 1
	before := w.Grid.Wild[w.Grid.Index(entity.Pos{X: 1, Y: 2})]
	if !run(w, a, Hunt) {
		t.Fatal("hunting should be possible with tools and a forest")
	}
	if !(a.Inventory[entity.Tools] < 1) {
		t.Fatal("hunting should wear the tool")
	}
	if !(w.Grid.Wild[w.Grid.Index(a.Pos)] < before-forageTake) {
		t.Fatal("hunting should take more from the forest than foraging")
	}
	if !(huntYield(w, 1) > forageYield(1)) {
		t.Fatal("a full forest should give a hunter more than a forager")
	}
}

func TestIrrigationRaisesWhatAFieldCanHold(t *testing.T) {
	w, a := shore(t)
	run(w, a, Clear)
	f := w.Grid.Index(a.Field)
	if Irrigate.Available(a, w) {
		t.Fatal("irrigating without wood should not be possible")
	}
	a.Inventory[entity.Wood] = 2
	if !run(w, a, Irrigate) {
		t.Fatal("irrigating a field near water should be possible")
	}
	if !(w.Grid.Rich[f] > 0.3) || !(w.Grid.Fertility[f] > 0.3) {
		t.Fatalf("irrigation should raise the field: rich %v, fertility %v", w.Grid.Rich[f], w.Grid.Fertility[f])
	}
	if a.Inventory[entity.Wood] != 2-irrigationCost {
		t.Fatal("irrigation should cost wood")
	}
	for w.Grid.Rich[f] < 1 {
		a.Inventory[entity.Wood] = 2
		run(w, a, Irrigate)
	}
	if Irrigate.Available(a, w) {
		t.Fatal("a field that holds all it can should not be irrigated further")
	}
}

func TestPlantingMakesAForest(t *testing.T) {
	w, a := shore(t)
	before := w.Grid.Count(func(t *world.Tile) bool { return t.Terrain == world.Forest })
	if !run(w, a, PlantTrees) {
		t.Fatal("planting should be possible on open ground")
	}
	after := w.Grid.Count(func(t *world.Tile) bool { return t.Terrain == world.Forest })
	if after != before+1 {
		t.Fatalf("planting should add a forest tile: %d -> %d", before, after)
	}
	// A planting is a planting. There is nothing on it to forage and nothing
	// to fell, and the good of it goes to whoever is still here when it has
	// grown; the brush comes back first and the timber long after.
	i := w.Grid.Index(a.Pos)
	if w.Grid.Wild[i] > 0 || w.Grid.Wood[i] > 0 {
		t.Fatalf("a planting gives %v wild and %v timber the day it is put in", w.Grid.Wild[i], w.Grid.Wood[i])
	}
	season(w, ontology.Brush.Full())
	if w.Grid.Grown(i, ontology.Brush.Full()) < 1 {
		t.Fatal("a stand that has stood a brush's lifetime should be grown")
	}
	if w.Grid.Grown(i, ontology.Timbering.Full()) >= 1 {
		t.Fatal("timber should take longer than brush to come on")
	}
}

func TestTheLandsAnswersStartOutOfReach(t *testing.T) {
	for _, d := range []*Def{Fish, Hunt, Irrigate, PlantTrees} {
		if d.Reach0 >= 1 {
			t.Errorf("%s should begin out of reach", d.Name)
		}
	}
	if ForSkill(entity.Fishing) != Fish {
		t.Fatal("fishing should be the skill fish exercises")
	}
}
