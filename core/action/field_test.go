package action

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/ontology"
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

// till is a season of a farmer working the land: breaking the next strip
// while the holding is short of what the household eats, and harvesting
// what they hold otherwise. Clearing and farming are two acts, but one
// livelihood.
func till(w *world.World, a *entity.Agent) bool {
	if run(w, a, Clear) {
		return true
	}
	season(w, ontology.Crop.Full()) // the crop comes on between one harvest and the next
	return run(w, a, Farm)
}

// A house is one tile. A holding is not: a family eats far more ground than
// it sleeps on, so a farmer goes on breaking strips beside the ones they have
// until the holding is as much land as the household lives off.
func TestAHoldingGrowsToWhatAFamilyEats(t *testing.T) {
	w, a := shore(t)
	for i := 0; i < 3*fieldTiles; i++ {
		if !till(w, a) {
			t.Fatalf("farming should be possible on open ground (season %d)", i)
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

// A holding is one farm, not eight fields. The harvest comes off all of it,
// so one harvest takes one harvest's worth out of the ground however much
// ground it came off - which is what lets the fallow keep up with it.
func TestAHoldingIsWorkedAsOneFarm(t *testing.T) {
	w, a := shore(t)
	for i := 0; i < 3*fieldTiles; i++ {
		till(w, a)
	}
	holding := held(t, w, a)
	before := make([]float64, len(holding))
	for i, p := range holding {
		before[i] = w.Grid.At(p).Fertility
	}
	season(w, ontology.Crop.Full())
	if !run(w, a, Farm) {
		t.Fatal("a farmer with a holding should always have somewhere to work")
	}
	if !a.Holds(a.Pos) {
		t.Fatalf("a full holding should be worked, not added to: went to %v", a.Pos)
	}
	taken := 0.0
	for i, p := range holding {
		worn := before[i] - w.Grid.At(p).Fertility
		if worn <= 0 {
			t.Fatalf("strip %v gave nothing to the harvest", p)
		}
		taken += worn
	}
	if taken > farmWear*1.001 {
		t.Fatalf("one harvest took %v out of the ground, more than the %v a harvest takes", taken, farmWear)
	}
}

// Fields go outside the built ground, not through it: a strip is never
// broken against somebody's wall, so the houses keep the gaps the lanes are
// laid along.
func TestAHoldingDoesNotPloughUpTheNeighbourhood(t *testing.T) {
	w, a := shore(t)
	if !run(w, a, Clear) {
		t.Fatal("clearing should be possible on open ground")
	}
	// Somebody builds two tiles from the farmer's first furrow: the holding
	// has to grow the other way.
	neighbour := w.Grid.At(entity.Pos{X: a.Field.X - 2, Y: a.Field.Y})
	neighbour.Structure, neighbour.Owner = world.House, 99
	for i := 0; i < 3*fieldTiles; i++ {
		till(w, a)
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

// A strip cut is bare ground, and bare ground is not cut again. A household
// with a holding works it in turn - the strip in ear, then the next - and one
// with a single strip waits for it. Nobody was told to rotate a crop; it
// falls out of the crop taking a season to come on.
func TestAHoldingIsWorkedStripByStrip(t *testing.T) {
	w, a := shore(t)
	for i := 0; i < 3*fieldTiles; i++ {
		till(w, a)
	}
	holding := held(t, w, a)
	if len(holding) < 2 {
		t.Fatal("this wants a holding of more than one strip")
	}

	// Every strip in ear: the farmer takes them one after another and never
	// the same one twice.
	season(w, ontology.Crop.Full())
	worked := map[entity.Pos]int{}
	for range holding {
		if !Farm.Available(a, w) {
			t.Fatalf("a holding in ear should be workable; worked %d of %d strips", len(worked), len(holding))
		}
		if !run(w, a, Farm) {
			t.Fatal("a ripe strip should be there to work")
		}
		worked[a.Pos]++
	}
	if len(worked) != len(holding) {
		t.Fatalf("the farmer worked %d strips of a holding of %d: %v", len(worked), len(holding), worked)
	}
	// And with the whole holding just cut, there is nothing to come back to.
	if Farm.Available(a, w) {
		t.Fatal("a holding cut to the ground still offers a harvest")
	}
	season(w, ontology.Crop.Full())
	if !Farm.Available(a, w) {
		t.Fatal("a season on, the holding should be in ear again")
	}
}
