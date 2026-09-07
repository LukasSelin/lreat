package system

import (
	"testing"

	"lreat/core/clock"
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
	for i := 0; i < 20*clock.Year && w.Grid.At(home).Structure == world.House; i++ {
		Upkeep(w)
		fell = i
	}
	if w.Grid.At(home).Structure != world.None {
		t.Fatal("an empty house was still standing after twenty years")
	}
	if !w.Grid.At(home).Buildable() {
		t.Fatalf("the ground under a fallen house is not open again: %+v", *w.Grid.At(home))
	}
	if fell < clock.Season {
		t.Fatalf("an empty house fell in after %d days; ruin should take a while", fell)
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
	for i := 0; i < 20*clock.Year && w.Grid.At(p).Terrain == world.Field; i++ {
		Upkeep(w)
	}
	if got := w.Grid.At(p); got.Terrain != world.Grass || got.Owner != 0 || got.Fertility != 0.5 {
		t.Fatalf("an abandoned field came back as %+v, want open grass still worth 0.5", *got)
	}
}
