package ascii

import (
	"fmt"
	"testing"

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
		if r.Low == "" || r.High == "" {
			t.Errorf("%s does not say which way its shading runs", r.Name)
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
			if View(i) == Settlement {
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
