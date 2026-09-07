package world

import (
	"testing"

	"lreat/core/entity"
)

// The tree line is a share of the land, read off the map rather than fixed:
// the ground that suits trees best will hold a wood and the rest will not,
// whatever shape the map happens to have.
func TestTheTreeLineCoversAShareOfTheLand(t *testing.T) {
	for seed := uint64(1); seed <= 4; seed++ {
		w := New(seed)
		g := w.Grid
		land, holds := 0, 0
		for i := range g.Tiles {
			if g.Tiles[i].Terrain == Water {
				continue
			}
			land++
			if g.HoldsWood(entity.Pos{X: i % g.W, Y: i / g.W}) {
				holds++
			}
		}
		if share := float64(holds) / float64(land); share < woodsShare*0.8 || share > woodsShare*1.3 {
			t.Fatalf("seed %d: %.2f of the land holds wood, want about %.2f", seed, share, woodsShare)
		}
	}
}

// Woods want damp ground. Across a slope running from the water's edge up to
// a dry shoulder, the wood stands on the damp end of it and stops: the line
// is where the ground stops suiting trees, not where the map ends.
func TestWoodsStopWhereTheGroundDries(t *testing.T) {
	g := NewGrid(20, 20)
	for i := range g.Tiles {
		// Dry in proportion to the distance from the water along the left.
		g.Tiles[i].Drain = 2 * FloodDepth * float64(i%20) / 19
	}
	g.readWoods()
	if !g.HoldsWood(entity.Pos{X: 0, Y: 10}) {
		t.Fatal("the ground at the water's edge will not hold a wood")
	}
	if g.HoldsWood(entity.Pos{X: 19, Y: 10}) {
		t.Fatal("the dry shoulder holds a wood")
	}
	// And the line sits where the share says, a fifth of the way up.
	edge := 0
	for x := 0; x < 20; x++ {
		if g.HoldsWood(entity.Pos{X: x, Y: 10}) {
			edge = x
		}
	}
	if edge < 2 || edge > 6 {
		t.Fatalf("the wood reaches x=%d up the slope, want about a fifth of 20", edge)
	}
}
