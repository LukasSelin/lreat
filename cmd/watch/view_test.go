package main

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"lreat/core/entity"
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

// The settlement's figures are a grid: whatever the numbers are, each
// label sits in its own column, so the panel can be read down as well as
// across.
func TestPanelFiguresLineUp(t *testing.T) {
	w := world.NewSized(1, 40, 12)
	for i := 0; i < 8; i++ {
		w.Spawn("a", need.Neutral())
	}

	sc := tcell.NewSimulationScreen("UTF-8")
	if err := sc.Init(); err != nil {
		t.Fatal(err)
	}
	defer sc.Fini()
	sc.SetSize(120, 40)

	s := observe.Take(w)
	v := &view{screen: sc, snap: &s}
	v.draw()

	px := s.Map.W + 2
	rows := strings.Split(screenText(sc), "\n")
	for _, pair := range [][2]string{
		{"mean age", "elders"},
		{"friends", "feuds"},
		{"houses", "fields"},
		{"forest", "roads"},
		{"food price", "safety"},
		{"knowledge", "gini"},
		{"reach", "spread"},
	} {
		row := ""
		for _, r := range rows {
			if runes := []rune(r); len(runes) > px && strings.HasPrefix(string(runes[px:]), pair[0]) {
				row = string(runes[px:])
				break
			}
		}
		if row == "" {
			t.Fatalf("no panel row starts with %q", pair[0])
		}
		if got := strings.Index(row, pair[1]); got != statCol {
			t.Errorf("%q sits at column %d on the %q row, want %d", pair[1], got, pair[0], statCol)
		}
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

// Picking a figure out of the crowd opens it up beside the map: who it is,
// what it is good at, what it is doing, and what it weighed before it set
// out. Without the last of those the view shows movement and no reason.
func TestCardShowsTheFollowedAgent(t *testing.T) {
	w := world.NewSized(1, 40, 12)
	a := w.Spawn("Ada", need.Neutral())
	a.AddSkill(entity.Farming, 0.6)
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
	v := &view{screen: sc, snap: &s, sel: a.ID, look: func(id entity.ID) *observe.Portrait { return observe.Look(w, id) }}
	v.draw()

	text := screenText(sc)
	for _, want := range []string{"Ada", "farming", "farm", "forage", "62%", "spring 10, year 1"} {
		if !strings.Contains(text, want) {
			t.Fatalf("the card does not show %q:\n%s", want, text)
		}
	}
}

// Following somebody who has died says so rather than quietly following
// whoever is standing where they were.
func TestCardSaysWhenTheAgentIsGone(t *testing.T) {
	w := world.NewSized(1, 40, 12)
	w.Spawn("Ada", need.Neutral())

	sc := tcell.NewSimulationScreen("UTF-8")
	if err := sc.Init(); err != nil {
		t.Fatal(err)
	}
	defer sc.Fini()
	sc.SetSize(140, 48)

	s := observe.Take(w)
	v := &view{screen: sc, snap: &s, sel: 99, look: func(entity.ID) *observe.Portrait { return nil }}
	v.draw()
	if !strings.Contains(screenText(sc), "gone") {
		t.Fatalf("the card does not say the agent is gone:\n%s", screenText(sc))
	}
}

// Tab walks the population and tells the simulation who to keep the
// thinking of; esc lets go again.
func TestPickWalksThePopulation(t *testing.T) {
	w := world.NewSized(1, 40, 12)
	w.Spawn("Ada", need.Neutral())
	w.Spawn("Bo", need.Neutral())
	s := observe.Take(w)

	var followed []entity.ID
	v := &view{snap: &s, follow: func(id entity.ID) { followed = append(followed, id) }}
	v.pick(1)
	first := v.sel
	v.pick(1)
	second := v.sel
	if first == 0 || second == 0 || first == second {
		t.Fatalf("tab does not move through the population: %d then %d", first, second)
	}
	v.pick(1)
	if v.sel != first {
		t.Fatal("tab does not come back round to the first agent")
	}
	v.choose(0)
	if v.sel != 0 || v.pic != nil {
		t.Fatal("letting go left somebody selected")
	}
	if len(followed) != 4 {
		t.Fatalf("the simulation was told to watch %d times, want 4", len(followed))
	}
}

func screenText(sc tcell.SimulationScreen) string {
	cells, width, _ := sc.GetContents()
	var text strings.Builder
	for i, c := range cells {
		if i%width == 0 {
			text.WriteByte('\n')
		}
		text.WriteRune(c.Runes[0])
	}
	return text.String()
}
