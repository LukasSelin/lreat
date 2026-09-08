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

// A street can go round whatever is in its way; a river cannot be gone round.
// So where the fording is heavy enough, the crossing is where the timber goes
// even though the busiest ground in a settlement is always a street.
func TestACrossingComesBeforeAStreet(t *testing.T) {
	w := world.New(9)
	a := townsfolk(t, w)
	a.Inventory[entity.Wood] = 4

	// A ford one way, a well-walked lane the other, the lane busier - and
	// nothing else worn anywhere, so that the two tiles being compared are
	// the only two in the running. The way townsfolk wears beside the agent
	// is a third candidate otherwise, and on a map that put a wood there it
	// won on what a road through trees saves rather than on being walked.
	for i := range w.Grid.Tiles {
		w.Grid.Tiles[i].Traffic = 0
	}
	ford := entity.Pos{X: a.Pos.X - 2, Y: a.Pos.Y}
	w.Grid.At(ford).Terrain = world.Water
	for i := 0; i < 90; i++ {
		w.Grid.Tread(ford)
	}
	lane := entity.Pos{X: a.Pos.X + 2, Y: a.Pos.Y}
	for i := 0; i < 300; i++ {
		w.Grid.Tread(lane)
	}
	got, ok := Pave.Target(a, w)
	if !ok {
		t.Fatal("nowhere to build at all")
	}
	if got != ford {
		t.Fatalf("aimed at %v, want the ford at %v even though the lane is busier", got, ford)
	}

	// Without the timber for a bridge, the lane is what gets paved.
	a.Inventory[entity.Wood] = 1
	if got, _ := Pave.Target(a, w); got != lane {
		t.Fatalf("with only a length of wood, aimed at %v, want the lane at %v", got, lane)
	}
}
