package ascii

import (
	"fmt"
	"testing"

	"lreat/core/entity"

	"lreat/core/observe"
	"lreat/core/system"
	"lreat/core/world"
)

// livedIn is a settlement that has been going long enough to have marked the
// ground it stands on. A reading of wear on a map at tick zero is flat by
// definition - nobody has walked anywhere yet - so the readings are checked
// on a world that has been lived in rather than on one just made.
func livedIn(t *testing.T, seed uint64, days int) *observe.MapView {
	t.Helper()
	w := world.New(seed)
	for i := 0; i < 12; i++ {
		w.Spawn(fmt.Sprintf("settler%d", i), w.RandomPersonality())
	}
	for i := 0; i < days; i++ {
		system.Step(w)
	}
	return observe.Take(w).Map
}

// Every reading has to be drawable and has to be named. A view with no draw
// falls back to the settlement, which is a switch that silently does nothing.
func TestEveryViewIsDrawnAndNamed(t *testing.T) {
	for i, r := range Views {
		if r.Name == "" {
			t.Errorf("view %d has no name", i)
		}
		if View(i) == Settlement {
			continue
		}
		if r.draw == nil {
			t.Errorf("%s has no drawing, so it would fall back to the settlement", r.Name)
		}
		switch r.Legend {
		case Scale:
			if r.Low == "" || r.High == "" {
				t.Errorf("%s is a scale and does not say which way it runs", r.Name)
			}
		case Key:
			if r.Says == "" {
				t.Errorf("%s is a key and does not say what its colours mean", r.Name)
			}
		}
		if r.Ramp == ([Bands]Color{}) {
			t.Errorf("%s has no ramp, so its legend cannot show its own scale", r.Name)
		}
	}
}

// A reading has to actually vary over a map. One that comes out all one shade
// is telling the reader nothing, and it is the failure to expect: a scale
// that saturates, or a divisor that puts every real value in the same step.
// Three steps of six is a low bar deliberately - what is being caught is a
// reading that is flat, not one that is merely uneven.
func TestEveryReadingVaries(t *testing.T) {
	for _, seed := range []uint64{1, 2, 3} {
		m := livedIn(t, seed, 400)
		for i, r := range Views {
			// A key is not a scale: it draws in two glyphs by design - held
			// and not held - and asking it to use the shade ramp would be
			// asking it to be a quantity, which is the one thing it is not.
			//
			// Fish is left out for a different and less comfortable reason.
			// It is a scale and it draws correctly across its range - see
			// TestTheFishReadingSpansItsRange - but no settlement ever moves
			// it: a fresh map runs from 0.70 to 1.00 and by four hundred days
			// every water tile has regrown to exactly 1.00 and stays there
			// through six thousand. The reading is flat because the water
			// really is untouched, and asserting otherwise here would be
			// asserting something about fishing that is not true.
			if View(i) == Settlement || r.Legend == Key || View(i) == Fish {
				continue
			}
			seen := map[rune]bool{}
			for _, row := range RenderView(m, View(i)) {
				for _, c := range row {
					seen[c.Ch] = true
				}
			}
			steps := 0
			for _, s := range shades {
				if seen[s] {
					steps++
				}
			}
			if steps < 3 {
				t.Errorf("seed %d: %s uses only %d of %d shades, so it is nearly flat",
					seed, r.Name, steps, Bands)
			}
		}
	}
}

// The readings are the land, not the settlement, but the market stays on them
// so a reader can find where the people are on a map of the soil.
func TestAReadingKeepsTheMarket(t *testing.T) {
	m := livedIn(t, 1, 20)
	rows := RenderView(m, Soil)
	if got := rows[m.Market.Y][m.Market.X]; got.Ch != 'M' {
		t.Errorf("the market draws as %q on the soil reading, want 'M'", got.Ch)
	}
}

// Cycling has to come back round to where it started, so that somebody who
// has pressed the key too many times is never stuck off the settlement.
func TestCyclingReturnsToTheSettlement(t *testing.T) {
	v := Settlement
	for i := 0; i < len(Views); i++ {
		v = (v + 1) % View(len(Views))
	}
	if v != Settlement {
		t.Errorf("a full cycle of %d views landed on %v, want the settlement", len(Views), v)
	}
}

// A key has to actually tell things apart. Ownership drawn in one colour is a
// map of whether anybody owns anything, which is a far poorer question than
// the one it is meant to answer.
func TestAKeyTellsHoldersApart(t *testing.T) {
	m := livedIn(t, 1, 900)
	held, colours := 0, map[Color]bool{}
	for _, row := range RenderView(m, Holdings) {
		for _, c := range row {
			if c.Color == Bare || c.Ch == 'M' {
				continue
			}
			held++
			colours[c.Color] = true
		}
	}
	if held == 0 {
		t.Fatal("nothing on the map is held after nine hundred days")
	}
	if len(colours) < 2 {
		t.Errorf("%d held tiles are drawn in %d colour(s); a key that cannot "+
			"tell two holdings apart is a map of whether anyone owns anything", held, len(colours))
	}
}

// Ground nobody has claimed has to recede, or the reading is a picture of the
// terrain with some colour thrown over it.
func TestUnclaimedGroundIsQuiet(t *testing.T) {
	m := livedIn(t, 1, 200)
	for y, row := range RenderView(m, Holdings) {
		for x, c := range row {
			if c.Ch == 'M' {
				continue
			}
			owner := m.At(entityPos(x, y)).Owner
			if owner == 0 && c.Color != Bare {
				t.Fatalf("unclaimed ground at %d,%d is drawn in %v, not the quiet colour", x, y, c.Color)
			}
			if owner != 0 && c.Color == Bare {
				t.Fatalf("held ground at %d,%d is drawn as unclaimed", x, y)
			}
		}
	}
}

func entityPos(x, y int) entity.Pos { return entity.Pos{X: x, Y: y} }

// The fish reading has to shade across the whole range fish can take, even
// though no settlement has yet been seen to take them below full. What is
// checked here is the drawing and not the world: if fishing ever does start
// to bite, this is what will show it.
func TestTheFishReadingSpansItsRange(t *testing.T) {
	m := livedIn(t, 1, 0)
	var water []int
	for i := range m.Tiles {
		if m.Tiles[i].Terrain == world.Water {
			water = append(water, i)
		}
	}
	if len(water) < Bands {
		t.Fatalf("only %d water tiles to shade", len(water))
	}
	for n, i := range water {
		m.Tiles[i].Fish = float64(n) / float64(len(water)-1)
	}
	seen := map[rune]bool{}
	for _, row := range RenderView(m, Fish) {
		for _, c := range row {
			seen[c.Ch] = true
		}
	}
	for step, glyph := range shades {
		if !seen[glyph] {
			t.Errorf("fish never draws step %d (%q) over its whole range", step, glyph)
		}
	}
}
