package main

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"lreat/core/observe"
	"lreat/core/world"
)

// land fills in what a snapshot says about the ground and the market, which
// is what the world page is made of.
func land(s *observe.Snapshot, forest, fields, houses int, price float64) {
	s.Forest, s.Forest0, s.Fields, s.Houses, s.FoodPrice = forest, 900, fields, houses, price
	s.Knowledge, s.Safety, s.Growth = float64(s.Tick), 1, 1
	s.Techs = []world.Tech{"pottery"}
}

// The world page says what became of the ground: the wood that came down,
// the fields that went in, what food cost, and when the settlement worked
// each thing out. A settlement is not only its people.
func TestWorldPageShowsWhatBecameOfTheGround(t *testing.T) {
	v := &view{screen: vitalScreen(t), world: true}
	for i := 1; i <= 1000; i++ {
		s := tick(i, 30, world.Vitals{Births: i / 100})
		// The wood comes down as the fields go in.
		land(&s, 900-i/2, i/50, i/100, 1+float64(i)/1000)
		v.snap = &s
		v.record(&s)
	}
	v.draw()

	text := screenText(v.screen.(tcell.SimulationScreen))
	for _, want := range []string{
		"the world", "forest", "fields", "houses", "food price", "knowledge",
		"of the 900 it started with", // what the wood was before anyone cut it
		"t1 pottery",                 // and when the settlement worked things out
		"what they have been doing with themselves",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("the world page never says %q:\n%s", want, text)
		}
	}
}

// Discoveries are kept with the tick they happened on, whichever page was up
// and however long ago it was: the event log is bounded and a long run drops
// its own beginning, which is where the first of them are.
func TestDiscoveriesAreKeptWithTheirTick(t *testing.T) {
	v := &view{}
	first := tick(100, 5, world.Vitals{})
	first.Techs = []world.Tech{"fishing"}
	v.record(&first)
	later := tick(700, 5, world.Vitals{})
	later.Techs = []world.Tech{"fishing", "pottery"}
	v.record(&later)
	again := tick(900, 5, world.Vitals{})
	again.Techs = []world.Tech{"fishing", "pottery"}
	v.record(&again)

	if len(v.techs) != 2 {
		t.Fatalf("kept %d discoveries, want 2", len(v.techs))
	}
	if v.techs[0].tick != 100 || v.techs[1].tick != 700 {
		t.Fatalf("discovered at %d and %d, want 100 and 700", v.techs[0].tick, v.techs[1].tick)
	}
}

// The written report carries the same history, because that is the only part
// of it that outlives the screen.
func TestReportCarriesTheWorld(t *testing.T) {
	v := &view{}
	for i := 1; i <= 600; i++ {
		s := tick(i, 10, world.Vitals{})
		land(&s, 900-i, i/60, i/100, 2)
		v.snap = &s
		v.record(&s)
	}
	got := v.Report()
	for _, want := range []string{"─ the world ─", "forest", "houses", "worked out", "pottery"} {
		if !strings.Contains(got, want) {
			t.Fatalf("the report never says %q:\n%s", want, got)
		}
	}
}
