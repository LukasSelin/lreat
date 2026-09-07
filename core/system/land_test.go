package system

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/world"
)

// A holding is many times the ground a house stands on, so ground the dead
// hold on to is ground the living cannot plough. A field nobody is left to
// work closes over and is common again; a field whose farmer is alive is
// theirs however long they leave it.
func TestAFieldNobodyWorksGoesBackToGrass(t *testing.T) {
	w := fitWorld(3)
	farmer := w.SpawnAt("farmer", need.Neutral(), w.MarketPos)

	var dead, kept entity.Pos
	found := 0
	for y := 0; y < w.Grid.H && found < 2; y++ {
		for x := 0; x < w.Grid.W && found < 2; x++ {
			p := entity.Pos{X: x, Y: y}
			if !w.Grid.At(p).Buildable() {
				continue
			}
			tile := w.Grid.At(p)
			tile.Terrain = world.Field
			if found == 0 {
				dead, tile.Owner = p, farmer.ID+1000 // a farmer long buried
			} else {
				kept, tile.Owner = p, farmer.ID
			}
			found++
		}
	}
	if found < 2 {
		t.Fatal("no open ground to lay out two fields on")
	}

	for i := 0; i < 6000 && w.Grid.At(dead).Terrain == world.Field; i++ {
		Land(w)
	}
	if w.Grid.At(dead).Terrain == world.Field {
		t.Fatal("a field nobody is left to work stayed a field for a lifetime")
	}
	if w.Grid.At(dead).Owner != 0 {
		t.Fatal("ground that went back to grass is still claimed")
	}
	if k := w.Grid.At(kept); k.Terrain != world.Field || k.Owner != farmer.ID {
		t.Fatalf("a living farmer's field was taken from under them: %v", k.Terrain)
	}
}
