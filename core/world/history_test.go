package world

import (
	"testing"

	"lreat/core/entity"
)

// historyConfig is the default valley made out of its own history rather than
// drawn, which is the comparison every test here is about.
func historyConfig(epochs int) Config {
	cfg := DefaultConfig()
	cfg.Epochs = epochs
	return cfg
}

// shape is the few numbers that say what kind of map something is: how much
// of it can be ploughed, how much is water, and how the ground lies.
type shape struct {
	open, wet int
	slope50   float64
	slope90   float64
}

func shapeOf(g *Grid) shape {
	var s shape
	slopes := make([]float64, 0, len(g.Tiles))
	for i := range g.Tiles {
		p := entity.Pos{X: i % g.W, Y: i / g.W}
		slopes = append(slopes, g.Slope(p))
		if g.Tiles[i].Buildable() {
			s.open++
		}
		if g.Tiles[i].Wet() {
			s.wet++
		}
	}
	s.slope50, s.slope90 = quantile(slopes, 0.5), quantile(slopes, 0.9)
	return s
}

// The whole settlement model is tuned against maps the drawn generator makes:
// how much of one can be ploughed, how steep the steep part of it is, how far
// a person walks uphill to get anywhere. A history has no reason to land on
// any of that and every reason not to, so normalise makes it - and this is
// the test of that join rather than of the history.
//
// The bands are wide on purpose. What is being asked is not that a made world
// is a drawn one, which would make the whole thing pointless, but that it is
// the same kind of place: a valley somebody could live in, with mountains at
// the edges of it rather than through the middle of everything.
func TestAHistoryLeavesAMapTheSettlementCanUse(t *testing.T) {
	for _, seed := range []uint64{1, 2, 3} {
		drawn := shapeOf(NewWith(seed, DefaultConfig()).Grid)
		made := shapeOf(NewWith(seed, historyConfig(16)).Grid)

		if made.open < drawn.open*8/10 {
			t.Errorf("seed %d: %d tiles can be ploughed on a made world against %d on a drawn one",
				seed, made.open, drawn.open)
		}
		if made.wet > drawn.wet*3 {
			t.Errorf("seed %d: %d tiles of water against %d", seed, made.wet, drawn.wet)
		}
		// The steepest tenth is what decides whether a map is country or a
		// set of walls, and it is the reading that caught every wrong turn
		// this generator took: plate settling flattening the interiors, seams
		// raised as knife edges, and a normalise that squeezed the lowland
		// while leaving the mountains alone.
		// Two and a half, and not two, because that is where the seeds
		// actually fall and a band should say what was measured: over five
		// seeds the worst made world runs about twice its drawn twin and the
		// best runs under it. Tightening this is worth doing - a made valley
		// is still the steeper place - but it should be done by making
		// gentler ground, not by moving the line.
		if made.slope90 > 2.5*drawn.slope90 {
			t.Errorf("seed %d: the steepest tenth of a made world is %.3f against %.3f drawn",
				seed, made.slope90, drawn.slope90)
		}
	}
}

// Basalt and schist are the rocks that have to happen to a place: one comes
// up and the other is buried and squeezed, and no lattice that knows only
// where a tile is can lay either. A history that makes neither has not made a
// history.
func TestAHistoryMakesTheRocksThatHaveToHappen(t *testing.T) {
	for _, seed := range []uint64{1, 2, 3} {
		w := NewWith(seed, historyConfig(16))
		var seen [BedrockCount]int
		for i := range w.Grid.Tiles {
			seen[w.Grid.Tiles[i].Bedrock]++
		}
		for _, b := range []Bedrock{Basalt, Schist} {
			if seen[b] == 0 {
				t.Errorf("seed %d made no %s", seed, b)
			}
		}
		kinds := 0
		for _, b := range Bedrocks() {
			if seen[b] > 0 {
				kinds++
			}
		}
		if kinds < 4 {
			t.Errorf("seed %d came out with %d kinds of rock on it", seed, kinds)
		}
	}
}

// A rock knows when it was made, and not all of it was made at once. If every
// tile dated from the same epoch there would be no history in the record,
// only in the making of it.
func TestRockIsDatedToWhenItWasMade(t *testing.T) {
	w := NewWith(1, historyConfig(16))
	seen := map[uint8]int{}
	for i := range w.Grid.Tiles {
		seen[w.Grid.Tiles[i].Formed]++
	}
	if len(seen) < 3 {
		t.Errorf("the whole map dates from %d epochs: %v", len(seen), seen)
	}
}

// Every tile rides a plate, and a history that ran at all has more than one.
func TestEveryTileRidesAPlate(t *testing.T) {
	w := NewWith(1, historyConfig(16))
	seen := map[uint8]bool{}
	for i := range w.Grid.Tiles {
		seen[w.Grid.Tiles[i].Plate] = true
	}
	if len(seen) < 3 {
		t.Errorf("the map broke into %d plates", len(seen))
	}
}

// A history is drawn from the world's own luck and nothing else, so the same
// seed has to give the same world - the whole model is deterministic and a
// generator that was not would take that away from it.
func TestTheSameSeedRunsTheSameHistory(t *testing.T) {
	a := NewWith(7, historyConfig(12)).Grid
	b := NewWith(7, historyConfig(12)).Grid
	for i := range a.Tiles {
		if a.Tiles[i] != b.Tiles[i] {
			t.Fatalf("tile %d came out %+v one time and %+v the next", i, a.Tiles[i], b.Tiles[i])
		}
	}
}

// A drawn world is what it always was: no plates, no dated rock, and none of
// the two rocks a history makes. Nothing about this change may reach a map
// that did not ask for it.
func TestADrawnWorldIsUntouched(t *testing.T) {
	w := New(1)
	for i := range w.Grid.Tiles {
		t2 := &w.Grid.Tiles[i]
		if t2.Plate != 0 || t2.Formed != 0 {
			t.Fatalf("tile %d of a drawn world rides plate %d and dates from %d", i, t2.Plate, t2.Formed)
		}
		if t2.Bedrock == Basalt || t2.Bedrock == Schist {
			t.Fatalf("tile %d of a drawn world is %s", i, t2.Bedrock)
		}
	}
}
