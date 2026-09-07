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
