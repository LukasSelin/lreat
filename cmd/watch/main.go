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
// Keys: space pauses, + and - change speed, . steps once while paused,
// r lays streets through the settlement, tab and shift-tab pick an agent
// (or click one), esc drops it, q quits.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"

	"lreat/core/entity"
	"lreat/core/need"
	"lreat/core/observe"
	"lreat/core/sim"
	"lreat/core/world"
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
	// column is the mean of graphTicks ticks; graphMax columns are kept so
	// a wider map simply shows more of the same history.
	graphHeight = 6
	graphTicks  = 5
	graphMax    = 320
)

func main() {
	seed := flag.Uint64("seed", 1, "world seed")
	agents := flag.Int("agents", 20, "starting population")
	tps := flag.Float64("tps", 20, "initial ticks per second")
	width := flag.Int("width", world.DefaultWidth, "map width")
	height := flag.Int("height", world.DefaultHeight, "map height")
	flag.Parse()

	w := world.NewSized(*seed, *width, *height)
	for i := 0; i < *agents; i++ {
		w.Spawn(fmt.Sprintf("%s%d", names[i%len(names)], i/len(names)), w.RandomPersonality())
	}
	runner := sim.New(w, *tps)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go runner.Run(ctx)

	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err := screen.Init(); err != nil {
		log.Fatal(err)
	}
	defer screen.Fini()
	screen.EnableMouse(tcell.MouseButtonEvents) // clicking a figure picks it

	keys := make(chan tcell.Event, 16)
	go func() {
		for {
			keys <- screen.PollEvent()
		}
	}()

	v := &view{
		screen: screen,
		speed:  *tps,
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
}

// record folds a tick's activity into the history. Ticks are averaged into
// columns because a single tick's answer to what everyone is doing changes
// faster than it can be read: agents swap between farm and forage and eat
// several times a second. A column holds still.
func (v *view) record(s *observe.Snapshot) {
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
}

// handleKey reacts to a key press; it returns false when the user quits.
func (v *view) handleKey(r *sim.Runner, ev *tcell.EventKey) bool {
	switch {
	case ev.Key() == tcell.KeyCtrlC || ev.Rune() == 'q':
		return false
	case ev.Key() == tcell.KeyEscape:
		// Esc lets go of whoever is being followed, and only quits when
		// nobody is: dropping back to the settlement is the commoner move.
		if v.sel == 0 {
			return false
		}
		v.choose(0)
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
	ascii.Grass:       tcell.StyleDefault.Foreground(tcell.Color(22)),
	ascii.GrassLow:    tcell.StyleDefault.Foreground(tcell.Color(28)),
	ascii.GrassHigh:   tcell.StyleDefault.Foreground(tcell.Color(101)),
	ascii.ForestRich:  tcell.StyleDefault.Foreground(tcell.ColorGreen),
	ascii.ForestPoor:  tcell.StyleDefault.Foreground(tcell.ColorOlive),
	ascii.Field:       tcell.StyleDefault.Foreground(tcell.ColorYellow),
	ascii.House:       tcell.StyleDefault.Foreground(tcell.ColorWhite).Bold(true),
	ascii.Market:      tcell.StyleDefault.Foreground(tcell.ColorFuchsia).Bold(true),
	ascii.Road:        tcell.StyleDefault.Foreground(tcell.Color137),
	ascii.Rock:        tcell.StyleDefault.Foreground(tcell.ColorGray),
	ascii.Granary:     tcell.StyleDefault.Foreground(tcell.ColorOrange).Bold(true),
	ascii.Tavern:      tcell.StyleDefault.Foreground(tcell.ColorFuchsia).Bold(true),
	ascii.AgentFood:   tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true),
	ascii.AgentBuild:  tcell.StyleDefault.Foreground(tcell.ColorOrange).Bold(true),
	ascii.AgentTrade:  tcell.StyleDefault.Foreground(tcell.ColorLime).Bold(true),
	ascii.AgentGuard:  tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true),
	ascii.AgentSocial: tcell.StyleDefault.Foreground(tcell.ColorFuchsia).Bold(true),
	ascii.AgentStudy:  tcell.StyleDefault.Foreground(tcell.ColorAqua).Bold(true),
	ascii.AgentIdle:   tcell.StyleDefault.Foreground(tcell.ColorWhite).Bold(true),
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
	if v.sel != 0 && v.look != nil {
		v.pic = v.look(v.sel)
	}
	sw, sh := sc.Size()
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
	put(bold, "tick %-7d pop %-5d %s", s.Tick, s.Population, state)
	line++
	for _, t := range need.Tiers() {
		put(tcell.StyleDefault, "%-13s %s %.2f", t, bar(s.MeanNeeds[t], 12), s.MeanNeeds[t])
	}
	// Health sits under the needs it is made of, set apart by colour: it is
	// not something anyone wants, it is what living on those needs has done
	// to the population's bodies.
	put(healthStyle(s.MeanHealth), "%-13s %s %.2f", "health", bar(s.MeanHealth, 12), s.MeanHealth)
	line++
	put(tcell.StyleDefault, "mean age %-5d elders %d", s.MeanAge, s.Elders)
	put(tcell.StyleDefault, "houses %-4d fields %-4d forest %d", s.Houses, s.Fields, s.Forest)
	put(tcell.StyleDefault, "roads  %-4d", s.Roads)
	put(tcell.StyleDefault, "safety %.2f  food price %.2f", s.Safety, s.FoodPrice)
	put(tcell.StyleDefault, "knowledge %.0f  gini %.2f", s.Knowledge, s.WealthGini)
	put(tcell.StyleDefault, "friends %-4d feuds %-4d hearsay %d", s.Friendships, s.Feuds, s.Hearsay)
	put(tcell.StyleDefault, "reach %.2f  spread %.2f  open %.2f", s.GatedReach, s.HabitSpread, s.ChoiceEntropy)
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
		style := tcell.StyleDefault
		if counts[i] == 0 {
			style = dim
		}
		puts(sc, lx+2, s.Map.H, style, trim(fmt.Sprintf("%s %d", g.Name, counts[i]), cell-3))
	}
	v.drawGraph(0, s.Map.H+1, s.Map.W)
	puts(sc, 0, s.Map.H+1+graphHeight, dim, fmt.Sprintf("%d ticks →", min(len(v.hist), s.Map.W)*graphTicks))
	puts(sc, px, sh-2, dim, "space pause  +/- speed  . step  r pave")
	puts(sc, px, sh-1, dim, "tab/click pick  esc drop  q quit")
	sc.Show()
}

// drawGraph weaves the kinds of work into a w-wide band of history, oldest
// column on the left, the full height being the whole population and the gap
// at the top those with nothing planned. What a single-tick list could never
// show is here: whether a band is widening.
func (v *view) drawGraph(x, y, w int) {
	hist := v.hist
	if len(hist) > w {
		hist = hist[len(hist)-w:] // only what fits, the recent past
	}
	for i, col := range hist {
		cx := x + w - len(hist) + i
		for r := 0; r < graphHeight; r++ {
			share := (float64(graphHeight-r) - 0.5) / float64(graphHeight)
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

func puts(sc tcell.Screen, x, y int, style tcell.Style, text string) {
	for i, r := range text {
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
	put(bold, "weighed at tick %d  open %.2f", last.Tick, last.Entropy)
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
