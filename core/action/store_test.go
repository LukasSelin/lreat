package action

import (
	"math"
	"testing"

	"lreat/core/entity"
	"lreat/core/ontology"
)

// A material is somewhere: the same class resolves to a number in the
// ground, in a pack, in a purse, or on the shelf, and moving it is the
// same act wherever it is.
func TestAMaterialIsSomewhere(t *testing.T) {
	w, a := shore(t)
	tile := w.Grid.At(entity.Pos{X: 1, Y: 1})
	tile.Wood = 0.7

	// Berries are food in a pack: the good is found up the tree.
	mine, ok := pack(a, ontology.Berries)
	if !ok {
		t.Fatal("berries should be something in a pack")
	}
	a.Inventory[entity.Food] = 2
	if mine.Held() != 2 {
		t.Fatalf("pack holds %v, want the food", mine.Held())
	}
	mine.Move(-3)
	if a.Inventory[entity.Food] != 0 {
		t.Fatal("a pack never holds less than nothing")
	}

	// Coin is the purse.
	purse, _ := pack(a, ontology.Coin)
	purse.Move(5)
	if a.Wealth != 5 {
		t.Fatalf("purse holds %v, want 5", a.Wealth)
	}

	// The market's shelf, and its bottomless purse.
	theirs, _ := shelf(w, ontology.Timber)
	theirs.Move(3)
	if w.Market.Stock[entity.Wood] != 3 {
		t.Fatal("the shelf should hold what was put on it")
	}
	till, _ := shelf(w, ontology.Coin)
	till.Move(-1e9)
	if !math.IsInf(till.Held(), 1) {
		t.Fatal("the market's coin should have no end")
	}

	// The ground: a stock where there is one, bottomless where there is
	// not.
	if got := soil(tile, ontology.Timber).Held(); got != 0.7 {
		t.Fatalf("the wood holds %v timber, want 0.7", got)
	}
	soil(tile, ontology.Timber).Move(-1)
	if tile.Wood != 0 {
		t.Fatal("the ground never holds less than nothing")
	}
	if !math.IsInf(soil(tile, ontology.Stone).Held(), 1) {
		t.Fatal("stone in the ground should have no end")
	}

	// A thing nobody can carry is in no pack and on no shelf.
	if _, ok := pack(a, ontology.Practice); ok {
		t.Fatal("a practice is nothing in a pack")
	}
	if _, ok := shelf(w, ontology.Person); ok {
		t.Fatal("a person is nothing on a shelf")
	}
}

// The trees say who holds what: the ground its stocks, a person and the
// market any material.
func TestTheTreesSayWhoHoldsWhat(t *testing.T) {
	for _, c := range []struct {
		holder, material *ontology.Class
		want             bool
	}{
		{ontology.Wood, ontology.Timber, true},
		{ontology.Wood, ontology.Stone, false},
		{ontology.Water, ontology.Fish, true},
		{ontology.Person, ontology.Coin, true},
		{ontology.Person, ontology.Meal, true},
		{ontology.Market, ontology.Tool, true},
		{ontology.Market, ontology.Practice, false},
	} {
		if got := ontology.Offers(c.holder, c.material); got != c.want {
			t.Errorf("%s holds %s: %v, want %v", c.holder.Name, c.material.Name, got, c.want)
		}
	}
}
