package main

import (
	"slices"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"lreat/core/observe"
	"lreat/core/world"
)

// run2000 is a settlement with two thousand days behind it: the wood coming
// down, the fields going in, and most of the people at their food while a
// couple study. It is what every page here is drawn from.
func run2000(t *testing.T) *view {
	t.Helper()
	v := &view{screen: vitalScreen(t)}
	for i := 1; i <= 2000; i++ {
		s := tick(i, 20+i/100, world.Vitals{Births: i / 50, Starved: i / 200})
		land(&s, 900-i/3, i/50, i/100, 1+float64(i)/900)
		s.Activity = []observe.Activity{
			{Action: "eat", Agents: 12}, {Action: "build", Agents: 5}, {Action: "study", Agents: 1},
		}
		v.snap = &s
		v.record(&s)
	}
	return v
}

// Stepping goes through everything the page has and off the end of it, back
// to the page as a whole. Reaching the end is how one stops reading a single
// measure without having to remember which key drops it.
func TestStepGoesThroughAndOffTheEnd(t *testing.T) {
	v := run2000(t)
	n := len(v.graphs())
	if n < 2 {
		t.Fatalf("the map page offers %d graphs to step through", n)
	}
	for i := 1; i <= n; i++ {
		v.step(1)
		if v.focus != i {
			t.Fatalf("stepping %d times opened graph %d, want %d", i, v.focus, i)
		}
		if _, ok := v.focused(); !ok {
			t.Fatalf("graph %d says nothing is opened", i)
		}
	}
	v.step(1)
	if v.focus != 0 {
		t.Fatalf("stepping past the last left graph %d open, want the page back", v.focus)
	}
	// And backwards from the page takes the last, so the far end of the
	// list is one press away rather than a dozen.
	v.step(-1)
	if v.focus != n {
		t.Fatalf("stepping back from the page opened graph %d, want the last, %d", v.focus, n)
	}
}

// Every page has its own graphs, and changing page closes whatever was open
// on the one being left: a focus kept across the change would land on
// whatever happened to sit at that number on the new page.
func TestEachPageHasItsOwnGraphs(t *testing.T) {
	v := run2000(t)
	pages := []struct {
		name  string
		keys  string
		wants []string
	}{
		{"map", "", []string{"everyone", "food", "study"}},
		{"world", "w", []string{"forest", "food price", "gini"}},
		{"vitals", "d", []string{"population", "born", "died"}},
	}
	for _, p := range pages {
		if p.keys != "" {
			v.focus = 3 // something open on the page being left
			v.handleKey(nil, rune_(rune(p.keys[0])))
		}
		if v.focus != 0 {
			t.Fatalf("changing to the %s page left graph %d open", p.name, v.focus)
		}
		var names []string
		for _, g := range v.graphs() {
			names = append(names, g.name)
		}
		for _, want := range p.wants {
			if !slices.Contains(names, want) {
				t.Fatalf("the %s page cannot step to %q; it offers %v", p.name, want, names)
			}
		}
	}
}

// Esc backs off the graph before it backs off the page. Whoever is reading
// one measure closely means to go back to the page it is on, not out of the
// page altogether.
func TestEscapeDropsTheGraphBeforeThePage(t *testing.T) {
	v := run2000(t)
	v.world, v.focus = true, 2
	v.handleKey(nil, key(tcell.KeyEscape))
	if v.focus != 0 {
		t.Fatal("esc did not close the opened graph")
	}
	if !v.world {
		t.Fatal("esc left the world page while a graph was still open on it")
	}
	v.handleKey(nil, key(tcell.KeyEscape))
	if v.world {
		t.Fatal("esc did not leave the world page once nothing was open")
	}
}

// The arrows are what step, on whichever page is up.
func TestArrowsStep(t *testing.T) {
	v := run2000(t)
	v.handleKey(nil, key(tcell.KeyDown))
	if v.focus != 1 {
		t.Fatalf("down opened graph %d, want the first", v.focus)
	}
	v.handleKey(nil, key(tcell.KeyUp))
	if v.focus != 0 {
		t.Fatalf("up from the first left graph %d open, want the page back", v.focus)
	}
}

// A measure opened out is drawn over the page's whole space and says what
// its height stands for. Squeezed into one row it says whether the wood went
// down; opened out it says how fast, and whether it has stopped.
func TestOpenedMeasureIsDrawnOut(t *testing.T) {
	v := run2000(t)
	v.world = true
	v.focus = focusOfMeasure(t, "forest")
	v.draw()

	text := screenText(v.screen.(tcell.SimulationScreen))
	if !strings.Contains(text, "forest   now") || !strings.Contains(text, "full height is") {
		t.Fatalf("the opened measure does not say where it stands or what its height means:\n%s", text)
	}
	// The row it was opened from is still in its place, marked.
	if !strings.Contains(text, "▸") {
		t.Fatal("nothing on the page says which row is opened out")
	}
	// The weave's legend belongs to the weave, which is not up.
	if strings.Contains(text, "build 0%") {
		t.Fatal("the weave's legend is still under a chart that is not the weave")
	}
}

// The point of opening a band out: a kind of work that holds a twentieth of
// the population is not drawn at all in a six-row weave shared with the
// rest, and is a chart of its own when it is stepped onto.
func TestThinBandIsVisibleOpenedOut(t *testing.T) {
	v := run2000(t)
	v.focus = focusOfBand(t, "study")
	v.draw()

	sc := v.screen.(tcell.SimulationScreen)
	if got := v.snap; got == nil {
		t.Fatal("no snapshot to draw the map from")
	}
	// The band's rows under the map should carry something now.
	cells, w, _ := sc.GetContents()
	filled := 0
	for y := v.snap.Map.H + 1; y < v.snap.Map.H+1+graphHeight; y++ {
		for x := 0; x < v.snap.Map.W; x++ {
			if r := cells[y*w+x].Runes[0]; r != ' ' && r != 0 {
				filled++
			}
		}
	}
	if filled == 0 {
		t.Fatalf("the study band opened out drew nothing:\n%s", screenText(sc))
	}
	if !strings.Contains(screenText(sc), "study   now") {
		t.Fatal("the band opened out does not name itself")
	}
}

// focusOfMeasure and focusOfBand are what focus has to be for a named graph
// to be the one opened out. Focus counts from one, and both lists carry more
// than the thing being looked for, so neither is its own index.
func focusOfMeasure(t *testing.T, name string) int {
	t.Helper()
	return focusOf(t, worldGraphs(), name)
}

func focusOfBand(t *testing.T, name string) int {
	t.Helper()
	return focusOf(t, mapGraphs(), name)
}

func focusOf(t *testing.T, gs []graph, name string) int {
	t.Helper()
	for i, g := range gs {
		if g.name == name {
			return i + 1
		}
	}
	t.Fatalf("nothing on the page is called %q", name)
	return 0
}
