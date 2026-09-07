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

// knows walks the agent over each plot so that it has been appraised and
// carried home, which is the only way anybody comes to know anywhere.
func knows(w *world.World, a *entity.Agent, ps ...entity.Pos) {
	here := a.Pos
	for _, p := range ps {
		a.Pos = p
		Notice(a, w)
	}
	a.Pos = here
}

// tilth sets the ground at p and all round it, since an appraisal reads the
// best field a door could open onto rather than the dirt under the floor.
func tilth(w *world.World, p entity.Pos, v float64) {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			q := entity.Pos{X: p.X + dx, Y: p.Y + dy}
			if w.Grid.In(q) {
				t := w.Grid.At(q)
				t.Fertility, t.Rich = v, v
			}
		}
	}
}

// richen makes the ground at p worth living on, which on this flat test map
// is entirely a matter of soil.
func richen(w *world.World, p entity.Pos) { tilth(w, p, 1) }

// A house on ground as good as anything its owner knows of is a house they
// stay in. Nobody moves for the sake of moving.
func TestNobodyMovesOutOfAGoodHouse(t *testing.T) {
	w, _ := shore(t)
	home := entity.Pos{X: 8, Y: 2}
	richen(w, home)
	a := householder(t, w, home)
	knows(w, a, entity.Pos{X: 2, Y: 4}, entity.Pos{X: 3, Y: 5})
	if p, ok := MoveHouse.Target(a, w); ok {
		t.Fatalf("a householder on the best ground they know moved to %v", p)
	}
}

// Better ground is worth moving onto - but only ground the owner has been
// over. A house is not sited off a survey nobody carried out.
func TestAHouseMovesOntoBetterGround(t *testing.T) {
	w, _ := shore(t)
	good := entity.Pos{X: 3, Y: 4}
	richen(w, good)
	home := entity.Pos{X: 0, Y: 5}
	tilth(w, home, 0)
	a := householder(t, w, home)
	if _, ok := MoveHouse.Target(a, w); ok {
		t.Fatal("a householder moved onto ground they had never seen")
	}
	knows(w, a, good)
	p, ok := MoveHouse.Target(a, w)
	if !ok {
		t.Fatal("a householder who has seen better ground found nowhere better")
	}
	if p != good {
		t.Fatalf("moved to %v, want the good ground at %v", p, good)
	}
}

// A way worn through the parlour is the settlement saying a road wants to run
// exactly there. The house steps aside, which is how a street gets straight.
func TestAHouseInTheWayStepsAside(t *testing.T) {
	w, _ := shore(t)
	home := entity.Pos{X: 8, Y: 2}
	a := householder(t, w, home)
	aside := entity.Pos{X: 8, Y: 4}
	knows(w, a, aside)
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
	good := entity.Pos{X: 3, Y: 4}
	richen(w, good)
	tilth(w, old, 0)
	a := householder(t, w, old)
	knows(w, a, good)
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
