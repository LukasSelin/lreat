package world

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/ontology"
)

// Heavy is a fact with a consequence, not a label. A creel of stone is not a
// sack of grain, and neither is a length of timber; what tells the difference
// is the trait on the class rather than a number written out again here.
func TestWhatIsHeavyIsCarriedAsSuch(t *testing.T) {
	armfuls := func(g entity.Good, n float64) float64 {
		a := &entity.Agent{}
		a.Inventory[g] = n
		return Hauled(a)
	}
	heavy := float64(ontology.HeavyLoad)

	// Food is the ordinary armful and the unit everything else is read in.
	if got := armfuls(entity.Food, 3); got != 3 {
		t.Fatalf("three sacks of grain hauled as %v, want 3 armfuls", got)
	}
	// Stone and timber are both heavy, and for the same reason.
	if got := armfuls(entity.Stone, 3); got != 3*heavy {
		t.Fatalf("three of stone hauled as %v, want %v", got, 3*heavy)
	}
	if got := armfuls(entity.Wood, 3); got != 3*heavy {
		t.Fatalf("three lengths of timber hauled as %v, want %v", got, 3*heavy)
	}
	if armfuls(entity.Wood, 3) <= armfuls(entity.Food, 3) {
		t.Fatal("timber is no more to carry than grain")
	}

	// Load is the other question and is meant to stay coarse: it asks only
	// whether the walker may take to the water, where a crumb shuts one out
	// as firmly as a harvest.
	light, laden := &entity.Agent{}, &entity.Agent{}
	light.Inventory[entity.Food] = 3
	laden.Inventory[entity.Stone] = 3
	if light.Load() != laden.Load() {
		t.Fatalf("Load told grain from stone at %v and %v; that is Hauled's question",
			light.Load(), laden.Load())
	}

	// A mixed pack adds up without a special case, so anything the trees
	// stop calling heavy stops counting for two with no edit here.
	c := &entity.Agent{}
	c.Inventory[entity.Food] = 1
	c.Inventory[entity.Wood] = 1
	c.Inventory[entity.Stone] = 1
	if got, want := Hauled(c), 1+2*heavy; got != want {
		t.Fatalf("a pack of grain, timber and stone hauled as %v, want %v", got, want)
	}
}
