package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"lreat/core/event"
	"lreat/core/observe"
	"lreat/core/world"
	"lreat/report"
)

// tick makes a snapshot of a settlement at one moment, enough of one for the
// vitals page: a headcount, what it has cost so far, and a map to sit on.
func tick(t int, pop int, v world.Vitals) observe.Snapshot {
	return observe.Snapshot{
		Tick: t, Population: pop, Deaths: v.Starved + v.Failed, Vitals: v,
		Map: &observe.MapView{W: 40, H: 12, Tiles: make([]world.Tile, 40*12), Layers: world.NewLayers(40 * 12)},
	}
}

func vitalScreen(t *testing.T) tcell.SimulationScreen {
	t.Helper()
	sc := tcell.NewSimulationScreen("UTF-8")
	if err := sc.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(sc.Fini)
	sc.SetSize(160, 50)
	return sc
}

// The whole point of the page: a settlement that has ended says when it
// ended, what its people died of, and what stood between the last of them
// and a child. A run that stops with none of that is a mystery.
func TestVitalsPageSaysWhyItDiedOut(t *testing.T) {
	v := &view{screen: vitalScreen(t), vitals: true}
	for i := 1; i <= 200; i++ {
		s := tick(i, 8, world.Vitals{Births: 3, Starved: 1})
		v.snap = &s
		v.record(&s)
	}
	end := tick(400, 0, world.Vitals{
		Births: 3, Starved: 9, Failed: 2,
		Gates: [world.GateCount]int{world.Spent: 4, world.Hungry: 2},
	})
	end.Chronicle = []world.Note{{Tick: 399, Kind: event.Died, Text: "Ada starved"}}
	v.snap = &end
	v.record(&end)
	v.draw()

	text := screenText(v.screen.(tcell.SimulationScreen))
	for _, want := range []string{
		"DIED OUT",             // that it ended, and when
		"starved",              // what of
		"past prime", "hungry", // and what stopped anyone replacing them
		"Ada starved", // in the settlement's own words
		"peak 8",      // against what it once was
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("the vitals page never says %q:\n%s", want, text)
		}
	}
}

// The history is kept whether the page is open or not. A settlement dies out
// once, and nobody is on the right page when it does.
func TestVitalsHistoryIsKeptWhileTheMapIsUp(t *testing.T) {
	v := &view{}
	for i := 1; i <= graphTicks*4; i++ {
		s := tick(i, 5, world.Vitals{})
		v.record(&s)
	}
	if len(v.traces) != 4 {
		t.Fatalf("%d columns of history after %d ticks on the map page, want 4", len(v.traces), graphTicks*4)
	}
	if v.peak != 5 {
		t.Fatalf("peak population %d, want 5", v.peak)
	}
	for i := 0; i < traceMax*graphTicks*3; i++ {
		s := tick(graphTicks*4+i, 5, world.Vitals{})
		v.record(&s)
	}
	if len(v.traces) > traceMax {
		t.Fatalf("history is not bounded: %d columns", len(v.traces))
	}
	// Bounded, but still the whole run: a long run is thinned rather than
	// cut off at the front, because the front is where a settlement's end
	// was decided.
	if v.traces[0].tick > graphTicks*8 {
		t.Fatalf("history starts at tick %d, having dropped the beginning of the run", v.traces[0].tick)
	}
	if last := v.traces[len(v.traces)-1].tick; last < traceMax*graphTicks*3-graphTicks*2 {
		t.Fatalf("history ends at tick %d, short of the present", last)
	}
}

// The rates are read over a window, so that a settlement losing people now
// does not hide behind the generations it raised earlier in the run.
func TestVitalsSpanReadsRecentTurnover(t *testing.T) {
	v := &view{}
	// A thousand ticks of growth, then five hundred of nothing but deaths.
	for i := 1; i <= 1000; i++ {
		s := tick(i, 20, world.Vitals{Births: i / 10})
		v.record(&s)
	}
	for i := 1001; i <= 1500; i++ {
		s := tick(i, 20, world.Vitals{Births: 100, Starved: (i - 1000) / 10})
		v.record(&s)
	}
	born, died, ok := v.span(rateSpan)
	if !ok {
		t.Fatal("a 1500-tick run has no rate over its last 500 ticks")
	}
	if born != 0 {
		t.Errorf("%d born in the last %d ticks, want none", born, rateSpan)
	}
	if died < 45 || died > 55 {
		t.Errorf("%d died in the last %d ticks, want about 50", died, rateSpan)
	}
}

// A run too young to have a window says so rather than quoting a rate over
// the three ticks it has.
func TestVitalsSpanAdmitsAShortRun(t *testing.T) {
	v := &view{}
	for i := 1; i <= 50; i++ {
		s := tick(i, 4, world.Vitals{Births: 1})
		v.record(&s)
	}
	if _, _, ok := v.span(rateSpan); ok {
		t.Fatal("a 50-tick run claims a rate over 500 ticks")
	}
}

// The map page says a settlement has ended and where to read why, rather
// than simply going still and looking like a hung run.
func TestMapPageSaysTheSettlementEnded(t *testing.T) {
	v := &view{screen: vitalScreen(t)}
	alive := tick(10, 3, world.Vitals{})
	v.snap = &alive
	for i := 0; i < graphTicks; i++ {
		v.record(&alive)
	}
	end := tick(20, 0, world.Vitals{Starved: 3})
	v.snap = &end
	for i := 0; i < graphTicks; i++ {
		v.record(&end)
	}
	v.draw()
	if text := screenText(v.screen.(tcell.SimulationScreen)); !strings.Contains(text, "died out at tick 20") {
		t.Fatalf("the map page does not say the settlement ended:\n%s", text)
	}
}

// A run that has ended says how it went in plain words too. The screen is
// torn down on the way out, so whatever the terminal keeps is the whole of
// what anyone can go back to.
func TestReportSaysHowTheRunWent(t *testing.T) {
	v := &view{}
	for i := 1; i <= 1000; i++ {
		s := tick(i, 12, world.Vitals{Births: i / 100})
		v.record(&s)
	}
	end := tick(1200, 0, world.Vitals{
		Births: 10, Starved: 20, Failed: 2,
		Gates: [world.GateCount]int{world.Hungry: 3},
	})
	end.Chronicle = []world.Note{{Tick: 1199, Kind: event.Died, Text: "Ada starved"}}
	v.snap = &end
	v.record(&end)

	got := v.Report()
	for _, want := range []string{
		"DIED OUT at tick 1200",
		"born 10, died 22 (starved 20, old age 2)",
		"hungry 3",
		"peak",
		"Ada starved",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("the report never says %q:\n%s", want, got)
		}
	}
}

// A run nobody watched still reports, rather than panicking on its way out.
func TestReportSurvivesAnEmptyRun(t *testing.T) {
	v := &view{}
	if got := v.Report(); !strings.Contains(got, "nothing ran") {
		t.Fatalf("an empty run reports %q", got)
	}
}

// The kept report carries the same run as numbers, so that a folder of them
// can be read without opening any. A run nobody watched scores nothing and
// does not panic trying.
func TestScoreCarriesTheRunAsNumbers(t *testing.T) {
	t.Setenv("LREAT_RUNS", t.TempDir())
	rep := report.Open("watch")

	empty := &view{}
	empty.score(rep)

	v := &view{}
	for i := 1; i <= 1000; i++ {
		s := tick(i, 12, world.Vitals{Births: i / 100})
		v.record(&s)
	}
	end := tick(1200, 0, world.Vitals{Births: 10, Starved: 20, Failed: 2})
	v.snap = &end
	v.record(&end)
	v.score(rep)

	if line := rep.Close(); line == "" {
		t.Fatal("the run was not kept")
	}
	paths, _ := filepath.Glob(filepath.Join(os.Getenv("LREAT_RUNS"), "*", "*.md"))
	if len(paths) != 1 {
		t.Fatalf("expected one report, found %v", paths)
	}
	b, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"died-out-at", "1200.000", "starved", "20.000", "peak"} {
		if !strings.Contains(string(b), want) {
			t.Errorf("the report never says %q:\n%s", want, b)
		}
	}
}
