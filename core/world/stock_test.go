package world

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/ontology"
)

// Heavy is a fact with a consequence, not a label. A creel of stone is not a
// sack of grain, and what tells the difference is the trait on the class
// rather than a number written out again here.
func TestStoneIsCarriedAsTheHeavyThingItIs(t *testing.T) {
	a := &entity.Agent{}
	a.Inventory[entity.Food] = 3
	if got := Hauled(a); got != 3 {
		t.Fatalf("three sacks of grain hauled as %v, want 3 armfuls", got)
	}

	b := &entity.Agent{}
	b.Inventory[entity.Stone] = 3
	if got, want := Hauled(b), 3*float64(ontology.HeavyLoad); got != want {
		t.Fatalf("three of stone hauled as %v, want %v", got, want)
	}
	if Hauled(b) <= Hauled(a) {
		t.Fatal("stone is no more to carry than grain")
	}

	// Load is the other question and is meant to stay coarse: it asks only
	// whether the walker may take to the water, where a crumb shuts it out
	// as firmly as a harvest.
	if a.Load() != b.Load() {
		t.Fatalf("Load told grain from stone at %v and %v; that is Hauled's question",
			a.Load(), b.Load())
	}

	// Every good is weighed, and anything the trees do not call heavy is an
	// ordinary armful, so a mixed pack adds up without a special case.
	c := &entity.Agent{}
	c.Inventory[entity.Food] = 1
	c.Inventory[entity.Wood] = 1
	c.Inventory[entity.Stone] = 1
	if got, want := Hauled(c), 2+float64(ontology.HeavyLoad); got != want {
		t.Fatalf("a pack of grain, timber and stone hauled as %v, want %v", got, want)
	}
}
