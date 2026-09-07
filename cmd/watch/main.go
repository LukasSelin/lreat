// Command watch is a live terminal view of a settlement developing, in the
// spirit of Dwarf Fortress. The map is on the left with aggregate state
// beside it, and under the map a woven band of what the settlement has been
// spending itself on, running the map's whole width. It is an omniscient
// view for insight while tuning; the player's own view will be far narrower.
//
// Keys: space pauses, + and - change speed, . steps once while paused,
// r lays streets through the settlement, q quits.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"

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

	keys := make(chan tcell.Event, 16)
	go func() {
		for {
			keys <- screen.PollEvent()
		}
	}()

	v := &view{screen: screen, speed: *tps}
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
	case ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC || ev.Rune() == 'q':
		return false
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

var palette = map[ascii.Color]tcell.Style{
	ascii.Default:     tcell.StyleDefault,
	ascii.Water:       tcell.StyleDefault.Foreground(tcell.ColorBlue),
	ascii.Grass:       tcell.StyleDefault.Foreground(tcell.Color(22)),
	ascii.ForestRich:  tcell.StyleDefault.Foreground(tcell.ColorGreen),
	ascii.ForestPoor:  tcell.StyleDefault.Foreground(tcell.ColorOlive),
	ascii.Field:       tcell.StyleDefault.Foreground(tcell.ColorYellow),
	ascii.House:       tcell.StyleDefault.Foreground(tcell.ColorWhite).Bold(true),
	ascii.Market:      tcell.StyleDefault.Foreground(tcell.ColorFuchsia).Bold(true),
	ascii.Road:        tcell.StyleDefault.Foreground(tcell.Color137),
	ascii.Rock:        tcell.StyleDefault.Foreground(tcell.ColorGray),
	ascii.Granary:     tcell.StyleDefault.Foreground(tcell.ColorOrange).Bold(true),
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
	// The weather gets a line of its own: it is the one thing on the panel
	// that moves on its own schedule rather than the settlement's, and the
	// growth figure says what the season is doing to the land.
	put(tcell.StyleDefault, "%-7s %+5.1f deg  growth %.2f", s.Season, s.Temp, s.Growth)
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
	puts(sc, px, sh-1, dim, "space pause  +/- speed  . step  r pave  q quit")
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

func trim(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
