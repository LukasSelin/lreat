package system

import (
	"testing"

	"lreat/core/world"
)

func woods(w *world.World) int {
	return w.Grid.Count(func(t *world.Tile) bool { return t.Terrain == world.Forest })
}

// Left entirely alone, the woods creep back over the ground that will hold
// them and stop there. Before the tree line they did not stop: over six
// thousand ticks a seed took wherever it landed until the forest had closed
// over three quarters of the map and there was nowhere left to break a field.
func TestTheWoodsCreepBackAndThenStop(t *testing.T) {
	w := world.New(2)
	g := w.Grid
	land := g.Count(func(t *world.Tile) bool { return t.Terrain != world.Water })
	// A settlement's worth of clearing, so that there is ground to creep
	// back over: every other wood is felled.
	cleared := 0
	for i := range g.Tiles {
		if t := &g.Tiles[i]; t.Terrain == world.Forest && i%2 == 0 {
			t.Terrain, t.Wood, t.Wild = world.Grass, 0, 0
			cleared++
		}
	}
	if cleared == 0 {
		t.Fatal("the map was made without woods to fell")
	}
	start := woods(w)

	for i := 0; i < 6000; i++ {
		w.Tick++
		w.Climate.Advance(w.Tick, w.RNG)
		Land(w)
	}

	grown := woods(w)
	if grown <= start {
		t.Fatalf("the woods went from %d to %d tiles, want them to creep back", start, grown)
	}
	if share := float64(grown) / float64(land); share > 0.3 {
		t.Fatalf("the woods took %.2f of the land, want them to stop at the tree line", share)
	}
}
