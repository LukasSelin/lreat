package system

import (
	"math"
	"testing"

	"lreat/core/entity"
	"lreat/core/world"
)

// The default map is what a settlement is measured on, and catching up
// sleeping ground is not the day-by-day pass to the last bit, so nothing on
// the default map may ever sleep. The market's chunk is always occupied and
// the other is beside it; this says so for twenty years, house moves and
// market moves included.
// wooded is how many tiles of this chunk are under trees, walked rather
// than read off a count: what kind of ground a tile is is the patches'
// tally now, at a size finer than a chunk, and a test that wants one
// chunk's worth can afford to look.
func wooded(g *world.Grid, c *world.Chunk) int {
	n := 0
	for y := c.Y0; y < c.Y0+c.H; y++ {
		for x := c.X0; x < c.X0+c.W; x++ {
			if g.At(entity.Pos{X: x, Y: y}).Terrain == world.Forest {
				n++
			}
		}
	}
	return n
}

func TestTheDefaultMapNeverSleeps(t *testing.T) {
	w := world.New(6)
	for i := 0; i < 30; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	for tick := 0; tick < 7200; tick++ {
		Step(w)
		for i := range w.Grid.Chunks {
			if !w.Grid.Active[i] {
				t.Fatalf("chunk %d was asleep on day %d", i, w.Tick)
			}
		}
	}
}

// Ground that sleeps a season and is caught up in one go comes out where
// the day-by-day pass would have put it, to rounding, for the woods a map
// is made with.
func TestDormantLandCatchesUpToWithinRounding(t *testing.T) {
	w := world.NewWith(8, world.Config{Width: 384, Height: 256, Wrap: true})
	for i := 0; i < 20; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	Step(w) // wake once, so the active set exists
	g := w.Grid
	// The sleeping wooded chunk farthest from the market, so that nobody
	// wanders into it while it is being watched.
	sleeping, farthest := -1, -1
	for i := range g.Chunks {
		c := &g.Chunks[i]
		mid := entity.Pos{X: c.X0 + c.W/2, Y: c.Y0 + c.H/2}
		if d := g.Dist(mid, w.MarketPos); !g.Active[i] && wooded(g, c) > 20 && d > farthest {
			sleeping, farthest = i, d
		}
	}
	if sleeping < 0 {
		t.Fatal("no sleeping chunk with woods on it to test with")
	}
	c := &g.Chunks[sleeping]
	// The chunk's tiles, in the order the day walks them.
	var tiles []int
	for y := c.Y0; y < c.Y0+c.H; y++ {
		for x := c.X0; x < c.X0+c.W; x++ {
			tiles = append(tiles, g.Index(entity.Pos{X: x, Y: y}))
		}
	}
	// Take a copy of the map and pass over the chunk day by day with the
	// weather the world actually has, alongside the world sleeping.
	w.CatchUp(sleeping) // start level
	byDay := g.Clone()
	for day := 0; day < 90; day++ {
		Step(w)
		k := w.Rates()[sleeping] // the rate the world applies to this chunk
		for _, i := range tiles {
			byDay.Ripen(i, k)
			byDay.Replenish(i, k)
		}
		if g.Active[sleeping] {
			t.Fatalf("the chunk woke on day %d; pick another", w.Tick)
		}
	}
	w.CatchUp(sleeping)
	compared := 0
	for _, i := range tiles {
		a, b := &byDay.Tiles[i], &g.Tiles[i]
		if a.Terrain != b.Terrain {
			continue // seed fell here while it slept; the copy saw no seed
		}
		compared++
		if math.Abs(byDay.Age[i]-g.Age[i]) > 1e-9 || math.Abs(byDay.Wood[i]-g.Wood[i]) > 1e-6 || math.Abs(byDay.Wild[i]-g.Wild[i]) > 1e-6 || math.Abs(byDay.Fish[i]-g.Fish[i]) > 1e-9 {
			t.Fatalf("tile %d by day %+v %+v, at once %+v %+v", i, *a, byDay.Read(i), *b, g.Read(i))
		}
	}
	if compared < len(tiles)/2 {
		t.Fatalf("only %d of %d tiles were left to compare", compared, len(tiles))
	}
}
