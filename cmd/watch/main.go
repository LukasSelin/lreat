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
// A map the terminal cannot hold whole — a globe is sixteen chunks round and
// no terminal is — is looked at through a window that moves over it, and the
// arrows move it. Z draws it at twice as much ground to the cell and Z back
// in again, up to the scale that holds the whole world: panning answers
// where else to look and only zooming answers what shape the place is.
// Everything else here is written as though the map were the screen because
// for a valley it still is: the window is the whole map at the only scale it
// has, and the keys that move it have nowhere to go. See camera.go.
//
// The arrows are given to whichever of the two the page has. Over a map with
// somewhere to look they look, and the graphs answer to pgup and pgdn; on a
// valley drawn whole, and on both pages of graphs, they step the graphs as
// they always did, and pgup and pgdn do the same. The line at the foot of
// the panel says which case it is in.
//
// Keys: space pauses, + and - change speed, . steps once while paused,
// m turns the map to the next reading of the land and M to the last,
// the arrows (or hjkl) look around a map larger than the screen, z and Z
// draw it at more and less ground to the cell (as does the wheel), c comes
// back to the settlement and f keeps the view on whoever is being followed,
// r lays streets through the settlement, tab and shift-tab pick an agent
// (or click one), pgup and pgdn open a graph out, esc backs off the graph
// and then the page, d shows the vitals, w the world, q quits.
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
	"lreat/core/system"
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
	ceiling := flag.Int("cap", d.ceiling, "how many people this program will carry: a guard on the machine, not a fact about the world (0 takes it off and lets the land do the stopping)")
	temp := flag.Float64("temp", d.temp, "base temperature of recognition; 0 always takes the best fit")
	snug := flag.Bool("fit", d.snug, "size the map to the terminal; -fit=false takes -width and -height instead")
	preset := flag.String("preset", d.preset, "which world: valley, a map with edges; ancient, that valley made out of its own history; or globe, a cylinder with no edges")
	skip := flag.Bool("start", false, "start straight away, without the menu")
	flag.Parse()
	if _, ok := world.Preset(*preset); !ok {
		fmt.Fprintf(os.Stderr, "no such preset: %q\n", *preset)
		os.Exit(2)
	}
	s := setup{
		seed: *seed, agents: *agents, tps: *tps,
		width: *width, height: *height, snug: *snug,
		fit: !*value, temp: *temp, preset: *preset,
		ceiling: *ceiling,
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
	// the map it was given rather than the map it would now ask for. A
	// globe is not measured against anything: it comes at the size that
	// makes it a globe, and the terminal shows as much of it as it can.
	sw, sh := screen.Size()
	s.measure(sw, sh)
	// Half a million tiles are raised, flooded, drained, incised and sorted
	// before there is anything to look at, and on a globe that is a couple
	// of seconds of a screen that has just been cleared. Saying what is
	// happening costs one line and is the difference between a wait and a
	// program that has hung.
	screen.Clear()
	puts(screen, 0, 0, tcell.StyleDefault, fmt.Sprintf("raising the %s: %d by %d...", s.world(), s.width, s.height))
	screen.Show()
	system.MaxPopulation = s.ceiling
	w := world.NewWith(s.seed, s.config())
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
		fmt.Fprintf(out, "\nfounded on seed %d in the %s, %d figures, %dx%d, temp %.2f, %s\n",
			s.seed, s.world(), s.agents, s.width, s.height, s.temp, choosing(s.fit))
		fmt.Fprint(out, v.Report())
		v.score(rep)
		fmt.Print(rep.Close())
	}()
	defer screen.Fini()
	screen.EnableMouse(tcell.MouseButtonEvents) // clicking a figure picks it

	keys := make(chan tcell.Event, 16)
	go func() {
		for {
			ev := screen.PollEvent()
			if ev == nil {
				return // the screen has been finished; there is nothing more to poll
			}
			keys <- ev
		}
	}()

	v.screen = screen
	// Every event draws, and drawing a large terminal is eleven
	// milliseconds. An arrow key held down, or a run of ticks arriving
	// together, is a burst of events, and drawing each of them meant a
	// frame apiece - every one of them but the last drawn for a state that
	// had already been left behind, while the presses queued up behind the
	// screen. So a frame is skipped while anything is already waiting to be
	// dealt with; the last event of a burst finds nothing waiting and draws
	// the once, from where the burst left off. See view.draw.
	v.busy = func() bool { return len(keys) > 0 || len(runner.Snapshots()) > 0 }
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
	// view is which reading of the land the map is drawn as. The settlement
	// is the one to watch a run on; the rest answer one question about the
	// ground over the whole map at once. See ui/ascii/view.go.
	view ascii.View
	// cam is which part of the map is on the screen. A map the terminal can
	// hold whole has only one answer and the camera never moves; a globe is
	// mostly off the screen at any moment and this is how the rest of it is
	// reached. See camera.go.
	cam camera

	sel    entity.ID
	pic    *observe.Portrait
	look   func(entity.ID) *observe.Portrait
	follow func(entity.ID)

	// busy says whether another event is already waiting, in which case
	// this frame would be replaced before it could be read and is not drawn
	// at all. Nil in tests, which draw whenever they say to.
	busy func() bool

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
	techs     []observe.Worked
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
		v.arrow(0, -1)
	case ev.Key() == tcell.KeyDown:
		v.arrow(0, 1)
	case ev.Key() == tcell.KeyLeft:
		v.arrow(-1, 0)
	case ev.Key() == tcell.KeyRight:
		v.arrow(1, 0)
	case ev.Key() == tcell.KeyPgUp:
		// The graphs answer to these wherever the arrows have gone, so that
		// stepping through them is one pair of keys and not a pair that
		// depends on how much of the map is on the screen.
		v.step(-1)
	case ev.Key() == tcell.KeyPgDn:
		v.step(1)
	case ev.Rune() == 'd':
		// Every page has its own graphs to step through, so changing page
		// closes whatever was opened out on the one being left.
		v.vitals, v.world, v.focus = !v.vitals, false, 0
	case ev.Rune() == 'w':
		v.world, v.vitals, v.focus = !v.world, false, 0
	case ev.Rune() == 'm':
		v.view = (v.view + 1) % ascii.View(len(ascii.Views))
	case ev.Rune() == 'M':
		v.view = (v.view + ascii.View(len(ascii.Views)) - 1) % ascii.View(len(ascii.Views))
	case ev.Rune() == 'h' || ev.Rune() == 'j' || ev.Rune() == 'k' || ev.Rune() == 'l':
		// Looking around, for a hand already on the letters. The arrows do
		// the same and are what the keys at the foot of the panel name; on a
		// map the screen already holds whole both do nothing, which is
		// right: there is nowhere else to look.
		dx, dy := 0, 0
		switch ev.Rune() {
		case 'h':
			dx = -1
		case 'l':
			dx = 1
		case 'k':
			dy = -1
		case 'j':
			dy = 1
		}
		v.pan(dx, dy)
	case ev.Rune() == 'z':
		// Out: twice as much ground to the cell. A globe is sixteen chunks
		// round and panning over it a third of a screen at a time never
		// shows anybody the shape of it.
		v.zoom(1)
	case ev.Rune() == 'Z':
		// And back in, to the tile-for-a-cell map the settlement is watched
		// on. The pair go the way m and M do: the small letter forward
		// through the scales, the capital back.
		v.zoom(-1)
	case ev.Rune() == 'c':
		// Back to the settlement. On a globe it is a fifth of a per cent of
		// the map and the only part of it anybody is watching; without a way
		// back, one pan too far east is a run abandoned.
		v.onMap(func(m *observe.MapView, w, h int) {
			v.cam.lock = false
			v.cam.center(m, m.Market, w, h)
		})
	case ev.Rune() == 'f':
		// Keep the window on whoever is being followed. It is the other half
		// of tab: naming a figure in the panel is no use on a map larger
		// than the screen if the figure itself is a hundred tiles away.
		v.cam.lock = !v.cam.lock
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

// arrow is one press of an arrow key: looking around where there is more map
// than screen, and stepping through the page's graphs where there is not.
// Left and right had nothing to do on a page of graphs and still have not.
func (v *view) arrow(dx, dy int) {
	if v.looking() {
		v.pan(dx, dy)
		return
	}
	if dy != 0 {
		v.step(dy)
	}
}

// pan moves the window a third of a screen.
func (v *view) pan(dx, dy int) {
	v.onMap(func(m *observe.MapView, w, h int) {
		v.cam.pan(m, dx*step(w), dy*step(h), w, h)
	})
}

// onMap runs a piece of camera work against the map as it is being shown:
// the ground in the latest snapshot, and how much of it the terminal is
// holding. Both are wanted by every key that moves the view and neither is
// kept anywhere, because the window is measured afresh each frame and the
// terminal may have been resized since the last one.
func (v *view) onMap(f func(m *observe.MapView, w, h int)) {
	if v.snap == nil || v.screen == nil {
		return
	}
	sw, sh := v.screen.Size()
	w, h := mapArea(v.snap.Map, sw, sh, v.cam.scale())
	if w < minMapW || h < minMapH {
		return
	}
	f(v.snap.Map, w, h)
}

// zoom draws the map at the next scale out or in, keeping the middle of the
// window where it is. What the eye is holding when the key goes down is
// whatever is in the middle of the screen, and a zoom that moves it is a
// zoom that has to be panned back afterwards.
//
// It lets go of whoever is being followed for the same reason panning does
// not: zooming out to see the coast and having the view snatched back to a
// figure on the next tick is the view refusing the question. Following is
// one press of f away again.
func (v *view) zoom(by int) {
	if v.snap == nil || v.screen == nil {
		return
	}
	m := v.snap.Map
	sw, sh := v.screen.Size()
	z := v.cam.scale()
	w, h := mapArea(m, sw, sh, z)
	if w < minMapW || h < minMapH {
		return
	}
	next, ok := zoomed(m, sw, sh, z, by)
	if !ok {
		return
	}
	mid := entity.Pos{X: v.cam.x + w*z/2, Y: v.cam.y + h*z/2}
	v.cam.z, v.cam.lock = next, false
	nw, nh := mapArea(m, sw, sh, next)
	v.cam.center(m, mid, nw, nh)
}

// handleMouse picks the figure under a click. It is the shortest way from
// "what is that one doing over there" to an answer. The click is in screen
// cells and the figures are on the ground, so it goes back through the
// window the ground is being drawn in: on a globe the cell in the corner of
// the screen is not tile 0,0 and need not even be east of it.
//
// The wheel zooms, which is the one thing a wheel does on every map anybody
// has ever used.
func (v *view) handleMouse(ev *tcell.EventMouse) {
	if v.snap == nil {
		return
	}
	switch {
	case ev.Buttons()&tcell.WheelDown != 0:
		v.zoom(1)
		v.draw()
		return
	case ev.Buttons()&tcell.WheelUp != 0:
		v.zoom(-1)
		v.draw()
		return
	case ev.Buttons()&tcell.Button1 == 0:
		return
	}
	x, y := ev.Position()
	v.onMap(func(m *observe.MapView, w, h int) {
		win := v.cam.window(m, w, h)
		if _, on := win.Tile(m, x, y); !on {
			return
		}
		// Whoever is drawn in the cell that was clicked, which zoomed out is
		// whoever is anywhere in the block of ground under it. Asking the
		// window where each figure is drawn rather than which tile the click
		// was of is the same question put the way the picture answers it.
		for _, mark := range m.Agents {
			if mx, my, ok := win.Screen(m, mark.Pos); ok && mx == x && my == y {
				v.choose(mark.ID)
				v.draw()
				return
			}
		}
	})
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
	ascii.Ice:         tcell.StyleDefault.Foreground(tcell.PaletteColor(195)),
	ascii.Field:       tcell.StyleDefault.Foreground(tcell.ColorYellow),
	ascii.FieldFenced: tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true),
	ascii.House:       tcell.StyleDefault.Foreground(tcell.ColorWhite).Bold(true),
	ascii.Market:      tcell.StyleDefault.Foreground(tcell.ColorFuchsia).Bold(true),
	ascii.Road:        tcell.StyleDefault.Foreground(tcell.Color137),
	ascii.Rock:        tcell.StyleDefault.Foreground(tcell.ColorGray),
	ascii.RockHigh:    tcell.StyleDefault.Foreground(tcell.PaletteColor(250)),
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
	// Both stay green through the lowland and leave it only above the lowland,
	// which is the same split the ground itself is built on: bands 0 to 2 are
	// the valley's own sixty metres and bands 3 to 5 are the high country
	// standing over it. The first try desaturated from the second band on, and
	// on the map that made two thirds of the ground a grey-green and a khaki -
	// a settlement lives in the valley, so the valley is most of what is on
	// screen, and washing it out washed out the whole picture to buy
	// distinctions among mountains that are a twentieth of it.
	//
	// The woods stay green the whole way up, because a wood is green: a
	// hillside of trees should read as a wooded hillside and not as another
	// shade of rock. What rises with them is brightness, which is what thinner
	// montane cover actually looks like.
	ascii.Ground0: tcell.StyleDefault.Foreground(tcell.PaletteColor(22)),  // the water meadow
	ascii.Ground1: tcell.StyleDefault.Foreground(tcell.PaletteColor(28)),  // the valley floor
	ascii.Ground2: tcell.StyleDefault.Foreground(tcell.PaletteColor(64)),  // the dry shoulder of it
	ascii.Ground3: tcell.StyleDefault.Foreground(tcell.PaletteColor(101)), // the foothills
	ascii.Ground4: tcell.StyleDefault.Foreground(tcell.PaletteColor(145)), // the mountainside
	ascii.Ground5: tcell.StyleDefault.Foreground(tcell.PaletteColor(252)), // the tops

	ascii.Wood0: tcell.StyleDefault.Foreground(tcell.PaletteColor(22)),
	ascii.Wood1: tcell.StyleDefault.Foreground(tcell.PaletteColor(28)),
	ascii.Wood2: tcell.StyleDefault.Foreground(tcell.PaletteColor(34)),
	ascii.Wood3: tcell.StyleDefault.Foreground(tcell.PaletteColor(40)),
	ascii.Wood4: tcell.StyleDefault.Foreground(tcell.PaletteColor(71)),
	ascii.Wood5: tcell.StyleDefault.Foreground(tcell.PaletteColor(108)),

	// The three ramps the readings are drawn in. Each is one hue darkening
	// or brightening the whole way, because a reading is a quantity and the
	// eye reads a quantity out of one colour getting stronger far better than
	// out of a rainbow: what is wanted here is "more than there", not "a
	// different kind of thing from there".
	//
	// Moisture runs from the grey of dry ground into deep water-blue.
	ascii.Wet0: tcell.StyleDefault.Foreground(tcell.PaletteColor(101)),
	ascii.Wet1: tcell.StyleDefault.Foreground(tcell.PaletteColor(66)),
	ascii.Wet2: tcell.StyleDefault.Foreground(tcell.PaletteColor(31)),
	ascii.Wet3: tcell.StyleDefault.Foreground(tcell.PaletteColor(32)),
	ascii.Wet4: tcell.StyleDefault.Foreground(tcell.PaletteColor(33)),
	ascii.Wet5: tcell.StyleDefault.Foreground(tcell.PaletteColor(39)),

	// Soil runs from bare tan into the green of ground worth ploughing.
	ascii.Crop0: tcell.StyleDefault.Foreground(tcell.PaletteColor(137)),
	ascii.Crop1: tcell.StyleDefault.Foreground(tcell.PaletteColor(143)),
	ascii.Crop2: tcell.StyleDefault.Foreground(tcell.PaletteColor(107)),
	ascii.Crop3: tcell.StyleDefault.Foreground(tcell.PaletteColor(71)),
	ascii.Crop4: tcell.StyleDefault.Foreground(tcell.PaletteColor(40)),
	ascii.Crop5: tcell.StyleDefault.Foreground(tcell.PaletteColor(46)),

	// Wear runs from ground nobody crosses into the bright of a thoroughfare.
	ascii.Worn0: tcell.StyleDefault.Foreground(tcell.PaletteColor(238)),
	ascii.Worn1: tcell.StyleDefault.Foreground(tcell.PaletteColor(94)),
	ascii.Worn2: tcell.StyleDefault.Foreground(tcell.PaletteColor(136)),
	ascii.Worn3: tcell.StyleDefault.Foreground(tcell.PaletteColor(178)),
	ascii.Worn4: tcell.StyleDefault.Foreground(tcell.PaletteColor(214)),
	ascii.Worn5: tcell.StyleDefault.Foreground(tcell.PaletteColor(220)),

	// Fish runs from the dark of water that has been emptied into the bright
	// of water that is full. It is a teal rather than the moisture blue on
	// purpose: the two readings are both about water and would otherwise be
	// the same picture at a glance.
	ascii.Shoal0: tcell.StyleDefault.Foreground(tcell.PaletteColor(23)),
	ascii.Shoal1: tcell.StyleDefault.Foreground(tcell.PaletteColor(29)),
	ascii.Shoal2: tcell.StyleDefault.Foreground(tcell.PaletteColor(36)),
	ascii.Shoal3: tcell.StyleDefault.Foreground(tcell.PaletteColor(43)),
	ascii.Shoal4: tcell.StyleDefault.Foreground(tcell.PaletteColor(50)),
	ascii.Shoal5: tcell.StyleDefault.Foreground(tcell.PaletteColor(51)),

	// Holdings are not a ramp. They stand for different owners rather than
	// for more and less of one thing, so they are six hues chosen to be told
	// apart rather than to be put in order - a scale here would say that one
	// farmer is somehow more than another.
	ascii.Held0: tcell.StyleDefault.Foreground(tcell.PaletteColor(203)),
	ascii.Held1: tcell.StyleDefault.Foreground(tcell.PaletteColor(214)),
	ascii.Held2: tcell.StyleDefault.Foreground(tcell.PaletteColor(227)),
	ascii.Held3: tcell.StyleDefault.Foreground(tcell.PaletteColor(84)),
	ascii.Held4: tcell.StyleDefault.Foreground(tcell.PaletteColor(87)),
	ascii.Held5: tcell.StyleDefault.Foreground(tcell.PaletteColor(177)),

	// Ground nobody has claimed, and the quietest thing on the map.
	ascii.Bare: tcell.StyleDefault.Foreground(tcell.PaletteColor(238)),
}

func (v *view) draw() {
	if v.busy != nil && v.busy() {
		return // whatever comes next will draw, and this frame is already old
	}
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
	// How much of the map is on the screen. A map that fits is drawn whole,
	// as it always was; one that does not — a globe is sixteen chunks round
	// and no terminal is — is drawn through a window that moves. What the
	// view refuses now is a screen with no room for a map at all, rather
	// than every screen too small to hold the whole of one.
	mw, mh := mapArea(s.Map, sw, sh, v.cam.scale())
	if mw < minMapW || mh < minMapH {
		need := minMapW + panelWidth
		puts(sc, 0, 0, tcell.StyleDefault, fmt.Sprintf("terminal too small: need %dx%d, have %dx%d", need, minMapH+graphHeight+3, sw, sh))
		sc.Show()
		return
	}
	// Where to look, before anything is drawn: at the settlement on the
	// first frame, and afterwards at whoever is being followed, unless the
	// user has taken the view somewhere themselves.
	if !v.cam.placed {
		v.cam.center(s.Map, s.Map.Market, mw, mh)
	}
	if v.cam.lock && v.sel != 0 {
		for _, m := range s.Map.Agents {
			if m.ID == v.sel {
				v.cam.center(s.Map, m.Pos, mw, mh)
				break
			}
		}
	}
	win := v.cam.window(s.Map, mw, mh)

	for y, row := range ascii.RenderWindow(s.Map, v.view, win) {
		for x, c := range row {
			sc.SetContent(x, y, c.Ch, nil, palette[c.Color])
		}
	}
	// The figure being followed is turned inside out on the map, so that the
	// panel beside it is plainly about somebody in particular and the eye
	// can find them again in the crowd after they have walked.
	for _, m := range s.Map.Agents {
		if m.ID == v.sel {
			if x, y, ok := win.Screen(s.Map, m.Pos); ok {
				sc.SetContent(x, y, '@', nil, palette[ascii.AgentColor(m.Action)].Reverse(true))
			}
			break
		}
	}

	// The keys stand at the foot of the panel, in the panel's own width:
	// what is written past the edge of the screen is not a reminder of
	// anything. The line about looking around is only there when there is
	// somewhere else to look, which on a map drawn whole there is not.
	keys := []string{"space pause  +/- speed  . step", "m map  r pave  q quit"}
	if v.looking() {
		keys = append(keys, "arrows look  c settlement  f follow")
	}
	// The scales are worth naming wherever there is one to step to, which on
	// a globe zoomed all the way out is inward and nowhere else.
	if _, out := zoomed(s.Map, sw, sh, v.cam.scale(), 1); out || v.cam.scale() > 1 {
		keys = append(keys, "z/Z zoom out and in")
	}
	keys = append(keys, "tab pick  d vitals  w world", v.graphKeys())

	px := mw + 2
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
	// What the settlement has worked out, one to a line, with the year it
	// came to it and the year it first had a master of the craft. It used
	// to be a single line of names, which at eleven technologies was a
	// comma and a truncation: the names are the least interesting part,
	// because a settlement that has held masonry since year four and still
	// has no mason is in a different position from one that mastered it
	// last spring, and the old line could not tell them apart.
	//
	// The panel is thirty-eight columns and shares its height with the card
	// of whoever is being followed, so this takes what it can and counts
	// the rest. Newest first: the old ones are on the world page and what
	// is worth seeing here is what has just happened.
	if len(s.Worked) == 0 {
		put(tcell.StyleDefault.Dim(true), "worked out: nothing yet")
	} else {
		put(tcell.StyleDefault, "worked out:")
		room := max(1, (sh-len(keys)-1-line)/2)
		shown := 0
		for i := len(s.Worked) - 1; i >= 0 && shown < room; i-- {
			f := s.Worked[i]
			mark, style := "—", tcell.StyleDefault.Dim(true)
			if f.Mastered != 0 {
				mark, style = fmt.Sprintf("y%d", clock.At(f.Mastered).Year), tcell.StyleDefault
			}
			put(style, " %-12s y%-4d %s", trim(string(f.Tech), 12), clock.At(f.Found).Year, mark)
			shown++
		}
		if rest := len(s.Worked) - shown; rest > 0 {
			put(tcell.StyleDefault.Dim(true), " and %d more, on the world page", rest)
		}
	}
	line++
	if v.sel != 0 {
		v.drawCard(px, &line, sh-len(keys)-1)
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
	//
	// Under a reading the row says what the reading is instead, because the
	// map above it is no longer showing anybody working: a legend about work
	// over a picture of the soil is worse than no legend.
	if v.view != ascii.Settlement {
		v.drawReading(mw, mh)
	} else {
		v.drawWorkLegend(s, mw, mh, dim)
	}
	span := fmt.Sprintf("%d days →", min(len(v.hist), mw)*graphTicks)
	if g, ok := v.focused(); ok {
		v.drawOpen(0, mh+1, mw, graphHeight, g)
		span = trim(v.headline(g)+"   "+span, mw)
	} else {
		v.drawGraph(0, mh+1, mw, graphHeight)
	}
	// On a map larger than the screen the span shares its line with where
	// the window is standing: on a globe every view looks alike, and a
	// reader who cannot say which corner of the world is on the screen
	// cannot come back to it either.
	puts(sc, 0, mh+1+graphHeight, dim, span)
	if z := v.cam.scale(); z > 1 || mw < s.Map.W || mh < s.Map.H {
		at := fmt.Sprintf("looking at %d,%d of %dx%d", win.X+mw*z/2, win.Y+mh*z/2, s.Map.W, s.Map.H)
		// The scale, whenever it is not the one everything else assumes. A
		// map drawn at eight tiles to the cell and one drawn at one look
		// alike, and a reader who takes the second for the first has the
		// world eight times the size it is.
		if z > 1 {
			at += fmt.Sprintf(" at %d tiles to the cell", z)
		}
		if v.cam.lock && v.sel != 0 {
			at += " (following)"
		}
		// It shares the row with the graph's span and gives way to it: two
		// lines of dim text run together say less than either of them.
		if room := mw - len([]rune(span)) - 2; room >= len([]rune(at)) {
			puts(sc, mw-len([]rune(at)), mh+1+graphHeight, dim, at)
		}
	}
	for i, line := range keys {
		puts(sc, px, sh-len(keys)+i, dim, trim(line, sw-px))
	}
	// A settlement that has ended says so across the empty map it left, and
	// says where to go and read why. Without this the map simply stops
	// moving and a finished run looks like a hung one.
	if s.Population == 0 && v.gone != 0 {
		note := fmt.Sprintf("the settlement died out at tick %d — press d for why", v.gone)
		puts(sc, max(0, (mw-len([]rune(note)))/2), mh/2,
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

// drawWorkLegend is the row under the map on the settlement view: every kind
// of work, always in the same place in the same colour.
func (v *view) drawWorkLegend(s *observe.Snapshot, w, h int, dim tcell.Style) {
	sc := v.screen
	var counts [len(ascii.Groups)]int
	for _, a := range s.Activity {
		counts[ascii.GroupOf(a.Action)] += a.Agents
	}
	cell := w / len(ascii.Groups)
	for i, g := range ascii.Groups {
		lx := i * cell
		puts(sc, lx, h, palette[g.Color], "█")
		// A kind of work nobody is doing is dim, and so is one that is
		// simply not the one being read: while a band is opened out the
		// legend says which of them it is. See focus.go.
		style := v.markGroup(i)
		if counts[i] == 0 && v.focus == 0 {
			style = dim
		}
		puts(sc, lx+2, h, style, trim(fmt.Sprintf("%s %d", g.Name, counts[i]), cell-3))
	}
}

// drawReading is that same row under a reading: what is being read, and which
// way the shading runs. A ramp whose ends are not named is a pattern and not
// a map - the reader can see that one place differs from another and cannot
// tell which of them is the good ground.
func (v *view) drawReading(w, y int) {
	sc := v.screen
	r := ascii.Views[v.view]
	x := 0
	put := func(style tcell.Style, text string) {
		puts(sc, x, y, style, text)
		x += len(text)
	}
	put(tcell.StyleDefault.Bold(true), r.Name)
	dim := tcell.StyleDefault.Dim(true)
	swatches := func() {
		for i := 0; i < ascii.Bands; i++ {
			puts(sc, x, y, palette[ascii.RampOf(v.view)[i]], "█")
			x++
		}
	}
	// A key is not a scale, so it is not given ends to read between: the
	// colours stand for different holders and putting "low" at one end of
	// them would say that one holding is less than another.
	if r.Legend == ascii.Key {
		put(dim, "   ")
		swatches()
		put(dim, " "+r.Says)
	} else {
		put(dim, "   "+r.Low+" ")
		swatches()
		put(dim, " "+r.High)
	}
	put(dim, "   m for the next, M for the last")
}
