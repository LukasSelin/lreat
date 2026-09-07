package action

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/world"
)

// townsfolk puts an agent in the moment paving belongs to: housed, with wood
// past what the roof took, among neighbours, on ground people have worn.
func townsfolk(t *testing.T, w *world.World) *entity.Agent {
	t.Helper()
	a := blank(w, "settler")
	blank(w, "neighbour").Pos = a.Pos
	a.Shelter = 1
	a.HasHome = true
	a.Inventory[entity.Wood] = 3
	a.Inventory[entity.Food] = 4
	a.Needs = need.Levels{0.9, 0.9, 0.9, 0.5, 0.5}
	for i := 0; i < 120; i++ {
		w.Grid.Tread(entity.Pos{X: a.Pos.X + 1, Y: a.Pos.Y})
	}
	return a
}

// Nobody paves wilderness. Until the ground shows that people go that way,
// there is no road to be laid.
func TestNobodyPavesUntroddenGround(t *testing.T) {
	w := world.New(9)
	a := townsfolk(t, w)
	for i := range w.Grid.Tiles {
		w.Grid.Tiles[i].Traffic = 0
	}
	if _, ok := Pave.Target(a, w); ok {
		t.Fatal("an agent found somewhere worth paving with nowhere worn")
	}
}

// The way people already take is the way that gets made.
func TestPavingAimsAtTheWornWay(t *testing.T) {
	w := world.New(9)
	a := townsfolk(t, w)
	want := entity.Pos{X: a.Pos.X + 1, Y: a.Pos.Y}
	got, ok := Pave.Target(a, w)
	if !ok {
		t.Fatal("an agent standing beside a worn way found nowhere to pave")
	}
	if got != want {
		t.Fatalf("paving aimed at %v, want the worn tile at %v", got, want)
	}
}

// Laying a road takes timber, so somebody with none cannot.
func TestPavingTakesTimber(t *testing.T) {
	w := world.New(9)
	a := townsfolk(t, w)
	if !Pave.Available(a, w) {
		t.Fatal("an agent with wood in hand cannot lay a road")
	}
	a.Inventory[entity.Wood] = 0
	if Pave.Available(a, w) {
		t.Fatal("an agent with no wood at all can still lay a road")
	}
}

// Paving is a settled person's act. Somebody with no roof of their own has
// more pressing uses for the same timber, and should recognise those first.
func TestTheUnshelteredBuildBeforeTheyPave(t *testing.T) {
	w := world.New(9)
	a := townsfolk(t, w)
	a.Reach[Index(Pave)] = 1 // as if masonry had opened it
	a.Shelter, a.HasHome = 0, false
	a.Needs[need.Safety] = 0.3
	r := Rank(a, w)
	if r[0].Def == Pave {
		t.Fatalf("an unsheltered agent ranked paving first: %v", names(r))
	}
}
