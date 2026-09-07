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
