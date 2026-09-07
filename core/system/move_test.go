package system

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/world"
)

// An agent crossing water should take several ticks per tile and pay for the
// slog out of the physiological tier.
func TestHardGroundIsSlowAndTiring(t *testing.T) {
	w := world.NewSized(1, 12, 3)
	for i := range w.Grid.Tiles {
		w.Grid.Tiles[i].Terrain = world.Grass
	}
	a := w.SpawnAt("walker", need.Neutral(), entity.Pos{X: 0, Y: 1})
	a.Vitality, a.Health = 1, 1 // an ordinary body in good condition: one tick per grass tile
	a.Plan = &entity.Plan{Action: "rest", Target: entity.Pos{X: 4, Y: 1}, Remaining: 1, Total: 1}
	before := a.Needs[need.Physiological]
	for i := 0; i < 4; i++ {
		Act(w)
	}
	if a.Pos.X != 4 {
		t.Fatalf("over grass the walker reached x=%d after 4 ticks, want 4", a.Pos.X)
	}
	drain := before - a.Needs[need.Physiological]
	if drain <= 0 {
		t.Fatalf("walking cost %v of the physiological tier, want a positive drain", drain)
	}

	for y := 0; y < 3; y++ {
		w.Grid.At(entity.Pos{X: 6, Y: y}).Terrain = world.Water
	}
	a.Pos = entity.Pos{X: 5, Y: 1}
	a.Travel = 0
	a.Plan = &entity.Plan{Action: "rest", Target: entity.Pos{X: 7, Y: 1}, Remaining: 1, Total: 1}
	Act(w)
	if a.Pos.X != 5 {
		t.Fatalf("the walker forded the river in one tick, reaching x=%d", a.Pos.X)
	}
}

// A strong body walks faster and pays less for it than a frail one over
// exactly the same ground.
func TestStrongerBodiesTravelCheaper(t *testing.T) {
	w := world.NewSized(2, 20, 3)
	for i := range w.Grid.Tiles {
		w.Grid.Tiles[i].Terrain = world.Forest
	}
	target := entity.Pos{X: 8, Y: 1}
	hale := w.SpawnAt("hale", need.Neutral(), entity.Pos{X: 0, Y: 1})
	worn := w.SpawnAt("worn", need.Neutral(), entity.Pos{X: 0, Y: 1})
	hale.Vitality, hale.Health = 1.3, 1
	worn.Vitality, worn.Health = 0.7, 0.2
	for _, a := range []*entity.Agent{hale, worn} {
		a.Needs[need.Physiological] = 1
		a.Plan = &entity.Plan{Action: "rest", Target: target, Remaining: 1, Total: 1}
	}
	for i := 0; i < 12; i++ {
		Act(w)
	}
	if hale.Pos.X <= worn.Pos.X {
		t.Fatalf("hale reached x=%d, worn x=%d; the stronger body should be ahead", hale.Pos.X, worn.Pos.X)
	}
	if hale.Needs[need.Physiological] <= worn.Needs[need.Physiological] {
		t.Fatalf("hale spent %.3f, worn %.3f; the stronger body should have more left",
			1-hale.Needs[need.Physiological], 1-worn.Needs[need.Physiological])
	}
}

// Health is a record of how an agent has lived, so it moves slowly and lands
// where circumstances put it.
func TestHealthFollowsFeedingAndHousing(t *testing.T) {
	w := world.NewSized(3, 10, 3)
	fed := w.SpawnAt("fed", need.Neutral(), entity.Pos{X: 1, Y: 1})
	lean := w.SpawnAt("lean", need.Neutral(), entity.Pos{X: 2, Y: 1})
	fed.Health, lean.Health = 0.5, 0.5
	fed.Shelter = 1
	for i := 0; i < 200; i++ {
		fed.Needs[need.Physiological], lean.Needs[need.Physiological] = 1, 0
		lean.Shelter = 0
		Decay(w)
	}
	if fed.Health < 0.85 {
		t.Fatalf("a fed and housed agent settled at health %.2f, want it near whole", fed.Health)
	}
	if lean.Health > 0.5 {
		t.Fatalf("a starving unhoused agent held health %.2f, want it worn down", lean.Health)
	}
}

// The point of a road is what it saves whoever walks it. The same errand down
// a street should arrive sooner and take less out of the walker, which is what
// makes paving worth the labour and what will make a settlement's streets show
// up in the condition of its people.
func TestTheSameErrandIsCheaperOnAStreet(t *testing.T) {
	walk := func(paved bool) (int, float64) {
		w := world.NewSized(3, 30, 5)
		for i := range w.Grid.Tiles {
			w.Grid.Tiles[i].Terrain = world.Grass
		}
		to := entity.Pos{X: 20, Y: 2}
		if paved {
			for x := 0; x < 30; x++ {
				w.Grid.Pave(entity.Pos{X: x, Y: 2})
			}
		}
		a := w.SpawnAt("walker", need.Neutral(), entity.Pos{X: 0, Y: 2})
		a.Vitality, a.Health = 1, 1
		a.Needs[need.Physiological] = 1
		a.Plan = &entity.Plan{Action: "rest", Target: to, Remaining: 1, Total: 1}
		ticks := 0
		for a.Pos != to && ticks < 200 {
			Act(w)
			ticks++
		}
		if a.Pos != to {
			t.Fatalf("the walker never arrived (paved=%v)", paved)
		}
		return ticks, 1 - a.Needs[need.Physiological]
	}

	openTicks, openSpent := walk(false)
	roadTicks, roadSpent := walk(true)
	if roadTicks >= openTicks {
		t.Fatalf("the walk took %d ticks on the road and %d over grass; want the road quicker", roadTicks, openTicks)
	}
	if roadSpent >= openSpent {
		t.Fatalf("the walk cost %.4f on the road and %.4f over grass; want the road cheaper", roadSpent, openSpent)
	}
}
