package main

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"lreat/core/need"
	"lreat/core/observe"
	"lreat/core/world"
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
