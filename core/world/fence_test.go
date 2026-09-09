package world

import (
	"testing"

	"lreat/core/entity"
)

// sow turns a block of tiles into one farmer's field.
func sow(g *Grid, owner entity.ID, x0, y0, x1, y1 int) {
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			t := g.At(entity.Pos{X: x, Y: y})
			t.Terrain, t.Owner = Field, owner
		}
	}
}

// A patch a farmer can see across is left open; a block of worked ground is
// hedged. Nobody decides either: it follows from how much of it there is.
func TestOnlyBigBlocksAreFenced(t *testing.T) {
	g := NewGrid(20, 10)
	sow(g, 1, 2, 2, 4, 2) // three strips in a row: one holding
	sow(g, 2, 10, 2, 12, 3)
	g.Fence()
	if g.At(entity.Pos{X: 3, Y: 2}).Fenced {
		t.Fatal("a holding of three strips was fenced; want it left open")
	}
	for y := 2; y <= 3; y++ {
		for x := 10; x <= 12; x++ {
			if !g.At(entity.Pos{X: x, Y: y}).Fenced {
				t.Fatalf("strip %d,%d of a six-strip block is not fenced", x, y)
			}
		}
	}
}

// Two holdings that touch are inside one fence. What is enclosed is the block
// of worked ground, not one man's title.
func TestNeighbouringHoldingsShareAFence(t *testing.T) {
	g := NewGrid(20, 10)
	sow(g, 1, 2, 2, 4, 2)
	sow(g, 2, 2, 3, 4, 3)
	g.Fence()
	if !g.At(entity.Pos{X: 3, Y: 2}).Fenced || !g.At(entity.Pos{X: 3, Y: 3}).Fenced {
		t.Fatal("two holdings lying together were not enclosed")
	}
}

// The ground under a fence is still ground: when the crop and the claim go,
// so does the hedge.
func TestGivingUpTheGroundTakesTheFence(t *testing.T) {
	g := NewGrid(20, 10)
	sow(g, 1, 2, 2, 4, 3)
	g.Fence()
	p := entity.Pos{X: 3, Y: 2}
	if !g.At(p).Fenced {
		t.Fatal("the block was not fenced to begin with")
	}
	g.Raze(p)
	if g.At(p).Fenced {
		t.Fatal("razed ground is still fenced")
	}
	sow(g, 0, 2, 2, 4, 3) // and with the owner gone, back to grass everywhere
	for y := 2; y <= 3; y++ {
		for x := 2; x <= 4; x++ {
			q := entity.Pos{X: x, Y: y}
			g.At(q).Terrain = Grass
		}
	}
	g.Fence()
	if g.At(p).Fenced {
		t.Fatal("grass is fenced")
	}
}

// The toll is on the line and not on the tile: crossing into the corn costs,
// walking about inside it does not, and the farmer whose corn it is has a
// gate.
func TestTheFenceIsChargedOnTheLine(t *testing.T) {
	g := NewGrid(20, 10)
	sow(g, 7, 4, 3, 6, 5)
	g.Fence()
	out, in, on := entity.Pos{X: 3, Y: 4}, entity.Pos{X: 4, Y: 4}, entity.Pos{X: 5, Y: 4}
	climbed := g.StepCost(out, in)
	inside := g.StepCost(in, on)
	if climbed <= inside+fenceToll-tie {
		t.Fatalf("stepping in cost %v against %v inside; want the toll on top", climbed, inside)
	}
	if leaving := g.StepCost(in, out); leaving <= g.MoveCost(out)+fenceToll-tie {
		t.Fatalf("stepping out cost %v; want the toll on the way out too", leaving)
	}
	if own := g.StepCostFor(out, in, 7); own >= climbed-tie {
		t.Fatalf("the farmer paid %v to enter their own field, a stranger %v", own, climbed)
	}
}

// What the fence is for: a walker crossing the settlement goes round the corn
// instead of through it, and the farmer whose corn it is still walks straight
// in at the gate.
func TestWalkersGoRoundAFencedHolding(t *testing.T) {
	g := NewGrid(20, 11)
	sow(g, 7, 6, 3, 10, 7)
	g.Fence()
	from, to := entity.Pos{X: 2, Y: 5}, entity.Pos{X: 14, Y: 5}
	for _, p := range g.Path(from, to) {
		if g.At(p).Fenced {
			t.Fatalf("a stranger's route went through the corn at %v", p)
		}
	}
	// And an errand that ends inside the corn: the stranger climbs in, the
	// farmer walks in.
	strip := entity.Pos{X: 8, Y: 5}
	stranger, farmer := g.TravelCost(from, strip), g.Holding(7).TravelCost(from, strip)
	if stranger <= farmer+fenceToll-tie {
		t.Fatalf("a stranger reached the strip for %v against the farmer's %v; want the fence between them", stranger, farmer)
	}
	for _, p := range g.Holding(7).Path(from, strip) {
		if g.At(p).Fenced && p != strip {
			return // straight in over their own ground
		}
	}
	t.Fatal("the farmer went round their own field to reach their own strip")
}

// A fence is a toll and not a wall. Where there is no way round, a walker
// climbs over and pays for it, so nothing on the map is ever shut off.
func TestAFenceIsNeverAWall(t *testing.T) {
	g := NewGrid(9, 9)
	sow(g, 7, 4, 0, 4, 8)
	g.Fence()
	from, to := entity.Pos{X: 1, Y: 4}, entity.Pos{X: 7, Y: 4}
	if c := g.TravelCost(from, to); isInf(c) {
		t.Fatal("a fenced strip across the map cut the far side off")
	}
	if len(g.Path(from, to)) == 0 {
		t.Fatal("no way through a fence with no way round it")
	}
}
