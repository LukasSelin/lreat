package system

import (
	"testing"

	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
	"lreat/core/world"
)

// A death is recorded under what it was of. Starving to death and a body
// giving out with age are the same line on a population curve and nothing
// alike as reasons a settlement ended, so the two are counted apart.
func TestDeathsAreCountedByCause(t *testing.T) {
	w := world.NewSized(1, 40, 12)
	starved := w.Spawn("Ada", need.Neutral())
	starved.Starving = Starvation + 1
	old := w.Spawn("Bo", need.Neutral())
	// Far enough past its prime that frailty really is certain. At a hundred
	// thousand, which this was, it is not: the risk climbs as the square of
	// how far through its decline a body is, and a hundred thousand ticks
	// works out at about one chance in seven on the tick. The test passed on
	// the draw, and stopped passing the moment anything upstream of it took a
	// different number of draws out of the world's luck.
	old.Born = w.Tick - 1_000_000
	old.Health = 0

	Population(w)

	if w.Vitals.Starved != 1 || w.Vitals.Failed != 1 {
		t.Fatalf("starved %d, gave out %d, want one of each", w.Vitals.Starved, w.Vitals.Failed)
	}
	if w.Vitals.Died != 2 || w.Deaths != 2 {
		t.Fatalf("this tick %d, all told %d, want two and two", w.Vitals.Died, w.Deaths)
	}
	if len(w.Agents) != 0 {
		t.Fatalf("%d agents outlived their own deaths", len(w.Agents))
	}
	// The chronicle is what a viewer reads afterwards, and it must hold the
	// deaths whether or not anybody was watching the tick they happened on.
	var died int
	for _, n := range w.Chronicle {
		if n.Kind == event.Died {
			died++
		}
	}
	if died != 2 {
		t.Fatalf("the chronicle kept %d of the two deaths", died)
	}
}

// Everyone alive is counted once under the first thing standing between them
// and a child. Without this a settlement that has stopped bearing looks
// exactly like one that is merely unlucky.
func TestFertilityGatesAccountForEveryone(t *testing.T) {
	w := world.NewSized(1, 40, 12)
	ready := func(a *entity.Agent) *entity.Agent {
		a.Born = w.Tick - entity.Maturity - 1
		a.Needs = need.Levels{0.9, 0.9, 0.9, 0.5, 0.5}
		return a
	}
	child := w.Spawn("child", need.Neutral())
	child.Born = w.Tick
	elder := ready(w.Spawn("elder", need.Neutral()))
	elder.Born = w.Tick - entity.Prime
	hungry := ready(w.Spawn("hungry", need.Neutral()))
	hungry.Needs[need.Physiological] = 0.1
	alone := ready(w.Spawn("alone", need.Neutral()))
	alone.Needs[need.Belonging] = 0.1
	ready(w.Spawn("ready", need.Neutral()))

	Population(w)

	for g, want := range map[world.Gate]int{
		world.Young: 1, world.Spent: 1, world.Hungry: 1, world.Alone: 1, world.Ready: 1,
		world.Unsafe: 0, world.Crowded: 0,
	} {
		if got := w.Vitals.Gates[g]; got != want {
			t.Errorf("%s: %d, want %d", g, got, want)
		}
	}
	var total int
	for _, n := range w.Vitals.Gates {
		total += n
	}
	if total != 5 {
		t.Fatalf("the gates account for %d of the 5 alive", total)
	}
}

// The gates are this tick's answer and not the run's: whoever was blocked
// last tick and is not blocked now must not still be counted.
func TestFertilityGatesAreThisTickOnly(t *testing.T) {
	w := world.NewSized(1, 40, 12)
	a := w.Spawn("Ada", need.Neutral())
	a.Born = w.Tick - entity.Maturity - 1
	a.Needs[need.Physiological] = 0.1
	Population(w)
	if w.Vitals.Gates[world.Hungry] != 1 {
		t.Fatalf("a hungry agent was not counted as hungry")
	}
	a.Needs = need.Levels{0.9, 0.9, 0.9, 0.5, 0.5}
	Population(w)
	if w.Vitals.Gates[world.Hungry] != 0 || w.Vitals.Gates[world.Ready] != 1 {
		t.Fatalf("hungry %d, ready %d after the larder filled, want 0 and 1",
			w.Vitals.Gates[world.Hungry], w.Vitals.Gates[world.Ready])
	}
}
