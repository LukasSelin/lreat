package action

import (
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
		a.Inventory[entity.Wood] = 4
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
