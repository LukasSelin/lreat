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
		if d := g.Dist(mid, w.MarketPos); !g.Active[i] && c.Kinds[world.Forest] > 20 && d > farthest {
			sleeping, farthest = i, d
		}
	}
	if sleeping < 0 {
		t.Fatal("no sleeping chunk with woods on it to test with")
	}
	c := &g.Chunks[sleeping]
	// Take a copy of the chunk's tiles and pass over them day by day with
	// the weather the world actually has, alongside the world sleeping.
	copyOf := func() []world.Tile {
		var out []world.Tile
		for y := c.Y0; y < c.Y0+c.H; y++ {
			for x := c.X0; x < c.X0+c.W; x++ {
				out = append(out, *g.At(entity.Pos{X: x, Y: y}))
			}
		}
		return out
	}
	w.CatchUp(sleeping) // start level
	byDay := copyOf()
	for day := 0; day < 90; day++ {
		Step(w)
		k := w.Rates()[sleeping] // the rate the world applies to this chunk
		for i := range byDay {
			byDay[i].Ripen(k)
			byDay[i].Replenish(k)
		}
		if g.Active[sleeping] {
			t.Fatalf("the chunk woke on day %d; pick another", w.Tick)
		}
	}
	w.CatchUp(sleeping)
	atOnce := copyOf()
	compared := 0
	for i := range byDay {
		a, b := byDay[i], atOnce[i]
		if a.Terrain != b.Terrain {
			continue // seed fell here while it slept; the copy saw no seed
		}
		compared++
		if math.Abs(a.Age-b.Age) > 1e-9 || math.Abs(a.Wood-b.Wood) > 1e-6 || math.Abs(a.Wild-b.Wild) > 1e-6 || math.Abs(a.Fish-b.Fish) > 1e-9 {
			t.Fatalf("tile %d by day %+v, at once %+v", i, a, b)
		}
	}
	if compared < len(byDay)/2 {
		t.Fatalf("only %d of %d tiles were left to compare", compared, len(byDay))
	}
}
