package world

import (
	"testing"

	"lreat/core/entity"
)

func TestStepTowardPrefersEasierGround(t *testing.T) {
	g := NewGrid(10, 10)
	from, to := entity.Pos{X: 2, Y: 5}, entity.Pos{X: 8, Y: 5}
	g.At(entity.Pos{X: 3, Y: 5}).Terrain = Forest // straight ahead, slow
	if p := g.StepToward(from, to); p != (entity.Pos{X: 3, Y: 4}) && p != (entity.Pos{X: 3, Y: 6}) {
		t.Fatalf("step = %v, want a detour around the forest", p)
	}
	g.At(entity.Pos{X: 3, Y: 4}).Terrain = Forest
	g.At(entity.Pos{X: 3, Y: 6}).Terrain = Forest
	if p := g.StepToward(from, to); p != (entity.Pos{X: 3, Y: 5}) {
		t.Fatalf("step = %v, want the straight line when every way is equally hard", p)
	}
}

func TestTravelCostRisesWithHardGround(t *testing.T) {
	g := NewGrid(10, 3)
	from, to := entity.Pos{X: 0, Y: 1}, entity.Pos{X: 5, Y: 1}
	open := g.TravelCost(from, to)
	if open != 5 {
		t.Fatalf("open travel cost = %v, want 5", open)
	}
	for y := 0; y < 3; y++ {
		g.At(entity.Pos{X: 3, Y: y}).Terrain = Water
	}
	if crossed := g.TravelCost(from, to); crossed <= open {
		t.Fatalf("travel cost across the river = %v, want more than %v", crossed, open)
	}
	if c := g.MoveCost(entity.Pos{X: -1, Y: 0}); !isInf(c) {
		t.Fatalf("off-map cost = %v, want infinite", c)
	}
}

func isInf(v float64) bool { return v > 1e308 }
