// Command watch is a live terminal view of a settlement developing, in the
// spirit of Dwarf Fortress. The map is on the left, aggregate state and a
// feed of notable events on the right. It is an omniscient view for insight
// while tuning; the player's own view will be far narrower.
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

	"lreat/core/event"
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
	feedLength = 14
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
			for _, e := range s.Notable {
				if e.Kind == event.Traded || e.Kind == event.Guarded {
					continue // too frequent to be informative in the feed
				}
				v.feed = append(v.feed, e)
			}
			if len(v.feed) > feedLength {
				v.feed = v.feed[len(v.feed)-feedLength:]
			}
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

type view struct {
	screen tcell.Screen
	snap   *observe.Snapshot
	feed   []event.Event
	speed  float64
	paused bool
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
		// Lay the streets by hand. Roads are a material the settlement can
		// have; deciding to want one is not yet anybody's to make, so for now
		// the observer spawns them and watches what changes.
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
	if sw < s.Map.W+panelWidth || sh < s.Map.H {
		puts(sc, 0, 0, tcell.StyleDefault, fmt.Sprintf("terminal too small: need %dx%d, have %dx%d", s.Map.W+panelWidth, s.Map.H, sw, sh))
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
	put(tcell.StyleDefault, "knowledge %.0f  gini %.2f", s.Knowledge, s.WealthGini)
	put(tcell.StyleDefault, "friends %-4d feuds %-4d hearsay %d", s.Friendships, s.Feuds, s.Hearsay)
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
	put(bold, "doing")
	for i, a := range s.Activity {
		if i == 6 {
			break
		}
		puts(sc, px, line, palette[ascii.AgentColor(a.Action)], "@")
		puts(sc, px+2, line, tcell.StyleDefault, fmt.Sprintf("%-14s %d", a.Action, a.Agents))
		line++
	}
	line++
	put(bold, "recently")
	for _, e := range v.feed {
		if line >= sh-1 {
			break
		}
		style := tcell.StyleDefault
		switch e.Kind {
		case event.Discovered:
			style = bold.Foreground(tcell.ColorAqua)
		case event.Died, event.Stolen, event.Avenged:
			style = tcell.StyleDefault.Foreground(tcell.ColorRed)
		case event.Born:
			style = tcell.StyleDefault.Foreground(tcell.ColorLime)
		case event.Met, event.Taught:
			style = dim
		}
		put(style, "%s", trim(fmt.Sprintf("%6d %s", e.Tick, e.Text), panelWidth-2))
	}
	puts(sc, px, sh-1, dim, "space pause  +/- speed  . step  r pave  q quit")
	sc.Show()
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
