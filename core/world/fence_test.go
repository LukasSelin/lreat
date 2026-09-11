package world

import (
	"math/rand/v2"
	"testing"

	"lreat/core/entity"
)

// sow turns a block of tiles into one farmer's field, the way the ground is
// turned and claimed in a settlement, so that the map's counts of what lies
// where keep up: the fence pass reads them to find the fields at all.
func sow(g *Grid, owner entity.ID, x0, y0, x1, y1 int) {
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			p := entity.Pos{X: x, Y: y}
			g.Turn(p, Field)
			g.Claim(p, owner)
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
			g.Turn(entity.Pos{X: x, Y: y}, Grass)
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

// hedgesWalkingAllTheGround is what the fence pass came to when it walked
// every awake tile: the definition the pass over the fields is held to.
func hedgesWalkingAllTheGround(g *Grid) []bool {
	seen := make([]bool, len(g.Tiles))
	out := make([]bool, len(g.Tiles))
	var block, stack []int32
	g.EachActive(nil, func(i, _ int, t *Tile) {
		if t.Terrain != Field || seen[i] {
			return
		}
		block, stack = block[:0], append(stack[:0], int32(i))
		seen[i] = true
		for len(stack) > 0 {
			j := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			block = append(block, j)
			p := g.PosOf(int(j))
			for d := range dirs {
				q := entity.Pos{X: p.X + dirs[d].X, Y: p.Y + dirs[d].Y}
				if !g.In(q) {
					continue
				}
				k := int32(g.Index(q))
				if seen[k] || g.Tiles[k].Terrain != Field {
					continue
				}
				seen[k] = true
				stack = append(stack, k)
			}
		}
		enclosed := len(block) >= fenceSize
		for _, j := range block {
			out[j] = enclosed
		}
	})
	return out
}

// The pass that visits only the fields comes to the same hedges as the one
// that walked all the ground, on a map with holdings scattered over it,
// and again after some of them are given up and others broken, so that the
// day's stamp on the ground already answered for is exercised too.
func TestFencingTheFieldsIsFencingAllTheGround(t *testing.T) {
	w := New(11)
	g := w.Grid
	rng := rand.New(rand.NewPCG(3, 4))
	strips := func(n int) {
		for k := 0; k < n; k++ {
			p := entity.Pos{X: rng.IntN(g.W), Y: rng.IntN(g.H)}
			if g.At(p).Buildable() {
				g.Turn(p, Field)
				g.Claim(p, entity.ID(1+k%9))
			}
		}
	}
	same := func(when string) {
		t.Helper()
		g.Fence()
		want := hedgesWalkingAllTheGround(g)
		for i := range g.Tiles {
			if g.Tiles[i].Fenced != want[i] {
				t.Fatalf("%s: tile %d fenced %v, walking all the ground says %v", when, i, g.Tiles[i].Fenced, want[i])
			}
		}
	}
	strips(400)
	same("first sown")
	for k := 0; k < 120; k++ {
		p := entity.Pos{X: rng.IntN(g.W), Y: rng.IntN(g.H)}
		if g.At(p).Terrain == Field {
			g.Raze(p)
		}
	}
	strips(150)
	same("given up and broken again")
}

// On a map with hundreds of patches of fields the fields are found by
// several goroutines at once and the blocks read off them in turn. However
// many goroutines, the hedges must be the ones walking all the ground
// would raise, and the ones one goroutine raises.
func TestFindingTheFieldsSideBySideRaisesTheSameHedges(t *testing.T) {
	was := Workers
	defer func() { Workers = was }()
	hedges := func(workers int) []bool {
		Workers = workers
		g := NewGrid(512, 256)
		strips := rand.New(rand.NewPCG(5, 6))
		for k := 0; k < 6000; k++ {
			p := entity.Pos{X: strips.IntN(g.W), Y: strips.IntN(g.H)}
			if g.At(p).Buildable() {
				g.Turn(p, Field)
				g.Claim(p, entity.ID(1+k%9))
			}
		}
		g.Fence()
		want := hedgesWalkingAllTheGround(g)
		out := make([]bool, len(g.Tiles))
		for i := range g.Tiles {
			out[i] = g.Tiles[i].Fenced
			if out[i] != want[i] {
				t.Fatalf("%d workers: tile %d fenced %v, walking all the ground says %v", workers, i, out[i], want[i])
			}
		}
		return out
	}
	alone := hedges(1)
	fenced := 0
	for _, f := range alone {
		if f {
			fenced++
		}
	}
	if fenced == 0 {
		t.Fatal("nothing was hedged; this exercised nothing")
	}
	for _, workers := range []int{2, 8, 24} {
		got := hedges(workers)
		for i := range alone {
			if got[i] != alone[i] {
				t.Fatalf("%d workers: tile %d fenced %v, one worker says %v", workers, i, got[i], alone[i])
			}
		}
	}
}
