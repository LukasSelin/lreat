package system

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/world"
)

// A house is kept by whoever lives in it, so while they live it stands.
func TestAKeptHouseStands(t *testing.T) {
	w := world.New(3)
	a := w.Spawn("keeper", w.RandomPersonality())
	home := entity.Pos{X: a.Pos.X, Y: a.Pos.Y}
	tile := w.Grid.At(home)
	tile.Terrain, tile.Structure, tile.Owner = world.Grass, world.House, a.ID
	for i := 0; i < 2000; i++ {
		Upkeep(w)
	}
	if w.Grid.At(home).Structure != world.House {
		t.Fatal("a house fell in with its owner alive in it")
	}
}

// What nobody keeps goes back to the land, and the ground it stood on is
// open to whoever comes next.
func TestAnEmptyHouseFallsIn(t *testing.T) {
	w := world.New(3)
	home := entity.Pos{X: 5, Y: 5}
	tile := w.Grid.At(home)
	tile.Terrain, tile.Structure, tile.Owner = world.Grass, world.House, entity.ID(999)
	fell := 0
	for i := 0; i < 4000 && w.Grid.At(home).Structure == world.House; i++ {
		Upkeep(w)
		fell = i
	}
	if w.Grid.At(home).Structure != world.None {
		t.Fatal("an empty house was still standing after four thousand ticks")
	}
	if !w.Grid.At(home).Buildable() {
		t.Fatalf("the ground under a fallen house is not open again: %+v", *w.Grid.At(home))
	}
	if fell < 20 {
		t.Fatalf("an empty house fell in after %d ticks; ruin should take a while", fell)
	}
}

// A field is a field because somebody works it. Left alone it goes to grass,
// and the fertility the ground holds is still there for whoever clears it
// next.
func TestAnUnworkedFieldGoesBackToGrass(t *testing.T) {
	w := world.New(3)
	p := entity.Pos{X: 6, Y: 6}
	tile := w.Grid.At(p)
	tile.Terrain, tile.Owner, tile.Fertility = world.Field, entity.ID(999), 0.5
	for i := 0; i < 4000 && w.Grid.At(p).Terrain == world.Field; i++ {
		Upkeep(w)
	}
	if got := w.Grid.At(p); got.Terrain != world.Grass || got.Owner != 0 || got.Fertility != 0.5 {
		t.Fatalf("an abandoned field came back as %+v, want open grass still worth 0.5", *got)
	}
}

// A lapsed claim costs the world no luck.
//
// This is the one thing about ruin that is not about ruin. What is claimed
// and neither lived in nor sown goes at once, and a certainty needs no
// draw - so wither settles what becomes of a tile before it spends any
// chance on it. If it drew anyway, the number would be thrown away and
// every later draw in the run would come out somewhere else: the same seed
// would be a different settlement three generations on. Nothing in the
// world's own state would look wrong, which is why it is asserted here.
func TestALapsedClaimCostsNoLuck(t *testing.T) {
	w := world.New(3)
	for i, p := range []entity.Pos{{X: 3, Y: 3}, {X: 4, Y: 3}, {X: 5, Y: 3}} {
		tile := w.Grid.At(p)
		tile.Terrain, tile.Structure, tile.Owner = world.Grass, world.None, entity.ID(900+i)
	}
	before := w.RNG.Float64()
	w2 := world.New(3)
	for i, p := range []entity.Pos{{X: 3, Y: 3}, {X: 4, Y: 3}, {X: 5, Y: 3}} {
		tile := w2.Grid.At(p)
		tile.Terrain, tile.Structure, tile.Owner = world.Grass, world.None, entity.ID(900+i)
	}
	Upkeep(w2)
	if got := w2.RNG.Float64(); got != before {
		t.Fatalf("razing three lapsed claims spent the world's luck: next draw %v, want %v", got, before)
	}
	for _, p := range []entity.Pos{{X: 3, Y: 3}, {X: 4, Y: 3}, {X: 5, Y: 3}} {
		if w2.Grid.At(p).Owner != 0 {
			t.Fatalf("a claim nobody is left to press at %v is still standing", p)
		}
	}
}
