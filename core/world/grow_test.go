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
	for terrain, c := range grounds {
		tile := &Tile{Terrain: terrain}
		if tile.Alive() != c.Has(ontology.Living) {
			t.Fatalf("terrain %d grows as %s, which the ontology does not call living", terrain, c.Path())
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
	tile := g.At(p)
	if tile.Alive() {
		t.Fatal("open grass carries a standing crop")
	}
	tile.Terrain = Forest
	if !tile.Alive() {
		t.Fatal("a wood carries nothing growing")
	}
	if tile.Grown(ontology.Timbering.Full()) != 0 {
		t.Fatal("a stand sown this tick is already grown")
	}
	tile.Age = ontology.Brush.Full()
	if tile.Grown(ontology.Brush.Full()) != 1 {
		t.Fatal("brush that has had its years is not grown")
	}
	if tile.Grown(ontology.Timbering.Full()) >= 1 {
		t.Fatal("timber comes on as fast as brush")
	}
	tile.Age = 10 * ontology.Timbering.Full()
	if tile.Grown(ontology.Timbering.Full()) != 1 {
		t.Fatal("an old wood is more than grown")
	}
	tile.Sow()
	if tile.Grown(ontology.Timbering.Full()) != 0 {
		t.Fatal("sowing did not start the stand over")
	}
}

// The woods a map is made with are old woods: a settlement founded among
// them has timber to build with on its first day.
func TestFoundingWoodsAreOldWoods(t *testing.T) {
	g := New(2).Grid
	for i := range g.Tiles {
		if t2 := &g.Tiles[i]; t2.Terrain == Forest && t2.Grown(ontology.Timbering.Full()) < 1 {
			t.Fatalf("a founding wood is only %.2f grown", t2.Grown(ontology.Timbering.Full()))
		}
	}
}
