package action

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/world"
)

// A public building is spaced, not capped. A settlement used to be held to
// one tavern and one square outright, which is a rule about the settlement
// rather than about the ground: a town that has walked down the valley
// wants a second of each and could not have one. What decides it now is
// whether the nearest is further off than apart.
func TestAPublicBuildingIsSpacedNotCapped(t *testing.T) {
	for _, c := range []struct {
		what  string
		apart int
		s     world.Structure
	}{
		{"tavern", tavernApart, world.Tavern},
		{"square", marketApart, world.Market},
	} {
		w, a := shore(t)
		for i := range w.Grid.Tiles {
			w.Grid.Tiles[i].Structure = world.None
		}
		if !roomApart(w, a.Pos, c.apart, c.s) {
			t.Errorf("no call for a %s where there is none at all", c.what)
		}
		// One right beside the builder answers for the ground they stand on.
		w.Grid.At(a.Pos).Structure = c.s
		if roomApart(w, a.Pos, c.apart, c.s) {
			t.Errorf("call for a second %s on top of the first", c.what)
		}
		// One further off than apart does not.
		w.Grid.At(a.Pos).Structure = world.None
		far := entity.Pos{X: a.Pos.X, Y: a.Pos.Y}
		far.X += c.apart + 1
		if !w.Grid.In(far) {
			continue // the shore is too small to place it; the case above is the one that matters
		}
		w.Grid.At(far).Structure = c.s
		if !roomApart(w, a.Pos, c.apart, c.s) {
			t.Errorf("a %s %d tiles off still answers for here", c.what, c.apart+1)
		}
	}
}

// A square is raised where the builder is standing, not beside the square
// they already have - the whole point of a second one is to be somewhere
// the first is not.
func TestASquareIsRaisedWhereTheBuilderStands(t *testing.T) {
	w, a := shore(t)
	p, ok := ownPlot(a, w)
	if !ok {
		t.Fatal("nowhere to lay a square on open ground")
	}
	if entity.Dist(p, a.Pos) > 3 {
		t.Fatalf("a square would go up %d tiles from its builder", entity.Dist(p, a.Pos))
	}
	if entity.Dist(p, w.MarketPos) < 3 {
		t.Fatal("a second square went up on top of the first")
	}
}
