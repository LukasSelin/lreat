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
	// Worn by people carrying things, because that is what wears a way worth
	// paving: an armful counts as a crossing, so hauling-1 armfuls make each
	// passage worth hauling and the count reads in the units the bars are
	// written in. See world.Haul.
	for i := 0; i < 200; i++ {
		w.Grid.Tread(entity.Pos{X: a.Pos.X + 1, Y: a.Pos.Y}, hauling-1)
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

	// A ford one way, a well-walked lane the other, the lane far busier - and
	// busier in a way the ford cannot answer. Everybody fording carries
	// nothing, because the water is shut to anybody who does not (SwimLoad),
	// while the lane is hauled along. That is the whole difficulty a crossing
	// is under: the errands a bridge would carry are going the long way round
	// and marking the ground somewhere else. See fordEnough.
	//
	// Nothing else is worn anywhere, so the two tiles being compared are the
	// only two in the running. The way townsfolk wears beside the agent is a
	// third candidate otherwise, and since the ground grew mountains the map
	// this seed makes puts a wood on it - where a road saves the walker more
	// than either of these does, so it won on the trees rather than on being
	// walked and the test read as a crossing losing to a street.
	for i := range w.Grid.Tiles {
		w.Grid.Tiles[i].Traffic = 0
	}
	// The two tiles are found rather than counted off from the agent. Fixed
	// offsets have now been wrong twice: the ground grew mountains and put a
	// wood on one of them, and then the rock underneath began to decide how
	// fast the water cuts, which moved the market onto one and a river onto
	// the other. Neither time was this test about any of that.
	ford, lane := clearGround(t, w, a.Pos, entity.Pos{}), entity.Pos{}
	lane = clearGround(t, w, a.Pos, ford)
	w.Grid.At(ford).Terrain = world.Water
	// The ford's count is written in terms of hauling and the lane's load is
	// not, which is the asymmetry itself: the lane can be worn harder by
	// carrying more, and the ford can only ever be worn by more journeys,
	// so what it takes to make a crossing's case has to rise with what a
	// laden crossing is worth elsewhere.
	for i := 0; i < 25*hauling; i++ {
		w.Grid.Tread(ford, 0)
	}
	for i := 0; i < 300; i++ {
		w.Grid.Tread(lane, hauling-1)
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

// clearGround is a tile near p that the test can do what it likes with: open
// ground nobody has built on or claimed, not the tile itself and not next to
// away, so that two of them can be told apart by anything reading the
// neighbourhood. It walks out in rings, so what comes back is as near as such
// ground gets.
func clearGround(t *testing.T, w *world.World, p, away entity.Pos) entity.Pos {
	t.Helper()
	for r := 2; r < 12; r++ {
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				if max(abs(dx), abs(dy)) != r {
					continue
				}
				q := entity.Pos{X: p.X + dx, Y: p.Y + dy}
				if !w.Grid.In(q) || !w.Grid.At(q).Buildable() {
					continue
				}
				if away != (entity.Pos{}) && w.Grid.Dist(q, away) < 3 {
					continue
				}
				return q
			}
		}
	}
	t.Fatalf("no clear ground within twelve tiles of %v", p)
	return entity.Pos{}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
