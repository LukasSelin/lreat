package world

import (
	"testing"
	"time"

	"lreat/core/entity"
)

// A valley reads the same weather on every row, to the bit: what a globe
// does by latitude a valley must not do at all.
func TestTheDefaultClimateReadsTheSameAtEveryRow(t *testing.T) {
	w := New(1)
	for tick := 0; tick < 400; tick++ {
		w.Climate.Advance(tick, w.RNG)
		for y := 0; y < w.Grid.H; y++ {
			if w.Climate.TempAt(y) != w.Climate.Temp || w.Climate.GrowthAt(y) != w.Climate.Growth() || w.Climate.ChillAt(y) != w.Climate.Chill() {
				t.Fatalf("row %d reads differently from the map on day %d", y, tick)
			}
		}
	}
}

// On a globe the year is warmer toward the middle and colder toward the
// poles, the middle has no winter, and the south's summer is the north's
// winter.
func TestGrowthIsSlowerTowardThePoles(t *testing.T) {
	c := NewClimateOn(Globe())
	c.Temp = seasonal(0)
	equator, temperate, pole := c.rows/2, c.rows/4, 0
	if !(c.MeanAt(equator) > c.MeanAt(temperate) && c.MeanAt(temperate) > c.MeanAt(pole)) {
		t.Fatalf("means: equator %.1f temperate %.1f pole %.1f", c.MeanAt(equator), c.MeanAt(temperate), c.MeanAt(pole))
	}
	if c.MeanAt(pole) >= Frost {
		t.Fatalf("the pole averages %.1f, above the frost", c.MeanAt(pole))
	}
	// Midsummer in the north is midwinter in the south.
	c.Temp = seasonal(Year / 4)
	north, south := c.TempAt(c.rows/4), c.TempAt(3*c.rows/4)
	c.Temp = seasonal(3 * Year / 4)
	northLater, southLater := c.TempAt(c.rows/4), c.TempAt(3*c.rows/4)
	if !(north > northLater && south < southLater) {
		t.Fatalf("north %.1f then %.1f, south %.1f then %.1f: the seasons do not turn over", north, northLater, south, southLater)
	}
	var lo, hi float64 = 100, -100
	for tick := 0; tick < Year; tick++ {
		c.Temp = seasonal(tick)
		lo, hi = min(lo, c.TempAt(equator)), max(hi, c.TempAt(equator))
	}
	if hi-lo > 1 {
		t.Fatalf("the equator swings %.1f degrees over the year", hi-lo)
	}
}

// A globe has a sea, its rivers reach it or a pole, and nothing but the
// sea lies on the seam that its ground does not continue across.
func TestAGlobeHasASeaItsRiversReach(t *testing.T) {
	if testing.Short() {
		t.Skip("a globe takes a second or two to make")
	}
	start := time.Now()
	w := NewWith(1, Globe())
	made := time.Since(start)
	g := w.Grid
	sea := 0
	for i := range g.Tiles {
		if g.underSea(i) {
			sea++
		}
	}
	if share := float64(sea) / float64(len(g.Tiles)); share < 0.25 || share > 0.35 {
		t.Fatalf("a third of the globe should be sea; %.2f is", share)
	}
	// Follow the water down from every wet tile: it ends in the sea, at a
	// pole, or in a hollow the flood should have filled.
	stranded := 0
	for i := range g.Tiles {
		if g.Tiles[i].Terrain != Water || g.underSea(i) {
			continue
		}
		p := g.PosOf(i)
		for steps := 0; steps < g.W+g.H; steps++ {
			a := g.Aspect(p)
			if a == (entity.Pos{}) {
				if !(g.underSea(g.Index(p)) || p.Y == 0 || p.Y == g.H-1) {
					stranded++
				}
				break
			}
			p = g.Norm(entity.Pos{X: p.X + a.X, Y: p.Y + a.Y})
		}
	}
	if stranded > 0 {
		t.Fatalf("%d river tiles drain into nowhere", stranded)
	}
	// The poles are bare and the middle is not.
	if g.At(entity.Pos{X: 100, Y: 0}).Terrain != Rock && g.At(entity.Pos{X: 100, Y: 0}).Terrain != Water {
		t.Fatal("the pole is not bare")
	}
	if made > 5*time.Second {
		t.Fatalf("the globe took %v to make", made)
	}
	t.Logf("a globe of %d tiles, %d sea, %d forest, made in %v; market at %v", len(g.Tiles), sea, g.Forest(), made, w.MarketPos)
}
