package system

import (
	"sort"
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
	for i := 0; i < 20*clock.Year && w.Grid.At(home).Structure == world.House; i++ {
		Upkeep(w)
	}
	if w.Grid.At(home).Structure != world.None {
		t.Fatal("an empty house was still standing after twenty years")
	}
	if !w.Grid.At(home).Buildable() {
		t.Fatalf("the ground under a fallen house is not open again: %+v", *w.Grid.At(home))
	}
	if median := ruinTimes(t, 21); median < clock.Season {
		t.Fatalf("half of twenty-one empty houses were down inside %d days; ruin should take a while", median)
	}
}

// ruinTimes stands n empty houses up on n maps and reports how long the
// middle one took to fall.
//
// One house is not enough to say anything with. Ruin is a draw against a
// rate, and at one chance in three years a house has better than a one in ten
// of being down inside the first season by luck alone - so the single house
// this used to be tested nothing but which way the world's dice had landed,
// and fell over the first time a change to the map moved them.
func ruinTimes(t *testing.T, n int) int {
	t.Helper()
	fell := make([]int, 0, n)
	for seed := 0; seed < n; seed++ {
		w := world.New(uint64(seed))
		home := entity.Pos{X: 5, Y: 5}
		tile := w.Grid.At(home)
		tile.Terrain, tile.Structure, tile.Owner = world.Grass, world.House, entity.ID(999)
		days := 0
		for ; days < 20*clock.Year && w.Grid.At(home).Structure == world.House; days++ {
			Upkeep(w)
		}
		fell = append(fell, days)
	}
	sort.Ints(fell)
	return fell[len(fell)/2]
}

// A field is a field because somebody works it. Left alone it goes to grass,
// and the fertility the ground holds is still there for whoever clears it
// next.
func TestAnUnworkedFieldGoesBackToGrass(t *testing.T) {
	w := world.New(3)
	p := entity.Pos{X: 6, Y: 6}
	tile, i := w.Grid.At(p), w.Grid.Index(p)
	tile.Terrain, tile.Owner, w.Grid.Fertility[i] = world.Field, entity.ID(999), 0.5
	for i := 0; i < 20*clock.Year && w.Grid.At(p).Terrain == world.Field; i++ {
		Upkeep(w)
	}
	if got := w.Grid.At(p); got.Terrain != world.Grass || got.Owner != 0 || w.Grid.Fertility[i] != 0.5 {
		t.Fatalf("an abandoned field came back as %+v worth %v, want open grass still worth 0.5", *got, w.Grid.Fertility[i])
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

// A public work stands a long time and then does not. Nothing ever owns a
// granary - the raising sets a structure and no owner - so before this it
// stood forever, and the settlement's keeping only ever improved.
func TestAGranaryFallsInEventually(t *testing.T) {
	w := world.New(3)
	a := w.Spawn("keeper", w.RandomPersonality())
	p := entity.Pos{X: 7, Y: 7}
	tile := w.Grid.At(p)
	tile.Terrain, tile.Structure = world.Grass, world.Granary
	stood := 0
	for i := 0; i < 200000 && w.Grid.At(p).Structure == world.Granary; i++ {
		Upkeep(w)
		stood = i
	}
	if w.Grid.At(p).Structure == world.Granary {
		t.Fatal("a granary nobody maintains stood for two thousand years")
	}
	// It is not kept by anybody, so the keeper being alive changes nothing.
	if !alive(w, a.ID) {
		t.Fatal("the test's keeper died, which is not what is being measured")
	}
	// A house goes in about three hundred ticks. A granary is built to last
	// and must not go at anything like that pace, or the settlement spends
	// its life rebuilding what it shares.
	if stood < 300 {
		t.Fatalf("a granary fell in after %d ticks, no better than a house", stood)
	}
}

func alive(w *world.World, id entity.ID) bool {
	for _, a := range w.Agents {
		if a.ID == id {
			return true
		}
	}
	return false
}
