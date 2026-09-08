package main

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
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
	press(m, rune_('4'), rune_('2'), rune_('7'))
	if s.seed == 427 {
		t.Fatal("a half-typed number took effect before the cursor left the line")
	}
	press(m, key(tcell.KeyDown))
	if s.seed != 427 {
		t.Fatalf("seed is %d after typing 427 and moving on, want 427", s.seed)
	}
	// The cursor is on the figures now: one right adds a settler.
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
	m := &menuState{s: &s, opts: true}
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

// Choosing is a toggle: enter is what moves it, and the rule it is left on
// is the one the world is given.
func TestMenuTogglesChoiceRule(t *testing.T) {
	s := defaults()
	m := &menuState{s: &s, opts: true, at: 5}
	if options()[m.at].name != "choosing" {
		t.Fatalf("line 5 is %q, not the choice rule", options()[m.at].name)
	}
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
	m := &menuState{screen: sc, s: &s, opts: true}
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
