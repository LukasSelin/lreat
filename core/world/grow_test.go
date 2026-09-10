package world

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/ontology"
)

// The map's living ground and the ontology's Living sites are the same
// ground. If a site is declared to carry something growing, the world has to
// know which terrain that is, and nothing else may claim to.
func TestLivingGroundMatchesTheOntology(t *testing.T) {
	named := map[*ontology.Class]bool{}
	for _, c := range ontology.Site.Leaves() {
		if c.Has(ontology.Living) {
			named[c] = true
		}
	}
	for _, terrain := range Terrains() {
		c := terrain.Class()
		tile := &Tile{Terrain: terrain}
		if tile.Alive() != c.Has(ontology.Living) {
			t.Fatalf("%s grows as %s, which the ontology does not call living", terrain, c.Path())
		}
		delete(named, c)
	}
	for c := range named {
		t.Fatalf("%s is living in the ontology and is no terrain in the world", c.Path())
	}
}

// A stand comes on with age and then holds. Nothing here takes growth away
// again: what ends a stand is the ground being cleared and sown afresh.
func TestAStandComesOnAndHolds(t *testing.T) {
	g := NewGrid(3, 3)
	p := entity.Pos{X: 1, Y: 1}
	tile, i := g.At(p), g.Index(p)
	if tile.Alive() {
		t.Fatal("open grass carries a standing crop")
	}
	tile.Terrain = Forest
	if !tile.Alive() {
		t.Fatal("a wood carries nothing growing")
	}
	if g.Grown(i, ontology.Timbering.Full()) != 0 {
		t.Fatal("a stand sown this tick is already grown")
	}
	g.Age[i] = ontology.Brush.Full()
	if g.Grown(i, ontology.Brush.Full()) != 1 {
		t.Fatal("brush that has had its years is not grown")
	}
	if g.Grown(i, ontology.Timbering.Full()) >= 1 {
		t.Fatal("timber comes on as fast as brush")
	}
	g.Age[i] = 10 * ontology.Timbering.Full()
	if g.Grown(i, ontology.Timbering.Full()) != 1 {
		t.Fatal("an old wood is more than grown")
	}
	g.Sow(i)
	if g.Grown(i, ontology.Timbering.Full()) != 0 {
		t.Fatal("sowing did not start the stand over")
	}
}

// The woods a map is made with are old woods: a settlement founded among
// them has timber to build with on its first day.
func TestFoundingWoodsAreOldWoods(t *testing.T) {
	g := New(2).Grid
	for i := range g.Tiles {
		if g.Tiles[i].Terrain == Forest && g.Grown(i, ontology.Timbering.Full()) < 1 {
			t.Fatalf("a founding wood is only %.2f grown", g.Grown(i, ontology.Timbering.Full()))
		}
	}
}

// What the green reading says is what there is to be had, and it says it
// about the two kinds of stand differently because the two are taken
// differently. A wood is drawn down an armful at a time, so what is left is
// the count on the tile; a crop is cut once and wholly, so what is there is
// how far along it is. Both are the number an act already reads before it
// takes anything.
func TestGreenReadsWhatIsStandingAndNotWhatCouldBe(t *testing.T) {
	// One tile of each kind of ground, side by side on a strip of map.
	g := NewGrid(7, 1)
	kind := func(i int, terrain Terrain, wood, wild float64) int {
		g.Tiles[i].Terrain = terrain
		g.Tiles[i].Wood, g.Tiles[i].Wild = wood, wild
		return i
	}
	bare := kind(0, Grass, 0, 0)
	if g.Green(bare) != 0 {
		t.Errorf("open grass reads %v; it carries no crop anybody can take", g.Green(bare))
	}
	water := kind(1, Water, 0, 0)
	if g.Green(water) != 0 {
		t.Errorf("open water reads %v", g.Green(water))
	}

	// A wood with everything standing, and the same wood gathered out. It is
	// a wood on both days, and only one of them has anything on it.
	wood := kind(2, Forest, 1, 1)
	g.Standing(wood)
	if g.Green(wood) != 1 {
		t.Errorf("a full wood reads %v, want 1", g.Green(wood))
	}
	felled := kind(3, Forest, 0, 0)
	g.Standing(felled)
	if g.Green(felled) != 0 {
		t.Errorf("a wood gathered to nothing reads %v; the count is what a taking draws down", g.Green(felled))
	}

	// A strip keeps no count: what it has to give is how far it has come, so
	// it is bare the day it is sown and full when it is in ear.
	sown := kind(4, Field, 0, 0)
	g.Sow(sown)
	if g.Green(sown) != 0 {
		t.Errorf("a strip sown this morning reads %v, want 0", g.Green(sown))
	}
	ripe := kind(5, Field, 0, 0)
	g.Age[ripe] = ontology.Crop.Full()
	if g.Green(ripe) != 1 {
		t.Errorf("a strip in ear reads %v, want 1", g.Green(ripe))
	}
	if !(g.Green(ripe) > g.Green(sown)) {
		t.Error("a strip in ear should read greener than one just sown")
	}

	// A house on ground that used to grow something reads as nothing: what
	// is under a roof is not a crop.
	roofed := kind(6, Grass, 0, 0)
	g.Tiles[roofed].Structure = House
	if g.Green(roofed) != 0 {
		t.Errorf("a house reads %v", g.Green(roofed))
	}
}
