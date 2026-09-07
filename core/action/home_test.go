package action

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/world"
)

// householder is somebody under a roof at home with timber to spare, which is the
// only moment a move is ever weighed in.
func householder(t *testing.T, w *world.World, home entity.Pos) *entity.Agent {
	t.Helper()
	a := blank(w, "householder")
	tile := w.Grid.At(home)
	tile.Structure, tile.Owner = world.House, a.ID
	a.Home, a.HasHome, a.Pos = home, true, home
	a.Shelter = 1
	a.Inventory[entity.Wood] = 2
	return a
}

// A house that suits its owner is a house they stay in.
func TestNobodyMovesOutOfAGoodHouse(t *testing.T) {
	w, _ := shore(t)
	a := householder(t, w, entity.Pos{X: 8, Y: 2}) // next door to the market
	if p, ok := MoveHouse.Target(a, w); ok {
		t.Fatalf("a householder beside the market moved to %v", p)
	}
}

// The far side of the settlement is worth leaving for the near side.
func TestAHouseMovesTowardTheLifeItsOwnerLeads(t *testing.T) {
	w, _ := shore(t)
	a := householder(t, w, entity.Pos{X: 0, Y: 5})
	p, ok := MoveHouse.Target(a, w)
	if !ok {
		t.Fatal("a householder in the far corner found nowhere better")
	}
	if entity.Dist(p, w.MarketPos) >= entity.Dist(a.Home, w.MarketPos)-worthMoving {
		t.Fatalf("moved to %v, no nearer the market than %v was", p, a.Home)
	}
}

// A way worn through the parlour is the settlement saying a road wants to run
// exactly there. The house steps aside, which is how a street gets straight.
func TestAHouseInTheWayStepsAside(t *testing.T) {
	w, _ := shore(t)
	home := entity.Pos{X: 8, Y: 2}
	a := householder(t, w, home)
	if _, ok := MoveHouse.Target(a, w); ok {
		t.Fatal("the house moved before anyone had walked through it")
	}
	for i := 0; i < 200; i++ {
		w.Grid.Tread(home)
	}
	p, ok := MoveHouse.Target(a, w)
	if !ok {
		t.Fatal("a house with a way worn through it stayed put")
	}
	if w.Grid.At(p).Traffic > w.Grid.At(home).Traffic {
		t.Fatalf("moved from worn ground at %v onto worse at %v", home, p)
	}
}

// Moving takes the roof with it: what it costs is the timber the taking down
// wastes and the day it takes, and what it leaves behind is open ground.
func TestMovingCarriesTheRoofAndFreesTheOldGround(t *testing.T) {
	w, _ := shore(t)
	old := entity.Pos{X: 0, Y: 5}
	a := householder(t, w, old)
	shelter, wood := a.Shelter, a.Inventory[entity.Wood]
	if !run(w, a, MoveHouse) {
		t.Fatal("a householder in the far corner could not move")
	}
	if a.Home == old {
		t.Fatal("the move left the house where it was")
	}
	if w.Grid.At(old).Structure != world.None || w.Grid.At(old).Owner != 0 {
		t.Fatalf("the old plot is still built on: %+v", *w.Grid.At(old))
	}
	if h := w.Grid.At(a.Home); h.Structure != world.House || h.Owner != a.ID {
		t.Fatalf("the new plot holds %+v, want their house", *h)
	}
	if a.Shelter != shelter {
		t.Fatalf("shelter %.2f after a move, want the %.2f they carried over", a.Shelter, shelter)
	}
	if a.Inventory[entity.Wood] != wood-movingWood {
		t.Fatalf("wood %.2f after a move, want %.2f", a.Inventory[entity.Wood], wood-movingWood)
	}
}

// A mover with no timber to waste stays where they are.
func TestMovingTakesTimber(t *testing.T) {
	w, _ := shore(t)
	a := householder(t, w, entity.Pos{X: 0, Y: 5})
	a.Inventory[entity.Wood] = 0
	if MoveHouse.Available(a, w) {
		t.Fatal("a householder with no timber can move house")
	}
}
