package main

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"

	"github.com/gdamore/tcell/v2"

	"lreat/core/world"
)

// The start screen is the front door. A settlement is founded on a handful
// of decisions — how many people, how much ground, which seed, and by what
// rule anyone chooses anything — and until now the only place to make them
// was the command line, which meant knowing every flag before the first run
// and restarting the program to change one. The menu asks the same
// questions in the place where the answers are needed, and answers them
// itself for anyone who would rather just watch a settlement.
//
// Options are laid out one to a line with what they do written under the
// one being looked at, because these are not settings so much as the terms
// the run is founded on: what temperature does to a settlement is not
// something the word "temperature" says.

// setup is everything decided before a world exists. Flags fill it, the
// menu edits it, and main founds the settlement out of it.
type setup struct {
	seed   uint64
	agents int
	width  int
	height int
	// snug takes the map's size off the terminal instead of the width and
	// height above, which are then whatever the window last came to. It is
	// how a settlement is founded unless somebody says otherwise: a map the
	// window cannot hold is drawn as an apology and nothing else, and one
	// much smaller than the window wastes ground nobody asked to do
	// without, and neither number is one anybody can know before the
	// program is looking at the terminal it was run in. Nothing that has to
	// be reproduced is founded here — headless and tune both take the fixed
	// default size — so what this view owes is the window it is in.
	snug bool
	tps  float64
	fit  bool
	temp float64
}

// fitMap is the largest map a terminal of this size will hold: the panel
// stands beside it and the legend, the activity graph and its span run
// under it, and all of that has to fit as well. It is the same reckoning
// the view makes before it refuses to draw — see draw in main.go — read
// the other way round, from the window to the map rather than from the map
// to a complaint.
func fitMap(w, h int) (int, int) {
	return clampInt(w-panelWidth, 20, 400), clampInt(h-graphHeight-3, 10, 200)
}

// defaults are what start founds a settlement on: the same run the command
// used to make with no flags at all.
func defaults() setup {
	r := world.DefaultRules()
	return setup{
		seed:   1,
		agents: 20,
		// The width and height stand behind the fitting as what a map is
		// when somebody takes it off the window's hands.
		width:  world.DefaultWidth,
		height: world.DefaultHeight,
		snug:   true,
		tps:    20,
		fit:    r.Fit,
		temp:   r.Temperature,
	}
}

// option is one line of the options page: what it is called, what it says
// about itself, how it reads, and how the arrow keys move it. digits is how
// a number typed straight in lands, nil on the lines that are not numbers.
type option struct {
	// group is the heading this line is drawn under. The lines are kept in
	// their groups' order, so a group is a run of them rather than a
	// scattering.
	group  string
	name   string
	help   string
	show   func(*setup) string
	step   func(*setup, int)
	digits func(*setup, uint64)
	// fixed says this line is being filled in from somewhere else at the
	// moment — the terminal's own size — and is read rather than set. It is
	// nil on the lines that are never anything but set by hand.
	fixed func(*setup) bool
}

// settable is whether a line takes a hand on it as things stand.
func (o option) settable(s *setup) bool { return o.fixed == nil || !o.fixed(s) }

// groups are the questions the options answer, in the order they are asked:
// what world this is, who is in it, and how it is to be watched. Everything
// on the page belongs to one of them.
var groups = []string{"the world", "the people", "the watching"}

func options() []option {
	all := []option{{
		group: "the world",
		name:  "seed",
		help:  "the world's one source of chance: the same seed is the same run, every time (r deals a fresh one)",
		show:  func(s *setup) string { return fmt.Sprintf("%d", s.seed) },
		step: func(s *setup, d int) {
			if d < 0 && s.seed == 0 {
				return
			}
			s.seed = uint64(int64(s.seed) + int64(d))
		},
		digits: func(s *setup, n uint64) { s.seed = n },
	}, {
		group:  "the people",
		name:   "figures",
		help:   "how many people the settlement is founded with; too few and one bad winter ends it",
		show:   func(s *setup) string { return fmt.Sprintf("%d", s.agents) },
		step:   func(s *setup, d int) { s.agents = clampInt(s.agents+d, 1, 500) },
		digits: func(s *setup, n uint64) { s.agents = clampInt(int(n), 1, 500) },
	}, {
		group: "the world",
		name:  "map",
		help:  "fit takes the ground off the window this is running in, as large as it will hold, and is how a settlement is founded unless this says otherwise; by hand keeps the two lines under it whatever the window is",
		show: func(s *setup) string {
			if s.snug {
				return "fit to terminal"
			}
			return "by hand"
		},
		step: func(s *setup, _ int) { s.snug = !s.snug },
	}, {
		group: "the world",
		name:  "map width",
		help:  "how wide the ground is; a bigger map is more forest to walk to and more room to spread into",
		show:  func(s *setup) string { return fmt.Sprintf("%d", s.width) },
		step: func(s *setup, d int) {
			if s.snug {
				return // the terminal is setting this; see the map line
			}
			s.width = clampInt(s.width+4*d, 20, 400)
		},
		digits: func(s *setup, n uint64) {
			if !s.snug {
				s.width = clampInt(int(n), 20, 400)
			}
		},
		// A line the terminal is filling in is shown as read rather than as
		// set: the number is true, and moving it would be a lie.
		fixed: func(s *setup) bool { return s.snug },
	}, {
		group: "the world",
		name:  "map height",
		help:  "how deep the ground is; the map, the panel beside it and the graph under it all have to fit the terminal",
		show:  func(s *setup) string { return fmt.Sprintf("%d", s.height) },
		step: func(s *setup, d int) {
			if s.snug {
				return
			}
			s.height = clampInt(s.height+2*d, 10, 200)
		},
		digits: func(s *setup, n uint64) {
			if !s.snug {
				s.height = clampInt(int(n), 10, 200)
			}
		},
		fixed: func(s *setup) bool { return s.snug },
	}, {
		group: "the watching",
		name:  "speed",
		help:  "days per second to begin at; + and - change it again while the settlement runs",
		show:  func(s *setup) string { return fmt.Sprintf("%.2f t/s", s.tps) },
		step: func(s *setup, d int) {
			if d > 0 {
				s.tps *= 2
			} else {
				s.tps /= 2
			}
			s.tps = clampFloat(s.tps, 0.25, 512)
		},
	}, {
		group: "the people",
		name:  "choosing",
		help:  "recognition takes the action whose habit fits the moment; value prices every option, the older rule",
		show: func(s *setup) string {
			if s.fit {
				return "recognition"
			}
			return "value"
		},
		step: func(s *setup, _ int) { s.fit = !s.fit },
	}, {
		group: "the people",
		name:  "temperature",
		help:  "how loosely recognition is followed: 0 always takes the best fit, higher wanders further from it",
		show: func(s *setup) string {
			if !s.fit {
				return "—"
			}
			return fmt.Sprintf("%.2f", s.temp)
		},
		step: func(s *setup, d int) { s.temp = clampFloat(s.temp+0.05*float64(d), 0, 2) },
	}}
	// The lines are written above in the order they read best beside one
	// another and come out in their groups' order, because a heading stands
	// over a run of lines: a stray line falling outside its own run would
	// have the page saying "the world" twice.
	slices.SortStableFunc(all, func(a, b option) int {
		return slices.Index(groups, a.group) - slices.Index(groups, b.group)
	})
	return all
}

func clampInt(v, lo, hi int) int { return min(max(v, lo), hi) }

func clampFloat(v, lo, hi float64) float64 { return min(max(v, lo), hi) }

// menu runs the start screen on an already-initialised screen and reports
// whether to go on and found the settlement; false is the user leaving
// before there is one. It edits s in place, so whatever the flags said is
// what the options page opens on.
func menu(sc tcell.Screen, s *setup) bool {
	m := &menuState{screen: sc, s: s}
	for {
		m.draw()
		ev, ok := sc.PollEvent().(*tcell.EventKey)
		if !ok {
			sc.Sync() // a resize or a click: take the screen again and redraw
			continue
		}
		if done, start := m.key(ev); done {
			return start
		}
	}
}

// The page is drawn with the same restraint as the settlement's own panel:
// a dim label, the value where the eye already is, and the cursor the only
// bright thing on the screen. The line under the options is what the option
// the cursor is on actually does, because a menu that only names its
// settings is a list of words to look up elsewhere.
const (
	// The page is laid out in a block of its own rather than against the
	// left edge: a start screen in a full-screen terminal is otherwise a
	// handful of words in one corner of an empty field.
	menuWidth = 64
	menuMark  = 2  // the cursor's own column
	menuName  = 4  // the names, indented under their heading
	menuValue = 18 // where the values stand, so the page reads down as well as across
)

// The page borrows the settlement's own colours rather than inventing any:
// the title in the green of the ground, the cursor in the yellow of a field,
// and everything that is there to be read rather than chosen kept dim. The
// cursor is the only moving thing on the screen and wants to be findable
// without being looked for.
var (
	titleStyle  = tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true)
	markStyle   = tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
	chosenStyle = tcell.StyleDefault.Bold(true)
	dimStyle    = tcell.StyleDefault.Dim(true)
)

func (m *menuState) draw() {
	sc := m.screen
	sc.Clear()
	w, h := sc.Size()
	// A fitted map is measured every time the page is drawn rather than
	// once when it is asked for, so that a window resized with the menu
	// open shows the ground it would now be given.
	if m.s.snug {
		m.s.width, m.s.height = fitMap(w, h)
	}
	m.x = max(0, (w-menuWidth)/2)
	line := max(1, (h-m.depth())/2)

	puts(sc, m.x, line, titleStyle, "lreat")
	line++
	puts(sc, m.x, line, dimStyle, "a settlement of people living in a world, and nobody playing it")
	line += 2

	if m.opts {
		m.drawOptions(&line, h)
	} else {
		m.drawFront(&line)
	}

	// The keys are named in words for the same reason the page has no rules
	// drawn on it: an arrow is a glyph some terminals give two columns to,
	// and the legend would come apart on those.
	keys := [][2]string{{"arrows", "move"}, {"enter", "choose"}, {"q", "quit"}}
	if m.opts {
		keys = [][2]string{{"arrows", "move and change"}, {"0-9", "type"},
			{"d", "defaults"}, {"esc", "back"}, {"q", "quit"}}
	}
	legend(sc, m.x, max(line+1, h-1), keys)
	sc.Show()
}

// depth is roughly how tall the page is, which is all the centring needs:
// a line or two out either way is not something the eye picks up, and a
// page that jumped as its help text rewrapped would be.
func (m *menuState) depth() int {
	if m.opts {
		return len(options()) + len(groups) + 13
	}
	return len(front) + 10
}

// heading is what a run of lines is gathered under. It is the only thing
// dividing one part of the page from another: a drawn rule would be a line
// of box-drawing glyphs, and this page is laid out by counting columns —
// anything the terminal decides to give two of them to walks the rest of
// the line sideways. Blank space and a dim word do the same work and cannot
// be measured wrong.
func (m *menuState) heading(line *int, s string) {
	puts(m.screen, m.x, *line, dimStyle, s)
	*line++
}

// row draws one line of a page: the cursor in its own column, the name, and
// the value out at the column they all share. The cursor keeps a column to
// itself because a mark that pushed its line sideways would make the page
// shuffle under the eye every time it moved.
func (m *menuState) row(line *int, on bool, name string, nameStyle tcell.Style, value string, valueStyle tcell.Style) {
	sc := m.screen
	if on {
		puts(sc, m.x+menuMark, *line, markStyle, "▸")
	}
	puts(sc, m.x+menuName, *line, nameStyle, name)
	puts(sc, m.x+menuValue, *line, valueStyle, trim(value, max(1, menuWidth-menuValue)))
	*line++
}

// legend writes the keys along the foot of the page, each standing out of
// the dim word for what it does.
func legend(sc tcell.Screen, x, y int, keys [][2]string) {
	for _, k := range keys {
		puts(sc, x, y, tcell.StyleDefault, k[0])
		x += len([]rune(k[0])) + 1
		puts(sc, x, y, dimStyle, k[1])
		x += len([]rune(k[1])) + 3
	}
}

// drawFront is the front page: the three things to do, and under them the
// terms the settlement would be founded on if it were started now. Start is
// never a leap in the dark — what it would do is written under it.
func (m *menuState) drawFront(line *int) {
	labels := map[string]string{
		"start":   "found a settlement and watch it",
		"options": "set the terms it is founded on",
		"quit":    "leave",
	}
	for i, name := range front {
		style := tcell.StyleDefault
		if i == m.at {
			style = chosenStyle
		}
		m.row(line, i == m.at, name, style, labels[name], dimStyle)
	}
	*line++
	m.heading(line, "as it stands")

	s := m.s
	ground := fmt.Sprintf("%d by %d", s.width, s.height)
	if s.snug {
		ground += " from the window"
	}
	rule := "recognition"
	if !s.fit {
		rule = "value"
	}
	puts(m.screen, m.x+menuName, *line, dimStyle,
		fmt.Sprintf("seed %d   %d figures   %s", s.seed, s.agents, ground))
	*line++
	puts(m.screen, m.x+menuName, *line, dimStyle,
		fmt.Sprintf("%.2f t/s   %s", s.tps, rule))
	*line += 2
}

// drawOptions is the options page: the terms under the headings they belong
// to, the start under them, and what the line the cursor is on means under
// that. The headings are there because these are not eight settings but
// three questions — what world, what people, and how it is to be watched.
func (m *menuState) drawOptions(line *int, h int) {
	opts := options()
	group := ""
	for i, o := range opts {
		if o.group != group {
			group = o.group
			if i > 0 {
				*line++
			}
			m.heading(line, group)
		}
		style := tcell.StyleDefault
		if i == m.at {
			style = chosenStyle
		}
		value := o.show(m.s)
		if i == m.at && m.typed != "" {
			// What is being typed stands where the value does, marked as
			// unfinished: it is not the setting until the cursor leaves.
			value = m.typed + "_"
		}
		if !o.settable(m.s) {
			// The terminal is filling this one in: the number is true and
			// there is nothing to be done to it, so it is dim even under
			// the cursor, and says where it came from.
			style = dimStyle
			value += "   from the window"
		}
		m.row(line, i == m.at, o.name, dimStyle, value, style)
	}
	*line += 2

	at := m.at >= len(opts)
	style := tcell.StyleDefault
	if at {
		style = chosenStyle
	}
	m.row(line, at, "start", style, "found the settlement on these terms", dimStyle)
	*line += 2

	help := "found the settlement on the terms above"
	if m.at < len(opts) {
		help = opts[m.at].help
	}
	for _, s := range wrap(help, menuWidth-menuName) {
		if *line >= h-1 {
			break
		}
		puts(m.screen, m.x+menuName, *line, dimStyle, s)
		*line++
	}
}

// wrap breaks a line of help into lines that fit, on spaces. The help is
// written as sentences rather than as labels, and a sentence cut off at the
// edge of the terminal explains nothing.
func wrap(s string, width int) []string {
	var lines []string
	line := ""
	for _, word := range strings.Fields(s) {
		switch {
		case line == "":
			line = word
		case len([]rune(line))+1+len([]rune(word)) <= width:
			line += " " + word
		default:
			lines = append(lines, line)
			line = word
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

// menuState is where the cursor is: which page, which line, and the digits
// typed at that line so far. Typing is kept apart from the value itself so
// that a half-typed number is never a world size.
type menuState struct {
	screen tcell.Screen
	s      *setup
	// opts is the options page rather than the front page, at is the line
	// the cursor is on, and typed is the number being entered there.
	opts  bool
	at    int
	typed string
	// x is the left edge of the block the page was last drawn in, which
	// follows the window's width. See draw.
	x int
}

// front is the three things that can be done before a settlement exists.
var front = []string{"start", "options", "quit"}

// rows is how many lines the current page has. The options page carries one
// more than it has options: the start at the foot of it, so that tuning and
// founding are one movement down the page rather than a trip back.
func (m *menuState) rows() int {
	if m.opts {
		return len(options()) + 1
	}
	return len(front)
}

// key acts on a press. It returns done when the menu is over, and start when
// what is over it is a settlement rather than the program.
func (m *menuState) key(ev *tcell.EventKey) (done, start bool) {
	switch {
	case ev.Key() == tcell.KeyCtrlC:
		return true, false
	case ev.Key() == tcell.KeyUp || ev.Key() == tcell.KeyBacktab:
		m.move(-1)
	case ev.Key() == tcell.KeyDown || ev.Key() == tcell.KeyTab:
		m.move(1)
	case ev.Key() == tcell.KeyEscape:
		if !m.opts {
			return true, false
		}
		// Esc backs out of the options onto the line that opened them,
		// keeping everything tuned there: leaving the page is not
		// changing one's mind about it.
		m.commit()
		m.opts, m.at = false, 1
	case ev.Key() == tcell.KeyLeft:
		m.adjust(-1)
	case ev.Key() == tcell.KeyRight:
		m.adjust(1)
	case ev.Key() == tcell.KeyEnter || ev.Rune() == ' ':
		return m.enter()
	case ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2:
		if m.typed != "" {
			m.typed = m.typed[:len(m.typed)-1]
		}
	case ev.Rune() >= '0' && ev.Rune() <= '9':
		m.digit(ev.Rune())
	case ev.Rune() == 'r' && m.opts && m.at == 0:
		// A seed nobody chose is the commonest thing to want and the most
		// tedious to type: r deals a fresh one, short enough to write down.
		m.s.seed, m.typed = rand.Uint64N(100000), ""
	case ev.Rune() == 'd':
		*m.s, m.typed = defaults(), ""
	case ev.Rune() == 'q':
		return true, false
	}
	return false, false
}

func (m *menuState) move(d int) {
	m.commit()
	n := m.rows()
	m.at = (m.at + d + n) % n
}

// enter acts on the line the cursor is on. On the front page that is the
// choice itself; on the options page it founds the settlement from the last
// line and otherwise only moves the toggles, which have nowhere to be
// adjusted to but the other side.
func (m *menuState) enter() (done, start bool) {
	m.commit()
	if !m.opts {
		switch front[m.at] {
		case "start":
			return true, true
		case "options":
			m.opts, m.at = true, 0
		case "quit":
			return true, false
		}
		return false, false
	}
	opts := options()
	if m.at >= len(opts) {
		return true, true // the start at the foot of the options
	}
	if opts[m.at].digits == nil {
		opts[m.at].step(m.s, 1)
	}
	return false, false
}

func (m *menuState) adjust(d int) {
	if !m.opts {
		return
	}
	opts := options()
	if m.at >= len(opts) {
		return
	}
	m.typed = "" // the arrows move the value itself, not the number being typed
	opts[m.at].step(m.s, d)
}

// digit takes a number typed straight at a line. It is held as text until
// the cursor leaves the line, so that typing 120 does not pass through 1
// and 12 and get clamped on the way.
func (m *menuState) digit(r rune) {
	if !m.opts {
		return
	}
	opts := options()
	if m.at >= len(opts) || opts[m.at].digits == nil || !opts[m.at].settable(m.s) {
		return
	}
	if len(m.typed) < 18 { // beyond that no uint64 will hold it
		m.typed += string(r)
	}
}

// commit lands whatever has been typed at the current line. Everything that
// moves the cursor or founds a settlement goes through it, so a number
// typed and walked away from is a number that took.
func (m *menuState) commit() {
	typed := m.typed
	m.typed = ""
	if typed == "" || !m.opts {
		return
	}
	opts := options()
	if m.at >= len(opts) || opts[m.at].digits == nil {
		return
	}
	var n uint64
	if _, err := fmt.Sscanf(typed, "%d", &n); err == nil {
		opts[m.at].digits(m.s, n)
	}
}
