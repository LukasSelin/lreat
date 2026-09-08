package system

import (
	"testing"

	"lreat/core/clock"
	"lreat/core/entity"
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

	for i := 0; i < 15*clock.Year; i++ {
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

// A planting is not a wood. What goes in the ground gives nothing on the day
// it is put there; the brush under it comes back within a few years and the
// timber takes a lifetime, which is what makes planting an act for whoever
// comes after.
func TestAPlantedStandComesOnBrushFirst(t *testing.T) {
	w := world.New(4)
	// A steady growing season and no age of weather, so that what is
	// measured is the stand and not the map moving under it.
	w.Tick = 1
	w.Climate = world.Climate{Temp: world.Thrive}
	p := entity.Pos{X: w.MarketPos.X, Y: w.MarketPos.Y}
	tile := w.Grid.At(p)
	tile.Terrain, tile.Structure, tile.Wood, tile.Wild = world.Forest, world.None, 0, 0
	tile.Sow()

	for i := 0; i < 20; i++ {
		Land(w)
	}
	if tile.Wood > 0.05 || tile.Wild > 0.05 {
		t.Fatalf("a planting gives %.2f timber and %.2f wild food in its first year", tile.Wood, tile.Wild)
	}
	for i := 0; i < int(world.BrushAge); i++ {
		Land(w)
	}
	brush, timber := tile.Wild, tile.Wood
	if brush < 0.8 {
		t.Fatalf("brush is only %.2f grown after its whole span", brush)
	}
	if timber >= brush {
		t.Fatalf("timber (%.2f) came on as fast as brush (%.2f)", timber, brush)
	}
	if timber <= 0 {
		t.Fatal("the stand has made no timber at all")
	}
}
