package main

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"

	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/observe"
	"lreat/core/sim"
	"lreat/core/world"
)

// The panel says what the run is costing as well as what the settlement is
// doing. Without it a world that has slowed to a third of the speed it was
// asked for is indistinguishable from a quiet one.
func TestPanelShowsWhatTheRunCosts(t *testing.T) {
	w := world.NewSized(1, 40, 12)
	w.Spawn("a", need.Neutral())

	sc := tcell.NewSimulationScreen("UTF-8")
	if err := sc.Init(); err != nil {
		t.Fatal(err)
	}
	defer sc.Fini()
	sc.SetSize(120, 40)

	s := observe.Take(w)
	v := &view{screen: sc, snap: &s, speed: 16}
	v.load = func() sim.Load {
		return sim.Load{Rate: 15.5, Step: 2100 * time.Microsecond, Peak: 9 * time.Millisecond, Busy: 0.38}
	}
	v.draw()

	text := screenText(sc)
	for _, want := range []string{
		"what it costs",
		"480 tiles", // the map this cost was paid on
		"15.5",      // the rate actually managed
		"38%",       // the share of the wall clock spent stepping
		"2.1ms",     // what a tick took
		"9ms",       // and the worst one in the window
		"memory", "churn", "cpu", "gc",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("the machine block does not show %q:\n%s", want, text)
		}
	}
}

// A run that cannot keep up with the speed it was asked for says so in
// colour, because the number alone is only alarming to somebody who
// remembers what was asked for.
func TestSlowRunIsMarked(t *testing.T) {
	for _, c := range []struct {
		name  string
		rate  float64
		want  tcell.Color
		speed float64
	}{
		{"keeping up", 60, tcell.ColorDefault, 60},
		{"a tenth behind", 50, tcell.ColorYellow, 60},
		{"half behind", 20, tcell.ColorRed, 60},
	} {
		t.Run(c.name, func(t *testing.T) {
			w := world.NewSized(1, 40, 12)
			w.Spawn("a", need.Neutral())
			sc := tcell.NewSimulationScreen("UTF-8")
			if err := sc.Init(); err != nil {
				t.Fatal(err)
			}
			defer sc.Fini()
			sc.SetSize(120, 40)
			s := observe.Take(w)
			v := &view{screen: sc, snap: &s, speed: c.speed}
			v.load = func() sim.Load { return sim.Load{Rate: c.rate, Busy: 0.5} }
			v.draw()

			cells, width, _ := sc.GetContents()
			var got tcell.Color
			for i, cell := range cells {
				if cell.Runes[0] != 't' || i%width == 0 {
					continue
				}
				// The rate sits in the value column of the "t/s" row.
				if string(cells[i+1].Runes) != "/" || string(cells[i+2].Runes) != "s" {
					continue
				}
				fg, _, _ := cells[i+statLabel+statValue-1].Style.Decompose()
				got = fg
			}
			if got != c.want {
				t.Fatalf("a rate of %.0f against %.0f asked for is %v, want %v", c.rate, c.speed, got, c.want)
			}
		})
	}
}

// The block is anchored at the foot of the panel and takes its height off
// what is above it, so that adding it did not push the end of an agent's
// thinking off the bottom of an ordinary terminal. That thinking is why the
// card is there at all.
func TestCostBlockLeavesRoomForTheCard(t *testing.T) {
	w := world.NewSized(1, 40, 12)
	a := w.Spawn("Ada", need.Neutral())
	a.Plan = &entity.Plan{Action: "farm", Target: a.Pos, Remaining: 2, Total: 4}
	w.Watch(a.ID)
	w.Remember(world.Deliberation{Tick: 9, Agent: a.ID, Rule: "fit", Entropy: 1.2,
		Weighed: []world.Weighed{
			{Action: "farm", Weight: 0.41, Chance: 0.62, Chosen: true},
			{Action: "forage", Weight: 0.33, Chance: 0.21},
		}})

	sc := tcell.NewSimulationScreen("UTF-8")
	if err := sc.Init(); err != nil {
		t.Fatal(err)
	}
	defer sc.Fini()
	sc.SetSize(140, 48)

	s := observe.Take(w)
	v := &view{screen: sc, snap: &s, sel: a.ID, speed: 16,
		look: func(id entity.ID) *observe.Portrait { return observe.Look(w, id) }}
	v.draw()

	text := screenText(sc)
	if !strings.Contains(text, "forage") {
		t.Fatalf("the cost block crowded the card's thinking off the panel:\n%s", text)
	}
	if !strings.Contains(text, "what it costs") {
		t.Fatalf("the cost block is missing beside the card:\n%s", text)
	}
}

// A frame abandoned because the next event was already waiting is counted,
// not passed over: how often it happens is half of how old what is on the
// screen is.
func TestDroppedFramesAreCounted(t *testing.T) {
	var g gauge
	start := time.Now().Add(-2 * frameWindow)
	g.dropped()
	g.dropped()
	g.dropped()
	g.drew(start) // one frame drawn, three dropped, over a full window
	if g.missed != 0.75 {
		t.Fatalf("three frames dropped in four is 0.75, have %.2f", g.missed)
	}
	if g.frame < frameWindow {
		t.Fatalf("the frame took %s and was measured at %s", 2*frameWindow, g.frame)
	}
}

// The runtime half of the block is read off the process, so it says
// something about a process that is certainly running.
func TestMachineReadsTheRuntime(t *testing.T) {
	var g gauge
	g.machine()
	if g.memory == 0 {
		t.Fatal("the program is using no memory at all, which it is not")
	}
	// The second reading inside the sampling interval is the same one: the
	// figures hold still between samples rather than flickering per frame.
	was := g.taken
	g.machine()
	if g.taken != was {
		t.Fatal("the runtime was sampled twice inside one sampling interval")
	}
}

// Every figure in the block has to fit the column the panel gives it, in
// every magnitude it can plausibly reach: a globe's tick is milliseconds
// and a valley's is microseconds, and both are read in the same six cells.
func TestFiguresFitTheirColumn(t *testing.T) {
	for _, got := range []string{
		span(0), span(400 * time.Nanosecond), span(12 * time.Microsecond),
		span(2100 * time.Microsecond), span(940 * time.Millisecond), span(3 * time.Second),
		size(0), size(512), size(40 << 10), size(184 << 20), size(3 << 30),
		pct(0), pct(0.004), pct(0.38), pct(1),
		count(480), count(262144), count(4000000),
		rate(0), rate(12 << 20), rate(940),
	} {
		if n := len([]rune(got)); n > statValue {
			t.Fatalf("%q is %d cells wide and the column is %d", got, n, statValue)
		}
	}
}
