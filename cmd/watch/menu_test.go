package main

import (
	"slices"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"

	"lreat/core/world"
)

func rune_(r rune) *tcell.EventKey { return tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone) }

func key(k tcell.Key) *tcell.EventKey { return tcell.NewEventKey(k, 0, tcell.ModNone) }

// press runs a sequence of keys through the menu and reports what it ended
// on, so that a test reads as the run of presses somebody would make.
func press(m *menuState, evs ...*tcell.EventKey) (done, start bool) {
	for _, ev := range evs {
		if done, start = m.key(ev); done {
			return done, start
		}
	}
	return false, false
}

// Enter on the front page's first line founds a settlement on the defaults
// without anyone having to visit the options at all.
func TestMenuStartsOnDefaults(t *testing.T) {
	s := defaults()
	m := &menuState{s: &s}
	done, start := press(m, key(tcell.KeyEnter))
	if !done || !start {
		t.Fatalf("enter on start gave done=%v start=%v, want both true", done, start)
	}
	if s != defaults() {
		t.Fatalf("starting changed the terms: %+v", s)
	}
}

// Quit leaves before there is a settlement, and so does esc on the front
// page: the menu is the last thing standing between the program and the
// terminal it was run from.
func TestMenuQuits(t *testing.T) {
	for _, ev := range []*tcell.EventKey{rune_('q'), key(tcell.KeyEscape)} {
		s := defaults()
		m := &menuState{s: &s}
		if done, start := press(m, ev); !done || start {
			t.Fatalf("%v gave done=%v start=%v, want done without starting", ev.Name(), done, start)
		}
	}
}

// The options are reached from the front page, and what is tuned there is
// what the settlement is founded on.
func TestMenuOptionsTune(t *testing.T) {
	s := defaults()
	m := &menuState{s: &s}
	// Down to "options", enter, then a seed typed straight in.
	press(m, key(tcell.KeyDown), key(tcell.KeyEnter))
	if !m.opts {
		t.Fatal("enter on options did not open the options page")
	}
	m.at = lineOf(t, "seed")
	press(m, rune_('4'), rune_('2'), rune_('7'))
	if s.seed == 427 {
		t.Fatal("a half-typed number took effect before the cursor left the line")
	}
	press(m, key(tcell.KeyDown))
	if s.seed != 427 {
		t.Fatalf("seed is %d after typing 427 and moving on, want 427", s.seed)
	}
	// One right on the figures adds a settler.
	m.at = lineOf(t, "figures")
	press(m, key(tcell.KeyRight))
	if s.agents != defaults().agents+1 {
		t.Fatalf("figures is %d, want %d", s.agents, defaults().agents+1)
	}
}

// Esc backs out of the options keeping everything set there — leaving the
// page is not changing one's mind about it — and lands on the line that
// opened them.
func TestMenuOptionsKeptOnEscape(t *testing.T) {
	s := defaults()
	m := &menuState{s: &s, opts: true, at: lineOf(t, "seed")}
	press(m, rune_('9'), key(tcell.KeyEscape))
	if m.opts {
		t.Fatal("esc did not leave the options page")
	}
	if front[m.at] != "options" {
		t.Fatalf("esc landed on %q, want options", front[m.at])
	}
	if s.seed != 9 {
		t.Fatalf("seed is %d after esc, want the typed 9 to have been kept", s.seed)
	}
}

// The start at the foot of the options is the whole point of tuning: a run
// is founded from the page it was set up on rather than by going back.
func TestMenuStartsFromOptions(t *testing.T) {
	s := defaults()
	m := &menuState{s: &s, opts: true, at: len(options())}
	done, start := press(m, key(tcell.KeyEnter))
	if !done || !start {
		t.Fatalf("enter on the options' start gave done=%v start=%v, want both true", done, start)
	}
}

// lineOf is where an option sits on the page. Tests ask for it by name:
// the page grows a line now and then, and a test that counted them would
// fail for the wrong reason every time it did.
func lineOf(t *testing.T, name string) int {
	t.Helper()
	for i, o := range options() {
		if o.name == name {
			return i
		}
	}
	t.Fatalf("no %q line on the options page", name)
	return 0
}

// Choosing is a toggle: enter is what moves it, and the rule it is left on
// is the one the world is given.
func TestMenuTogglesChoiceRule(t *testing.T) {
	s := defaults()
	m := &menuState{s: &s, opts: true, at: lineOf(t, "choosing")}
	press(m, key(tcell.KeyEnter))
	if s.fit {
		t.Fatal("enter did not move the choice rule off recognition")
	}
	if got := options()[m.at].show(&s); got != "value" {
		t.Fatalf("the line reads %q, want value", got)
	}
}

// d puts everything back, which is what makes tuning safe to poke at.
func TestMenuDefaultsRestore(t *testing.T) {
	s := defaults()
	s.agents, s.seed, s.fit = 300, 77, false
	m := &menuState{s: &s, opts: true}
	press(m, rune_('d'))
	if s != defaults() {
		t.Fatalf("d left %+v, want the defaults back", s)
	}
}

// The page says what the option under the cursor does. A menu that only
// names its settings is a list of words to look up somewhere else.
func TestMenuDrawsHelpForTheLine(t *testing.T) {
	sc := tcell.NewSimulationScreen("UTF-8")
	if err := sc.Init(); err != nil {
		t.Fatal(err)
	}
	defer sc.Fini()
	sc.SetSize(100, 30)

	s := defaults()
	m := &menuState{screen: sc, s: &s, opts: true, at: lineOf(t, "seed")}
	m.draw()

	text := screenText(sc)
	if !strings.Contains(text, "seed") || !strings.Contains(text, "source of chance") {
		t.Fatalf("options page does not explain the seed:\n%s", text)
	}
	if !strings.Contains(text, "start") {
		t.Fatal("options page offers no way to start")
	}
}

// The front page says what starting would do before anyone commits to it.
func TestMenuFrontShowsTerms(t *testing.T) {
	sc := tcell.NewSimulationScreen("UTF-8")
	if err := sc.Init(); err != nil {
		t.Fatal(err)
	}
	defer sc.Fini()
	sc.SetSize(100, 30)

	s := defaults()
	s.seed, s.agents = 7, 33
	m := &menuState{screen: sc, s: &s}
	m.draw()

	text := screenText(sc)
	for _, want := range []string{"start", "options", "quit", "seed 7", "33 figures"} {
		if !strings.Contains(text, want) {
			t.Fatalf("front page has no %q:\n%s", want, text)
		}
	}
}

// A fitted map leaves room for everything drawn around it: the panel beside
// it, and the legend, graph and span under it. Anything larger and the view
// refuses to draw at all, which is the failure this option exists to end.
func TestFitMapLeavesRoomForTheView(t *testing.T) {
	for _, size := range [][2]int{{120, 40}, {200, 60}, {90, 30}} {
		w, h := fitMap(size[0], size[1])
		if w+panelWidth > size[0] {
			t.Fatalf("a %dx%d terminal was given a %d-wide map, %d too wide",
				size[0], size[1], w, w+panelWidth-size[0])
		}
		if h+graphHeight+3 > size[1] {
			t.Fatalf("a %dx%d terminal was given a %d-deep map, %d too deep",
				size[0], size[1], h, h+graphHeight+3-size[1])
		}
	}
}

// A terminal too small to hold even the smallest map still gets one: a bad
// map is a settlement that cannot be read, but no map is a program that
// will not run.
func TestFitMapHasAFloor(t *testing.T) {
	w, h := fitMap(10, 4)
	if w < 20 || h < 10 {
		t.Fatalf("a tiny terminal got a %dx%d map, want the floor of 20x10", w, h)
	}
}

// With fitting on, the width and height are read off the window every time
// the page is drawn, and there is no setting them by hand.
func TestMenuFitsMapToTerminal(t *testing.T) {
	sc := tcell.NewSimulationScreen("UTF-8")
	if err := sc.Init(); err != nil {
		t.Fatal(err)
	}
	defer sc.Fini()
	sc.SetSize(140, 50)

	s := defaults()
	if !s.snug {
		t.Fatal("a settlement is founded on the window by default; it was not")
	}
	m := &menuState{screen: sc, s: &s, opts: true, at: lineOf(t, "map")}
	m.draw()
	wantW, wantH := fitMap(140, 50)
	if s.width != wantW || s.height != wantH {
		t.Fatalf("fitting a 140x50 terminal gave a %dx%d map, want %dx%d", s.width, s.height, wantW, wantH)
	}

	// The window is what says how big the map is now: the arrows and a
	// typed number both leave it alone.
	m.at = lineOf(t, "map width")
	press(m, key(tcell.KeyRight), rune_('5'), rune_('0'), key(tcell.KeyDown))
	if s.width != wantW {
		t.Fatalf("map width is %d after being pushed at by hand, want the window's %d", s.width, wantW)
	}

	// A window resized under the menu is answered by the next draw.
	sc.SetSize(100, 30)
	m.draw()
	wantW, wantH = fitMap(100, 30)
	if s.width != wantW || s.height != wantH {
		t.Fatalf("after a resize the map is %dx%d, want %dx%d", s.width, s.height, wantW, wantH)
	}
	if text := screenText(sc); !strings.Contains(text, "from the window") {
		t.Fatalf("the page does not say where the size came from:\n%s", text)
	}

	// Taken off the window's hands, the two lines answer to a hand again.
	m.at = lineOf(t, "map")
	press(m, key(tcell.KeyEnter))
	m.at = lineOf(t, "map width")
	press(m, rune_('9'), rune_('0'), key(tcell.KeyDown))
	if s.width != 90 {
		t.Fatalf("map width is %d after fitting was turned off and 90 typed, want 90", s.width)
	}
}

// What the command line asks for about the map's size, and what it means.
// A stated size is meant, so it takes the map off the window; -fit said
// outright settles it either way, including in favour of the window over a
// size stated alongside it.
func TestSnugFromFlags(t *testing.T) {
	for _, c := range []struct {
		what              string
		fit, sized, given bool
		want              bool
	}{
		{"nothing said at all", true, false, false, true},
		{"-fit=false", false, false, true, false},
		{"-width 120", true, true, false, false},
		{"-width 120 -fit", true, true, true, true},
		{"-width 120 -fit=false", false, true, true, false},
	} {
		if got := snugFrom(c.fit, c.sized, c.given); got != c.want {
			t.Errorf("%s: the map comes off the window = %v, want %v", c.what, got, c.want)
		}
	}
}

// A heading stands over a run of lines, so every option belongs to one of
// the known groups and each group comes out in one piece. Written in any
// order, they are sorted into their groups; a line whose group were
// misspelt would fall out of every run and the page would say a heading
// twice.
func TestOptionsComeOutGrouped(t *testing.T) {
	var order []string
	for _, o := range options() {
		if !slices.Contains(groups, o.group) {
			t.Fatalf("the %q line is under %q, which is no heading on the page", o.name, o.group)
		}
		if len(order) == 0 || order[len(order)-1] != o.group {
			order = append(order, o.group)
		}
	}
	for i, g := range order {
		if slices.Index(order, g) != i {
			t.Fatalf("%q is a heading in two places: %v", g, order)
		}
	}
}

// The options page is read down under its headings rather than as one list
// of eight settings.
func TestMenuDrawsHeadings(t *testing.T) {
	sc := tcell.NewSimulationScreen("UTF-8")
	if err := sc.Init(); err != nil {
		t.Fatal(err)
	}
	defer sc.Fini()
	sc.SetSize(110, 34)

	s := defaults()
	m := &menuState{screen: sc, s: &s, opts: true}
	m.draw()

	text := screenText(sc)
	for _, g := range groups {
		if !strings.Contains(text, g) {
			t.Fatalf("the page has no %q heading:\n%s", g, text)
		}
	}
}

// The world line chooses what is being founded. A globe is not sized to the
// terminal like a valley is: it comes at the size that makes it a globe, and
// the two lines under the map line say so rather than offering a number
// nobody can move.
func TestChoosingAGlobeTakesItsOwnSize(t *testing.T) {
	s := defaults()
	m := &menuState{s: &s, opts: true, at: lineOf(t, "world")}
	stepTo(t, m, &s, "globe")
	cfg := s.config()
	if !cfg.Wrap {
		t.Fatal("a globe was founded with edges")
	}
	// A terminal's size has no say in it.
	s.measure(80, 24)
	if s.width != cfg.Width || s.height != cfg.Height {
		t.Fatalf("a globe measured against an 80x24 terminal came to %dx%d, want %dx%d",
			s.width, s.height, cfg.Width, cfg.Height)
	}
	if !s.sized() {
		t.Fatal("the map size is offered as something to set by hand on a globe")
	}
	for _, name := range []string{"map width", "map height"} {
		o := options()[lineOf(t, name)]
		if o.settable(&s) {
			t.Fatalf("the %s line takes a hand on it on a globe", name)
		}
	}
	// And back again: the valley is sized to the terminal as it always was.
	stepTo(t, m, &s, "valley")
	s.measure(120, 40)
	w, h := fitMap(120, 40)
	if s.width != w || s.height != h {
		t.Fatalf("the valley came to %dx%d, want the window's %dx%d", s.width, s.height, w, h)
	}
}

// Every world the menu offers is one the world package will actually found.
// The two lists are apart, so this is what keeps them from drifting.
func TestEveryWorldOfferedCanBeFounded(t *testing.T) {
	for _, name := range presets {
		if _, ok := world.Preset(name); !ok {
			t.Fatalf("the menu offers %q, which is no world", name)
		}
	}
	s := defaults()
	if !slices.Contains(presets, s.world()) {
		t.Fatalf("a settlement is founded in the %s by default, which the menu cannot reach", s.world())
	}
}

// stepTo walks the world line round to the world named, however many others
// stand between: a test about what a globe is should not also be a test of
// what order the worlds are offered in.
func stepTo(t *testing.T, m *menuState, s *setup, name string) {
	t.Helper()
	for range presets {
		if s.world() == name {
			return
		}
		press(m, key(tcell.KeyRight))
	}
	t.Fatalf("stepping the world line never reached %q; it is on %q", name, s.world())
}
