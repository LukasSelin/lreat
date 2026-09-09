// Command watch is a live terminal view of a settlement developing, in the
// spirit of Dwarf Fortress. The map is on the left with aggregate state
// beside it, and under the map a woven band of what the settlement has been
// spending itself on, running the map's whole width. It is an omniscient
// view for insight while tuning; the player's own view will be far narrower.
//
// Tab picks a figure out of the crowd and opens it up beside the map: who it
// is, what it has grown good at, what errand it is on, and — decision by
// decision — everything it weighed before setting out. A settlement is only
// ever the sum of those; without a way to read one of them the map is a
// weather system.
//
// D swaps the map for the settlement's vital record: the population curve,
// what people have died of, and what stood between everyone still alive and
// a child. W swaps it for what became of the ground it stands on: the wood
// it has taken, the fields and houses and roads it has put there, what food
// costs, what it knows, and what it has been spending its people on. The map
// says a settlement has stopped; those two say why, and they are the reason
// a run that ends is worth reading rather than restarting.
//
// Every page here shows a dozen measurements at once and squeezes each of
// them into a row or a band. The arrows step through whatever the page has
// and open the one stepped onto out over the page's largest space, scaled to
// its own high-water mark: a kind of work holding a twentieth of the
// population is not drawn at all in a weave shared with six others, and is a
// chart of its own when it is stepped onto. See focus.go.
//
// Keys: space pauses, + and - change speed, . steps once while paused,
// r lays streets through the settlement, tab and shift-tab pick an agent
// (or click one), up and down open a graph out, esc backs off the graph and
// then the page, d shows the vitals, w the world, q quits.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"

	"lreat/core/clock"
	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/observe"
	"lreat/core/sim"
	"lreat/core/world"
	"lreat/report"
	"lreat/ui/ascii"
)

var names = []string{
	"Ada", "Bo", "Cai", "Dag", "Eli", "Fen", "Gus", "Hal", "Ivo", "Jun",
	"Kai", "Lin", "Mo", "Nia", "Odd", "Pim", "Quin", "Rui", "Sol", "Tam",
	"Uma", "Vic", "Wen", "Xin", "Yara", "Zed",
}

const (
	panelWidth = 38
	// The activity graph runs under the map at the map's own width. Each
	// column is the mean of graphTicks days; graphMax columns are kept so
	// a wider map simply shows more of the same history. At twenty days a
	// column a map-wide graph covers some four years, which is the span a
	// settlement's changes of habit actually show up over.
	graphHeight = 6
	graphTicks  = 20
	graphMax    = 320
	// The settlement's figures sit in two columns of label and
	// right-aligned number, both halves the same shape so the numbers line
	// up down the panel.
	statLabel = 10
	statValue = 6
	statCol   = statLabel + statValue + 2
)

// main opens the start screen and, if a settlement is started from it,
// founds one on whatever terms the menu was left holding. The flags are
// still there and still mean what they meant; they are the menu's opening
// position now rather than the only way to say any of it, and -start skips
// the menu for the runs that are launched from a shell script rather than
// by hand. See menu.go.
func main() {
	d := defaults()
	seed := flag.Uint64("seed", d.seed, "world seed")
	agents := flag.Int("agents", d.agents, "starting population")
	tps := flag.Float64("tps", d.tps, "initial ticks per second")
	width := flag.Int("width", d.width, "map width")
	height := flag.Int("height", d.height, "map height")
	value := flag.Bool("value", false, "agents choose by expected value, the original rule, instead of by recognition")
	temp := flag.Float64("temp", d.temp, "base temperature of recognition; 0 always takes the best fit")
	snug := flag.Bool("fit", d.snug, "size the map to the terminal; -fit=false takes -width and -height instead")
	skip := flag.Bool("start", false, "start straight away, without the menu")
	flag.Parse()
	s := setup{
		seed: *seed, agents: *agents, tps: *tps,
		width: *width, height: *height, snug: *snug,
		fit: !*value, temp: *temp,
	}
	s.snug = snugFrom(*snug, given("width") || given("height"), given("fit"))

	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err := screen.Init(); err != nil {
		log.Fatal(err)
	}
	if !*skip && !menu(screen, &s) {
		screen.Fini() // left from the menu: there was never a settlement
		return
	}
	run(screen, s)
}

// choosing names the rule the figures decided under, for a report that has
// to say afterwards what the run was of.
func choosing(fit bool) string {
	if fit {
		return "recognition"
	}
	return "value"
}

// given says whether a flag was named on the command line rather than left
// at what it defaults to.
func given(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// snugFrom is whether the map comes off the window, out of what -fit says
// and what else was named alongside it. A size asked for by name is a size
// meant: -width or -height takes the map off the window without anyone
// having to say -fit=false as well. Naming -fit outright settles it either
// way, so that a run can ask for both and get the window.
func snugFrom(fit, sized, fitGiven bool) bool {
	if sized && !fitGiven {
		return false
	}
	return fit
}

// run founds the settlement and watches it until the user quits.
func run(screen tcell.Screen, s setup) {
	// A fitted map is measured here, against the terminal as it stands at
	// the moment of founding, because that is the last moment it can be:
	// the ground is generated once and a window resized afterwards finds
	// the map it was given rather than the map it would now ask for.
	if s.snug {
		s.width, s.height = fitMap(screen.Size())
	}
	w := world.NewSized(s.seed, s.width, s.height)
	w.Rules.Fit = s.fit
	w.Rules.Temperature = s.temp
	for i := 0; i < s.agents; i++ {
		w.Spawn(fmt.Sprintf("%s%d", names[i%len(names)], i/len(names)), w.RandomPersonality())
	}
	runner := sim.New(w, s.tps)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go runner.Run(ctx)

	v := &view{
		speed: s.tps,
		// The panel reads one agent at a time straight off the simulation
		// goroutine rather than out of the snapshot: a portrait is far more
		// than the map needs, and nobody is looking at all of them at once.
		look: func(id entity.ID) *observe.Portrait {
			var p *observe.Portrait
			runner.Inspect(func(w *world.World) { p = observe.Look(w, id) })
			return p
		},
		follow: func(id entity.ID) {
			runner.Send(sim.Func(func(w *world.World) { w.Watch(id) }))
		},
	}
	// However the run ends, it says how it went on the way out — and it is
	// registered before the screen is handed back, so that, deferred calls
	// running in reverse, it prints to the terminal the screen has just
	// given up rather than into a display about to be torn down. A
	// settlement nobody can report on afterwards is one nobody can tune
	// against. See Report in vitals.go.
	//
	// The same words are kept on disk, because a watched run is founded on
	// whatever the menu was left at rather than on flags, and a terminal
	// scrolled past is the only other place that ever said so. The terms
	// go in first: without them the numbers under them are of nothing.
	rep := report.Open("watch")
	defer func() {
		out := rep.Out(os.Stdout)
		fmt.Fprintf(out, "\nfounded on seed %d, %d figures, %dx%d, temp %.2f, %s\n",
			s.seed, s.agents, s.width, s.height, s.temp, choosing(s.fit))
		fmt.Fprint(out, v.Report())
		v.score(rep)
		fmt.Print(rep.Close())
	}()
	defer screen.Fini()
	screen.EnableMouse(tcell.MouseButtonEvents) // clicking a figure picks it

	keys := make(chan tcell.Event, 16)
	go func() {
		for {
			keys <- screen.PollEvent()
		}
	}()

	v.screen = screen
	for {
		select {
		case s := <-runner.Snapshots():
			v.snap = &s
			v.record(&s)
			v.draw()
		case ev := <-keys:
			switch ev := ev.(type) {
			case *tcell.EventResize:
				screen.Sync()
				v.draw()
			case *tcell.EventKey:
				if !v.handleKey(runner, ev) {
					return
				}
			case *tcell.EventMouse:
				v.handleMouse(ev)
			}
		}
	}
}

// column is one bar of the activity graph: the share of the population in
// each kind of work, averaged over graphTicks ticks. The shares need not
// reach 1; what is left is everyone with nothing planned.
type column [len(ascii.Groups)]float64

type view struct {
	screen tcell.Screen
	snap   *observe.Snapshot
	// hist is the graph's history, oldest column first. acc gathers the
	// ticks of the column still being filled.
	hist     []column
	acc      column
	accTotal float64
	accTicks int
	speed    float64
	paused   bool

	// sel is the agent the panel is opened up on, zero for nobody, and pic
	// is the last portrait drawn of it. look and follow reach into the
	// simulation; both are nil in tests, which draw from a snapshot alone.
	sel    entity.ID
	pic    *observe.Portrait
	look   func(entity.ID) *observe.Portrait
	follow func(entity.ID)

	// vitals swaps the map for the settlement's demographic record and
	// world for what became of the ground it stands on. The history behind
	// both is kept whether either page is open or not: a settlement dies
	// out once, and nobody is watching the right page when it does. See
	// vitals.go and world.go.
	vitals bool
	world  bool
	// focus is the graph on the page now up that is opened out, counted
	// from one, and zero for none. See focus.go.
	focus     int
	traces    []trace
	techs     []found
	peak      int
	peakAt    int
	gone      int // the tick the last person died, zero while anyone lives
	births    int
	deaths    int
	lastBirth int
	lastDeath int
}

// record folds a tick's activity into the history. Ticks are averaged into
// columns because a single tick's answer to what everyone is doing changes
// faster than it can be read: agents swap between farm and forage and eat
// several times a second. A column holds still.
func (v *view) record(s *observe.Snapshot) {
	v.keep(s)
	for _, a := range s.Activity {
		v.acc[ascii.GroupOf(a.Action)] += float64(a.Agents)
	}
	v.accTotal += float64(s.Population)
	v.accTicks++
	if v.accTicks < graphTicks {
		return
	}
	var col column
	if v.accTotal > 0 {
		for i, n := range v.acc {
			col[i] = n / v.accTotal
		}
	}
	v.hist = append(v.hist, col)
	if len(v.hist) > graphMax {
		v.hist = v.hist[len(v.hist)-graphMax:]
	}
	v.acc, v.accTotal, v.accTicks = column{}, 0, 0
	// The run's own history is kept last, so that the column of work just
	// closed is the one it takes.
	v.plot(s)
}

// handleKey reacts to a key press; it returns false when the user quits.
func (v *view) handleKey(r *sim.Runner, ev *tcell.EventKey) bool {
	switch {
	case ev.Key() == tcell.KeyCtrlC || ev.Rune() == 'q':
		return false
	case ev.Key() == tcell.KeyEscape:
		// Esc backs out one step at a time — off whichever graph is opened
		// out, then off whichever page is up, then off whoever is being
		// followed — and only quits when there is nothing left to back out
		// of: dropping back to the settlement is the commoner move.
		switch {
		case v.focus != 0:
			v.focus = 0
		case v.vitals || v.world:
			v.vitals, v.world, v.focus = false, false, 0
		case v.sel != 0:
			v.choose(0)
		default:
			return false
		}
	case ev.Rune() == ' ':
		v.paused = !v.paused
		if v.paused {
			r.Pause()
		} else {
			r.Resume()
		}
	case ev.Rune() == '+' || ev.Rune() == '=':
		v.speed *= 2
		r.SetSpeed(v.speed)
	case ev.Rune() == '-':
		v.speed /= 2
		if v.speed < 0.25 {
			v.speed = 0.25
		}
		r.SetSpeed(v.speed)
	case ev.Rune() == '.':
		r.StepOnce()
	case ev.Key() == tcell.KeyTab:
		v.pick(1)
	case ev.Key() == tcell.KeyBacktab:
		v.pick(-1)
	case ev.Key() == tcell.KeyUp:
		v.step(-1)
	case ev.Key() == tcell.KeyDown:
		v.step(1)
	case ev.Rune() == 'd':
		// Every page has its own graphs to step through, so changing page
		// closes whatever was opened out on the one being left.
		v.vitals, v.world, v.focus = !v.vitals, false, 0
	case ev.Rune() == 'w':
		v.world, v.vitals, v.focus = !v.world, false, 0
	case ev.Rune() == 'r':
		// Lay the whole street network at once. Agents pave for themselves
		// now, a length at a time where they have worn the ground; this is
		// the operator's shortcut, for seeing what a finished network does to
		// a settlement without waiting for one to be built.
		r.Send(sim.Func(func(w *world.World) { w.PaveStreets() }))
	}
	v.draw()
	return true
}

// handleMouse picks the figure under a click. It is the shortest way from
// "what is that one doing over there" to an answer.
func (v *view) handleMouse(ev *tcell.EventMouse) {
	if ev.Buttons()&tcell.Button1 == 0 || v.snap == nil {
		return
	}
	x, y := ev.Position()
	for _, m := range v.snap.Map.Agents {
		if m.Pos.X == x && m.Pos.Y == y {
			v.choose(m.ID)
			v.draw()
			return
		}
	}
}

// pick moves the selection step places along the population, in the order
// agents were born, and starts from the beginning when nobody is selected.
func (v *view) pick(step int) {
	if v.snap == nil || len(v.snap.Map.Agents) == 0 {
		return
	}
	marks := v.snap.Map.Agents
	at := -1
	for i, m := range marks {
		if m.ID == v.sel {
			at = i
			break
		}
	}
	if at < 0 {
		// Whoever was being followed is gone, or nobody was: take the first
		// going forward and the last going back.
		if step > 0 {
			v.choose(marks[0].ID)
		} else {
			v.choose(marks[len(marks)-1].ID)
		}
		return
	}
	next := (at + step + len(marks)) % len(marks)
	v.choose(marks[next].ID)
}

// choose follows an agent, telling the simulation to keep that one's
// deliberations. The thinking of whoever was followed before is forgotten:
// it is a window, not a record.
func (v *view) choose(id entity.ID) {
	v.sel, v.pic = id, nil
	if v.follow != nil {
		v.follow(id)
	}
}

var palette = map[ascii.Color]tcell.Style{
	ascii.Default:     tcell.StyleDefault,
	ascii.Water:       tcell.StyleDefault.Foreground(tcell.ColorBlue),
	ascii.Field:       tcell.StyleDefault.Foreground(tcell.ColorYellow),
	ascii.FieldFenced: tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true),
	ascii.House:       tcell.StyleDefault.Foreground(tcell.ColorWhite).Bold(true),
	ascii.Market:      tcell.StyleDefault.Foreground(tcell.ColorFuchsia).Bold(true),
	ascii.Road:        tcell.StyleDefault.Foreground(tcell.Color137),
	ascii.Rock:        tcell.StyleDefault.Foreground(tcell.ColorGray),
	ascii.RockHigh:    tcell.StyleDefault.Foreground(tcell.Color(250)),
	ascii.Granary:     tcell.StyleDefault.Foreground(tcell.ColorOrange).Bold(true),
	ascii.Tavern:      tcell.StyleDefault.Foreground(tcell.ColorFuchsia).Bold(true),
	ascii.AgentFood:   tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true),
	ascii.AgentBuild:  tcell.StyleDefault.Foreground(tcell.ColorOrange).Bold(true),
	ascii.AgentTrade:  tcell.StyleDefault.Foreground(tcell.ColorLime).Bold(true),
	ascii.AgentGuard:  tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true),
	ascii.AgentSocial: tcell.StyleDefault.Foreground(tcell.ColorFuchsia).Bold(true),
	ascii.AgentStudy:  tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true),
	ascii.AgentIdle:   tcell.StyleDefault.Foreground(tcell.ColorWhite).Bold(true),

	// The two ramps: open country and woodland, valley floor to skyline. The
	// land is what most of the screen is, so these are most of what the map
	// looks like, and they are the whole of how high ground reads as high.
	//
	// Both run from a damp green through a dry khaki to the bare grey of a
	// mountainside, which is what ground does as it rises: the colour is
	// carrying height, so it has to leave green behind or the top of the map
	// looks like a meadow. The woods keep more green than the open ground
	// for longer, because a wood is green - a hillside of trees should read
	// as a wooded hillside and not as another shade of rock.
	ascii.Ground0: tcell.StyleDefault.Foreground(tcell.Color(22)),  // the water meadow
	ascii.Ground1: tcell.StyleDefault.Foreground(tcell.Color(65)),  // the valley floor
	ascii.Ground2: tcell.StyleDefault.Foreground(tcell.Color(101)), // the shoulder of it
	ascii.Ground3: tcell.StyleDefault.Foreground(tcell.Color(138)), // the foothills
	ascii.Ground4: tcell.StyleDefault.Foreground(tcell.Color(145)), // the mountainside
	ascii.Ground5: tcell.StyleDefault.Foreground(tcell.Color(252)), // the tops

	ascii.Wood0: tcell.StyleDefault.Foreground(tcell.Color(22)),
	ascii.Wood1: tcell.StyleDefault.Foreground(tcell.Color(28)),
	ascii.Wood2: tcell.StyleDefault.Foreground(tcell.Color(34)),
	ascii.Wood3: tcell.StyleDefault.Foreground(tcell.Color(71)),
	ascii.Wood4: tcell.StyleDefault.Foreground(tcell.Color(108)),
	ascii.Wood5: tcell.StyleDefault.Foreground(tcell.Color(144)),
}

func (v *view) draw() {
	sc := v.screen
	sc.Clear()
	if v.snap == nil {
		puts(sc, 0, 0, tcell.StyleDefault, "waiting for first tick...")
		sc.Show()
		return
	}
	s := v.snap
	sw, sh := sc.Size()
	if v.vitals {
		v.drawVitals(sw, sh)
		return
	}
	if v.world {
		v.drawWorld(sw, sh)
		return
	}
	if v.sel != 0 && v.look != nil {
		v.pic = v.look(v.sel)
	}
	needH := s.Map.H + graphHeight + 3 // map, the legend, the graph, its span, the keys
	if sw < s.Map.W+panelWidth || sh < needH {
		puts(sc, 0, 0, tcell.StyleDefault, fmt.Sprintf("terminal too small: need %dx%d, have %dx%d", s.Map.W+panelWidth, needH, sw, sh))
		sc.Show()
		return
	}

	for y, row := range ascii.Render(s.Map) {
		for x, c := range row {
			sc.SetContent(x, y, c.Ch, nil, palette[c.Color])
		}
	}
	// The figure being followed is turned inside out on the map, so that the
	// panel beside it is plainly about somebody in particular and the eye
	// can find them again in the crowd after they have walked.
	for _, m := range s.Map.Agents {
		if m.ID == v.sel {
			sc.SetContent(m.Pos.X, m.Pos.Y, '@', nil, palette[ascii.AgentColor(m.Action)].Reverse(true))
			break
		}
	}

	px := s.Map.W + 2
	line := 0
	put := func(style tcell.Style, format string, args ...any) {
		puts(sc, px, line, style, fmt.Sprintf(format, args...))
		line++
	}
	bold := tcell.StyleDefault.Bold(true)
	dim := tcell.StyleDefault.Dim(true)

	state := fmt.Sprintf("%.0f t/s", v.speed)
	if v.paused {
		state = "PAUSED"
	}
	// The date rather than the tick count: what a settlement is living
	// through is a year and a season, and the raw day is only useful for
	// lining a run up against a log.
	put(bold, "%-20s pop %-5d %s", s.Date, s.Population, state)
	put(dim, "day %d", s.Tick)
	line++
	for _, t := range need.Tiers() {
		put(tcell.StyleDefault, "%-13s %s %.2f", t, bar(s.MeanNeeds[t], 12), s.MeanNeeds[t])
	}
	// Health sits under the needs it is made of, set apart by colour: it is
	// not something anyone wants, it is what living on those needs has done
	// to the population's bodies.
	put(healthStyle(s.MeanHealth), "%-13s %s %.2f", "health", bar(s.MeanHealth, 12), s.MeanHealth)
	line++
	// The settlement's figures are laid out as a grid rather than a run of
	// sentences: dim label on the left, number right-aligned against a
	// fixed column, two to a line, and the lines in the order of what they
	// are about — the people, then their land, then their dealings, then
	// what is getting through to them. Read this way the eye runs down a
	// column of numbers instead of picking each one out of the middle of a
	// phrase, and a figure that has moved since the last glance is still in
	// the place it was before.
	stat := func(l1, v1, l2, v2 string) {
		puts(sc, px, line, dim, l1)
		puts(sc, px+statLabel, line, tcell.StyleDefault, fmt.Sprintf("%*s", statValue, v1))
		if l2 != "" {
			puts(sc, px+statCol, line, dim, l2)
			puts(sc, px+statCol+statLabel, line, tcell.StyleDefault, fmt.Sprintf("%*s", statValue, v2))
		}
		line++
	}
	n := func(v int) string { return fmt.Sprintf("%d", v) }
	f := func(v float64) string { return fmt.Sprintf("%.2f", v) }

	stat("mean age", n(s.MeanAge), "elders", n(s.Elders))
	stat("friends", n(s.Friendships), "feuds", n(s.Feuds))
	stat("hearsay", n(s.Hearsay), "", "")
	stat("houses", n(s.Houses), "fields", n(s.Fields))
	stat("forest", n(s.Forest), "roads", n(s.Roads))
	stat("food price", f(s.FoodPrice), "safety", f(s.Safety))
	stat("knowledge", fmt.Sprintf("%.0f", s.Knowledge), "gini", f(s.WealthGini))
	stat("reach", f(s.GatedReach), "spread", f(s.HabitSpread))
	stat("open", f(s.ChoiceEntropy), "", "")
	// The weather keeps a line of its own, out of the grid: it is the one
	// thing on the panel that moves on its own schedule rather than the
	// settlement's, and the growth figure says what the season is doing to
	// the land.
	put(tcell.StyleDefault, "%-7s %+5.1f deg  growth %.2f", s.Date.Season, s.Temp, s.Growth)
	techs := "none yet"
	if len(s.Techs) > 0 {
		parts := make([]string, len(s.Techs))
		for i, t := range s.Techs {
			parts[i] = string(t)
		}
		techs = strings.Join(parts, ", ")
	}
	put(tcell.StyleDefault, "techs: %s", trim(techs, panelWidth-9))
	line++
	if v.sel != 0 {
		v.drawCard(px, &line, sh-3)
	}
	// What everyone is doing goes under the map rather than in the panel:
	// given the map's whole width it is hundreds of ticks of history at
	// once, and the eye reads a weave of bands widening and giving way far
	// better than it reads a column of numbers.
	//
	// The legend is fixed: every kind of work always in the same place in
	// the same colour, whether anyone is doing it or not. A legend that
	// reshuffles itself is one more thing moving on a view meant to be read
	// at a glance, and the colours have to mean the same thing from one
	// frame to the next for the weave under them to be legible at all.
	var counts [len(ascii.Groups)]int
	for _, a := range s.Activity {
		counts[ascii.GroupOf(a.Action)] += a.Agents
	}
	cell := s.Map.W / len(ascii.Groups)
	for i, g := range ascii.Groups {
		lx := i * cell
		puts(sc, lx, s.Map.H, palette[g.Color], "█")
		// A kind of work nobody is doing is dim, and so is one that is
		// simply not the one being read: while a band is opened out the
		// legend says which of them it is. See focus.go.
		style := v.markGroup(i)
		if counts[i] == 0 && v.focus == 0 {
			style = dim
		}
		puts(sc, lx+2, s.Map.H, style, trim(fmt.Sprintf("%s %d", g.Name, counts[i]), cell-3))
	}
	// The band under the map is either the whole weave or the one kind of
	// work stepped onto, drawn in the same rows and scaled to its own high
	// so that a thin one can be read at all.
	span := fmt.Sprintf("%d days →", min(len(v.hist), s.Map.W)*graphTicks)
	if g, ok := v.focused(); ok {
		v.drawOpen(0, s.Map.H+1, s.Map.W, graphHeight, g)
		span = trim(v.headline(g)+"   "+span, s.Map.W)
	} else {
		v.drawGraph(0, s.Map.H+1, s.Map.W, graphHeight)
	}
	puts(sc, 0, s.Map.H+1+graphHeight, dim, span)
	puts(sc, px, sh-2, dim, "space pause  +/- speed  . step  r pave")
	puts(sc, px, sh-1, dim, trim(v.keyed("tab pick  d vitals  w world  q quit"), panelWidth))
	// A settlement that has ended says so across the empty map it left, and
	// says where to go and read why. Without this the map simply stops
	// moving and a finished run looks like a hung one.
	if s.Population == 0 && v.gone != 0 {
		note := fmt.Sprintf("the settlement died out at tick %d — press d for why", v.gone)
		puts(sc, max(0, (s.Map.W-len([]rune(note)))/2), s.Map.H/2,
			tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true), note)
	}
	sc.Show()
}

// drawGraph weaves the kinds of work into a w-wide band of history, oldest
// column on the left, the full height being the whole population and the gap
// at the top those with nothing planned. What a single-tick list could never
// show is here: whether a band is widening.
func (v *view) drawGraph(x, y, w, h int) {
	hist := v.hist
	if len(hist) > w {
		hist = hist[len(hist)-w:] // only what fits, the recent past
	}
	for i, col := range hist {
		cx := x + w - len(hist) + i
		for r := 0; r < h; r++ {
			share := (float64(h-r) - 0.5) / float64(h)
			var cum float64
			for gi, g := range ascii.Groups {
				cum += col[gi]
				if share <= cum {
					v.screen.SetContent(cx, y+r, '█', nil, palette[g.Color])
					break
				}
			}
		}
	}
}

// puts writes a line of text from x, one cell to the rune. Ranging over the
// string itself would count in bytes, which is the same thing only while the
// text is ASCII: the moment a line carries a rule, an arrow or a degree sign
// it tears open, every glyph after the first multi-byte one pushed two cells
// further right than it belongs and the tail of the line run off the width
// it was trimmed to. Every header on every page is drawn through here.
func puts(sc tcell.Screen, x, y int, style tcell.Style, text string) {
	for i, r := range []rune(text) {
		sc.SetContent(x+i, y, r, nil, style)
	}
}

// healthStyle colours the health bar by how worn the population is, so a
// settlement grinding its people down reads at a glance instead of only in
// the death feed a few hundred ticks later.
func healthStyle(h float64) tcell.Style {
	switch {
	case h < 0.5:
		return tcell.StyleDefault.Foreground(tcell.ColorRed)
	case h < 0.75:
		return tcell.StyleDefault.Foreground(tcell.ColorYellow)
	}
	return tcell.StyleDefault.Foreground(tcell.ColorLime)
}

func bar(v float64, width int) string {
	n := int(v*float64(width) + 0.5)
	if n > width {
		n = width
	}
	return strings.Repeat("█", n) + strings.Repeat("░", width-n)
}

// trim cuts a string to at most max columns, counted in runes: the panel is
// full of bars and arrows, and cutting those by the byte both truncates far
// too early and can cut a glyph in half.
func trim(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}

// cardTail is how many lines the end of the card wants: the errand, the
// heading over the thinking, the candidates under it, and the run of
// choices behind them.
const cardTail = 9

// drawCard opens up the followed agent under the settlement's own figures,
// down to at most the bottom line. What is drawn first matters: on a short
// terminal the card is simply cut off, so it runs from who this is, through
// what it is doing, to what it thought about doing instead.
func (v *view) drawCard(x int, line *int, bottom int) {
	sc := v.screen
	bold := tcell.StyleDefault.Bold(true)
	dim := tcell.StyleDefault.Dim(true)
	put := func(style tcell.Style, format string, args ...any) {
		if *line > bottom {
			return
		}
		puts(sc, x, *line, style, trim(fmt.Sprintf(format, args...), panelWidth))
		*line++
	}

	p := v.pic
	if p == nil {
		put(dim, "─ #%d ─ gone", v.sel)
		return
	}
	put(bold, "─ %s ─ #%d ─ age %d", p.Name, p.ID, p.Age)
	// What an agent is good at is the nearest thing this world has to a
	// role: nobody is given one, and one is arrived at all the same. What it
	// believes it is good at is shown beside it, because that, and not the
	// truth, is what it acts on.
	put(tcell.StyleDefault, "%-11s %.2f  believes %.2f", p.Calling, p.Level, p.Efficacy[p.Calling])
	put(healthStyle(p.Health), "%-13s %s %.2f", "health", bar(p.Health, 8), p.Health)
	// Beside each need is how hard it is pulling: the level is what the
	// agent has, the arrow is what that lack is worth to this particular
	// agent, which is the whole of what it decides on.
	for _, t := range need.Tiers() {
		put(tcell.StyleDefault, "%-13s %s %.2f ▲%.2f", t, bar(p.Needs[t], 8), p.Needs[t], p.Urgency[t]*p.Personality[t])
	}
	put(dim, "shelter %.2f  wealth %.1f  rep %.1f", p.Shelter, p.Wealth, p.Reputation)
	// Everything from here to the errand is worth knowing and none of it is
	// worth crowding out the thinking, which is the point of the card. On a
	// short terminal it gives way instead of pushing that off the bottom.
	spare := bottom - *line - cardTail
	if carrying := goods(p); carrying != "" && spare > 0 {
		put(dim, "carrying %s", carrying)
		spare--
	}
	if spare > 0 {
		put(dim, "knows %d  friends %d  feuds %d  said %d", p.Known, p.Friends, p.Feuds, p.Hearsay)
		spare--
	}
	for _, t := range p.Ties {
		if spare <= 0 {
			break
		}
		spare--
		if t.Met == 0 {
			put(dim, "  %-8s heard of, regard %+.2f", trim(t.Name, 8), t.Regard)
			continue
		}
		put(dim, "  %-8s bond %.2f regard %+.2f", trim(t.Name, 8), t.Strength, t.Regard)
	}
	*line++
	put(bold, "%s", errand(p))
	if len(p.Thinking) == 0 {
		put(dim, "thinking: waiting for its next choice")
		return
	}
	last := p.Thinking[len(p.Thinking)-1]
	put(bold, "weighed on %s  open %.2f", clock.At(last.Tick), last.Entropy)
	for i, c := range last.Weighed {
		if i >= 5 && !c.Chosen {
			continue
		}
		style := palette[ascii.AgentColor(c.Action)]
		mark := " "
		if c.Chosen {
			mark = "▸"
			style = style.Bold(true)
		}
		// An action out of reach says so. It is the one way a candidate can
		// stand at the top of the list and still be passed over, and
		// without the number it looks like the draw misbehaving.
		var reach string
		if c.Reach < 1 {
			reach = fmt.Sprintf("  reach %.2f", c.Reach)
		}
		put(style, "%s %-11s %+.2f %3.0f%%%s", mark, trim(c.Action, 11), c.Weight, c.Chance*100, reach)
	}
	put(dim, "since  %s", recent(p.Thinking))
}

// errand puts an agent's plan into a line: an agent is either on its way
// somewhere, at work, or between the two.
func errand(p *observe.Portrait) string {
	if p.Errand == nil {
		return "deciding what to do"
	}
	e := p.Errand
	if e.Walking {
		return fmt.Sprintf("%s: %d steps to (%d,%d)", e.Action, e.Steps, e.Target.X, e.Target.Y)
	}
	return fmt.Sprintf("%s: %d of %d ticks left", e.Action, e.Remaining, e.Total)
}

// goods lists what an agent has on it, leaving out what it has none of.
func goods(p *observe.Portrait) string {
	var parts []string
	for g := entity.Good(0); g < entity.GoodCount; g++ {
		if p.Inventory[g] >= 0.05 {
			parts = append(parts, fmt.Sprintf("%s %.1f", g, p.Inventory[g]))
		}
	}
	return strings.Join(parts, " ")
}

// recent is the run of choices behind the latest one, newest first. One
// decision says what an agent did; a run of them says whether it is getting
// anywhere, or turning on the spot between the same two errands.
func recent(ds []world.Deliberation) string {
	var parts []string
	for i := len(ds) - 2; i >= 0 && len(parts) < 6; i-- {
		parts = append(parts, ds[i].Chose())
	}
	if len(parts) == 0 {
		return "nothing yet"
	}
	return strings.Join(parts, " ")
}
