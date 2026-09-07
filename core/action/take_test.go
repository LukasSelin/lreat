package action

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/ontology"
	"lreat/core/world"
)

// A taking is composed from a lode and a ground, not written: any material
// with a lode, at any site with a ground, is an act without code of its
// own. Timber does not lie in outcrops, so the ontology never entails
// this one; that it can be carried out all the same is the point.
func TestATakingIsComposedFromLodeAndGround(t *testing.T) {
	w, a := shore(t)
	for y := 3; y < 6; y++ {
		tile := w.Grid.At(entity.Pos{X: 2, Y: y})
		tile.Terrain, tile.Wood = world.Rock, 1
	}
	d := taking(ontology.Instance{Key: "take/timber@outcrop", Object: ontology.Timber, Site: ontology.Outcrop, Ticks: 2})
	if d == nil {
		t.Fatal("a lode at a ground should be an act")
	}
	if d.Name != "gather wood" || d.Ticks != 2 {
		t.Fatalf("composed %q over %d ticks", d.Name, d.Ticks)
	}
	if !run(w, a, d) {
		t.Fatal("the taking should find the outcrop")
	}
	if w.Grid.At(a.Pos).Terrain != world.Rock {
		t.Fatalf("stood on %v, not the outcrop", a.Pos)
	}
	if a.Inventory[entity.Wood] != armful {
		t.Fatalf("brought in %v wood, want an armful", a.Inventory[entity.Wood])
	}
	if w.Grid.At(a.Pos).Wood >= 1 {
		t.Fatal("the taking should draw the stock down")
	}
}

// What the ontology does not entail, nothing carries: a material with no
// lode is not an act, however the trees are arranged.
func TestAMaterialWithoutALodeIsNoAct(t *testing.T) {
	if d := taking(ontology.Instance{Object: ontology.Coin, Site: ontology.Wood}); d != nil {
		t.Fatalf("coin lies in no wood, yet %q was composed", d.Name)
	}
}
