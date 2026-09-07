package main

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"lreat/core/need"
	"lreat/core/observe"
	"lreat/core/world"
	"lreat/ui/ascii"
)

// The panel should show the population's condition next to its needs.
func TestPanelShowsHealth(t *testing.T) {
	w := world.NewSized(1, 40, 12)
	a := w.Spawn("a", need.Neutral())
	a.Health = 0.42

	sc := tcell.NewSimulationScreen("UTF-8")
	if err := sc.Init(); err != nil {
		t.Fatal(err)
	}
	defer sc.Fini()
	sc.SetSize(120, 40)

	s := observe.Take(w)
	v := &view{screen: sc, snap: &s}
	v.draw()

	cells, width, _ := sc.GetContents()
	var text strings.Builder
	for i, c := range cells {
		if i%width == 0 {
			text.WriteByte('\n')
		}
		text.WriteRune(c.Runes[0])
	}
	if !strings.Contains(text.String(), "health") {
		t.Fatalf("panel has no health row:\n%s", text.String())
	}
	if !strings.Contains(text.String(), "0.42") {
		t.Fatal("panel does not show the mean health value")
	}
}

// What everyone is doing is shown as history, not as this tick's answer:
// a column stands for several ticks, so a band lasts long enough to read.
func TestActivityGraphKeepsHistory(t *testing.T) {
	v := &view{}
	s := observe.Snapshot{Population: 4, Activity: []observe.Activity{{Action: "farm", Agents: 3}}}
	for i := 0; i < graphTicks-1; i++ {
		v.record(&s)
	}
	if len(v.hist) != 0 {
		t.Fatalf("a column appeared before %d ticks were in it", graphTicks)
	}
	v.record(&s)
	if len(v.hist) != 1 {
		t.Fatalf("want one column after %d ticks, have %d", graphTicks, len(v.hist))
	}
	if got := v.hist[0][ascii.GroupOf("farm")]; got != 0.75 {
		t.Fatalf("three of four farming is a share of 0.75, have %.2f", got)
	}
	for i := 0; i < graphMax*graphTicks*2; i++ {
		v.record(&s)
	}
	if len(v.hist) != graphMax {
		t.Fatalf("history is not bounded: %d columns", len(v.hist))
	}
}
