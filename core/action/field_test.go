package action

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/world"
)

// held returns the tiles of an agent's holding, checked against the map: a
// strip is only theirs if the ground says so too.
func held(t *testing.T, w *world.World, a *entity.Agent) []entity.Pos {
	t.Helper()
	for _, p := range a.Parcel {
		tile := w.Grid.At(p)
		if tile.Terrain != world.Field || tile.Owner != a.ID {
			t.Fatalf("holding lists %v, which the map says is not theirs", p)
		}
	}
	return a.Parcel
}

// A house is one tile. A holding is not: a family eats far more ground than
// it sleeps on, so a farmer goes on breaking strips beside the ones they have
// until the holding is as much land as the household lives off.
func TestAHoldingGrowsToWhatAFamilyEats(t *testing.T) {
	w, a := shore(t)
	for i := 0; i < 3*fieldTiles; i++ {
		if !run(w, a, Farm) {
			t.Fatalf("farming should be possible on open ground (harvest %d)", i)
		}
	}
	holding := held(t, w, a)
	if len(holding) != fieldTiles {
		t.Fatalf("a farmer worked %d harvests and holds %d strips, wanted %d", 3*fieldTiles, len(holding), fieldTiles)
	}
	if len(holding) <= 1 {
		t.Fatal("a holding the size of a house is not a holding")
	}
	// The strips are one piece of ground, not a scatter of claims.
	for i, p := range holding[1:] {
		touching := false
		for _, q := range holding[:i+1] {
			if entity.Dist(p, q) <= 1 {
				touching = true
			}
		}
		if !touching {
			t.Fatalf("strip %v does not touch the rest of the holding", p)
		}
	}
}

// A holding large enough to rotate is a holding that does not wear out: the
// farmer turns to the richest ground they hold and the rest lies fallow
// while it waits.
func TestAFarmerWorksTheRichestStripAndRestsTheOthers(t *testing.T) {
	w, a := shore(t)
	for i := 0; i < fieldTiles; i++ {
		run(w, a, Farm)
	}
	worn := map[entity.Pos]bool{}
	rich := w.Grid.At(a.Field).Fertility
	for i := 0; i < fieldTiles; i++ {
		if !run(w, a, Farm) {
			t.Fatal("a farmer with a holding should always have somewhere to work")
		}
		if !a.Holds(a.Pos) {
			t.Fatalf("a full holding should be worked, not added to: went to %v", a.Pos)
		}
		worn[a.Pos] = true
	}
	if len(worn) != fieldTiles {
		t.Fatalf("%d harvests fell on %d of %d strips: the holding is not being rotated", fieldTiles, len(worn), fieldTiles)
	}
	if !(w.Grid.At(a.Field).Fertility < rich) {
		t.Fatal("harvests should wear the ground they come off")
	}
}

// Fields go outside the built ground, not through it: a strip is never
// broken against somebody's wall, so the houses keep the gaps the lanes are
// laid along.
func TestAHoldingDoesNotPloughUpTheNeighbourhood(t *testing.T) {
	w, a := shore(t)
	if !run(w, a, Farm) {
		t.Fatal("farming should be possible on open ground")
	}
	// Somebody builds two tiles from the farmer's first furrow: the holding
	// has to grow the other way.
	neighbour := w.Grid.At(entity.Pos{X: a.Field.X - 2, Y: a.Field.Y})
	neighbour.Structure, neighbour.Owner = world.House, 99
	for i := 0; i < 3*fieldTiles; i++ {
		run(w, a, Farm)
	}
	holding := held(t, w, a)
	if len(holding) != fieldTiles {
		t.Fatalf("a farmer beside a house holds %d strips, wanted %d: a holding goes round a neighbour, it does not stop at one", len(holding), fieldTiles)
	}
	for _, p := range holding {
		if w.Grid.HasNeighbor(p, (*world.Tile).Roofed) {
			t.Fatalf("a strip was broken at %v, against a neighbour's wall", p)
		}
	}
}
