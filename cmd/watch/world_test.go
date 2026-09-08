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

// puts counts in cells, not in bytes. Ranging over a string gives byte
// offsets, which is the same thing only while the text is ASCII: with a
// rule or an arrow in the line every glyph after it lands two cells too far
// right and the tail runs off the width the line was trimmed to. Every
// header on every page is drawn through puts, so this is all of them.
func TestPutsWritesOneCellPerRune(t *testing.T) {
	sc := vitalScreen(t)
	sc.Clear()
	text := "─ the world ─ 40 by 12 →"
	puts(sc, 0, 0, tcell.StyleDefault, text)
	sc.Show()

	cells, w, _ := sc.GetContents()
	for i, r := range []rune(text) {
		if got := cells[i].Runes[0]; got != r {
			t.Fatalf("cell %d holds %q, want %q — the line is spread over %d cells",
				i, string(got), string(r), len(text))
		}
	}
	if got := cells[len([]rune(text))].Runes[0]; got != ' ' {
		t.Fatalf("the line runs past its last rune, into %q", string(got))
	}
	_ = w
}

// row is where a piece of text stands on the screen, counted from the top.
func row(t *testing.T, sc tcell.SimulationScreen, want string) int {
	t.Helper()
	for i, line := range strings.Split(screenText(sc), "\n") {
		if strings.Contains(line, want) {
			return i
		}
	}
	t.Fatalf("nothing on the screen says %q:\n%s", want, screenText(sc))
	return 0
}

// The world page spends the window it is given. The weave of what people
// have been doing is the one thing on it with no natural height, so it
// takes whatever the measures above it have not: a taller terminal buys a
// finer weave rather than a taller blank.
func TestWorldPageFillsTheWindow(t *testing.T) {
	sc := vitalScreen(t)
	v := &view{screen: sc, world: true}
	for i := 1; i <= 1000; i++ {
		s := tick(i, 30, world.Vitals{Births: i / 100})
		land(&s, 900-i/2, i/50, i/100, 1+float64(i)/1000)
		s.Activity = []observe.Activity{{Action: "eat", Agents: 20}}
		v.snap = &s
		v.record(&s)
	}
	for _, h := range []int{30, 46, 60} {
		sc.SetSize(100, h)
		v.draw()
		// The legend under the weave is the last thing before the run's
		// discoveries and the keys, so where it sits says how much of the
		// window the page above it took.
		if at := row(t, sc, "other"); at < h-6 {
			t.Errorf("on a %d-row terminal the weave ends at row %d, leaving %d rows blank under it",
				h, at, h-at)
		}
	}
}
