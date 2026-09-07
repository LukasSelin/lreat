package action

import (
	"lreat/core/ontology"
	"testing"

	"lreat/core/entity"
	"lreat/core/world"
)

// houses returns every tile somebody has roofed.
func houses(w *world.World) []entity.Pos {
	var out []entity.Pos
	for y := 0; y < w.Grid.H; y++ {
		for x := 0; x < w.Grid.W; x++ {
			p := entity.Pos{X: x, Y: y}
			if w.Grid.At(p).Structure == world.House {
				out = append(out, p)
			}
		}
	}
	return out
}

// settle has n newcomers each raise a roof, in turn, with timber to spare.
func settle(t *testing.T, w *world.World, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		a := blank(w, "settler")
		a.Pos = w.MarketPos
		a.Inventory[entity.Wood] = raisingTimber + 1
		if !run(w, a, BuildShelter) {
			t.Fatalf("settler %d could not build", i)
		}
	}
}

// A settlement that builds wall to wall has nowhere left to put a street, so
// a house takes a plot with its own ground round it.
func TestHousesLeaveRoomForARoad(t *testing.T) {
	w, _ := shore(t)
	settle(t, w, 5)
	built := houses(w)
	if len(built) != 5 {
		t.Fatalf("five settlers raised %d roofs", len(built))
	}
	for i, p := range built {
		for j, q := range built {
			if i != j && entity.Dist(p, q) <= 1 {
				t.Fatalf("houses at %v and %v stand wall to wall", p, q)
			}
		}
	}
}

// The spacing is what a settlement wants, not what it owes: when there is no
// plot left within reach, a roof beside a neighbour's beats no roof at all.
func TestACrowdedSettlementStillGetsARoof(t *testing.T) {
	w, _ := shore(t)
	gap := entity.Pos{X: 4, Y: 2}
	for y := 0; y < w.Grid.H; y++ {
		for x := 0; x < w.Grid.W; x++ {
			p := entity.Pos{X: x, Y: y}
			if p != gap && w.Grid.At(p).Buildable() {
				w.Grid.At(p).Structure = world.House
			}
		}
	}
	settle(t, w, 1)
	if w.Grid.At(gap).Structure != world.House {
		t.Fatal("the last open tile in a crowded settlement went unbuilt")
	}
}

// A house is the largest thing anybody in a settlement makes, and it should
// read that way against a day's work in the woods. Before these prices a
// single tree housed a family twice over and everyone was under a roof
// within the first season; the frame has to be worth several trees, and
// several days of felling, or the founding years have no scarcity in them at
// all. Keeping the roof on afterwards is the opposite kind of act: an armful
// and an afternoon, so that owning a house does not become its own treadmill.
func TestAFrameCostsSeveralTreesAndAPatchDoesNot(t *testing.T) {
	// A forest tile carries one length of standing timber, and treeTake is
	// what one day's felling brings down of it.
	perTree := armful / treeTake
	if trees := raisingTimber / perTree; trees < 3 {
		t.Fatalf("a house's frame takes %.1f trees; it should take several", trees)
	}
	if days := raisingTimber / armful; days < 8 {
		t.Fatalf("a house's frame takes %.0f days in the woods; it should take many", days)
	}
	if roofingTimber > armful {
		t.Fatalf("patching a roof costs %v, more than the %v a day carries home", roofingTimber, armful)
	}
	// Wood has to stay something to be short of right up to the frame's
	// price, or nobody ever stands in front of enough timber to raise one.
	if knee(ontology.Timber) != raisingTimber || cookReserve != raisingTimber {
		t.Fatalf("wood knee %v and cook reserve %v should both be the frame's price %v",
			knee(ontology.Timber), cookReserve, raisingTimber)
	}
}

// Raising and keeping are one act to the builder and two prices to the
// forest: the frame is charged once, and only from somebody who has no house.
func TestRaisingIsChargedOnceAndKeepingIsCheap(t *testing.T) {
	w, _ := shore(t)
	a := blank(w, "a")
	a.Pos = w.MarketPos
	a.Inventory[entity.Wood] = roofingTimber
	if BuildShelter.Available(a, w) {
		t.Fatal("an armful should not raise a house")
	}
	a.Inventory[entity.Wood] = raisingTimber
	if !run(w, a, BuildShelter) {
		t.Fatal("a frame's worth of timber should raise one")
	}
	if !a.HasHome || a.Inventory[entity.Wood] != 0 {
		t.Fatalf("after raising: home %v, wood %v", a.HasHome, a.Inventory[entity.Wood])
	}
	a.Inventory[entity.Wood] = roofingTimber
	before := a.Shelter
	if !run(w, a, BuildShelter) {
		t.Fatal("an armful should patch the roof of a house that stands")
	}
	if !(a.Shelter > before) || a.Inventory[entity.Wood] != 0 {
		t.Fatalf("after patching: shelter %v from %v, wood %v", a.Shelter, before, a.Inventory[entity.Wood])
	}
}
